---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 30450220439b15c4d9abbd4890b71cdbfcbcb60226fcd118cdaf0b6ae813bbc625a979d1022100b60af663ef6ee4e4b7cdf2096eb07744455257e405f97217425c263f0536b9a8
    ReservedCode2: 30460221009aa6df1ef6eb18ca8fb2da78b6f24e8ab6feb2fd3a2204ab6de171707f4fac2a022100e39965943b75098360f4102d9f6597a6d4ffde862713e1a93e5b59469b7a45e8
---

# GitHub Pitfall Scraper 数据库架构

本项目为 GitHub Pitfall Scraper 设计了完整的 SQLite 数据库架构，支持 Issues 的去重、分类、时间序列分析等功能。

## 文件结构

```
internal/database/
├── schema.sql      # 数据库表结构定义
├── models.go       # Go 数据结构定义
├── init.go         # 数据库初始化和连接管理
└── example.go      # 使用示例
```

## 数据库设计

### 核心表结构

#### 1. repositories 表
存储被爬取的 GitHub 仓库信息：
- 基础信息：owner、name、full_name、description、url
- 统计数据：stars、forks、issues_count、language
- 状态信息：created_at、updated_at、last_scraped_at、is_active
- 扩展字段：metadata (JSON格式)

#### 2. categories 表
管理问题的分类和标签：
- 分类信息：name、description、color
- 状态控制：is_active、priority
- 时间戳：created_at、updated_at

#### 3. issues 表
存储 GitHub Issues 信息：
- 基础信息：issue_id、repository_id、number、title、body
- 状态信息：state、author_login、author_type
- 关系数据：labels、assignees、milestone
- 统计数据：reactions、comments_count、severity_score、score
- 识别结果：is_pitfall、category_id、is_duplicate
- 时间信息：created_at、updated_at、closed_at、first_seen_at、last_seen_at
- 扩展字段：metadata (JSON格式)

#### 4. time_series 表
存储按时间维度聚合的数据：
- 时间维度：date、year、month、day、week_of_year、day_of_week
- 统计数据：new_issues_count、closed_issues_count、active_issues_count、pitfall_issues_count
- 计算指标：avg_severity_score、total_comments

### 索引优化

数据库设计了全面的索引策略：

#### 单列索引
- `idx_repositories_owner` - 仓库所有者查询
- `idx_issues_created_at` - 按创建时间排序
- `idx_issues_severity_score` - 严重程度排序
- `idx_issues_is_pitfall` - 坑点筛选

#### 复合索引
- `idx_issues_repo_state` - 仓库+状态组合查询
- `idx_issues_repo_category` - 仓库+分类组合查询
- `idx_issues_pitfall_severity` - 坑点+严重程度查询

#### 唯一索引
- `idx_issues_github_id` - GitHub原始ID去重
- `idx_time_series_repo_date` - 时间序列唯一性保证

### 视图

#### 1. active_issues
活跃 Issues 视图，包含仓库和分类信息

#### 2. pitfall_issues
坑点 Issues 视图，便于筛选

#### 3. duplicate_issues
重复 Issues 视图，便于处理

#### 4. repository_stats
仓库统计视图，包含完整的统计信息

### 触发器

#### 自动时间戳更新
- `update_repositories_timestamp`
- `update_issues_timestamp`
- `update_categories_timestamp`
- `update_time_series_timestamp`

#### 自动时间序列更新
- `insert_issue_time_series` - 新建Issue时更新时间序列
- `update_issue_time_series` - Issue状态变化时更新统计

## 使用方法

### 1. 初始化数据库

```go
import "github.com/your-repo/gh-pitfall-scraper/internal/database"

// 配置数据库
config := database.DefaultDatabaseConfig()
config.Path = "./data/gh-pitfall-scraper.db"
config.MaxConnections = 20

// 初始化数据库管理器
dbManager, err := database.NewDatabaseManager(config, logger)
if err != nil {
    log.Fatal(err)
}
defer dbManager.Close()
```

### 2. 仓库操作

```go
ctx := context.Background()

// 创建仓库
repo := &database.Repository{
    Owner:       "octocat",
    Name:        "Hello-World",
    FullName:    "octocat/Hello-World",
    Description: "这是一个示例仓库",
    URL:         "https://github.com/octocat/Hello-World",
    Language:    "Go",
    Stars:       1000,
    IsActive:    true,
}

if err := dbManager.CreateRepository(ctx, repo); err != nil {
    log.Printf("创建仓库失败: %v", err)
}

// 获取仓库
retrievedRepo, err := dbManager.GetRepositoryByID(ctx, repo.ID)
```

### 3. Issue 操作

```go
// 创建 Issue
issue := &database.Issue{
    IssueID:      12345,
    RepositoryID: repo.ID,
    Number:       1,
    Title:        "示例Issue：这是一个需要修复的Bug",
    Body:         "这是Issue的详细内容描述",
    State:        "open",
    AuthorLogin:  "contributor1",
    IsPitfall:    true,
    SeverityScore: 8.5,
    Score:         9.2,
    Labels:       []string{"bug", "priority-high"},
    URL:          "https://github.com/octocat/Hello-World/issues/1",
}

if err := dbManager.CreateIssue(ctx, issue); err != nil {
    log.Printf("创建Issue失败: %v", err)
}

// 查询 Issues
issues, err := dbManager.GetIssuesByRepository(ctx, repo.ID, 10, 0)
```

### 4. 统计分析

```go
// 仓库统计
stats, err := dbManager.GetRepositoryStats(ctx, repo.ID)
log.Printf("仓库 %s 有 %d 个 Issues，%d 个坑点", 
    stats.FullName, stats.TotalIssues, stats.PitfallIssues)

// 全局统计
overallStats, err := dbManager.GetOverallStats(ctx)
log.Printf("总计 %d 个 Issues，平均分数 %.2f", 
    overallStats.TotalCount, overallStats.AverageScore)
```

### 5. 事务操作

```go
err := dbManager.Transaction(ctx, func(tx *sql.Tx) error {
    // 在事务中执行多个操作
    // 如果任何操作失败，整个事务会回滚
    
    // 示例：批量创建 Issues
    for _, issueData := range issuesData {
        if err := createIssueInTx(tx, issueData); err != nil {
            return err // 会触发回滚
        }
    }
    return nil
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

## 数据模型

### 自定义类型

#### 1. CustomTime
适配 SQLite 的 DATETIME 类型，支持 RFC3339 格式

#### 2. JSONSlice
JSON 数组类型，支持字符串切片和数据库之间的转换

#### 3. JSONMap
JSON 对象类型，支持 map[string]interface{} 和数据库之间的转换

#### 4. ReactionCount
GitHub reactions 统计，包含各种反应类型的计数

### 模型方法

每个模型都提供了便捷的方法：

#### Issue 模型
- `IsOpen()` / `IsClosed()` - 检查状态
- `IsHighSeverity()` - 检查严重程度
- `IsHighScore()` - 检查评分
- `HasLabels()` - 检查标签
- `GetAgeInDays()` - 计算存在天数
- `GetContentHash()` - 生成内容哈希
- `IsSimilar()` - 计算相似度

#### Repository 模型
- `GetFullName()` - 获取完整仓库名
- `IsStale()` - 检查是否需要重新爬取

#### Category 模型
- `IsHighPriority()` - 检查是否为高优先级

#### TimeSeries 模型
- `IsToday()` / `IsThisWeek()` / `IsThisMonth()` - 时间范围检查
- `GetActivityScore()` - 计算活动分数

## 性能优化

### 1. 连接池配置
- `MaxOpenConns` - 最大打开连接数
- `MaxIdleConns` - 最大空闲连接数
- `ConnMaxLifetime` - 连接生命周期

### 2. SQLite 优化
- WAL 模式：提高并发性能
- 同步模式：NORMAL 平衡性能和数据安全
- 缓存大小：10000 页提高查询性能

### 3. 查询优化
- 预编译语句：减少 SQL 解析开销
- 索引利用：确保查询使用合适索引
- 分页查询：避免大量数据一次性加载

### 4. 维护操作
- `Analyze()` - 更新统计信息，优化查询计划
- `Vacuum()` - 清理碎片，优化存储空间

## 错误处理

### 1. 连接错误
```go
if err := dbManager.HealthCheck(ctx); err != nil {
    log.Fatalf("数据库连接失败: %v", err)
}
```

### 2. 事务错误
```go
err := dbManager.Transaction(ctx, func(tx *sql.Tx) error {
    // 事务操作
    return err // 如果返回错误，会自动回滚
})

if err != nil {
    log.Printf("事务失败: %v", err)
    // 处理错误或重试
}
```

### 3. 查询错误
```go
issues, err := dbManager.GetIssuesByRepository(ctx, repoID, 10, 0)
if err != nil {
    if err == sql.ErrNoRows {
        log.Println("未找到任何 Issues")
    } else {
        log.Printf("查询失败: %v", err)
    }
    return
}
```

## 配置选项

```go
type DatabaseConfig struct {
    Path            string        // 数据库文件路径
    MaxConnections  int           // 最大连接数
    Timeout         time.Duration // 连接超时
    EnableWAL       bool          // 启用 WAL 模式
    EnableForeignKeys bool        // 启用外键约束
    BusyTimeout     time.Duration // 忙等待超时
}
```

## 最佳实践

### 1. 连接管理
- 使用数据库管理器而不是直接创建连接
- 合理配置连接池大小
- 及时关闭连接释放资源

### 2. 事务使用
- 将相关操作放在同一事务中
- 避免长时间持有事务
- 适当处理事务错误和重试

### 3. 查询优化
- 使用预编译语句防止 SQL 注入
- 利用索引优化查询性能
- 使用分页避免内存溢出

### 4. 数据一致性
- 利用外键约束保证引用完整性
- 使用触发器自动维护衍生数据
- 定期进行数据完整性检查

### 5. 错误处理
- 区分不同类型的错误
- 适当记录错误日志
- 提供有意义的错误信息

## 扩展功能

数据库架构支持以下扩展：

### 1. 数据迁移
可以轻松添加新表和字段，通过版本控制管理数据库演化

### 2. 分区表
对于大数据量场景，可以考虑按时间或仓库分区

### 3. 读写分离
可以配置主从数据库支持读写分离

### 4. 缓存层
结合 Redis 等缓存系统提高查询性能

### 5. 监控告警
集成数据库监控工具，及时发现性能问题