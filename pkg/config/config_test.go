package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhanceHelmChart(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建基本的 Chart 结构
	chartDir := filepath.Join(tempDir, "test-chart")
	err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0755)
	require.NoError(t, err)

	// 创建基本的 Chart.yaml
	chartYamlContent := `
apiVersion: v2
name: test-chart
version: 0.1.0
appVersion: 1.0.0
`
	err = os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(chartYamlContent), 0644)
	require.NoError(t, err)

	// 创建基本的 values.yaml
	valuesYamlContent := `
image:
  repository: test-app
  tag: latest
service:
  type: ClusterIP
  port: 8080
`
	err = os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte(valuesYamlContent), 0644)
	require.NoError(t, err)

	// 创建一个模板文件
	templateContent := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    spec:
      containers:
      - name: test
        image: test-app:latest
`
	err = os.WriteFile(filepath.Join(chartDir, "templates", "deployment.yaml"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// 执行增强
	err = EnhanceHelmChart(chartDir)
	require.NoError(t, err, "增强 Chart 应该成功")

	// 验证增强后的文件是否存在
	_, err = os.Stat(filepath.Join(chartDir, "environments", "development.yaml"))
	assert.NoError(t, err, "应该生成开发环境配置")

	_, err = os.Stat(filepath.Join(chartDir, "environments", "testing.yaml"))
	assert.NoError(t, err, "应该生成测试环境配置")

	_, err = os.Stat(filepath.Join(chartDir, "environments", "production.yaml"))
	assert.NoError(t, err, "应该生成生产环境配置")

	_, err = os.Stat(filepath.Join(chartDir, "docs", "configuration.md"))
	assert.NoError(t, err, "应该生成配置文档")

	_, err = os.Stat(filepath.Join(chartDir, "templates", "NOTES.txt"))
	assert.NoError(t, err, "应该生成 NOTES.txt")
}

func TestEnhanceHelmChart_NonExistentDirectory(t *testing.T) {
	// 测试目录不存在的情况
	err := EnhanceHelmChart("/nonexistent/directory")
	assert.Error(t, err, "目录不存在应该返回错误")
	assert.Contains(t, err.Error(), "不存在", "错误信息应该包含特定内容")
}

func TestEnhanceHelmChart_MissingChartYaml(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建目录但不创建 Chart.yaml
	chartDir := filepath.Join(tempDir, "test-chart")
	err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0755)
	require.NoError(t, err)

	// 执行增强（应该失败）
	err = EnhanceHelmChart(chartDir)
	assert.Error(t, err, "缺少 Chart.yaml 应该返回错误")
}

func TestEnhanceValuesStructure(t *testing.T) {
	// 测试增强 values 结构
	values := make(map[string]interface{})

	// 初始为空
	enhanced := enhanceValuesStructure(values)

	// 验证基本结构已添加
	_, ok := enhanced["image"]
	assert.True(t, ok, "应该有 image 配置")

	_, ok = enhanced["service"]
	assert.True(t, ok, "应该有 service 配置")

	_, ok = enhanced["replicaCount"]
	assert.True(t, ok, "应该有 replicaCount 配置")

	_, ok = enhanced["resources"]
	assert.True(t, ok, "应该有 resources 配置")

	_, ok = enhanced["env"]
	assert.True(t, ok, "应该有 env 配置")

	_, ok = enhanced["volumes"]
	assert.True(t, ok, "应该有 volumes 配置")

	_, ok = enhanced["ingress"]
	assert.True(t, ok, "应该有 ingress 配置")

	_, ok = enhanced["autoscaling"]
	assert.True(t, ok, "应该有 autoscaling 配置")

	_, ok = enhanced["config"]
	assert.True(t, ok, "应该有 config 配置")

	// 验证默认值
	assert.Equal(t, 1, enhanced["replicaCount"], "默认副本数应该是1")

	image, ok := enhanced["image"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "IfNotPresent", image["pullPolicy"], "默认镜像拉取策略应该是 IfNotPresent")

	// 验证嵌套配置
	config, ok := enhanced["config"].(map[string]interface{})
	require.True(t, ok)

	appConfig, ok := config["app"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, appConfig["debug"], "默认调试模式应该是关闭")
	assert.Equal(t, "info", appConfig["logLevel"], "默认日志级别应该是 info")
}

func TestEnhanceValuesStructure_PreserveExisting(t *testing.T) {
	// 测试保留现有配置
	values := map[string]interface{}{
		"replicaCount": 5,
		"image": map[string]interface{}{
			"repository": "custom-app",
			"tag":        "v2.0.0",
		},
	}

	enhanced := enhanceValuesStructure(values)

	// 验证现有配置被保留
	assert.Equal(t, 5, enhanced["replicaCount"], "应该保留现有副本数")

	image, ok := enhanced["image"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "custom-app", image["repository"], "应该保留现有镜像仓库")
	assert.Equal(t, "v2.0.0", image["tag"], "应该保留现有镜像标签")

	// 验证缺失的配置被添加
	_, ok = enhanced["service"]
	assert.True(t, ok, "应该添加缺失的 service 配置")
}

func TestWriteYamlFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试写入 YAML 文件
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": map[string]string{
			"nested": "value",
		},
	}

	filePath := filepath.Join(tempDir, "test.yaml")
	err := writeYamlFile(filePath, data)
	require.NoError(t, err, "写入 YAML 文件应该成功")

	// 验证文件存在
	_, err = os.Stat(filePath)
	assert.NoError(t, err, "文件应该存在")

	// 读取并验证内容
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "key1: value1", "内容应该包含 key1")
	assert.Contains(t, string(content), "key2: 123", "内容应该包含 key2")
}

func TestGenerateEnvironmentConfigs(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建 environments 目录
	envDir := filepath.Join(tempDir, "environments")
	err := os.MkdirAll(envDir, 0755)
	require.NoError(t, err)

	// 生成环境配置
	err = generateEnvironmentConfigs(tempDir)
	require.NoError(t, err, "生成环境配置应该成功")

	// 验证文件存在
	files, err := os.ReadDir(envDir)
	require.NoError(t, err)
	assert.Len(t, files, 3, "应该生成3个环境配置文件")

	// 验证文件名
	fileNames := make(map[string]bool)
	for _, file := range files {
		fileNames[file.Name()] = true
	}

	assert.True(t, fileNames["development.yaml"], "应该有开发环境配置")
	assert.True(t, fileNames["testing.yaml"], "应该有测试环境配置")
	assert.True(t, fileNames["production.yaml"], "应该有生产环境配置")
}

func TestGenerateConfigDocumentation(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建 docs 目录
	docsDir := filepath.Join(tempDir, "docs")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	// 生成配置文档
	err = generateConfigDocumentation(tempDir)
	require.NoError(t, err, "生成配置文档应该成功")

	// 验证文件存在
	configDoc := filepath.Join(docsDir, "configuration.md")
	_, err = os.Stat(configDoc)
	assert.NoError(t, err, "配置文档应该存在")

	// 读取并验证内容
	content, err := os.ReadFile(configDoc)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# 配置文档", "应该有标题")
	assert.Contains(t, contentStr, "## 概述", "应该有概述章节")
	assert.Contains(t, contentStr, "## 配置结构", "应该有配置结构章节")
	assert.Contains(t, contentStr, "## 详细配置", "应该有详细配置章节")
	assert.Contains(t, contentStr, "## 多环境配置", "应该有多环境配置章节")
	assert.Contains(t, contentStr, "## 自定义配置", "应该有自定义配置章节")
}

func TestEnhanceHelpersTpl(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建 _helpers.tpl 文件
	helpersFile := filepath.Join(tempDir, "_helpers.tpl")
	err := os.WriteFile(helpersFile, []byte(`{{/*
Test template
*/}}
{{- define "test.template" -}}
test-value
{{- end }}
`), 0644)
	require.NoError(t, err)

	// 增强 _helpers.tpl
	err = enhanceHelpersTpl(helpersFile)
	require.NoError(t, err, "增强 _helpers.tpl 应该成功")

	// 读取并验证内容
	content, err := os.ReadFile(helpersFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "test.template", "应该保留原有模板")
	assert.Contains(t, contentStr, "helmforge.name", "应该添加 helmforge.name")
	assert.Contains(t, contentStr, "helmforge.fullname", "应该添加 helmforge.fullname")
	assert.Contains(t, contentStr, "helmforge.image", "应该添加 helmforge.image")
	assert.Contains(t, contentStr, "helmforge.service.port", "应该添加 helmforge.service.port")
}

func TestCreateNotesTxt(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建 NOTES.txt 文件
	notesFile := filepath.Join(tempDir, "NOTES.txt")
	err := createNotesTxt(notesFile)
	require.NoError(t, err, "创建 NOTES.txt 应该成功")

	// 验证文件存在
	_, err = os.Stat(notesFile)
	assert.NoError(t, err, "NOTES.txt 应该存在")

	// 读取并验证内容
	content, err := os.ReadFile(notesFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "helmforge.fullname", "应该包含模板函数引用")
	assert.Contains(t, contentStr, "helmforge.name", "应该包含模板函数引用")
	assert.Contains(t, contentStr, "Release.Name", "应该包含 Release 引用")
	assert.Contains(t, contentStr, "Chart.AppVersion", "应该包含 Chart 引用")
}
