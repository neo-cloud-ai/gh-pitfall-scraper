package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// ReportConfig 报告配置
type ReportConfig struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`
	OutputPath  string                 `json:"output_path"`
	Format      string                 `json:"format"` // "html", "pdf", "json"
	Parameters  map[string]interface{} `json:"parameters"`
	Charts      []ChartConfig          `json:"charts"`
	Tables      []TableConfig          `json:"tables"`
	Metrics     []string               `json:"metrics"`
}

// ChartConfig 图表配置
type ChartConfig struct {
	Type       string            `json:"type"` // "line", "bar", "pie", "scatter"
	Title      string            `json:"title"`
	DataSource string            `json:"data_source"` // "time_series", "aggregation", "statistics"
	XAxis      string            `json:"x_axis"`
	YAxis      string            `json:"y_axis"`
	GroupBy    string            `json:"group_by"`
	Filter     SearchCriteria    `json:"filter"`
	Options    map[string]interface{} `json:"options"`
}

// TableConfig 表格配置
type TableConfig struct {
	Title      string            `json:"title"`
	Columns    []TableColumn     `json:"columns"`
	DataSource string            `json:"data_source"`
	Filter     SearchCriteria    `json:"filter"`
	SortBy     string            `json:"sort_by"`
	Limit      int               `json:"limit"`
}

// TableColumn 表格列配置
type TableColumn struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Type        string `json:"type"` // "string", "number", "date", "boolean"
	Format      string `json:"format"`
	Sortable    bool   `json:"sortable"`
	Filterable  bool   `json:"filterable"`
	Aggregated  bool   `json:"aggregated"`
}

// ReportData 报告数据
type ReportData struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	GeneratedAt time.Time                `json:"generated_at"`
	Filters     map[string]interface{}   `json:"filters"`
	
	// 统计信息
	Statistics  map[string]interface{}   `json:"statistics"`
	
	// 时间序列数据
	TimeSeries  []TimeSeriesPoint        `json:"time_series"`
	
	// 聚合数据
	Aggregations map[string][]map[string]interface{} `json:"aggregations"`
	
	// Issues数据
	Issues      []*Issue                 `json:"issues"`
	
	// 图表数据
	Charts      map[string]interface{}   `json:"charts"`
	
	// 表格数据
	Tables      map[string]interface{}   `json:"tables"`
	
	// 元数据
	Metadata    map[string]interface{}   `json:"metadata"`
}

// ReportGenerator 报告生成器
type ReportGenerator struct {
	db      *sql.DB
	logger  *log.Logger
	queryBuilder *QueryBuilder
	templates map[string]*template.Template
}

// ReportResult 报告生成结果
type ReportResult struct {
	ReportData *ReportData           `json:"report_data"`
	OutputPath string                `json:"output_path"`
	Format     string                `json:"format"`
	GeneratedAt time.Time            `json:"generated_at"`
	Duration   time.Duration         `json:"duration"`
	SizeBytes  int64                 `json:"size_bytes"`
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	Metric      string                 `json:"metric"`
	Period      string                 `json:"period"`
	Data        []TimeSeriesPoint      `json:"data"`
	
	// 趋势指标
	GrowthRate  float64                `json:"growth_rate"`        // 增长率
	Volatility  float64                `json:"volatility"`         // 波动性
	Trend       string                 `json:"trend"`              // 趋势方向: "increasing", "decreasing", "stable"
	
	// 统计指标
	Mean        float64                `json:"mean"`
	Median      float64                `json:"median"`
	StdDev      float64                `json:"std_dev"`
	Min         float64                `json:"min"`
	Max         float64                `json:"max"`
	
	// 预测
	Predictions []TimeSeriesPoint      `json:"predictions"`
	Confidence  float64                `json:"confidence"`
}

// ComparisonAnalysis 比较分析
type ComparisonAnalysis struct {
	Metric       string                   `json:"metric"`
	Groups       []string                 `json:"groups"`
	Data         map[string][]TimeSeriesPoint `json:"data"`
	
	// 比较结果
	Differences  map[string]float64       `json:"differences"`
	Ratios       map[string]float64       `json:"ratios"`
	Significance map[string]float64       `json:"significance"`
	
	// 统计检验
	PValues      map[string]float64       `json:"p_values"`
	EffectSizes  map[string]float64       `json:"effect_sizes"`
}

// NewReportGenerator 创建新的报告生成器
func NewReportGenerator(db *sql.DB) *ReportGenerator {
	rg := &ReportGenerator{
		db:           db,
		logger:       log.New(log.Writer(), "[Report] ", log.LstdFlags),
		queryBuilder: NewQueryBuilder(db),
		templates:    make(map[string]*template.Template),
	}
	
	// 初始化模板
	rg.initializeTemplates()
	
	return rg
}

// GenerateReport 生成报告
func (rg *ReportGenerator) GenerateReport(config ReportConfig) (*ReportResult, error) {
	startTime := time.Now()
	
	// 创建报告数据
	reportData, err := rg.buildReportData(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build report data")
	}
	
	// 确保输出目录存在
	outputDir := filepath.Dir(config.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create output directory")
	}
	
	// 生成报告文件
	var fileSize int64
	switch config.Format {
	case "html":
		fileSize, err = rg.generateHTMLReport(reportData, config)
	case "json":
		fileSize, err = rg.generateJSONReport(reportData, config)
	case "pdf":
		fileSize, err = rg.generatePDFReport(reportData, config)
	default:
		return nil, errors.Errorf("unsupported report format: %s", config.Format)
	}
	
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate report")
	}
	
	duration := time.Since(startTime)
	
	result := &ReportResult{
		ReportData:  reportData,
		OutputPath:  config.OutputPath,
		Format:      config.Format,
		GeneratedAt: startTime,
		Duration:    duration,
		SizeBytes:   fileSize,
	}
	
	rg.logger.Printf("报告生成完成: %s, 耗时: %v, 大小: %d bytes", config.OutputPath, duration, fileSize)
	return result, nil
}

// buildReportData 构建报告数据
func (rg *ReportGenerator) buildReportData(config ReportConfig) (*ReportData, error) {
	reportData := &ReportData{
		Title:        config.Title,
		Description:  config.Description,
		GeneratedAt:  time.Now(),
		Filters:      config.Parameters,
		Statistics:   make(map[string]interface{}),
		TimeSeries:   make([]TimeSeriesPoint, 0),
		Aggregations: make(map[string][]map[string]interface{}),
		Issues:       make([]*Issue, 0),
		Charts:       make(map[string]interface{}),
		Tables:       make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}
	
	// 生成基础统计
	stats, err := rg.generateStatistics(config)
	if err != nil {
		rg.logger.Printf("failed to generate statistics: %v", err)
	} else {
		reportData.Statistics = stats
	}
	
	// 生成时间序列数据
	if err := rg.generateTimeSeriesData(reportData, config); err != nil {
		rg.logger.Printf("failed to generate time series data: %v", err)
	}
	
	// 生成聚合数据
	if err := rg.generateAggregationData(reportData, config); err != nil {
		rg.logger.Printf("failed to generate aggregation data: %v", err)
	}
	
	// 生成Issues数据
	if err := rg.generateIssuesData(reportData, config); err != nil {
		rg.logger.Printf("failed to generate issues data: %v", err)
	}
	
	// 生成图表数据
	if err := rg.generateChartData(reportData, config); err != nil {
		rg.logger.Printf("failed to generate chart data: %v", err)
	}
	
	// 生成表格数据
	if err := rg.generateTableData(reportData, config); err != nil {
		rg.logger.Printf("failed to generate table data: %v", err)
	}
	
	// 添加元数据
	reportData.Metadata["config"] = config
	reportData.Metadata["query_count"] = len(reportData.TimeSeries) + len(reportData.Issues)
	
	return reportData, nil
}

// generateStatistics 生成统计信息
func (rg *ReportGenerator) generateStatistics(config ReportConfig) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 基础统计
	totalIssues, err := rg.getTotalIssuesCount()
	if err != nil {
		return nil, err
	}
	stats["total_issues"] = totalIssues
	
	openIssues, err := rg.getIssuesCountByState("open")
	if err != nil {
		return nil, err
	}
	stats["open_issues"] = openIssues
	
	closedIssues, err := rg.getIssuesCountByState("closed")
	if err != nil {
		return nil, err
	}
	stats["closed_issues"] = closedIssues
	
	pitfallIssues, err := rg.getPitfallIssuesCount()
	if err != nil {
		return nil, err
	}
	stats["pitfall_issues"] = pitfallIssues
	
	duplicateIssues, err := rg.getDuplicateIssuesCount()
	if err != nil {
		return nil, err
	}
	stats["duplicate_issues"] = duplicateIssues
	
	// 平均分数
	avgScore, err := rg.getAverageScore()
	if err != nil {
		return nil, err
	}
	stats["average_score"] = avgScore
	
	avgSeverity, err := rg.getAverageSeverity()
	if err != nil {
		return nil, err
	}
	stats["average_severity"] = avgSeverity
	
	// 分类统计
	categoryStats, err := rg.getCategoryStatistics()
	if err != nil {
		return nil, err
	}
	stats["category_distribution"] = categoryStats
	
	// 仓库统计
	repoStats, err := rg.getRepositoryStatistics()
	if err != nil {
		return nil, err
	}
	stats["repository_distribution"] = repoStats
	
	// 时间统计
	timeStats, err := rg.getTimeStatistics()
	if err != nil {
		return nil, err
	}
	stats["time_statistics"] = timeStats
	
	return stats, nil
}

// generateTimeSeriesData 生成时间序列数据
func (rg *ReportGenerator) generateTimeSeriesData(reportData *ReportData, config ReportConfig) error {
	// 获取时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // 默认过去一个月
	
	if dateFrom, ok := config.Parameters["date_from"]; ok {
		if dateFromTime, ok := dateFrom.(time.Time); ok {
			startDate = dateFromTime
		}
	}
	
	if dateTo, ok := config.Parameters["date_to"]; ok {
		if dateToTime, ok := dateTo.(time.Time); ok {
			endDate = dateToTime
		}
	}
	
	// 构建时间序列查询
	tsq := TimeSeriesQuery{
		StartDate:    startDate,
		EndDate:      endDate,
		Interval:     "day",
		Repositories: []string{},
		Categories:   []string{},
	}
	
	// 应用过滤条件
	if len(config.Parameters) > 0 {
		if repos, ok := config.Parameters["repositories"]; ok {
			if repoSlice, ok := repos.([]string); ok {
				tsq.Repositories = repoSlice
			}
		}
		
		if categories, ok := config.Parameters["categories"]; ok {
			if categorySlice, ok := categories.([]string); ok {
				tsq.Categories = categorySlice
			}
		}
	}
	
	// 执行查询
	tsResult, err := rg.queryBuilder.TimeSeriesQuery(tsq)
	if err != nil {
		return err
	}
	
	reportData.TimeSeries = tsResult.Points
	
	// 生成趋势分析
	if len(tsResult.Points) > 0 {
		trendAnalysis := rg.analyzeTrends(tsResult.Points, "new_issues")
		reportData.Metadata["trend_analysis"] = trendAnalysis
	}
	
	return nil
}

// generateAggregationData 生成聚合数据
func (rg *ReportGenerator) generateAggregationData(reportData *ReportData, config ReportConfig) error {
	// 分类聚合
	categoryAgg := AggregatedQuery{
		GroupBy:   "category",
		Metrics:   []string{"count", "avg_score", "avg_severity"},
		Filters:   DefaultSearchCriteria(),
	}
	
	if config.Parameters != nil {
		if filters, ok := config.Parameters["filters"]; ok {
			if filterMap, ok := filters.(map[string]interface{}); ok {
				// 转换过滤条件
				// 这里需要根据实际的参数结构进行转换
			}
		}
	}
	
	aggResult, err := rg.queryBuilder.AggregatedQuery(categoryAgg)
	if err != nil {
		return err
	}
	
	reportData.Aggregations["category"] = aggResult.Groups
	
	// 仓库聚合
	repoAgg := AggregatedQuery{
		GroupBy:   "repository",
		Metrics:   []string{"count", "avg_score"},
		Filters:   DefaultSearchCriteria(),
	}
	
	aggResult, err = rg.queryBuilder.AggregatedQuery(repoAgg)
	if err != nil {
		return err
	}
	
	reportData.Aggregations["repository"] = aggResult.Groups
	
	// 作者聚合
	authorAgg := AggregatedQuery{
		GroupBy:   "author",
		Metrics:   []string{"count", "avg_score"},
		Filters:   DefaultSearchCriteria(),
	}
	
	aggResult, err = rg.queryBuilder.AggregatedQuery(authorAgg)
	if err != nil {
		return err
	}
	
	reportData.Aggregations["author"] = aggResult.Groups
	
	return nil
}

// generateIssuesData 生成Issues数据
func (rg *ReportGenerator) generateIssuesData(reportData *ReportData, config ReportConfig) error {
	// 构建搜索条件
	criteria := DefaultSearchCriteria()
	criteria.Page = 1
	criteria.PageSize = 1000 // 默认获取前1000条
	
	// 应用过滤条件
	if config.Parameters != nil {
		// 这里可以根据配置参数调整搜索条件
	}
	
	// 执行查询
	queryResult, err := rg.queryBuilder.SimpleQuery(criteria)
	if err != nil {
		return err
	}
	
	reportData.Issues = queryResult.Issues
	
	return nil
}

// generateChartData 生成图表数据
func (rg *ReportGenerator) generateChartData(reportData *ReportData, config ReportConfig) error {
	reportData.Charts = make(map[string]interface{})
	
	for _, chartConfig := range config.Charts {
		var chartData interface{}
		var err error
		
		switch chartConfig.DataSource {
		case "time_series":
			chartData, err = rg.generateTimeSeriesChart(chartConfig)
		case "aggregation":
			chartData, err = rg.generateAggregationChart(chartConfig)
		case "statistics":
			chartData, err = rg.generateStatisticsChart(chartConfig)
		default:
			rg.logger.Printf("unknown chart data source: %s", chartConfig.DataSource)
			continue
		}
		
		if err != nil {
			rg.logger.Printf("failed to generate chart %s: %v", chartConfig.Title, err)
			continue
		}
		
		reportData.Charts[chartConfig.Title] = chartData
	}
	
	return nil
}

// generateTableData 生成表格数据
func (rg *ReportGenerator) generateTableData(reportData *ReportData, config ReportConfig) error {
	reportData.Tables = make(map[string]interface{})
	
	for _, tableConfig := range config.Tables {
		var tableData interface{}
		var err error
		
		switch tableConfig.DataSource {
		case "issues":
			tableData, err = rg.generateIssuesTable(tableConfig)
		case "aggregation":
			tableData, err = rg.generateAggregationTable(tableConfig)
		default:
			rg.logger.Printf("unknown table data source: %s", tableConfig.DataSource)
			continue
		}
		
		if err != nil {
			rg.logger.Printf("failed to generate table %s: %v", tableConfig.Title, err)
			continue
		}
		
		reportData.Tables[tableConfig.Title] = tableData
	}
	
	return nil
}

// generateHTMLReport 生成HTML报告
func (rg *ReportGenerator) generateHTMLReport(reportData *ReportData, config ReportConfig) (int64, error) {
	file, err := os.Create(config.OutputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	// HTML模板
	htmlTemplate := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { border-bottom: 2px solid #333; padding-bottom: 20px; margin-bottom: 30px; }
        .metric { display: inline-block; margin: 10px; padding: 15px; background: #f5f5f5; border-radius: 5px; }
        .chart { margin: 30px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .table { margin: 30px 0; overflow-x: auto; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        .summary { background: #e7f3ff; padding: 20px; border-radius: 5px; margin: 20px 0; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="header">
        <h1>{{ .Title }}</h1>
        <p>{{ .Description }}</p>
        <p><small>生成时间: {{ .GeneratedAt.Format "2006-01-02 15:04:05" }}</small></p>
    </div>
    
    <div class="summary">
        <h2>概览</h2>
        {{ range $key, $value := .Statistics }}
        <div class="metric">
            <strong>{{ $key }}:</strong> {{ $value }}
        </div>
        {{ end }}
    </div>
    
    {{ range $chartTitle, $chartData := .Charts }}
    <div class="chart">
        <h3>{{ $chartTitle }}</h3>
        <canvas id="chart-{{ $chartTitle }}" width="400" height="200"></canvas>
        <script>
            var ctx = document.getElementById('chart-{{ $chartTitle }}').getContext('2d');
            new Chart(ctx, {
                type: 'line',
                data: {{ $chartData }},
                options: { responsive: true }
            });
        </script>
    </div>
    {{ end }}
    
    {{ range $tableTitle, $tableData := .Tables }}
    <div class="table">
        <h3>{{ $tableTitle }}</h3>
        {{ $tableData }}
    </div>
    {{ end }}
    
    <div style="margin-top: 50px; padding-top: 20px; border-top: 1px solid #ddd; text-align: center; color: #666;">
        <small>Generated by GitHub Issues Analytics Platform</small>
    </div>
</body>
</html>
`
	
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return 0, err
	}
	
	if err := tmpl.Execute(file, reportData); err != nil {
		return 0, err
	}
	
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}
	
	return fileInfo.Size(), nil
}

// generateJSONReport 生成JSON报告
func (rg *ReportGenerator) generateJSONReport(reportData *ReportData, config ReportConfig) (int64, error) {
	file, err := os.Create(config.OutputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(reportData); err != nil {
		return 0, err
	}
	
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}
	
	return fileInfo.Size(), nil
}

// generatePDFReport 生成PDF报告
func (rg *ReportGenerator) generatePDFReport(reportData *ReportData, config ReportConfig) (int64, error) {
	// 这里可以实现PDF生成逻辑
	// 可以使用如 go-pdf 或其他PDF库
	return rg.generateHTMLReport(reportData, config) // 暂时返回HTML大小
}

// 辅助方法
func (rg *ReportGenerator) getTotalIssuesCount() (int, error) {
	var count int
	err := rg.db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	return count, err
}

func (rg *ReportGenerator) getIssuesCountByState(state string) (int, error) {
	var count int
	err := rg.db.QueryRow("SELECT COUNT(*) FROM issues WHERE state = ?", state).Scan(&count)
	return count, err
}

func (rg *ReportGenerator) getPitfallIssuesCount() (int, error) {
	var count int
	err := rg.db.QueryRow("SELECT COUNT(*) FROM issues WHERE is_pitfall = true").Scan(&count)
	return count, err
}

