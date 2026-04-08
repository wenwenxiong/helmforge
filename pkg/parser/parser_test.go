package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wenwenxiong/HelmForge/internal/models"
)

func TestParseDockerCompose(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试用的 docker-compose 文件
	dockerComposeContent := `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      - NGINX_HOST=localhost
    volumes:
      - ./data:/data
  api:
    image: node:16-alpine
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    depends_on:
      - db
  db:
    image: postgres:13-alpine
    environment:
      - POSTGRES_DB=app
      - POSTGRES_USER=app
      - POSTGRES_PASSWORD=secret
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  postgres-data:

networks:
  default:
    name: app-network
`
	dockerComposeFile := filepath.Join(tempDir, "docker-compose.yaml")
	err := os.WriteFile(dockerComposeFile, []byte(dockerComposeContent), 0644)
	require.NoError(t, err, "创建测试文件失败")

	// 测试解析
	config, err := ParseDockerCompose(dockerComposeFile)
	require.NoError(t, err, "解析 Docker Compose 文件失败")
	require.NotNil(t, config, "解析结果不应为 nil")

	// 验证基本信息
	assert.Equal(t, "3.8", config.Version, "版本号应该正确")
	assert.Len(t, config.Services, 3, "应该解析出3个服务")
	assert.Len(t, config.Volumes, 1, "应该解析出1个卷")
	assert.Len(t, config.Networks, 1, "应该解析出1个网络")

	// 验证 web 服务
	webService, ok := config.Services["web"]
	require.True(t, ok, "应该存在 web 服务")
	assert.Equal(t, "nginx:latest", webService.Image, "镜像应该正确")
	assert.Len(t, webService.Ports, 1, "应该有1个端口映射")
	assert.Equal(t, "8080:80", webService.Ports[0], "端口映射应该正确")
	assert.Len(t, webService.Environment, 1, "应该有1个环境变量")
	assert.Equal(t, "localhost", webService.Environment["NGINX_HOST"], "环境变量应该正确")
	assert.Len(t, webService.Volumes, 1, "应该有1个卷映射")

	// 验证 api 服务
	apiService, ok := config.Services["api"]
	require.True(t, ok, "应该存在 api 服务")
	assert.Equal(t, "node:16-alpine", apiService.Image, "镜像应该正确")
	assert.Len(t, apiService.Ports, 1, "应该有1个端口映射")
	assert.Len(t, apiService.DependsOn, 1, "应该有1个依赖")
	assert.Equal(t, "db", apiService.DependsOn[0], "依赖应该正确")

	// 验证 db 服务
	dbService, ok := config.Services["db"]
	require.True(t, ok, "应该存在 db 服务")
	assert.Equal(t, "postgres:13-alpine", dbService.Image, "镜像应该正确")
	assert.Len(t, dbService.Environment, 3, "应该有3个环境变量")
	assert.Equal(t, "secret", dbService.Environment["POSTGRES_PASSWORD"], "密码应该正确")
}

func TestParseDockerCompose_InvalidFile(t *testing.T) {
	// 测试文件不存在的情况
	_, err := ParseDockerCompose("/nonexistent/file.yaml")
	assert.Error(t, err, "文件不存在应该返回错误")
}

func TestParseDockerCompose_EmptyFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建空文件
	emptyFile := filepath.Join(tempDir, "empty.yaml")
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	require.NoError(t, err)

	// 测试解析空文件
	_, err = ParseDockerCompose(emptyFile)
	assert.Error(t, err, "空文件应该返回错误")
}

func TestParseDockerCompose_NoServices(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建没有服务的文件
	noServicesContent := `
version: '3.8'
networks:
  default:
    name: app-network
`
	noServicesFile := filepath.Join(tempDir, "no-services.yaml")
	err := os.WriteFile(noServicesFile, []byte(noServicesContent), 0644)
	require.NoError(t, err)

	// 测试解析没有服务的文件
	_, err = ParseDockerCompose(noServicesFile)
	assert.Error(t, err, "没有服务应该返回错误")
	assert.Contains(t, err.Error(), "没有定义服务", "错误信息应该包含特定内容")
}

func TestParseDockerCompose_WithHealthcheck(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建带有健康检查的配置
	healthcheckContent := `
version: '3.8'
services:
  app:
    image: app:latest
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
`
	healthcheckFile := filepath.Join(tempDir, "healthcheck.yaml")
	err := os.WriteFile(healthcheckFile, []byte(healthcheckContent), 0644)
	require.NoError(t, err)

	// 测试解析健康检查
	config, err := ParseDockerCompose(healthcheckFile)
	require.NoError(t, err, "解析健康检查配置失败")
	require.NotNil(t, config.Services["app"].Healthcheck, "健康检查配置不应为 nil")

	// 验证健康检查配置
	healthcheck := config.Services["app"].Healthcheck
	assert.Len(t, healthcheck.Test, 3, "测试命令应该有3个部分")
	assert.Equal(t, "30s", healthcheck.Interval, "间隔时间应该正确")
	assert.Equal(t, "10s", healthcheck.Timeout, "超时时间应该正确")
	assert.Equal(t, 3, healthcheck.Retries, "重试次数应该正确")
	assert.Equal(t, "40s", healthcheck.StartPeriod, "启动期应该正确")
}

func TestParseDockerCompose_WithBuild(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建带有构建配置的文件
	buildContent := `
version: '3.8'
services:
  app:
    build:
      context: ./app
      dockerfile: Dockerfile.dev
      args:
        - BUILD_ENV=development
        - NODE_ENV=dev
`
	buildFile := filepath.Join(tempDir, "build.yaml")
	err := os.WriteFile(buildFile, []byte(buildContent), 0644)
	require.NoError(t, err)

	// 测试解析构建配置
	config, err := ParseDockerCompose(buildFile)
	require.NoError(t, err, "解析构建配置失败")
	require.NotNil(t, config.Services["app"].Build, "构建配置不应为 nil")

	// 验证构建配置
	build := config.Services["app"].Build
	assert.Equal(t, "./app", build.Context, "构建上下文应该正确")
	assert.Equal(t, "Dockerfile.dev", build.Dockerfile, "Dockerfile 应该正确")
	assert.Len(t, build.Args, 2, "构建参数应该有2个")
	assert.Equal(t, "development", build.Args["BUILD_ENV"], "构建参数应该正确")
}
