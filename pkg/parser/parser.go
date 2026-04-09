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
	// Go语言规范：len()对nil切片返回0，因此只需要检查len > 0
	// 这样可以同时处理nil和[]string{}两种情况
	processedServices := make(map[string]models.Service)
	for serviceName, service := range config.Services {
		// 创建服务副本，避免range循环中修改原map导致副作用
		processedService := service

		// 处理环境变量数组格式：[KEY=VALUE, ...]
		// 转换为map格式：EnvVars: {KEY: VALUE, ...}
		if len(processedService.Environment) > 0 {
			envMap := make(map[string]string)

			// 解析每个环境变量
			for _, env := range processedService.Environment {
				// 支持两种格式：KEY=VALUE 和 KEY
				if strings.Contains(env, "=") {
					// 格式：KEY=VALUE
					parts := strings.SplitN(env, "=", 2)
					if len(parts) == 2 {
						envMap[parts[0]] = parts[1]
					} else {
						// 格式不正确，忽略这个环境变量
						continue
					}
				} else {
					// 格式：KEY（只有键名，值为空字符串）
					envMap[env] = ""
				}
			}

			// 转换为map格式：清空Environment数组，设置EnvVars map
			processedService.Environment = nil
			processedService.EnvVars = envMap
		}

		// 将处理后的服务添加到新的map中
		processedServices[serviceName] = processedService
	}

	// 使用处理后的服务配置替换原有配置
	// 这样可以确保所有服务字段都正确保留，避免range循环副作用
	config.Services = processedServices

	// 验证配置
	if len(config.Services) == 0 {
		return nil, errors.InvalidConfig("Docker Compose 文件中没有定义服务")
	}

	return &config, nil
}
