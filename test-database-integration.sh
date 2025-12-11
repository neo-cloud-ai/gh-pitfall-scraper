#!/bin/bash
# gh-pitfall-scraper 数据库集成测试脚本

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
GO_MOD_FILE="$PROJECT_DIR/go.mod"
MAIN_GO="$PROJECT_DIR/main.go"
TEST_DB_NAME="gh_pitfall_scraper_test"
TEST_CONFIG="$PROJECT_DIR/config_test.yaml"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
check_go_environment() {
    log_info "检查Go环境..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请安装Go 1.19或更高版本"
        return 1
    fi
    
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
    log_success "Go版本: $GO_VERSION"
    
    # 检查模块文件
    if [ ! -f "$GO_MOD_FILE" ]; then
        log_error "go.mod文件不存在: $GO_MOD_FILE"
        return 1
    fi
    
    log_success "Go环境检查通过"
    return 0
}

# 检查PostgreSQL
check_postgresql() {
    log_info "检查PostgreSQL..."
    
    if ! command -v psql &> /dev/null; then
        log_warning "PostgreSQL客户端未安装，测试将跳过数据库相关功能"
        return 1
    fi
    
    # 测试连接
    if ! psql -h localhost -U postgres -c "SELECT 1;" &> /dev/null; then
        log_warning "无法连接到PostgreSQL，测试将跳过数据库相关功能"
        return 1
    fi
    
    log_success "PostgreSQL连接正常"
    return 0
}

# 创建测试配置
create_test_config() {
    log_info "创建测试配置..."
    
    cat > "$TEST_CONFIG" << 'EOF'
# 测试配置文件
app:
  name: "gh-pitfall-scraper-test"
  version: "2.0.0"
  log_level: "debug"
  data_dir: "./test_data"
  output_dir: "./test_output"
  max_workers: 2
  worker_queue: 50

github_token: "test_token"
request_interval: 1000
timeout: 10

repos:
  - owner: test-owner
    name: test-repo

keywords:
  - "test"
  - "example"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: ""
  dbname: "gh_pitfall_scraper_test"
  sslmode: "disable"
  max_open_conns: 5
  max_idle_conns: 2
  conn_max_lifetime: "60s"
  conn_max_idle_time: "30s"
  cache_enabled: false
  cache_size: 100
  cache_ttl: "300s"
  auto_cleanup_enabled: false
  cleanup_interval: "1h"
  data_retention: "24h"
  backup_enabled: false
  backup_interval: "1h"
  backup_path: "./test_backups"
  retention_days: 1
EOF
    
    log_success "测试配置已创建: $TEST_CONFIG"
}

# 创建测试数据库
create_test_database() {
    log_info "创建测试数据库..."
    
    # 删除现有测试数据库
    psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" &> /dev/null || true
    
    # 创建新测试数据库
    if ! psql -h localhost -U postgres -c "CREATE DATABASE $TEST_DB_NAME;" &> /dev/null; then
        log_error "创建测试数据库失败"
        return 1
    fi
    
    log_success "测试数据库已创建: $TEST_DB_NAME"
    return 0
}

# 清理测试数据库
cleanup_test_database() {
    log_info "清理测试数据库..."
    
    psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" &> /dev/null || true
    
    # 清理测试文件
    rm -f "$TEST_CONFIG"
    rm -rf "$PROJECT_DIR/test_data"
    rm -rf "$PROJECT_DIR/test_output"
    rm -rf "$PROJECT_DIR/test_backups"
    
    log_success "测试环境已清理"
}

# 测试代码编译
test_compilation() {
    log_info "测试代码编译..."
    
    cd "$PROJECT_DIR"
    
    # 下载依赖
    log_info "下载Go依赖..."
    if ! go mod download; then
        log_error "下载依赖失败"
        return 1
    fi
    
    # 编译主程序
    log_info "编译主程序..."
    if ! go build -o gh-pitfall-scraper main.go; then
        log_error "编译失败"
        return 1
    fi
    
    log_success "编译成功"
    return 0
}

# 测试命令行参数
test_command_line_args() {
    log_info "测试命令行参数..."
    
    cd "$PROJECT_DIR"
    
    # 测试帮助信息
    log_info "测试 --help 参数..."
    if ! ./gh-pitfall-scraper --help; then
        log_error "--help 测试失败"
        return 1
    fi
    
    # 测试版本信息
    log_info "测试 --version 参数..."
    if ! ./gh-pitfall-scraper --version; then
        log_error "--version 测试失败"
        return 1
    fi
    
    log_success "命令行参数测试通过"
    return 0
}

