#!/bin/bash

# HelmForge Git Submodules 初始化脚本
# 该脚本用于安全地初始化和更新 Git Submodules

set -e

# 超时和重试配置
SUBMODULE_TIMEOUT=300  # 5分钟超时
SUBMODULE_RETRIES=3   # 最大重试次数
SUBMODULE_RETRY_DELAY=20  # 重试间隔（秒）

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
                sleep $SUBMODULE_RETRY_DELAY
            else
                log_warn "达到最大重试次数（$max_retries），将使用本地版本"
                return 0
            fi
        fi
    done
}

# 验证 Submodule 状态
verify_submodules() {
    log_info "验证 Submodule 状态..."
    
    local has_critical_errors=false
    local has_warnings=false
    
    # 检查每个 submodule 的状态
    git submodule status | while read -r line; do
        if [[ $line == *"U"* ]]; then
            # 合并冲突是严重错误，需要手动解决
            log_error "Submodule 有合并冲突（需要手动解决）：$line"
            has_critical_errors=true
        elif [[ $line == *"-"* ]]; then
            # 未初始化是警告，降级策略会处理
            log_warn "Submodule 未初始化或部分失败（降级策略将处理）：$line"
            has_warnings=true
        elif [[ $line == *"+"* ]]; then
            # 版本不匹配是警告
            log_warn "Submodule 版本不匹配（降级策略将使用本地版本）：$line"
            has_warnings=true
        fi
    done
    
    # 总结验证结果
    if [ "$has_critical_errors" = false ] && [ "$has_warnings" = false ]; then
        log_info "所有 Submodules 状态完美"
        return 0
    elif [ "$has_critical_errors" = false ] && [ "$has_warnings" = true ]; then
        log_info "有警告但无严重错误，降级策略将处理"
        return 0
    else
        log_error "存在严重错误，需要手动解决"
        return 1
    fi
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
    
    # 执行初始化步骤（严格模式，严重错误会中断）
    if ! check_git_repo; then
        log_error "Git 仓库检查失败，无法继续"
        return 1
    fi
    
    if ! init_submodules; then
        log_error "Submodules 初始化失败，无法继续"
        return 1
    fi
    
    # 更新 Submodules（容忍模式，失败会降级）
    if ! update_submodules; then
        log_warn "Submodules 更新有问题，将使用本地版本"
    fi
    
    # 验证状态（容忍模式，只有严重错误会中断）
    if ! verify_submodules; then
        log_error "Submodules 验证失败（存在严重错误）"
        return 1
    fi
    
    # 显示最终状态
    show_submodule_info
    
    log_info "Submodules 初始化流程完成（容忍部分失败）"
    
    # 显示后续建议
    echo ""
    log_info "后续步骤:"
    echo "  1. 运行 'go mod tidy' 更新 Go 依赖"
    echo "  2. 运行 'go build' 验证构建"
    echo "  3. 运行 'go test ./...' 执行测试"
    
    return 0
}

# 脚本入口
main "$@"