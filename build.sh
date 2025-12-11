#!/bin/bash

# gh-pitfall-scraper 构建和运行脚本

set -e

echo "🚀 gh-pitfall-scraper 构建脚本"
echo "==============================="

# 检查参数
SKIP_DB_INIT=false
SKIP_TESTS=false
SKIP_DEPS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-db-init)
            SKIP_DB_INIT=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        -h|--help)
            echo "使用方法: $0 [选项]"
            echo "选项:"
            echo "  --skip-db-init  跳过数据库初始化"
            echo "  --skip-tests    跳过测试"
            echo "  --skip-deps     跳过依赖下载"
            echo "  -h, --help      显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 -h 或 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21 或更高版本"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 创建必要目录
mkdir -p output data backups

# 安装依赖
if [ "$SKIP_DEPS" = false ]; then
    echo "📦 下载依赖..."
    go mod tidy
else
    echo "⏭️  跳过依赖下载"
fi

# 运行测试
if [ "$SKIP_TESTS" = false ]; then
    echo "🧪 运行单元测试..."
    go test ./internal/scraper/... -v
    
    # 如果数据库相关代码存在，运行数据库测试
    if [ -d "./internal/database" ]; then
        echo "🧪 运行数据库测试..."
        go test ./internal/database/... -v || echo "⚠️  数据库测试部分失败，继续构建"
    fi
else
    echo "⏭️  跳过测试"
fi

# 初始化数据库
if [ "$SKIP_DB_INIT" = false ]; then
    echo "🗄️  初始化数据库..."
    
    # 检查数据库文件是否存在
    if [ ! -f "data/gh-pitfall-scraper.db" ]; then
        echo "📝 创建数据库文件..."
        # 尝试运行数据库初始化脚本
        if [ -f "./scripts/init-database.sh" ]; then
            chmod +x ./scripts/init-database.sh
            ./scripts/init-database.sh
        else
            echo "⚠️  未找到数据库初始化脚本，跳过初始化"
            echo "💡 运行 'make db-init' 手动初始化数据库"
        fi
    else
        echo "✅ 数据库文件已存在，跳过初始化"
    fi
else
    echo "⏭️  跳过数据库初始化"
fi

# 构建可执行文件
echo "🔨 构建可执行文件..."
go build -o gh-pitfall-scraper main.go

# 设置可执行权限
chmod +x gh-pitfall-scraper

echo "✅ 构建完成！"
echo ""
echo "📋 使用方法:"
echo "1. 编辑 config.yaml 文件，设置您的 GitHub Token"
echo "2. 运行: ./gh-pitfall-scraper"
echo "3. 查看输出: output/issues.json"
echo ""

# 配置文件检查
echo "🎯 配置检查:"
if grep -q "ghp_xxx" config.yaml; then
    echo "⚠️  请在 config.yaml 中设置您的真实 GitHub Token"
    echo "💡 访问 https://github.com/settings/tokens 创建 Personal Access Token"
else
    echo "✅ GitHub Token 已配置"
fi

# 数据库检查
if [ -f "data/gh-pitfall-scraper.db" ]; then
    echo "✅ 数据库已准备就绪"
    echo "💡 运行 'make db-stats' 查看数据库统计"
else
    echo "⚠️  数据库未初始化"
    echo "💡 运行 'make db-init' 初始化数据库"
fi

echo ""
echo "🎉 准备就绪！运行 ./gh-pitfall-scraper 开始抓取"
echo ""
echo "🔧 其他有用命令:"
echo "  make db-backup    - 创建数据库备份"
echo "  make db-restore   - 恢复数据库备份"
echo "  make db-stats     - 查看数据库统计"
echo "  make help         - 查看所有可用命令"