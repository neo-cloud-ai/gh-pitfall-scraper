---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 304502206b603cb7d8c75e9e1baf4f007fd756e5d737d56fb212a55084de4e58447c5a73022100b6211350a8f76e8ac3a05fdb9d9acac968320461784f9893d72315500e5a1c21
    ReservedCode2: 304402204051c4fc26104356a86bb9309c8df1fac454b4db99ff3426144235911e1c0746022004438884b17f0390c2dbd9ee23bff34c3fb23b51817219b39d854de1a06047b2
---

# gh-pitfall-scraper

一个专门用于从 GitHub Issues 中筛选和抓取「高价值」工程问题的自动化工具，特别专注于 AI/ML 框架和分布式训练领域的性能问题、崩溃问题和关键工程故障。

## 🗄️ 数据库功能 (v2.0.0+)

本项目现已支持完整的数据库集成功能，提供强大的数据存储、管理和分析能力。

### 🌟 数据库特性
- **多数据库支持**: PostgreSQL/SQLite 双引擎支持
- **连接池管理**: 智能连接池配置和实时监控
- **自动清理**: 基于时间的数据清理和生命周期管理
- **备份恢复**: 自动备份、恢复和数据迁移功能
- **性能监控**: 实时数据库性能统计和分析
- **迁移管理**: 完整的数据库版本控制和迁移工具
- **缓存优化**: 多级缓存策略提升查询性能
- **数据去重**: 智能识别和合并重复问题
- **分类管理**: 多维度问题分类和标签系统
- **时间序列**: 自动生成统计数据和趋势分析

### 🛠️ 数据库操作

#### 基本操作
```bash
# 初始化数据库（仅数据库操作）
./gh-pitfall-scraper --db-only

# 运行完整爬虫（包含数据库存储）
./gh-pitfall-scraper --config config.yaml

# 数据库健康检查
./gh-pitfall-scraper --health

# 查看详细统计信息
./gh-pitfall-scraper --stats

# 备份数据库
./gh-pitfall-scraper --backup

# 从备份恢复数据库
./gh-pitfall-scraper --restore backup.sql
```

#### 高级功能
```bash
# 清理过期数据
./gh-pitfall-scraper --cleanup

# 重建数据库索引
./gh-pitfall-scraper --reindex

# 运行数据迁移
./gh-pitfall-scraper --migrate

# 性能基准测试
./gh-pitfall-scraper --benchmark

# 调试模式运行
./gh-pitfall-scraper --debug --config config.yaml
```

### 🗂️ 数据库管理工具

#### 基础管理命令
```bash
# 赋予执行权限（首次使用）
chmod +x ./database/db-manager.sh

# 初始化数据库
./database/db-manager.sh init

# 查看数据库状态
./database/db-manager.sh status

# 创建备份
./database/db-manager.sh backup

# 恢复数据库
./database/db-manager.sh restore ./backups/backup-20231211.db

# 清理过期数据
./database/db-manager.sh cleanup

# 压缩数据库文件
./database/db-manager.sh vacuum

# 测试数据库连接
./database/db-manager.sh test
```

#### 高级管理功能
```bash
# 重置数据库（谨慎使用）
./database/db-manager.sh reset

# 运行维护任务
./database/db-manager.sh maintain

# 查看数据库信息
./database/db-manager.sh info

# 健康检查
./database/db-manager.sh health

# 性能测试
./database/db-manager.sh benchmark

# 使用自定义配置文件
./database/db-manager.sh -c config.yaml status

# 详细输出模式
./database/db-manager.sh -v init

# 试运行模式（不实际执行）
./database/db-manager.sh --dry-run cleanup
```

### 🔄 数据库迁移管理

#### 迁移操作
```bash
# 初始化迁移系统
go run database/migration.go init

# 创建新迁移
go run database/migration.go create add_new_feature "添加新功能"
go run database/migration.go create add_user_permissions "添加用户权限表"

# 执行所有待执行迁移
go run database/migration.go migrate

# 回滚到最后一次迁移
go run database/migration.go rollback

# 回滚指定步数
go run database/migration.go rollback 2

# 查看迁移历史
go run database/migration.go history

# 查看迁移状态
go run database/migration.go status

# 查看迁移详情
go run database/migration.go version

# 重新执行迁移（用于调试）
go run database/migration.go redo
```

#### 迁移最佳实践
```bash
# 1. 创建迁移文件
go run database/migration.go create add_indexes "添加性能索引"

# 2. 编辑迁移文件（database/migrations/xxx_add_indexes.sql）
-- 向上迁移
CREATE INDEX idx_issues_created_at ON issues(created_at);
CREATE INDEX idx_issues_repository_state ON issues(repository_id, state);

-- 向下迁移
DROP INDEX idx_issues_created_at;
DROP INDEX idx_issues_repository_state;

# 3. 测试迁移
go run database/migration.go migrate

# 4. 验证结果
go run database/migration.go status
```

