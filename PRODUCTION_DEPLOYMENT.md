# HelmForge 生产部署指南

本文档提供了 HelmForge 工具的生产容器化部署指南，包括 Docker 容器构建、运行和最佳实践。

## 目录

- [容器化架构](#容器化架构)
- [构建 Docker 镜像](#构建-docker-镜像)
- [运行 HelmForge 容器](#运行-helmforge-容器)
- [Docker Compose 部署](#docker-compose-部署)
- [生产环境配置](#生产环境配置)
- [安全最佳实践](#安全最佳实践)
- [常见问题排查](#常见问题排查)

## 容器化架构

HelmForge 使用多阶段构建策略，确保生产镜像最小化且安全：

1. **构建阶段** (`builder`):
   - 使用 `golang:1.20-alpine` 作为基础镜像
   - 安装构建依赖（git、build-base）
   - 下载 Go 依赖并构建应用

2. **运行时阶段**:
   - 使用 `alpine:3.18` 作为基础镜像
   - 安装运行时依赖（bash、curl、git、docker、docker-compose、helm）
   - 安装外部工具（kompose、helmify）
   - 创建非 root 用户并设置权限
   - 复制构建产物和示例文件

## 构建 Docker 镜像

### 基本构建

```bash
# 构建默认镜像
docker build -t helmforge:latest .

# 构建生产镜像
docker build -t helmforge:production .

# 构建特定版本
docker build -t helmforge:v1.0.0 .
```

### 构建参数

可以使用以下构建参数来自定义构建过程：

```bash
# 使用自定义 Go 版本
docker build --build-arg GO_VERSION=1.21 -t helmforge:latest .

# 使用自定义 Alpine 版本
docker build --build-arg ALPINE_VERSION=3.19 -t helmforge:latest .
```

## 运行 HelmForge 容器

### 基本运行

```bash
# 查看帮助信息
docker run --rm helmforge:latest --help

# 从 Docker Compose 文件生成 Helm Chart
docker run --rm -v $(pwd):/app helmforge:latest compose-to-chart -i examples/docker-compose.yaml -o output-chart

# 从源代码生成 Helm Chart
docker run --rm -v $(pwd):/app helmforge:latest code-to-chart -s ./path/to/source -o output-chart

# 从部署手册生成 Helm Chart
docker run --rm -v $(pwd):/app helmforge:latest manual-to-chart -m examples/deployment-manual.md -o output-chart
```

### 挂载 Docker 套接字

为了使 HelmForge 能够构建 Docker 镜像，需要挂载 Docker 套接字：

```bash
docker run --rm \
  -v $(pwd):/app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  helmforge:latest code-to-chart -s ./path/to/source -o output-chart
```

### 持久化 Helm 配置

```bash
docker run --rm \
  -v $(pwd):/app \
  -v ${HOME}/.helm:/app/.helm \
  helmforge:latest compose-to-chart -i docker-compose.yaml -o output-chart
```

## Docker Compose 部署

使用提供的 `docker-compose.yaml` 文件可以更方便地管理 HelmForge 容器：

### 开发环境

```bash
# 启动开发容器
docker-compose up helmforge

# 在开发容器中执行命令
docker-compose run --rm helmforge compose-to-chart -i examples/docker-compose.yaml -o output-chart
```

### 生产环境

```bash
# 启动生产容器
docker-compose up -d helmforge-production

# 检查容器状态
docker-compose ps

# 查看容器日志
docker-compose logs helmforge-production

# 停止容器
docker-compose down
```

## 生产环境配置

### 资源限制

在生产环境中，建议为容器设置资源限制：

```yaml
# 在 docker-compose.yaml 中添加
resources:
  limits:
    cpus: '1.0'
    memory: 1G
  reservations:
    cpus: '0.5'
    memory: 512M
```

### 环境变量

| 环境变量 | 描述 | 默认值 |
|---------|------|-------|
| `DOCKER_HOST` | Docker 主机地址 | `unix:///var/run/docker.sock` |
| `HELM_HOME` | Helm 配置目录 | `/app/.helm` |
| `KUBECONFIG` | Kubernetes 配置文件路径 | - |

### 健康检查

生产容器已配置健康检查：

```yaml
healthcheck:
  test: ["CMD", "helmforge", "--help"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

## 安全最佳实践

### 容器安全

1. **使用非 root 用户**：容器以 `helmforge` 用户（UID 1001）运行
2. **最小化基础镜像**：使用 Alpine 作为基础镜像，减小攻击面
3. **定期更新依赖**：定期更新基础镜像和安装的工具
4. **限制容器能力**：在生产环境中使用 `--cap-drop=all` 和 `--security-opt=no-new-privileges`

### 运行时安全

1. **避免挂载敏感目录**：只挂载必要的目录
2. **使用 Docker 机密**：对于敏感信息，使用 Docker 机密管理
3. **网络隔离**：使用自定义网络隔离容器
4. **监控容器**：实施容器监控和日志记录

### 示例：安全运行

```bash
docker run --rm \
  --user 1001:1001 \
  --cap-drop=all \
  --security-opt=no-new-privileges \
  --network=helmforge-network \
  -v $(pwd):/app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  helmforge:production compose-to-chart -i docker-compose.yaml -o output-chart
```

## 常见问题排查

### Docker 权限问题

**症状**：`Got permission denied while trying to connect to the Docker daemon socket`

**解决方案**：
1. 确保挂载了 Docker 套接字：`-v /var/run/docker.sock:/var/run/docker.sock`
2. 确保容器内用户有权限访问套接字
3. 在主机上，将当前用户添加到 docker 组：`sudo usermod -aG docker $USER`

### 工具安装失败

**症状**：`Error: exec: "kompose": executable file not found in $PATH`

**解决方案**：
1. 检查 Dockerfile 中的工具安装命令
2. 确保网络连接正常
3. 尝试手动构建镜像：`docker build -t helmforge:latest .`

### 权限被拒绝

**症状**：`permission denied` 当访问文件或目录时

**解决方案**：
1. 确保挂载的卷有正确的权限
2. 使用 `--user $(id -u):$(id -g)` 运行容器
3. 检查主机文件系统权限

### Helm 命令失败

**症状**：`Error: Kubernetes cluster unreachable`

**解决方案**：
1. 确保 Kubernetes 集群可访问
2. 挂载 KUBECONFIG 文件：`-v ${HOME}/.kube/config:/app/.kube/config -e KUBECONFIG=/app/.kube/config`
3. 检查集群凭证是否有效

## 性能优化

### 构建优化

1. **使用 BuildKit**：`DOCKER_BUILDKIT=1 docker build -t helmforge:latest .`
2. **利用缓存**：保持 Dockerfile 指令顺序，频繁变化的指令放在后面
3. **并行构建**：使用 `docker buildx` 进行并行构建

### 运行时优化

1. **限制资源**：为容器设置适当的资源限制
2. **使用体积挂载**：对于频繁访问的文件，使用体积挂载
3. **避免不必要的文件复制**：只复制必要的文件到容器

## 多平台构建

使用 `docker buildx` 可以构建多平台镜像：

```bash
# 创建构建器
docker buildx create --name multiarch --use

# 构建多平台镜像
docker buildx build --platform linux/amd64,linux/arm64 -t helmforge:latest --push .
```

## 结论

通过本文档提供的容器化部署方案，您可以在生产环境中安全、高效地运行 HelmForge 工具。使用多阶段构建确保了镜像的最小化和安全性，而 Docker Compose 配置则简化了容器管理。

如需进一步定制或有任何问题，请参考项目文档或联系开发团队。
