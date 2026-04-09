# 构建阶段
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git build-base

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod tidy

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o helmforge ./cmd/helmforge

# 运行时阶段
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    bash \
    curl \
    git \
    docker \
    docker-compose \
    helm \
    && rm -rf /var/cache/apk/*

# 安装kompose (如果需要)
RUN curl -L https://github.com/kubernetes/kompose/releases/download/v1.32.0/kompose-linux-amd64 -o /usr/local/bin/kompose \
    && chmod +x /usr/local/bin/kompose

# 安装helmify (如果需要)
RUN curl -L https://github.com/arttor/helmify/releases/download/v0.4.19/helmify_Linux_x86_64.tar.gz -o helmify.tar.gz \
    && tar -xzf helmify.tar.gz \
    && mv helmify /usr/local/bin/ \
    && chmod +x /usr/local/bin/helmify \
    && rm helmify.tar.gz

# 创建非root用户
RUN addgroup -g 1001 -S helmforge \
    && adduser -S helmforge -u 1001 -G helmforge

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder --chown=helmforge:helmforge /app/helmforge /usr/local/bin/

# 复制示例文件
COPY --chown=helmforge:helmforge examples/ /app/examples/

# 设置用户
USER helmforge

# 暴露端口（如果需要）
# EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD helmforge --help || exit 1

# 入口点
ENTRYPOINT ["helmforge"]
# 默认命令
CMD ["--help"]