# 测试数据库初始化
test_database_init() {
    log_info "测试数据库初始化..."
    
    cd "$PROJECT_DIR"
    
    # 测试数据库初始化（仅初始化，不执行爬虫）
    log_info "执行 --db-only 参数测试..."
    export DB_NAME="$TEST_DB_NAME"
    
    if ! ./gh-pitfall-scraper --config "$TEST_CONFIG" --db-only; then
        log_error "数据库初始化测试失败"
        return 1
    fi
    
    log_success "数据库初始化测试通过"
    return 0
}

# 测试数据库健康检查
test_database_health() {
    log_info "测试数据库健康检查..."
    
    cd "$PROJECT_DIR"
    
    export DB_NAME="$TEST_DB_NAME"
    
    if ! ./gh-pitfall-scraper --config "$TEST_CONFIG" --health; then
        log_error "数据库健康检查测试失败"
        return 1
    fi
    
    log_success "数据库健康检查测试通过"
    return 0
}

# 测试数据库统计信息
test_database_stats() {
    log_info "测试数据库统计信息..."
    
    cd "$PROJECT_DIR"
    
    export DB_NAME="$TEST_DB_NAME"
    
    if ! ./gh-pitfall-scraper --config "$TEST_CONFIG" --stats; then
        log_error "数据库统计信息测试失败"
        return 1
    fi
    
    log_success "数据库统计信息测试通过"
    return 0
}

# 运行所有测试
run_all_tests() {
    local tests_passed=0
    local tests_total=0
    
    # 测试列表
    local test_functions=(
        "check_go_environment"
        "test_compilation"
        "test_command_line_args"
    )
    
    # 数据库相关测试（需要PostgreSQL）
    if check_postgresql; then
        test_functions+=(
            "create_test_config"
            "create_test_database"
            "test_database_init"
            "test_database_health"
            "test_database_stats"
        )
    fi
    
    # 执行测试
    for test_func in "${test_functions[@]}"; do
        tests_total=$((tests_total + 1))
        log_info "执行测试: $test_func"
        
        if $test_func; then
            tests_passed=$((tests_passed + 1))
            log_success "测试通过: $test_func"
        else
            log_error "测试失败: $test_func"
        fi
        echo
    done
    
    # 清理测试环境
    if check_postgresql; then
        cleanup_test_database
    fi
    
    # 输出测试结果
    echo "=== 测试结果 ==="
    echo "总测试数: $tests_total"
    echo "通过测试: $tests_passed"
    echo "失败测试: $((tests_total - tests_passed))"
    
    if [ $tests_passed -eq $tests_total ]; then
        log_success "所有测试通过！"
        return 0
    else
        log_error "部分测试失败"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
gh-pitfall-scraper 数据库集成测试脚本

用法: $0 [选项] [测试名称]

测试名称:
  all              运行所有测试（默认）
  go               测试Go环境
  compile          测试代码编译
  args             测试命令行参数
  db-init          测试数据库初始化
  db-health        测试数据库健康检查
  db-stats         测试数据库统计信息
  cleanup          清理测试环境

选项:
  --help           显示此帮助信息

示例:
  $0               # 运行所有测试
  $0 compile       # 仅测试代码编译
  $0 db-init       # 仅测试数据库初始化

EOF
}

# 主函数
main() {
    # 解析参数
    test_name="all"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                test_name="$1"
                shift
                ;;
        esac
    done
    
    # 设置错误处理
    trap 'log_error "测试过程中发生错误"' ERR
    
    # 执行测试
    case $test_name in
        all)
            run_all_tests
            ;;
        go)
            check_go_environment
            ;;
        compile)
            check_go_environment && test_compilation
            ;;
        args)
            check_go_environment && test_compilation && test_command_line_args
            ;;
        db-init)
            check_postgresql && create_test_config && create_test_database && test_database_init
            ;;
        db-health)
            check_postgresql && create_test_config && create_test_database && test_database_health
            ;;
        db-stats)
            check_postgresql && create_test_config && create_test_database && test_database_stats
            ;;
        cleanup)
            cleanup_test_database
            ;;
        *)
            log_error "未知测试: $test_name"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi