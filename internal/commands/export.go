package commands

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gh-pitfall-scraper/internal/database"
)

// ExportCommand 导出命令
type ExportCommand struct {
	db            *sql.DB
	exporter      *database.Exporter
	queryBuilder  *database.QueryBuilder
	reportGen     *database.ReportGenerator
	logger        *log.Logger
}

// ExportOptions 导出选项
type ExportOptions struct {
	// 基础选项
	Format       string            `json:"format"`       // "json", "csv", "md"
	OutputPath   string            `json:"output_path"`
	TemplatePath string            `json:"template_path"`
	
	// 过滤选项
	DateFrom     string            `json:"date_from"`
	DateTo       string            `json:"date_to"`
	Categories   []string          `json:"categories"`
	Tags         []string          `json:"tags"`
	Repositories []string          `json:"repositories"`
	States       []string          `json:"states"`
	MinScore     *float64          `json:"min_score"`
	MaxScore     *float64          `json:"max_score"`
	IsPitfall    *bool             `json:"is_pitfall"`
	IsDuplicate  *bool             `json:"is_duplicate"`
	
	// 查询选项
	Query        string            `json:"query"`
	Keywords     []string          `json:"keywords"`
	Authors      []string          `json:"authors"`
	Assignees    []string          `json:"assignees"`
	
	// 分页选项
	Page         int               `json:"page"`
	PageSize     int               `json:"page_size"`
	SortBy       string            `json:"sort_by"`
	SortOrder    string            `json:"sort_order"`
	
	// 报告选项
	GenerateReport bool            `json:"generate_report"`
	ReportTitle    string          `json:"report_title"`
	ReportFormat   string          `json:"report_format"` // "html", "pdf", "json"
	
	// 其他选项
	Verbose       bool             `json:"verbose"`
	IncludeMetadata bool           `json:"include_metadata"`
	ExcludeDuplicates bool         `json:"exclude_duplicates"`
}

// ReportOptions 报告选项
type ReportOptions struct {
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	OutputPath     string                 `json:"output_path"`
	Format         string                 `json:"format"` // "html", "pdf", "json"
	
	// 图表配置
	Charts         []ChartConfig          `json:"charts"`
	
	// 表格配置
	Tables         []TableConfig          `json:"tables"`
	
	// 时间序列配置
	TimeSeries     bool                   `json:"time_series"`
	StartDate      string                 `json:"start_date"`
	EndDate        string                 `json:"end_date"`
	Interval       string                 `json:"interval"` // "day", "week", "month"
	
	// 聚合配置
	Aggregations   []string               `json:"aggregations"` // "category", "repository", "author"
	
	// 过滤条件
	Filters        database.SearchCriteria `json:"filters"`
	
	// 高级选项
	Template       string                 `json:"template"`
	IncludeCharts  bool                   `json:"include_charts"`
	IncludeTables  bool                   `json:"include_tables"`
}

// ChartConfig 图表配置
type ChartConfig struct {
	Type          string                 `json:"type"` // "line", "bar", "pie"
	Title         string                 `json:"title"`
	DataSource    string                 `json:"data_source"` // "time_series", "aggregation"
	Metric        string                 `json:"metric"`
	GroupBy       string                 `json:"group_by"`
	Repository    string                 `json:"repository"`
	Category      string                 `json:"category"`
}

// TableConfig 表格配置
type TableConfig struct {
	Title          string                 `json:"title"`
	Columns        []string               `json:"columns"`
	DataSource     string                 `json:"data_source"` // "issues", "aggregation"
	SortBy         string                 `json:"sort_by"`
	Limit          int                    `json:"limit"`
}

// NewExportCommand 创建新的导出命令
func NewExportCommand(db *sql.DB) *ExportCommand {
	return &ExportCommand{
		db:            db,
		exporter:      database.NewExporter(db),
		queryBuilder:  database.NewQueryBuilder(db),
		reportGen:     database.NewReportGenerator(db),
		logger:        log.New(log.Writer(), "[ExportCmd] ", log.LstdFlags),
	}
}

// Run 运行导出命令
func (ec *ExportCommand) Run(args []string) error {
	// 解析命令行参数
	opts, err := ec.parseFlags(args)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	
	// 验证选项
	if err := ec.validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	
	// 执行导出
	if err := ec.executeExport(opts); err != nil {
		return fmt.Errorf("failed to execute export: %w", err)
	}
	
	return nil
}