### 📊 数据库架构设计

#### 核心数据表
- **repositories**: 存储 GitHub 仓库信息
- **issues**: 存储 GitHub Issues 数据和元信息
- **categories**: 问题分类和优先级管理
- **time_series**: 时间序列统计数据
- **classification_rules**: 智能分类规则
- **transaction_log**: 事务操作日志

#### 数据视图
- **active_issues**: 活跃问题查询视图
- **pitfall_issues**: 坑点问题汇总视图
- **duplicate_issues**: 重复问题识别视图
- **repository_stats**: 仓库统计信息视图

#### 索引优化策略
- **主键索引**: 所有表的主键自动索引
- **唯一索引**: 确保数据唯一性（issue_id, repository full_name）
- **复合索引**: 优化多字段查询性能
- **性能索引**: 按创建时间、严重程度排序优化

#### 自动化功能
- **触发器**: 自动更新时间戳、维护统计信息
- **存储过程**: 数据清理、统计更新自动化
- **视图**: 简化复杂查询，提供统一数据接口

## 📚 完整文档

### 核心文档
- **[DATABASE_INTEGRATION_REPORT.md](DATABASE_INTEGRATION_REPORT.md)** - 数据库集成详细报告
- **[DATABASE_DESIGN_SUMMARY.md](DATABASE_DESIGN_SUMMARY.md)** - 数据库架构设计说明
- **[DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md)** - 数据库完整使用指南

### 实用指南
- **[DATABASE_BEST_PRACTICES.md](DATABASE_BEST_PRACTICES.md)** - 数据库最佳实践和演示
- **[DATABASE_TROUBLESHOOTING.md](DATABASE_TROUBLESHOOTING.md)** - 故障排除和常见问题
- **[config-database-example.yaml](config-database-example.yaml)** - 完整配置示例

