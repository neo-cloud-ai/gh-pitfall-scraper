package main

import (
	"fmt"
	"log"
	"time"

	"github.com/your-repo/gh-pitfall-scraper/internal/database"
)

func main() {
	// 1. 配置数据库
	config := database.DefaultDatabaseConfig()
	config.Path = "./data/issues.db"
	config.MaxConnections = 10
	
	// 2. 创建数据库实例
	db, err := database.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// 3. 初始化数据库
	fmt.Println("Initializing database...")
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// 4. 基本 CRUD 操作示例
	basicCRUDExample(db)
	
	// 5. 批量操作示例
	batchOperationsExample(db)
	
	// 6. 搜索和过滤示例
	searchExample(db)
	
	// 7. 去重功能示例
	deduplicationExample(db)
	
	// 8. 分类功能示例
	classificationExample(db)
	
	// 9. 事务操作示例
	transactionExample(db)
	
	// 10. 统计信息示例
	statsExample(db)
	
	fmt.Println("All examples completed successfully!")
}

func basicCRUDExample(db *database.Database) {
	fmt.Println("\n=== Basic CRUD Operations ===")
	
	crud := db.CRUD()
	
	// 创建问题
	issue := &database.Issue{
		Number:      1,
		Title:       "Memory leak in cache implementation",
		Body:        "There is a memory leak in the cache implementation that causes memory usage to grow over time. This affects performance significantly.",
		URL:         "https://github.com/example/repo/issues/1",
		State:       "open",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		Comments:    5,
		Reactions:   12,
		Assignee:    "developer1",
		Milestone:   "v1.2",
		RepoOwner:   "example",
		RepoName:    "repo",
		Keywords:    database.StringArray{"memory", "leak", "cache", "performance"},
		Score:       25.0,
		Category:    "performance",
		Priority:    "high",
		TechStack:   database.StringArray{"go", "cache", "memory"},
		Labels:      database.StringArray{"bug", "performance"},
		IsDuplicate: false,
	}
	
	// 插入问题
	id, err := crud.CreateIssue(issue)
	if err != nil {
		log.Printf("Failed to create issue: %v", err)
		return
	}
	fmt.Printf("Created issue with ID: %d\n", id)
	
	// 获取问题
	retrievedIssue, err := crud.GetIssue(id)
	if err != nil {
		log.Printf("Failed to get issue: %v", err)
		return
	}
	fmt.Printf("Retrieved issue: %s (Score: %.1f, Category: %s)\n", 
		retrievedIssue.Title, retrievedIssue.Score, retrievedIssue.Category)
	
	// 更新问题
	retrievedIssue.Score = 27.5
	retrievedIssue.Priority = "critical"
	if err := crud.UpdateIssue(retrievedIssue); err != nil {
		log.Printf("Failed to update issue: %v", err)
		return
	}
	fmt.Printf("Updated issue score to: %.1f\n", retrievedIssue.Score)
	
	// 检查问题是否存在
	exists, err := crud.Exists(id)
	if err != nil {
		log.Printf("Failed to check existence: %v", err)
		return
	}
	fmt.Printf("Issue exists: %v\n", exists)
	
	// 删除问题
	if err := crud.DeleteIssue(id); err != nil {
		log.Printf("Failed to delete issue: %v", err)
		return
	}
	fmt.Println("Deleted issue")
}

