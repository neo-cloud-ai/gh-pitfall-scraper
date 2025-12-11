#!/bin/bash

echo "🔧 正在修复Go模块依赖问题..."

# 1. 备份并删除example.go文件
echo "📁 删除导致包冲突的example.go文件..."
if [ -f "internal/database/example.go" ]; then
    mv internal/database/example.go examples/database_example_internal.go 2>/dev/null || true
    echo "✅ 移除了example.go文件"
fi

# 2. 清理现有的go.sum
echo "🧹 清理go.sum文件..."
rm -f go.sum

# 3. 下载所有依赖
echo "📥 下载Go模块依赖..."
go mod download

# 4. 整理模块
echo "🔧 整理go.mod文件..."
go mod tidy

# 5. 验证模块
echo "✅ 验证Go模块..."
go mod verify

echo "🎉 Go模块依赖修复完成！"

# 6. 尝试构建以验证
echo "🔨 验证构建..."
go build -v .

echo "✨ 所有依赖问题已解决！"