// parseFlags 解析命令行参数
func (ec *ExportCommand) parseFlags(args []string) (*ExportOptions, error) {
	flags := flag.NewFlagSet("export", flag.ExitOnError)
	
	// 基础选项
	format := flags.String("format", "json", "导出格式 (json, csv, md)")
	outputPath := flags.String("output", "", "输出文件路径 (必需)")
	templatePath := flags.String("template", "", "模板文件路径")
	
	// 过滤选项
	dateFrom := flags.String("date-from", "", "开始日期 (YYYY-MM-DD)")
	dateTo := flags.String("date-to", "", "结束日期 (YYYY-MM-DD)")
	categories := flags.String("categories", "", "分类列表 (逗号分隔)")
	tags := flags.String("tags", "", "标签列表 (逗号分隔)")
	repositories := flags.String("repos", "", "仓库列表 (逗号分隔, owner/name)")
	states := flags.String("states", "", "状态列表 (open, closed)")
	minScore := flags.Float64("min-score", 0, "最小分数")
	maxScore := flags.Float64("max-score", 10, "最大分数")
	isPitfall := flags.Bool("pitfall", false, "仅导出坑点问题")
	isDuplicate := flags.Bool("duplicate", false, "仅导出重复问题")
	
	// 查询选项
	query := flags.String("query", "", "文本搜索查询")
	keywords := flags.String("keywords", "", "关键词列表 (逗号分隔)")
	authors := flags.String("authors", "", "作者列表 (逗号分隔)")
	assignees := flags.String("assignees", "", "指派人列表 (逗号分隔)")
	
	// 分页选项
	page := flags.Int("page", 1, "页码")
	pageSize := flags.Int("page-size", 1000, "页面大小")
	sortBy := flags.String("sort-by", "score", "排序字段")
	sortOrder := flags.String("sort-order", "DESC", "排序方向 (ASC, DESC)")
	
	// 报告选项
	generateReport := flags.Bool("report", false, "生成报告")
	reportTitle := flags.String("report-title", "Issues Export Report", "报告标题")
	reportFormat := flags.String("report-format", "html", "报告格式 (html, pdf, json)")
	
	// 其他选项
	verbose := flags.Bool("verbose", false, "详细输出")
	includeMetadata := flags.Bool("include-metadata", false, "包含元数据")
	excludeDuplicates := flags.Bool("exclude-duplicates", false, "排除重复问题")
	
	// 解析标志
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	
	// 转换为选项结构
	opts := &ExportOptions{
		Format:           *format,
		OutputPath:       *outputPath,
		TemplatePath:     *templatePath,
		DateFrom:         *dateFrom,
		DateTo:           *dateTo,
		Categories:       parseStringList(*categories),
		Tags:             parseStringList(*tags),
		Repositories:     parseStringList(*repositories),
		States:           parseStringList(*states),
		Query:            *query,
		Keywords:         parseStringList(*keywords),
		Authors:          parseStringList(*authors),
		Assignees:        parseStringList(*assignees),
		Page:             *page,
		PageSize:         *pageSize,
		SortBy:           *sortBy,
		SortOrder:        *sortOrder,
		GenerateReport:   *generateReport,
		ReportTitle:      *reportTitle,
		ReportFormat:     *reportFormat,
		Verbose:          *verbose,
		IncludeMetadata:  *includeMetadata,
		ExcludeDuplicates: *excludeDuplicates,
	}
	
	// 设置数值选项
	if *minScore > 0 {
		opts.MinScore = float64Ptr(*minScore)
	}
	if *maxScore < 10 {
		opts.MaxScore = float64Ptr(*maxScore)
	}
	if *isPitfall {
		opts.IsPitfall = boolPtr(true)
	}
	if *isDuplicate {
		opts.IsDuplicate = boolPtr(true)
	}
	
	return opts, nil
}

