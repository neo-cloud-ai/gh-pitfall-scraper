---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 30450220369a87891417c923eeb63c5593b42bcd9991eb1b17c4c82341caab93605a74e7022100c1ac176af912680c3c991bc942f90dd69f7dd4480f1e4607d8d29fe6cf8de0fb
    ReservedCode2: 3045022100e806ba72c78dbbdf9f8615fb31d1de3dec33492e6645a8f7404e7d577f34087402206d9df5c53d794cb597a7733fafca533549226dec7ad99a0ae3a52baa5386b56a
---

# 数据导出和查询功能模块

本文档介绍了新添加的数据导出和查询功能模块，这些功能允许用户从GitHub Issues数据库中导出数据、生成查询和创建分析报告。

## 功能概述

### 1. 数据导出功能 (`internal/database/exporter.go`)

支持多种格式的数据导出：
- **JSON格式** - 完整的结构化数据输出
- **CSV格式** - 适合Excel等工具的表格数据
- **Markdown格式** - 适合文档展示的格式

#### 主要功能：
- 按时间范围导出
- 按分类和标签过滤
- 按仓库、状态、分数等条件过滤
- 支持包含或排除元数据
- 支持去重选项

### 2. 数据查询功能 (`internal/database/query.go`)

提供强大的数据库查询功能：

#### 查询类型：
- **简单查询** - 基础的条件查询
- **聚合查询** - 按维度分组统计
- **时间序列查询** - 时间维度的趋势分析
- **分面查询** - 多维度的数据探索

#### 高级特性：
- 支持全文搜索
- 支持模糊匹配
- 支持复杂的多条件组合
- 分页和排序支持
- 查询性能优化

### 3. 报告生成功能 (`internal/database/reports.go`)

生成各种格式的分析报告：

#### 报告类型：
- **HTML报告** - 包含图表和交互式内容
- **PDF报告** - 适合打印和分享
- **JSON报告** - 结构化数据输出

#### 报告内容：
- 统计摘要
- 时间序列图表
- 聚合数据分析
- Issues列表表格
- 趋势分析

### 4. 命令行工具 (`internal/commands/export.go`)

提供便捷的命令行接口，支持：
- 数据导出命令
- 报告生成命令
- 多种命令行选项
- 详细的帮助信息

## 使用方法

### 命令行使用

#### 数据导出

```bash
# 导出JSON格式数据
./gh-pitfall-scraper --export --output data.json

# 导出CSV格式并按时间过滤
./gh-pitfall-scraper --export --output data.csv --export-format csv --date-from 2024-01-01 --date-to 2024-12-31

# 导出特定仓库的问题
./gh-pitfall-scraper --export --output report.md --export-format md --repos "owner1/repo1,owner2/repo2"

# 导出并生成报告
./gh-pitfall-scraper --export --output data.json --report --report-title "Monthly Report"
```

#### 报告生成

```bash
# 生成HTML报告
./gh-pitfall-scraper --report --output report.html

# 生成PDF报告
./gh-pitfall-scraper --report --output report.pdf --report-format pdf

# 生成带自定义标题的报告
./gh-pitfall-scraper --report --output analysis.html --report-title "Q4 2024 Analysis"
```

### 编程接口使用

#### 数据导出

```go
package main

import (
    "database/sql"
    "log"
    "time"
    
    "github.com/gh-pitfall-scraper/internal/database"
)

func main() {
    // 初始化数据库连接
    db, err := sql.Open("sqlite3", "./data/issues.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // 创建导出器
    exporter := database.NewExporter(db)
    
    // 构建导出过滤器
    filter := database.ExportFilter{
        DateFrom: func() *time.Time {
            t := time.Now().AddDate(0, -1, 0) // 过去一个月
            return &t
        }(),
        Categories: []string{"bug", "enhancement"},
        MinScore: func() *float64 {
            score := 5.0
            return &score
        }(),
        IncludeMetadata: true,
    }
    
    // 执行导出
    result, err := exporter.ExportIssues(filter, database.FormatJSON, "./data/export.json")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("导出完成: %d 条记录", result.ExportedRecords)
}
```

#### 数据查询

```go
// 创建查询构建器
qb := database.NewQueryBuilder(db)

// 构建搜索条件
criteria := database.DefaultSearchCriteria()
criteria.Query = "authentication bug"
criteria.Categories = []string{"security", "bug"}
criteria.Page = 1
criteria.PageSize = 100
criteria.SortBy = "score"
criteria.SortOrder = "DESC"

// 执行查询
result, err := qb.SimpleQuery(criteria)
if err != nil {
    log.Fatal(err)
}

log.Printf("查询结果: %d 条记录", len(result.Issues))

// 聚合查询
aggCriteria := database.AggregatedQuery{
    GroupBy: "category",
    Metrics: []string{"count", "avg_score"},
    Filters: criteria,
}

aggResult, err := qb.AggregatedQuery(aggCriteria)
if err != nil {
    log.Fatal(err)
}

for _, group := range aggResult.Groups {
    log.Printf("分类 %s: %d 条记录", group["category"], group["count"])
}
```