func batchOperationsExample(db *database.Database) {
	fmt.Println("\n=== Batch Operations ===")
	
	crud := db.CRUD()
	
	// 创建多个问题
	issues := make([]*database.Issue, 3)
	for i := 0; i < 3; i++ {
		issue := &database.Issue{
			Number:    100 + i,
			Title:     fmt.Sprintf("Batch Test Issue %d", i),
			Body:      fmt.Sprintf("This is test issue number %d for batch operations", i),
			URL:       fmt.Sprintf("https://github.com/example/repo/issues/%d", 100+i),
			State:     "open",
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt: time.Now(),
			Comments:  i,
			Reactions: i * 2,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"test", "batch"},
			Score:     float64(10 + i*5),
			IsDuplicate: false,
		}
		issues[i] = issue
	}
	
	// 批量创建
	ids, err := crud.CreateIssues(issues)
	if err != nil {
		log.Printf("Failed to batch create issues: %v", err)
		return
	}
	fmt.Printf("Created %d issues with IDs: %v\n", len(ids), ids)
	
	// 批量更新
	for i, issue := range issues {
		issue.Score += 10.0
		issue.Category = "updated"
	}
	
	if err := crud.UpdateIssues(issues); err != nil {
		log.Printf("Failed to batch update issues: %v", err)
		return
	}
	fmt.Println("Updated all issues")
	
	// 批量删除
	if err := crud.DeleteIssues(ids); err != nil {
		log.Printf("Failed to batch delete issues: %v", err)
		return
	}
	fmt.Println("Deleted all issues")
}

func searchExample(db *database.Database) {
	fmt.Println("\n=== Search and Filter Operations ===")
	
	crud := db.CRUD()
	
	// 先创建一些测试数据
	testIssues := []*database.Issue{
		{
			Number:    200,
			Title:     "SQL injection vulnerability",
			Body:      "Found SQL injection vulnerability in user authentication",
			URL:       "https://github.com/example/repo/issues/200",
			State:     "open",
			CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
			UpdatedAt: time.Now(),
			Comments:  8,
			Reactions: 15,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"security", "sql", "injection"},
			Score:     30.0,
			Category:  "security",
			Priority:  "critical",
			TechStack: database.StringArray{"go", "database", "security"},
			IsDuplicate: false,
		},
		{
			Number:    201,
			Title:     "Performance issue with large datasets",
			Body:      "The application becomes slow when processing large datasets",
			URL:       "https://github.com/example/repo/issues/201",
			State:     "open",
			CreatedAt: time.Now().Add(-1 * 24 * time.Hour),
			UpdatedAt: time.Now(),
			Comments:  3,
			Reactions: 7,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"performance", "dataset", "optimization"},
			Score:     22.0,
			Category:  "performance",
			Priority:  "high",
			TechStack: database.StringArray{"go", "performance"},
			IsDuplicate: false,
		},
	}
	
	crud.CreateIssues(testIssues)
	
	// 按仓库查询
	repoIssues, err := crud.GetIssuesByRepository("example", "repo", 10, 0)
	if err != nil {
		log.Printf("Failed to get repository issues: %v", err)
		return
	}
	fmt.Printf("Found %d issues in repository\n", len(repoIssues))
	
	// 按分数范围查询
	scoreIssues, err := crud.GetIssuesByScore(20.0, 35.0, 10, 0)
	if err != nil {
		log.Printf("Failed to get issues by score: %v", err)
		return
	}
	fmt.Printf("Found %d issues with score 20-35\n", len(scoreIssues))
	
	// 按分类查询
	categoryIssues, err := crud.GetIssuesByCategory("security", 10, 0)
	if err != nil {
		log.Printf("Failed to get issues by category: %v", err)
		return
	}
	fmt.Printf("Found %d security issues\n", len(categoryIssues))
	
	// 按关键词查询
	keywordIssues, err := crud.GetIssuesByKeywords([]string{"performance", "optimization"}, 10, 0)
	if err != nil {
		log.Printf("Failed to get issues by keywords: %v", err)
		return
	}
	fmt.Printf("Found %d issues with performance keywords\n", len(keywordIssues))
	
	// 文本搜索
	searchIssues, err := crud.SearchIssues("vulnerability", 10, 0)
	if err != nil {
		log.Printf("Failed to search issues: %v", err)
		return
	}
	fmt.Printf("Found %d issues containing 'vulnerability'\n", len(searchIssues))
	
	// 高级搜索
	advancedSearch := &database.AdvancedSearch{
		Query:       "security",
		Categories:  []string{"security", "performance"},
		SortBy:      "score",
		SortOrder:   "DESC",
		Limit:       10,
		Offset:      0,
		ExcludeDuplicates: true,
	}
	
	advancedIssues, err := crud.SearchIssuesAdvanced(advancedSearch)
	if err != nil {
		log.Printf("Failed to perform advanced search: %v", err)
		return
	}
	fmt.Printf("Found %d issues with advanced search\n", len(advancedIssues))
}

