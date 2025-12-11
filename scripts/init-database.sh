#!/bin/bash

# gh-pitfall-scraper 数据库初始化脚本
# 作者: gh-pitfall-scraper 项目组
# 版本: 1.0.0
# 描述: 初始化数据库，创建必要的表结构和索引

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEFAULT_CONFIG_FILE="$PROJECT_ROOT/config.yaml"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印彩色信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
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

# 显示帮助信息
show_help() {
    cat << EOF
gh-pitfall-scraper 数据库初始化脚本

用法: $0 [选项]

选项:
    -c, --config <文件>     指定配置文件路径 (默认: $DEFAULT_CONFIG_FILE)
    -d, --drop              删除现有表结构并重新创建
    -s, --seed              创建测试数据
    -v, --verbose          详细输出
    -f, --force            强制执行 (跳过确认)
    --dry-run              试运行模式 (不实际执行)
    -h, --help            显示帮助信息

示例:
    $0                      # 初始化数据库
    $0 -d                   # 重新创建所有表
    $0 -s                   # 初始化并添加测试数据
    $0 -f                   # 强制初始化 (跳过确认)

EOF
}

# 解析命令行参数
parse_args() {
    CONFIG_FILE="$DEFAULT_CONFIG_FILE"
    DROP_TABLES=false
    SEED_DATA=false
    VERBOSE=false
    FORCE=false
    DRY_RUN=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -d|--drop)
                DROP_TABLES=true
                shift
                ;;
            -s|--seed)
                SEED_DATA=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查配置文件
check_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        print_warning "配置文件不存在: $CONFIG_FILE，将使用默认配置"
    else
        if [[ "$VERBOSE" == "true" ]]; then
            print_info "使用配置文件: $CONFIG_FILE"
        fi
    fi
}

