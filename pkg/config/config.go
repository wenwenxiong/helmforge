package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnhanceHelmChart 增强 Helm Chart 配置
func EnhanceHelmChart(chartDir string) error {
	fmt.Printf("开始增强 Helm Chart 配置，Chart 目录: %s\n", chartDir)

	// 检查 Chart 目录是否存在
	if _, err := os.Stat(chartDir); os.IsNotExist(err) {
		return fmt.Errorf("Chart 目录不存在: %s", chartDir)
	}

	// 增强 values.yaml 文件
	if err := enhanceValuesYaml(chartDir); err != nil {
		return fmt.Errorf("增强 values.yaml 失败: %v", err)
	}

	// 增强模板文件
	if err := enhanceTemplateFiles(chartDir); err != nil {
		return fmt.Errorf("增强模板文件失败: %v", err)
	}

	// 生成多环境配置文件
	if err := generateEnvironmentConfigs(chartDir); err != nil {
		return fmt.Errorf("生成多环境配置文件失败: %v", err)
	}

	// 生成配置文档
	if err := generateConfigDocumentation(chartDir); err != nil {
		return fmt.Errorf("生成配置文档失败: %v", err)
	}

	fmt.Println("✓ 成功增强 Helm Chart 配置")
	return nil
}

// enhanceValuesYaml 增强 values.yaml 文件
func enhanceValuesYaml(chartDir string) error {
	valuesFile := filepath.Join(chartDir, "values.yaml")

	// 读取当前 values.yaml 文件
	content, err := os.ReadFile(valuesFile)
	if err != nil {
		return fmt.Errorf("读取 values.yaml 失败: %v", err)
	}

	// 解析 values.yaml
	var values map[string]interface{}
	if err := yaml.Unmarshal(content, &values); err != nil {
		return fmt.Errorf("解析 values.yaml 失败: %v", err)
	}

	// 增强 values.yaml 结构
	enhancedValues := enhanceValuesStructure(values)

	// 写回增强后的 values.yaml
	enhancedContent, err := yaml.Marshal(enhancedValues)
	if err != nil {
		return fmt.Errorf("序列化增强后的 values.yaml 失败: %v", err)
	}

	if err := os.WriteFile(valuesFile, enhancedContent, 0644); err != nil {
		return fmt.Errorf("写入 values.yaml 失败: %v", err)
	}

	fmt.Println("✓ 增强 values.yaml 文件")
	return nil
}

// enhanceValuesStructure 增强 values 结构
func enhanceValuesStructure(values map[string]interface{}) map[string]interface{} {
	// 确保基本结构存在
	if _, ok := values["image"]; !ok {
		values["image"] = map[string]interface{}{
			"registry":   "",
			"repository": "app",
			"tag":        "",
			"pullPolicy": "IfNotPresent",
		}
	}

	if _, ok := values["service"]; !ok {
		values["service"] = map[string]interface{}{
			"type":        "ClusterIP",
			"port":        8080,
			"annotations": map[string]interface{}{},
		}
	}

	if _, ok := values["replicaCount"]; !ok {
		values["replicaCount"] = 1
	}

	if _, ok := values["resources"]; !ok {
		values["resources"] = map[string]interface{}{
			"limits": map[string]string{
				"cpu":    "100m",
				"memory": "128Mi",
			},
			"requests": map[string]string{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		}
	}

	if _, ok := values["env"]; !ok {
		values["env"] = map[string]string{}
	}

	if _, ok := values["volumes"]; !ok {
		values["volumes"] = []map[string]interface{}{}
	}

	if _, ok := values["ingress"]; !ok {
		values["ingress"] = map[string]interface{}{
			"enabled":     false,
			"className":   "",
			"annotations": map[string]interface{}{},
			"hosts": []map[string]interface{}{
				{
					"host": "chart-example.local",
					"paths": []map[string]interface{}{
						{
							"path":     "/",
							"pathType": "ImplementationSpecific",
						},
					},
				},
			},
			"tls": []interface{}{},
		}
	}

	if _, ok := values["autoscaling"]; !ok {
		values["autoscaling"] = map[string]interface{}{
			"enabled":                           false,
			"minReplicas":                       1,
			"maxReplicas":                       100,
			"targetCPUUtilizationPercentage":    80,
			"targetMemoryUtilizationPercentage": 80,
		}
	}

	if _, ok := values["nodeSelector"]; !ok {
		values["nodeSelector"] = map[string]string{}
	}

	if _, ok := values["tolerations"]; !ok {
		values["tolerations"] = []interface{}{}
	}

	if _, ok := values["affinity"]; !ok {
		values["affinity"] = map[string]interface{}{}
	}

	// 添加配置管理部分
	values["config"] = map[string]interface{}{
		"app": map[string]interface{}{
			"debug":    false,
			"logLevel": "info",
			"timeout":  "30s",
		},
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"name":     "app",
			"username": "app",
			"password": "",
			"sslMode":  "disable",
		},
		"cache": map[string]interface{}{
			"enabled":  false,
			"host":     "localhost",
			"port":     6379,
			"password": "",
			"database": 0,
		},
		"storage": map[string]interface{}{
			"enabled": false,
			"type":    "local",
			"path":    "/data",
		},
	}

	return values
}

