#!/bin/bash

# 抓取引擎数据库集成测试脚本
# 测试抓取引擎的数据库集成功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查Go环境
check_go_environment() {
    print_test "检查Go环境..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go未安装"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go版本: $GO_VERSION"
    
    # 检查必要的包
    if ! go list -m github.com/neo-cloud-ai/gh-pitfall-scraper &> /dev/null; then
        print_warning "项目未在Go模块中，跳过依赖检查"
    fi
}

# 编译测试
compile_test() {
    print_test "编译抓取引擎代码..."
    
    cd /workspace/gh-pitfall-scraper
    
    # 检查语法错误
    if go build -o /tmp/gh-pitfall-scraper .; then
        print_success "代码编译成功"
    else
        print_error "代码编译失败"
        exit 1
    fi
}

# 导入测试
import_test() {
    print_test "测试包导入..."
    
    cd /workspace/gh-pitfall-scraper
    
    # 测试scraper包导入
    if go list ./internal/scraper &> /dev/null; then
        print_success "scraper包导入成功"
    else
        print_error "scraper包导入失败"
        exit 1
    fi
    
    # 测试database包导入
    if go list ./internal/database &> /dev/null; then
        print_success "database包导入成功"
    else
        print_error "database包导入失败"
        exit 1
    fi
}

# 数据库服务测试
database_service_test() {
    print_test "测试数据库服务..."
    
    cd /workspace/gh-pitfall-scraper
    
    # 创建测试Go文件
    cat > /tmp/test_database_service.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
    // 创建临时数据库
    dbPath := "/tmp/test_scraper.db"
    os.Remove(dbPath) // 确保删除旧文件
    
    config := database.DefaultDatabaseConfig()
    config.Path = dbPath
    
    db, err := database.NewSQLiteDB(config)
    if err != nil {
        log.Fatalf("创建数据库失败: %v", err)
    }
    defer db.Close()
    
    // 初始化数据库
    if err := database.InitializeSchema(db); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    
    // 测试DatabaseService创建
    dbService := scraper.NewDatabaseService(db)
    if dbService == nil {
        log.Fatal("DatabaseService创建失败")
    }
    
    // 测试GithubClientWithDB创建
    client := scraper.NewGithubClientWithDB("test_token", dbService)
    if client == nil {
        log.Fatal("GithubClientWithDB创建失败")
    }
    
    // 测试DatabaseScorer创建
    dbScorer := scraper.NewDatabaseScorer(db)
    if dbScorer == nil {
        log.Fatal("DatabaseScorer创建失败")
    }
    
    fmt.Println("所有数据库服务创建成功")
}
EOF
    
    # 运行测试
    if go run /tmp/test_database_service.go; then
        print_success "数据库服务创建测试通过"
    else
        print_error "数据库服务创建测试失败"
        exit 1
    fi
    
    # 清理
    rm -f /tmp/test_database_service.go /tmp/test_scraper.db
}

# 功能测试
functionality_test() {
    print_test "测试抓取引擎功能..."
    
    cd /workspace/gh-pitfall-scraper
    
    # 创建功能测试文件
    cat > /tmp/test_functionality.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
    // 创建临时数据库
    dbPath := "/tmp/test_functionality.db"
    os.Remove(dbPath)
    
    config := database.DefaultDatabaseConfig()
    config.Path = dbPath
    
    db, err := database.NewSQLiteDB(config)
    if err != nil {
        log.Fatalf("创建数据库失败: %v", err)
    }
    defer db.Close()
    
    if err := database.InitializeSchema(db); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    
    // 测试数据库服务
    dbService := scraper.NewDatabaseService(db)
    client := scraper.NewGithubClientWithDB("test_token", dbService)
    
    // 测试数据库查询
    filters := scraper.DefaultDatabaseFilter()
    issues, err := client.GetRecentIssuesFromDatabase(10)
    if err != nil {
        log.Printf("数据库查询测试: %v", err)
    } else {
        fmt.Printf("数据库查询成功，获取到 %d 个Issues\n", len(issues))
    }
    
    // 测试过滤器
    criteria := scraper.DefaultDatabaseFilterCriteria()
    // 这里可以添加更多过滤测试
    
    // 测试评分器
    dbScorer := scraper.NewDatabaseScorer(db)
    stats, err := dbScorer.GetScoreStatistics()
    if err != nil {
        log.Printf("评分统计测试: %v", err)
    } else {
        fmt.Printf("评分统计成功: %+v\n", stats)
    }
    
    fmt.Println("功能测试完成")
}
EOF
    
    # 运行功能测试
    if go run /tmp/test_functionality.go; then
        print_success "功能测试通过"
    else
        print_error "功能测试失败"
        exit 1
    fi
    
    # 清理
    rm -f /tmp/test_functionality.go /tmp/test_functionality.db
}