// validateOptions 验证选项
func (ec *ExportCommand) validateOptions(opts *ExportOptions) error {
	if opts.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	
	// 验证格式
	validFormats := map[string]bool{"json": true, "csv": true, "md": true}
	if !validFormats[opts.Format] {
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
	
	// 验证排序字段
	validSortFields := map[string]bool{
		"score": true, "severity": true, "created": true, "updated": true,
		"comments": true, "title": true, "number": true,
	}
	if !validSortFields[opts.SortBy] {
		return fmt.Errorf("invalid sort field: %s", opts.SortBy)
	}
	
	// 验证排序方向
	if opts.SortOrder != "ASC" && opts.SortOrder != "DESC" {
		return fmt.Errorf("sort order must be ASC or DESC")
	}
	
	// 验证分页参数
	if opts.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if opts.PageSize < 1 || opts.PageSize > 10000 {
		return fmt.Errorf("page size must be between 1 and 10000")
	}
	
	return nil
}

// executeExport 执行导出
func (ec *ExportCommand) executeExport(opts *ExportOptions) error {
	if opts.Verbose {
		ec.logger.Printf("开始导出，格式: %s, 输出: %s", opts.Format, opts.OutputPath)
	}
	
	// 构建过滤条件
	filter := database.ExportFilter{
		Categories:      opts.Categories,
		Tags:            opts.Tags,
		Repositories:    opts.Repositories,
		States:          opts.States,
		MinScore:        opts.MinScore,
		MaxScore:        opts.MaxScore,
		IsPitfall:       opts.IsPitfall,
		IsDuplicate:     opts.IsDuplicate,
		IncludeMetadata: opts.IncludeMetadata,
	}
	
	// 处理日期
	if opts.DateFrom != "" {
		if date, err := time.Parse("2006-01-02", opts.DateFrom); err != nil {
			return fmt.Errorf("invalid date-from format: %s", opts.DateFrom)
		} else {
			filter.DateFrom = &date
		}
	}
	
	if opts.DateTo != "" {
		if date, err := time.Parse("2006-01-02", opts.DateTo); err != nil {
			return fmt.Errorf("invalid date-to format: %s", opts.DateTo)
		} else {
			filter.DateTo = &date
		}
	}
	
	// 转换为导出格式
	var format database.ExportFormat
	switch opts.Format {
	case "json":
		format = database.FormatJSON
	case "csv":
		format = database.FormatCSV
	case "md":
		format = database.FormatMarkdown
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
	
	// 执行导出
	result, err := ec.exporter.ExportIssues(filter, format, opts.OutputPath)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	
	if opts.Verbose {
		ec.logger.Printf("导出完成:")
		ec.logger.Printf("  总记录数: %d", result.TotalRecords)
		ec.logger.Printf("  导出记录数: %d", result.ExportedRecords)
		ec.logger.Printf("  耗时: %v", result.Duration)
		ec.logger.Printf("  输出文件: %s", result.OutputPath)
	}
	
	// 如果需要生成报告
	if opts.GenerateReport {
		if err := ec.generateReport(opts, result); err != nil {
			return fmt.Errorf("report generation failed: %w", err)
		}
	}
	
	return nil
}

// generateReport 生成报告
func (ec *ExportCommand) generateReport(opts *ExportOptions, exportResult *database.ExportResult) error {
	if opts.Verbose {
		ec.logger.Printf("开始生成报告: %s", opts.ReportTitle)
	}
	
	// 确定报告输出路径
	reportOutputPath := opts.OutputPath
	if opts.ReportFormat != "" {
		ext := fmt.Sprintf(".%s", opts.ReportFormat)
		if !strings.HasSuffix(reportOutputPath, ext) {
			baseName := strings.TrimSuffix(reportOutputPath, filepath.Ext(reportOutputPath))
			reportOutputPath = baseName + "_report" + ext
		}
	}
	
	// 构建报告配置
	config := database.ReportConfig{
		Title:       opts.ReportTitle,
		Description: fmt.Sprintf("Generated from export of %d issues", exportResult.ExportedRecords),
		OutputPath:  reportOutputPath,
		Format:      opts.ReportFormat,
		Parameters: map[string]interface{}{
			"export_result": exportResult,
			"filters":       exportResult.Filter,
		},
		Charts: []database.ChartConfig{
			{
				Type:       "line",
				Title:      "Issues Over Time",
				DataSource: "time_series",
				XAxis:      "timestamp",
				YAxis:      "count",
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
	reportResult, err := ec.reportGen.GenerateReport(config)
	if err != nil {
		return err
	}
	
	if opts.Verbose {
		ec.logger.Printf("报告生成完成:")
		ec.logger.Printf("  报告文件: %s", reportResult.OutputPath)
		ec.logger.Printf("  格式: %s", reportResult.Format)
		ec.logger.Printf("  大小: %d bytes", reportResult.SizeBytes)
	}
	
	return nil
}

// RunReportCommand 运行报告命令
func (ec *ExportCommand) RunReportCommand(args []string) error {
	// 解析报告命令行参数
	opts, err := ec.parseReportFlags(args)
	if err != nil {
		return fmt.Errorf("failed to parse report flags: %w", err)
	}
	
	// 验证选项
	if err := ec.validateReportOptions(opts); err != nil {
		return fmt.Errorf("invalid report options: %w", err)
	}
	
	// 生成报告
	if err := ec.executeReport(opts); err != nil {
		return fmt.Errorf("failed to execute report: %w", err)
	}
	
	return nil
}

// parseReportFlags 解析报告命令行参数
func (ec *ExportCommand) parseReportFlags(args []string) (*ReportOptions, error) {
	flags := flag.NewFlagSet("report", flag.ExitOnError)
	
	// 基础选项
	title := flags.String("title", "Analytics Report", "报告标题")
	description := flags.String("description", "", "报告描述")
	outputPath := flags.String("output", "", "输出文件路径 (必需)")
	format := flags.String("format", "html", "报告格式 (html, pdf, json)")
	
	// 图表选项
	includeCharts := flags.Bool("charts", true, "包含图表")
	chartTypes := flags.String("chart-types", "line,bar", "图表类型")
	metrics := flags.String("metrics", "count,avg_score", "指标列表")
	
	// 表格选项
	includeTables := flags.Bool("tables", true, "包含表格")
	tableLimit := flags.Int("table-limit", 100, "表格行数限制")
	
	// 时间序列选项
	timeSeries := flags.Bool("time-series", true, "包含时间序列")
	startDate := flags.String("start-date", "", "开始日期")
	endDate := flags.String("end-date", "", "结束日期")
	interval := flags.String("interval", "day", "时间间隔 (day, week, month)")
	
	// 聚合选项
	aggregations := flags.String("aggregations", "category,repository", "聚合维度")
	
	// 过滤选项
	query := flags.String("query", "", "搜索查询")
	categories := flags.String("categories", "", "分类过滤")
	repositories := flags.String("repos", "", "仓库过滤")
	minScore := flags.Float64("min-score", 0, "最小分数")
	
	// 解析标志
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	
	// 转换为选项结构
	opts := &ReportOptions{
		Title:          *title,
		Description:    *description,
		OutputPath:     *outputPath,
		Format:         *format,
		IncludeCharts:  *includeCharts,
		IncludeTables:  *includeTables,
		TimeSeries:     *timeSeries,
		Interval:       *interval,
		Template:       "",
	}
	
	// 构建图表配置
	if opts.IncludeCharts {
		chartTypesList := parseStringList(*chartTypes)
		metricsList := parseStringList(*metrics)
		
		for _, chartType := range chartTypesList {
			for _, metric := range metricsList {
				config := ChartConfig{
					Type:       chartType,
					Title:      fmt.Sprintf("%s by %s", metric, chartType),
					DataSource: "time_series",
					Metric:     metric,
				}
				opts.Charts = append(opts.Charts, config)
			}
		}
	}
	
	// 构建表格配置
	if opts.IncludeTables {
		opts.Tables = []TableConfig{
			{
				Title:      "Issues Summary",
				Columns:    []string{"id", "title", "score", "state", "created_at"},
				DataSource: "issues",
				SortBy:     "score",
				Limit:      *tableLimit,
			},
		}
	}
	
	// 构建过滤条件
	filters := database.DefaultSearchCriteria()
	if *query != "" {
		filters.Query = *query
	}
	if *categories != "" {
		filters.Categories = parseStringList(*categories)
	}
	if *repositories != "" {
		filters.Repositories = parseStringList(*repositories)
	}
	if *minScore > 0 {
		filters.MinScore = float64Ptr(*minScore)
	}
	
	opts.Filters = filters
	
	return opts, nil
}

// validateReportOptions 验证报告选项
func (ec *ExportCommand) validateReportOptions(opts *ReportOptions) error {
	if opts.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	
	validFormats := map[string]bool{"html": true, "pdf": true, "json": true}
	if !validFormats[opts.Format] {
		return fmt.Errorf("invalid report format: %s", opts.Format)
	}
	
	return nil
}

// executeReport 执行报告生成
func (ec *ExportCommand) executeReport(opts *ReportOptions) error {
	if opts.Verbose {
		ec.logger.Printf("开始生成报告: %s", opts.Title)
	}
	
	// 构建报告配置
	config := database.ReportConfig{
		Title:       opts.Title,
		Description: opts.Description,
		OutputPath:  opts.OutputPath,
		Format:      opts.Format,
		Parameters: map[string]interface{}{
			"start_date": opts.StartDate,
			"end_date":   opts.EndDate,
			"interval":   opts.Interval,
			"filters":    opts.Filters,
		},
		Charts: make([]database.ChartConfig, 0),
		Tables: make([]database.TableConfig, 0),
	}
	
	// 添加图表配置
	for _, chart := range opts.Charts {
		config.Charts = append(config.Charts, database.ChartConfig{
			Type:       chart.Type,
			Title:      chart.Title,
			DataSource: chart.DataSource,
			XAxis:      "timestamp",
			YAxis:      chart.Metric,
			GroupBy:    chart.GroupBy,
		})
	}
	
	// 添加表格配置
	for _, table := range opts.Tables {
		config.Tables = append(config.Tables, database.TableConfig{
			Title:      table.Title,
			DataSource: table.DataSource,
			SortBy:     table.SortBy,
			Limit:      table.Limit,
		})
	}
	
	// 生成报告
	_, err := ec.reportGen.GenerateReport(config)
	if err != nil {
		return err
	}
	
	return nil
}

// 辅助函数
func parseStringList(s string) []string {
	if s == "" {
		return nil
	}
	
	var result []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

// PrintHelp 打印帮助信息
func (ec *ExportCommand) PrintHelp() {
	fmt.Println(`导出命令用法:

导出Issues数据:
  ./gh-pitfall-scraper export [选项]

选项:
  --format string      导出格式 (json, csv, md) [默认: json]
  --output string      输出文件路径 (必需)
  --template string    模板文件路径
  
  过滤选项:
  --date-from string   开始日期 (YYYY-MM-DD)
  --date-to string     结束日期 (YYYY-MM-DD)
  --categories string  分类列表 (逗号分隔)
  --tags string        标签列表 (逗号分隔)
  --repos string       仓库列表 (逗号分隔, owner/name)
  --states string      状态列表 (open, closed)
  --min-score float64  最小分数
  --max-score float64  最大分数
  --pitfall           仅导出坑点问题
  --duplicate         仅导出重复问题
  
  查询选项:
  --query string       文本搜索查询
  --keywords string    关键词列表 (逗号分隔)
  --authors string     作者列表 (逗号分隔)
  --assignees string   指派人列表 (逗号分隔)
  
  分页选项:
  --page int          页码 [默认: 1]
  --page-size int     页面大小 [默认: 1000]
  --sort-by string    排序字段 [默认: score]
  --sort-order string 排序方向 (ASC, DESC) [默认: DESC]
  
  报告选项:
  --report            生成报告
  --report-title string 报告标题 [默认: Issues Export Report]
  --report-format string 报告格式 (html, pdf, json) [默认: html]
  
  其他选项:
  --verbose           详细输出
  --include-metadata  包含元数据
  --exclude-duplicates 排除重复问题

示例:
  # 导出JSON格式
  ./gh-pitfall-scraper export --format json --output issues.json
  
  # 导出CSV格式并按时间过滤
  ./gh-pitfall-scraper export --format csv --output issues.csv --date-from 2024-01-01 --date-to 2024-12-31
  
  # 导出特定仓库的问题
  ./gh-pitfall-scraper export --format md --output report.md --repos "owner1/repo1,owner2/repo2"
  
  # 生成带报告的导出
  ./gh-pitfall-scraper export --format json --output data.json --report --report-title "Monthly Report"
`)
}