func deduplicationExample(db *database.Database) {
	fmt.Println("\n=== Deduplication Operations ===")
	
	deduplicator := db.Deduplication()
	
	// 创建一些相似的问题用于去重测试
	similarIssues := make([]*database.Issue, 3)
	for i := 0; i < 3; i++ {
		issue := &database.Issue{
			Number:    300 + i,
			Title:     "Memory leak in cache implementation", // 相同标题
			Body:      "There is a memory leak that causes high memory usage", // 相似内容
			URL:       fmt.Sprintf("https://github.com/example/repo/issues/%d", 300+i),
			State:     "open",
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt: time.Now(),
			Comments:  i,
			Reactions: i * 2,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"memory", "leak", "cache"},
			Score:     25.0,
			ContentHash: "similar_content_hash", // 相同内容哈希
			IsDuplicate: false,
		}
		similarIssues[i] = issue
	}
	
	db.CRUD().CreateIssues(similarIssues)
	
	// 执行去重
	result, err := deduplicator.FindDuplicates()
	if err != nil {
		log.Printf("Failed to perform deduplication: %v", err)
		return
	}
	
	fmt.Printf("Deduplication Results:\n")
	fmt.Printf("  Total Processed: %d\n", result.TotalProcessed)
	fmt.Printf("  Duplicates Found: %d\n", result.DuplicatesFound)
	fmt.Printf("  Unique Issues: %d\n", result.UniqueIssues)
	fmt.Printf("  Duplicate Groups: %d\n", len(result.DuplicateGroups))
	
	for i, group := range result.DuplicateGroups {
		fmt.Printf("  Group %d: Master=%s, Duplicates=%d, Similarity=%.2f\n",
			i+1, group.MasterIssue.Title, len(group.Duplicates), group.Similarity)
	}
	
	// 获取去重统计
	stats, err := deduplicator.GetDuplicateStats()
	if err != nil {
		log.Printf("Failed to get duplicate stats: %v", err)
		return
	}
	
	fmt.Printf("Duplicate Statistics:\n")
	fmt.Printf("  Total Duplicates: %d\n", stats["total_duplicates"])
	fmt.Printf("  Unique Issues: %d\n", stats["unique_issues"])
	if duplicateRate, ok := stats["duplicate_rate"]; ok {
		fmt.Printf("  Duplicate Rate: %.2f%%\n", duplicateRate.(float64)*100)
	}
	
	// 移除重复项
	removed, err := deduplicator.RemoveDuplicates()
	if err != nil {
		log.Printf("Failed to remove duplicates: %v", err)
		return
	}
	fmt.Printf("Removed %d duplicate issues\n", removed)
}