// enhanceTemplateFiles 增强模板文件
func enhanceTemplateFiles(chartDir string) error {
	templatesDir := filepath.Join(chartDir, "templates")

	// 检查 templates 目录是否存在
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return fmt.Errorf("templates 目录不存在: %s", templatesDir)
	}

	// 读取 templates 目录下的所有文件
	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return fmt.Errorf("读取 templates 目录失败: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		// 跳过 _helpers.tpl 和 NOTES.txt 文件
		if strings.HasPrefix(fileName, "_") || fileName == "NOTES.txt" {
			continue
		}

		filePath := filepath.Join(templatesDir, fileName)
		if err := enhanceTemplateFile(filePath); err != nil {
			return fmt.Errorf("增强模板文件 %s 失败: %v", fileName, err)
		}
	}

	// 增强或创建 _helpers.tpl 文件
	helpersFile := filepath.Join(templatesDir, "_helpers.tpl")
	if err := enhanceHelpersTpl(helpersFile); err != nil {
		return fmt.Errorf("增强 _helpers.tpl 失败: %v", err)
	}

	// 创建 NOTES.txt 文件
	notesFile := filepath.Join(templatesDir, "NOTES.txt")
	if err := createNotesTxt(notesFile); err != nil {
		return fmt.Errorf("创建 NOTES.txt 失败: %v", err)
	}

	fmt.Println("✓ 增强模板文件")
	return nil
}

// enhanceTemplateFile 增强单个模板文件
func enhanceTemplateFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 增强内容
	enhancedContent := string(content)

	// 添加配置映射
	enhancedContent = addConfigMaps(enhancedContent)

	// 添加环境变量从 values 中获取
	enhancedContent = addEnvFromValues(enhancedContent)

	// 写回文件
	if err := os.WriteFile(filePath, []byte(enhancedContent), 0644); err != nil {
		return err
	}

	return nil
}

// addConfigMaps 添加配置映射
func addConfigMaps(content string) string {
	// 如果内容中包含 ConfigMap 的键名，生成 ConfigMap 模板
	if strings.Contains(content, "CONFIG_FROM_CONFIGMAP") {
		// 生成 ConfigMap 模板
		configMapTemplate := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "helmforge.configmap.name" . }}
  labels:
    {{- include "helmforge.labels" . | nindent 4 }}