# 从配置文件提取数据库路径
get_db_path() {
    if [[ -f "$CONFIG_FILE" ]]; then
        if command -v yq >/dev/null 2>&1; then
            DB_PATH=$(yq eval '.database.file_path' "$CONFIG_FILE" 2>/dev/null || echo "")
        elif command -v python3 >/dev/null 2>&1; then
            DB_PATH=$(python3 -c "
import yaml
try:
    with open('$CONFIG_FILE', 'r') as f:
        config = yaml.safe_load(f)
        print(config.get('database', {}).get('file_path', './data/gh-pitfall-scraper.db'))
except Exception as e:
    print('./data/gh-pitfall-scraper.db')
" 2>/dev/null || echo "./data/gh-pitfall-scraper.db")
        else
            DB_PATH="./data/gh-pitfall-scraper.db"
        fi
    else
        DB_PATH="./data/gh-pitfall-scraper.db"
    fi
    
    # 转换为绝对路径
    if [[ ! "$DB_PATH" = /* ]]; then
        DB_PATH="$PROJECT_ROOT/$DB_PATH"
    fi
    
    echo "$DB_PATH"
}

# 创建数据库目录
create_db_directory() {
    DB_DIR=$(dirname "$DB_PATH")
    
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "创建数据库目录: $DB_DIR"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将创建目录 $DB_DIR"
        return
    fi
    
    mkdir -p "$DB_DIR"
    print_success "数据库目录准备完成"
}

# 生成数据库表结构 SQL
generate_schema_sql() {
    local drop_tables="$1"
    local sql=""
    
    if [[ "$drop_tables" == "true" ]]; then
        sql="
-- 删除现有表
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS commits;
DROP TABLE IF EXISTS pulls;
DROP TABLE IF EXISTS scraper_metadata;
DROP TABLE IF EXISTS error_patterns;
DROP TABLE IF EXISTS performance_metrics;
"
    fi
    
    sql="${sql}
-- 仓库表
CREATE TABLE IF NOT EXISTS repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    full_name TEXT UNIQUE NOT NULL,
    description TEXT,
    url TEXT,
    stars INTEGER DEFAULT 0,
    forks INTEGER DEFAULT 0,
    watchers INTEGER DEFAULT 0,
    language TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    archived BOOLEAN DEFAULT FALSE,
    disabled BOOLEAN DEFAULT FALSE,
    UNIQUE(owner, name)
);

-- GitHub Issues 表
CREATE TABLE IF NOT EXISTS issues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id INTEGER NOT NULL,
    issue_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL,
    author TEXT NOT NULL,
    author_type TEXT,
    labels TEXT, -- JSON array of labels
    milestone TEXT,
    assignees TEXT, -- JSON array of assignees
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    comments_count INTEGER DEFAULT 0,
    url TEXT,
    html_url TEXT,
    FOREIGN KEY (repo_id) REFERENCES repositories (id),
    UNIQUE(repo_id, issue_id)
);

-- Issue 评论表
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    body TEXT NOT NULL,
    author TEXT NOT NULL,
    author_type TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    url TEXT,
    FOREIGN KEY (issue_id) REFERENCES issues (id),
    UNIQUE(issue_id, comment_id)
);

-- 提交表
CREATE TABLE IF NOT EXISTS commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id INTEGER NOT NULL,
    sha TEXT UNIQUE NOT NULL,
    message TEXT NOT NULL,
    author TEXT NOT NULL,
    author_email TEXT,
    committer TEXT,
    committer_email TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    url TEXT,
    stats_additions INTEGER DEFAULT 0,
    stats_deletions INTEGER DEFAULT 0,
    stats_total INTEGER DEFAULT 0,
    FOREIGN KEY (repo_id) REFERENCES repositories (id)
);

-- Pull Request 表
CREATE TABLE IF NOT EXISTS pulls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id INTEGER NOT NULL,
    pull_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL,
    author TEXT NOT NULL,
    base_branch TEXT,
    head_branch TEXT,
    mergeable BOOLEAN,
    merged BOOLEAN DEFAULT FALSE,
    merged_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    url TEXT,
    html_url TEXT,
    FOREIGN KEY (repo_id) REFERENCES repositories (id),
    UNIQUE(repo_id, pull_id)
);

-- 爬虫元数据表
CREATE TABLE IF NOT EXISTS scraper_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 错误模式表
CREATE TABLE IF NOT EXISTS error_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern TEXT UNIQUE NOT NULL,
    description TEXT,
    severity TEXT DEFAULT 'medium',
    category TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 性能指标表
CREATE TABLE IF NOT EXISTS performance_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id INTEGER NOT NULL,
    metric_type TEXT NOT NULL,
    value REAL NOT NULL,
    unit TEXT,
    description TEXT,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    source TEXT,
    FOREIGN KEY (repo_id) REFERENCES repositories (id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_repositories_owner_name ON repositories (owner, name);
CREATE INDEX IF NOT EXISTS idx_repositories_full_name ON repositories (full_name);
CREATE INDEX IF NOT EXISTS idx_repositories_language ON repositories (language);

CREATE INDEX IF NOT EXISTS idx_issues_repo_id ON issues (repo_id);
CREATE INDEX IF NOT EXISTS idx_issues_number ON issues (number);
CREATE INDEX IF NOT EXISTS idx_issues_state ON issues (state);
CREATE INDEX IF NOT EXISTS idx_issues_author ON issues (author);
CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues (created_at);
CREATE INDEX IF NOT EXISTS idx_issues_updated_at ON issues (updated_at);
CREATE INDEX IF NOT EXISTS idx_issues_closed_at ON issues (closed_at);

CREATE INDEX IF NOT EXISTS idx_comments_issue_id ON comments (issue_id);
CREATE INDEX IF NOT EXISTS idx_comments_author ON comments (author);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments (created_at);

CREATE INDEX IF NOT EXISTS idx_commits_repo_id ON commits (repo_id);
CREATE INDEX IF NOT EXISTS idx_commits_sha ON commits (sha);
CREATE INDEX IF NOT EXISTS idx_commits_author ON commits (author);
CREATE INDEX IF NOT EXISTS idx_commits_created_at ON commits (created_at);

CREATE INDEX IF NOT EXISTS idx_pulls_repo_id ON pulls (repo_id);
CREATE INDEX IF NOT EXISTS idx_pulls_number ON pulls (number);
CREATE INDEX IF NOT EXISTS idx_pulls_state ON pulls (state);
CREATE INDEX IF NOT EXISTS idx_pulls_author ON pulls (author);
CREATE INDEX IF NOT EXISTS idx_pulls_created_at ON pulls (created_at);

CREATE INDEX IF NOT EXISTS idx_scraper_metadata_key ON scraper_metadata (key);

CREATE INDEX IF NOT EXISTS idx_error_patterns_pattern ON error_patterns (pattern);
CREATE INDEX IF NOT EXISTS idx_error_patterns_enabled ON error_patterns (enabled);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_repo_id ON performance_metrics (repo_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_type ON performance_metrics (metric_type);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_recorded_at ON performance_metrics (recorded_at);

-- 触发器：自动更新 updated_at 字段
CREATE TRIGGER IF NOT EXISTS update_repositories_updated_at 
    AFTER UPDATE ON repositories
    BEGIN
        UPDATE repositories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_issues_updated_at 
    AFTER UPDATE ON issues
    BEGIN
        UPDATE issues SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_comments_updated_at 
    AFTER UPDATE ON comments
    BEGIN
        UPDATE comments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_pulls_updated_at 
    AFTER UPDATE ON pulls
    BEGIN
        UPDATE pulls SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_scraper_metadata_updated_at 
    AFTER UPDATE ON scraper_metadata
    BEGIN
        UPDATE scraper_metadata SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
"
    
    echo "$sql"
}

# 执行 SQL 语句
execute_sql() {
    local sql="$1"
    local description="$2"
    
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "$description"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: $description"
        return
    fi
    
    if command -v sqlite3 >/dev/null 2>&1; then
        echo "$sql" | sqlite3 "$DB_PATH"
        if [[ $? -eq 0 ]]; then
            print_success "$description 完成"
        else
            print_error "$description 失败"
            return 1
        fi
    else
        print_error "sqlite3 命令未找到，无法执行 SQL"
        return 1
    fi
}

# 初始化数据库
init_database() {
    print_info "开始初始化数据库..."
    
    # 生成并执行数据库表结构 SQL
    local schema_sql=$(generate_schema_sql "$DROP_TABLES")
    execute_sql "$schema_sql" "创建数据库表结构"
    
    # 插入初始元数据
    insert_initial_metadata
    
    # 插入默认错误模式
    insert_default_error_patterns
    
    print_success "数据库初始化完成"
}

# 插入初始元数据
insert_initial_metadata() {
    local sql="
INSERT OR IGNORE INTO scraper_metadata (key, value, description) VALUES
('schema_version', '1.0.0', '数据库模式版本'),
('init_timestamp', '$(date -Iseconds)', '数据库初始化时间'),
('created_by', 'gh-pitfall-scraper-init-script', '初始化脚本名称'),
('db_type', 'sqlite', '数据库类型');

SELECT '插入初始元数据完成' as status;
"
    
    execute_sql "$sql" "插入初始元数据"
}

# 插入默认错误模式
insert_default_error_patterns() {
    local patterns=(
        "performance|性能问题"
        "regression|回归问题"
        "latency|延迟问题"
        "throughput|吞吐量问题"
        "OOM|内存溢出"
        "memory leak|内存泄漏"
        "CUDA|CUDA相关错误"
        "kernel|内核问题"
        "NCCL|NCCL通信错误"
        "hang|程序挂起"
        "deadlock|死锁"
        "kv cache|KV缓存问题"
    )
    
    local sql="INSERT OR IGNORE INTO error_patterns (pattern, description, severity, category) VALUES "
    local values=()
    
    for pattern_info in "${patterns[@]}"; do
        IFS='|' read -r pattern description <<< "$pattern_info"
        values+=("('$pattern', '$description', 'medium', 'performance')")
    done
    
    sql="${sql}${values[*]};"
    
    execute_sql "$sql" "插入默认错误模式"
}

# 创建测试数据
create_seed_data() {
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "创建测试数据..."
    fi
    
    # 插入示例仓库数据
    local sql="
INSERT OR IGNORE INTO repositories (owner, name, full_name, description, language, stars) VALUES
('test-owner', 'test-repo', 'test-owner/test-repo', '测试仓库', 'Go', 100),
('example', 'demo-project', 'example/demo-project', '示例项目', 'Python', 250);

INSERT OR IGNORE INTO issues (repo_id, issue_id, number, title, body, state, author, labels) VALUES
(1, 1000, 1, '性能问题测试', '这是一个测试性能问题的issue', 'open', 'test-user', '["performance", "bug"]'),
(1, 1001, 2, '内存泄漏问题', '发现内存泄漏问题', 'open', 'test-user', '["memory-leak", "critical"]'),
(2, 2000, 1, 'CUDA错误', 'CUDA内核执行失败', 'open', 'dev-user', '["CUDA", "error"]');

INSERT OR IGNORE INTO error_patterns (pattern, description, severity, category) VALUES
('test-pattern', '测试错误模式', 'low', 'test');
"
    
    execute_sql "$sql" "创建测试数据"
    
    print_success "测试数据创建完成"
}

# 验证数据库
validate_database() {
    print_info "验证数据库结构..."
    
    local sql="
SELECT 
    'repositories' as table_name, 
    COUNT(*) as record_count 
FROM repositories
UNION ALL
SELECT 
    'issues' as table_name, 
    COUNT(*) as record_count 
FROM issues
UNION ALL
SELECT 
    'error_patterns' as table_name, 
    COUNT(*) as record_count 
FROM error_patterns;
"
    
    if [[ "$VERBOSE" == "true" ]]; then
        execute_sql "$sql" "验证数据库记录"
    else
        echo "$sql" | sqlite3 "$DB_PATH" | while read line; do
            print_info "表统计: $line"
        done
    fi
}

# 主函数
main() {
    parse_args "$@"
    check_config
    
    # 获取数据库路径
    DB_PATH=$(get_db_path)
    
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "数据库文件路径: $DB_PATH"
    fi
    
    # 创建数据库目录
    create_db_directory
    
    # 检查数据库是否已存在
    if [[ -f "$DB_PATH" && "$DROP_TABLES" != "true" ]]; then
        if [[ "$FORCE" != "true" ]]; then
            print_warning "数据库文件已存在: $DB_PATH"
            read -p "是否继续? 这将创建新表或更新现有表结构. (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_info "操作已取消"
                return 0
            fi
        fi
    fi
    
    # 初始化数据库
    init_database
    
    # 创建测试数据
    if [[ "$SEED_DATA" == "true" ]]; then
        create_seed_data
    fi
    
    # 验证数据库
    validate_database
    
    print_success "数据库初始化脚本执行完成"
    print_info "数据库文件: $DB_PATH"
}

# 检查依赖
check_dependencies() {
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "检查依赖..."
    fi
    
    if ! command -v sqlite3 >/dev/null 2>&1; then
        print_error "sqlite3 命令未找到，请安装 SQLite"
        exit 1
    fi
    
    # 检查可选依赖
    if ! command -v yq >/dev/null 2>&1; then
        if [[ "$VERBOSE" == "true" ]]; then
            print_warning "yq 命令未找到，将使用默认配置"
        fi
    fi
    
    if ! command -v python3 >/dev/null 2>&1; then
        if [[ "$VERBOSE" == "true" ]]; then
            print_warning "python3 命令未找到，将使用默认配置"
        fi
    fi
}

# 执行主函数
check_dependencies
main "$@"