func classificationExample(db *database.Database) {
	fmt.Println("\n=== Classification Operations ===")
	
	classifier := db.Classification()
	
	// 创建不同类型的问题
	issues := []*database.Issue{
		{
			Number:    400,
			Title:     "Bug in authentication system",
			Body:      "The authentication system has a bug that prevents users from logging in",
			URL:       "https://github.com/example/repo/issues/400",
			State:     "open",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now(),
			Comments:  5,
			Reactions: 8,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"bug", "authentication", "login"},
			Score:     20.0,
			IsDuplicate: false,
		},
		{
			Number:    401,
			Title:     "Performance optimization needed",
			Body:      "The application is slow and needs performance optimization",
			URL:       "https://github.com/example/repo/issues/401",
			State:     "open",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now(),
			Comments:  3,
			Reactions: 6,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"performance", "optimization", "speed"},
			Score:     18.0,
			IsDuplicate: false,
		},
		{
			Number:    402,
			Title:     "Security vulnerability in payment processing",
			Body:      "There is a security vulnerability in the payment processing module",
			URL:       "https://github.com/example/repo/issues/402",
			State:     "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Comments:  12,
			Reactions: 20,
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"security", "vulnerability", "payment"},
			Score:     35.0,
			IsDuplicate: false,
		},
	}
	
	db.CRUD().CreateIssues(issues)
	
	// 单个分类
	result, err := classifier.ClassifySingleIssue(issues[0])
	if err != nil {
		log.Printf("Failed to classify single issue: %v", err)
		return
	}
	
	fmt.Printf("Single Classification Result:\n")
	fmt.Printf("  Issue: %s\n", result.Issue.Title)
	fmt.Printf("  Category: %s (Confidence: %.2f)\n", result.PredictedCategory, result.Confidence)
	fmt.Printf("  Priority: %s\n", result.PredictedPriority)
	fmt.Printf("  Tech Stack: %v\n", result.PredictedTechStack)
	
	// 批量分类
	stats, err := classifier.ClassifyIssues(issues)
	if err != nil {
		log.Printf("Failed to classify issues: %v", err)
		return
	}
	
	fmt.Printf("\nBatch Classification Results:\n")
	fmt.Printf("  Total Processed: %d\n", stats.TotalProcessed)
	fmt.Printf("  Auto Classified: %d\n", stats.AutoClassified)
	fmt.Printf("  Average Confidence: %.2f\n", stats.Confidence)
	
	fmt.Printf("  Categories: %v\n", stats.ByCategory)
	fmt.Printf("  Priorities: %v\n", stats.ByPriority)
	fmt.Printf("  Tech Stacks: %v\n", stats.ByTechStack)
	
	// 获取分类统计
	clsStats, err := classifier.GetClassificationStats()
	if err != nil {
		log.Printf("Failed to get classification stats: %v", err)
		return
	}
	
	fmt.Printf("\nClassification Statistics:\n")
	fmt.Printf("  Categories: %v\n", clsStats["categories"])
	fmt.Printf("  Priorities: %v\n", clsStats["priorities"])
	fmt.Printf("  Tech Stacks: %v\n", clsStats["tech_stacks"])
}

func transactionExample(db *database.Database) {
	fmt.Println("\n=== Transaction Operations ===")
	
	transaction := db.Transaction()
	
	// 成功事务
	fmt.Println("Testing successful transaction...")
	err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
		// 在事务中执行多个操作
		_, err := tx.Exec("INSERT INTO issues (number, title, repo_owner, repo_name) VALUES (?, ?, ?, ?)",
			500, "Transaction Test Issue", "example", "repo")
		if err != nil {
			return err
		}
		
		_, err = tx.Exec("INSERT INTO issues (number, title, repo_owner, repo_name) VALUES (?, ?, ?, ?)",
			501, "Another Transaction Issue", "example", "repo")
		if err != nil {
			return err
		}
		
		return nil
	})
	
	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}
	fmt.Println("Transaction committed successfully")
	
	// 回滚事务
	fmt.Println("Testing rollback transaction...")
	err = transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO issues (number, title, repo_owner, repo_name) VALUES (?, ?, ?, ?)",
			502, "Rollback Test Issue", "example", "repo")
		if err != nil {
			return err
		}
		
		// 模拟错误导致回滚
		return fmt.Errorf("simulated error for rollback test")
	})
	
	if err != nil {
		fmt.Printf("Transaction rolled back as expected: %v\n", err)
	}
	
	// 批量插入事务
	fmt.Println("Testing batch insert transaction...")
	issues := make([]*database.Issue, 3)
	for i := 0; i < 3; i++ {
		issues[i] = &database.Issue{
			Number:    600 + i,
			Title:     fmt.Sprintf("Batch Transaction Issue %d", i),
			Body:      "This is a test issue for batch transaction",
			URL:       fmt.Sprintf("https://github.com/example/repo/issues/%d", 600+i),
			State:     "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			RepoOwner: "example",
			RepoName:  "repo",
			Keywords:  database.StringArray{"test", "transaction"},
			Score:     15.0,
			IsDuplicate: false,
		}
	}
	
	ids, err := transaction.BatchInsertIssues(issues)
	if err != nil {
		log.Printf("Batch insert transaction failed: %v", err)
		return
	}
	fmt.Printf("Batch inserted %d issues with IDs: %v\n", len(ids), ids)
	
	// 获取事务统计
	txStats := transaction.GetTransactionStats()
	fmt.Printf("Transaction Statistics: %v\n", txStats)
}