data:
  {{- if .Values.config.app }}
  app-config.yaml: |
    debug: {{ .Values.config.app.debug | default false }}
    logLevel: {{ .Values.config.app.logLevel | default "info" }}
    timeout: {{ .Values.config.app.timeout | default "30s" }}
  {{- end }}
  {{- if .Values.config.database }}
  database-config.yaml: |
    host: {{ .Values.config.database.host | default "localhost" }}
    port: {{ .Values.config.database.port | default 5432 }}
    name: {{ .Values.config.database.name | default "app" }}
    username: {{ .Values.config.database.username | default "app" }}
    sslMode: {{ .Values.config.database.sslMode | default "disable" }}
  {{- end }}
  {{- if .Values.config.cache }}
  cache-config.yaml: |
    enabled: {{ .Values.config.cache.enabled | default false }}
    host: {{ .Values.config.cache.host | default "localhost" }}
    port: {{ .Values.config.cache.port | default 6379 }}
    database: {{ .Values.config.cache.database | default 0 }}
  {{- end }}
`

		// 将 ConfigMap 模板添加到文件末尾
		content += configMapTemplate
	}

	return content
}

// addEnvFromValues 添加环境变量从 values 中获取
func addEnvFromValues(content string) string {
	// 检查是否已经有环境变量配置
	if !strings.Contains(content, "env:") && !strings.Contains(content, "valueFrom:") {
		return content
	}

	// 在 env 部分添加模板化支持
	// 将硬编码的环境变量值替换为 values 引用
	content = strings.ReplaceAll(content, "value: \"{{", "value: \"{{ .Values.env.")
	content = strings.ReplaceAll(content, "value: '", "value: '{{ .Values.env.")

	// 添加环境变量从 ConfigMap 或 Secret 获取的示例
	if strings.Contains(content, "env:") && !strings.Contains(content, "valueFrom:") {
		envSection := `
        env:
        - name: CONFIG_FROM_VALUES
          value: "{{ .Values.env.appConfig | default \"default\" }}"
        - name: SECRET_FROM_SECRET
          valueFrom:
            secretKeyRef:
              name: "{{ include \"helmforge.secret.name\" . }}"
              key: secret-key
        - name: CONFIG_FROM_CONFIGMAP
          valueFrom:
            configMapKeyRef:
               name: "{{ include \"helmforge.configmap.name\" . }}"
               key: config-key
`

		// 在 env: 之后插入增强的环境变量配置
		content = strings.Replace(content, "env:", envSection+"\n        env:", 1)
	}

	return content
}

// enhanceHelpersTpl 增强 _helpers.tpl 文件
func enhanceHelpersTpl(filePath string) error {
	// 检查文件是否存在
	var content string
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，读取内容
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		content = string(fileContent)
	} else {
		// 文件不存在，创建新内容
		content = `{{/*
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
	}

	// 添加额外的 helper 函数
	additionalHelpers := `
{{/*
Config map name
*/}}
{{- define "helmforge.configmap.name" -}}
{{- printf "%s-config" (include "helmforge.fullname" .) }}
{{- end }}

{{/*
Secret name
*/}}
{{- define "helmforge.secret.name" -}}
{{- printf "%s-secret" (include "helmforge.fullname" .) }}
{{- end }}

{{/*
Database host
*/}}
{{- define "helmforge.database.host" -}}
{{- .Values.config.database.host | default "localhost" }}
{{- end }}

{{/*
Database port
*/}}
{{- define "helmforge.database.port" -}}
{{- .Values.config.database.port | default 5432 }}
{{- end }}
`

	// 检查是否已经包含额外的 helper 函数
	if !strings.Contains(content, "helmforge.configmap.name") {
		content += additionalHelpers
	}

	// 写回文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}

// createNotesTxt 创建 NOTES.txt 文件
func createNotesTxt(filePath string) error {
	notesContent := `{{- /* vim: set filetype=mustache: */ -}}
{{- /*
Copyright 2024 The HelmForge Authors. All rights reserved.
*/ -}}

