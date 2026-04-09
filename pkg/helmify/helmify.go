package helmify

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	helmifyapp "github.com/arttor/helmify/pkg/app"
	helmifyconfig "github.com/arttor/helmify/pkg/config"
	"github.com/wenwenxiong/HelmForge/pkg/kompose"
	"gopkg.in/yaml.v3"
)

// HelmifyOptions Helmify 配置选项
type HelmifyOptions struct {
	OutputDir        string
	ChartName        string
	ChartVersion     string
	AppVersion       string
	Description      string
	KeepK8sManifests bool
	Namespace        string
	Release          string
	ValuesPrefix     string
	ExtraValues      []string
	SkipTests        bool
	NoHooks          bool
	IncludeCRDs      bool
	SkipDependencies bool
}

// ConvertToHelmChart 将 Kubernetes 资源转换为 Helm Chart
func ConvertToHelmChart(resources []kompose.KubernetesResource, outputDir string) error {
	fmt.Printf("开始转换 Kubernetes 资源为 Helm Chart，输出目录: %s\n", outputDir)

	// 默认选项
	options := HelmifyOptions{
		OutputDir:    outputDir,
		ChartName:    "helmforge-app",
		ChartVersion: "0.1.0",
		AppVersion:   "1.0.0",
	}

	// 使用 Helmify Go 包进行转换
	err := convertUsingHelmifyPackage(resources, options)
	if err != nil {
		// 如果 Helmify Go 包不可用，尝试调用命令行工具
		fmt.Printf("警告: Helmify Go 包不可用，尝试使用命令行工具: %v\n", err)
		err = RunHelmify(resources, options)
		if err != nil {
			// 如果命令行工具也不可用，使用模拟实现
			fmt.Printf("警告: Helmify 工具不可用，使用模拟实现: %v\n", err)
			err = createHelmChartFallback(resources, options)
			if err != nil {
				return fmt.Errorf("创建 Helm Chart 失败: %v", err)
			}
		}
	}

	fmt.Printf("✓ 成功转换为 Helm Chart，输出目录: %s\n", outputDir)
	return nil
}

// ConvertToHelmChartWithOptions 使用指定选项将 Kubernetes 资源转换为 Helm Chart
func ConvertToHelmChartWithOptions(resources []kompose.KubernetesResource, options HelmifyOptions) error {
	fmt.Printf("开始转换 Kubernetes 资源为 Helm Chart（使用自定义选项），输出目录: %s\n", options.OutputDir)

	// 使用 Helmify Go 包进行转换
	err := convertUsingHelmifyPackage(resources, options)
	if err != nil {
		// 如果 Helmify Go 包不可用，尝试调用命令行工具
		fmt.Printf("警告: Helmify Go 包不可用，尝试使用命令行工具: %v\n", err)
		err = RunHelmify(resources, options)
		if err != nil {
			// 如果命令行工具也不可用，使用模拟实现
			fmt.Printf("警告: Helmify 工具不可用，使用模拟实现: %v\n", err)
			err = createHelmChartFallback(resources, options)
			if err != nil {
				return fmt.Errorf("创建 Helm Chart 失败: %v", err)
			}
		}
	}

	fmt.Printf("✓ 成功转换为 Helm Chart，输出目录: %s\n", options.OutputDir)
	return nil
}

// IsHelmifyInstalled 检查 helmify 工具是否已安装
func IsHelmifyInstalled() bool {
	_, err := exec.LookPath("helmify")
	return err == nil
}

