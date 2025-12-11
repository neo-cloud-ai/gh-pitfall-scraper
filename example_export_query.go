package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gh-pitfall-scraper/internal/database"
)

// 演示数据导出和查询功能
func demonstrateExportAndQuery() {
	// 这里假设已经有了一个可用的数据库连接
	// 在实际使用中，你需要先初始化数据库
	
	// 创建数据库连接（示例）
	db, err := sql.Open("sqlite3", "./data/issues.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	fmt.Println("=== 数据导出和查询功能演示 ===")
	
	// 1. 演示查询功能
	demonstrateQuery(db)
	
	// 2. 演示导出功能
	demonstrateExport(db)
	
	// 3. 演示报告生成功能
	demonstrateReportGeneration(db)
}

// 演示查询功能
func demonstrateQuery(db *sql.DB) {
	fmt.Println("\n1. 演示查询功能")
	
	// 创建查询构建器
	qb := database.NewQueryBuilder(db)
	
	// 构建搜索条件
	criteria := database.DefaultSearchCriteria()
	criteria.Query = "bug"
	criteria.Page = 1
	criteria.PageSize = 10
	criteria.SortBy = "score"
	criteria.SortOrder = "DESC"
	
	// 执行简单查询
	result, err := qb.SimpleQuery(criteria)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	
	fmt.Printf("查询结果:\n")
	fmt.Printf("  总记录数: %d\n", result.TotalCount)
	fmt.Printf("  返回记录数: %d\n", len(result.Issues))
	fmt.Printf("  查询耗时: %v\n", result.QueryTime)
	
	// 显示前几个结果
	for i, issue := range result.Issues {
		if i >= 3 { // 只显示前3个
			break
		}
		fmt.Printf("  Issue #%d: %s (Score: %.2f)\n", issue.Number, issue.Title, issue.Score)
	}
	
	// 演示聚合查询
	fmt.Println("\n  演示分类聚合查询:")
	aggCriteria := database.AggregatedQuery{
		GroupBy: "category",
		Metrics: []string{"count", "avg_score"},
		Filters: database.DefaultSearchCriteria(),
	}
	
	aggResult, err := qb.AggregatedQuery(aggCriteria)
	if err != nil {
		fmt.Printf("聚合查询失败: %v\n", err)
		return
	}
	
	fmt.Printf("  聚合结果组数: %d\n", aggResult.TotalGroups)
	for i, group := range aggResult.Groups {
		if i >= 3 { // 只显示前3个
			break
		}
		fmt.Printf("    %s: %v 条记录\n", group["category"], group["count"])
	}
}

// 演示导出功能
func demonstrateExport(db *sql.DB) {
	fmt.Println("\n2. 演示导出功能")
	
	// 创建导出器
	exporter := database.NewExporter(db)
	
	// 构建导出过滤器
	filter := database.ExportFilter{
		DateFrom: func() *time.Time {
			t := time.Now().AddDate(0, -1, 0) // 过去一个月
			return &t
		}(),
		IncludeMetadata: true,
	}
	
	// 演示JSON导出
	fmt.Println("  导出JSON格式...")
	result, err := exporter.ExportIssues(filter, database.FormatJSON, "./data/export_demo.json")
	if err != nil {
		fmt.Printf("导出失败: %v\n", err)
		return
	}
	
	fmt.Printf("  导出结果:\n")
	fmt.Printf("    输出文件: %s\n", result.OutputPath)
	fmt.Printf("    总记录数: %d\n", result.TotalRecords)
	fmt.Printf("    导出记录数: %d\n", result.ExportedRecords)
	fmt.Printf("    耗时: %v\n", result.Duration)
	
	// 演示CSV导出
	fmt.Println("  导出CSV格式...")
	_, err = exporter.ExportIssues(filter, database.FormatCSV, "./data/export_demo.csv")
	if err != nil {
		fmt.Printf("CSV导出失败: %v\n", err)
		return
	}
	fmt.Println("    CSV导出完成")
	
	// 演示Markdown导出
	fmt.Println("  导出Markdown格式...")
	_, err = exporter.ExportIssues(filter, database.FormatMarkdown, "./data/export_demo.md")
	if err != nil {
		fmt.Printf("Markdown导出失败: %v\n", err)
		return
	}
	fmt.Println("    Markdown导出完成")
}

// 演示报告生成功能
func demonstrateReportGeneration(db *sql.DB) {
	fmt.Println("\n3. 演示报告生成功能")
	
	// 创建报告生成器
	reportGen := database.NewReportGenerator(db)
	
	// 构建报告配置
	config := database.ReportConfig{
		Title:       "Issues Analytics Demo Report",
		Description: "这是一个演示报告，展示了Issues数据的分析结果",
		OutputPath:  "./data/demo_report.html",
		Format:      "html",
		Parameters: map[string]interface{}{
			"demo":           true,
			"generated_time": time.Now(),
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
				Limit:      20,
			},
		},
	}
	
	// 生成报告
	fmt.Println("  生成HTML报告...")
	result, err := reportGen.GenerateReport(config)
	if err != nil {
		fmt.Printf("报告生成失败: %v\n", err)
		return
	}
	
	fmt.Printf("  报告生成结果:\n")
	fmt.Printf("    输出文件: %s\n", result.OutputPath)
	fmt.Printf("    格式: %s\n", result.Format)
	fmt.Printf("    大小: %d bytes\n", result.SizeBytes)
	fmt.Printf("    耗时: %v\n", result.Duration)
	
	// 演示多格式导出
	fmt.Println("  生成多种格式的报告...")
	results, err := reportGen.ExportToMultipleFormats(config)
	if err != nil {
		fmt.Printf("多格式导出失败: %v\n", err)
		return
	}
	
	for format, result := range results {
		fmt.Printf("    %s: %s (%d bytes)\n", format, result.OutputPath, result.SizeBytes)
	}
}

// 演示高级查询功能
func demonstrateAdvancedQueries(db *sql.DB) {
	fmt.Println("\n4. 演示高级查询功能")
	
	qb := database.NewQueryBuilder(db)
	
	// 时间序列查询
	fmt.Println("  时间序列查询:")
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0) // 过去3个月
	
	tsq := database.TimeSeriesQuery{
		StartDate: startDate,
		EndDate:   endDate,
		Interval:  "week",
	}
	
	tsResult, err := qb.TimeSeriesQuery(tsq)
	if err != nil {
		fmt.Printf("时间序列查询失败: %v\n", err)
		return
	}
	
	fmt.Printf("    时间序列点数: %d\n", tsResult.TotalPoints)
	fmt.Printf("    查询耗时: %v\n", tsResult.QueryTime)
	
	// 分面查询
	fmt.Println("  分面查询:")
	fq := database.FacetedQuery{
		BaseCriteria: database.DefaultSearchCriteria(),
		Facets:       []string{"category", "repository", "state"},
	}
	
	facetedResult, err := qb.FacetedQuery(fq)
	if err != nil {
		fmt.Printf("分面查询失败: %v\n", err)
		return
	}
	
	fmt.Printf("    总记录数: %d\n", facetedResult.TotalCount)
	fmt.Printf("    分面数量: %d\n", len(facetedResult.Facets))
	for facet, values := range facetedResult.Facets {
		fmt.Printf("      %s: %d 个不同的值\n", facet, len(values))
	}
}

func main() {
	demonstrateExportAndQuery()
	demonstrateAdvancedQueries(nil) // 传入nil作为示例
}