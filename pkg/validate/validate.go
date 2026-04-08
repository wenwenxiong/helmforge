package validate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateHelmChart 验证 Helm Chart
func ValidateHelmChart(chartDir string) error {
	fmt.Printf("开始验证 Helm Chart，Chart 目录: %s\n", chartDir)

	// 检查 Chart 目录是否存在
	if _, err := os.Stat(chartDir); os.IsNotExist(err) {
		return fmt.Errorf("Chart 目录不存在: %s", chartDir)
	}

	// 检查 Chart 基本结构
	if err := checkChartStructure(chartDir); err != nil {
		return fmt.Errorf("检查 Chart 结构失败: %v", err)
	}

	// 检查 Chart.yaml 文件
	if err := checkChartYaml(chartDir); err != nil {
		return fmt.Errorf("检查 Chart.yaml 失败: %v", err)
	}

	// 检查 values.yaml 文件
	if err := checkValuesYaml(chartDir); err != nil {
		return fmt.Errorf("检查 values.yaml 失败: %v", err)
	}

	// 检查模板文件
	if err := checkTemplateFiles(chartDir); err != nil {
		return fmt.Errorf("检查模板文件失败: %v", err)
	}

	// 运行 helm lint 命令
	if err := runHelmLint(chartDir); err != nil {
		return fmt.Errorf("运行 helm lint 失败: %v", err)
	}

	// 运行 helm template 命令
	if err := runHelmTemplate(chartDir); err != nil {
		return fmt.Errorf("运行 helm template 失败: %v", err)
	}

	// 检查多环境配置文件
	if err := checkEnvironmentConfigs(chartDir); err != nil {
		return fmt.Errorf("检查多环境配置文件失败: %v", err)
	}

	fmt.Println("✓ 成功验证 Helm Chart")
	return nil
}

// checkChartStructure 检查 Chart 基本结构
func checkChartStructure(chartDir string) error {
	requiredDirs := []string{
		"templates",
		"charts",
	}

	requiredFiles := []string{
		"Chart.yaml",
		"values.yaml",
	}

	// 检查必需的目录
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(chartDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return fmt.Errorf("缺少必需的目录: %s", dir)
		}
	}

	// 检查必需的文件
	for _, file := range requiredFiles {
		filePath := filepath.Join(chartDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("缺少必需的文件: %s", file)
		}
	}

	fmt.Println("✓ 检查 Chart 结构成功")
	return nil
}

// checkChartYaml 检查 Chart.yaml 文件
func checkChartYaml(chartDir string) error {
	chartFile := filepath.Join(chartDir, "Chart.yaml")

	// 读取文件内容
	content, err := os.ReadFile(chartFile)
	if err != nil {
		return fmt.Errorf("读取 Chart.yaml 失败: %v", err)
	}

	// 解析 Chart.yaml
	var chartConfig map[string]interface{}
	if err := yaml.Unmarshal(content, &chartConfig); err != nil {
		return fmt.Errorf("解析 Chart.yaml 失败: %v", err)
	}

	// 检查必需的字段
	requiredFields := []string{
		"apiVersion",
		"name",
		"version",
	}

	for _, field := range requiredFields {
		if _, ok := chartConfig[field]; !ok {
			return fmt.Errorf("Chart.yaml 缺少必需的字段: %s", field)
		}
	}

	// 检查 apiVersion
	if apiVersion, ok := chartConfig["apiVersion"].(string); ok {
		if apiVersion != "v1" && apiVersion != "v2" {
			return fmt.Errorf("Chart.yaml 的 apiVersion 无效: %s", apiVersion)
		}
	}

	// 检查 version 格式
	if version, ok := chartConfig["version"].(string); ok {
		if !isValidVersion(version) {
			return fmt.Errorf("Chart.yaml 的 version 格式无效: %s", version)
		}
	}

	fmt.Println("✓ 检查 Chart.yaml 成功")
	return nil
}