{{- if .Values.ingress.enabled }}
1. Get the application URL by running these commands:
{{- range .Values.ingress.hosts }}
  {{- range .paths }}
  * {{- if $.Values.ingress.tls }}
    https://{{ .host }}{{ .path }}
  {{- else }}
    http://{{ .host }}{{ .path }}
  {{- end }}
  {{- end }}
{{- end }}
{{- else if contains "NodePort" .Values.service.type }}
1. Get the application URL by running these commands:
  export NODE_PORT=$(kubectl get --namespace {{ .Release.Namespace }} -o jsonpath="{.spec.ports[0].nodePort}" services {{ include "helmforge.fullname" . }})
  export NODE_IP=$(kubectl get nodes --namespace {{ .Release.Namespace }} -o jsonpath="{.items[0].status.addresses[0].address}")
  echo http://$NODE_IP:$NODE_PORT
{{- else if contains "LoadBalancer" .Values.service.type }}
1. Get the application URL by running these commands:
  NOTE: It may take a few minutes for the LoadBalancer IP to be available.
        You can watch the status of by running 'kubectl get svc -w {{ include "helmforge.fullname" . }}'
  export SERVICE_IP=$(kubectl get svc --namespace {{ .Release.Namespace }} {{ include "helmforge.fullname" . }} -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo http://$SERVICE_IP:{{ .Values.service.port }}
{{- else if contains "ClusterIP" .Values.service.type }}
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace {{ .Release.Namespace }} -l "app.kubernetes.io/name={{ include "helmforge.name" . }},app.kubernetes.io/instance={{ .Release.Name }}" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace {{ .Release.Namespace }} $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  kubectl --namespace {{ .Release.Namespace }} port-forward $POD_NAME 8080:$CONTAINER_PORT
  echo http://127.0.0.1:8080
{{- end }}

2. Configuration

The following table lists the configurable parameters of the {{ include "helmforge.name" . }} chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| image.repository | Image repository | app |
| image.tag | Image tag | {{ .Chart.AppVersion }} |
| image.pullPolicy | Image pull policy | IfNotPresent |
| service.type | Service type | ClusterIP |
| service.port | Service port | 8080 |
| replicaCount | Number of replicas | 1 |
| resources | Resource requests and limits | See values.yaml |
| config.app.debug | Enable debug mode | false |
| config.database.host | Database host | localhost |
| config.database.port | Database port | 5432 |

For more information, see the configuration documentation in the chart directory.
`

	// 写回文件
	if err := os.WriteFile(filePath, []byte(notesContent), 0644); err != nil {
		return err
	}

	return nil
}

// generateEnvironmentConfigs 生成多环境配置文件
func generateEnvironmentConfigs(chartDir string) error {
	envDir := filepath.Join(chartDir, "environments")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return err
	}

	// 生成开发环境配置
	devConfig := map[string]interface{}{
		"replicaCount": 1,
		"image": map[string]interface{}{
			"tag": "dev",
		},
		"config": map[string]interface{}{
			"app": map[string]interface{}{
				"debug":    true,
				"logLevel": "debug",
			},
			"database": map[string]interface{}{
				"host": "dev-db",
				"name": "app_dev",
			},
		},
	}

	devFile := filepath.Join(envDir, "development.yaml")
	if err := writeYamlFile(devFile, devConfig); err != nil {
		return err
	}

	// 生成测试环境配置
	testConfig := map[string]interface{}{
		"replicaCount": 2,
		"image": map[string]interface{}{
			"tag": "test",
		},
		"config": map[string]interface{}{
			"app": map[string]interface{}{
				"debug":    false,
				"logLevel": "info",
			},
			"database": map[string]interface{}{
				"host": "test-db",
				"name": "app_test",
			},
		},
	}

	testFile := filepath.Join(envDir, "testing.yaml")
	if err := writeYamlFile(testFile, testConfig); err != nil {
		return err
	}

	// 生成生产环境配置
	prodConfig := map[string]interface{}{
		"replicaCount": 3,
		"image": map[string]interface{}{
			"tag": "prod",
		},
		"resources": map[string]interface{}{
			"limits": map[string]string{
				"cpu":    "500m",
				"memory": "512Mi",
			},
			"requests": map[string]string{
				"cpu":    "200m",
				"memory": "256Mi",
			},
		},
		"config": map[string]interface{}{
			"app": map[string]interface{}{
				"debug":    false,
				"logLevel": "warn",
			},
			"database": map[string]interface{}{
				"host": "prod-db",
				"name": "app_prod",
			},
		},
	}

	prodFile := filepath.Join(envDir, "production.yaml")
	if err := writeYamlFile(prodFile, prodConfig); err != nil {
		return err
	}

	fmt.Println("✓ 生成多环境配置文件")
	return nil
}

// generateConfigDocumentation 生成配置文档
func generateConfigDocumentation(chartDir string) error {
	docsDir := filepath.Join(chartDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}

	configDoc := "# 配置文档\n\n## 概述\n\n本文档描述了 Helm Chart 的配置选项。\n\n## 配置结构\n\nChart 的配置分为以下几个主要部分：\n\n- **image**: 容器镜像配置\n- **service**: 服务配置\n- **replicaCount**: 副本数量\n- **resources**: 资源限制和请求\n- **env**: 环境变量\n- **config**: 应用配置\n- **ingress**:  ingress 配置\n- **autoscaling**: 自动伸缩配置\n- **nodeSelector**: 节点选择器\n- **tolerations**: 容忍度\n- **affinity**: 亲和性\n\n## 详细配置\n\n### image\n\n| 参数 | 描述 | 默认值 |\n|------|------|--------|\n| image.registry | 镜像仓库 | \"\" |\n| image.repository | 镜像名称 | \"app\" |\n| image.tag | 镜像标签 | \"\" |\n| image.pullPolicy | 镜像拉取策略 | \"IfNotPresent\" |\n\n### service\n\n| 参数 | 描述 | 默认值 |\n|------|------|--------|\n| service.type | 服务类型 | \"ClusterIP\" |\n| service.port | 服务端口 | 8080 |\n| service.annotations | 服务注解 | {} |\n\n### config\n\n| 参数 | 描述 | 默认值 |\n|------|------|--------|\n| config.app.debug | 启用调试模式 | false |\n| config.app.logLevel | 日志级别 | \"info\" |\n| config.app.timeout | 超时时间 | \"30s\" |\n| config.database.host | 数据库主机 | \"localhost\" |\n| config.database.port | 数据库端口 | 5432 |\n| config.database.name | 数据库名称 | \"app\" |\n| config.database.username | 数据库用户名 | \"app\" |\n| config.database.password | 数据库密码 | \"\" |\n| config.database.sslMode | SSL 模式 | \"disable\" |\n| config.cache.enabled | 启用缓存 | false |\n| config.cache.host | 缓存主机 | \"localhost\" |\n| config.cache.port | 缓存端口 | 6379 |\n| config.storage.enabled | 启用存储 | false |\n| config.storage.type | 存储类型 | \"local\" |\n| config.storage.path | 存储路径 | \"/data\" |\n\n## 多环境配置\n\nChart 提供了以下环境的配置文件：\n\n- **development.yaml**: 开发环境配置\n- **testing.yaml**: 测试环境配置\n- **production.yaml**: 生产环境配置\n\n使用方法：\n\n```bash\n# 使用开发环境配置\nhelm install my-app . -f environments/development.yaml\n\n# 使用测试环境配置\nhelm install my-app . -f environments/testing.yaml\n\n# 使用生产环境配置\nhelm install my-app . -f environments/production.yaml\n```\n\n## 自定义配置\n\n您可以通过创建自定义的 values 文件来覆盖默认配置：\n\n```yaml\n# custom-values.yaml\nreplicaCount: 5\n\nimage:\n  tag: v1.2.3\n\nconfig:\n  app:\n    logLevel: error\n  database:\n    host: my-db\n```\n\n然后使用该文件安装 Chart：\n\n```bash\nhelm install my-app . -f custom-values.yaml\n```\n"

	configDocFile := filepath.Join(docsDir, "configuration.md")
	if err := os.WriteFile(configDocFile, []byte(configDoc), 0644); err != nil {
		return err
	}

	fmt.Println("✓ 生成配置文档")
	return nil
}

// writeYamlFile 写入 YAML 文件
func writeYamlFile(filePath string, data interface{}) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return err
	}

	return nil
}
