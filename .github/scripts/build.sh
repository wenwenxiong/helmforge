#!/bin/bash

# HelmForge 跨平台构建脚本
# 支持多个操作系统和架构的交叉编译

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 默认配置
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
OUTPUT_DIR="${OUTPUT_DIR:-./dist}"
MAIN_PKG="./cmd/helmforge"

# 构建矩阵
declare -A BUILD_MATRIX
BUILD_MATRIX["linux-amd64"]="linux amd64"
BUILD_MATRIX["darwin-amd64"]="darwin amd64"
BUILD_MATRIX["darwin-arm64"]="darwin arm64"
BUILD_MATRIX["windows-amd64"]="windows amd64"

# 检查依赖
check_dependencies() {
    log_step "检查构建依赖..."

    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        return 1
    fi

    local go_version=$(go version | awk '{print $3}')
    log_info "Go 版本: $go_version"

    log_info "依赖检查通过"
}

# 设置构建环境
setup_build_env() {
    log_step "设置构建环境..."

    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"

    # 设置构建标志
    export CGO_ENABLED=0
    export GOFLAGS="-trimpath -ldflags=-X=main.Version=${VERSION} -X=main.BuildTime=${BUILD_TIME}"

    log_info "版本: $VERSION"
    log_info "构建时间: $BUILD_TIME"
    log_info "输出目录: $OUTPUT_DIR"
}

# 清理旧构建产物
clean_build() {
    log_step "清理旧构建产物..."

    rm -rf "$OUTPUT_DIR"/*
    log_info "清理完成"
}

# 构建单个平台
build_platform() {
    local platform_key=$1
    local os_type=$2
    local arch=$3

    log_step "构建 $platform_key..."

    export GOOS="$os_type"
    export GOARCH="$arch"

    local binary_name="helmforge"
    if [ "$os_type" = "windows" ]; then
        binary_name="helmforge.exe"
    fi

    local output_path="$OUTPUT_DIR/${platform_key}/${binary_name}"
    mkdir -p "$(dirname "$output_path")"

    if go build -o "$output_path" "$MAIN_PKG"; then
        log_info "✓ $platform_key 构建成功"

        # 生成 checksum
        cd "$(dirname "$output_path")"
        if command -v sha256sum &> /dev/null; then
            sha256sum "$(basename "$binary_name")" > "$(basename "$binary_name").sha256"
            log_info "✓ Checksum 生成成功"
        elif command -v shasum &> /dev/null; then
            shasum -a 256 "$(basename "$binary_name")" > "$(basename "$binary_name").sha256"
            log_info "✓ Checksum 生成成功"
        fi
        cd - > /dev/null
    else
        log_error "✗ $platform_key 构建失败"
        return 1
    fi
}

# 构建所有平台
build_all_platforms() {
    log_step "构建所有平台..."

    for platform_key in "${!BUILD_MATRIX[@]}"; do
        IFS=' ' read -r os_type arch <<< "${BUILD_MATRIX[$platform_key]}"
        build_platform "$platform_key" "$os_type" "$arch"
    done

    log_info "所有平台构建完成"
}

# 构建指定平台
build_specific_platforms() {
    log_step "构建指定平台: $PLATFORMS..."

    for platform_key in $(echo $PLATFORMS | tr ',' ' '); do
        if [ -n "${BUILD_MATRIX[$platform_key]}" ]; then
            IFS=' ' read -r os_type arch <<< "${BUILD_MATRIX[$platform_key]}"
            build_platform "$platform_key" "$os_type" "$arch"
        else
            log_warn "不支持的平台: $platform_key"
        fi
    done
}

# 构建当前平台
build_current_platform() {
    log_step "构建当前平台..."

    local os_type=$(go env GOOS)
    local arch=$(go env GOARCH)
    local platform_key="${os_type}-${arch}"

    build_platform "$platform_key" "$os_type" "$arch"
}

# 打包构建产物
package_artifacts() {
    log_step "打包构建产物..."

    cd "$OUTPUT_DIR"

    for dir in */; do
        if [ -d "$dir" ]; then
            local platform=$(basename "$dir")
            local archive_name="helmforge-${VERSION}-${platform}.tar.gz"

            if [ -f "${dir}helmforge" ] || [ -f "${dir}helmforge.exe" ]; then
                tar -czf "$archive_name" -C "$dir" .
                log_info "✓ 打包: $archive_name"
            fi
        fi
    done

    cd - > /dev/null
}

# 生成构建信息
generate_build_info() {
    log_step "生成构建信息..."

    local build_info_file="$OUTPUT_DIR/build-info.json"

    cat > "$build_info_file" <<EOF
{
  "version": "$VERSION",
  "build_time": "$BUILD_TIME",
  "git_commit": "$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")",
  "git_branch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")",
  "platforms": [
EOF

    local first=true
    for platform_key in "${!BUILD_MATRIX[@]}"; do
        if [ "$first" = false ]; then
            echo "," >> "$build_info_file"
        fi
        first=false

        IFS=' ' read -r os_type arch <<< "${BUILD_MATRIX[$platform_key]}"
        echo "    {\"platform\": \"$platform_key\", \"os\": \"$os_type\", \"arch\": \"$arch\"}" >> "$build_info_file"
    done

    cat >> "$build_info_file" <<EOF

  ]
}
EOF

    log_info "✓ 构建信息生成成功"
}

# 显示帮助信息
show_help() {
    cat << EOF
HelmForge 构建脚本

用法: $0 [选项]

选项:
    -h, --help              显示此帮助信息
    -c, --clean             清理构建产物
    -a, --all               构建所有平台
    -p, --platforms <list>  构建指定平台（逗号分隔）
    -v, --version <string>  设置版本号
    -o, --output <dir>      设置输出目录

支持的平台:
    linux-amd64, darwin-amd64, darwin-arm64, windows-amd64

示例:
    $0 --all                              构建所有平台
    $0 --platforms linux-amd64,windows-amd64  构建指定平台
    $0 --version v1.0.0 --all            使用特定版本构建
    $0 --clean                             清理构建产物

环境变量:
    VERSION          版本号（默认从 Git 标签获取）
    BUILD_TIME       构建时间（默认为当前时间）
    OUTPUT_DIR       输出目录（默认为 ./dist）
EOF
}

# 主函数
main() {
    # 解析参数
    local clean_only=false
    local build_all=false
    local build_platforms_list=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--clean)
                clean_only=true
                shift
                ;;
            -a|--all)
                build_all=true
                shift
                ;;
            -p|--platforms)
                PLATFORMS="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    log_info "HelmForge 构建脚本启动"
    log_info "版本: $VERSION"
    log_info "构建时间: $BUILD_TIME"

    # 执行构建步骤
    check_dependencies
    setup_build_env

    if [ "$clean_only" = true ]; then
        clean_build
        log_info "清理完成"
        exit 0
    fi

    if [ -n "$PLATFORMS" ]; then
        build_specific_platforms
    elif [ "$build_all" = true ]; then
        build_all_platforms
    else
        build_current_platform
    fi

    package_artifacts
    generate_build_info

    log_info "构建完成！"
    log_info "构建产物位置: $OUTPUT_DIR"
    ls -lh "$OUTPUT_DIR"
}

# 脚本入口
main "$@"