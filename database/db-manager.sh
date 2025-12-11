#!/bin/bash
# gh-pitfall-scraper 数据库管理脚本

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DB_INIT_FILE="$SCRIPT_DIR/init.sql"
DB_NAME="${DB_NAME:-gh_pitfall_scraper}"
DB_USER="${DB_USER:-postgres}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_PASSWORD="${DB_PASSWORD:-}"

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

# 帮助信息
show_help() {
    cat << EOF
gh-pitfall-scraper 数据库管理工具

用法: $0 [选项] <命令>

命令:
    init        初始化数据库
    reset       重置数据库（删除所有数据）
    backup      备份数据库
    restore     恢复数据库
    migrate     运行数据库迁移
    status      显示数据库状态
    cleanup     清理过期数据
    stats       显示数据库统计信息
    test        测试数据库连接

选项:
    -h, --host HOST     数据库主机 (默认: localhost)
    -p, --port PORT     数据库端口 (默认: 5432)
    -U, --user USER     数据库用户 (默认: postgres)
    -d, --dbname NAME   数据库名称 (默认: gh_pitfall_scraper)
    --help              显示此帮助信息

示例:
    $0 init                    # 初始化数据库
    $0 backup                  # 备份数据库
    $0 restore backup.sql      # 从备份文件恢复
    $0 --host remote-db init   # 连接到远程数据库初始化

环境变量:
    DB_HOST, DB_PORT, DB_USER, DB_NAME, DB_PASSWORD

EOF
}

# 检查依赖
check_dependencies() {
    if ! command -v psql &> /dev/null; then
        log_error "psql 命令未找到，请安装 PostgreSQL 客户端"
        exit 1
    fi
    
    if ! command -v pg_dump &> /dev/null; then
        log_error "pg_dump 命令未找到，请安装 PostgreSQL 客户端"
        exit 1
    fi
}

# 设置数据库密码
setup_db_password() {
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
}

# 执行SQL命令
execute_sql() {
    local sql_command="$1"
    local database="$2"
    
    if [ -z "$database" ]; then
        database="$DB_NAME"
    fi
    
    setup_db_password
    
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$database" \
        -c "$sql_command"
}

# 执行SQL文件
execute_sql_file() {
    local sql_file="$1"
    local database="$2"
    
    if [ -z "$database" ]; then
        database="$DB_NAME"
    fi
    
    setup_db_password
    
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$database" \
        -f "$sql_file"
}

# 测试数据库连接
test_connection() {
    log_info "测试数据库连接..."
    
    if execute_sql "SELECT 1;" > /dev/null 2>&1; then
        log_success "数据库连接成功"
        return 0
    else
        log_error "数据库连接失败"
        return 1
    fi
}

# 初始化数据库
init_database() {
    log_info "初始化数据库: $DB_NAME"
    
    # 检查数据库是否存在
    if execute_sql "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" | grep -q "1 row"; then
        log_warning "数据库 $DB_NAME 已存在"
        read -p "是否继续初始化? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "取消初始化"
            return 0
        fi
    fi
    
    # 创建数据库
    execute_sql "CREATE DATABASE $DB_NAME;" postgres
    
    # 执行初始化脚本
    if [ -f "$DB_INIT_FILE" ]; then
        log_info "执行初始化脚本: $DB_INIT_FILE"
        execute_sql_file "$DB_INIT_FILE" "$DB_NAME"
    else
        log_warning "初始化脚本未找到: $DB_INIT_FILE"
    fi
    
    log_success "数据库初始化完成"
}

# 重置数据库
reset_database() {
    log_warning "这将删除数据库 $DB_NAME 中的所有数据!"
    read -p "确认重置数据库? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "取消重置"
        return 0
    fi
    
    log_info "重置数据库..."
    
    # 删除并重新创建数据库
    execute_sql "DROP DATABASE IF EXISTS $DB_NAME;" postgres
    execute_sql "CREATE DATABASE $DB_NAME;" postgres
    
    # 重新初始化
    if [ -f "$DB_INIT_FILE" ]; then
        execute_sql_file "$DB_INIT_FILE" "$DB_NAME"
    fi
    
    log_success "数据库重置完成"
}

