#!/bin/bash

# HelmForge Git Submodules 初始化脚本
# 该脚本用于安全地初始化和更新 Git Submodules

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 检查是否在 Git 仓库中
check_git_repo() {
    if [ ! -d ".git" ]; then
        log_error "当前目录不是 Git 仓库"
        exit 1
    fi
    log_info "Git 仓库检查通过"
}

# 初始化 Submodules
init_submodules() {
    log_info "初始化 Git Submodules..."

    # 检查是否有 submodules
    if [ ! -f ".gitmodules" ]; then
        log_warn "没有找到 .gitmodules 文件"
        return 0
    fi

    # 显示当前 submodule 配置
    log_info "Submodule 配置:"
    cat .gitmodules

    # 初始化 submodules
    if ! git submodule init; then
        log_error "Submodule 初始化失败"
        return 1
    fi

    log_info "Submodule 初始化成功"
}

# 更新 Submodules（带重试机制）
update_submodules() {
    log_info "更新 Git Submodules..."

    local max_retries=3
    local retry_count=0

    while [ $retry_count -lt $max_retries ]; do
        if git submodule update --recursive --remote --merge; then
            log_info "Submodule 更新成功"
            return 0
        else
            retry_count=$((retry_count + 1))
            if [ $retry_count -lt $max_retries ]; then
                log_warn "Submodule 更新失败，重试中... ($retry_count/$max_retries)"
                sleep 5
            else
                log_error "Submodule 更新失败，已达到最大重试次数"
                return 1
            fi
        fi
    done
}

# 验证 Submodule 状态
verify_submodules() {
    log_info "验证 Submodule 状态..."

    # 检查每个 submodule 的状态
    git submodule status | while read -r line; do
        if [[ $line == *"-"* ]]; then
            log_error "Submodule 未初始化: $line"
            return 1
        elif [[ $line == *"+"* ]]; then
            log_warn "Submodule 版本不匹配: $line"
        elif [[ $line == *"U"* ]]; then
            log_error "Submodule 有冲突: $line"
            return 1
        fi
    done

    log_info "Submodule 状态检查完成"
}

# 显示 Submodule 信息
show_submodule_info() {
    log_info "=== Submodule 信息 ==="

    if [ -f ".gitmodules" ]; then
        while IFS= read -r line; do
            if [[ $line == *"path"* ]]; then
                local path=$(echo "$line" | awk '{print $2}')
                local url=$(echo "$line" | awk '{print $3}')
                local branch=$(cd "$path" 2>/dev/null && git branch --show-current 2>/dev/null || echo "unknown")
                local commit=$(cd "$path" 2>/dev/null && git rev-parse --short HEAD 2>/dev/null || echo "unknown")

                echo "  路径: $path"
                echo "  URL: $url"
                echo "  分支: $branch"
                echo "  提交: $commit"
                echo "  ---"
            fi
        done < .gitmodules
    else
        log_warn "没有找到 Submodules"
    fi

    log_info "=== Submodule 信息结束 ==="
}

# 主函数
main() {
    log_info "开始 HelmForge Submodules 初始化流程"

    # 执行初始化步骤
    check_git_repo
    init_submodules
    update_submodules
    verify_submodules
    show_submodule_info

    log_info "Submodules 初始化流程完成！"

    # 显示后续建议
    echo ""
    log_info "后续步骤:"
    echo "  1. 运行 'go mod tidy' 更新 Go 依赖"
    echo "  2. 运行 'go build' 验证构建"
    echo "  3. 运行 'go test ./...' 执行测试"
}

# 脚本入口
main "$@"