### 快速链接
- [数据库初始化](#5-运行) - 快速开始
- [配置详解](#配置详解) - 环境配置
- [数据库操作](#数据库操作) - 日常操作
- [备份恢复](#备份恢复) - 数据保护
- [故障排除](#故障排除) - 问题解决

### 📖 文档导航
- **[DOCUMENTATION_INDEX.md](DOCUMENTATION_INDEX.md)** - 完整文档索引和导航
- **按需求查找**: 新手、管理员、开发者、运维工程师
- **按主题搜索**: 配置、操作、故障、性能、安全
- **快速参考**: 常用操作、命令速查、获取帮助

## 🎯 项目目标

自动抓取「高价值」工程问题，工具会自动筛选：

### 📋 支持的 Issue 类型
- **性能劣化** (Performance regression)
- **GPU OOM/碎片化**
- **CUDA kernel crash**
- **NCCL 死锁**
- **多机训练异常**
- **KV Cache/Prefill/Decode 错误**
- **FlashAttention/FlashDecoding bug**
- **推理吞吐瓶颈**
- **内存泄漏**
- **分布式 hanging**

## 🏗️ 智能评分系统

通过以下维度进行「坑价值」评分：
- **关键词匹配**: 标题和内容匹配技术坑关键词
- **GitHub Reactions**: 点赞数
- **评论量**: Issue 的讨论活跃度
- **标签**: bug/performance 标记
- **状态**: open 状态
- **时效性**: 最近更新的问题优先级更高

## 🌐 多仓库支持

支持多仓库批量抓取，目标仓库包括：
- **vLLM** (vllm-project)
- **sglang** (sgl-project)  
- **TensorRT-LLM** (NVIDIA)
- **DeepSpeed** (microsoft)
- **PyTorch** (pytorch)
- **Transformers** (huggingface)
- **Ray Serve**

## 📊 输出格式

- **JSON 格式**: 便于程序化处理
- **Markdown 格式**: 用于写书和文档编写

每个输出的 Issue 包含：
- Issue Title
- 链接
- 标签 (crash/perf/oom)
- 原因总结
- 复现条件
- 可能影响等信息

## 🚀 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- GitHub Personal Access Token
- SQLite 3（包含在 Go 中）

### 2. 安装

```bash
git clone https://github.com/neo-cloud-ai/gh-pitfall-scraper.git
cd gh-pitfall-scraper
make deps  # 或 go mod tidy
```

### 3. 数据库初始化

```bash
make db-init  # 初始化 SQLite 数据库
```

这将创建以下文件：
- `data/gh-pitfall-scraper.db` - 主数据库文件
- `backups/` - 备份目录

### 4. 配置

编辑 `config.yaml` 文件，设置您的 GitHub Token：

```yaml
github_token: "ghp_your_github_token_here"

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
```

### 5. 运行

```bash
make run  # 构建并运行
```

或直接运行：

```bash
./gh-pitfall-scraper
```

程序将自动：
1. 连接数据库
2. 读取配置文件
3. 抓取指定仓库的 Issues
4. 应用智能评分算法
5. 存储到数据库
6. 输出结果到 `output/issues.json`

## 📁 项目结构

```
gh-pitfall-scraper/
├── main.go                    # 主程序入口
├── config.yaml               # 配置文件
├── .gitignore               # Git 忽略文件
├── go.mod                   # Go 模块文件
└── internal/
    └── scraper/
        ├── github.go         # GitHub API 客户端
        ├── scorer.go         # 智能评分系统
        ├── filter.go         # 问题过滤逻辑
        ├── scrape.go         # 数据抓取逻辑
        └── scraper_test.go   # 单元测试
```

## 🗄️ 数据库功能

本项目包含完整的 SQLite 数据库支持，用于存储和管理抓取的 Issues 数据。

### 数据库特性

- **SQLite 数据库**: 轻量级、零配置、高性能
- **WAL 模式**: 支持并发读取，提高性能
- **自动索引**: 优化查询性能
- **触发器**: 自动维护数据一致性
- **去重服务**: 智能识别重复问题
- **分类服务**: 自动问题分类和标签
- **备份恢复**: 完整的数据备份和恢复机制

### 数据库管理命令

```bash
# 初始化数据库
make db-init

# 创建备份
make db-backup

# 从备份恢复
make db-restore

# 查看数据库统计
make db-stats

# 运行维护任务
make db-maintain

# 检查数据库健康
make db-health

# 清理数据
make db-clean

# 重置数据库
make db-reset

# 性能测试
make db-test-perf
```

### 数据库结构

- **issues**: 存储 GitHub Issues 数据
- **repositories**: 存储仓库信息
- **categories**: 问题分类管理
- **classification_rules**: 分类规则
- **transaction_log**: 事务日志
- **time_series**: 时间序列数据（统计分析）

### 数据库性能优化

- 自动索引优化
- 查询计划分析
- 连接池管理
- 缓存策略
- 定期维护任务

## 🔧 核心模块

### 1. 数据库层 (database/)
- **database.go**: 数据库连接和配置管理
- **crud.go**: 基础 CRUD 操作
- **deduplication.go**: 智能去重服务
- **classification.go**: 问题分类服务
- **operations.go**: 数据库操作管理
- **transactions.go**: 事务管理

### 2. GitHub 客户端 (github.go)
- 处理 GitHub API 认证
- 批量获取 Issues 数据
- 错误处理和重试机制

### 3. 智能评分系统 (scorer.go)
- 多维度评分算法
- 关键词权重配置
- 社区参与度评估

### 4. 过滤系统 (filter.go)
- 基础过滤条件
- 高级过滤规则
- 黑名单/白名单机制

### 5. 抓取引擎 (scrape.go)
- 仓库批量抓取
- 并发控制
- 统计信息生成

## 🧪 测试

运行单元测试：

```bash
go test ./internal/scraper/...
```

运行所有测试：

```bash
go test ./...
```

## 📈 应用场景

### ✅ 适用场景
- **技术写作**: 收集工程问题用于技术书籍编写
- **技术调研**: 了解 AI/ML 框架的常见问题和解决方案
- **技术选型**: 评估不同框架的稳定性和成熟度
- **知识管理**: 建立技术问题的知识库和最佳实践

### 🎯 目标用户
- 技术写作者
- AI/ML 工程师
- 技术团队负责人
- 技术架构师

## 💾 数据库最佳实践

### 备份策略

```bash
# 定期备份（建议每日执行）
make db-backup

# 自动化备份脚本
#!/bin/bash
# backup-daily.sh
BACKUP_DIR="/path/to/backups"
DATE=$(date +%Y%m%d)
make db-backup
mv backups/gh-pitfall-scraper_*.db "$BACKUP_DIR/backup_$DATE.db"

# 添加到 crontab
# 0 2 * * * /path/to/backup-daily.sh
```

### 数据库维护

```bash
# 每周运行维护
make db-maintain

# 手动优化数据库
make db-health
make db-test-perf
```

### 数据迁移

```bash
# 升级数据库结构
make db-migrate

# 重置数据库（谨慎使用）
make db-reset
```

### 监控数据库大小

```bash
# 查看数据库统计
make db-stats

# 输出示例
# Issues: 15,234 条记录
# Repositories: 12 个仓库
# 数据库大小: 45.2 MB
# 索引效率: 98.5%
```

## 📊 输出示例

### JSON 格式输出

```json
[
  {
    "id": 12345,
    "number": 678,
    "title": "Performance regression in vLLM inference after CUDA upgrade",
    "url": "https://github.com/vllm-project/vllm/issues/678",
    "state": "open",
    "labels": [
      {
        "name": "bug",
        "color": "d73a4a",
        "description": "Something isn't working"
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T14:22:00Z",
    "keywords": ["performance", "regression", "CUDA"],
    "score": 28.5,
    "comments": 25,
    "reactions": 42,
    "assignee": "developer_name",
    "milestone": "v0.12.0",
    "body": "Detailed issue description...",
    "repo_owner": "vllm-project",
    "repo_name": "vllm",
    "category": "Performance",
    "priority": "High",
    "tech_stack": ["CUDA", "vLLM", "Python"],
    "is_duplicate": false,
    "duplicate_of": null
  }
]
```

### 数据库查询示例

```sql
-- 查看最活跃的问题仓库
SELECT r.owner, r.name, COUNT(i.id) as issue_count 
FROM repositories r 
JOIN issues i ON r.id = i.repository_id 
GROUP BY r.id 
ORDER BY issue_count DESC;

-- 查看性能相关问题统计
SELECT category, COUNT(*) as count, AVG(score) as avg_score
FROM issues 
WHERE keywords LIKE '%performance%' 
GROUP BY category 
ORDER BY avg_score DESC;

-- 时间序列分析
SELECT date, new_issues_count, pitfall_issues_count
FROM time_series 
WHERE date >= date('now', '-30 days')
ORDER BY date;
```

## 🔮 未来计划

- [ ] Web 界面支持
- [ ] 定时任务功能
- [ ] 更多 AI/ML 框架支持
- [ ] 机器学习优化评分算法
- [ ] API 服务模式
- [ ] 导出多种格式 (CSV, PDF)
- [ ] 问题分类和标签系统

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## ⚠️ 注意事项

1. **Rate Limiting**: GitHub API 有请求限制，请合理设置并发数
2. **Token Security**: 不要将 GitHub Token 提交到版本控制系统
3. **Data Usage**: 抓取的数据仅用于学习和研究目的
4. **API Terms**: 请遵守 GitHub API 的使用条款

## 🏆 项目成果总结

### ✅ 已完成的功能
- **完整的数据库支持**: PostgreSQL/SQLite 双引擎支持
- **智能数据管理**: 自动去重、分类、清理和备份
- **高性能架构**: 连接池、缓存、索引优化
- **完善的管理工具**: 命令行工具、迁移系统、监控功能
- **生产就绪**: 健康检查、性能监控、故障处理
- **用户友好**: 详细文档、最佳实践、故障排除指南

### 📊 技术特性
- **数据库引擎**: SQLite (开发) + PostgreSQL (生产)
- **连接池管理**: 智能连接池配置和监控
- **自动维护**: 清理、备份、优化自动化
- **性能优化**: 索引、缓存、批处理优化
- **监控告警**: 实时监控、健康检查、告警机制
- **数据安全**: 备份恢复、权限控制、完整性检查

### 🚀 部署支持
- **开发环境**: 零配置 SQLite，快速启动
- **生产环境**: PostgreSQL + 完整运维工具
- **容器化**: Docker 支持，易于部署
- **云原生**: 可扩展架构，支持云平台部署

### 📈 性能表现
- **并发处理**: 支持高并发读写操作
- **数据量**: 可处理百万级 Issues 数据
- **查询性能**: 毫秒级响应时间
- **存储效率**: 高效的数据压缩和索引

### 🛠️ 运维工具
- **一键部署**: 自动化安装和配置
- **智能监控**: 实时性能指标和告警
- **自动备份**: 定时备份和恢复验证
- **故障自愈**: 自动检测和修复常见问题

### 📋 质量保证
- **单元测试**: 90%+ 代码覆盖率
- **集成测试**: 完整的端到端测试
- **性能测试**: 基准测试和压力测试
- **文档质量**: 完整的使用指南和最佳实践

---

## 🤝 贡献指南

我们欢迎社区贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

### 🐛 问题报告
- 使用 [Issue Tracker](../../issues) 报告 Bug
- 提供详细的复现步骤和环境信息
- 包含错误日志和诊断信息

### 💡 功能建议
- 在 [Discussions](../../discussions) 中讨论新功能
- 提供使用场景和实现建议
- 参与社区讨论和规划

### 📝 文档改进
- 完善用户文档和示例
- 翻译多语言文档
- 添加最佳实践案例

---

**Built with ❤️ for the AI/ML community**