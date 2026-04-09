package parser

import (
	"os"
	"strings"

	"github.com/wenwenxiong/HelmForge/internal/models"
	"github.com/wenwenxiong/HelmForge/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ParseDockerCompose 解析 Docker Compose 文件
func ParseDockerCompose(filePath string) (*models.DockerComposeConfig, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.FileReadError(filePath, err)
	}

	// 解析 YAML 文件
	var config models.DockerComposeConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidFormat, "解析 YAML 文件失败")
	}

	// 预处理环境变量：将数组格式转换为map格式
	// 使用新的map来避免直接修改原有数据
	processedServices := make(map[string]models.Service)
	for serviceName, service := range config.Services {
		processedService := service // 创建副本

		if len(processedService.Environment) > 0 {
			envMap := make(map[string]string)
			for _, env := range processedService.Environment {
				// 解析 KEY=VALUE 格式
				if strings.Contains(env, "=") {
					parts := strings.SplitN(env, "=", 2)
					if len(parts) == 2 {
						envMap[parts[0]] = parts[1]
					}
				} else {
					// 处理只有KEY的情况
					envMap[env] = ""
				}
			}
			// 转换为map格式
			processedService.Environment = nil
			processedService.EnvVars = envMap
		}

		processedServices[serviceName] = processedService
	}

	// 使用处理后的服务配置替换原有配置
	config.Services = processedServices

	// 验证配置
	if len(config.Services) == 0 {
		return nil, errors.InvalidConfig("Docker Compose 文件中没有定义服务")
	}

	return &config, nil
}
