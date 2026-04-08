# 集成 Kompose 和 Helmify Go 包

## 说明

项目已经配置好使用 Kompose 和 Helmify 的 Go 包，但需要您手动下载这两个项目的代码到 third-party 目录。

## 下载步骤

### 1. 下载 Kompose 代码

```bash
# 进入 third-party 目录
cd third-party

# 克隆 Kompose 仓库
git clone https://github.com/kubernetes/kompose.git

# 切换到指定版本
cd kompose
git checkout v1.34.0
```

### 2. 下载 Helmify 代码

```bash
# 返回 third-party 目录
cd ..

# 克隆 Helmify 仓库
git clone https://github.com/arttor/helmify.git

# 切换到指定版本
cd helmify
git checkout v0.48.0
```

### 3. 更新依赖

下载完成后，运行以下命令更新依赖：

```bash
cd ..
go mod tidy
```

### 4. 构建项目

```bash
go build -o helmforge.exe ./cmd/helmforge
```

## 项目结构

下载完成后的目录结构应该是：

```
HelmForge/
├── third-party/
│   ├── kompose/          # Kompose 源代码
│   │   ├── pkg/
│   │   │   ├── kobject/
│   │   │   └── transform/
│   │   └── go.mod
│   └── helmify/          # Helmify 源代码
│       ├── pkg/
│       │   └── helmify/
│       └── go.mod
├── pkg/
│   ├── kompose/
│   │   └── kompose.go    # 已集成 Kompose Go 包
│   └── helmify/
│       └── helmify.go    # 已集成 Helmify Go 包
└── go.mod                # 已配置 replace 指令
```

## 集成说明

### Kompose 集成

- [pkg/kompose/kompose.go](pkg/kompose/kompose.go) 已更新为使用 Kompose Go 包
- 主要函数：`convertUsingKomposePackage()`
- 备用方案：如果 Go 包不可用，会自动降级到命令行工具

### Helmify 集成

- [pkg/helmify/helmify.go](pkg/helmify/helmify.go) 已更新为使用 Helmify Go 包
- 主要函数：`convertUsingHelmifyPackage()`
- 备用方案：如果 Go 包不可用，会自动降级到命令行工具

## go.mod 配置

```go
module github.com/wenwenxiong/HelmForge

go 1.20

require (
	github.com/arttor/helmify v0.48.0
	github.com/kubernetes/kompose v1.34.0
	github.com/spf13/cobra v1.8.0
	gopkg.in/yaml.v3 v3.0.1
)

// 本地集成 kompose 和 helmify
replace github.com/kubernetes/kompose => ./third-party/kompose
replace github.com/arttor/helmify => ./third-party/helmify
```

## 优势

使用 Go 包集成相比命令行工具的优势：

1. **更好的性能**：不需要启动外部进程
2. **更灵活的控制**：可以直接调用内部 API
3. **更好的错误处理**：可以捕获和处理更详细的错误信息
4. **更容易二次开发**：可以直接访问和修改转换逻辑
5. **无依赖外部工具**：不需要用户安装 kompose 和 helmify 命令行工具

## 注意事项

1. 确保 Go 版本为 1.20 或更高
2. 下载代码后需要运行 `go mod tidy` 更新依赖
3. 如果遇到依赖问题，可以尝试删除 go.sum 文件后重新运行 `go mod tidy`
4. 项目已经实现了降级机制，即使 Go 包不可用，也可以使用命令行工具