func statsExample(db *database.Database) {
	fmt.Println("\n=== Statistics and Monitoring ===")
	
	crud := db.CRUD()
	
	// 获取问题统计
	issueStats, err := crud.GetIssueStats()
	if err != nil {
		log.Printf("Failed to get issue stats: %v", err)
		return
	}
	
	fmt.Printf("Issue Statistics:\n")
	fmt.Printf("  Total Count: %d\n", issueStats.TotalCount)
	fmt.Printf("  Average Score: %.2f\n", issueStats.AverageScore)
	fmt.Printf("  Categories: %v\n", issueStats.ByCategory)
	fmt.Printf("  Priorities: %v\n", issueStats.ByPriority)
	fmt.Printf("  States: %v\n", issueStats.ByState)
	fmt.Printf("  Score Distribution: %v\n", issueStats.ScoreDistribution)
	
	// 获取仓库统计
	repoStats, err := crud.GetRepositoryStats("example", "repo")
	if err != nil {
		log.Printf("Failed to get repository stats: %v", err)
		return
	}
	
	fmt.Printf("\nRepository Statistics (example/repo):\n")
	fmt.Printf("  Total Issues: %d\n", repoStats.TotalCount)
	fmt.Printf("  Categories: %v\n", repoStats.ByCategory)
	
	// 获取数据库统计
	dbStats, err := db.GetStats()
	if err != nil {
		log.Printf("Failed to get database stats: %v", err)
		return
	}
	
	fmt.Printf("\nDatabase Statistics:\n")
	fmt.Printf("  Issues Count: %d\n", dbStats["issues_count"])
	fmt.Printf("  Repositories Count: %d\n", dbStats["repositories_count"])
	fmt.Printf("  Database Size: %.2f MB\n", dbStats["database_size_mb"])
	fmt.Printf("  Journal Mode: %v\n", dbStats["journal_mode"])
	fmt.Printf("  Foreign Keys Enabled: %v\n", dbStats["foreign_keys_enabled"])
	
	// 健康检查
	fmt.Println("\nHealth Check:")
	if err := db.HealthCheck(); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("Database health check passed")
	}
	
	// 优化数据库
	fmt.Println("\nOptimizing database...")
	if err := db.Optimize(); err != nil {
		log.Printf("Database optimization failed: %v", err)
	} else {
		fmt.Println("Database optimization completed")
	}
	
	// 获取去重统计
	dedupStats, err := db.Deduplication().GetDuplicateStats()
	if err != nil {
		log.Printf("Failed to get deduplication stats: %v", err)
	} else {
		fmt.Printf("\nDeduplication Statistics: %v\n", dedupStats)
	}
	
	// 获取分类统计
	clsStats, err := db.Classification().GetClassificationStats()
	if err != nil {
		log.Printf("Failed to get classification stats: %v", err)
	} else {
		fmt.Printf("\nClassification Statistics: %v\n", clsStats)
	}
}