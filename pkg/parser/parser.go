package parser

import (
	"os"

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

	// 验证配置
	if len(config.Services) == 0 {
		return nil, errors.InvalidConfig("Docker Compose 文件中没有定义服务")
	}

	return &config, nil
}
