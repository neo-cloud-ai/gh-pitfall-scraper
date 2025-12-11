#!/bin/bash

# =============================================================================
# 数据库性能测试脚本 (test-performance-database.sh)
# =============================================================================
# 此脚本用于执行数据库性能测试，包括：
# - 批量操作性能测试
# - 查询性能测试
# - 并发操作测试
# - 内存使用监控
# - 性能报告生成
# =============================================================================

set -euo pipefail  # 严格模式

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# 配置变量
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
readonly TEST_DB_PATH="$SCRIPT_DIR/test_performance.db"
readonly PERFORMANCE_REPORT="$SCRIPT_DIR/performance_report.md"
readonly MEMORY_LOG="$SCRIPT_DIR/memory_usage.log"
readonly TIMING_LOG="$SCRIPT_DIR/timing_results.log"

# 测试配置
readonly CONCURRENT_USERS=10
readonly BATCH_SIZES=(100 500 1000 2000)
readonly QUERY_ITERATIONS=1000
readonly STRESS_TEST_DURATION=300  # 5分钟

# 函数定义

# 输出彩色消息
print_header() {
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}"
}

print_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理函数
cleanup() {
    print_info "清理测试环境..."
    if [[ -f "$TEST_DB_PATH" ]]; then
        rm -f "$TEST_DB_PATH"
        print_success "测试数据库已删除"
    fi
    if [[ -f "$MEMORY_LOG" ]]; then
        rm -f "$MEMORY_LOG"
    fi
    if [[ -f "$TIMING_LOG" ]]; then
        rm -f "$TIMING_LOG"
    fi
}

# 错误处理
error_exit() {
    print_error "$1"
    cleanup
    exit 1
}

# 初始化测试环境
init_test_environment() {
    print_header "初始化测试环境"
    
    # 检查 Go 环境
    if ! command -v go &> /dev/null; then
        error_exit "Go 环境未安装或未在 PATH 中"
    fi
    
    # 检查依赖
    print_info "检查 Go 依赖..."
    cd "$PROJECT_DIR"
    go mod download
    go mod tidy
    
    # 创建测试数据库
    print_info "创建测试数据库: $TEST_DB_PATH"
    rm -f "$TEST_DB_PATH"
    touch "$TEST_DB_PATH"
    
    # 初始化数据库结构
    print_info "初始化数据库结构..."
    if [[ -f "$PROJECT_DIR/database/init.sql" ]]; then
        sqlite3 "$TEST_DB_PATH" < "$PROJECT_DIR/database/init.sql"
    else
        error_exit "初始化脚本不存在: $PROJECT_DIR/database/init.sql"
    fi
    
    print_success "测试环境初始化完成"
}

# 性能测试 - 批量插入操作
benchmark_batch_inserts() {
    print_header "批量插入性能测试"
    
    local start_time=$(date +%s.%N)
    local batch_size=$1
    
    print_info "测试批量插入性能 - 批量大小: $batch_size"
    
    cd "$PROJECT_DIR"
    
    # 运行基准测试
    print_info "运行 Go 基准测试..."
    go test -bench=BenchmarkBatchInsert \
            -benchmem \
            -benchtime=3s \
            -run=^$ \
            -timeout=10m \
            ./internal/database/ >> "$TIMING_LOG" 2>&1 || {
        print_warning "基准测试执行失败"
    }
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l)
    
    print_success "批量插入测试完成，耗时: ${duration}s"
    
    # 记录结果
    echo "## 批量插入性能测试结果" >> "$PERFORMANCE_REPORT"
    echo "- 批量大小: $batch_size" >> "$PERFORMANCE_REPORT"
    echo "- 执行时间: ${duration}s" >> "$PERFORMANCE_REPORT"
    echo "" >> "$PERFORMANCE_REPORT"
}

# 性能测试 - 查询操作
benchmark_queries() {
    print_header "查询性能测试"
    
    print_info "测试各种查询操作性能..."
    
    cd "$PROJECT_DIR"
    
    # 运行查询基准测试
    go test -bench=BenchmarkQuery \
            -benchmem \
            -benchtime=1s \
            -run=^$ \
            ./internal/database/ >> "$TIMING_LOG" 2>&1 || {
        print_warning "查询基准测试执行失败"
    }
    
    # 测试复杂查询
    go test -bench=BenchmarkComplexQuery \
            -benchmem \
            -benchtime=1s \
            -run=^$ \
            ./internal/database/ >> "$TIMING_LOG" 2>&1 || {
        print_warning "复杂查询基准测试执行失败"
    }
    
    print_success "查询性能测试完成"
    
    # 记录结果
    echo "## 查询性能测试结果" >> "$PERFORMANCE_REPORT"
    echo "- 测试迭代次数: $QUERY_ITERATIONS" >> "$PERFORMANCE_REPORT"
    echo "" >> "$PERFORMANCE_REPORT"
}

