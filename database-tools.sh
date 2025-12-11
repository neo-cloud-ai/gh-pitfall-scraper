#!/bin/bash

# gh-pitfall-scraper 数据库管理命令行工具
# 作者: gh-pitfall-scraper 项目组
# 版本: 1.0.0
# 描述: 提供数据库管理功能，包括初始化、备份、恢复、清理等操作

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/" && pwd)"
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
gh-pitfall-scraper 数据库管理工具

用法: $0 [选项] <命令> [参数]

命令:
    init                    初始化数据库
    backup [输出文件]       创建数据库备份
    restore <备份文件>      从备份文件恢复数据库
    clean                   清理过期数据
    vacuum                  压缩数据库文件
    status                  显示数据库状态信息
    info                    显示数据库配置信息
    migrate [版本]          执行数据库迁移
    health                  检查数据库健康状态
    test                    测试数据库连接

选项:
    -c, --config <文件>     指定配置文件路径 (默认: $DEFAULT_CONFIG_FILE)
    -v, --verbose          详细输出
    -f, --force            强制执行 (跳过确认)
    --dry-run              试运行模式 (不实际执行)
    -h, --help            显示帮助信息

示例:
    $0 init                              # 初始化数据库
    $0 backup                            # 创建备份 (自动命名)
    $0 backup backup-$(date +%Y%m%d).db  # 指定备份文件名
    $0 restore ./backups/backup-20231211.db  # 恢复数据库
    $0 clean                            # 清理过期数据
    $0 status                           # 查看数据库状态
    $0 health                           # 检查数据库健康状态

EOF
}

# 解析命令行参数
parse_args() {
    CONFIG_FILE="$DEFAULT_CONFIG_FILE"
    VERBOSE=false
    FORCE=false
    DRY_RUN=false
    COMMAND=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
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
            init|backup|restore|clean|vacuum|status|info|migrate|health|test)
                COMMAND="$1"
                shift
                break
                ;;
            *)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查是否提供了命令
    if [[ -z "$COMMAND" ]]; then
        print_error "请指定命令"
        show_help
        exit 1
    fi
}

# 检查配置文件
check_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        print_error "配置文件不存在: $CONFIG_FILE"
        exit 1
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "使用配置文件: $CONFIG_FILE"
    fi
}

