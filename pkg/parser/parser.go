package parser

import (
	"fmt"
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
	// 创建全新的配置对象，避免修改原map导致的range循环副作用
	fmt.Printf("=== 开始环境变量预处理 ===\n")
	fmt.Printf("原始服务数量: %d\n", len(config.Services))

	newConfig := &models.DockerComposeConfig{
		Version:  config.Version,
		Services: make(map[string]models.Service),
		Networks: config.Networks,
		Volumes:  config.Volumes,
	}

	// 第一步：完全复制所有服务，包括所有字段
	for serviceName, service := range config.Services {
		fmt.Printf("处理服务: %s\n", serviceName)
		fmt.Printf("  原始 - Image: %s\n", service.Image)
		fmt.Printf("  原始 - Environment: %v (len: %d, isNil: %v)\n",
			service.Environment, len(service.Environment), service.Environment == nil)
		fmt.Printf("  原始 - DependsOn: %v (len: %d)\n",
			service.DependsOn, len(service.DependsOn))

		// 完整复制所有字段，确保没有任何字段丢失
		newService := models.Service{
			Image:       service.Image,
			Build:       service.Build,
			Ports:       service.Ports,
			Environment: service.Environment, // 先保持原始状态
			Volumes:     service.Volumes,
			Networks:    service.Networks,
			DependsOn:   service.DependsOn, // 确保这个字段被完整复制
			Healthcheck: service.Healthcheck,
			EnvVars:     make(map[string]string), // 初始化空的EnvVars
		}

		// 第二步：处理环境变量数组格式
		if len(service.Environment) > 0 {
			fmt.Printf("  开始处理环境变量，数量: %d\n", len(service.Environment))
			envMap := make(map[string]string)

			for _, env := range service.Environment {
				fmt.Printf("    处理环境变量: %s\n", env)

				// 支持两种格式：KEY=VALUE 和 KEY
				if strings.Contains(env, "=") {
					// 格式：KEY=VALUE
					parts := strings.SplitN(env, "=", 2)
					if len(parts) == 2 {
						fmt.Printf("      解析为: key=%s, value=%s\n", parts[0], parts[1])
						envMap[parts[0]] = parts[1]
					} else {
						fmt.Printf("      格式不正确，跳过\n")
						continue
					}
				} else {
					// 格式：KEY（只有键名，值为空字符串）
					fmt.Printf("      无等号，设置为: key=%s, value=\"\"\n", env)
					envMap[env] = ""
				}
			}

			fmt.Printf("  环境变量处理完成，共%d个变量\n", len(envMap))

			// 转换：清空Environment数组，设置EnvVars map
			newService.Environment = nil
			newService.EnvVars = envMap
		} else {
			fmt.Printf("  没有环境变量需要处理\n")
		}

		// 验证关键字段完整性
		fmt.Printf("  处理后 - DependsOn: %v (len: %d)\n",
			newService.DependsOn, len(newService.DependsOn))
		fmt.Printf("  处理后 - EnvVars: %v (len: %d)\n",
			newService.EnvVars, len(newService.EnvVars))

		newConfig.Services[serviceName] = newService
		fmt.Printf(" 服务 %s 已添加到新配置\n", serviceName)
	}

	// 验证新配置的服务数量
	if len(newConfig.Services) == 0 {
		fmt.Printf("警告：新配置中没有任何服务！\n")
	}

	// 验证配置的服务数量（检查config，因为已经将newConfig赋值给了config）
	if len(config.Services) == 0 {
		return nil, errors.InvalidConfig("Docker Compose 文件中没有定义服务")
	}

	fmt.Printf("=== 环境变量预处理完成 ===\n")
	fmt.Printf("最终服务数量: %d\n", len(config.Services))

	// 验证配置
	if len(newConfig.Services) == 0 {
		return nil, errors.InvalidConfig("Docker Compose 文件中没有定义服务")
	}

	return newConfig, nil
}
