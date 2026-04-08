# GitHub CI/CD 配置文档

## 📋 概述

HelmForge 使用 GitHub Actions 实现自动化 CI/CD 流程，包括代码质量检查、测试、构建和部署。

## 🔧 配置文件

### 主要工作流

#### 1. **CI 工作流** (`.github/workflows/ci.yml`)

**触发条件**:
- Push 到 main/master/develop 分支
- Pull Request 到 main/master/develop 分支

**执行步骤**:
1. **代码检出**: 包含 Git Submodules 递归克隆
2. **Submodules 初始化**: 使用自定义脚本初始化第三方依赖
3. **Go 环境设置**: 支持 Go 1.20, 1.21, 1.22 矩阵
4. **依赖缓存**: 缓存 Go modules 和构建缓存
5. **代码检查**:
   - gofmt 代码格式检查
   - go vet 静态分析
   - golangci-lint 严格代码质量检查
6. **测试执行**: 运行单元测试并生成覆盖率报告
7. **覆盖率上传**: 上传到 Codecov
8. **项目构建**: 验证构建过程
9. **Docker 构建**: (仅 main 分支) 构建并推送 Docker 镜像
10. **安全扫描**: Trivy 镜像漏洞扫描

**优化特性**:
- Go modules 缓存，减少依赖下载时间
- 矩阵策略，并行执行测试
- 条件执行，避免不必要的 Docker 构建
- GitHub Actions 缓存 Docker layers

#### 2. **golangci-lint 配置** (`.github/golangci.yml`)

**严格检查策略**:
- 启用 30+ 个 linters
- 所有警告视为错误
- 复杂度限制：最大 20
- 函数长度限制：最大 50 行
- 代码覆盖率：建议但不强制

**主要 Linters**:
- **安全性**: gosec, govet
- **性能**: ineffassign, prealloc
- **代码质量**: errcheck, staticcheck, revive
- **复杂度**: gocyclo, cyclop, funlen
- **风格**: gofmt, goimports, misspell

**自定义规则**:
- 允许测试中的 TODO 注释
- 允许测试中的 fmt.Print
- 排除第三方代码检查
- 针对外部包的特定规则放宽

#### 3. **Dependabot 配置** (`.github/dependabot.yml`)

**自动更新策略**:
- 检查频率：每日
- Go 依赖：自动 PR
- Docker 基础镜像：每周检查
- 最大并发 PR 数：10
- 忽略 kompose/helmify 的主版本更新

### 辅助脚本

#### 1. **Submodules 初始化脚本** (`.github/scripts/setup-submodules.sh`)

**功能**:
- 安全初始化 Git Submodules
- 带重试机制的更新
- Submodule 状态验证
- 详细的状态报告

**使用方式**:
```bash
# 手动运行脚本
bash .github/scripts/setup-submodules.sh

# 或在 CI 中自动调用
- name: 初始化 Submodules
  run: bash .github/scripts/setup-submodules.sh
```

**特性**:
- 彩色日志输出
- 错误重试机制（最多3次）
- 详细的状态信息
- 后续步骤建议

#### 2. **跨平台构建脚本** (`.github/scripts/build.sh`)

**功能**:
- 支持多平台交叉编译
- 自动版本号生成
- Checksum 生成
- 构建产物打包
- 构建信息生成

**支持的平台**:
- Linux/amd64
- macOS/amd64
- macOS/arm64
- Windows/amd64

**使用方式**:
```bash
# 构建当前平台
bash .github/scripts/build.sh

# 构建所有平台
bash .github/scripts/build.sh --all

# 构建指定平台
bash .github/scripts/build.sh --platforms linux-amd64,windows-amd64

# 使用特定版本构建
bash .github/scripts/build.sh --version v1.0.0 --all

# 清理构建产物
bash .github/scripts/build.sh --clean

# 查看帮助
bash .github/scripts/build.sh --help
```

**输出产物**:
```
dist/
├── build-info.json           # 构建信息
├── linux-amd64/
│   ├── helmforge
│   ├── helmforge.sha256
│   └── helmforge-v1.0.0-linux-amd64.tar.gz
├── darwin-amd64/
│   ├── helmforge
│   ├── helmforge.sha256
│   └── helmforge-v1.0.0-darwin-amd64.tar.gz
└── windows-amd64/
    ├── helmforge.exe
    ├── helmforge.exe.sha256
    └── helmforge-v1.0.0-windows-amd64.tar.gz
```

## 🔐 GitHub Secrets 配置

### 必需的 Secrets

在 GitHub 仓库设置中配置以下 Secrets：

#### `DOCKER_USERNAME`
- **用途**: Docker Hub 登录用户名
- **获取方式**: Docker Hub 账户用户名
- **配置位置**: Settings → Secrets and variables → Actions → New repository secret

