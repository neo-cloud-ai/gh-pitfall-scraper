// gh-pitfall-scraper 主程序 - 数据库集成版本
// 包含完整的数据库配置、初始化、监控和维护功能
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gh-pitfall-scraper/internal/database"
	"gh-pitfall-scraper/internal/scraper"
	
	_ "github.com/lib/pq" // PostgreSQL驱动
	"gopkg.in/yaml.v3"
)

// 版本信息
const (
	AppName    = "gh-pitfall-scraper"
	AppVersion = "2.0.0"
	AppAuthor  = "neo-cloud-ai"
)

// 全局变量
var (
	logger *log.Logger
	appCtx context.Context
	cancel context.CancelFunc
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// 连接配置
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	
	// 连接池配置
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	
	// 缓存配置
	CacheEnabled bool `yaml:"cache_enabled"`
	CacheSize    int  `yaml:"cache_size"`
	CacheTTL     time.Duration `yaml:"cache_ttl"`
	
	// 自动清理配置
	AutoCleanupEnabled bool          `yaml:"auto_cleanup_enabled"`
	CleanupInterval    time.Duration `yaml:"cleanup_interval"`
	DataRetention      time.Duration `yaml:"data_retention"`
	
	// 备份配置
	BackupEnabled  bool          `yaml:"backup_enabled"`
	BackupInterval time.Duration `yaml:"backup_interval"`
	BackupPath     string        `yaml:"backup_path"`
	RetentionDays  int           `yaml:"retention_days"`
}

// Config 主配置结构体
type Config struct {
	// 应用配置
	App struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		LogLevel    string `yaml:"log_level"`
		DataDir     string `yaml:"data_dir"`
		OutputDir   string `yaml:"output_dir"`
		MaxWorkers  int    `yaml:"max_workers"`
		WorkerQueue int    `yaml:"worker_queue"`
	} `yaml:"app"`
	
	// GitHub配置
	GithubToken string `yaml:"github_token"`
	RequestInterval int `yaml:"request_interval"`
	Timeout        int `yaml:"timeout"`
	
	// 爬虫配置
	Repos []struct {
		Owner string `yaml:"owner"`
		Name  string `yaml:"name"`
	} `yaml:"repos"`
	
	Keywords []string `yaml:"keywords"`
	
	// 数据库配置
	Database DatabaseConfig `yaml:"database"`
}

// DatabaseManager 数据库管理器包装
type DatabaseManager struct {
	*database.DatabaseManager
	analytics    *database.DatabaseAnalytics
	maintenance  *database.DatabaseMaintenance
	config       *DatabaseConfig
	logger       *log.Logger
	stopCh       chan struct{}
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(dbConfig *DatabaseConfig, logger *log.Logger) (*DatabaseManager, error) {
	if logger == nil {
		logger = log.New(os.Stdout, "[DB] ", log.LstdFlags|log.Lmsgprefix)
	}
	
	// 配置数据库管理器
	config := &database.Config{
		Host:            dbConfig.Host,
		Port:            dbConfig.Port,
		User:            dbConfig.User,
		Password:        dbConfig.Password,
		DBName:          dbConfig.DBName,
		SSLMode:         dbConfig.SSLMode,
		MaxOpenConns:    dbConfig.MaxOpenConns,
		MaxIdleConns:    dbConfig.MaxIdleConns,
		ConnMaxLifetime: dbConfig.ConnMaxLifetime,
		ConnMaxIdleTime: dbConfig.ConnMaxIdleTime,
	}
	
	// 创建数据库管理器
	dbManager, err := database.NewDatabaseManager(config, logger)
	if err != nil {
		return nil, fmt.Errorf("创建数据库管理器失败: %w", err)
	}
	
	// 创建分析工具
	analyticsConfig := &database.AnalyticsConfig{
		DataRetentionDays: int(dbConfig.DataRetention.Hours() / 24),
		CollectionInterval: 1 * time.Hour,
		BatchSize:         1000,
		ReportFormats:     []string{"json"},
		ReportDirectory:   "./reports",
		CacheEnabled:      dbConfig.CacheEnabled,
		CacheTTL:          dbConfig.CacheTTL,
	}
	
	analytics, err := database.NewDatabaseAnalytics(dbManager.GetDB(), analyticsConfig, logger)
	if err != nil {
		logger.Printf("创建数据库分析工具失败: %v", err)
	}
	
	// 创建维护工具
	maintenanceConfig := &database.MaintenanceConfig{
		AutoCleanupEnabled:    dbConfig.AutoCleanupEnabled,
		CleanupInterval:       dbConfig.CleanupInterval,
		RetentionPeriod:       dbConfig.DataRetention,
		IndexOptimizationEnabled: true,
		AnalyzeEnabled:        true,
		VacuumEnabled:         true,
	}
	
	maintenance, err := database.NewDatabaseMaintenance(dbManager.GetDB(), maintenanceConfig, logger)
	if err != nil {
		logger.Printf("创建数据库维护工具失败: %v", err)
	}
	
	return &DatabaseManager{
		DatabaseManager: dbManager,
		analytics:       analytics,
		maintenance:     maintenance,
		config:          dbConfig,
		logger:          logger,
		stopCh:          make(chan struct{}),
	}, nil
}

// Start 启动数据库管理器
func (dm *DatabaseManager) Start(ctx context.Context) error {
	dm.logger.Println("启动数据库管理器...")
	
	// 健康检查
	if err := dm.HealthCheck(ctx); err != nil {
		return fmt.Errorf("数据库健康检查失败: %w", err)
	}
	
	// 启动维护任务
	if dm.maintenance != nil {
		if err := dm.maintenance.Start(ctx); err != nil {
			dm.logger.Printf("启动维护任务失败: %v", err)
		} else {
			dm.logger.Println("数据库维护任务已启动")
		}
	}
	
	// 启动分析任务
	if dm.analytics != nil {
		if err := dm.analytics.Start(ctx); err != nil {
			dm.logger.Printf("启动分析任务失败: %v", err)
		} else {
			dm.logger.Println("数据库分析任务已启动")
		}
	}
	
	// 启动监控协程
	go dm.monitor(ctx)
	
	dm.logger.Println("数据库管理器启动完成")
	return nil
}

// monitor 监控数据库状态
func (dm *DatabaseManager) monitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 健康检查
			if err := dm.HealthCheck(ctx); err != nil {
				dm.logger.Printf("数据库健康检查失败: %v", err)
			}
			
			// 连接池统计
			stats := dm.GetStats()
			if stats.OpenConnections > 0 {
				dm.logger.Printf("连接池状态: 开放=%d, 使用中=%d, 空闲=%d",
					stats.OpenConnections, stats.InUse, stats.Idle)
			}
		}
	}
}

// Stop 停止数据库管理器
func (dm *DatabaseManager) Stop() error {
	close(dm.stopCh)
	
	dm.logger.Println("停止数据库管理器...")
	
	// 停止维护任务
	if dm.maintenance != nil {
		dm.maintenance.Stop()
	}
	
	// 停止分析任务
	if dm.analytics != nil {
		dm.analytics.Stop()
	}
	
	// 关闭数据库连接
	if err := dm.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}
	
	dm.logger.Println("数据库管理器已停止")
	return nil
}

// ShowStats 显示数据库统计信息
func (dm *DatabaseManager) ShowStats() {
	stats := dm.GetStats()
	
	fmt.Println("=== 数据库连接池统计 ===")
	fmt.Printf("开放连接: %d\n", stats.OpenConnections)
	fmt.Printf("使用中: %d\n", stats.InUse)
	fmt.Printf("空闲: %d\n", stats.Idle)
	fmt.Printf("等待次数: %d\n", stats.WaitCount)
	fmt.Printf("等待时间: %v\n", stats.WaitDuration)
	fmt.Printf("最大空闲关闭: %d\n", stats.MaxIdleClosed)
	fmt.Printf("最大生命周期关闭: %d\n", stats.MaxLifetimeClosed)
	fmt.Println()
}

// AppConfig 应用配置
type AppConfig struct {
	ConfigFile string
	ShowVersion bool
	ShowHelp    bool
	DatabaseOnly bool
	Backup      bool
	Restore     string
	Stats       bool
	HealthCheck bool
	Debug       bool
	
	// 导出相关选项
	Export       bool
	ExportFormat string
	ExportOutput string
	Report       bool
	ReportFormat string
	ReportTitle  string
}

// parseCommandLine 解析命令行参数
func parseCommandLine() *AppConfig {
	var config AppConfig
	
	flag.StringVar(&config.ConfigFile, "config", "config.yaml", "配置文件路径")
	flag.BoolVar(&config.ShowVersion, "version", false, "显示版本信息")
	flag.BoolVar(&config.ShowHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&config.DatabaseOnly, "db-only", false, "仅初始化数据库，不执行爬虫")
	flag.BoolVar(&config.Backup, "backup", false, "执行数据库备份")
	flag.StringVar(&config.Restore, "restore", "", "从备份文件恢复数据库")
	flag.BoolVar(&config.Stats, "stats", false, "显示数据库统计信息")
	flag.BoolVar(&config.HealthCheck, "health", false, "执行数据库健康检查")
	flag.BoolVar(&config.Debug, "debug", false, "启用调试模式")
	
	// 导出相关选项
	flag.BoolVar(&config.Export, "export", false, "导出数据")
	flag.StringVar(&config.ExportFormat, "export-format", "json", "导出格式 (json, csv, md)")
	flag.StringVar(&config.ExportOutput, "output", "", "导出输出文件路径")
	flag.BoolVar(&config.Report, "report", false, "生成报告")
	flag.StringVar(&config.ReportFormat, "report-format", "html", "报告格式 (html, pdf, json)")
	flag.StringVar(&config.ReportTitle, "report-title", "Issues Analytics Report", "报告标题")
	
	flag.Parse()
	
	return &config
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf("%s v%s\n", AppName, AppVersion)
	fmt.Printf("作者: %s\n", AppAuthor)
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s [选项]\n", AppName)
	fmt.Println()
	fmt.Println("选项:")
	flag.CommandLine.PrintDefaults()
	fmt.Println()
	fmt.Println("配置文件示例:")
	fmt.Println("  config.yaml 包含数据库、GitHub、爬虫等配置")
	fmt.Println()
	fmt.Println("数据库命令:")
	fmt.Printf("  %s --db-only                    # 仅初始化数据库\n", AppName)
	fmt.Printf("  %s --stats                      # 显示数据库统计\n", AppName)
	fmt.Printf("  %s --health                     # 执行健康检查\n", AppName)
	fmt.Printf("  %s --backup                     # 执行数据库备份\n", AppName)
	fmt.Printf("  %s --restore backup.sql         # 从备份恢复\n", AppName)
	fmt.Println()
	fmt.Println("数据导出命令:")
	fmt.Printf("  %s --export --output data.json  # 导出JSON格式数据\n", AppName)
	fmt.Printf("  %s --export --output data.csv --export-format csv  # 导出CSV格式\n", AppName)
	fmt.Printf("  %s --report --output report.html  # 生成HTML报告\n", AppName)
	fmt.Printf("  %s --export --output data.json --report  # 导出并生成报告\n", AppName)
}

// loadConfig 加载配置文件
func loadConfig(configFile string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	// 解析配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	// 验证必需的配置
	if config.GithubToken == "" {
		return nil, fmt.Errorf("GitHub token 不能为空")
	}
	
	if len(config.Repos) == 0 {
		return nil, fmt.Errorf("至少需要配置一个仓库")
	}
	
	if len(config.Keywords) == 0 {
		return nil, fmt.Errorf("至少需要配置一个关键词")
	}
	
	return &config, nil
}

// setupLogger 设置日志
func setupLogger(level string, debug bool) {
	var logFlags log.LstdFlags | log.Lmsgprefix
	
	if debug {
		logFlags |= log.Lshortfile
	}
	
	logger = log.New(os.Stdout, "[Main] ", logFlags)
	
	// 设置日志级别
	if level == "debug" && debug {
		logger.Println("调试模式已启用")
	}
}

// createDirectories 创建必要的目录
func createDirectories(config *Config) error {
	dirs := []string{
		config.App.DataDir,
		config.App.OutputDir,
	}
	
	// 添加数据库备份目录
	if config.Database.BackupEnabled && config.Database.BackupPath != "" {
		dirs = append(dirs, config.Database.BackupPath)
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}
	
	return nil
}

// initializeDatabase 初始化数据库
func initializeDatabase(config *Config) (*DatabaseManager, error) {
	logger.Println("正在初始化数据库...")
	
	// 创建数据库管理器
	dbManager, err := NewDatabaseManager(&config.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("创建数据库管理器失败: %w", err)
	}
	
	// 启动数据库管理器
	if err := dbManager.Start(appCtx); err != nil {
		return nil, fmt.Errorf("启动数据库管理器失败: %w", err)
	}
	
	logger.Println("数据库初始化完成")
	return dbManager, nil
}

// performHealthCheck 执行健康检查
func performHealthCheck(dbManager *DatabaseManager) error {
	logger.Println("执行数据库健康检查...")
	
	ctx, cancel := context.WithTimeout(appCtx, 30*time.Second)
	defer cancel()
	
	if err := dbManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("数据库健康检查失败: %w", err)
	}
	
	logger.Println("数据库健康检查通过")
	return nil
}

// executeBackup 执行数据库备份
func executeBackup(dbManager *DatabaseManager, config *Config) error {
	logger.Println("执行数据库备份...")
	
	if config.Database.BackupPath == "" {
		return fmt.Errorf("备份路径未配置")
	}
	
	ctx, cancel := context.WithTimeout(appCtx, 5*time.Minute)
	defer cancel()
	
	// 创建备份
	backupPath, err := dbManager.Backup(ctx, config.Database.BackupPath)
	if err != nil {
		return fmt.Errorf("数据库备份失败: %w", err)
	}
	
	logger.Printf("数据库备份完成: %s", backupPath)
	return nil
}

// executeRestore 执行数据库恢复
func executeRestore(dbManager *DatabaseManager, backupFile string) error {
	logger.Printf("从备份文件恢复数据库: %s", backupFile)
	
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", backupFile)
	}
	
	ctx, cancel := context.WithTimeout(appCtx, 5*time.Minute)
	defer cancel()
	
	if err := dbManager.Restore(ctx, backupFile); err != nil {
		return fmt.Errorf("数据库恢复失败: %w", err)
	}
	
	logger.Println("数据库恢复完成")
	return nil
}

// runScraper 运行爬虫
func runScraper(config *Config, dbManager *DatabaseManager) error {
	logger.Println("启动爬虫...")
	
	// 创建GitHub客户端
	client := scraper.NewGithubClient(config.GithubToken)
	
	var totalIssues int
	var results []scraper.PitfallIssue
	
	// 爬取每个仓库
	for i, repo := range config.Repos {
		select {
		case <-appCtx.Done():
			return appCtx.Err()
		default:
		}
		
		progress := fmt.Sprintf("📦 进度: %d/%d", i+1, len(config.Repos))
		logger.Printf("%s - 正在爬取: %s/%s", progress, repo.Owner, repo.Name)
		
		issues, err := scraper.ScrapeRepo(
			client,
			repo.Owner,
			repo.Name,
			config.Keywords,
		)
		
		if err != nil {
			logger.Printf("爬取 %s/%s 失败: %v", repo.Owner, repo.Name, err)
			continue
		}
		
		// 保存到数据库
		if dbManager != nil {
			ctx, cancel := context.WithTimeout(appCtx, 30*time.Second)
			for _, issue := range issues {
				if err := saveIssueToDatabase(ctx, dbManager, issue); err != nil {
					logger.Printf("保存Issue到数据库失败: %v", err)
				}
			}
			cancel()
		}
		
		results = append(results, issues...)
		totalIssues += len(issues)
		
		// 请求间隔
		if i < len(config.Repos)-1 && config.RequestInterval > 0 {
			time.Sleep(time.Duration(config.RequestInterval) * time.Millisecond)
		}
	}
	
	logger.Printf("爬取完成，共获取 %d 个Issue", totalIssues)
	
	// 保存结果到文件
	return saveResultsToFile(config, results)
}

// saveIssueToDatabase 保存Issue到数据库
func saveIssueToDatabase(ctx context.Context, dbManager *DatabaseManager, issue scraper.PitfallIssue) error {
	query := `
		INSERT INTO issues (
			repo_owner, repo_name, issue_number, title, body, labels,
			state, created_at, updated_at, html_url, score, severity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (repo_owner, repo_name, issue_number) 
		DO UPDATE SET 
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			labels = EXCLUDED.labels,
			state = EXCLUDED.state,
			updated_at = EXCLUDED.updated_at,
			score = EXCLUDED.score,
			severity = EXCLUDED.severity
	`
	
	_, err := dbManager.ExecuteExec(ctx, query,
		issue.RepoOwner, issue.RepoName, issue.Number, issue.Title,
		issue.Body, strings.Join(issue.Labels, ","), issue.State,
		issue.CreatedAt, issue.UpdatedAt, issue.HTMLURL, issue.Score, issue.Severity,
	)
	
	return err
}

// saveResultsToFile 保存结果到文件
func saveResultsToFile(config *Config, results []scraper.PitfallIssue) error {
	outputPath := filepath.Join(config.App.OutputDir, "issues.json")
	
	// 确保输出目录存在
	if err := os.MkdirAll(config.App.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}
	
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	
	logger.Printf("结果已保存到: %s", outputPath)
	return nil
}

// setupSignalHandlers 设置信号处理器
func setupSignalHandlers(dbManager *DatabaseManager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		logger.Println("接收到退出信号，正在清理资源...")
		
		if cancel != nil {
			cancel()
		}
		
		if dbManager != nil {
			if err := dbManager.Stop(); err != nil {
				logger.Printf("停止数据库管理器失败: %v", err)
			}
		}
		
		os.Exit(0)
	}()
}

// main 主函数
func main() {
	// 解析命令行参数
	appConfig := parseCommandLine()
	
	// 显示版本信息
	if appConfig.ShowVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		fmt.Printf("Go版本: %s\n", runtime.Version())
		fmt.Printf("作者: %s\n", AppAuthor)
		return
	}
	
	// 显示帮助信息
	if appConfig.ShowHelp {
		showHelp()
		return
	}
	
	// 设置日志
	setupLogger("info", appConfig.Debug)
	
	// 创建应用上下文
	appCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	
	// 加载配置
	config, err := loadConfig(appConfig.ConfigFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 创建必要目录
	if err := createDirectories(config); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}
	
	// 设置信号处理器
	var dbManager *DatabaseManager
	setupSignalHandlers(dbManager)
	
	// 初始化数据库
	if !appConfig.DatabaseOnly {
		dbManager, err = initializeDatabase(config)
		if err != nil {
			log.Fatalf("初始化数据库失败: %v", err)
		}
		defer dbManager.Stop()
	}
	
	// 执行数据库操作
	switch {
	case appConfig.HealthCheck:
		if err := performHealthCheck(dbManager); err != nil {
			log.Fatalf("健康检查失败: %v", err)
		}
		return
		
	case appConfig.Stats:
		dbManager.ShowStats()
		return
		
	case appConfig.Backup:
		if err := executeBackup(dbManager, config); err != nil {
			log.Fatalf("备份失败: %v", err)
		}
		return
		
	case appConfig.Restore != "":
		if err := executeRestore(dbManager, appConfig.Restore); err != nil {
			log.Fatalf("恢复失败: %v", err)
		}
		return
		
	case appConfig.Export:
		if err := executeExport(dbManager, appConfig); err != nil {
			log.Fatalf("导出失败: %v", err)
		}
		return
		
	case appConfig.Report:
		if err := executeReport(dbManager, appConfig); err != nil {
			log.Fatalf("报告生成失败: %v", err)
		}
		return
		
	case appConfig.DatabaseOnly:
		logger.Println("数据库初始化完成（--db-only模式）")
		return
	}
	
	// 运行爬虫
	if err := runScraper(config, dbManager); err != nil {
		log.Fatalf("爬虫运行失败: %v", err)
	}
	
	logger.Println("程序执行完成")
}

// executeExport 执行数据导出
func executeExport(dbManager *DatabaseManager, config *AppConfig) error {
	logger.Printf("开始导出数据...")
	
	// 获取数据库连接
	db := dbManager.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接不可用")
	}
	
	// 创建导出器
	exporter := database.NewExporter(db)
	
	// 构建导出过滤器
	filter := database.ExportFilter{
		IncludeMetadata: true,
	}
	
	// 确定导出格式
	var format database.ExportFormat
	switch config.ExportFormat {
	case "json":
		format = database.FormatJSON
	case "csv":
		format = database.FormatCSV
	case "md":
		format = database.FormatMarkdown
	default:
		return fmt.Errorf("不支持的导出格式: %s", config.ExportFormat)
	}
	
	// 执行导出
	result, err := exporter.ExportIssues(filter, format, config.ExportOutput)
	if err != nil {
		return err
	}
	
	logger.Printf("导出完成:")
	logger.Printf("  输出文件: %s", result.OutputPath)
	logger.Printf("  格式: %s", result.Format)
	logger.Printf("  总记录数: %d", result.TotalRecords)
	logger.Printf("  导出记录数: %d", result.ExportedRecords)
	logger.Printf("  耗时: %v", result.Duration)
	
	// 如果同时需要生成报告
	if config.Report {
		return executeReport(dbManager, config)
	}
	
	return nil
}

// executeReport 执行报告生成
func executeReport(dbManager *DatabaseManager, config *AppConfig) error {
	logger.Printf("开始生成报告...")
	
	// 获取数据库连接
	db := dbManager.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接不可用")
	}
	
	// 创建报告生成器
	reportGen := database.NewReportGenerator(db)
	
	// 构建报告配置
	reportConfig := database.ReportConfig{
		Title:       config.ReportTitle,
		Description: "GitHub Issues Analytics Report",
		OutputPath:  config.ExportOutput,
		Format:      config.ReportFormat,
		Parameters: map[string]interface{}{
			"generated_by": "gh-pitfall-scraper",
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
				XAxis:      "category",
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
	result, err := reportGen.GenerateReport(reportConfig)
	if err != nil {
		return err
	}
	
	logger.Printf("报告生成完成:")
	logger.Printf("  输出文件: %s", result.OutputPath)
	logger.Printf("  格式: %s", result.Format)
	logger.Printf("  大小: %d bytes", result.SizeBytes)
	logger.Printf("  耗时: %v", result.Duration)
	
	return nil
}