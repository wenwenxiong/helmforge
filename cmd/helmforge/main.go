package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wenwenxiong/HelmForge/pkg/config"
	"github.com/wenwenxiong/HelmForge/pkg/helmify"
	"github.com/wenwenxiong/HelmForge/pkg/kompose"
	"github.com/wenwenxiong/HelmForge/pkg/parser"
	"github.com/wenwenxiong/HelmForge/pkg/validate"
)

var (
	inputFile string
	outputDir string
)

var rootCmd = &cobra.Command{
	Use:   "helmforge",
	Short: "HelmForge - 将 Docker Compose 文件转换为 Helm Chart 的工具",
	Long:  `HelmForge 是一个基于 kompose 和 helmify 的命令行工具，旨在帮助用户将 Docker Compose 文件优雅合规地转换为 Helm Chart，简化 Kubernetes 部署流程。`,
}

var composeToChartCmd = &cobra.Command{
	Use:   "compose-to-chart",
	Short: "从 Docker Compose 文件生成 Helm Chart",
	Long:  `将 Docker Compose 文件转换为可部署的 Helm Chart`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== HelmForge: Docker Compose 转 Helm Chart ===")
		fmt.Printf("输入文件: %s\n", inputFile)
		fmt.Printf("输出目录: %s\n", outputDir)

		// 解析 Docker Compose 文件
		composeConfig, err := parser.ParseDockerCompose(inputFile)
		if err != nil {
			fmt.Printf("解析 Docker Compose 文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 解析 Docker Compose 文件成功")

		// 使用 Kompose 转换为 Kubernetes 资源
		k8sResources, err := kompose.ConvertToKubernetes(composeConfig)
		if err != nil {
			fmt.Printf("转换为 Kubernetes 资源失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 转换为 Kubernetes 资源成功")

		// 使用 Helmify 转换为 Helm Chart
		err = helmify.ConvertToHelmChart(k8sResources, outputDir)
		if err != nil {
			fmt.Printf("转换为 Helm Chart 失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 转换为 Helm Chart 成功")

		// 增强 Helm Chart（配置参数化）
		err = config.EnhanceHelmChart(outputDir)
		if err != nil {
			fmt.Printf("增强 Helm Chart 失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 增强 Helm Chart 成功")

		// 验证 Helm Chart
		err = validate.ValidateHelmChart(outputDir)
		if err != nil {
			fmt.Printf("验证 Helm Chart 失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 验证 Helm Chart 成功")

		fmt.Println("\n=== 转换完成 ===")
		fmt.Printf("Helm Chart 已生成到: %s\n", outputDir)
	},
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "./output-chart", "输出 Helm Chart 目录")

	// compose-to-chart 命令
	composeToChartCmd.Flags().StringVarP(&inputFile, "input", "i", "docker-compose.yaml", "Docker Compose 文件路径")
	rootCmd.AddCommand(composeToChartCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
