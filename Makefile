# gh-pitfall-scraper Makefile

.PHONY: all build test clean run help db-init db-backup db-restore db-clean db-migrate db-test-perf db-stats

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
	@echo "  构建和测试:"
	@echo "    make build         - 构建可执行文件"
	@echo "    make test          - 运行单元测试"
	@echo "    make fmt           - 格式化代码"
	@echo "    make lint          - 代码质量检查"
	@echo "    make clean         - 清理构建文件"
	@echo "    make run           - 构建并运行程序"
	@echo "    make deps          - 安装项目依赖"
	@echo ""
	@echo "  数据库操作:"
	@echo "    make db-init       - 初始化数据库"
	@echo "    make db-backup     - 创建数据库备份"
	@echo "    make db-restore    - 从备份恢复数据库"
	@echo "    make db-clean      - 清理数据库数据"
	@echo "    make db-migrate    - 运行数据库迁移"
	@echo "    make db-test-perf  - 数据库性能测试"
	@echo "    make db-stats      - 显示数据库统计"
	@echo "    make db-maintain   - 运行数据库维护"
	@echo "    make db-health     - 检查数据库健康"
	@echo "    make db-reset      - 重置数据库（危险）"
	@echo ""
	@echo "  工具和配置:"
	@echo "    make example-config - 创建示例配置文件"
	@echo "    make coverage      - 生成测试覆盖率报告"
	@echo "    make check-config  - 检查配置文件"
	@echo "    make help          - 显示此帮助信息"
	@echo ""
	@echo "🚀 快速开始:"
	@echo "  1. make deps       # 安装依赖"
	@echo "  2. make db-init    # 初始化数据库"
	@echo "  3. 编辑 config.yaml # 设置 GitHub Token"
	@echo "  4. make run        # 运行程序"
	@echo ""
	@echo "💾 数据库管理:"
	@echo "  - 数据文件: data/gh-pitfall-scraper.db"
	@echo "  - 备份目录: backups/"
	@echo "  - 配置示例: config-database-example.yaml"

# =====================
# 数据库相关目标
# =====================

# 初始化数据库
db-init:
	@echo "🗄️ 初始化数据库..."
	@mkdir -p data backups
	@if [ ! -f "data/gh-pitfall-scraper.db" ]; then \
		go run -tags=scripts ./cmd/initdb.go; \
		echo "✅ 数据库初始化完成"; \
	else \
		echo "⚠️ 数据库已存在，跳过初始化"; \
	fi

# 数据库备份
db-backup:
	@echo "💾 创建数据库备份..."
	@mkdir -p backups
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		timestamp=$$(date +"%Y%m%d_%H%M%S"); \
		cp "data/gh-pitfall-scraper.db" "backups/gh-pitfall-scraper_$${timestamp}.db"; \
		echo "✅ 备份完成: backups/gh-pitfall-scraper_$${timestamp}.db"; \
	else \
		echo "❌ 数据库文件不存在，无法备份"; \
	fi

# 数据库恢复
db-restore:
	@echo "🔄 从备份恢复数据库..."
	@echo "可用备份文件:"
	@ls -la backups/*.db 2>/dev/null || echo "  未找到备份文件"
	@echo ""
	@read -p "请输入要恢复的备份文件名: " backup_file; \
	if [ -f "backups/$$backup_file" ]; then \
		cp "backups/$$backup_file" "data/gh-pitfall-scraper.db"; \
		echo "✅ 数据库恢复完成"; \
	else \
		echo "❌ 备份文件不存在"; \
	fi

# 清理数据库
db-clean:
	@echo "🧹 清理数据库数据..."
	@echo "⚠️ 这将删除所有数据，确定要继续吗？ (y/N)"
	@read -r confirm && [ "$$confirm" = "y" ]
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		rm "data/gh-pitfall-scraper.db"; \
		echo "✅ 数据库已清理"; \
		echo "💡 运行 'make db-init' 重新初始化数据库"; \
	else \
		echo "⚠️ 数据库文件不存在"; \
	fi

# 数据库迁移
db-migrate:
	@echo "🔄 运行数据库迁移..."
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		go run -tags=scripts ./cmd/migrate.go; \
		echo "✅ 迁移完成"; \
	else \
		echo "❌ 数据库未初始化，请先运行 'make db-init'"; \
	fi

# 数据库性能测试
db-test-perf:
	@echo "⚡ 运行数据库性能测试..."
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		go test -run=TestDatabasePerformance ./internal/database/... -v; \
	else \
		echo "❌ 数据库未初始化，请先运行 'make db-init'"; \
	fi

# 显示数据库统计信息
db-stats:
	@echo "📊 数据库统计信息:"
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		go run -tags=scripts ./cmd/db-stats.go; \
	else \
		echo "❌ 数据库未初始化，请先运行 'make db-init'"; \
	fi

# 数据库维护
db-maintain:
	@echo "🔧 运行数据库维护..."
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		go run -tags=scripts ./cmd/maintenance.go; \
		echo "✅ 维护完成"; \
	else \
		echo "❌ 数据库未初始化，请先运行 'make db-init'"; \
	fi

# 数据库重置（危险操作）
db-reset: db-clean db-init
	@echo "🔄 数据库重置完成"

# 检查数据库健康状态
db-health:
	@echo "🏥 检查数据库健康状态..."
	@if [ -f "data/gh-pitfall-scraper.db" ]; then \
		go run -tags=scripts ./cmd/health-check.go; \
	else \
		echo "❌ 数据库文件不存在"; \
	fi

# 完整构建流程
ci: clean deps fmt lint test build
	@echo "🎉 CI 构建完成"