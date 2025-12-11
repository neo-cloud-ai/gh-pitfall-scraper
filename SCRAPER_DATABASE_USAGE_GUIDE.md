---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 3046022100818952fea970d4c9876c10582cdcd34e0d8c5ce35ef10d9c5ddcaddf8da1ea2d022100bfd9557cc86aad3a84a8b51ba6341d684f138ae9a53ac6f79be180b51cbfb09f
    ReservedCode2: 3046022100df414d6803ac1686cc3c020a37a316d5d3e5616fae208d2f872daa31e3fb96a5022100a2465efec561d0a14fbdedfdb7b95b64e5d12ba394c60e5f213f32a55662917c
---

# 抓取引擎数据库集成使用指南

## 快速开始

### 1. 基本使用

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
    // 1. 初始化数据库
    dbConfig := database.DefaultDatabaseConfig()
    dbConfig.Path = "./data/issues.db"
    
    db, err := database.NewSQLiteDB(dbConfig)
    if err != nil {
        log.Fatalf("数据库初始化失败: %v", err)
    }
    defer db.Close()
    
    // 2. 初始化数据库架构
    if err := database.InitializeSchema(db); err != nil {
        log.Fatalf("数据库架构初始化失败: %v", err)
    }
    
    // 3. 创建数据库服务
    dbService := scraper.NewDatabaseService(db)
    
    // 4. 创建GitHub客户端（带数据库集成）
    client := scraper.NewGithubClientWithDB("your_github_token", dbService)
    
    // 5. 定义仓库和关键词
    repositories := []scraper.Repository{
        {Owner: "vllm-project", Name: "vllm"},
        {Owner: "sgl-project", Name: "sglang"},
    }
    
    keywords := []string{
        "performance", "regression", "latency", "OOM", 
        "memory leak", "CUDA", "kernel", "NCCL",
    }
    
    // 6. 开始抓取（自动存储到数据库）
    issues, err := scraper.ScrapeMultipleRepos(client, dbService, repositories, keywords)
    if err != nil {
        log.Printf("抓取错误: %v", err)
    }
    
    log.Printf("抓取完成，发现 %d 个坑点问题", len(issues))
}
```

### 2. 高级使用

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func advancedExample() {
    // 初始化数据库
    db, _ := initDatabase()
    defer db.Close()
    
    dbService := scraper.NewDatabaseService(db)
    client := scraper.NewGithubClientWithDB("token", dbService)
    
    // 高级抓取选项
    options := scraper.DefaultScrapingOptions()
    options.MaxPages = 3
    options.Concurrency = 2
    
    advancedScraper := scraper.NewAdvancedScraper(client, dbService, options)
    
    // 执行高级抓取
    issues, err := advancedScraper.ScrapeWithAdvancedOptions("vllm-project", "vllm", keywords)
    if err != nil {
        log.Fatal(err)
    }
    
    // 数据库操作
    performDatabaseOperations(client, dbService)
}

func performDatabaseOperations(client *scraper.GithubClient, dbService *scraper.DatabaseService) {
    // 1. 执行去重
    dedupResult, err := client.RunDeduplication()
    if err != nil {
        log.Printf("去重失败: %v", err)
    } else {
        fmt.Printf("去重完成: %d 个重复项\n", dedupResult.DuplicatesFound)
    }
    
    // 2. 执行分类
    recentIssues, _ := client.GetRecentIssuesFromDatabase(100)
    classResult, err := client.RunClassification(recentIssues)
    if err != nil {
        log.Printf("分类失败: %v", err)
    } else {
        fmt.Printf("分类完成: %d 个已分类\n", classResult.Classified)
    }
    
    // 3. 查询高分数问题
    highScoreIssues, err := client.GetHighScoreIssuesFromDatabase(20.0, 50)
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        fmt.Printf("高分数问题: %d 个\n", len(highScoreIssues))
    }
    
    // 4. 获取统计信息
    getStatistics(client, dbService)
}

func getStatistics(client *scraper.GithubClient, dbService *scraper.DatabaseService) {
    // 去重统计
    dupStats, err := client.GetDuplicateStatsFromDatabase()
    if err == nil {
        fmt.Printf("去重统计: %+v\n", dupStats)
    }
    
    // 分类统计
    classStats, err := client.GetClassificationStatsFromDatabase()
    if err == nil {
        fmt.Printf("分类统计: %+v\n", classStats)
    }
}
```

### 3. 数据库过滤示例

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "time"
    
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
    "github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func filteringExample() {
    db, _ := initDatabase()
    defer db.Close()
    
    // 创建过滤服务
    filterService := scraper.NewDatabaseFilterService(db)
    
    // 定义过滤条件
    criteria := scraper.DefaultDatabaseFilterCriteria()
    
    // 设置高分数阈值
    minScore := 20.0
    criteria.MinScore = &minScore
    
    // 设置关键词
    criteria.Keywords = []string{"performance", "OOM", "memory"}
    
    // 设置分类
    criteria.Categories = []string{"bug", "performance"}
    
    // 设置时间范围
    weekAgo := time.Now().AddDate(0, 0, -7)
    criteria.DateFrom = &weekAgo
    
    // 执行过滤
    result, err := filterService.FilterFromDatabase(criteria)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("过滤结果: %d 个问题，耗时 %v\n", 
        result.FilteredCount, result.ProcessingTime)
    
    // 获取过滤统计
    stats, err := filterService.GetFilteredStatsFromDatabase(criteria)
    if err == nil {
        fmt.Printf("过滤统计: %+v\n", stats)
    }
}
```

## API 参考

### DatabaseService

```go
type DatabaseService struct {
    // 数据库操作
    PersistIssue(issue *PitfallIssue) error
    PersistIssuesBatch(issues []*PitfallIssue) (int, error)
    
    // 内部服务
    crudOps               database.CRUDOperations
    dedupService          *database.DeduplicationService
    classificationService *database.ClassificationService
}
```

### GithubClient (数据库扩展)

```go
type GithubClient struct {
    // 原有功能...
    
    // 数据库查询
    GetIssuesFromDatabase(filters DatabaseFilter, limit, offset int) ([]*database.Issue, error)
    GetRecentIssuesFromDatabase(limit int) ([]*database.Issue, error)
    GetHighScoreIssuesFromDatabase(minScore float64, limit int) ([]*database.Issue, error)
    
    // 数据库操作
    RunDeduplication() (*database.DeduplicationResult, error)
    RunClassification(issues []*database.Issue) (*database.ClassificationResult, error)
    
    // 统计信息
    GetDuplicateStatsFromDatabase() (map[string]interface{}, error)
    GetClassificationStatsFromDatabase() (map[string]interface{}, error)
}
```

### DatabaseFilterService

```go
type DatabaseFilterService struct {
    // 过滤操作
    FilterFromDatabase(criteria DatabaseFilterCriteria) (*DatabaseFilterResult, error)
    GetFilteredStatsFromDatabase(criteria DatabaseFilterCriteria) (map[string]interface{}, error)
}
```

### DatabaseScorer

```go
type DatabaseScorer struct {
    // 评分和存储
    ScoreAndStore(issue *Issue, keywords []string) (float64, error)
    ScoreAndStoreBatch(issues []*Issue, keywords []string) ([]float64, error)
    
    // 更新操作
    UpdateIssueScore(issueID int64, issue *Issue, keywords []string) error
    RescoreIssues(limit int) (int, error)
    
    // 查询操作
    GetTopScoredIssues(limit int) ([]*database.Issue, error)
    GetScoreStatistics() (map[string]interface{}, error)
}
```

## 配置说明

### 数据库配置

```yaml
database:
  path: "./data/issues.db"          # 数据库文件路径
  max_open_conns: 25               # 最大连接数
  max_idle_conns: 5                # 最大空闲连接数
  conn_max_lifetime: "300s"        # 连接生命周期
```

### 抓取配置

```yaml
scraping:
  max_pages: 5                     # 最大页数
  issues_per_page: 100            # 每页问题数
  rate_limit_delay: "100ms"       # 速率限制延迟
  min_score: 10.0                 # 最小分数
  min_comments: 1                 # 最少评论数
  concurrency: 3                  # 并发数
```

## 最佳实践

### 1. 错误处理

```go
// 使用上下文进行超时控制
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// 错误重试
for attempt := 0; attempt < 3; attempt++ {
    issues, err := scraper.ScrapeRepo(client, dbService, owner, repoName, keywords)
    if err == nil {
        return issues, nil
    }
    if attempt < 2 {
        time.Sleep(time.Second * time.Duration(attempt+1))
        continue
    }
    return nil, err
}
```

### 2. 批量处理

```go
// 使用批量操作提高性能
batchSize := 50
for i := 0; i < len(issues); i += batchSize {
    end := i + batchSize
    if end > len(issues) {
        end = len(issues)
    }
    
    batch := issues[i:end]
    count, err := dbService.PersistIssuesBatch(batch)
    if err != nil {
        log.Printf("批次 %d 失败: %v", i/batchSize, err)
        continue
    }
    
    log.Printf("批次 %d 完成，处理 %d 个问题", i/batchSize, count)
}
```

### 3. 监控和日志

```go
// 添加适当的日志记录
logger := log.New(os.Stdout, "[Scraper] ", log.LstdFlags)

// 记录性能指标
start := time.Now()
defer func() {
    logger.Printf("操作耗时: %v", time.Since(start))
}()

// 记录错误
if err != nil {
    logger.Printf("错误: %v", err)
}
```

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库文件路径
   - 确认数据库文件权限
   - 验证磁盘空间

2. **内存使用过高**
   - 减少批处理大小
   - 启用连接池
   - 优化查询条件

3. **性能问题**
   - 增加并发数
   - 使用批量操作
   - 添加数据库索引

### 调试模式

```go
// 启用调试日志
logger := log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)

// 记录详细操作
logger.Printf("开始抓取 %s/%s", owner, repoName)
logger.Printf("发现 %d 个问题", len(issues))
logger.Printf("数据库操作耗时: %v", dbOperationTime)
```

## 扩展功能

### 自定义分类规则

```go
// 添加自定义分类规则
customConfig := database.DefaultClassificationConfig()
customConfig.AutoClassificationThreshold = 0.9
classificationService := database.NewClassificationService(db, customConfig)
```

### 自定义去重配置

```go
// 配置去重参数
dedupConfig := database.DefaultDeduplicationConfig()
dedupConfig.SimilarityThreshold = 0.8
dedupConfig.BatchSize = 200
dedupService := database.NewDeduplicationService(db, dedupConfig)
```

这个集成提供了完整的数据库功能，同时保持了与现有抓取引擎的兼容性。