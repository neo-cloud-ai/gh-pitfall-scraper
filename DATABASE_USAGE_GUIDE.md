---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 3045022100d2c7c82d36fe9e766c603f799cb2ff5b45bf06be8fa224bd3c889a6841e47e5d02204742ee23f05fe79d8ea93201d7be1860f645da2f9aad32872a4d292e2e2756cd
    ReservedCode2: 3046022100e60d7e25df8732fccb61f7a211edcc53f306c00f2d970ed2c3be08cc0a45a5e2022100cf0d4c3d2acb4046b7160975668540624cd14e6c88098e33ebdc14664e1a5a28
---

# gh-pitfall-scraper 数据库使用指南

## 目录

1. [快速开始](#快速开始)
2. [配置详解](#配置详解)
3. [数据库操作](#数据库操作)
4. [数据查询](#数据查询)
5. [性能优化](#性能优化)
6. [备份恢复](#备份恢复)
7. [故障排除](#故障排除)
8. [最佳实践](#最佳实践)
9. [高级功能](#高级功能)
10. [监控维护](#监控维护)

## 快速开始

### 1. 环境准备

确保系统已安装必要的依赖：

```bash
# 检查 Go 版本（需要 1.21+）
go version

# 检查 SQLite3（如果使用 SQLite）
sqlite3 --version

# 检查 PostgreSQL（如果使用 PostgreSQL）
psql --version

# 克隆项目
git clone https://github.com/neo-cloud-ai/gh-pitfall-scraper.git
cd gh-pitfall-scraper
```

### 2. 快速初始化

```bash
# 方法一：使用 Makefile（推荐）
make deps        # 安装依赖
make db-init     # 初始化数据库
make run         # 运行程序

# 方法二：使用数据库管理工具
chmod +x ./database/db-manager.sh
./database/db-manager.sh init
./gh-pitfall-scraper --db-only

# 方法三：直接运行主程序
./gh-pitfall-scraper --config config.yaml --db-only
```

### 3. 验证安装

```bash
# 检查数据库状态
./gh-pitfall-scraper --health

# 查看统计信息
./gh-pitfall-scraper --stats

# 运行完整爬虫
./gh-pitfall-scraper --config config.yaml
```

## 配置详解

### 1. 完整配置文件示例

创建 `config.yaml` 文件：

```yaml
# 应用配置
app:
  name: "gh-pitfall-scraper"
  version: "2.0.0"
  log_level: "info"
  data_dir: "./data"
  
# GitHub API 配置
github:
  token: "ghp_your_github_token_here"
  api_url: "https://api.github.com"
  rate_limit: 5000
  timeout: 30
  
# 数据库配置
database:
  # 数据库类型：sqlite 或 postgres
  driver: "sqlite"
  
  # SQLite 配置
  sqlite:
    file_path: "./data/gh-pitfall-scraper.db"
    wal_mode: true          # 启用 WAL 模式提高并发性能
    busy_timeout: 30000     # 忙等待超时（毫秒）
    
  # PostgreSQL 配置
  postgres:
    host: "localhost"
    port: 5432
    user: "postgres"
    password: "your_password"
    dbname: "gh_pitfall_scraper"
    sslmode: "disable"
    
  # 连接池配置
  connection_pool:
    max_open_conns: 25      # 最大打开连接数
    max_idle_conns: 5       # 最大空闲连接数
    conn_max_lifetime: "300s"   # 连接最大生命周期
    conn_max_idle_time: "60s"   # 连接最大空闲时间
    
  # 缓存配置
  cache:
    enabled: true           # 是否启用缓存
    size: 1000             # 缓存大小（条目数）
    ttl: "3600s"           # 缓存过期时间
    
  # 清理策略配置
  cleanup:
    enabled: true           # 是否启用自动清理
    interval: "24h"         # 清理间隔
    max_age: "720h"         # 数据最大保留时间（30天）
    batch_size: 1000        # 批处理大小
    
  # 备份策略配置
  backup:
    enabled: true           # 是否启用自动备份
    interval: "12h"         # 备份间隔
    retention_days: 7       # 备份保留天数
    path: "./backups"       # 备份文件路径
    compress: true          # 是否压缩备份文件
    
  # 监控配置
  monitoring:
    enabled: true           # 是否启用监控
    metrics_interval: "5m"  # 指标收集间隔
    health_check_interval: "1m"  # 健康检查间隔
    
# 仓库配置
repos:
  - owner: vllm-project
    name: vllm
  - owner: sgl-project
    name: sglang
  - owner: NVIDIA
    name: TensorRT-LLM
  - owner: microsoft
    name: DeepSpeed
  - owner: pytorch
    name: pytorch
  - owner: huggingface
    name: transformers

# 关键词配置
keywords:
  - "performance"
  - "regression"
  - "latency"
  - "throughput"
  - "OOM"
  - "memory leak"
  - "CUDA"
  - "kernel"
  - "NCCL"
  - "hang"
  - "deadlock"
  - "kv cache"
  - "flash attention"
  - "distributed"
  - "training"
  - "inference"

# 评分配置
scoring:
  # 关键词权重
  keyword_weights:
    critical: 10.0
    high: 5.0
    medium: 2.0
    low: 1.0
    
  # 参与度权重
  engagement_weights:
    reactions: 0.3
    comments: 0.2
    assignees: 0.1
    milestone: 0.4
    
  # 时效性权重
  recency_weights:
    days_old_7: 1.0
    days_old_30: 0.8
    days_old_90: 0.6
    days_old_365: 0.4
    days_old_9999: 0.2
```

### 2. 配置验证

```bash
# 验证配置文件语法
yq eval . config.yaml

# 测试数据库连接
./gh-pitfall-scraper --config config.yaml --health

# 验证所有配置项
./database/db-manager.sh -c config.yaml info
```

## 数据库操作

### 1. 基础操作

#### 初始化数据库
```bash
# 初始化新数据库
./database/db-manager.sh init

# 强制重新初始化（删除现有数据）
./database/db-manager.sh reset

# 初始化并创建测试数据
./database/db-manager.sh init --with-sample-data
```

#### 状态检查
```bash
# 查看数据库状态
./database/db-manager.sh status

# 健康检查
./database/db-manager.sh health

# 查看详细信息
./database/db-manager.sh info

# 查看配置信息
./database/db-manager.sh config
```

#### 数据操作
```bash
# 清理过期数据
./database/db-manager.sh cleanup

# 压缩数据库
./database/db-manager.sh vacuum

# 重建索引
./database/db-manager.sh reindex

# 分析查询性能
./database/db-manager.sh analyze
```

### 2. 备份恢复

#### 创建备份
```bash
# 创建完整备份
./database/db-manager.sh backup

# 创建带时间戳的备份
./database/db-manager.sh backup backup-$(date +%Y%m%d_%H%M%S).db

# 备份到指定目录
./database/db-manager.sh backup /path/to/backups/custom-backup.db

# 压缩备份
./database/db-manager.sh backup --compress
```

#### 恢复数据
```bash
# 从备份恢复
./database/db-manager.sh restore ./backups/backup-20231211.db

# 恢复到指定数据库文件
./database/db-manager.sh restore --target ./data/backup.db ./backups/backup-20231211.db

# 恢复并验证完整性
./database/db-manager.sh restore --verify ./backups/backup-20231211.db
```

### 3. 迁移操作

#### 数据库迁移
```bash
# 初始化迁移系统
go run database/migration.go init

# 创建新迁移
go run database/migration.go create add_user_table "添加用户管理功能"

# 执行迁移
go run database/migration.go migrate

# 查看迁移状态
go run database/migration.go status

# 回滚迁移
go run database/migration.go rollback 1

# 查看迁移历史
go run database/migration.go history
```

## 数据查询

### 1. 通过命令行查询

```bash
# 查看总体统计
./gh-pitfall-scraper --stats

# 查看特定仓库统计
./gh-pitfall-scraper --stats --repo=vllm-project/vllm

# 查看最近的数据
./gh-pitfall-scraper --stats --days=7

# 查看高优先级问题
./gh-pitfall-scraper --stats --min-score=20
```

### 2. 直接数据库查询

#### SQLite 查询
```bash
# 连接到 SQLite 数据库
sqlite3 ./data/gh-pitfall-scraper.db

# 查看所有表
.tables

# 查看表结构
.schema issues

# 查询最近的问题
SELECT title, score, created_at 
FROM issues 
ORDER BY created_at DESC 
LIMIT 10;

# 查看各仓库问题统计
SELECT r.owner, r.name, COUNT(i.id) as issue_count 
FROM repositories r 
JOIN issues i ON r.id = i.repository_id 
GROUP BY r.id 
ORDER BY issue_count DESC;
```

#### PostgreSQL 查询
```bash
# 连接到 PostgreSQL 数据库
psql -h localhost -U postgres -d gh_pitfall_scraper

# 查看所有表
\dt

# 查看表结构
\d issues

# 查询示例
SELECT title, score, created_at 
FROM issues 
ORDER BY created_at DESC 
LIMIT 10;
```

### 3. 常用查询示例

#### 问题统计分析
```sql
-- 查看各类型问题分布
SELECT 
    c.name as category,
    COUNT(i.id) as count,
    AVG(i.score) as avg_score,
    MAX(i.severity_score) as max_severity
FROM issues i
LEFT JOIN categories c ON i.category_id = c.id
GROUP BY c.id, c.name
ORDER BY count DESC;

-- 查看最近30天的趋势
SELECT 
    date_trunc('day', created_at) as date,
    COUNT(*) as daily_issues,
    COUNT(CASE WHEN is_pitfall = 1 THEN 1 END) as pitfall_issues
FROM issues 
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY date_trunc('day', created_at)
ORDER BY date;

-- 查看最活跃的仓库
SELECT 
    r.owner,
    r.name,
    COUNT(i.id) as total_issues,
    COUNT(CASE WHEN i.state = 'open' THEN 1 END) as open_issues,
    COUNT(CASE WHEN i.is_pitfall = 1 THEN 1 END) as pitfall_issues,
    AVG(i.score) as avg_score
FROM repositories r
JOIN issues i ON r.id = i.repository_id
GROUP BY r.id
ORDER BY pitfall_issues DESC;
```

#### 性能分析查询
```sql
-- 查看查询性能统计
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- 查看索引使用情况
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- 查看数据库大小
SELECT 
    pg_size_pretty(pg_database_size('gh_pitfall_scraper')) as database_size;
```

## 性能优化

### 1. 配置优化

#### 连接池优化
```yaml
database:
  connection_pool:
    # 高并发场景
    max_open_conns: 50      # 增加最大连接数
    max_idle_conns: 20      # 增加空闲连接数
    conn_max_lifetime: "600s"   # 延长连接生命周期
    
    # 低并发场景
    max_open_conns: 10      # 减少连接数
    max_idle_conns: 3       # 减少空闲连接数
    conn_max_lifetime: "300s"   # 标准连接生命周期
```

#### 缓存优化
```yaml
database:
  cache:
    # 大数据量场景
    enabled: true
    size: 5000             # 增加缓存大小
    ttl: "7200s"           # 延长缓存时间
    
    # 小数据量场景
    enabled: true
    size: 500              # 减少缓存大小
    ttl: "1800s"           # 缩短缓存时间
```

### 2. 数据库优化

#### SQLite 优化
```bash
# 启用 WAL 模式
./database/db-manager.sh set-wal-mode

# 设置合适的缓存大小
./database/db-manager.sh set-cache-size 10000

# 优化数据库
./database/db-manager.sh optimize

# 重建索引
./database/db-manager.sh reindex
```

#### PostgreSQL 优化
```sql
-- 更新统计信息
ANALYZE;

-- 重建索引
REINDEX DATABASE gh_pitfall_scraper;

-- 查看慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- 查看索引效率
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    idx_scan::float / NULLIF(idx_tup_read, 0) as selectivity
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

### 3. 查询优化

#### 优化建议
1. **使用索引**: 确保常用查询字段有索引
2. **分页查询**: 使用 LIMIT 和 OFFSET 进行分页
3. **避免 SELECT ***: 只查询需要的字段
4. **使用连接**: 用 JOIN 替代子查询
5. **批量操作**: 使用批量插入和更新

#### 示例优化
```sql
-- 优化前（慢查询）
SELECT * FROM issues WHERE title LIKE '%performance%';

-- 优化后（使用索引）
SELECT id, title, score, created_at 
FROM issues 
WHERE title LIKE 'performance%'  -- 前缀匹配可以使用索引
LIMIT 100;

-- 使用覆盖索引
CREATE INDEX idx_issues_title_score ON issues(title, score, created_at);

-- 批量插入优化
INSERT INTO issues (...) VALUES 
(...), (...), (...)  -- 一次性插入多条记录
```

## 故障排除

### 1. 常见问题

#### 数据库连接问题
```bash
# 检查数据库文件权限
ls -la ./data/gh-pitfall-scraper.db

# 检查磁盘空间
df -h

# 检查数据库进程
lsof ./data/gh-pitfall-scraper.db

# 重启数据库连接
./database/db-manager.sh restart
```

#### 性能问题
```bash
# 查看数据库状态
./database/db-manager.sh status --verbose

# 运行性能测试
./database/db-manager.sh benchmark

# 分析查询性能
./database/db-manager.sh analyze

# 查看慢查询日志
tail -f ./logs/slow-queries.log
```

#### 内存问题
```bash
# 查看内存使用
./database/db-manager.sh memory-usage

# 清理缓存
./database/db-manager.sh clear-cache

# 重建数据库
./database/db-manager.sh rebuild
```

### 2. 错误处理

#### 数据库锁定
```bash
# 检查锁定状态
./database/db-manager.sh check-locks

# 强制解锁
./database/db-manager.sh unlock

# 重启应用
pkill -f gh-pitfall-scraper
./gh-pitfall-scraper --config config.yaml
```

#### 数据损坏
```bash
# 检查数据库完整性
./database/db-manager.sh integrity-check

# 修复数据库
./database/db-manager.sh repair

# 从备份恢复
./database/db-manager.sh restore ./backups/latest-backup.db
```

#### 迁移失败
```bash
# 查看迁移状态
go run database/migration.go status

# 回滚失败的迁移
go run database/migration.go rollback 1

# 重新执行迁移
go run database/migration.go migrate
```

### 3. 调试模式

```bash
# 启用详细日志
./gh-pitfall-scraper --config config.yaml --debug --verbose

# 查看数据库日志
tail -f ./logs/database.log

# 启用 SQL 跟踪
./database/db-manager.sh enable-sql-trace

# 查看实时统计
./database/db-manager.sh watch-stats
```

## 最佳实践

### 1. 数据管理

#### 数据生命周期
- **保留策略**: 根据业务需求设置数据保留时间
- **定期清理**: 自动化清理过期和无效数据
- **归档策略**: 对历史数据进行归档处理
- **备份策略**: 制定完整的备份和恢复计划

#### 数据质量
- **数据验证**: 在数据入库前进行验证
- **去重处理**: 避免重复数据影响分析结果
- **数据一致性**: 确保数据在各个表中的一致性
- **完整性检查**: 定期检查数据完整性

### 2. 性能管理

#### 监控指标
- **连接数**: 监控数据库连接池使用情况
- **查询性能**: 跟踪慢查询和性能瓶颈
- **存储空间**: 监控数据库文件大小增长
- **缓存命中率**: 监控缓存效率

#### 优化策略
- **索引优化**: 根据查询模式创建合适索引
- **查询优化**: 优化常用查询语句
- **配置调优**: 根据负载调整数据库配置
- **硬件优化**: 使用 SSD 存储增加 IO 性能

### 3. 安全实践

#### 访问控制
```bash
# 设置文件权限
chmod 600 ./data/gh-pitfall-scraper.db
chmod 755 ./database/

# 使用环境变量存储敏感信息
export GITHUB_TOKEN="your_token"
export DB_PASSWORD="your_password"

# 配置文件权限
chmod 644 config.yaml
```

#### 数据保护
- **加密存储**: 对敏感数据进行加密
- **访问日志**: 记录所有数据库访问
- **定期审计**: 定期检查数据库访问权限
- **备份安全**: 确保备份文件的安全性

### 4. 运维建议

#### 监控告警
```bash
# 设置监控脚本
#!/bin/bash
# monitor-db.sh

# 检查数据库健康
if ! ./database/db-manager.sh health > /dev/null; then
    echo "数据库健康检查失败" | mail -s "DB Alert" admin@example.com
fi

# 检查磁盘空间
USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $USAGE -gt 80 ]; then
    echo "磁盘空间不足: ${USAGE}%" | mail -s "Disk Alert" admin@example.com
fi

# 检查数据库大小
SIZE=$(./database/db-manager.sh size)
if [ $SIZE -gt 10737418240 ]; then  # 10GB
    echo "数据库大小超过10GB" | mail -s "Size Alert" admin@example.com
fi
```

#### 自动化任务
```bash
# 添加到 crontab
# 0 2 * * * /path/to/backup-daily.sh
# 0 */6 * * * /path/to/monitor-db.sh
# 0 0 * * 0 /path/to/weekly-maintenance.sh

# 每日备份脚本
#!/bin/bash
# backup-daily.sh
DATE=$(date +%Y%m%d)
BACKUP_DIR="/backups"
mkdir -p $BACKUP_DIR
./database/db-manager.sh backup "$BACKUP_DIR/backup-$DATE.db"
find $BACKUP_DIR -name "backup-*.db" -mtime +7 -delete
```

## 高级功能

### 1. 自定义查询

#### 创建数据库视图
```sql
-- 创建高优先级问题视图
CREATE VIEW high_priority_issues AS
SELECT 
    i.*,
    r.owner,
    r.name as repo_name,
    c.name as category_name
FROM issues i
JOIN repositories r ON i.repository_id = r.id
LEFT JOIN categories c ON i.category_id = c.id
WHERE i.severity_score >= 7.0 
    AND i.state = 'open'
ORDER BY i.severity_score DESC, i.score DESC;

-- 使用视图查询
SELECT * FROM high_priority_issues WHERE repo_name = 'vllm';
```

#### 自定义函数
```sql
-- 创建计算问题年龄的函数
CREATE FUNCTION get_issue_age_days(created_at DATETIME) 
RETURNS INTEGER AS $$
BEGIN
    RETURN CAST((julianday('now') - julianday(created_at)) AS INTEGER);
END;
$$ LANGUAGE SQL;

-- 使用自定义函数
SELECT 
    title,
    get_issue_age_days(created_at) as age_days
FROM issues;
```

### 2. 数据导出

#### 导出为 CSV
```bash
# 导出问题数据
./gh-pitfall-scraper --export-csv issues.csv --format=issues

# 导出特定仓库数据
./gh-pitfall-scraper --export-csv vllm-issues.csv --repo=vllm-project/vllm

# 导出统计报告
./gh-pitfall-scraper --export-csv stats.csv --format=statistics
```

#### 导出为 JSON
```bash
# 导出为 JSON 格式
./gh-pitfall-scraper --export-json issues.json --format=issues

# 导出特定时间范围
./gh-pitfall-scraper --export-json recent-issues.json --since="2024-01-01" --until="2024-12-31"
```

### 3. 批量操作

#### 批量更新
```sql
-- 批量更新问题分类
UPDATE issues 
SET category_id = (
    SELECT id FROM categories WHERE name = 'Performance'
)
WHERE keywords LIKE '%performance%' 
    OR keywords LIKE '%latency%'
    OR keywords LIKE '%throughput%';

-- 批量更新严重程度
UPDATE issues 
SET severity_score = CASE 
    WHEN score >= 20 THEN 9.0
    WHEN score >= 15 THEN 7.0
    WHEN score >= 10 THEN 5.0
    ELSE 3.0
END;
```

#### 批量删除
```sql
-- 删除重复问题
DELETE FROM issues 
WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY title, repository_id 
            ORDER BY created_at DESC
        ) as rn
        FROM issues
    ) t 
    WHERE rn > 1
);

-- 删除过期数据
DELETE FROM issues 
WHERE created_at < datetime('now', '-365 days')
    AND state = 'closed'
    AND comments_count = 0;
```

## 监控维护

### 1. 性能监控

#### 实时监控
```bash
# 实时查看数据库状态
./database/db-manager.sh watch

# 监控连接池
./database/db-manager.sh monitor-connections

# 监控查询性能
./database/db-manager.sh monitor-queries

# 监控存储使用
./database/db-manager.sh monitor-storage
```

#### 性能报告
```bash
# 生成性能报告
./database/db-manager.sh performance-report

# 生成使用统计报告
./database/db-manager.sh usage-report

# 生成健康检查报告
./database/db-manager.sh health-report
```

### 2. 定期维护

#### 每日维护任务
```bash
#!/bin/bash
# daily-maintenance.sh

echo "开始每日维护任务 - $(date)"

# 健康检查
./database/db-manager.sh health
if [ $? -ne 0 ]; then
    echo "健康检查失败"
    exit 1
fi

# 清理过期数据
./database/db-manager.sh cleanup

# 更新统计信息
./database/db-manager.sh analyze

# 创建备份
./database/db-manager.sh backup

echo "每日维护任务完成 - $(date)"
```

#### 每周维护任务
```bash
#!/bin/bash
# weekly-maintenance.sh

echo "开始每周维护任务 - $(date)"

# 数据库优化
./database/db-manager.sh optimize

# 重建索引
./database/db-manager.sh reindex

# 压缩数据库
./database/db-manager.sh vacuum

# 性能基准测试
./database/db-manager.sh benchmark

# 生成维护报告
./database/db-manager.sh maintenance-report

echo "每周维护任务完成 - $(date)"
```

### 3. 告警系统

#### 设置告警阈值
```yaml
# alerting.yaml
database:
  alerts:
    disk_usage_warning: 80      # 磁盘使用警告阈值（%）
    disk_usage_critical: 90     # 磁盘使用严重阈值（%）
    db_size_warning: "5GB"      # 数据库大小警告阈值
    connection_pool_warning: 80 # 连接池使用警告阈值（%）
    query_time_warning: 1000    # 查询时间警告阈值（毫秒）
    
  notifications:
    email:
      enabled: true
      smtp_server: "smtp.example.com"
      smtp_port: 587
      username: "alerts@example.com"
      password: "your_password"
      recipients: ["admin@example.com", "devops@example.com"]
    
    webhook:
      enabled: true
      url: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
      channel: "#alerts"
```

#### 告警脚本
```bash
#!/bin/bash
# alert-system.sh

# 检查磁盘使用
DISK_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"🚨 磁盘使用率警告: '${DISK_USAGE}'%"}' \
        $SLACK_WEBHOOK_URL
fi

# 检查数据库大小
DB_SIZE=$(./database/db-manager.sh size-bytes)
if [ $DB_SIZE -gt 5368709120 ]; then  # 5GB
    echo "数据库大小超过5GB: $(./database/db-manager.sh size)" | \
    mail -s "数据库大小警告" admin@example.com
fi

# 检查连接池
POOL_USAGE=$(./database/db-manager.sh pool-usage)
if [ $POOL_USAGE -gt 80 ]; then
    echo "连接池使用率过高: ${POOL_USAGE}%" | \
    mail -s "连接池警告" admin@example.com
fi
```

## 总结

本指南涵盖了 gh-pitfall-scraper 数据库的完整使用方法，从基础配置到高级功能，从性能优化到故障排除。通过遵循本指南，您可以：

1. **快速上手**: 掌握数据库的基本配置和操作
2. **高效使用**: 了解查询优化和性能调优技巧
3. **稳定运行**: 学会监控、维护和故障处理
4. **安全可靠**: 实施数据保护和备份策略
5. **持续改进**: 监控性能指标并持续优化

如有问题，请参考：
- [项目 README](README.md)
- [数据库集成报告](DATABASE_INTEGRATION_REPORT.md)
- [数据库设计总结](DATABASE_DESIGN_SUMMARY.md)

或者提交 Issue 获取帮助。