#### `DOCKER_PASSWORD`
- **用途**: Docker Hub 访问令牌
- **获取方式**: Docker Hub 账户设置 → Access Tokens
- **权限建议**: Read, Write, Delete (仅仓库访问)

### 自动提供的 Secrets

#### `GITHUB_TOKEN`
- **用途**: GitHub API 访问
- **来源**: GitHub Actions 自动提供
- **权限**: `contents: write`, `packages: write`

## 🚀 使用指南

### 本地开发

#### 初始化 Submodules
```bash
git clone --recursive https://github.com/wenwenxiong/HelmForge.git
cd HelmForge

# 或手动初始化
git submodule init
git submodule update
```

#### 运行代码检查
```bash
# 格式检查
gofmt -s -l .

# 静态分析
go vet ./...

# 运行 golangci-lint
golangci-lint run

# 或使用自定义配置
golangci-lint run --config .github/golangci.yml
```

#### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# 查看覆盖率
go tool cover -html=coverage.out
```

#### 构建项目
```bash
# 构建当前平台
go build -o helmforge ./cmd/helmforge

# 使用构建脚本
bash .github/scripts/build.sh
```

### CI/CD 工作流程

#### 推送代码触发 CI
```bash
# 创建功能分支
git checkout -b feature/my-feature

# 提交更改
git add .
git commit -m "feat: 添加新功能"

# 推送并触发 CI
git push origin feature/my-feature

# 创建 Pull Request
```

#### 发布新版本
```bash
# 打标签
git tag -a v1.0.0 -m "Release v1.0.0"

# 推送标签（触发发布流程）
git push origin v1.0.0
```

## 📊 监控和报告

### GitHub Actions 监控

**查看执行历史**:
1. 访问仓库的 Actions 标签页
2. 查看工作流执行历史
3. 点击具体的执行查看详细日志

**失败排查**:
1. 查看失败步骤的详细日志
2. 检查 golangci-lint 错误报告
3. 验证测试失败原因
4. 检查 Docker 构建日志

### 代码质量监控

**golangci-lint 报告**:
- 在 CI 执行中查看详细的 lint 报告
- 下载 linter 结果文件
- 使用 `golangci-lint run --out-format=html > report.html` 本地生成报告

**测试覆盖率**:
- 查看 Codecov 报告
- 检查覆盖率趋势
- 识别未覆盖的代码区域

### Docker 镜像监控

**镜像扫描结果**:
- 查看在 CI 中生成的安全扫描报告
- 在 Docker Hub 查看镜像扫描状态
- 使用 `docker scan wenwenxiong/helmforge:latest` 本地扫描

## 🛠️ 故障排查

### 常见问题

#### 1. Submodules 初始化失败
**症状**: `fatal: not a git repository`
**解决**:
```bash
rm -rf .git/modules/*
git submodule deinit --all
git submodule update --init --recursive
```

#### 2. golangci-lint 检查失败
**症状**: 某些 linters 报告错误
**解决**:
```bash
# 本地运行 lint
golangci-lint run --config .github/golangci.yml

# 修复问题后重新运行
golangci-lint run --fix
```

#### 3. Docker 推送失败
**症状**: `denied: insufficient permissions`
**解决**:
1. 验证 Docker Hub 凭据是否正确
2. 检查 Secrets 配置
3. 确认用户名和密码权限

#### 4. 测试失败
**症状**: 某些测试用例失败
**解决**:
```bash
# 运行失败的测试
go test -v ./pkg/parser -run TestParseDockerCompose

# 查看详细输出
go test -v -race ./...
```

### 性能优化

#### 加速 CI 执行
1. **启用缓存**: Go modules 和 Docker layers 缓存
2. **并行执行**: 矩阵策略并行化测试
3. **条件执行**: 避免不必要的步骤
4. **超时设置**: 合理设置各步骤超时

#### 减少构建时间
1. **使用 Buildx**: Docker Buildx 加速镜像构建
2. **缓存 Docker layers**: 复用之前的构建缓存
3. **多阶段构建**: 减小最终镜像大小

## 📈 后续优化建议

### 短期优化
1. **添加更多平台**: 支持 ARM64 架构
2. **集成更多测试**: 集成测试和端到端测试
3. **性能基准测试**: 添加性能测试基准

### 长期规划
1. **自动发布流程**: Git tag 触发自动发布
2. **多环境部署**: 开发/测试/生产环境
3. **集成监控**: Prometheus/Grafana 监控
4. **自动化文档**: API 文档自动生成

## 🔗 相关资源

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [golangci-lint 配置](https://golangci-lint.run/usage/configuration/)
- [Docker Buildx 文档](https://docs.docker.com/buildx/working-with-buildx/)
- [Dependabot 文档](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates)
- [Codecov 文档](https://docs.codecov.com/)

## 📞 支持和反馈

如有问题或建议，请：
1. 创建 GitHub Issue
2. 查看 GitHub Actions 执行日志
3. 检查文档和故障排查指南