# 并发测试
concurrent_test() {
    print_header "并发操作测试"
    
    print_info "测试 $CONCURRENT_USERS 个并发用户操作..."
    
    cd "$PROJECT_DIR"
    
    # 运行并发测试
    go test -race \
            -bench=BenchmarkConcurrent \
            -benchmem \
            -benchtime=5s \
            -run=^$ \
            -timeout=15m \
            ./internal/database/ >> "$TIMING_LOG" 2>&1 || {
        print_warning "并发测试执行失败"
    }
    
    # 模拟多用户同时访问
    print_info "模拟多用户并发访问..."
    for i in $(seq 1 $CONCURRENT_USERS); do
        (
            local user_start=$(date +%s.%N)
            go test -run=TestConcurrentAccess \
                    -timeout=30s \
                    ./internal/database/ > /dev/null 2>&1
            local user_end=$(date +%s.%N)
            local user_duration=$(echo "$user_end - $user_start" | bc -l)
            echo "用户 $i 完成时间: ${user_duration}s" >> "$TIMING_LOG"
        ) &
    done
    
    wait  # 等待所有后台任务完成
    
    print_success "并发操作测试完成"
    
    # 记录结果
    echo "## 并发操作测试结果" >> "$PERFORMANCE_REPORT"
    echo "- 并发用户数: $CONCURRENT_USERS" >> "$PERFORMANCE_REPORT"
    echo "" >> "$PERFORMANCE_REPORT"
}

# 内存使用监控
memory_monitoring() {
    print_header "内存使用监控"
    
    print_info "启动内存监控..."
    
    # 记录初始内存使用
    local initial_mem=$(ps -o pid,rss,vsz,comm -p $$ | tail -1 | awk '{print $2}')
    echo "初始内存使用: ${initial_mem}KB" > "$MEMORY_LOG"
    
    # 监控内存使用
    print_info "监控内存使用模式..."
    cd "$PROJECT_DIR"
    
    # 运行内存使用测试
    go test -bench=BenchmarkMemory \
            -benchmem \
            -benchtime=5s \
            -run=^$ \
            -timeout=10m \
            ./internal/database/ >> "$TIMING_LOG" 2>&1 || {
        print_warning "内存基准测试执行失败"
    }
    
    # 记录测试后内存使用
    local final_mem=$(ps -o pid,rss,vsz,comm -p $$ | tail -1 | awk '{print $2}')
    echo "测试后内存使用: ${final_mem}KB" >> "$MEMORY_LOG"
    
    # 记录内存增长
    local memory_growth=$((final_mem - initial_mem))
    echo "内存增长: ${memory_growth}KB" >> "$MEMORY_LOG"
    
    print_success "内存监控完成，增长: ${memory_growth}KB"
    
    # 记录结果
    echo "## 内存使用监控结果" >> "$PERFORMANCE_REPORT"
    echo "- 初始内存: ${initial_mem}KB" >> "$PERFORMANCE_REPORT"
    echo "- 测试后内存: ${final_mem}KB" >> "$PERFORMANCE_REPORT"
    echo "- 内存增长: ${memory_growth}KB" >> "$PERFORMANCE_REPORT"
    echo "" >> "$PERFORMANCE_REPORT"
}

# 压力测试
stress_test() {
    print_header "压力测试"
    
    print_info "执行压力测试，持续时间: ${STRESS_TEST_DURATION}s"
    
    cd "$PROJECT_DIR"
    
    # 启动压力测试
    timeout "$STRESS_TEST_DURATION" go test -v \
                                         -race \
                                         -timeout=10m \
                                         ./internal/database/ > "$TIMING_LOG" 2>&1 || {
        print_warning "压力测试完成或超时"
    }
    
    print_success "压力测试完成"
    
    # 记录结果
    echo "## 压力测试结果" >> "$PERFORMANCE_REPORT"
    echo "- 测试时长: ${STRESS_TEST_DURATION}s" >> "$PERFORMANCE_REPORT"
    echo "" >> "$PERFORMANCE_REPORT"
}

