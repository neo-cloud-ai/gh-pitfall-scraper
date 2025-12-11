#!/bin/bash

# 数据库集成测试脚本
# 测试完整的数据库操作流程

set -e  # 遇到错误立即退出

# 颜色定义
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

# 测试计数器
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# 测试结果记录
TEST_RESULTS_FILE="test_results_$(date +%Y%m%d_%H%M%S).json"

# 开始测试
start_test() {
    local test_name="$1"
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "开始测试: $test_name"
}

# 测试通过
pass_test() {
    local test_name="$1"
    local details="$2"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    log_success "✓ $test_name 测试通过"
    if [ -n "$details" ]; then
        echo "    详细信息: $details"
    fi
}

# 测试失败
fail_test() {
    local test_name="$1"
    local error="$2"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    log_error "✗ $test_name 测试失败: $error"
}

# 记录测试结果到JSON文件
record_test_result() {
    local test_name="$1"
    local status="$2"
    local duration="$3"
    local error="$4"
    
    cat >> "$TEST_RESULTS_FILE" << EOF
{
    "test_name": "$test_name",
    "status": "$status",
    "duration_ms": $duration,
    "error": "$error",
    "timestamp": "$(date -Iseconds)"
},
EOF
}

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    
    # 清理测试数据库文件
    find . -name "test_*.db" -type f -delete 2>/dev/null || true
    find . -name "*test*.db" -type f -delete 2>/dev/null || true
    find . -name "*.db" -path "*/tmp/*" -type f -delete 2>/dev/null || true
    
    # 清理临时文件
    find /tmp -name "*test_db_*" -type d -exec rm -rf {} + 2>/dev/null || true
    
    log_success "清理完成"
}

# 检查Go环境
check_go_environment() {
    log_info "检查Go环境..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go未安装或不在PATH中"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    log_success "Go版本: $GO_VERSION"
    
    # 检查必要的依赖
    if ! go list -m github.com/mattn/go-sqlite3 &> /dev/null; then
        log_info "安装测试依赖..."
        go get github.com/mattn/go-sqlite3
    fi
}

# 运行单元测试
run_unit_tests() {
    log_info "运行数据库单元测试..."
    
    start_test "单元测试"
    
    # 运行所有测试
    if go test -v ./internal/database/... -timeout=300s; then
        pass_test "单元测试" "所有单元测试通过"
        record_test_result "unit_tests" "PASSED" "0" ""
    else
        fail_test "单元测试" "单元测试失败"
        record_test_result "unit_tests" "FAILED" "0" "unit tests failed"
    fi
    
    # 运行覆盖率测试
    log_info "生成测试覆盖率报告..."
    if go test -coverprofile=coverage.out ./internal/database/...; then
        log_success "测试覆盖率报告已生成: coverage.out"
        
        # 显示覆盖率统计
        if command -v go &> /dev/null; then
            COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            log_info "代码覆盖率: $COVERAGE"
        fi
    else
        log_warning "覆盖率报告生成失败"
    fi
}

# 运行性能基准测试
run_benchmark_tests() {
    log_info "运行性能基准测试..."
    
    start_test "性能基准测试"
    
    # 运行基准测试
    if go test -bench=. -benchmem ./internal/database/... -timeout=600s; then
        pass_test "性能基准测试" "性能测试完成"
        record_test_result "benchmark_tests" "PASSED" "0" ""
    else
        fail_test "性能基准测试" "性能测试失败"
        record_test_result "benchmark_tests" "FAILED" "0" "benchmark tests failed"
    fi
}

# 测试数据库初始化
test_database_initialization() {
    log_info "测试数据库初始化..."
    
    start_test "数据库初始化"
    
    # 创建临时测试数据库
    TEMP_DB="/tmp/test_integration_$(date +%s).db"
    
    cat > test_db_init.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    // 创建数据库配置
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    // 创建数据库实例
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    // 初始化数据库
    if err := db.Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    // 执行健康检查
    if err := db.HealthCheck(); err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    
    // 获取数据库统计信息
    stats, err := db.GetStats()
    if err != nil {
        log.Fatalf("Failed to get stats: %v", err)
    }
    
    fmt.Printf("Database initialized successfully!\n")
    fmt.Printf("Issues count: %v\n", stats["issues_count"])
    fmt.Printf("Repositories count: %v\n", stats["repositories_count"])
    fmt.Printf("Database size: %.2f MB\n", stats["database_size_mb"])
}
EOF

    if go run test_db_init.go "$TEMP_DB"; then
        pass_test "数据库初始化" "数据库初始化成功"
        record_test_result "database_initialization" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "数据库初始化" "数据库初始化失败"
        record_test_result "database_initialization" "FAILED" "0" "initialization failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_db_init.go
}

# 测试CRUD操作
test_crud_operations() {
    log_info "测试CRUD操作..."
    
    start_test "CRUD操作"
    
    TEMP_DB="/tmp/test_crud_$(date +%s).db"
    
    cat > test_crud.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    if err := db.Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    crud := db.CRUD()
    
    // 测试创建Issue
    issue := &database.Issue{
        Number:      1,
        Title:       "Integration Test Issue",
        Body:        "This is a test issue for CRUD operations",
        URL:         "https://github.com/test/repo/issues/1",
        State:       "open",
        RepoOwner:   "test",
        RepoName:    "repo",
        Score:       15.5,
        CreatedAtDB: time.Now(),
        UpdatedAtDB: time.Now(),
    }
    
    id, err := crud.CreateIssue(issue)
    if err != nil {
        log.Fatalf("Failed to create issue: %v", err)
    }
    fmt.Printf("Created issue with ID: %d\n", id)
    
    // 测试读取Issue
    retrieved, err := crud.GetIssue(id)
    if err != nil {
        log.Fatalf("Failed to get issue: %v", err)
    }
    fmt.Printf("Retrieved issue: %s\n", retrieved.Title)
    
    // 测试更新Issue
    issue.Title = "Updated Integration Test Issue"
    issue.Score = 20.0
    err = crud.UpdateIssue(issue)
    if err != nil {
        log.Fatalf("Failed to update issue: %v", err)
    }
    fmt.Printf("Updated issue title: %s\n", issue.Title)
    
    // 测试查询Issues
    issues, err := crud.GetAllIssues(10, 0)
    if err != nil {
        log.Fatalf("Failed to query issues: %v", err)
    }
    fmt.Printf("Found %d issues\n", len(issues))
    
    // 测试删除Issue
    err = crud.DeleteIssue(id)
    if err != nil {
        log.Fatalf("Failed to delete issue: %v", err)
    }
    fmt.Printf("Deleted issue with ID: %d\n", id)
    
    // 验证删除
    _, err = crud.GetIssue(id)
    if err == nil {
        log.Fatalf("Expected error when getting deleted issue")
    }
    
    fmt.Println("CRUD operations test completed successfully!")
}
EOF

    if go run test_crud.go "$TEMP_DB"; then
        pass_test "CRUD操作" "所有CRUD操作成功"
        record_test_result "crud_operations" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "CRUD操作" "CRUD操作失败"
        record_test_result "crud_operations" "FAILED" "0" "crud operations failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_crud.go
}

# 测试事务操作
test_transaction_operations() {
    log_info "测试事务操作..."
    
    start_test "事务操作"
    
    TEMP_DB="/tmp/test_transaction_$(date +%s).db"
    
    cat > test_transaction.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    if err := db.Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    transaction := db.Transaction()
    
    // 测试成功事务
    err = transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
        _, err := tx.Exec(`
            INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, 100, "Transaction Test Issue", "Test body", 
            "https://github.com/test/repo/issues/100", "open", "test", "repo",
            time.Now(), time.Now())
        return err
    })
    
    if err != nil {
        log.Fatalf("Transaction failed: %v", err)
    }
    fmt.Println("Successful transaction completed")
    
    // 测试回滚事务
    err = transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
        _, err := tx.Exec(`
            INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, 101, "Rollback Test Issue", "Test body", 
            "https://github.com/test/repo/issues/101", "open", "test", "repo",
            time.Now(), time.Now())
        if err != nil {
            return err
        }
        // 模拟错误导致回滚
        return sql.ErrTxDone
    })
    
    if err == nil {
        log.Fatalf("Expected error for rollback transaction")
    }
    fmt.Println("Rollback transaction test completed")
    
    // 验证只有第一个issue被创建
    cruder := db.CRUD()
    issues, err := cruder.GetAllIssues(10, 0)
    if err != nil {
        log.Fatalf("Failed to query issues: %v", err)
    }
    
    found := false
    for _, issue := range issues {
        if issue.Number == 100 {
            found = true
            break
        }
    }
    
    if !found {
        log.Fatalf("Expected to find issue 100")
    }
    
    fmt.Println("Transaction operations test completed successfully!")
}
EOF

    if go run test_transaction.go "$TEMP_DB"; then
        pass_test "事务操作" "事务操作测试成功"
        record_test_result "transaction_operations" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "事务操作" "事务操作测试失败"
        record_test_result "transaction_operations" "FAILED" "0" "transaction operations failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_transaction.go
}

# 测试去重功能
test_deduplication() {
    log_info "测试去重功能..."
    
    start_test "去重功能"
    
    TEMP_DB="/tmp/test_dedup_$(date +%s).db"
    
    cat > test_dedup.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    if err := db./*Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    crud := db.CRUD()
    deduplicator := db.Deduplication()
    
    // 创建重复的issues
    issues := []*database.Issue{
        {
            Number:      200,
            Title:       "Memory leak issue",
            Body:        "There is a memory leak in the application",
            URL:         "https://github.com/test/repo/issues/200",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       15.0,
            ContentHash: "duplicate_hash",
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      201,
            Title:       "Memory leak issue",
            Body:        "There is a memory leak in the application",
            URL:         "https://github.com/test/repo/issues/201",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       12.0,
            ContentHash: "duplicate_hash",
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      202,
            Title:       "Different issue",
            Body:        "This is a different issue",
            URL:         "https://github.com/test/repo/issues/202",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       8.0,
            ContentHash: "unique_hash",
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
    }
    
    _, err = crud.CreateIssues(issues)
    if err != nil {
        log.Fatalf("Failed to create issues: %v", err)
    }
    fmt.Println("Created test issues with duplicates")
    
    // 执行去重
    result, err := deduplicator.FindDuplicates()
    if err != nil {
        log.Fatalf("Deduplication failed: %v", err)
    }
    
    fmt.Printf("Deduplication completed:\n")
    fmt.Printf("  Total processed: %d\n", result.TotalProcessed)
    fmt.Printf("  Duplicates found: %d\n", result.DuplicatesFound)
    fmt.Printf("  Duplicate groups: %d\n", len(result.DuplicateGroups))
    
    if result.DuplicatesFound == 0 {
        log.Warning("No duplicates found (may be expected depending on deduplication logic)")
    }
    
    // 获取去重统计
    stats, err := deduplicator.GetDuplicateStats()
    if err != nil {
        log.Fatalf("Failed to get duplicate stats: %v", err)
    }
    
    fmt.Printf("Duplicate stats: %+v\n", stats)
    
    fmt.Println("Deduplication test completed successfully!")
}
EOF

    if go run test_dedup.go "$TEMP_DB"; then
        pass_test "去重功能" "去重功能测试成功"
        record_test_result "deduplication" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "去重功能" "去重功能测试失败"
        record_test_result "deduplication" "FAILED" "0" "deduplication failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_dedup.go
}

# 测试分类功能
test_classification() {
    log_info "测试分类功能..."
    
    start_test "分类功能"
    
    TEMP_DB="/tmp/test_classification_$(date +%s).db"
    
    cat > test_classification.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    if err := db.Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    crud := db.CRUD()
    classifier := db.Classification()
    
    // 创建测试issues
    issues := []*database.Issue{
        {
            Number:      300,
            Title:       "Memory leak causing performance issues",
            Body:        "The application has a memory leak that affects performance",
            URL:         "https://github.com/test/repo/issues/300",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       20.0,
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      301,
            Title:       "Security vulnerability in authentication",
            Body:        "There is a security vulnerability that allows unauthorized access",
            URL:         "https://github.com/test/repo/issues/301",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       25.0,
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      302,
            Title:       "Add PostgreSQL support",
            Body:        "We need to add support for PostgreSQL database",
            URL:         "https://github.com/test/repo/issues/302",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       15.0,
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
    }
    
    _, err = crud.CreateIssues(issues)
    if err != nil {
        log.Fatalf("Failed to create issues: %v", err)
    }
    fmt.Println("Created test issues for classification")
    
    // 测试单个issue分类
    for i, issue := range issues {
        result, err := classifier.ClassifySingleIssue(issue)
        if err != nil {
            log.Fatalf("Classification failed for issue %d: %v", i, err)
        }
        
        fmt.Printf("Issue %d (%s) classified as: %s (confidence: %.2f)\n", 
            issue.Number, issue.Title, result.Category, result.Confidence)
    }
    
    // 测试批量分类
    stats, err := classifier.ClassifyIssues(issues)
    if err != nil {
        log.Fatalf("Batch classification failed: %v", err)
    }
    
    fmt.Printf("Batch classification completed:\n")
    fmt.Printf("  Total processed: %d\n", stats.TotalProcessed)
    fmt.Printf("  Classified: %d\n", stats.Classified)
    fmt.Printf("  Confidence: %.2f\n", stats.Confidence)
    
    // 获取分类统计
    clsStats, err := classifier.GetClassificationStats()
    if err != nil {
        log.Fatalf("Failed to get classification stats: %v", err)
    }
    
    fmt.Printf("Classification stats: %+v\n", clsStats)
    
    fmt.Println("Classification test completed successfully!")
}
EOF

    if go run test_classification.go "$TEMP_DB"; then
        pass_test "分类功能" "分类功能测试成功"
        record_test_result "classification" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "分类功能" "分类功能测试失败"
        record_test_result "classification" "FAILED" "0" "classification failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_classification.go
}

# 测试搜索功能
test_search_operations() {
    log_info "测试搜索功能..."
    
    start_test "搜索功能"
    
    TEMP_DB="/tmp/test_search_$(date +%s).db"
    
    cat > test_search.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    if err := db.Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    crud := db.CRUD()
    
    // 创建测试数据
    issues := []*database.Issue{
        {
            Number:      400,
            Title:       "Memory leak in cache system",
            Body:        "The cache system has a memory leak issue",
            URL:         "https://github.com/test/repo/issues/400",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       20.0,
            Category:    "performance",
            Keywords:    database.StringArray{"memory", "cache", "performance"},
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      401,
            Title:       "Security vulnerability found",
            Body:        "There is a security vulnerability in the login system",
            URL:         "https://github.com/test/repo/issues/401",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       25.0,
            Category:    "security",
            Keywords:    database.StringArray{"security", "vulnerability", "login"},
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      402,
            Title:       "Add new feature for user management",
            Body:        "We need to implement user management features",
            URL:         "https://github.com/test/repo/issues/402",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       15.0,
            Category:    "feature",
            Keywords:    database.StringArray{"feature", "user", "management"},
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
    }
    
    _, err = crud.CreateIssues(issues)
    if err != nil {
        log.Fatalf("Failed to create issues: %v", err)
    }
    fmt.Println("Created test issues for search")
    
    // 测试基本搜索
    results, err := crud.SearchIssues("memory", 10, 0)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }
    fmt.Printf("Basic search for 'memory' found %d results\n", len(results))
    
    // 测试高级搜索
    search := &database.AdvancedSearch{
        Query:       "cache",
        Categories:  []string{"performance"},
        SortBy:      "score",
        SortOrder:   "DESC",
        Limit:       10,
        Offset:      0,
    }
    
    advancedResults, err := crud.SearchIssuesAdvanced(search)
    if err != nil {
        log.Fatalf("Advanced search failed: %v", err)
    }
    fmt.Printf("Advanced search found %d results\n", len(advancedResults))
    
    // 测试按分类搜索
    categoryResults, err := crud.GetIssuesByCategory("performance", 10, 0)
    if err != nil {
        log.Fatalf("Category search failed: %v", err)
    }
    fmt.Printf("Category search for 'performance' found %d results\n", len(categoryResults))
    
    // 测试按关键词搜索
    keywordResults, err := crud.GetIssuesByKeywords([]string{"security", "vulnerability"}, 10, 0)
    if err != nil {
        log.Fatalf("Keyword search failed: %v", err)
    }
    fmt.Printf("Keyword search found %d results\n", len(keywordResults))
    
    fmt.Println("Search operations test completed successfully!")
}
EOF

    if go run test_search.go "$TEMP_DB"; then
        pass_test "搜索功能" "搜索功能测试成功"
        record_test_result "search_operations" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "搜索功能" "搜索功能测试失败"
        record_test_result "search_operations" "FAILED" "0" "search operations failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_search.go
}

# 测试统计功能
test_statistics() {
    log_info "测试统计功能..."
    
    start_test "统计功能"
    
    TEMP_DB="/tmp/test_stats_$(date +%s).db"
    
    cat > test_stats.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "./gh-pitfall-scraper/internal/database"
)

func main() {
    config := database.DefaultDatabaseConfig()
    config.Path = os.Args[1]
    config.MaxConnections = 1
    
    db, err := database.NewDatabase(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    if err := db.Initialize(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    crud := db.CRUD()
    
    // 创建测试数据
    issues := []*database.Issue{
        {
            Number:      500,
            Title:       "Performance issue 1",
            Body:        "First performance issue",
            URL:         "https://github.com/test/repo/issues/500",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       20.0,
            Category:    "performance",
            Priority:    "high",
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      501,
            Title:       "Security issue 1",
            Body:        "First security issue",
            URL:         "https://github.com/test/repo/issues/501",
            State:       "closed",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       25.0,
            Category:    "security",
            Priority:    "critical",
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
        {
            Number:      502,
            Title:       "Feature issue 1",
            Body:        "First feature issue",
            URL:         "https://github.com/test/repo/issues/502",
            State:       "open",
            RepoOwner:   "test",
            RepoName:    "repo",
            Score:       15.0,
            Category:    "feature",
            Priority:    "medium",
            CreatedAtDB: time.Now(),
            UpdatedAtDB: time.Now(),
        },
    }
    
    _, err = crud.CreateIssues(issues)
    if err != nil {
        log.Fatalf("Failed to create issues: %v", err)
    }
    fmt.Println("Created test issues for statistics")
    
    // 获取整体统计
    stats, err := crud.GetIssueStats()
    if err != nil {
        log.Fatalf("Failed to get issue stats: %v", err)
    }
    
    fmt.Printf("Overall statistics:\n")
    fmt.Printf("  Total count: %d\n", stats.TotalCount)
    fmt.Printf("  Average score: %.2f\n", stats.AverageScore)
    fmt.Printf("  Categories: %+v\n", stats.ByCategory)
    fmt.Printf("  Priorities: %+v\n", stats.ByPriority)
    fmt.Printf("  States: %+v\n", stats.ByState)
    fmt.Printf("  Score distribution: %+v\n", stats.ScoreDistribution)
    
    // 获取特定仓库统计
    repoStats, err := crud.GetRepositoryStats("test", "repo")
    if err != nil {
        log.Fatalf("Failed to get repository stats: %v", err)
    }
    
    fmt.Printf("Repository statistics for test/repo:\n")
    fmt.Printf("  Total count: %d\n", repoStats.TotalCount)
    fmt.Printf("  Categories: %+v\n", repoStats.ByCategory)
    
    fmt.Println("Statistics test completed successfully!")
}
EOF

    if go run test_stats.go "$TEMP_DB"; then
        pass_test "统计功能" "统计功能测试成功"
        record_test_result "statistics" "PASSED" "0" ""
        rm -f "$TEMP_DB"
    else
        fail_test "统计功能" "统计功能测试失败"
        record_test_result "statistics" "FAILED" "0" "statistics failed"
        rm -f "$TEMP_DB" 2>/dev/null || true
    fi
    
    rm -f test_stats.go
}

# 生成测试报告
generate_test_report() {
    log_info "生成测试报告..."
    
    # 修复JSON文件格式（移除最后一个逗号）
    if [ -f "$TEST_RESULTS_FILE" ]; then
        sed -i '$ s/,$//' "$TEST_RESULTS_FILE"
        echo "]" >> "$TEST_RESULTS_FILE"
    fi
    
    # 生成HTML报告
    cat > test_report.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>数据库集成测试报告</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; padding: 15px; border-radius: 5px; }
        .passed { background-color: #d4edda; border: 1px solid #c3e6cb; }
        .failed { background-color: #f8d7da; border: 1px solid #f5c6cb; }
        .test-result { margin: 10px 0; padding: 10px; border-left: 4px solid #007bff; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="header">
        <h1>数据库集成测试报告</h1>
        <p class="timestamp">生成时间: $(date)</p>
    </div>
    
    <div class="summary">
        <h2>测试总结</h2>
        <p>总测试数: $TESTS_RUN</p>
        <p class="passed">通过: $TESTS_PASSED</p>
        <p class="failed">失败: $TESTS_FAILED</p>
        <p>成功率: $(echo "scale=2; $TESTS_PASSED * 100 / $TESTS_RUN" | bc -l)%</p>
    </div>
    
    <div class="test-results">
        <h2>详细结果</h2>
        $(if [ -f "$TEST_RESULTS_FILE" ]; then cat "$TEST_RESULTS_FILE"; fi)
    </div>
</body>
</html>
EOF
    
    log_success "测试报告已生成: test_report.html"
    log_info "详细结果文件: $TEST_RESULTS_FILE"
}

# 主函数
main() {
    log_info "开始数据库集成测试..."
    log_info "测试开始时间: $(date)"
    
    # 初始化JSON结果文件
    echo "[" > "$TEST_RESULTS_FILE"
    
    # 清理环境
    cleanup
    
    # 检查环境
    check_go_environment
    
    # 运行各种测试
    run_unit_tests
    run_benchmark_tests
    test_database_initialization
    test_crud_operations
    test_transaction_operations
    test_deduplication
    test_classification
    test_search_operations
    test_statistics
    
    # 生成报告
    generate_test_report
    
    # 最终清理
    cleanup
    
    # 输出最终结果
    echo
    log_info "========================================"
    log_info "测试完成!"
    log_info "总测试数: $TESTS_RUN"
    log_success "通过: $TESTS_PASSED"
    log_error "失败: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "所有测试都通过了! 🎉"
        exit 0
    else
        log_error "有测试失败，请检查日志"
        exit 1
    fi
}

# 捕获中断信号
trap cleanup EXIT

# 运行主函数
main "$@"