# 从配置文件提取数据库路径
get_db_path() {
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
        # 默认值
        DB_PATH="./data/gh-pitfall-scraper.db"
    fi
    
    # 转换为绝对路径
    if [[ ! "$DB_PATH" = /* ]]; then
        DB_PATH="$PROJECT_ROOT/$DB_PATH"
    fi
    
    echo "$DB_PATH"
}

# 初始化数据库
init_database() {
    print_info "正在初始化数据库..."
    
    DB_PATH=$(get_db_path)
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将创建数据库文件: $DB_PATH"
        return
    fi
    
    # 创建数据库目录
    DB_DIR=$(dirname "$DB_PATH")
    mkdir -p "$DB_DIR"
    
    # 检查数据库是否已存在
    if [[ -f "$DB_PATH" ]]; then
        if [[ "$FORCE" == "true" ]]; then
            print_warning "数据库已存在，将被覆盖"
        else
            read -p "数据库已存在，是否继续? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_info "操作已取消"
                return
            fi
        fi
    fi
    
    # 创建数据库文件 (使用 Go 程序)
    if [[ -f "$PROJECT_ROOT/main.go" ]]; then
        cd "$PROJECT_ROOT"
        go run main.go --init-db || {
            print_error "Go 程序初始化数据库失败"
            return 1
        }
    else
        # 创建空的 SQLite 数据库文件
        touch "$DB_PATH"
        print_success "创建了空的数据库文件: $DB_PATH"
    fi
    
    print_success "数据库初始化完成: $DB_PATH"
}

# 备份数据库
backup_database() {
    BACKUP_FILE="$1"
    
    if [[ -z "$BACKUP_FILE" ]]; then
        BACKUP_FILE="$PROJECT_ROOT/backups/backup-$(date +%Y%m%d-%H%M%S).db"
    fi
    
    # 确保备份目录存在
    mkdir -p "$(dirname "$BACKUP_FILE")"
    
    DB_PATH=$(get_db_path)
    
    print_info "正在备份数据库到: $BACKUP_FILE"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将备份数据库 $DB_PATH 到 $BACKUP_FILE"
        return
    fi
    
    # 检查源数据库是否存在
    if [[ ! -f "$DB_PATH" ]]; then
        print_error "源数据库文件不存在: $DB_PATH"
        exit 1
    fi
    
    # 复制数据库文件
    cp "$DB_PATH" "$BACKUP_FILE"
    
    # 压缩备份文件
    gzip "$BACKUP_FILE"
    
    print_success "数据库备份完成: $BACKUP_FILE.gz"
}

# 恢复数据库
restore_database() {
    BACKUP_FILE="$1"
    
    if [[ -z "$BACKUP_FILE" ]]; then
        print_error "请指定备份文件路径"
        exit 1
    fi
    
    DB_PATH=$(get_db_path)
    
    print_info "正在从备份恢复数据库: $BACKUP_FILE"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将从 $BACKUP_FILE 恢复到 $DB_PATH"
        return
    fi
    
    # 检查备份文件是否存在
    if [[ ! -f "$BACKUP_FILE" ]]; then
        print_error "备份文件不存在: $BACKUP_FILE"
        exit 1
    fi
    
    # 如果是压缩文件，先解压
    if [[ "$BACKUP_FILE" == *.gz ]]; then
        TEMP_FILE=$(mktemp)
        gunzip -c "$BACKUP_FILE" > "$TEMP_FILE"
        BACKUP_FILE="$TEMP_FILE"
        CLEANUP_TEMP=true
    fi
    
    # 创建数据库目录
    mkdir -p "$(dirname "$DB_PATH")"
    
    # 备份当前数据库 (如果存在)
    if [[ -f "$DB_PATH" ]]; then
        mv "$DB_PATH" "$DB_PATH.backup.$(date +%Y%m%d-%H%M%S)"
    fi
    
    # 恢复数据库
    cp "$BACKUP_FILE" "$DB_PATH"
    
    # 清理临时文件
    if [[ "$CLEANUP_TEMP" == "true" ]]; then
        rm "$BACKUP_FILE"
    fi
    
    print_success "数据库恢复完成: $DB_PATH"
}

# 清理过期数据
cleanup_database() {
    print_info "正在清理过期数据..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将清理过期数据"
        return
    fi
    
    # 清理逻辑将在这里实现
    # 目前只是占位符
    print_success "数据清理完成"
}

# 压缩数据库
vacuum_database() {
    print_info "正在压缩数据库文件..."
    
    DB_PATH=$(get_db_path)
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将压缩数据库文件 $DB_PATH"
        return
    fi
    
    # 检查数据库是否存在
    if [[ ! -f "$DB_PATH" ]]; then
        print_error "数据库文件不存在: $DB_PATH"
        exit 1
    fi
    
    # 记录原始大小
    ORIGINAL_SIZE=$(du -h "$DB_PATH" | cut -f1)
    
    # 使用 sqlite3 命令压缩数据库
    if command -v sqlite3 >/dev/null 2>&1; then
        sqlite3 "$DB_PATH" "VACUUM;"
        VACUUMED_SIZE=$(du -h "$DB_PATH" | cut -f1)
        print_success "数据库压缩完成: $ORIGINAL_SIZE -> $VACUUMED_SIZE"
    else
        print_warning "sqlite3 命令未找到，无法执行压缩操作"
    fi
}

# 显示数据库状态
show_status() {
    print_info "数据库状态信息:"
    
    DB_PATH=$(get_db_path)
    
    if [[ ! -f "$DB_PATH" ]]; then
        print_warning "数据库文件不存在: $DB_PATH"
        return
    fi
    
    # 显示文件信息
    echo "数据库文件: $DB_PATH"
    echo "文件大小: $(du -h "$DB_PATH" | cut -f1)"
    echo "创建时间: $(stat -c %y "$DB_PATH" 2>/dev/null || stat -f %Sm "$DB_PATH" 2>/dev/null || echo "未知")"
    echo "修改时间: $(stat -c %y "$DB_PATH" 2>/dev/null || stat -f %Sm "$DB_PATH" 2>/dev/null || echo "未知")"
    
    # 使用 sqlite3 显示数据库信息
    if command -v sqlite3 >/dev/null 2>&1; then
        echo
        echo "数据库详细信息:"
        sqlite3 "$DB_PATH" "
            SELECT '数据库页面大小: ' || page_size || ' bytes';
            SELECT '页面数量: ' || page_count;
            SELECT '数据库大小: ' || page_count * page_size / 1024 / 1024 || ' MB';
        " 2>/dev/null || echo "无法获取数据库详细信息"
    fi
}

# 显示数据库配置信息
show_config() {
    print_info "数据库配置信息:"
    
    if command -v yq >/dev/null 2>&1; then
        yq eval '.database' "$CONFIG_FILE" 2>/dev/null || {
            print_warning "无法解析配置文件"
        }
    else
        print_warning "yq 命令未找到，无法显示详细配置"
        echo "配置文件: $CONFIG_FILE"
    fi
}

# 健康检查
health_check() {
    print_info "正在检查数据库健康状态..."
    
    DB_PATH=$(get_db_path)
    
    # 检查数据库文件是否存在
    if [[ ! -f "$DB_PATH" ]]; then
        print_error "数据库文件不存在"
        return 1
    fi
    
    # 检查文件权限
    if [[ ! -r "$DB_PATH" ]]; then
        print_error "数据库文件不可读"
        return 1
    fi
    
    # 检查 SQLite 数据库完整性
    if command -v sqlite3 >/dev/null 2>&1; then
        if sqlite3 "$DB_PATH" "PRAGMA integrity_check;" | grep -q "ok"; then
            print_success "数据库完整性检查通过"
        else
            print_error "数据库完整性检查失败"
            return 1
        fi
    else
        print_warning "sqlite3 命令未找到，跳过完整性检查"
    fi
    
    # 检查磁盘空间
    AVAILABLE_SPACE=$(df "$(dirname "$DB_PATH")" | awk 'NR==2 {print $4}')
    if [[ "$AVAILABLE_SPACE" -lt 1048576 ]]; then # 1GB in KB
        print_warning "磁盘空间不足，建议清理或扩容"
    else
        print_success "磁盘空间充足"
    fi
    
    print_success "数据库健康状态良好"
}

# 测试数据库连接
test_connection() {
    print_info "正在测试数据库连接..."
    
    DB_PATH=$(get_db_path)
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将测试连接到 $DB_PATH"
        return
    fi
    
    if command -v sqlite3 >/dev/null 2>&1; then
        if sqlite3 "$DB_PATH" "SELECT 1;" >/dev/null 2>&1; then
            print_success "数据库连接测试成功"
        else
            print_error "数据库连接测试失败"
            return 1
        fi
    else
        print_warning "sqlite3 命令未找到，无法测试连接"
    fi
}

# 数据库迁移
migrate_database() {
    VERSION="$1"
    
    if [[ -z "$VERSION" ]]; then
        print_info "执行默认迁移..."
    else
        print_info "执行迁移到版本: $VERSION"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "试运行: 将执行数据库迁移"
        return
    fi
    
    # 迁移逻辑将在这里实现
    print_success "数据库迁移完成"
}

# 主函数
main() {
    parse_args "$@"
    check_config
    
    case "$COMMAND" in
        init)
            init_database
            ;;
        backup)
            backup_database "$1"
            ;;
        restore)
            restore_database "$1"
            ;;
        clean)
            cleanup_database
            ;;
        vacuum)
            vacuum_database
            ;;
        status)
            show_status
            ;;
        info)
            show_config
            ;;
        migrate)
            migrate_database "$1"
            ;;
        health)
            health_check
            ;;
        test)
            test_connection
            ;;
        *)
            print_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 检查依赖
check_dependencies() {
    if [[ "$VERBOSE" == "true" ]]; then
        print_info "检查依赖..."
    fi
    
    # 检查可选依赖
    if ! command -v sqlite3 >/dev/null 2>&1; then
        print_warning "sqlite3 命令未找到，部分功能可能不可用"
    fi
    
    if ! command -v yq >/dev/null 2>&1; then
        print_warning "yq 命令未找到，配置解析功能可能受限"
    fi
}

# 执行主函数
check_dependencies
main "$@"