func (rg *ReportGenerator) getDuplicateIssuesCount() (int, error) {
	var count int
	err := rg.db.QueryRow("SELECT COUNT(*) FROM issues WHERE is_duplicate = true").Scan(&count)
	return count, err
}

func (rg *ReportGenerator) getAverageScore() (float64, error) {
	var avg float64
	err := rg.db.QueryRow("SELECT AVG(score) FROM issues").Scan(&avg)
	return avg, err
}

func (rg *ReportGenerator) getAverageSeverity() (float64, error) {
	var avg float64
	err := rg.db.QueryRow("SELECT AVG(severity_score) FROM issues").Scan(&avg)
	return avg, err
}

func (rg *ReportGenerator) getCategoryStatistics() (map[string]int, error) {
	stats := make(map[string]int)
	rows, err := rg.db.Query(`
		SELECT COALESCE(c.name, 'Unknown') as category, COUNT(*) as count
		FROM issues i
		LEFT JOIN categories c ON i.category_id = c.id
		GROUP BY c.name
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			continue
		}
		stats[category] = count
	}
	
	return stats, nil
}

func (rg *ReportGenerator) getRepositoryStatistics() (map[string]int, error) {
	stats := make(map[string]int)
	rows, err := rg.db.Query(`
		SELECT r.full_name, COUNT(*) as count
		FROM issues i
		JOIN repositories r ON i.repository_id = r.id
		GROUP BY r.full_name
		ORDER BY count DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var repo string
		var count int
		if err := rows.Scan(&repo, &count); err != nil {
			continue
		}
		stats[repo] = count
	}
	
	return stats, nil
}

func (rg *ReportGenerator) getTimeStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 最近7天的统计
	var lastWeek int
	err := rg.db.QueryRow(`
		SELECT COUNT(*) FROM issues 
		WHERE created_at >= datetime('now', '-7 days')
	`).Scan(&lastWeek)
	if err != nil {
		return nil, err
	}
	stats["last_week"] = lastWeek
	
	// 最近30天的统计
	var lastMonth int
	err = rg.db.QueryRow(`
		SELECT COUNT(*) FROM issues 
		WHERE created_at >= datetime('now', '-30 days')
	`).Scan(&lastMonth)
	if err != nil {
		return nil, err
	}
	stats["last_month"] = lastMonth
	
	// 平均每日新增
	avgDaily, err := rg.db.QueryRow(`
		SELECT AVG(daily_count) FROM (
			SELECT DATE(created_at) as date, COUNT(*) as daily_count
			FROM issues
			GROUP BY DATE(created_at)
		)
	`).Scan(&avgDaily)
	if err != nil {
		return nil, err
	}
	stats["avg_daily_new"] = avgDaily
	
	return stats, nil
}

// analyzeTrends 分析趋势
func (rg *ReportGenerator) analyzeTrends(data []TimeSeriesPoint, metric string) *TrendAnalysis {
	if len(data) == 0 {
		return &TrendAnalysis{Metric: metric}
	}
	
	// 提取指标值
	values := make([]float64, len(data))
	for i, point := range data {
		if val, ok := point.Metrics[metric]; ok {
			values[i] = val
		} else {
			values[i] = 0
		}
	}
	
	// 计算统计指标
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// 计算标准差
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	stdDev := math.Sqrt(variance / float64(len(data)))
	
	// 计算增长率
	growthRate := 0.0
	if len(values) > 1 {
		if values[0] != 0 {
			growthRate = (values[len(values)-1] - values[0]) / values[0] * 100
		}
	}
	
	// 确定趋势方向
	trend := "stable"
	if growthRate > 5 {
		trend = "increasing"
	} else if growthRate < -5 {
		trend = "decreasing"
	}
	
	// 计算中位数
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	for i := 0; i < len(sortedValues)/2; i++ {
		j := len(sortedValues) - 1 - i
		sortedValues[i], sortedValues[j] = sortedValues[j], sortedValues[i]
	}
	
	median := 0.0
	if len(sortedValues)%2 == 0 {
		median = (sortedValues[len(sortedValues)/2-1] + sortedValues[len(sortedValues)/2]) / 2
	} else {
		median = sortedValues[len(sortedValues)/2]
	}
	
	// 计算波动性
	volatility := stdDev / mean * 100
	
	return &TrendAnalysis{
		Metric:      metric,
		Data:        data,
		GrowthRate:  growthRate,
		Volatility:  volatility,
		Trend:       trend,
		Mean:        mean,
		Median:      median,
		StdDev:      stdDev,
		Min:         values[0],
		Max:         values[len(values)-1],
		Confidence:  0.85, // 默认置信度
	}
}

// 生成图表和表格的辅助方法
func (rg *ReportGenerator) generateTimeSeriesChart(config ChartConfig) (interface{}, error) {
	// 实现时间序列图表数据生成
	return nil, nil
}

func (rg *ReportGenerator) generateAggregationChart(config ChartConfig) (interface{}, error) {
	// 实现聚合图表数据生成
	return nil, nil
}

func (rg *ReportGenerator) generateStatisticsChart(config ChartConfig) (interface{}, error) {
	// 实现统计图表数据生成
	return nil, nil
}

func (rg *ReportGenerator) generateIssuesTable(config TableConfig) (interface{}, error) {
	// 实现Issues表格数据生成
	return nil, nil
}

func (rg *ReportGenerator) generateAggregationTable(config TableConfig) (interface{}, error) {
	// 实现聚合表格数据生成
	return nil, nil
}

// initializeTemplates 初始化模板
func (rg *ReportGenerator) initializeTemplates() {
	// 初始化HTML模板等
}

// CompareReports 比较报告
func (rg *ReportGenerator) CompareReports(config1, config2 ReportConfig) (*ComparisonAnalysis, error) {
	// 生成两个报告的数据
	data1, err := rg.buildReportData(config1)
	if err != nil {
		return nil, err
	}
	
	data2, err := rg.buildReportData(config1)
	if err != nil {
		return nil, err
	}
	
	// 比较分析逻辑
	analysis := &ComparisonAnalysis{
		Metric: "count",
		Groups: []string{"report1", "report2"},
		Data: map[string][]TimeSeriesPoint{
			"report1": data1.TimeSeries,
			"report2": data2.TimeSeries,
		},
	}
	
	// 计算差异和比率
	analysis.Differences = make(map[string]float64)
	analysis.Ratios = make(map[string]float64)
	
	return analysis, nil
}

// ExportToMultipleFormats 导出多种格式
func (rg *ReportGenerator) ExportToMultipleFormats(config ReportConfig) (map[string]*ReportResult, error) {
	results := make(map[string]*ReportResult)
	
	formats := []string{"html", "json"}
	if config.Format != "" {
		formats = []string{config.Format}
	}
	
	for _, format := range formats {
		configCopy := config
		configCopy.Format = format
		
		// 修改输出路径
		ext := filepath.Ext(configCopy.OutputPath)
		baseName := strings.TrimSuffix(configCopy.OutputPath, ext)
		configCopy.OutputPath = fmt.Sprintf("%s_%s%s", baseName, format, ext)
		
		result, err := rg.GenerateReport(configCopy)
		if err != nil {
			rg.logger.Printf("failed to generate %s report: %v", format, err)
			continue
		}
		
		results[format] = result
	}
	
	return results, nil
}