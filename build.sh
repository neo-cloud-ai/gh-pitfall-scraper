#!/bin/bash

# gh-pitfall-scraper 构建和运行脚本

set -e

echo "🚀 gh-pitfall-scraper 构建脚本"
echo "==============================="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21 或更高版本"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 创建输出目录
mkdir -p output

echo "📦 下载依赖..."
go mod tidy

echo "🧪 运行单元测试..."
go test ./internal/scraper/... -v

echo "🔨 构建可执行文件..."
go build -o gh-pitfall-scraper main.go

echo "✅ 构建完成！"
echo ""
echo "📋 使用方法:"
echo "1. 编辑 config.yaml 文件，设置您的 GitHub Token"
echo "2. 运行: ./gh-pitfall-scraper"
echo "3. 查看输出: output/issues.json"
echo ""
echo "🎯 示例配置检查:"
if grep -q "ghp_xxx" config.yaml; then
    echo "⚠️  请在 config.yaml 中设置您的真实 GitHub Token"
else
    echo "✅ GitHub Token 已配置"
fi

echo ""
echo "🎉 准备就绪！运行 ./gh-pitfall-scraper 开始抓取"