// convertUsingHelmifyPackage 使用 helmify Go 包进行转换
func convertUsingHelmifyPackage(resources []kompose.KubernetesResource, options HelmifyOptions) error {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "helmforge-helmify-")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 将 Kubernetes 资源写入临时文件
	k8sDir := filepath.Join(tempDir, "k8s")
	if err := os.MkdirAll(k8sDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}

	for i, resource := range resources {
		fileName := fmt.Sprintf("%d-%s.yaml", i, strings.ToLower(resource.Kind))
		if metadata, ok := resource.Metadata["name"]; ok {
			if name, ok := metadata.(string); ok {
				fileName = fmt.Sprintf("%s-%s.yaml", name, strings.ToLower(resource.Kind))
			}
		}

		filePath := filepath.Join(k8sDir, fileName)
		content, err := yaml.Marshal(resource)
		if err != nil {
			return fmt.Errorf("序列化 Kubernetes 资源失败: %v", err)
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("写入临时文件失败: %v", err)
		}
	}

	// 调用 helmify 包处理
	helmifyConfig := helmifyconfig.Config{
		ChartName:        options.ChartName,
		ChartDir:         options.OutputDir,
		Files:            []string{k8sDir},
		FilesRecursively: true,
		OriginalName:     true,
		PreserveNs:       options.Namespace != "",
		AddWebhookOption: false,
		GenerateDefaults: true,
	}

	err = helmifyapp.Start(nil, helmifyConfig)
	if err != nil {
		return fmt.Errorf("helmify 包转换失败: %v", err)
	}

	// 更新 Chart.yaml 中的版本信息
	chartFile := filepath.Join(options.OutputDir, options.ChartName, "Chart.yaml")
	if _, err := os.Stat(chartFile); err == nil {
		content, err := os.ReadFile(chartFile)
		if err == nil {
			var chartConfig map[string]interface{}
			if err := yaml.Unmarshal(content, &chartConfig); err == nil {
				if options.ChartVersion != "" {
					chartConfig["version"] = options.ChartVersion
				}
				if options.AppVersion != "" {
					chartConfig["appVersion"] = options.AppVersion
				}
				if options.Description != "" {
					chartConfig["description"] = options.Description
				}

				updatedContent, err := yaml.Marshal(chartConfig)
				if err != nil {
					fmt.Printf("警告: 序列化Chart配置失败: %v\n", err)
				} else if err := os.WriteFile(chartFile, updatedContent, 0644); err != nil {
					fmt.Printf("警告: 写入Chart配置文件失败: %v\n", err)
				}
			}
		}
	}

	fmt.Printf("✓ helmify 包转换成功，输出目录: %s\n", options.OutputDir)
	return nil
}

// RunHelmify 运行 helmify 命令
func RunHelmify(resources []kompose.KubernetesResource, options HelmifyOptions) error {
	fmt.Println("执行 helmify 命令...")

	// 检查 helmify 是否已安装
	if !IsHelmifyInstalled() {
		return fmt.Errorf("helmify 工具未安装，请先安装 helmify (https://github.com/mumoshu/helmify)")
	}

	// 为转换创建临时目录
	tempDir, err := os.MkdirTemp("", "helmforge-helmify-")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 将 Kubernetes 资源写入临时文件
	k8sDir := filepath.Join(tempDir, "k8s")
	if err := os.MkdirAll(k8sDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}

	for i, resource := range resources {
		fileName := fmt.Sprintf("%d-%s.yaml", i, strings.ToLower(resource.Kind))
		if metadata, ok := resource.Metadata["name"]; ok {
			if name, ok := metadata.(string); ok {
				fileName = fmt.Sprintf("%s-%s.yaml", name, strings.ToLower(resource.Kind))
			}
		}

		filePath := filepath.Join(k8sDir, fileName)
		content, err := yaml.Marshal(resource)
		if err != nil {
			return fmt.Errorf("序列化 Kubernetes 资源失败: %v", err)
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("写入临时文件失败: %v", err)
		}
	}

	// 使用 helm template 创建 Chart
	releaseName := "helmforge"
	if options.Release != "" {
		releaseName = options.Release
	}

	args := []string{
		"template", releaseName,
		"--namespace", options.Namespace,
		"--output-dir", options.OutputDir,
		k8sDir,
	}

	// 执行 helm template 命令
	cmd := exec.Command("helm", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("helm template 命令执行失败: %v\nstderr: %s", err, stderr.String())
	}

	// 更新 Chart.yaml 中的版本信息
	chartFile := filepath.Join(options.OutputDir, options.ChartName, "Chart.yaml")
	if _, err := os.Stat(chartFile); err == nil {
		content, err := os.ReadFile(chartFile)
		if err == nil {
			var chartConfig map[string]interface{}
			if err := yaml.Unmarshal(content, &chartConfig); err == nil {
				if options.ChartVersion != "" {
					chartConfig["version"] = options.ChartVersion
				}
				if options.AppVersion != "" {
					chartConfig["appVersion"] = options.AppVersion
				}
				if options.Description != "" {
					chartConfig["description"] = options.Description
				}

				updatedContent, err := yaml.Marshal(chartConfig)
				if err != nil {
					fmt.Printf("警告: 序列化Chart配置失败: %v\n", err)
				} else if err := os.WriteFile(chartFile, updatedContent, 0644); err != nil {
					fmt.Printf("警告: 写入Chart配置文件失败: %v\n", err)
				}
			}
		}
	}

	fmt.Printf("✓ Helm template 转换成功，输出目录: %s\n", options.OutputDir)
	return nil
}