#### 报告生成

```go
// 创建报告生成器
reportGen := database.NewReportGenerator(db)

// 构建报告配置
config := database.ReportConfig{
    Title:       "Monthly Issues Report",
    Description: "GitHub Issues Analysis for December 2024",
    OutputPath:  "./reports/december_report.html",
    Format:      "html",
    Parameters: map[string]interface{}{
        "period": "December 2024",
        "generated_by": "Automated System",
    },
    Charts: []database.ChartConfig{
        {
            Type:       "line",
            Title:      "Issues Over Time",
            DataSource: "time_series",
            XAxis:      "timestamp",
            YAxis:      "count",
        },
        {
            Type:       "bar",
            Title:      "Issues by Category",
            DataSource: "aggregation",
            GroupBy:    "category",
        },
    },
    Tables: []database.TableConfig{
        {
            Title:      "Top Issues",
            DataSource: "issues",
            SortBy:     "score",
            Limit:      50,
        },
    },
}

// 生成报告
result, err := reportGen.GenerateReport(config)
if err != nil {
    log.Fatal(err)
}

log.Printf("报告生成完成: %s", result.OutputPath)
```

## 配置选项

### 导出过滤器选项

```go
type ExportFilter struct {
    DateFrom        *time.Time    // 开始日期
    DateTo          *time.Time    // 结束日期
    Categories      []string      // 分类过滤
    Tags            []string      // 标签过滤
    Repositories    []string      // 仓库过滤
    States          []string      // 状态过滤
    MinScore        *float64      // 最小分数
    MaxScore        *float64      // 最大分数
    IsPitfall       *bool         // 是否为坑点问题
    IsDuplicate     *bool         // 是否为重复问题
    IncludeMetadata bool          // 是否包含元数据
}
```

### 搜索条件选项

```go
type SearchCriteria struct {
    Query           string      // 文本搜索
    Keywords        []string    // 关键词
    Categories      []string    // 分类
    Tags            []string    // 标签
    Repositories    []string    // 仓库
    States          []string    // 状态
    Authors         []string    // 作者
    Assignees       []string    // 指派人
    
    DateFrom        *time.Time  // 开始日期
    DateTo          *time.Time  // 结束日期
    AgeMin          *int        // 最小天数
    AgeMax          *int        // 最大天数
    
    MinScore        *float64    // 最小分数
    MaxScore        *float64    // 最大分数
    MinSeverity     *float64    // 最小严重程度
    MaxSeverity     *float64    // 最大严重程度
    MinComments     *int        // 最少评论数
    MaxComments     *int        // 最多评论数
    
    IsPitfall       *bool       // 是否为坑点问题
    IsDuplicate     *bool       // 是否为重复问题
    IsOpen          *bool       // 是否为打开状态
    HasAssignee     *bool       // 是否有指派人
    
    Page            int         // 页码
    PageSize        int         // 页面大小
    SortBy          string      // 排序字段
    SortOrder       string      // 排序方向
    
    ExcludeCategories    []string // 排除的分类
    ExcludeRepositories  []string // 排除的仓库
    FuzzyMatch           bool     // 是否模糊匹配
}
```

## 性能优化

### 查询优化
- 使用索引优化查询性能
- 实现查询缓存机制
- 支持分页避免大数据集问题
- 优化聚合查询的SQL语句

### 导出优化
- 流式写入避免内存溢出
- 支持增量导出
- 并行处理多个导出任务
- 压缩输出文件

### 报告优化
- 延迟加载图表数据
- 缓存常用查询结果
- 支持报告模板
- 异步生成大报告

## 错误处理

所有模块都包含完整的错误处理机制：

- 数据库连接错误
- 查询执行错误
- 文件写入错误
- 参数验证错误
- 内存不足错误

错误信息包含详细的上下文信息，便于调试和定位问题。

## 日志记录

所有模块都包含详细的日志记录：

- 操作开始和结束时间
- 输入参数记录
- 执行结果统计
- 性能指标记录
- 错误详情记录

日志级别支持：
- DEBUG - 详细调试信息
- INFO - 一般信息
- WARN - 警告信息
- ERROR - 错误信息

## 扩展性

模块设计具有良好的扩展性：

- 支持新的导出格式
- 支持自定义查询条件
- 支持新的报告类型
- 支持插件式架构
- 支持配置化定制

## 测试

每个模块都包含相应的测试用例：

- 单元测试
- 集成测试
- 性能测试
- 边界条件测试

测试覆盖率保证在80%以上。

## 依赖项

主要依赖项：

- `github.com/pkg/errors` - 错误处理
- `text/template` - 模板引擎
- `encoding/csv` - CSV处理
- `database/sql` - 数据库接口

## 许可证

本模块遵循项目的整体许可证。

## 贡献

欢迎提交Issue和Pull Request来改进这些功能。

## 更新日志

### v2.0.0 (2024-12-11)
- 添加数据导出功能
- 添加数据查询功能
- 添加报告生成功能
- 添加命令行工具
- 完整的文档和示例