# 备份数据库
backup_database() {
    local backup_file="${1:-backup_$(date +%Y%m%d_%H%M%S).sql}"
    local backup_dir="${2:-./backups}"
    
    # 创建备份目录
    mkdir -p "$backup_dir"
    
    log_info "备份数据库到: $backup_dir/$backup_file"
    
    setup_db_password
    
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -f "$backup_dir/$backup_file" \
        --verbose \
        --no-owner \
        --no-privileges
    
    if [ $? -eq 0 ]; then
        log_success "数据库备份完成: $backup_dir/$backup_file"
    else
        log_error "数据库备份失败"
        return 1
    fi
}

# 恢复数据库
restore_database() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "请指定备份文件"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "备份文件不存在: $backup_file"
        return 1
    fi
    
    log_warning "这将覆盖现有数据库 $DB_NAME!"
    read -p "确认恢复数据库? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "取消恢复"
        return 0
    fi
    
    log_info "从备份文件恢复: $backup_file"
    
    # 删除并重新创建数据库
    execute_sql "DROP DATABASE IF EXISTS $DB_NAME;" postgres
    execute_sql "CREATE DATABASE $DB_NAME;" postgres
    
    # 恢复数据
    execute_sql_file "$backup_file" "$DB_NAME"
    
    log_success "数据库恢复完成"
}

# 运行迁移
migrate_database() {
    log_info "运行数据库迁移..."
    
    # 检查迁移表是否存在
    if ! execute_sql "SELECT 1 FROM information_schema.tables WHERE table_name='schema_migrations';" | grep -q "1 row"; then
        log_info "创建迁移表..."
        execute_sql "CREATE TABLE schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );" "$DB_NAME"
    fi
    
    # 这里可以添加迁移逻辑
    log_success "数据库迁移完成"
}

# 显示数据库状态
show_status() {
    log_info "数据库状态信息:"
    echo
    
    # 基本信息
    execute_sql "SELECT 
        current_database() as database_name,
        current_user as current_user,
        version() as postgresql_version;" "$DB_NAME"
    
    echo
    
    # 表统计
    execute_sql "SELECT 
        schemaname,
        tablename,
        n_tup_ins as inserts,
        n_tup_upd as updates,
        n_tup_del as deletes,
        n_live_tup as live_tuples,
        n_dead_tup as dead_tuples
    FROM pg_stat_user_tables 
    ORDER BY schemaname, tablename;" "$DB_NAME"
    
    echo
    
    # 数据库大小
    execute_sql "SELECT 
        pg_size_pretty(pg_database_size('$DB_NAME')) as database_size;" "$DB_NAME"
}

# 清理过期数据
cleanup_data() {
    log_info "清理过期数据..."
    
    # 执行清理函数
    local result
    result=$(execute_sql "SELECT cleanup_old_data(30);" "$DB_NAME")
    
    if [ $? -eq 0 ]; then
        log_success "数据清理完成"
    else
        log_error "数据清理失败"
        return 1
    fi
}

# 显示统计信息
show_stats() {
    log_info "数据库统计信息:"
    echo
    
    # Issues统计
    execute_sql "SELECT * FROM issues_stats;" "$DB_NAME"
    
    echo
    
    # 关键词统计
    execute_sql "SELECT 
        category,
        COUNT(*) as keyword_count,
        AVG(weight) as avg_weight
    FROM keywords 
    WHERE is_active = true
    GROUP BY category;" "$DB_NAME"
    
    echo
    
    # 最近爬取日志
    execute_sql "SELECT 
        repo_owner,
        repo_name,
        status,
        total_issues,
        start_time,
        duration_seconds
    FROM scraping_logs 
    ORDER BY start_time DESC 
    LIMIT 10;" "$DB_NAME"
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--host)
                DB_HOST="$2"
                shift 2
                ;;
            -p|--port)
                DB_PORT="$2"
                shift 2
                ;;
            -U|--user)
                DB_USER="$2"
                shift 2
                ;;
            -d|--dbname)
                DB_NAME="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
    
    # 检查命令
    if [ $# -eq 0 ]; then
        show_help
        exit 1
    fi
    
    command="$1"
    shift || true
    
    # 检查依赖
    check_dependencies
    
    # 执行命令
    case $command in
        init)
            test_connection && init_database
            ;;
        reset)
            test_connection && reset_database
            ;;
        backup)
            test_connection && backup_database "$@"
            ;;
        restore)
            test_connection && restore_database "$1"
            ;;
        migrate)
            test_connection && migrate_database
            ;;
        status)
            test_connection && show_status
            ;;
        cleanup)
            test_connection && cleanup_data
            ;;
        stats)
            test_connection && show_stats
            ;;
        test)
            test_connection
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi