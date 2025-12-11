---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 3046022100ed4ff7f3da818dbc20a7f88476051f3e97be2c54c5543c665d8415aa3b93e306022100cb1610f81cce6c33a7bf199eef2663cfa657a821c8bc9ebc922981a43fe65c04
    ReservedCode2: 3044022040a5387c6b218ff5dc7748bdcd267ad6780d4e335b38a123ded911c7fcbbc93502200edf9c021b2021f9568f943f4ab328f45a307fca67054f48615055a1335ad4b6
---

# GitHub Pitfall Scraper 数据库架构设计总结

## 设计完成情况

✅ **已完成的文件：**

1. **schema.sql** - 数据库表结构定义 (297行)
   - 4个核心表：repositories, categories, issues, time_series
   - 完整的索引优化策略
   - 4个视图：active_issues, pitfall_issues, duplicate_issues, repository_stats
   - 自动触发器：时间戳更新、时间序列维护
   - 默认分类数据初始化

2. **models.go** - Go数据结构定义 (651行)
   - 完整的模型结构：Repository, Issue, Category, TimeSeries
   - 自定义类型：CustomTime, JSONSlice, JSONMap, ReactionCount
   - 兼容现有代码结构
   - 丰富的方法：查询、验证、计算等
   - 查询参数和分页支持

3. **init.go** - 数据库初始化和连接管理 (711行)
   - DatabaseManager：完整的数据库管理器
   - 连接池管理和优化
   - 事务支持
   - 健康检查和维护功能
   - 错误处理和日志记录

4. **example.go** - 使用示例 (275行)
   - 完整的使用演示
   - 高级查询示例
   - 事务操作示例
   - 最佳实践展示

5. **database_test.go** - 测试文件 (510行)
   - 全面的单元测试
   - 性能基准测试
   - 事务测试
   - 维护功能测试

6. **README.md** - 详细文档 (353行)
   - 完整的使用指南
   - 架构设计说明
   - 性能优化建议
   - 最佳实践

## 架构特点

### 1. 数据完整性
- **外键约束**：确保数据一致性
- **唯一索引**：防止重复数据
- **触发器**：自动维护衍生数据

### 2. 性能优化
- **索引策略**：单列索引 + 复合索引 + 唯一索引
- **连接池**：合理的连接数配置
- **查询优化**：预编译语句和分页查询

### 3. 功能完整性
- **去重功能**：通过 content_hash 和唯一索引
- **分类管理**：多维度分类和优先级
- **时间序列**：自动统计和历史追踪
- **统计功能**：多维度数据统计

### 4. 扩展性
- **JSON字段**：存储复杂结构化数据
- **视图**：简化复杂查询
- **触发器**：自动化数据维护

## 核心表设计

### repositories 表
```sql
CREATE TABLE repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    full_name TEXT UNIQUE NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    stars INTEGER DEFAULT 0,
    forks INTEGER DEFAULT 0,
    issues_count INTEGER DEFAULT 0,
    language TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,
    metadata TEXT,
    CONSTRAINT unique_repository UNIQUE (owner, name)
);
```

### issues 表
```sql
CREATE TABLE issues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_id BIGINT UNIQUE NOT NULL,
    repository_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL,
    author_login TEXT NOT NULL,
    author_type TEXT DEFAULT 'User',
    labels TEXT,
    assignees TEXT,
    milestone TEXT,
    reactions TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    first_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_pitfall BOOLEAN DEFAULT 0,
    severity_score REAL DEFAULT 0,
    category_id INTEGER,
    score REAL DEFAULT 0,
    url TEXT NOT NULL,
    html_url TEXT NOT NULL,
    comments_count INTEGER DEFAULT 0,
    is_duplicate BOOLEAN DEFAULT 0,
    duplicate_of INTEGER,
    metadata TEXT,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FOREIGN KEY (duplicate_of) REFERENCES issues(issue_id) ON DELETE SET NULL
);
```

### time_series 表
```sql
CREATE TABLE time_series (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    date DATE NOT NULL,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    day INTEGER NOT NULL,
    week_of_year INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL,
    new_issues_count INTEGER DEFAULT 0,
    closed_issues_count INTEGER DEFAULT 0,
    active_issues_count INTEGER DEFAULT 0,
    pitfall_issues_count INTEGER DEFAULT 0,
    avg_severity_score REAL DEFAULT 0,
    total_comments INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repository_id, date),
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE
);
```

## 索引设计

### 关键索引
- **主键索引**：所有表的主键自动索引
- **唯一索引**：issues.issue_id（去重）、repositories.full_name（唯一性）
- **复合索引**：issues(repository_id, state)、issues(repository_id, category_id)
- **性能索引**：issues(created_at DESC)、issues(severity_score DESC)

### 索引统计
- 总共创建了 20+ 个索引
- 覆盖所有常用查询场景
- 优化了排序和筛选性能

## 自动化功能

### 触发器
1. **时间戳自动更新**：所有表的 updated_at 字段
2. **时间序列自动维护**：Issue 变更时更新统计

### 视图
1. **active_issues**：活跃 Issue 查询
2. **pitfall_issues**：坑点 Issue 查询
3. **duplicate_issues**：重复 Issue 查询
4. **repository_stats**：仓库统计视图

## 数据模型

### 自定义类型
- **CustomTime**：SQLite DATETIME 适配
- **JSONSlice**：JSON 数组类型
- **JSONMap**：JSON 对象类型
- **ReactionCount**：GitHub reactions 统计

### 模型方法
- **Issue**：IsOpen(), IsHighSeverity(), HasLabels(), GetAgeInDays()
- **Repository**：GetFullName(), IsStale()
- **Category**：IsHighPriority()
- **TimeSeries**：IsToday(), GetActivityScore()

## 使用示例

### 基本使用
```go
// 初始化数据库
config := database.DefaultDatabaseConfig()
dbManager, err := database.NewDatabaseManager(config, logger)

// 创建仓库
repo := &database.Repository{
    Owner:    "octocat",
    Name:     "Hello-World", 
    FullName: "octocat/Hello-World",
    URL:      "https://github.com/octocat/Hello-World",
}
dbManager.CreateRepository(ctx, repo)

// 创建 Issue
issue := &database.Issue{
    IssueID:      12345,
    RepositoryID: repo.ID,
    Number:       1,
    Title:        "Bug: Application crashes",
    State:        "open",
    IsPitfall:    true,
    SeverityScore: 8.5,
}
dbManager.CreateIssue(ctx, issue)

// 查询统计
stats, err := dbManager.GetRepositoryStats(ctx, repo.ID)
log.Printf("Issues: %d, Pitfalls: %d", stats.TotalIssues, stats.PitfallIssues)
```

### 事务操作
```go
err := dbManager.Transaction(ctx, func(tx *sql.Tx) error {
    // 在事务中执行多个相关操作
    if err := createRepo(tx, repo); err != nil {
        return err
    }
    if err := createIssue(tx, issue); err != nil {
        return err
    }
    return nil
})
```

## 性能优化

### SQLite 优化设置
- **WAL 模式**：提高并发性能
- **NORMAL 同步**：平衡性能和数据安全
- **缓存大小**：10000 页
- **忙等待超时**：30秒

### 连接池配置
- **最大打开连接数**：10-25
- **最大空闲连接数**：5-10
- **连接生命周期**：300秒

## 维护功能

### 自动维护
- **ANALYZE**：更新查询统计
- **VACUUM**：清理存储碎片
- **健康检查**：连接状态监控

### 备份策略
- **自动备份**：每12小时
- **保留天数**：7天
- **备份路径**：可配置

## 兼容性

### 与现有代码兼容
- 保留现有 Issue 模型结构
- 向后兼容的字段命名
- 平滑迁移路径

### 扩展性
- JSON 元数据字段支持未来扩展
- 模块化设计便于功能扩展
- 插件化的分类和评分系统

## 测试覆盖

### 单元测试
- ✅ 仓库操作测试
- ✅ Issue 操作测试
- ✅ 分类管理测试
- ✅ 统计功能测试
- ✅ 事务处理测试

### 性能测试
- ✅ 创建 Issue 性能测试
- ✅ 查询性能测试
- ✅ 统计查询性能测试

## 总结

数据库架构设计已完成，提供了：

1. **完整的功能**：去重、分类、时间序列、统计分析
2. **优秀的性能**：优化的索引和查询策略
3. **高可靠性**：事务支持和数据完整性保证
4. **易用性**：清晰的 API 和完整的文档
5. **扩展性**：灵活的设计支持未来发展

该架构能够支持大规模 GitHub Issues 的存储、分析和查询，为 GitHub Pitfall Scraper 项目提供强大的数据基础。