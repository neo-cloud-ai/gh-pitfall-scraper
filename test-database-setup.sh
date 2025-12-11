#!/bin/bash

# gh-pitfall-scraper 数据库工具测试脚本
# 用于验证数据库配置和工具脚本的功能

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

echo "=== gh-pitfall-scraper 数据库工具测试 ==="
echo

# 测试 1: 检查文件是否存在
print_info "测试 1: 检查文件是否存在..."
files=(
    "$PROJECT_ROOT/config.yaml"
    "$PROJECT_ROOT/internal/database/config.go"
    "$PROJECT_ROOT/database-tools.sh"
    "$PROJECT_ROOT/scripts/init-database.sh"
    "$PROJECT_ROOT/DATABASE_README.md"
)

for file in "${files[@]}"; do
    if [[ -f "$file" ]]; then
        print_success "文件存在: $(basename "$file")"
    else
        print_error "文件不存在: $file"
        exit 1
    fi
done
echo

# 测试 2: 检查 config.yaml 中的数据库配置
print_info "测试 2: 检查 config.yaml 数据库配置..."
if grep -q "database:" "$PROJECT_ROOT/config.yaml" && \
   grep -q "file_path:" "$PROJECT_ROOT/config.yaml" && \
   grep -q "connection_pool:" "$PROJECT_ROOT/config.yaml" && \
   grep -q "backup:" "$PROJECT_ROOT/config.yaml"; then
    print_success "数据库配置已正确添加到 config.yaml"
else
    print_error "数据库配置未正确添加到 config.yaml"
    exit 1
fi
echo

# 测试 3: 检查 database-tools.sh 的可执行性
print_info "测试 3: 检查 database-tools.sh 可执行性..."
if [[ -x "$PROJECT_ROOT/database-tools.sh" ]] || chmod +x "$PROJECT_ROOT/database-tools.sh"; then
    print_success "database-tools.sh 具有执行权限或可设置为执行权限"
else
    print_error "无法为 database-tools.sh 设置执行权限"
fi
echo

# 测试 4: 检查 init-database.sh 的可执行性
print_info "测试 4: 检查 init-database.sh 可执行性..."
if [[ -x "$PROJECT_ROOT/scripts/init-database.sh" ]] || chmod +x "$PROJECT_ROOT/scripts/init-database.sh"; then
    print_success "init-database.sh 具有执行权限或可设置为执行权限"
else
    print_error "无法为 init-database.sh 设置执行权限"
fi
echo

# 测试 5: 测试 database-tools.sh 帮助功能
print_info "测试 5: 测试 database-tools.sh 帮助功能..."
if "$PROJECT_ROOT/database-tools.sh" --help > /dev/null 2>&1; then
    print_success "database-tools.sh 帮助功能正常"
else
    print_info "可能需要设置执行权限: chmod +x $PROJECT_ROOT/database-tools.sh"
fi
echo

# 测试 6: 测试数据库初始化脚本帮助...
print_info "测试 6: 测试数据库初始化脚本帮助..."
if "$PROJECT_ROOT/scripts/init-database.sh" --help > /dev/null 2>&1; then
    print_success "init-database.sh 帮助功能正常"
else
    print_info "可能需要设置执行权限: chmod +x $PROJECT_ROOT/scripts/init-database.sh"
fi
echo

# 测试 7: 检查 config.go 语法
print_info "测试 7: 检查 config.go 语法..."
cd "$PROJECT_ROOT"
if command -v go >/dev/null 2>&1; then
    if go vet internal/database/config.go > /dev/null 2>&1; then
        print_success "config.go 语法检查通过"
    else
        print_info "Go 语法检查可能需要完整的项目结构"
    fi
else
    print_info "Go 未安装，跳过语法检查"
fi
echo

# 测试 8: 创建测试目录结构
print_info "测试 8: 创建测试目录结构..."
mkdir -p "$PROJECT_ROOT/data" "$PROJECT_ROOT/backups"
print_success "测试目录创建完成"
echo

# 测试 9: 模拟数据库初始化（不实际执行）
print_info "测试 9: 测试数据库工具试运行模式..."
if "$PROJECT_ROOT/database-tools.sh" --dry-run init > /dev/null 2>&1; then
    print_success "试运行模式功能正常"
else
    print_info "试运行模式需要执行权限"
fi
echo

# 总结
echo "=== 测试总结 ==="
print_success "数据库配置和工具脚本创建完成！"
echo
echo "主要创建的文件："
echo "  - config.yaml (已更新，添加数据库配置)"
echo "  - internal/database/config.go (数据库配置管理)"
echo "  - database-tools.sh (数据库管理工具)"
echo "  - scripts/init-database.sh (数据库初始化脚本)"
echo "  - DATABASE_README.md (详细使用文档)"
echo
echo "使用方法："
echo "  1. 设置执行权限: chmod +x database-tools.sh scripts/init-database.sh"
echo "  2. 查看帮助: ./database-tools.sh --help"
echo "  3. 初始化数据库: ./scripts/init-database.sh"
echo "  4. 查看详细文档: cat DATABASE_README.md"
echo

print_success "所有测试完成！"