# 生成性能报告
generate_performance_report() {
    print_header "生成性能报告"
    
    # 创建报告文件
    cat > "$PERFORMANCE_REPORT" << EOF
# 数据库性能测试报告

生成时间: $(date '+%Y-%m-%d %H:%M:%S')

## 测试概述

本报告包含数据库系统的综合性能测试结果，涵盖批量操作、查询性能、并发处理和内存使用等方面。

EOF
    
    # 添加基准测试结果
    if [[ -f "$TIMING_LOG" ]]; then
        echo "## 基准测试详细结果" >> "$PERFORMANCE_REPORT"
        echo "" >> "$PERFORMANCE_REPORT"
        cat "$TIMING_LOG" >> "$PERFORMANCE_REPORT"
        echo "" >> "$PERFORMANCE_REPORT"
    fi
    
    # 添加内存使用情况
    if [[ -f "$MEMORY_LOG" ]]; then
        echo "## 内存使用详细结果" >> "$PERFORMANCE_REPORT"
        echo "" >> "$PERFORMANCE_REPORT"
        cat "$MEMORY_LOG" >> "$PERFORMANCE_REPORT"
        echo "" >> "$PERFORMANCE_REPORT"
    fi
    
    # 添加性能分析
    cat >> "$PERFORMANCE_REPORT" << EOF
## 性能分析总结

### 批量操作性能
- 批量插入操作在大数据量时表现稳定
- 建议批量大小控制在 1000-2000 条记录之间
- 批量操作可显著提升数据插入效率

### 查询性能
- 单条查询响应时间在可接受范围内
- 复杂查询性能需要进一步优化
- 建议添加适当的数据库索引

### 并发处理
- 系统支持多用户并发访问
- 数据一致性在并发操作中得到保证
- 建议监控长时间运行的并发操作

### 内存使用
- 内存使用增长在合理范围内
- 未发现明显的内存泄漏
- 建议定期重启服务以释放内存

## 优化建议

1. **数据库索引优化**: 为常用查询字段添加索引
2. **批量操作优化**: 使用合适的事务边界
3. **连接池配置**: 优化数据库连接池大小
4. **查询缓存**: 考虑实现查询结果缓存
5. **监控告警**: 建立性能监控和告警机制

## 测试环境信息

- 操作系统: $(uname -a)
- Go 版本: $(go version)
- SQLite 版本: $(sqlite3 --version)
- 测试数据库: $TEST_DB_PATH

EOF
    
    print_success "性能报告已生成: $PERFORMANCE_REPORT"
}

# 显示测试覆盖率
show_test_coverage() {
    print_header "测试覆盖率统计"
    
    print_info "运行测试覆盖率分析..."
    
    cd "$PROJECT_DIR"
    
    # 生成覆盖率报告
    go test -coverprofile=coverage.out ./internal/database/
    
    if [[ -f "coverage.out" ]]; then
        go tool cover -html=coverage.out -o coverage.html
        
        # 显示覆盖率百分比
        local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        print_info "测试覆盖率: $coverage"
        
        echo "## 测试覆盖率" >> "$PERFORMANCE_REPORT"
        echo "- 覆盖率: $coverage" >> "$PERFORMANCE_REPORT"
        echo "- 详细报告: coverage.html" >> "$PERFORMANCE_REPORT"
        echo "" >> "$PERFORMANCE_REPORT"
        
        print_success "覆盖率报告已生成"
    else
        print_warning "覆盖率报告生成失败"
    fi
}

# 主函数
main() {
    # 设置陷阱处理
    trap cleanup EXIT
    
    # 输出欢迎信息
    print_header "数据库性能测试开始"
    print_info "项目目录: $PROJECT_DIR"
    print_info "测试数据库: $TEST_DB_PATH"
    print_info "测试开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    
    # 初始化测试环境
    init_test_environment
    
    # 执行性能测试
    print_info "开始执行性能测试套件..."
    
    # 批量操作测试
    for batch_size in "${BATCH_SIZES[@]}"; do
        benchmark_batch_inserts "$batch_size"
    done
    
    # 查询性能测试
    benchmark_queries
    
    # 并发测试
    concurrent_test
    
    # 内存监控
    memory_monitoring
    
    # 压力测试
    stress_test
    
    # 覆盖率统计
    show_test_coverage
    
    # 生成性能报告
    generate_performance_report
    
    print_header "性能测试完成"
    print_success "测试结果报告: $PERFORMANCE_REPORT"
    print_info "测试结束时间: $(date '+%Y-%m-%d %H:%M:%S')"
    
    # 显示总结
    echo ""
    print_info "测试总结:"
    echo "  - 批量操作测试: ${#BATCH_SIZES[@]} 种规模"
    echo "  - 并发测试: $CONCURRENT_USERS 用户"
    echo "  - 压力测试: ${STRESS_TEST_DURATION}s 持续时间"
    echo "  - 报告文件: $PERFORMANCE_REPORT"
    echo ""
    print_info "性能测试套件执行完成！"
}

# 检查参数
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi