---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 3045022100e33becc93010ed3dfefa734d5af4540cc81b37e5a0eba0b80e000f9a76c719cc02204eb0a62161a98ca5a12195f8a97febd7ccb0af7c4321d91f0c988c4cb727c6d0
    ReservedCode2: 30450220700d8aa4470ffeace79797046c98ef3af11e2a87267ecfa86be45f37f7383129022100c4303760011c1414f8a9b82aa59ef0a697d854f445b69592a5c2880d676eeadb
---

# gh-pitfall-scraper 数据库故障排除指南

## 目录

1. [快速诊断](#快速诊断)
2. [常见问题及解决方案](#常见问题及解决方案)
3. [错误代码说明](#错误代码说明)
4. [性能问题排查](#性能问题排查)
5. [数据问题处理](#数据问题处理)
6. [配置问题诊断](#配置问题诊断)
7. [环境问题排查](#环境问题排查)
8. [应急处理流程](#应急处理流程)
9. [预防措施](#预防措施)

## 快速诊断

### 1. 一键诊断脚本
```bash
#!/bin/bash
# quick-diagnosis.sh - 快速诊断脚本

echo "🔍 gh-pitfall-scraper 数据库快速诊断"
echo "======================================"

# 检查系统环境
echo ""
echo "📋 系统环境检查:"
echo "Go 版本: $(go version 2>/dev/null || echo 'Go 未安装')"
echo "SQLite: $(sqlite3 --version 2>/dev/null || echo 'SQLite3 命令行工具未安装')"
echo "磁盘空间: $(df -h . | tail -1 | awk '{print $4 " 可用"}')"
echo "内存使用: $(free -h | grep '^Mem:' | awk '{print $3 "/" $2}')"

# 检查数据库文件
echo ""
echo "🗄️  数据库文件检查:"
if [ -f "./data/gh-pitfall-scraper.db" ]; then
    echo "✅ 数据库文件存在"
    echo "文件大小: $(ls -lh ./data/gh-pitfall-scraper.db | awk '{print $5}')"
    echo "权限: $(ls -l ./data/gh-pitfall-scraper.db | awk '{print $1}')"
else
    echo "❌ 数据库文件不存在"
fi

# 检查配置文件
echo ""
echo "⚙️  配置文件检查:"
if [ -f "./config.yaml" ]; then
    echo "✅ 配置文件存在"
    echo "文件大小: $(ls -lh ./config.yaml | awk '{print $5}')"
else
    echo "❌ 配置文件不存在"
fi

# 检查进程
echo ""
echo "🔍 进程检查:"
if pgrep -f gh-pitfall-scraper > /dev/null; then
    echo "⚠️  gh-pitfall-scraper 进程正在运行"
    ps aux | grep gh-pitfall-scraper | grep -v grep
else
    echo "✅ gh-pitfall-scraper 进程未运行"
fi

# 数据库连接测试
echo ""
echo "🔗 数据库连接测试:"
if [ -f "./data/gh-pitfall-scraper.db" ]; then
    if sqlite3 ./data/gh-pitfall-scraper.db "PRAGMA integrity_check;" > /dev/null 2>&1; then
        echo "✅ 数据库连接正常"
    else
        echo "❌ 数据库连接失败"
    fi
else
    echo "⚠️  数据库文件不存在，无法测试连接"
fi

# 检查日志文件
echo ""
echo "📄 日志文件检查:"
if [ -d "./logs" ]; then
    echo "✅ 日志目录存在"
    echo "日志文件大小:"
    du -sh ./logs/* 2>/dev/null | sort -hr
else
    echo "⚠️  日志目录不存在"
fi

echo ""
echo "🏥 健康检查:"
./gh-pitfall-scraper --health --config config.yaml 2>/dev/null || echo "健康检查失败，请查看详细日志"

echo ""
echo "📊 快速诊断完成"
```

### 2. 详细诊断脚本
```bash
#!/bin/bash
# detailed-diagnosis.sh - 详细诊断脚本

generate_diagnosis_report() {
    local report_file="diagnosis-report-$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "🔍 gh-pitfall-scraper 详细诊断报告"
        echo "生成时间: $(date)"
        echo "=================================="
        echo ""
        
        # 系统信息
        echo "📋 系统信息:"
        echo "操作系统: $(uname -a)"
        echo "Go 版本: $(go version 2>/dev/null || echo 'Go 未安装')"
        echo "当前用户: $(whoami)"
        echo "当前目录: $(pwd)"
        echo "系统时间: $(date)"
        echo ""
        
        # 资源使用
        echo "💾 资源使用情况:"
        echo "磁盘空间:"
        df -h
        echo ""
        echo "内存使用:"
        free -h
        echo ""
        echo "CPU 信息:"
        cat /proc/cpuinfo | grep "model name" | head -1
        echo ""
        
        # 网络状态
        echo "🌐 网络状态:"
        echo "监听端口:"
        netstat -tuln 2>/dev/null | grep LISTEN || ss -tuln 2>/dev/null | grep LISTEN
        echo ""
        
        # 数据库状态
        echo "🗄️  数据库状态:"
        echo "数据库文件:"
        ls -la ./data/ 2>/dev/null || echo "data 目录不存在"
        echo ""
        
        if [ -f "./data/gh-pitfall-scraper.db" ]; then
            echo "数据库大小: $(du -h ./data/gh-pitfall-scraper.db | cut -f1)"
            echo "数据库权限: $(ls -l ./data/gh-pitfall-scraper.db | cut -d' ' -f1)"
            echo "文件锁定状态:"
            lsof ./data/gh-pitfall-scraper.db 2>/dev/null || echo "无锁定进程"
            echo ""
            
            echo "数据库完整性检查:"
            sqlite3 ./data/gh-pitfall-scraper.db "PRAGMA integrity_check;" 2>/dev/null || echo "完整性检查失败"
            echo ""
            
            echo "数据库表统计:"
            sqlite3 ./data/gh-pitfall-scraper.db "SELECT name, COUNT(*) as count FROM sqlite_master s JOIN pragma_table_info(s.name) GROUP BY name;" 2>/dev/null || echo "无法获取表统计"
            echo ""
        else
            echo "❌ 数据库文件不存在"
            echo ""
        fi
        
        # 配置文件
        echo "⚙️  配置文件状态:"
        if [ -f "./config.yaml" ]; then
            echo "配置文件存在: ./config.yaml"
            echo "文件大小: $(wc -l ./config.yaml | cut -d' ' -f1) 行"
            echo "最后修改: $(stat -c %y ./config.yaml)"
        else
            echo "❌ 配置文件不存在"
        fi
        echo ""
        
        # 进程信息
        echo "🔍 进程信息:"
        echo "gh-pitfall-scraper 进程:"
        ps aux | grep gh-pitfall-scraper | grep -v grep || echo "无相关进程"
        echo ""
        
        # 日志分析
        echo "📄 日志分析:"
        if [ -d "./logs" ]; then
            echo "最近的错误日志 (最近 20 行):"
            find ./logs -name "*.log" -exec tail -n 20 {} \; 2>/dev/null | grep -i error | tail -10
        else
            echo "日志目录不存在"
        fi
        echo ""
        
        echo "诊断报告生成完成: $report_file"
        
    } > "$report_file"
    
    echo "📄 详细诊断报告已生成: $report_file"
    echo "请将此报告发送给技术支持团队"
}

generate_diagnosis_report
```

## 常见问题及解决方案

### 1. 数据库连接问题

#### 问题1: 数据库文件权限错误
```bash
# 症状
Error: permission denied for database file

# 诊断
ls -la data/gh-pitfall-scraper.db

# 解决方案
chmod 644 data/gh-pitfall-scraper.db
chown $USER:$USER data/gh-pitfall-scraper.db

# 预防措施
# 1. 确保数据目录有正确的权限
# 2. 不要以 root 用户运行应用
# 3. 定期检查文件权限
```

#### 问题2: 数据库被锁定
```bash
# 症状
Error: database is locked

# 诊断
lsof data/gh-pitfall-scraper.db
ps aux | grep gh-pitfall-scraper

# 解决方案
# 1. 停止所有相关进程
pkill -f gh-pitfall-scraper

# 2. 等待释放锁
sleep 5

# 3. 重启应用
./gh-pitfall-scraper --config config.yaml

# 4. 如果仍然锁定，强制清除
rm data/gh-pitfall-scraper.db-wal data/gh-pitfall-scraper.db-shm 2>/dev/null
```

#### 问题3: 磁盘空间不足
```bash
# 症状
Error: no space left on device

# 诊断
df -h
du -sh data/

# 解决方案
# 1. 清理临时文件
find /tmp -name "*gh-pitfall*" -delete

# 2. 清理旧备份
find backups/ -name "*.db" -mtime +30 -delete

# 3. 压缩数据库
./database/db-manager.sh vacuum

# 4. 清理日志文件
find logs/ -name "*.log" -mtime +7 -delete

# 预防措施
# 1. 设置磁盘空间监控
# 2. 定期清理临时文件
# 3. 监控数据库大小增长
```

### 2. 性能问题

#### 问题4: 查询响应缓慢
```bash
# 症状
Queries taking too long to execute

# 诊断
./database/db-manager.sh benchmark
./database/db-manager.sh analyze-queries

# 解决方案
# 1. 重建索引
./database/db-manager.sh reindex

# 2. 更新统计信息
./database/db-manager.sh analyze

# 3. 清理碎片
./database/db-manager.sh vacuum

# 4. 调整缓存大小
# 在 config.yaml 中增加 cache_size
```

#### 问题5: 连接池耗尽
```bash
# 症状
Error: connection pool exhausted

# 诊断
./database/db-manager.sh pool-stats

# 解决方案
# 1. 增加连接池大小
# 在 config.yaml 中调整:
# database.connection_pool.max_open_conns: 50

# 2. 减少连接超时时间
# database.connection_pool.conn_max_lifetime: "300s"

# 3. 优化长时间运行的查询
# 查看慢查询日志并优化
```

#### 问题6: 内存使用过高
```bash
# 症状
System running out of memory

# 诊断
free -h
top -p $(pgrep gh-pitfall-scraper)

# 解决方案
# 1. 减少缓存大小
# database.cache.size: 500

# 2. 降低并发数
# scraping.concurrency.max_workers: 2

# 3. 增加批处理大小但减少频率
# database.performance.batch_size: 500
```

### 3. 数据问题

#### 问题7: 数据损坏
```bash
# 症状
Database corruption detected

# 诊断
./database/db-manager.sh integrity-check

# 解决方案
# 1. 从备份恢复
./database/db-manager.sh restore backups/latest-backup.db

# 2. 如果没有备份，尝试修复
sqlite3 data/gh-pitfall-scraper.db ".recover" > recovered.sql
sqlite3 data/gh-pitfall-scraper.db < recovered.sql

# 预防措施
# 1. 定期备份
# 2. 使用 WAL 模式
# 3. 避免强制关机
```

#### 问题8: 重复数据
```bash
# 症状
Duplicate records found

# 诊断
sqlite3 data/gh-pitfall-scraper.db "SELECT COUNT(*), title FROM issues GROUP BY title HAVING COUNT(*) > 1;"

# 解决方案
# 1. 启用自动去重
# deduplication.enabled: true

# 2. 手动清理重复数据
sqlite3 data/gh-pitfall-scraper.db << 'EOF'
DELETE FROM issues WHERE id NOT IN (
    SELECT MIN(id) FROM issues GROUP BY title, repository_id
);
EOF

# 预防措施
# 1. 启用内容哈希去重
# 2. 定期检查重复数据
# 3. 设置唯一性约束
```

#### 问题9: 数据不一致
```bash
# 症状
Data inconsistency detected

# 诊断
./database/db-manager.sh validate-data

# 解决方案
# 1. 修复外键约束
sqlite3 data/gh-pitfall-scraper.db "PRAGMA foreign_key_check;"

# 2. 同步统计信息
./database/db-manager.sh sync-stats

# 预防措施
# 1. 启用外键约束
# 2. 使用事务处理
# 3. 定期数据验证
```

## 错误代码说明

### 1. 应用程序错误代码

| 错误代码 | 含义 | 解决方案 |
|---------|------|----------|
| 1001 | 数据库连接失败 | 检查数据库文件权限和磁盘空间 |
| 1002 | 数据库锁定 | 停止相关进程并重启 |
| 1003 | SQL语法错误 | 检查SQL语句和表结构 |
| 1004 | 内存不足 | 减少缓存大小或增加系统内存 |
| 1005 | 磁盘空间不足 | 清理临时文件和旧备份 |
| 1006 | 网络连接失败 | 检查网络配置和防火墙设置 |
| 1007 | 配置文件错误 | 验证YAML语法和配置项 |
| 1008 | 权限不足 | 检查文件和目录权限 |
| 1009 | 服务不可用 | 重启服务或检查依赖 |
| 1010 | 超时错误 | 调整超时参数或优化查询 |

### 2. 数据库错误代码

| 错误代码 | 含义 | 解决方案 |
|---------|------|----------|
| SQLITE_CORRUPT | 数据库损坏 | 从备份恢复或修复数据库 |
| SQLITE_LOCKED | 数据库被锁定 | 停止其他进程并等待 |
| SQLITE_BUSY | 数据库忙 | 增加busy_timeout或优化查询 |
| SQLITE_FULL | 数据库满 | 清理数据或增加磁盘空间 |
| SQLITE_CANTOPEN | 无法打开数据库 | 检查文件路径和权限 |
| SQLITE_TOOBIG | SQL语句太长 | 分拆SQL语句或调整限制 |
| SQLITE_CONSTRAINT | 约束违反 | 检查数据完整性和约束 |
| SQLITE_MISMATCH | 数据类型不匹配 | 检查数据类型和转换 |
| SQLITE_MISUSE | API使用错误 | 检查API调用方式 |
| SQLITE_NOLFS | 大文件支持不可用 | 升级系统或使用其他数据库 |

### 3. 网络错误代码

| 错误代码 | 含义 | 解决方案 |
|---------|------|----------|
| 2001 | GitHub API限制 | 减少请求频率或升级API计划 |
| 2002 | 认证失败 | 检查GitHub Token |
| 2003 | 网络超时 | 调整超时参数或检查网络 |
| 2004 | 服务不可用 | 稍后重试或检查服务状态 |
| 2005 | 代理错误 | 检查代理设置 |

## 性能问题排查

### 1. 查询性能分析

#### 分析慢查询
```bash
#!/bin/bash
# analyze-slow-queries.sh

echo "🔍 分析慢查询性能"
echo "=================="

# 启用查询日志
echo "启用查询日志..."
sqlite3 data/gh-pitfall-scraper.db "PRAGMA analysis_limit=1000;"

# 运行分析
./database/db-manager.sh analyze-queries

# 查看执行计划
sqlite3 data/gh-pitfall-scraper.db << 'EOF'
EXPLAIN QUERY PLAN SELECT * FROM issues WHERE created_at > '2024-01-01';
EOF

# 建议优化
echo ""
echo "💡 性能优化建议:"
echo "1. 为常用查询字段创建索引"
echo "2. 使用 LIMIT 限制结果集大小"
echo "3. 避免 SELECT *，只查询需要的字段"
echo "4. 使用预编译语句"
echo "5. 定期执行 VACUUM 和 ANALYZE"
```

#### 索引优化
```bash
#!/bin/bash
# optimize-indexes.sh

echo "🔧 索引优化"
echo "==========="

# 查看索引使用情况
sqlite3 data/gh-pitfall-scraper.db "PRAGMA index_list(issues);"
sqlite3 data/gh-pitfall-scraper.db "PRAGMA index_info(idx_issues_created_at);"

# 重建未使用的索引
sqlite3 data/gh-pitfall-scraper.db << 'EOF'
-- 重建索引
REINDEX;

-- 分析索引使用
ANALYZE;

-- 查看索引统计
SELECT name, pgsz, sz, usage 
FROM dbstat 
WHERE name LIKE 'idx_%';
EOF

echo ""
echo "📊 索引优化完成"
```

### 2. 缓存性能分析

#### 缓存命中率分析
```bash
#!/bin/bash
# analyze-cache.sh

echo "💾 缓存性能分析"
echo "==============="

# 查看缓存统计
./database/db-manager.sh cache-stats

# 调整缓存大小建议
echo ""
echo "💡 缓存优化建议:"
echo "1. 如果缓存命中率 < 80%，增加缓存大小"
echo "2. 如果内存使用过高，减少缓存大小"
echo "3. 如果有大量相同查询，增加TTL"
echo "4. 定期清理过期缓存"
```

## 数据问题处理

### 1. 数据恢复

#### 从备份恢复
```bash
#!/bin/bash
# restore-from-backup.sh

BACKUP_FILE="$1"
TARGET_DB="data/gh-pitfall-scraper.db"

if [ -z "$BACKUP_FILE" ]; then
    echo "用法: $0 <backup_file>"
    echo "可用备份:"
    ls -la backups/*.db 2>/dev/null || echo "无备份文件"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    echo "错误: 备份文件不存在: $BACKUP_FILE"
    exit 1
fi

echo "🔄 从备份恢复数据库"
echo "备份文件: $BACKUP_FILE"
echo "目标文件: $TARGET_DB"

# 创建当前数据库的备份
if [ -f "$TARGET_DB" ]; then
    cp "$TARGET_DB" "${TARGET_DB}.backup.$(date +%Y%m%d_%H%M%S)"
    echo "✅ 已备份当前数据库"
fi

# 恢复备份
cp "$BACKUP_FILE" "$TARGET_DB"
echo "✅ 数据库恢复完成"

# 验证恢复
if ./database/db-manager.sh health; then
    echo "✅ 数据库恢复验证成功"
else
    echo "❌ 数据库恢复验证失败"
    exit 1
fi

echo "🎉 数据库恢复完成"
```

### 2. 数据迁移

#### SQLite 到 PostgreSQL
```bash
#!/bin/bash
# migrate-sqlite-to-postgres.sh

SQLITE_DB="data/gh-pitfall-scraper.db"
PG_HOST="localhost"
PG_DB="gh_pitfall_scraper"
PG_USER="postgres"

echo "🔄 数据迁移：SQLite -> PostgreSQL"
echo "=================================="

# 1. 导出SQLite数据
echo "📤 导出SQLite数据..."
sqlite3 $SQLITE_DB .dump > sqlite_export.sql

# 2. 转换SQL语句
echo "🔧 转换SQL语句..."
sed 's/INTEGER PRIMARY KEY AUTOINCREMENT/SERIAL PRIMARY KEY/g' sqlite_export.sql > postgres_import.sql
sed 's/AUTOINCREMENT/AUTO_INCREMENT/g' postgres_import.sql > temp.sql
mv temp.sql postgres_import.sql

# 3. 导入PostgreSQL
echo "📥 导入PostgreSQL..."
psql -h $PG_HOST -U $PG_USER -d $PG_DB -f postgres_import.sql

# 4. 验证迁移
echo "✅ 验证迁移结果..."
psql -h $PG_HOST -U $PG_USER -d $PG_DB -c "SELECT COUNT(*) FROM issues;"

echo "🎉 数据迁移完成"
```

## 配置问题诊断

### 1. 配置文件验证
```bash
#!/bin/bash
# validate-config.sh

CONFIG_FILE="${1:-config.yaml}"

echo "⚙️  配置文件验证"
echo "=================="

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

echo "📄 配置文件: $CONFIG_FILE"

# 1. YAML语法检查
if command -v yq > /dev/null 2>&1; then
    echo "🔍 检查YAML语法..."
    if yq eval . "$CONFIG_FILE" > /dev/null; then
        echo "✅ YAML语法正确"
    else
        echo "❌ YAML语法错误"
        yq eval . "$CONFIG_FILE"
    fi
else
    echo "⚠️  yq命令未安装，跳过YAML语法检查"
fi

# 2. 必需配置项检查
echo "🔍 检查必需配置项..."

required_fields=(
    "github_token"
    "database.type"
    "repos"
    "keywords"
)

for field in "${required_fields[@]}"; do
    if yq eval ".$field" "$CONFIG_FILE" | grep -q "."; then
        echo "✅ $field 已配置"
    else
        echo "❌ $field 未配置"
    fi
done

# 3. 配置合理性检查
echo "🔍 检查配置合理性..."

# 检查数据库路径
db_type=$(yq eval ".database.type" "$CONFIG_FILE")
if [ "$db_type" = "sqlite" ]; then
    db_file=$(yq eval ".database.sqlite.file_path" "$CONFIG_FILE")
    if [ -n "$db_file" ]; then
        echo "✅ SQLite数据库路径已配置: $db_file"
    else
        echo "❌ SQLite数据库路径未配置"
    fi
elif [ "$db_type" = "postgresql" ]; then
    pg_host=$(yq eval ".database.postgresql.host" "$CONFIG_FILE")
    if [ -n "$pg_host" ]; then
        echo "✅ PostgreSQL配置已配置"
    else
        echo "❌ PostgreSQL配置不完整"
    fi
fi

echo ""
echo "📊 配置文件验证完成"
```

## 环境问题排查

### 1. 系统资源检查
```bash
#!/bin/bash
# check-system-resources.sh

echo "💻 系统资源检查"
echo "================"

# CPU信息
echo "🖥️  CPU信息:"
lscpu | grep "Model name" | sed 's/Model name:\s*//'

# 内存信息
echo ""
echo "💾 内存信息:"
free -h

# 磁盘信息
echo ""
echo "💽 磁盘信息:"
df -h

# 系统负载
echo ""
echo "📈 系统负载:"
uptime

# 进程信息
echo ""
echo "🔍 相关进程:"
ps aux | grep -E "(gh-pitfall|sqlite|postgres)" | grep -v grep

# 网络信息
echo ""
echo "🌐 网络连接:"
netstat -tuln 2>/dev/null | grep -E "(5432|8080|3000)" || ss -tuln | grep -E "(5432|8080|3000)"

# 系统限制
echo ""
echo "🔒 系统限制:"
echo "打开文件数限制: $(ulimit -n)"
echo "进程数限制: $(ulimit -u)"
echo "内存限制: $(ulimit -m)"

echo ""
echo "📊 系统资源检查完成"
```

## 应急处理流程

### 1. 应急响应脚本
```bash
#!/bin/bash
# emergency-response.sh

LOG_FILE="emergency-$(date +%Y%m%d_%H%M%S).log"

log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

emergency_backup() {
    log_message "INFO: 创建紧急备份"
    if [ -f "data/gh-pitfall-scraper.db" ]; then
        cp "data/gh-pitfall-scraper.db" "emergency-backup-$(date +%Y%m%d_%H%M%S).db"
        log_message "INFO: 紧急备份创建完成"
    else
        log_message "ERROR: 数据库文件不存在，无法创建备份"
    fi
}

stop_services() {
    log_message "INFO: 停止相关服务"
    pkill -f gh-pitfall-scraper
    sleep 2
    log_message "INFO: 服务已停止"
}

check_recovery_options() {
    log_message "INFO: 检查恢复选项"
    
    echo "恢复选项:"
    echo "1. 从备份恢复"
    echo "2. 修复数据库"
    echo "3. 重建数据库"
    echo "4. 什么都不做"
    
    read -p "请选择操作 (1-4): " choice
    
    case $choice in
        1)
            echo "可用备份文件:"
            ls -la backups/*.db 2>/dev/null || echo "无备份文件"
            read -p "请输入备份文件路径: " backup_file
            if [ -f "$backup_file" ]; then
                cp "$backup_file" "data/gh-pitfall-scraper.db"
                log_message "INFO: 从备份恢复完成"
            else
                log_message "ERROR: 备份文件不存在"
            fi
            ;;
        2)
            log_message "INFO: 尝试修复数据库"
            sqlite3 "data/gh-pitfall-scraper.db" ".recover" > recovered.sql
            sqlite3 "data/gh-pitfall-scraper.db" < recovered.sql
            log_message "INFO: 数据库修复尝试完成"
            ;;
        3)
            log_message "INFO: 重建数据库"
            rm -f "data/gh-pitfall-scraper.db"
            ./gh-pitfall-scraper --db-only
            log_message "INFO: 数据库重建完成"
            ;;
        *)
            log_message "INFO: 用户选择不恢复"
            ;;
    esac
}

verify_recovery() {
    log_message "INFO: 验证恢复结果"
    
    if ./database/db-manager.sh health; then
        log_message "INFO: 数据库健康检查通过"
        echo "✅ 恢复成功！"
    else
        log_message "ERROR: 数据库健康检查失败"
        echo "❌ 恢复失败，请检查日志"
    fi
}

# 主流程
main() {
    log_message "INFO: 开始应急处理流程"
    
    echo "🚨 应急处理流程启动"
    echo "===================="
    
    emergency_backup
    stop_services
    check_recovery_options
    verify_recovery
    
    log_message "INFO: 应急处理流程完成"
    
    echo ""
    echo "📄 应急处理日志: $LOG_FILE"
    echo "📞 如需进一步帮助，请联系技术支持团队"
}

# 执行应急处理
main
```

## 预防措施

### 1. 定期健康检查脚本
```bash
#!/bin/bash
# health-check-scheduler.sh

HEALTH_CHECK_INTERVAL=3600  # 1小时
LOG_FILE="health-check.log"

perform_health_check() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "[$timestamp] 开始健康检查" >> "$LOG_FILE"
    
    # 数据库健康检查
    if ./database/db-manager.sh health > /dev/null 2>&1; then
        echo "[$timestamp] ✅ 数据库健康" >> "$LOG_FILE"
    else
        echo "[$timestamp] ❌ 数据库异常" >> "$LOG_FILE"
        # 发送告警
        echo "数据库健康检查失败" | mail -s "DB Alert" admin@example.com
    fi
    
    # 磁盘空间检查
    DISK_USAGE=$(df . | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ $DISK_USAGE -gt 80 ]; then
        echo "[$timestamp] ⚠️  磁盘使用率过高: ${DISK_USAGE}%" >> "$LOG_FILE"
        echo "磁盘使用率过高: ${DISK_USAGE}%" | mail -s "Disk Alert" admin@example.com
    fi
    
    # 内存使用检查
    MEMORY_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
    if [ $MEMORY_USAGE -gt 90 ]; then
        echo "[$timestamp] ⚠️  内存使用率过高: ${MEMORY_USAGE}%" >> "$LOG_FILE"
    fi
    
    # 数据库大小检查
    DB_SIZE=$(./database/db-manager.sh size-bytes)
    if [ $DB_SIZE -gt 5368709120 ]; then  # 5GB
        echo "[$timestamp] ⚠️  数据库大小超过5GB" >> "$LOG_FILE"
    fi
    
    echo "[$timestamp] 健康检查完成" >> "$LOG_FILE"
}

# 调度健康检查
schedule_health_checks() {
    echo "开始健康检查调度..."
    
    while true; do
        perform_health_check
        sleep $HEALTH_CHECK_INTERVAL
    done
}

# 根据参数执行不同操作
case "$1" in
    once)
        perform_health_check
        ;;
    schedule)
        schedule_health_checks
        ;;
    *)
        echo "使用方法: $0 {once|schedule}"
        exit 1
        ;;
esac
```

### 2. 自动维护脚本
```bash
#!/bin/bash
# auto-maintenance.sh

perform_maintenance() {
    local date=$(date '+%Y-%m-%d')
    local log_file="maintenance-$date.log"
    
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始自动维护" | tee "$log_file"
    
    # 1. 数据库维护
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行数据库维护" | tee -a "$log_file"
    ./database/db-manager.sh vacuum
    ./database/db-manager.sh reindex
    ./database/db-manager.sh analyze
    
    # 2. 清理过期数据
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 清理过期数据" | tee -a "$log_file"
    ./database/db-manager.sh cleanup
    
    # 3. 创建备份
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 创建备份" | tee -a "$log_file"
    ./database/db-manager.sh backup "backups/auto-backup-$date.db"
    
    # 4. 清理旧文件
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 清理旧文件" | tee -a "$log_file"
    find logs/ -name "*.log" -mtime +7 -delete
    find backups/ -name "*.db" -mtime +30 -delete
    
    # 5. 性能检查
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行性能检查" | tee -a "$log_file"
    ./database/db-manager.sh benchmark
    
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 自动维护完成" | tee -a "$log_file"
}

# 添加到crontab
setup_crontab() {
    echo "设置自动维护计划任务..."
    
    # 每天凌晨2点执行维护
    (crontab -l 2>/dev/null; echo "0 2 * * * $(pwd)/auto-maintenance.sh") | crontab -
    
    echo "✅ 自动维护计划已设置"
    echo "📋 查看当前crontab:"
    crontab -l
}

# 根据参数执行不同操作
case "$1" in
    now)
        perform_maintenance
        ;;
    setup)
        setup_crontab
        ;;
    *)
        echo "使用方法: $0 {now|setup}"
        exit 1
        ;;
esac
```

## 总结

本故障排除指南涵盖了 gh-pitfall-scraper 数据库的各种常见问题和解决方案：

1. **快速诊断**: 使用一键诊断脚本快速识别问题
2. **问题分类**: 按类型组织问题和解决方案
3. **应急处理**: 完善的应急响应流程
4. **预防措施**: 定期检查和维护机制

记住：**预防优于治疗，定期维护是避免故障的最佳方式**。

如遇到本文档未涵盖的问题，请：
1. 收集详细的错误信息
2. 运行诊断脚本生成报告
3. 查看相关日志文件
4. 联系技术支持团队