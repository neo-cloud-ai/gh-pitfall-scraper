package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/model"
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/output"
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func main() {
	app := &cli.App{
		Name:  "gh-pitfall-scraper",
		Usage: "自动筛选 GitHub Issues 中的高价值踩坑内容",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "config.yaml",
				Usage: "配置文件路径",
			},
			&cli.StringFlag{
				Name:  "token",
				Usage: "GitHub Token (可选)",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "./output",
				Usage: "输出目录",
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "markdown",
				Usage: "输出格式 (markdown/json)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "试运行模式 (不实际抓取数据)",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "详细输出",
			},
		},
		Action: runApp,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApp(c *cli.Context) error {
	configPath := c.String("config")
	token := c.String("token")
	outputDir := c.String("output")
	format := c.String("format")
	dryRun := c.Bool("dry-run")
	verbose := c.Bool("verbose")

	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Load configuration
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with command line flags
	if token != "" {
		config.GitHubToken = token
	}
	if outputDir != "" {
		config.Output.OutputDir = outputDir
	}
	if format != "" {
		config.Output.Format = format
	}

	log.Printf("🚀 启动 gh-pitfall-scraper...")
	log.Printf("📁 配置文件: %s", configPath)
	log.Printf("📤 输出目录: %s", config.Output.OutputDir)
	log.Printf("📄 输出格式: %s", config.Output.Format)

	if dryRun {
		log.Println("🔍 试运行模式 - 将模拟数据")
		return runDryRun(config)
	}

	return runScrape(config)
}

// loadConfig loads configuration from YAML file
func loadConfig(configPath string) (scraper.Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set defaults
	viper.SetDefault("filter.min_score", 20.0)
	viper.SetDefault("filter.required_state", "all")
	viper.SetDefault("filter.max_issues", 50)
	viper.SetDefault("output.format", "markdown")
	viper.SetDefault("output.output_dir", "./output")
	viper.SetDefault("output.sort_by", "score")
	viper.SetDefault("output.include_raw", false)

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		return scraper.Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config scraper.Config
	if err := viper.Unmarshal(&config); err != nil {
		return scraper.Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return scraper.Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validateConfig validates the configuration
func validateConfig(config scraper.Config) error {
	if len(config.Repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	for i, repo := range config.Repositories {
		if repo.Name == "" {
			return fmt.Errorf("repository %d has no name", i)
		}
		if !strings.Contains(repo.Name, "/") {
			return fmt.Errorf("repository name %s must be in format owner/repo", repo.Name)
		}
	}

	if config.Filter.MinScore < 0 || config.Filter.MinScore > 100 {
		return fmt.Errorf("min_score must be between 0 and 100")
	}

	validFormats := []string{"markdown", "json"}
	if !contains(validFormats, config.Output.Format) {
		return fmt.Errorf("output format must be one of: %v", validFormats)
	}

	return nil
}

// runScrape executes the main scraping logic
func runScrape(config scraper.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Create scraper
	scraperInstance := scraper.NewScraper(config)

	// Scrape repositories
	log.Println("🔍 开始抓取仓库数据...")
	allIssues, err := scraperInstance.ScrapeRepositories(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to scrape repositories: %w", err)
	}

	if len(allIssues) == 0 {
		log.Println("⚠️  没有抓取到任何数据")
		return nil
	}

	log.Printf("✅ 抓取完成，共获取 %d 个仓库的数据", len(allIssues))

	// Filter and score issues
	log.Println("🎯 开始过滤和评分...")
	filteredIssues := scraperInstance.FilterAndScoreIssues(allIssues, config)

	// Print statistics
	stats := scraperInstance.GetStatistics(allIssues, filteredIssues)
	printStatistics(stats)

	// Generate output
	log.Println("📝 生成输出文件...")
	formatter := output.NewFormatter()
	if err := formatter.FormatIssues(filteredIssues, config.Output.Format, config.Output.OutputDir); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Create summary report
	if err := createSummaryReport(config, allIssues, filteredIssues); err != nil {
		log.Printf("⚠️  警告: 未能创建摘要报告: %v", err)
	}

	log.Printf("🎉 处理完成！结果保存在: %s", config.Output.OutputDir)
	return nil
}

// runDryRun simulates the scraping process
func runDryRun(config scraper.Config) error {
	log.Println("🔍 生成模拟数据...")

	// Generate sample issues for demonstration
	sampleIssues := generateSampleIssues()
	allIssues := make(map[string][]model.Issue)

	for _, repo := range config.Repositories {
		if repo.Enabled {
			// Generate sample issues for each enabled repository
			repoIssues := make([]model.Issue, 0, len(sampleIssues))
			for i, sampleIssue := range sampleIssues {
				if i < repo.MaxIssues {
					issue := sampleIssue
					issue.Repository = repo.Name
					issue.Number = i + 1
					issue.URL = fmt.Sprintf("https://github.com/%s/issues/%d", repo.Name, i+1)
					repoIssues = append(repoIssues, issue)
				}
			}
			allIssues[repo.Name] = repoIssues
		}
	}

	// Create scraper instance for scoring
	scraperInstance := scraper.NewScraper(config)
	filteredIssues := scraperInstance.FilterAndScoreIssues(allIssues, config)

	// Print statistics
	stats := scraperInstance.GetStatistics(allIssues, filteredIssues)
	printStatistics(stats)

	// Generate output
	log.Println("📝 生成模拟输出文件...")
	formatter := output.NewFormatter()
	if err := formatter.FormatIssues(filteredIssues, config.Output.Format, config.Output.OutputDir); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	log.Printf("🎉 模拟完成！示例结果保存在: %s", config.Output.OutputDir)
	return nil
}

// generateSampleIssues creates sample issues for demonstration
func generateSampleIssues() []model.Issue {
	return []model.Issue{
		{
			ID:          1,
			Title:       "Performance regression in GPU memory usage after v0.4.0",
			Body:        "After upgrading to v0.4.0, we're seeing significant memory usage increase...",
			State:       "open",
			CreatedAt:   time.Now().AddDate(0, 0, -10),
			UpdatedAt:   time.Now().AddDate(0, 0, -2),
			Comments:    15,
			Reactions:   8,
			Labels: []model.Label{
				{Name: "bug", Color: "d73a4a"},
				{Name: "performance", Color: "fbca04"},
			},
		},
		{
			ID:          2,
			Title:       "CUDA kernel crash when using flash attention with large batch sizes",
			Body:        "The application crashes with CUDA error when batch size exceeds 32...",
			State:       "open",
			CreatedAt:   time.Now().AddDate(0, 0, -5),
			UpdatedAt:   time.Now().AddDate(0, 0, -1),
			Comments:    23,
			Reactions:   12,
			Labels: []model.Label{
				{Name: "critical", Color: "d73a4a"},
				{Name: "cuda", Color: "1d76db"},
			},
		},
		{
			ID:          3,
			Title:       "Memory leak in distributed training mode",
			Body:        "Memory usage keeps increasing during multi-node training...",
			State:       "open",
			CreatedAt:   time.Now().AddDate(0, 0, -7),
			UpdatedAt:   time.Now().AddDate(0, 0, -3),
			Comments:    8,
			Reactions:   5,
			Labels: []model.Label{
				{Name: "bug", Color: "d73a4a"},
				{Name: "distributed", Color: "0e8a16"},
			},
		},
	}
}

// printStatistics prints scraping statistics
func printStatistics(stats map[string]interface{}) {
	log.Println("📊 抓取统计:")
	log.Printf("   仓库数量: %v", stats["total_repositories"])
	log.Printf("   总问题数: %v", stats["total_issues"])
	log.Printf("   过滤后: %v", stats["filtered_issues"])
	log.Printf("   过滤率: %.2f%%", stats["overall_filter_rate"])
}

// createSummaryReport creates a summary report
func createSummaryReport(config scraper.Config, allIssues, filteredIssues map[string][]model.Issue) error {
	summaryPath := filepath.Join(config.Output.OutputDir, "scraping_summary.txt")
	
	file, err := os.Create(summaryPath)
	if err != nil {
		return err
	}
	defer file.Close()

	summary := fmt.Sprintf(`GitHub Issues 踩坑内容抓取报告
=====================================

抓取时间: %s
配置文件: %s

仓库统计:
`, time.Now().Format("2006-01-02 15:04:05"), "config.yaml")

	for repoName, issues := range allIssues {
		filtered := filteredIssues[repoName]
		summary += fmt.Sprintf("- %s: %d/%d 问题 (过滤率: %.1f%%)\n", 
			repoName, len(filtered), len(issues), 
			float64(len(filtered))/float64(len(issues))*100)
	}

	summary += fmt.Sprintf(`

总计: %d 个仓库, %d 个问题, %d 个高价值问题
过滤率: %.2f%%

工具: gh-pitfall-scraper
`, len(allIssues), getTotalIssues(allIssues), getTotalIssues(filteredIssues),
		float64(getTotalIssues(filteredIssues))/float64(getTotalIssues(allIssues))*100)

	_, err = file.WriteString(summary)
	return err
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getTotalIssues(issues map[string][]model.Issue) int {
	total := 0
	for _, repoIssues := range issues {
		total += len(repoIssues)
	}
	return total
}