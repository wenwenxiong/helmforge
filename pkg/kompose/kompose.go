package kompose

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	// "github.com/kubernetes/kompose/pkg/kobject" // 暂时禁用，包接口已变更
	// "github.com/kubernetes/kompose/pkg/transform" // 暂时禁用，包结构已变化
	"github.com/wenwenxiong/HelmForge/internal/models"
	"gopkg.in/yaml.v3"
)

// KubernetesResource Kubernetes 资源结构
type KubernetesResource struct {
	Kind       string                 `yaml:"kind"`
	APIVersion string                 `yaml:"apiVersion"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

// ConvertToKubernetes 将 Docker Compose 配置转换为 Kubernetes 资源
func ConvertToKubernetes(composeConfig *models.DockerComposeConfig) ([]KubernetesResource, error) {
	fmt.Println("开始转换 Docker Compose 配置为 Kubernetes 资源")

	// 为转换创建临时目录
	tempDir, err := os.MkdirTemp("", "helmforge-kompose-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 将 Docker Compose 配置写入临时文件
	composeFile := filepath.Join(tempDir, "docker-compose.yaml")
	content, err := yaml.Marshal(composeConfig)
	if err != nil {
		return nil, fmt.Errorf("序列化 Docker Compose 配置失败: %v", err)
	}

	if err := os.WriteFile(composeFile, content, 0644); err != nil {
		return nil, fmt.Errorf("写入临时 Docker Compose 文件失败: %v", err)
	}

	// 使用 Kompose Go 包进行转换
	resources, err := convertUsingKomposePackage(composeFile)
	if err != nil {
		// 如果 Kompose Go 包不可用，尝试调用命令行工具
		fmt.Printf("警告: Kompose Go 包不可用，尝试使用命令行工具: %v\n", err)
		resources, err = RunKomposeConvert(composeFile, tempDir)
		if err != nil {
			// 如果命令行工具也不可用，使用模拟实现
			fmt.Printf("警告: Kompose 工具不可用，使用模拟实现: %v\n", err)
			resources, err = generateKubernetesResources(composeConfig, tempDir)
			if err != nil {
				return nil, fmt.Errorf("生成 Kubernetes 资源失败: %v", err)
			}
		}
	}

	fmt.Printf("✓ 成功转换为 Kubernetes 资源，生成 %d 个资源文件\n", len(resources))
	return resources, nil
}

// convertUsingKomposePackage 使用 Kompose Go 包进行转换
func convertUsingKomposePackage(composeFile string) ([]KubernetesResource, error) {
	fmt.Println("Kompose Go 包暂时不可用，使用命令行工具")
	return nil, fmt.Errorf("Kompose Go 包接口已变更，暂时使用命令行工具")
}

// RunKomposeConvert 运行 kompose convert 命令
func RunKomposeConvert(composeFile, outputDir string) ([]KubernetesResource, error) {
	fmt.Println("执行 kompose convert 命令...")

	// 检查 Kompose 是否已安装
	if !IsKomposeInstalled() {
		return nil, fmt.Errorf("Kompose 工具未安装，请先安装 Kompose (https://kompose.io/)")
	}

	// 构建命令参数
	args := []string{"convert", "-f", composeFile, "-o", outputDir}

	// 执行命令
	cmd := exec.Command("kompose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("kompose convert 命令执行失败: %v", err)
	}

	// 读取生成的 Kubernetes 资源文件
	resources, err := readKubernetesResources(outputDir)
	if err != nil {
		return nil, fmt.Errorf("读取 Kubernetes 资源文件失败: %v", err)
	}

	return resources, nil
}

// IsKomposeInstalled 检查 Kompose 是否已安装
func IsKomposeInstalled() bool {
	cmd := exec.Command("kompose", "version")
	err := cmd.Run()
	return err == nil
}

// readKubernetesResources 从输出目录读取 Kubernetes 资源文件
func readKubernetesResources(outputDir string) ([]KubernetesResource, error) {
	var resources []KubernetesResource

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 只处理 YAML 文件
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		// 跳过辅助文件
		if strings.HasPrefix(info.Name(), "_") {
			return nil
		}

		// 读取并解析 YAML 文件
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var resource KubernetesResource
		if err := yaml.Unmarshal(content, &resource); err != nil {
			return err
		}

		resources = append(resources, resource)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

// RunKomposeConvertDirect 直接运行 kompose convert 命令（不通过临时文件）
func RunKomposeConvertDirect(composeFile, outputDir string, withHelm bool) error {
	fmt.Println("直接执行 kompose convert 命令...")

	// 检查 Kompose 是否已安装
	if !IsKomposeInstalled() {
		return fmt.Errorf("Kompose 工具未安装，请先安装 Kompose (https://kompose.io/)")
	}

	// 构建命令参数
	args := []string{"convert", "-f", composeFile, "-o", outputDir}

	if withHelm {
		args = append(args, "--with-helm")
	}

	// 执行命令
	cmd := exec.Command("kompose", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("kompose convert 命令执行失败: %v\nstderr: %s", err, stderr.String())
	}

	fmt.Printf("✓ Kompose 转换成功，输出目录: %s\n", outputDir)
	return nil
}

// generateKubernetesResources 生成 Kubernetes 资源（模拟实现，当 Kompose 不可用时使用）
func generateKubernetesResources(composeConfig *models.DockerComposeConfig, outputDir string) ([]KubernetesResource, error) {
	var resources []KubernetesResource

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 为每个服务生成 Deployment 和 Service
	for serviceName, service := range composeConfig.Services {
		// 生成 Deployment
		deployment := generateDeployment(serviceName, service)
		resources = append(resources, deployment)

		// 写入 Deployment 文件
		deploymentFile := filepath.Join(outputDir, fmt.Sprintf("%s-deployment.yaml", serviceName))
		content, err := yaml.Marshal(deployment)
		if err != nil {
			return nil, fmt.Errorf("序列化 Deployment 失败: %v", err)
		}
		if err := os.WriteFile(deploymentFile, content, 0644); err != nil {
			return nil, fmt.Errorf("写入 Deployment 文件失败: %v", err)
		}

		// 生成 Service（如果有端口映射）
		if len(service.Ports) > 0 {
			svc := generateService(serviceName, service)
			resources = append(resources, svc)

			// 写入 Service 文件
			serviceFile := filepath.Join(outputDir, fmt.Sprintf("%s-service.yaml", serviceName))
			content, err := yaml.Marshal(svc)
			if err != nil {
				return nil, fmt.Errorf("序列化 Service 失败: %v", err)
			}
			if err := os.WriteFile(serviceFile, content, 0644); err != nil {
				return nil, fmt.Errorf("写入 Service 文件失败: %v", err)
			}
		}
	}

	return resources, nil
}

// generateDeployment 生成 Deployment 资源
func generateDeployment(serviceName string, service models.Service) KubernetesResource {
	// 提取容器端口
	containerPorts := []map[string]interface{}{}
	for _, port := range service.Ports {
		portParts := strings.Split(port, "/")
		portMapping := strings.Split(portParts[0], ":")
		if len(portMapping) == 2 {
			containerPort := map[string]interface{}{
				"containerPort": parseInt(portMapping[1]),
				"protocol":      "TCP",
			}
			containerPorts = append(containerPorts, containerPort)
		}
	}

	// 提取环境变量
	envVars := []map[string]interface{}{}
	for key, value := range service.Environment {
		envVar := map[string]interface{}{
			"name":  key,
			"value": value,
		}
		envVars = append(envVars, envVar)
	}

	// 提取卷挂载
	volumeMounts := []map[string]interface{}{}
	volumes := []map[string]interface{}{}
	for i, volume := range service.Volumes {
		volumeParts := strings.Split(volume, ":")
		if len(volumeParts) >= 2 {
			hostPath := volumeParts[0]
			containerPath := volumeParts[1]

			volumeName := fmt.Sprintf("%s-volume-%d", serviceName, i)

			volumeMount := map[string]interface{}{
				"name":      volumeName,
				"mountPath": containerPath,
			}
			volumeMounts = append(volumeMounts, volumeMount)

			vol := map[string]interface{}{
				"name": volumeName,
				"hostPath": map[string]interface{}{
					"path": hostPath,
					"type": "DirectoryOrCreate",
				},
			}
			volumes = append(volumes, vol)
		}
	}

	// 构建 Deployment
	deployment := KubernetesResource{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
		Metadata: map[string]interface{}{
			"name": serviceName,
			"labels": map[string]interface{}{
				"app": serviceName,
			},
		},
		Spec: map[string]interface{}{
			"replicas": 1,
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": serviceName,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": serviceName,
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":            serviceName,
							"image":           service.Image,
							"ports":           containerPorts,
							"env":             envVars,
							"volumeMounts":    volumeMounts,
							"imagePullPolicy": "IfNotPresent",
						},
					},
					"volumes": volumes,
				},
			},
		},
	}

	return deployment
}

// generateService 生成 Service 资源
func generateService(serviceName string, service models.Service) KubernetesResource {
	// 提取端口
	ports := []map[string]interface{}{}
	for _, port := range service.Ports {
		portParts := strings.Split(port, "/")
		portMapping := strings.Split(portParts[0], ":")
		protocol := "TCP"
		if len(portParts) == 2 {
			protocol = strings.ToUpper(portParts[1])
		}

		if len(portMapping) == 2 {
			port := map[string]interface{}{
				"name":       fmt.Sprintf("port-%s", portMapping[1]),
				"port":       parseInt(portMapping[0]),
				"targetPort": parseInt(portMapping[1]),
				"protocol":   protocol,
			}
			ports = append(ports, port)
		}
	}

	// 构建 Service
	svc := KubernetesResource{
		Kind:       "Service",
		APIVersion: "v1",
		Metadata: map[string]interface{}{
			"name": serviceName,
			"labels": map[string]interface{}{
				"app": serviceName,
			},
		},
		Spec: map[string]interface{}{
			"selector": map[string]interface{}{
				"app": serviceName,
			},
			"ports": ports,
			"type":  "ClusterIP",
		},
	}

	return svc
}

// parseInt 将字符串转换为整数
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0 // 解析失败时返回默认值
	}
	return i
}
