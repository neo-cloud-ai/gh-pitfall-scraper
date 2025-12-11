# gh-pitfall-scraper Makefile

.PHONY: all build test clean run help

# 默认目标
all: build test

# 构建可执行文件
build:
	@echo "🔨 构建 gh-pitfall-scraper..."
	go build -o gh-pitfall-scraper main.go
	@echo "✅ 构建完成"

# 运行测试
test:
	@echo "🧪 运行单元测试..."
	go test ./internal/scraper/... -v
	@echo "✅ 测试完成"

# 格式化代码
fmt:
	@echo "📝 格式化 Go 代码..."
	go fmt ./...
	@echo "✅ 代码格式化完成"

# 代码检查
lint:
	@echo "🔍 代码质量检查..."
	go vet ./...
	@echo "✅ 代码检查完成"

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f gh-pitfall-scraper
	rm -rf output/
	@echo "✅ 清理完成"

# 运行程序
run: build
	@echo "🚀 运行 gh-pitfall-scraper..."
	@mkdir -p output
	./gh-pitfall-scraper

# 开发模式（监控文件变化并重新构建）
dev:
	@echo "👨‍💻 开发模式启动..."
	@echo "请使用第三方工具如 'go run main.go' 或配置文件监控工具"

# 安装依赖
deps:
	@echo "📦 下载项目依赖..."
	go mod download
	go mod tidy
	@echo "✅ 依赖安装完成"

# 创建示例配置
example-config:
	@echo "📋 创建示例配置文件..."
	@if [ ! -f config.yaml ]; then \
		cp config.yaml config.yaml.example; \
		echo "✅ 创建 config.yaml.example"; \
	else \
		echo "⚠️  config.yaml 已存在，跳过创建"; \
	fi

# 生成覆盖率报告
coverage:
	@echo "📊 生成测试覆盖率报告..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告生成完成: coverage.html"

# 检查配置
check-config:
	@echo "🔍 检查配置文件..."
	@if grep -q "ghp_xxx" config.yaml; then \
		echo "⚠️  请在 config.yaml 中设置您的真实 GitHub Token"; \
		echo "💡 访问 https://github.com/settings/tokens 创建 Personal Access Token"; \
	else \
		echo "✅ GitHub Token 已配置"; \
	fi

# 显示帮助信息
help:
	@echo "gh-pitfall-scraper Makefile"
	@echo "==========================="
	@echo ""
	@echo "可用命令:"
	@echo "  make build         - 构建可执行文件"
	@echo "  make test          - 运行单元测试"
	@echo "  make fmt           - 格式化代码"
	@echo "  make lint          - 代码质量检查"
	@echo "  make clean         - 清理构建文件"
	@echo "  make run           - 构建并运行程序"
	@echo "  make deps          - 安装项目依赖"
	@echo "  make example-config - 创建示例配置文件"
	@echo "  make coverage      - 生成测试覆盖率报告"
	@echo "  make check-config  - 检查配置文件"
	@echo "  make help          - 显示此帮助信息"
	@echo ""
	@echo "🚀 快速开始:"
	@echo "  1. make deps       # 安装依赖"
	@echo "  2. 编辑 config.yaml # 设置 GitHub Token"
	@echo "  3. make run        # 运行程序"

# 完整构建流程
ci: clean deps fmt lint test build
	@echo "🎉 CI 构建完成"