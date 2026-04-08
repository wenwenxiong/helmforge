# HelmForge

## 项目简介
HelmForge 是一个基于 kompose 和 helmify 的命令行工具，专注于将 Docker Compose 文件优雅合规地转换为 Helm Chart，简化 Kubernetes 部署流程。通过集成和增强 kompose 和 helmify 的能力，提供更强大、更易用的转换功能。

## 需求描述
目前有不少用户需要将现有的 Docker Compose 配置转换为 Helm Chart，以便在 Kubernetes 环境中部署应用。主要痛点包括：

- 不熟悉 Kubernetes 部署规范
- 手动编写 Helm Chart 复杂且容易出错
- 希望有一个简单易用的工具来自动化转换过程
- 需要生成符合最佳实践的 Helm Chart

## 核心工作流程
```
Docker Compose 文件
    ↓
Kompose 转换（Docker Compose → Kubernetes 资源）
    ↓
Helmify 转换（Kubernetes 资源 → Helm Chart）
    ↓
HelmForge 增强（配置参数化、模板优化）
    ↓
验证与测试
    ↓
Kubernetes 部署
```

## 功能模块

### Docker Compose 转 Helm Chart（基于 Kompose + Helmify）
- 集成 Kompose 进行 Docker Compose 到 Kubernetes 资源的转换
- 集成 Helmify 进行 Kubernetes 资源到 Helm Chart 的转换
- 解析 docker-compose.yaml 文件，提取服务定义、网络、卷等配置
- 增强对 Docker Compose 高级特性的支持
- 提供配置参数化和模板优化
- 集成 Helm Chart 验证功能

## 技术架构

### 核心依赖工具
- **Kompose**: Docker Compose 到 Kubernetes 资源的转换工具
  - GitHub: https://github.com/kubernetes/kompose
  - 用于将 docker-compose.yaml 转换为 Deployment、Service 等 Kubernetes 资源
- **Helmify**: Kubernetes 资源到 Helm Chart 的转换工具
  - GitHub: https://github.com/mumoshu/helmify
  - 用于将 Kubernetes 资源打包为 Helm Chart

### 输入格式支持
- Docker Compose YAML 文件

### 输出格式
- 标准 Helm Chart（符合 Helm v3 规范）

### 核心技术栈
- Go 语言（命令行工具开发）
- Kompose（Docker Compose → Kubernetes 转换）
- Helmify（Kubernetes → Helm Chart 转换）
- Helm（Chart 验证和部署）
- YAML 解析库

### 架构设计
```
HelmForge CLI
    ├── 输入解析模块（Docker Compose）
    ├── Kompose 集成模块
    ├── Helmify 集成模块
    ├── 配置参数化模块
    └── 验证与测试模块
```

## 使用场景

### 场景：已有 Docker Compose 配置
```bash
helmforge compose-to-chart -i docker-compose.yaml -o ./output-chart
```

## 使用示例

### 示例：从 Docker Compose 文件生成 Helm Chart

1. **准备 Docker Compose 文件**
   使用 `examples/docker-compose.yaml` 作为示例：
   ```bash
   cd examples
   helmforge compose-to-chart -i docker-compose.yaml -o ../output-chart
   ```

2. **查看生成的 Helm Chart**
   ```bash
   ls -la ../output-chart
   ```

3. **验证 Helm Chart**
   ```bash
   helm lint ../output-chart
   helm template test-release ../output-chart
   ```

## 项目结构

```
HelmForge/
├── cmd/
│   └── helmforge/           # 命令行工具入口
├── pkg/
│   ├── parser/              # Docker Compose 解析模块
│   ├── kompose/             # Kompose 集成模块
│   ├── helmify/             # Helmify 集成模块
│   ├── config/              # 配置参数化模块
│   └── validate/            # 验证与测试模块
├── internal/
│   └── models/              # 数据模型
├── examples/                # 示例文件
│   └── docker-compose.yaml  # 示例 Docker Compose 文件
├── go.mod                   # Go 模块文件
└── README.md                # 项目说明
```

## 安装与配置

### 安装

```bash
# 克隆项目
git clone https://github.com/wenwenxiong/HelmForge.git
cd HelmForge

# 构建项目
go build -o helmforge ./cmd/helmforge

# 安装到系统路径
sudo mv helmforge /usr/local/bin/

# 验证安装
helmforge --version
```

### 依赖要求

- Go 1.20+
- Helm 3+

## 开发指南

### 代码结构

- **cmd/helmforge/**: 命令行工具入口，定义命令和参数
- **pkg/parser/**: 解析 Docker Compose 文件
- **pkg/kompose/**: 集成 Kompose 工具，转换 Docker Compose 为 Kubernetes 资源
- **pkg/helmify/**: 集成 Helmify 工具，转换 Kubernetes 资源为 Helm Chart
- **pkg/config/**: 增强 Helm Chart 配置
- **pkg/validate/**: 验证生成的 Helm Chart 是否正确

### 开发流程

1. **修改代码**
2. **构建项目**：`go build -o helmforge ./cmd/helmforge`
3. **测试功能**：使用示例文件测试功能
4. **运行验证**：`helm lint` 和 `helm template` 验证生成的 Chart

## 故障排查

### 常见问题

1. **Kompose 转换失败**
   - 检查 Docker Compose 文件格式是否正确
   - 检查 Docker Compose 版本是否兼容

2. **Helm Chart 验证失败**
   - 检查生成的 Chart 结构是否正确
   - 检查模板文件语法是否正确
   - 检查 values.yaml 格式是否正确

### 日志查看

```bash
# 启用详细日志
helmforge --debug compose-to-chart -i docker-compose.yaml -o ./output-chart
```

## 项目目标

### 短期目标
- [x] 集成 Kompose 和 Helmify 到命令行工具
- [x] 实现 Docker Compose 到 Helm Chart 的完整转换流程
- [x] 提供基础的命令行工具框架
- [x] 支持基本的 Docker Compose 指令转换

### 中期目标
- [ ] 扩展 Kompose 支持更多 Docker Compose 高级特性
- [ ] 优化 Helmify 生成的 Chart 模板质量
- [ ] 实现智能配置参数化功能

### 长期目标
- [ ] 提供 Web 可视化界面
- [ ] 支持更多部署模板和场景
- [ ] 集成 CI/CD 流程

## 待解决问题

1. **Kompose 扩展**：如何扩展 Kompose 以支持更多 Docker Compose 高级特性（如 healthcheck、depends_on 条件等）
2. **Helmify 优化**：如何优化 Helmify 生成的 Chart 模板，使其更符合生产环境最佳实践
3. **配置参数化策略**：如何智能识别哪些配置需要参数化，并自动生成合理的 values.yaml 结构
4. **复杂场景处理**：如何处理有状态服务、初始化容器、ConfigMap/Secret 管理等复杂场景
5. **版本兼容性**：如何处理 Kompose 和 Helmify 版本升级带来的兼容性问题

## 参考资源

### 相关工具
- [Kompose 官方文档](https://kompose.io/)
- [Kompose GitHub](https://github.com/kubernetes/kompose)
- [Helmify GitHub](https://github.com/mumoshu/helmify)
- [Helm 官方文档](https://helm.sh/docs/)

### 设计理念
- 复用成熟工具（Kompose + Helmify）的核心能力
- 通过二次开发增强功能，而非从零开始
- 保持与上游工具的兼容性，便于升级维护
- 提供更友好的用户体验和错误处理

## 贡献指南
欢迎提交 Issue 和 Pull Request！