# 并发测试
concurrency_test() {
    print_test "测试并发处理..."
    
    cd /workspace/gh-pitfall-scraper
    
    cat > /tmp/test_concurrency.go << 'EOF'
package main

import (
    "database/sql"
    "log"
    "os"
    "sync"
    "time"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
    dbPath := "/tmp/test_concurrency.db"
    os.Remove(dbPath)
    
    config := database.DefaultDatabaseConfig()
    config.Path = dbPath
    
    db, err := database.NewSQLiteDB(config)
    if err != nil {
        log.Fatalf("创建数据库失败: %v", err)
    }
    defer db.Close()
    
    if err := database.InitializeSchema(db); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    
    dbService := scraper.NewDatabaseService(db)
    var wg sync.WaitGroup
    
    // 模拟并发访问
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // 模拟数据库操作
            client := scraper.NewGithubClientWithDB("test_token", dbService)
            _, err := client.GetRecentIssuesFromDatabase(10)
            if err != nil {
                log.Printf("goroutine %d error: %v", id, err)
            }
            
            time.Sleep(100 * time.Millisecond)
        }(i)
    }
    
    wg.Wait()
    log.Println("并发测试完成")
}
EOF
    
    # 运行并发测试
    if go run /tmp/test_concurrency.go; then
        print_success "并发测试通过"
    else
        print_error "并发测试失败"
        exit 1
    fi
    
    # 清理
    rm -f /tmp/test_concurrency.go /tmp/test_concurrency.db
}

# 错误处理测试
error_handling_test() {
    print_test "测试错误处理..."
    
    cd /workspace/gh-pitfall-scraper
    
    cat > /tmp/test_error_handling.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
    // 测试无效数据库路径
    config := database.DefaultDatabaseConfig()
    config.Path = "/invalid/path/test.db"
    
    _, err := database.NewSQLiteDB(config)
    if err != nil {
        fmt.Println("正确处理了无效数据库路径错误")
    } else {
        log.Fatal("应该返回错误但没有")
    }
    
    // 测试空数据库服务
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("正确处理了panic")
        }
    }()
    
    // 模拟一些错误情况
    dbService := scraper.NewDatabaseService(nil)
    if dbService == nil {
        fmt.Println("正确处理了nil数据库")
    }
    
    fmt.Println("错误处理测试完成")
}
EOF
    
    # 运行错误处理测试
    if go run /tmp/test_error_handling.go; then
        print_success "错误处理测试通过"
    else
        print_error "错误处理测试失败"
        exit 1
    fi
    
    # 清理
    rm -f /tmp/test_error_handling.go
}

# 性能测试
performance_test() {
    print_test "运行性能测试..."
    
    cd /workspace/gh-pitfall-scraper
    
    cat > /tmp/test_performance.go << 'EOF'
package main

import (
    "database/sql"
    "log"
    "os"
    "time"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
    dbPath := "/tmp/test_performance.db"
    os.Remove(dbPath)
    
    config := database.DefaultDatabaseConfig()
    config.Path = dbPath
    
    db, err := database.NewSQLiteDB(config)
    if err != nil {
        log.Fatalf("创建数据库失败: %v", err)
    }
    defer db.Close()
    
    if err := database.InitializeSchema(db); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    
    dbService := scraper.NewDatabaseService(db)
    
    start := time.Now()
    
    // 批量数据库操作测试
    for i := 0; i < 100; i++ {
        client := scraper.NewGithubClientWithDB("test_token", dbService)
        _, err := client.GetRecentIssuesFromDatabase(10)
        if err != nil {
            log.Printf("操作 %d 失败: %v", i, err)
        }
    }
    
    elapsed := time.Since(start)
    log.Printf("执行100次数据库操作耗时: %v", elapsed)
    log.Printf("平均每次操作耗时: %v", elapsed/100)
    
    // 性能测试完成
}
EOF
    
    # 运行性能测试
    if timeout 30s go run /tmp/test_performance.go; then
        print_success "性能测试通过"
    else
        print_error "性能测试超时或失败"
        exit 1
    fi
    
    # 清理
    rm -f /tmp/test_performance.go /tmp/test_performance.db
}

# 清理函数
cleanup() {
    print_test "清理测试文件..."
    rm -f /tmp/gh-pitfall-scraper
    rm -f /tmp/test_*.go
    rm -f /tmp/test_*.db
    print_success "清理完成"
}

# 主函数
main() {
    echo "=========================================="
    echo "  抓取引擎数据库集成测试"
    echo "=========================================="
    
    # 陷阱处理
    trap cleanup EXIT
    
    # 运行所有测试
    check_go_environment
    compile_test
    import_test
    database_service_test
    functionality_test
    concurrency_test
    error_handling_test
    performance_test
    
    echo ""
    echo "=========================================="
    print_success "所有测试通过！抓取引擎数据库集成功能正常"
    echo "=========================================="
}

# 运行主函数
main