// createHelmChartFallback 当 helmify 不可用时使用的回退实现
func createHelmChartFallback(resources []kompose.KubernetesResource, options HelmifyOptions) error {
	// 创建输出目录结构
	if err := createChartDirectoryStructure(options.OutputDir, options); err != nil {
		return fmt.Errorf("创建 Chart 目录结构失败: %v", err)
	}

	// 创建 Chart.yaml 文件
	if err := createChartYaml(options.OutputDir, options); err != nil {
		return fmt.Errorf("创建 Chart.yaml 失败: %v", err)
	}

	// 创建 values.yaml 文件
	if err := createValuesYaml(options.OutputDir); err != nil {
		return fmt.Errorf("创建 values.yaml 失败: %v", err)
	}

	// 创建 templates 目录下的资源文件
	if err := createTemplateFiles(options.OutputDir, resources); err != nil {
		return fmt.Errorf("创建模板文件失败: %v", err)
	}

	// 创建 .helmignore 文件
	if err := createHelmignore(options.OutputDir); err != nil {
		return fmt.Errorf("创建 .helmignore 失败: %v", err)
	}

	return nil
}

// createChartDirectoryStructure 创建 Chart 目录结构
func createChartDirectoryStructure(outputDir string, options HelmifyOptions) error {
	// 创建主目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// 创建 templates 目录
	templatesDir := filepath.Join(outputDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	// 创建 charts 目录
	chartsDir := filepath.Join(outputDir, "charts")
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		return err
	}

	// 创建 templates/_helpers.tpl 文件
	helpersFile := filepath.Join(templatesDir, "_helpers.tpl")
	helpersContent := `{{/*
Chart name
*/}}
{{- define "helmforge.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Chart full name
*/}}
{{- define "helmforge.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Container image
*/}}
{{- define "helmforge.image" -}}
{{- $registry := .Values.image.registry }}
{{- $repository := .Values.image.repository }}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Service port
*/}}
{{- define "helmforge.service.port" -}}
{{- .Values.service.port | default 8080 }}
{{- end }}
`
	if err := os.WriteFile(helpersFile, []byte(helpersContent), 0644); err != nil {
		return err
	}

	return nil
}

// createChartYaml 创建 Chart.yaml 文件
func createChartYaml(outputDir string, options HelmifyOptions) error {
	chartYaml := map[string]interface{}{
		"apiVersion":  "v2",
		"name":        options.ChartName,
		"version":     options.ChartVersion,
		"appVersion":  options.AppVersion,
		"description": options.Description,
		"type":        "application",
		"keywords": []string{
			"helmforge",
			"app",
		},
		"sources": []string{
			"https://github.com/wenwenxiong/HelmForge",
		},
		"maintainers": []map[string]string{
			{
				"name":  "HelmForge Team",
				"email": "team@helmforge.example.com",
			},
		},
		"annotations": map[string]string{
			"helm.sh/created": "HelmForge",
		},
	}

	content, err := yaml.Marshal(chartYaml)
	if err != nil {
		return err
	}

	chartFile := filepath.Join(outputDir, "Chart.yaml")
	return os.WriteFile(chartFile, content, 0644)
}

