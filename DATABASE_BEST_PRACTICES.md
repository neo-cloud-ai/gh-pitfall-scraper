---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 304402201e5b8a2f898d62de0ce8420003a3635eb9e79eb5a80135f06425ccc7fb6221ef0220486519e23c4a12fc23a69d3809bf417e3b4f07cb7b84f3be8efd0d390cb12b16
    ReservedCode2: 3045022100d2ad0fbc43988fcc0191e66d5279054b294fdfece9692c8c7c6804003190e1af022062a49da59793c95526d74fcbe51a1d53fed711ad42a3cc6cdae5f5217cb7d0c8
---

# gh-pitfall-scraper 数据库功能演示和最佳实践

## 目录

1. [功能演示](#功能演示)
2. [配置演示](#配置演示)
3. [操作演示](#操作演示)
4. [性能演示](#性能演示)
5. [故障处理演示](#故障处理演示)
6. [最佳实践](#最佳实践)
7. [生产环境部署](#生产环境部署)
8. [监控和维护](#监控和维护)

## 功能演示

### 1. 基础功能演示

#### 1.1 快速启动演示
```bash
# 演示环境准备
echo "🚀 开始 gh-pitfall-scraper 数据库功能演示"
echo "============================================="

# 检查环境
echo "📋 检查系统环境..."
go version
sqlite3 --version || echo "SQLite3 未安装，但 Go 内置支持"

# 创建演示目录
mkdir -p demo/{data,logs,backups,output}
cd demo

# 复制配置文件
cp ../config-database-example.yaml config.yaml

echo "✅ 环境准备完成"
echo "📁 当前目录: $(pwd)"
echo "🗂️  目录结构:"
ls -la
```

#### 1.2 数据库初始化演示
```bash
echo ""
echo "🗄️  数据库初始化演示"
echo "========================"

# 方法一：使用主程序初始化
echo "📝 方法一：使用主程序初始化数据库"
echo "./gh-pitfall-scraper --db-only --config config.yaml"
./gh-pitfall-scraper --db-only --config config.yaml

echo ""
echo "📊 查看数据库文件"
ls -la data/

echo ""
echo "🏥 检查数据库健康状态"
./gh-pitfall-scraper --health --config config.yaml

echo ""
echo "📈 查看数据库统计信息"
./gh-pitfall-scraper --stats --config config.yaml
```

#### 1.3 数据库管理工具演示
```bash
echo ""
echo "🔧 数据库管理工具演示"
echo "========================="

# 赋予执行权限
chmod +x ../database/db-manager.sh

echo "📋 查看管理工具帮助"
../database/db-manager.sh --help

echo ""
echo "🔍 查看数据库状态"
../database/db-manager.sh status

echo ""
echo "ℹ️  查看数据库详细信息"
../database/db-manager.sh info

echo ""
echo "🧪 测试数据库连接"
../database/db-manager.sh test
```

### 2. 数据操作演示

#### 2.1 运行爬虫并存储数据
```bash
echo ""
echo "🕷️  爬虫功能演示"
echo "=================="

# 编辑配置文件，设置GitHub Token
sed -i 's/ghp_xxx/ghp_your_token_here/' config.yaml

echo "🔄 运行爬虫（包含数据库存储）"
echo "./gh-pitfall-scraper --config config.yaml --max-issues=10"

# 运行爬虫（模拟）
echo "✅ 爬虫执行完成，数据已存储到数据库"
```

#### 2.2 数据库查询演示
```bash
echo ""
echo "🔍 数据库查询演示"
echo "=================="

# 使用命令行工具查询
echo "📊 获取总体统计信息"
./gh-pitfall-scraper --stats --config config.yaml

echo ""
echo "🏢 获取仓库统计信息"
echo "./gh-pitfall-scraper --stats --repo=vllm-project/vllm"

echo ""
echo "📅 获取最近数据"
echo "./gh-pitfall-scraper --stats --days=7"

# 直接SQL查询演示
echo ""
echo "🗃️  直接SQL查询演示"
sqlite3 data/gh-pitfall-scraper.db << 'EOF'
.mode column
.headers on
.tables
SELECT name FROM sqlite_master WHERE type='table';
SELECT 'repositories' as table_name, COUNT(*) as count FROM repositories
UNION ALL
SELECT 'issues', COUNT(*) FROM issues;
.quit
EOF
```

### 3. 备份恢复演示

#### 3.1 创建备份
```bash
echo ""
echo "💾 备份功能演示"
echo "================="

echo "🔄 创建数据库备份"
../database/db-manager.sh backup

echo ""
echo "📁 查看备份文件"
ls -la backups/

echo ""
echo "🗜️  创建压缩备份"
../database/db-manager.sh backup --compress demo-backup-compressed.db

echo ""
echo "📋 查看所有备份文件"
ls -la backups/
```

#### 3.2 恢复演示
```bash
echo ""
echo "🔄 恢复功能演示"
echo "================="

echo "⚠️  注意：这是一个演示，不会实际覆盖数据"

echo "📋 列出可用的备份文件"
ls -la backups/

echo "💡 恢复命令示例："
echo "../database/db-manager.sh restore backups/demo-backup-$(date +%Y%m%d_%H%M%S).db"
echo ""
echo "🔍 恢复前检查"
../database/db-manager.sh status

echo "⚠️  实际恢复操作（演示）"
echo "../database/db-manager.sh --dry-run restore backups/latest-backup.db"
```

## 配置演示

### 1. 不同环境配置演示

#### 1.1 开发环境配置
```bash
echo ""
echo "🛠️  开发环境配置演示"
echo "======================"

cat > config-dev.yaml << 'EOF'
# 开发环境配置示例
app:
  name: "gh-pitfall-scraper-dev"
  log_level: "debug"                    # 详细日志
  data_dir: "./dev-data"

database:
  type: "sqlite"
  sqlite:
    file_path: "./dev-data/dev.db"
    enable_wal: true
    cache_size: -5000                   # 5MB 缓存
    synchronous: "OFF"                  # 最快速度
  
  connection_pool:
    max_open_conns: 5                   # 少量连接
    max_idle_conns: 2
    
  cache:
    enabled: true
    size: 500                           # 小缓存
    ttl: "1800s"                        # 30分钟
    
  cleanup:
    enabled: false                      # 开发环境不清理
  
  backup:
    enabled: false                      # 开发环境不备份

github_token: "${GITHUB_TOKEN}"

repos:
  - owner: vllm-project
    name: vllm

keywords:
  - "performance"
  - "bug"
EOF

echo "✅ 开发环境配置创建完成: config-dev.yaml"
echo "🔧 特点："
echo "  - 详细日志 (debug)"
echo "  - SQLite 数据库"
echo "  - 关闭清理和备份"
echo "  - 小缓存设置"
```

#### 1.2 生产环境配置
```bash
echo ""
echo "🏢 生产环境配置演示"
echo "====================="

cat > config-prod.yaml << 'EOF'
# 生产环境配置示例
app:
  name: "gh-pitfall-scraper"
  log_level: "info"
  data_dir: "/var/lib/gh-pitfall-scraper"

database:
  type: "postgresql"
  postgresql:
    host: "localhost"
    port: 5432
    user: "gh_scraper"
    password: "${DB_PASSWORD}"
    dbname: "gh_pitfall_scraper"
    sslmode: "require"
  
  connection_pool:
    max_open_conns: 25                  # 根据负载调整
    max_idle_conns: 5
    conn_max_lifetime: "300s"
    
  cache:
    enabled: true
    size: 5000                          # 大缓存
    ttl: "7200s"                        # 2小时
    
  cleanup:
    enabled: true
    interval: "24h"
    max_age: "720h"                     # 30天保留
    
  backup:
    enabled: true
    interval: "12h"
    retention_days: 30
    path: "/var/backups/gh-pitfall-scraper"
    compress: true
    
  monitoring:
    enabled: true
    log_slow_queries: true
    slow_query_threshold: 1000

github_token: "${GITHUB_TOKEN}"

repos:
  - owner: vllm-project
    name: vllm
  - owner: sgl-project
    name: sglang
  - owner: NVIDIA
    name: TensorRT-LLM

keywords:
  - "performance regression"
  - "memory leak"
  - "CUDA"
  - "distributed"
  - "training"
EOF

echo "✅ 生产环境配置创建完成: config-prod.yaml"
echo "🏗️  特点："
echo "  - PostgreSQL 数据库"
echo "  - SSL 加密连接"
echo "  - 自动清理和备份"
echo "  - 监控和告警"
echo "  - 大缓存设置"
```

### 2. 性能调优配置演示

#### 2.1 高并发配置
```bash
echo ""
echo "⚡ 高并发配置演示"
echo "=================="

cat > config-high-concurrency.yaml << 'EOF'
# 高并发场景配置
database:
  connection_pool:
    max_open_conns: 50                  # 大量连接
    max_idle_conns: 20
    conn_max_lifetime: "600s"
    
  cache:
    enabled: true
    size: 10000                         # 大缓存
    ttl: "3600s"
    
  performance:
    batch_size: 2000                    # 大批量处理
    query_timeout: 60
    max_connections: 50
    enable_prepared_statements: true
    
scraping:
  concurrency:
    max_workers: 20                     # 大量工作线程
    requests_per_second: 50             # 高请求频率
EOF

echo "✅ 高并发配置创建完成: config-high-concurrency.yaml"
```

#### 2.2 低资源配置
```bash
echo ""
echo "💻 低资源环境配置演示"
echo "======================"

cat > config-low-resource.yaml << 'EOF'
# 低资源环境配置
database:
  connection_pool:
    max_open_conns: 5                   # 少量连接
    max_idle_conns: 2
    
  cache:
    enabled: true
    size: 100                           # 小缓存
    ttl: "600s"                         # 短时间
    
  performance:
    batch_size: 100                     # 小批量处理
    query_timeout: 15
    
scraping:
  concurrency:
    max_workers: 2                      # 少量工作线程
    requests_per_second: 2              # 低请求频率
    timeout: 15                         # 短超时
EOF

echo "✅ 低资源配置创建完成: config-low-resource.yaml"
```

## 操作演示

### 1. 迁移操作演示

#### 1.1 创建和执行迁移
```bash
echo ""
echo "🔄 数据库迁移演示"
echo "=================="

echo "📋 查看迁移工具帮助"
go run ../database/migration.go --help

echo ""
echo "🆕 创建新迁移"
go run ../database/migration.go create add_performance_indexes "添加性能优化索引"

echo ""
echo "📁 查看迁移文件"
ls -la ../database/migrations/

echo ""
echo "🔍 查看迁移状态"
go run ../database/migration.go status

echo ""
echo "⚡ 执行迁移（模拟）"
echo "go run ../database/migration.go migrate"

echo ""
echo "📜 查看迁移历史"
go run ../database/migration.go history
```

### 2. 维护操作演示

#### 2.1 数据库维护
```bash
echo ""
echo "🔧 数据库维护演示"
echo "=================="

echo "🧹 清理过期数据"
../database/db-manager.sh cleanup

echo ""
echo "🗜️  压缩数据库"
../database/db-manager.sh vacuum

echo ""
echo "🔍 重建索引"
../database/db-manager.sh reindex

echo ""
echo "📊 分析数据库"
../database/db-manager.sh analyze

echo ""
echo "🏥 健康检查"
../database/db-manager.sh health
```

#### 2.2 性能监控
```bash
echo ""
echo "📈 性能监控演示"
echo "================"

echo "💾 查看数据库大小"
../database/db-manager.sh size

echo ""
echo "⚡ 运行性能测试"
../database/db-manager.sh benchmark

echo ""
echo "📊 查看连接池状态"
echo "../database/db-manager.sh pool-stats"

echo ""
echo "🐌 查看慢查询"
echo "../database/db-manager.sh slow-queries"
```

## 故障处理演示

### 1. 常见问题诊断

#### 1.1 连接问题诊断
```bash
echo ""
echo "🔍 连接问题诊断演示"
echo "==================="

echo "🔍 检查数据库文件权限"
ls -la data/

echo ""
echo "🔍 检查磁盘空间"
df -h

echo ""
echo "🔍 检查数据库进程"
lsof data/gh-pitfall-scraper.db || echo "没有进程在使用数据库文件"

echo ""
echo "🧪 测试数据库连接"
../database/db-manager.sh test

echo ""
echo "🔄 重启数据库连接"
echo "../database/db-manager.sh restart"
```

#### 1.2 性能问题诊断
```bash
echo ""
echo "⚡ 性能问题诊断演示"
echo "==================="

echo "📊 查看数据库统计"
../database/db-manager.sh stats

echo ""
echo "📈 查看连接池使用情况"
echo "../database/db-manager.sh pool-usage"

echo ""
echo "🐌 分析慢查询"
echo "../database/db-manager.sh analyze-queries"

echo ""
echo "🔍 检查索引使用情况"
echo "../database/db-manager.sh index-stats"

echo ""
echo "💾 检查缓存命中率"
echo "../database/db-manager.sh cache-stats"
```

### 2. 数据恢复演示

#### 2.1 完整性检查
```bash
echo ""
echo "🔍 数据完整性检查演示"
echo "====================="

echo "🔍 检查数据库完整性"
../database/db-manager.sh integrity-check

echo ""
echo "🔍 验证数据一致性"
echo "../database/db-manager.sh validate-data"

echo ""
echo "🔍 检查外键约束"
echo "../database/db-manager.sh check-foreign-keys"
```

#### 2.2 恢复操作
```bash
echo ""
echo "🔄 数据恢复演示"
echo "================"

echo "📋 列出可用备份"
ls -la backups/

echo ""
echo "💡 恢复前验证"
../database/db-manager.sh verify-backup backups/latest-backup.db

echo ""
echo "🔄 执行恢复（模拟）"
echo "../database/db-manager.sh restore --verify backups/latest-backup.db"

echo ""
echo "✅ 恢复后验证"
../database/db-manager.sh health
```

## 最佳实践

### 1. 配置最佳实践

#### 1.1 开发环境最佳实践
```bash
cat > BEST-PRACTICES.md << 'EOF'
# gh-pitfall-scraper 数据库最佳实践

## 🛠️ 开发环境最佳实践

### 配置文件管理
- 使用 `.gitignore` 排除配置文件和敏感数据
- 创建 `config-dev.yaml` 开发专用配置
- 使用环境变量管理敏感信息
- 定期更新依赖和安全补丁

### 数据库配置
```yaml
# 开发环境推荐配置
database:
  type: "sqlite"
  sqlite:
    enable_wal: true          # 启用并发支持
    cache_size: -10000        # 10MB 缓存
    synchronous: "NORMAL"     # 平衡性能和安全
  connection_pool:
    max_open_conns: 5         # 少量连接
  cache:
    enabled: true
    size: 1000                # 小缓存
  cleanup:
    enabled: false            # 开发环境不自动清理
  backup:
    enabled: false            # 开发环境不自动备份
```

### 调试技巧
- 启用详细日志: `log_level: "debug"`
- 启用 SQL 跟踪: `log_queries: true`
- 使用性能分析: `profiling: true`
- 定期清理测试数据

## 🏢 生产环境最佳实践

### 安全配置
- 启用 SSL 连接: `sslmode: "require"`
- 使用强密码策略
- 定期更新凭据
- 限制网络访问
- 启用审计日志

### 性能优化
```yaml
# 生产环境推荐配置
database:
  type: "postgresql"
  connection_pool:
    max_open_conns: 25        # 根据负载调整
    max_idle_conns: 5
  cache:
    enabled: true
    size: 5000                # 大缓存
  cleanup:
    enabled: true
    interval: "24h"
    max_age: "720h"           # 30天保留
  backup:
    enabled: true
    interval: "12h"
    retention_days: 30
    compress: true
```

### 监控告警
- 启用健康检查: `health_check_interval: "60s"`
- 监控慢查询: `slow_query_threshold: 1000`
- 设置告警阈值
- 定期生成报告

## 💾 备份策略

### 备份类型
1. **全量备份**: 每周执行，保留30天
2. **增量备份**: 每天执行，保留7天
3. **事务日志备份**: 实时或每小时

### 备份脚本示例
```bash
#!/bin/bash
# backup-strategy.sh

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# 全量备份 (周日)
if [ $(date +%w) -eq 0 ]; then
    ./database/db-manager.sh backup "$BACKUP_DIR/full_backup_$DATE.db"
    find $BACKUP_DIR -name "full_backup_*.db" -mtime +30 -delete
else
    # 增量备份
    ./database/db-manager.sh backup "$BACKUP_DIR/incremental_backup_$DATE.db"
    find $BACKUP_DIR -name "incremental_backup_*.db" -mtime +7 -delete
fi

# 验证备份完整性
./database/db-manager.sh verify-backup "$BACKUP_DIR/latest_backup.db"
```

## 🔧 维护计划

### 每日维护
- [ ] 健康检查
- [ ] 清理过期数据
- [ ] 更新统计信息
- [ ] 创建备份

### 每周维护
- [ ] 数据库优化
- [ ] 重建索引
- [ ] 压缩数据库
- [ ] 性能基准测试

### 每月维护
- [ ] 容量规划
- [ ] 安全审计
- [ ] 备份恢复测试
- [ ] 性能调优

## 🚨 故障处理

### 常见问题快速解决
1. **数据库锁定**: 重启应用或增加 `busy_timeout`
2. **性能下降**: 运行 `ANALYZE` 和重建索引
3. **磁盘空间不足**: 清理旧备份和压缩数据库
4. **连接超时**: 调整连接池参数

### 应急处理流程
1. 立即创建备份
2. 检查错误日志
3. 隔离问题范围
4. 应用修复方案
5. 验证修复结果
6. 总结经验教训

## 📊 性能调优

### 监控指标
- 连接池使用率
- 查询响应时间
- 缓存命中率
- 磁盘 I/O 性能
- 内存使用情况

### 调优步骤
1. 收集基线数据
2. 识别瓶颈
3. 调整配置参数
4. 测试验证
5. 持续监控

EOF

echo "✅ 最佳实践文档创建完成: BEST-PRACTICES.md"
```

## 生产环境部署

### 1. 部署检查清单
```bash
cat > DEPLOYMENT-CHECKLIST.md << 'EOF'
# 生产环境部署检查清单

## 🔒 安全检查
- [ ] GitHub Token 已设置且权限正确
- [ ] 数据库密码已设置且符合安全要求
- [ ] SSL 证书已配置（PostgreSQL）
- [ ] 防火墙规则已设置
- [ ] 文件权限已正确设置
- [ ] 敏感信息未硬编码在配置中

## 🗄️ 数据库配置
- [ ] 数据库已创建且权限正确
- [ ] 连接池参数已根据负载调整
- [ ] 缓存配置已优化
- [ ] 自动清理已启用
- [ ] 备份策略已配置
- [ ] 监控已启用

## 📊 性能优化
- [ ] 索引已优化
- [ ] 查询性能已测试
- [ ] 连接池大小已调整
- [ ] 缓存命中率已验证
- [ ] 存储空间已规划

## 🔧 运维准备
- [ ] 日志配置已优化
- [ ] 监控告警已设置
- [ ] 备份恢复流程已测试
- [ ] 维护计划已制定
- [ ] 应急响应流程已准备

## 🧪 测试验证
- [ ] 功能测试通过
- [ ] 性能测试通过
- [ ] 压力测试通过
- [ ] 故障恢复测试通过
- [ ] 安全测试通过

EOF

echo "✅ 部署检查清单创建完成: DEPLOYMENT-CHECKLIST.md"
```

## 监控和维护

### 1. 监控脚本
```bash
cat > monitor.sh << 'EOF'
#!/bin/bash
# 生产环境监控脚本

LOG_FILE="/var/log/gh-pitfall-scraper-monitor.log"
ALERT_EMAIL="admin@example.com"

log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOG_FILE
}

check_database_health() {
    if ! ./database/db-manager.sh health > /dev/null 2>&1; then
        log_message "ERROR: 数据库健康检查失败"
        echo "数据库健康检查失败" | mail -s "DB Alert" $ALERT_EMAIL
        return 1
    fi
    log_message "INFO: 数据库健康检查通过"
}

check_disk_space() {
    USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ $USAGE -gt 80 ]; then
        log_message "WARNING: 磁盘使用率过高: ${USAGE}%"
        echo "磁盘使用率过高: ${USAGE}%" | mail -s "Disk Alert" $ALERT_EMAIL
    fi
}

check_database_size() {
    SIZE=$(./database/db-manager.sh size-bytes)
    if [ $SIZE -gt 10737418240 ]; then  # 10GB
        log_message "WARNING: 数据库大小超过10GB"
        echo "数据库大小超过10GB" | mail -s "Size Alert" $ALERT_EMAIL
    fi
}

check_connection_pool() {
    USAGE=$(./database/db-manager.sh pool-usage)
    if [ $USAGE -gt 80 ]; then
        log_message "WARNING: 连接池使用率过高: ${USAGE}%"
        echo "连接池使用率过高: ${USAGE}%" | mail -s "Pool Alert" $ALERT_EMAIL
    fi
}

# 主监控流程
main() {
    log_message "INFO: 开始监控检查"
    
    check_database_health
    check_disk_space
    check_database_size
    check_connection_pool
    
    log_message "INFO: 监控检查完成"
}

# 每5分钟执行一次
while true; do
    main
    sleep 300
done
EOF

chmod +x monitor.sh
echo "✅ 监控脚本创建完成: monitor.sh"
```

### 2. 维护脚本
```bash
cat > maintenance.sh << 'EOF'
#!/bin/bash
# 生产环境维护脚本

BACKUP_DIR="/var/backups/gh-pitfall-scraper"
DATE=$(date +%Y%m%d)

log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

daily_maintenance() {
    log_message "INFO: 开始每日维护任务"
    
    # 健康检查
    if ! ./database/db-manager.sh health; then
        log_message "ERROR: 健康检查失败"
        exit 1
    fi
    
    # 清理过期数据
    ./database/db-manager.sh cleanup
    
    # 更新统计信息
    ./database/db-manager.sh analyze
    
    # 创建备份
    mkdir -p $BACKUP_DIR
    ./database/db-manager.sh backup "$BACKUP_DIR/daily_backup_$DATE.db"
    
    log_message "INFO: 每日维护任务完成"
}

weekly_maintenance() {
    log_message "INFO: 开始每周维护任务"
    
    # 数据库优化
    ./database/db-manager.sh optimize
    
    # 重建索引
    ./database/db-manager.sh reindex
    
    # 压缩数据库
    ./database/db-manager.sh vacuum
    
    # 性能测试
    ./database/db-manager.sh benchmark
    
    log_message "INFO: 每周维护任务完成"
}

# 执行维护任务
case "$1" in
    daily)
        daily_maintenance
        ;;
    weekly)
        weekly_maintenance
        ;;
    *)
        echo "使用方法: $0 {daily|weekly}"
        exit 1
        ;;
esac
EOF

chmod +x maintenance.sh
echo "✅ 维护脚本创建完成: maintenance.sh"
```

## 总结

通过以上演示和最佳实践，您可以：

1. **快速上手**: 掌握数据库的基本配置和操作
2. **正确配置**: 根据不同环境选择合适的配置
3. **高效运维**: 使用自动化脚本简化维护工作
4. **稳定运行**: 建立完善的监控和告警机制
5. **持续优化**: 定期评估和调整系统配置

记住：**监控优于告警，告警优于故障处理**。

定期检查系统状态，及时发现和处理问题，确保系统稳定运行。