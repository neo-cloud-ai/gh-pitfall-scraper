---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 304402202d84c6ffc3d668e373db3d99694227f37b44ccaa2aec039a8bad38b50f9c1c0402207dbd917d6689a6ad69e08cc6ff2a5e7648c5749df42d731eee832fc3a7281c81
    ReservedCode2: 3045022046fa4d1003630fb2bcfbd2942135740e2e578408d77aa193ebd3f2da6cd95a20022100979e40566e5c2c0c17b17a4ff5bbdf49a4ee7abe3581413836f40a88f038980b
---

# gh-pitfall-scraper 数据库配置和工具

本文档介绍了 gh-pitfall-scraper 项目的数据库配置选项和管理工具。

## 目录结构

```
gh-pitfall-scraper/
├── config.yaml                     # 主配置文件
├── internal/
│   └── database/
│       └── config.go               # 数据库配置管理
├── scripts/
│   └── init-database.sh            # 数据库初始化脚本
└── database-tools.sh               # 数据库管理命令行工具
```

## 数据库配置

### config.yaml 中的数据库配置选项

```yaml
# 数据库配置
database:
  # 数据库文件路径配置
  file_path: "./data/gh-pitfall-scraper.db"
  
  # 连接池配置
  connection_pool:
    max_open_conns: 25      # 最大打开连接数
    max_idle_conns: 5       # 最大空闲连接数
    conn_max_lifetime: 300  # 连接最大生命周期(秒)
  
  # 缓存配置
  cache:
    enabled: true           # 是否启用缓存
    size: 1000             # 缓存大小(条目数)
    ttl: 3600              # 缓存过期时间(秒)
  
  # 清理策略配置
  cleanup:
    enabled: true           # 是否启用自动清理
    interval: 86400         # 清理间隔(秒，24小时)
    max_age: 2592000       # 数据最大保留时间(秒，30天)
  
  # 备份策略配置
  backup:
    enabled: true           # 是否启用自动备份
    interval: 43200         # 备份间隔(秒，12小时)
    retention_days: 7       # 备份保留天数
    path: "./backups"      # 备份文件路径
```

### 配置说明

#### 数据库文件路径
- `file_path`: SQLite 数据库文件的存储路径
- 支持相对路径和绝对路径
- 默认值: `./data/gh-pitfall-scraper.db`

#### 连接池配置
- `max_open_conns`: 同时打开的最大连接数
- `max_idle_conns`: 最大空闲连接数
- `conn_max_lifetime`: 连接的最大生命周期

#### 缓存配置
- `enabled`: 是否启用缓存功能
- `size`: 缓存中存储的最大条目数
- `ttl`: 缓存项的生存时间（秒）

#### 清理策略配置
- `enabled`: 是否启用自动数据清理
- `interval`: 执行清理操作的时间间隔（秒）
- `max_age`: 数据项的最大保留时间（秒）

#### 备份策略配置
- `enabled`: 是否启用自动备份
- `interval`: 执行备份操作的时间间隔（秒）
- `retention_days`: 备份文件的保留天数
- `path`: 备份文件的存储目录

## 使用指南

### 初始化数据库

```bash
# 赋予执行权限（首次使用）
chmod +x scripts/init-database.sh database-tools.sh

# 初始化数据库
./scripts/init-database.sh

# 初始化数据库并创建测试数据
./scripts/init-database.sh -s

# 重新创建所有表结构
./scripts/init-database.sh -d

# 强制初始化（跳过确认）
./scripts/init-database.sh -f
```

### 数据库管理工具

```bash
# 赋予执行权限（首次使用）
chmod +x database-tools.sh

# 查看帮助信息
./database-tools.sh --help

# 初始化数据库
./database-tools.sh init

# 创建备份
./database-tools.sh backup
./database-tools.sh backup backup-$(date +%Y%m%d).db

# 恢复数据库
./database-tools.sh restore ./backups/backup-20231211.db

# 清理过期数据
./database-tools.sh clean

# 压缩数据库文件
./database-tools.sh vacuum

# 查看数据库状态
./database-tools.sh status

# 查看数据库配置
./database-tools.sh info

# 健康检查
./database-tools.sh health

# 测试数据库连接
./database-tools.sh test

# 使用自定义配置文件
./database-tools.sh -c /path/to/config.yaml status

# 详细输出模式
./database-tools.sh -v init

# 试运行模式（不实际执行）
./database-tools.sh --dry-run clean
```

## 数据库表结构

数据库包含以下主要表：

- **repositories**: 仓库信息
- **issues**: GitHub Issues
- **comments**: Issue 评论
- **commits**: 代码提交
- **pulls**: Pull Requests
- **scraper_metadata**: 爬虫元数据
- **error_patterns**: 错误模式定义
- **performance_metrics**: 性能指标

### 关键索引

- 仓库信息按 owner/name 组合索引
- Issues 按 repo_id、state、author 索引
- 按创建时间和更新时间索引所有主要表
- 性能指标按 repo_id 和 metric_type 索引

### 触发器

- 自动更新 `updated_at` 字段
- 确保数据时间戳的准确性

## 依赖要求

### 必需依赖
- **SQLite3**: 用于数据库操作
- **Go**: 用于编译和运行主程序

### 可选依赖
- **yq**: 用于 YAML 文件解析
- **Python3**: 用于配置文件读取

### 安装依赖

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install sqlite3

# CentOS/RHEL
sudo yum install sqlite

# macOS
brew install sqlite

# Arch Linux
sudo pacman -S sqlite

# 验证安装
sqlite3 --version
```

## 最佳实践

### 配置建议
1. **连接池大小**: 根据应用负载调整连接池参数
2. **缓存策略**: 根据内存使用情况设置缓存大小和TTL
3. **清理策略**: 根据数据保留需求设置清理间隔
4. **备份策略**: 定期备份数据库文件

### 性能优化
1. **定期清理**: 使用 `clean` 命令清理过期数据
2. **数据库压缩**: 定期使用 `vacuum` 命令压缩数据库
3. **索引维护**: 保持适当的索引以提高查询性能
4. **监控空间**: 定期检查磁盘空间使用情况

### 安全建议
1. **文件权限**: 确保数据库文件有适当的访问权限
2. **备份验证**: 定期验证备份文件的完整性
3. **敏感信息**: 避免在配置文件中存储敏感信息
4. **访问控制**: 限制对数据库文件的访问

## 故障排除

### 常见问题

1. **权限错误**
   ```bash
   chmod +x database-tools.sh scripts/init-database.sh
   ```

2. **SQLite3 未找到**
   ```bash
   # Ubuntu/Debian
   sudo apt install sqlite3
   
   # macOS
   brew install sqlite
   ```

3. **配置文件解析错误**
   ```bash
   # 检查配置文件语法
   yq eval . config.yaml
   
   # 使用默认配置
   ./database-tools.sh --dry-run init
   ```

4. **数据库锁定**
   ```bash
   # 检查数据库进程
   lsof ./data/gh-pitfall-scraper.db
   
   # 强制结束进程或重启应用
   ```

### 调试模式

```bash
# 启用详细输出
./database-tools.sh -v init

# 试运行模式
./database-tools.sh --dry-run clean

# 使用配置文件验证
./database-tools.sh -c config.yaml info
```

## 更新日志

### v1.0.0
- 初始版本发布
- 完整的数据库配置系统
- 数据库管理命令行工具
- 初始化脚本
- 支持备份和恢复
- 自动化清理和优化功能

## 贡献

如需贡献代码或报告问题，请遵循项目的贡献指南。

## 许可证

本项目使用与 gh-pitfall-scraper 相同的许可证。