// createValuesYaml 创建 values.yaml 文件
func createValuesYaml(outputDir string) error {
	valuesYaml := map[string]interface{}{
		"nameOverride":     "",
		"fullnameOverride": "",
		"image": map[string]interface{}{
			"registry":   "",
			"repository": "app",
			"tag":        "",
			"pullPolicy": "IfNotPresent",
		},
		"service": map[string]interface{}{
			"type": "ClusterIP",
			"port": 8080,
		},
		"replicaCount": 1,
		"resources": map[string]interface{}{
			"limits": map[string]string{
				"cpu":    "100m",
				"memory": "128Mi",
			},
			"requests": map[string]string{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
		"env":          map[string]string{},
		"volumes":      []map[string]interface{}{},
		"nodeSelector": map[string]string{},
		"tolerations":  []interface{}{},
		"affinity":     map[string]interface{}{},
	}

	content, err := yaml.Marshal(valuesYaml)
	if err != nil {
		return err
	}

	valuesFile := filepath.Join(outputDir, "values.yaml")
	return os.WriteFile(valuesFile, content, 0644)
}

// createTemplateFiles 创建模板文件
func createTemplateFiles(outputDir string, resources []kompose.KubernetesResource) error {
	templatesDir := filepath.Join(outputDir, "templates")

	for i, resource := range resources {
		// 为资源文件生成唯一名称
		fileName := fmt.Sprintf("%d-%s.yaml", i, strings.ToLower(resource.Kind))
		if metadata, ok := resource.Metadata["name"]; ok {
			if name, ok := metadata.(string); ok {
				fileName = fmt.Sprintf("%s-%s.yaml", name, strings.ToLower(resource.Kind))
			}
		}

		// 将资源转换为模板，替换硬编码的值为 Helm 模板变量
		templatizedResource := templatizeResource(resource)

		// 序列化资源
		content, err := yaml.Marshal(templatizedResource)
		if err != nil {
			return fmt.Errorf("序列化资源失败: %v", err)
		}

		// 写入文件
		filePath := filepath.Join(templatesDir, fileName)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("写入模板文件失败: %v", err)
		}

		fmt.Printf("✓ 生成模板文件: %s\n", fileName)
	}

	return nil
}

// templatizeResource 将资源转换为 Helm 模板
func templatizeResource(resource kompose.KubernetesResource) kompose.KubernetesResource {
	// 替换 metadata.name
	if resource.Metadata != nil {
		if _, ok := resource.Metadata["name"]; ok {
			// 对于 Deployment 和 Service，使用模板函数生成名称
			if resource.Kind == "Deployment" || resource.Kind == "Service" {
				resource.Metadata["name"] = "{{ include \"helmforge.fullname\" . }}"
			}
		}

		// 替换 labels
		if labels, ok := resource.Metadata["labels"].(map[string]interface{}); ok {
			labels["app.kubernetes.io/name"] = "{{ include \"helmforge.name\" . }}"
			labels["app.kubernetes.io/instance"] = "{{ .Release.Name }}"
			labels["app.kubernetes.io/version"] = "{{ .Chart.AppVersion }}"
			labels["app.kubernetes.io/managed-by"] = "{{ .Release.Service }}"
			resource.Metadata["labels"] = labels
		}
	}

	// 替换 spec 中的值
	if resource.Spec != nil {
		// 对于 Deployment
		if resource.Kind == "Deployment" {
			if template, ok := resource.Spec["template"].(map[string]interface{}); ok {
				if spec, ok := template["spec"].(map[string]interface{}); ok {
					if containers, ok := spec["containers"].([]interface{}); ok {
						for _, container := range containers {
							if c, ok := container.(map[string]interface{}); ok {
								// 替换 image
								if _, ok := c["image"]; ok {
									c["image"] = "{{ include \"helmforge.image\" . }}"
								}

								// 替换 resources
								if _, ok := c["resources"]; !ok {
									c["resources"] = map[string]interface{}{
										"limits":   "{{ toJson .Values.resources.limits }}",
										"requests": "{{ toJson .Values.resources.requests }}",
									}
								}
							}
						}
					}
				}
			}
		}

		// 对于 Service
		if resource.Kind == "Service" {
			if ports, ok := resource.Spec["ports"].([]interface{}); ok {
				for _, port := range ports {
					if p, ok := port.(map[string]interface{}); ok {
						if _, ok := p["port"]; ok {
							p["port"] = "{{ .Values.service.port }}"
						}
					}
				}
			}

			if _, ok := resource.Spec["type"]; ok {
				resource.Spec["type"] = "{{ .Values.service.type }}"
			}
		}
	}

	return resource
}

// createHelmignore 创建 .helmignore 文件
func createHelmignore(outputDir string) error {
	helmignoreContent := `# Dependency directories
node_modules/
jspm_packages/

# Build outputs
build/
dist/

# IDE and editor files
.idea/
.vscode/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# HelmForge generated files
helmforge-output/
`

	helmignoreFile := filepath.Join(outputDir, ".helmignore")
	return os.WriteFile(helmignoreFile, []byte(helmignoreContent), 0644)
}