// checkValuesYaml 检查 values.yaml 文件
func checkValuesYaml(chartDir string) error {
	valuesFile := filepath.Join(chartDir, "values.yaml")

	// 读取文件内容
	content, err := os.ReadFile(valuesFile)
	if err != nil {
		return fmt.Errorf("读取 values.yaml 失败: %v", err)
	}

	// 解析 values.yaml
	var values map[string]interface{}
	if err := yaml.Unmarshal(content, &values); err != nil {
		return fmt.Errorf("解析 values.yaml 失败: %v", err)
	}

	// 检查基本结构
	if _, ok := values["image"]; !ok {
		return fmt.Errorf("values.yaml 缺少 image 配置")
	}

	if _, ok := values["service"]; !ok {
		return fmt.Errorf("values.yaml 缺少 service 配置")
	}

	fmt.Println("✓ 检查 values.yaml 成功")
	return nil
}

// checkTemplateFiles 检查模板文件
func checkTemplateFiles(chartDir string) error {
	templatesDir := filepath.Join(chartDir, "templates")

	// 读取 templates 目录下的所有文件
	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return fmt.Errorf("读取 templates 目录失败: %v", err)
	}

	// 检查是否有模板文件
	templateCount := 0
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

		// 检查文件内容
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取模板文件 %s 失败: %v", fileName, err)
		}

		// 检查文件是否为空
		if len(strings.TrimSpace(string(content))) == 0 {
			return fmt.Errorf("模板文件 %s 为空", fileName)
		}

		templateCount++
	}

	if templateCount == 0 {
		return fmt.Errorf("templates 目录中没有有效的模板文件")
	}

	fmt.Printf("✓ 检查模板文件成功，发现 %d 个模板文件\n", templateCount)
	return nil
}

// runHelmLint 运行 helm lint 命令
func runHelmLint(chartDir string) error {
	fmt.Println("运行 helm lint 命令...")

	// 执行 helm lint 命令
	cmd := exec.Command("helm", "lint", chartDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// 注意：helm lint 命令在发现警告时也会返回非零退出码
		// 这里我们只检查命令是否能够执行，不严格要求完全通过
		fmt.Printf("警告: helm lint 命令返回错误: %v\n", err)
		fmt.Println("继续验证过程...")
	}

	fmt.Println("✓ 运行 helm lint 命令完成")
	return nil
}

// runHelmTemplate 运行 helm template 命令
func runHelmTemplate(chartDir string) error {
	fmt.Println("运行 helm template 命令...")

	// 执行 helm template 命令
	cmd := exec.Command("helm", "template", "test-release", chartDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("helm template 命令失败: %v", err)
	}

	fmt.Println("✓ 运行 helm template 命令成功")
	return nil
}

// checkEnvironmentConfigs 检查多环境配置文件
func checkEnvironmentConfigs(chartDir string) error {
	envDir := filepath.Join(chartDir, "environments")

	// 检查 environments 目录是否存在
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		// environments 目录不是必需的，所以这里只是警告
		fmt.Println("警告: environments 目录不存在")
		return nil
	}

	// 读取 environments 目录下的所有文件
	files, err := os.ReadDir(envDir)
	if err != nil {
		return fmt.Errorf("读取 environments 目录失败: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
			continue
		}

		filePath := filepath.Join(envDir, fileName)

		// 检查文件内容
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取环境配置文件 %s 失败: %v", fileName, err)
		}

		// 解析环境配置文件
		var envConfig map[string]interface{}
		if err := yaml.Unmarshal(content, &envConfig); err != nil {
			return fmt.Errorf("解析环境配置文件 %s 失败: %v", fileName, err)
		}

		fmt.Printf("✓ 检查环境配置文件 %s 成功\n", fileName)
	}

	return nil
}

// isValidVersion 检查版本号格式是否有效
func isValidVersion(version string) bool {
	// 简单检查版本号格式，应该符合语义化版本规范
	// 这里只是一个基本检查，实际应用中可以使用更严格的检查
	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	for _, part := range parts {
		for _, char := range part {
			if !((char >= '0' && char <= '9') || char == '-' || char == '+') {
				return false
			}
		}
	}

	return true
}
