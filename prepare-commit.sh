#!/bin/bash

# gh-pitfall-scraper Git 提交准备脚本

echo "🚀 gh-pitfall-scraper Git 提交准备"
echo "=================================="

# 检查是否在git仓库中
if [ ! -d ".git" ]; then
    echo "📁 初始化Git仓库..."
    git init
    echo "✅ Git仓库已初始化"
else
    echo "✅ 已存在Git仓库"
fi

# 检查文件状态
echo ""
echo "📋 检查文件状态..."
git status

# 添加所有文件
echo ""
echo "📦 添加所有文件到暂存区..."
git add .

# 显示将要提交的文件
echo ""
echo "📝 准备提交的文件:"
git status --porcelain

# 提交代码
echo ""
echo "💾 提交代码..."
COMMIT_MESSAGE="feat: 初始版本 gh-pitfall-scraper

✨ 新功能:
- 完整的GitHub Issues抓取工具
- 智能多维度评分系统
- 高级过滤和筛选功能
- 支持6个主流AI/ML仓库
- 完整的单元测试覆盖

🏗️ 架构:
- 模块化设计，职责清晰
- GitHub API客户端封装
- 智能评分算法实现
- 批量抓取引擎

📚 文档:
- 详细README文档
- 完整API说明
- 使用示例和指南

🧪 测试:
- 单元测试覆盖核心逻辑
- 代码质量验证脚本
- 构建和部署脚本

🎯 用途:
专门用于从GitHub Issues中筛选高价值工程问题，
特别关注AI/ML领域的性能问题、GPU问题、分布式训练异常等。"

git commit -m "$COMMIT_MESSAGE"

echo ""
echo "✅ 代码提交完成！"
echo ""
echo "📋 提交信息:"
git log --oneline -1

echo ""
echo "🌐 如果需要推送到远程仓库，请运行:"
echo "git remote add origin https://github.com/neo-cloud-ai/gh-pitfall-scraper.git"
echo "git push -u origin main"
echo ""
echo "🎉 项目已准备就绪！"