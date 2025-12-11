package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

// BenchmarkCreateIssue 测试创建Issue的性能
func BenchmarkCreateIssue(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()
	issue := CreateTestIssue()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		issue.Number = i + 10000 // 确保唯一编号
		issue.ContentHash = fmt.Sprintf("benchmark_hash_%d", i)
		_, err := crud.CreateIssue(issue)
		if err != nil {
			b.Fatalf("Failed to create issue: %v", err)
		}
	}
}

// BenchmarkBatchCreateIssues 测试批量创建Issues的性能
func BenchmarkBatchCreateIssues(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()
	batchSizes := []int{10, 50, 100, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			issues := make([]*Issue, batchSize)
			
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// 准备批量数据
				for j := 0; j < batchSize; j++ {
					issue := CreateTestIssue()
					issue.Number = i*batchSize + j + 20000
					issue.ContentHash = fmt.Sprintf("batch_hash_%d_%d", i, j)
					issues[j] = issue
				}

				_, err := crud.CreateIssues(issues)
				if err != nil {
					b.Fatalf("Failed to create batch issues: %v", err)
				}
			}
		})
	}
}

// BenchmarkQueryIssues 测试查询Issues的性能
func BenchmarkQueryIssues(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 先准备大量测试数据
	issues := make([]*Issue, 1000)
	for i := 0; i < 1000; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 30000
		issue.Title = fmt.Sprintf("Performance Test Issue %d", i)
		issue.Score = float64(i % 100) // 0-99的分数
		issue.ContentHash = fmt.Sprintf("perf_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create test data: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := crud.GetAllIssues(100, 0)
		if err != nil {
			b.Fatalf("Failed to query issues: %v", err)
		}
		_ = result // 使用结果避免优化
	}
}

// BenchmarkSearchIssues 测试搜索Issues的性能
func BenchmarkSearchIssues(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 准备包含关键词的测试数据
	keywords := []string{"performance", "memory", "bug", "feature", "security"}
	issues := make([]*Issue, 500)
	for i := 0; i < 500; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 40000
		keyword := keywords[i%len(keywords)]
		issue.Title = fmt.Sprintf("%s issue number %d", keyword, i)
		issue.Body = fmt.Sprintf("This is a %s related issue with detailed description", keyword)
		issue.Keywords = StringArray{keyword, "test"}
		issue.ContentHash = fmt.Sprintf("search_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create search test data: %v", err)
	}

	searchTerms := []string{"performance", "memory", "bug", "feature", "nonexistent"}

	for _, term := range searchTerms {
		b.Run(fmt.Sprintf("Search_%s", term), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result, err := crud.SearchIssues(term, 50, 0)
				if err != nil {
					b.Fatalf("Failed to search issues: %v", err)
				}
				_ = result // 使用结果避免优化
			}
		})
	}
}

// BenchmarkAdvancedSearch 测试高级搜索的性能
func BenchmarkAdvancedSearch(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 准备复杂的测试数据
	issues := make([]*Issue, 200)
	for i := 0; i < 200; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 50000
		issue.Category = []string{"performance", "security", "feature", "bug"}[i%4]
		issue.Priority = []string{"high", "medium", "low", "critical"}[i%4]
		issue.State = []string{"open", "closed"}[i%2]
		issue.Score = float64(i % 50) // 0-49的分数
		issue.TechStack = StringArray{[]string{"go", "python", "javascript", "java"}[i%4]}
		issue.ContentHash = fmt.Sprintf("adv_search_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create advanced search test data: %v", err)
	}

	// 定义不同的搜索场景
	scenarios := []struct {
		name   string
		search *AdvancedSearch
	}{
		{
			name: "Simple_Query",
			search: &AdvancedSearch{
				Query:    "issue",
				Limit:    50,
				Offset:   0,
				SortBy:   "score",
				SortOrder: "DESC",
			},
		},
		{
			name: "Category_Filter",
			search: &AdvancedSearch{
				Categories: []string{"performance", "security"},
				Limit:      50,
				Offset:     0,
				SortBy:     "score",
				SortOrder:  "DESC",
			},
		},
		{
			name: "Score_Range",
			search: &AdvancedSearch{
				MinScore:  func() *float64 { score := 20.0; return &score }(),
				MaxScore:  func() *float64 { score := 40.0; return &score }(),
				Limit:     50,
				Offset:    0,
				SortBy:    "score",
				SortOrder: "DESC",
			},
		},
		{
			name: "Complex_Filter",
			search: &AdvancedSearch{
				Categories:      []string{"performance"},
				Priorities:      []string{"high", "critical"},
				States:          []string{"open"},
				ExcludeDuplicates: true,
				Limit:           50,
				Offset:          0,
				SortBy:          "score",
				SortOrder:       "DESC",
			},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result, err := crud.SearchIssuesAdvanced(scenario.search)
				if err != nil {
					b.Fatalf("Failed to perform advanced search: %v", err)
				}
				_ = result // 使用结果避免优化
			}
		})
	}
}

// BenchmarkConcurrentReads 测试并发读取性能
func BenchmarkConcurrentReads(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 准备测试数据
	issues := make([]*Issue, 100)
	for i := 0; i < 100; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 60000
		issue.ContentHash = fmt.Sprintf("concurrent_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create concurrent test data: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 模拟并发读取
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := crud.GetAllIssues(10, 0)
			if err != nil {
				b.Fatalf("Failed to query issues: %v", err)
			}
			_ = result
		}
	})
}

// BenchmarkConcurrentWrites 测试并发写入性能
func BenchmarkConcurrentWrites(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	b.ResetTimer()
	b.ReportAllocs()

	// 模拟并发写入
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			issue := CreateTestIssue()
			issue.Number = 70000 + counter
			issue.ContentHash = fmt.Sprintf("concurrent_write_%d", counter)
			counter++

			_, err := crud.CreateIssue(issue)
			if err != nil {
				b.Fatalf("Failed to create issue: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentMixed 测试并发混合操作性能
func BenchmarkConcurrentMixed(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 准备基础数据
	baseIssues := make([]*Issue, 50)
	for i := 0; i < 50; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 80000
		issue.ContentHash = fmt.Sprintf("mixed_hash_%d", i)
		baseIssues[i] = issue
	}

	_, err = crud.CreateIssues(baseIssues)
	if err != nil {
		b.Fatalf("Failed to create base data: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 混合读写操作
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			operation := counter % 3
			counter++

			switch operation {
			case 0: // 读取
				result, err := crud.GetAllIssues(5, 0)
				if err != nil {
					b.Fatalf("Failed to read issues: %v", err)
				}
				_ = result
			case 1: // 创建
				issue := CreateTestIssue()
				issue.Number = 90000 + counter
				issue.ContentHash = fmt.Sprintf("mixed_create_%d", counter)
				_, err := crud.CreateIssue(issue)
				if err != nil {
					b.Fatalf("Failed to create issue: %v", err)
				}
			case 2: // 搜索
				result, err := crud.SearchIssues("test", 5, 0)
				if err != nil {
					b.Fatalf("Failed to search issues: %v", err)
				}
				_ = result
			}
		}
	})
}

// BenchmarkDeduplicationPerformance 测试去重性能
func BenchmarkDeduplicationPerformance(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()
	deduplicator := tdb.Deduplication()

	// 准备包含重复数据的测试集
	dataSizes := []int{100, 500, 1000}

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("DataSize_%d", size), func(b *testing.B) {
			// 创建重复数据
			issues := make([]*Issue, size)
			for i := 0; i < size; i++ {
				issue := CreateTestIssue()
				issue.Number = 100000 + i
				// 每10个issue共享相同的标题和内容hash，创建重复
				issue.Title = fmt.Sprintf("Duplicate Issue %d", i%10)
				issue.Body = fmt.Sprintf("Duplicate body content %d", i%10)
				issue.ContentHash = fmt.Sprintf("duplicate_hash_%d", i%10)
				issues[i] = issue
			}

			_, err := crud.CreateIssues(issues)
			if err != nil {
				b.Fatalf("Failed to create deduplication test data: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result, err := deduplicator.FindDuplicates()
				if err != nil {
					b.Fatalf("Deduplication failed: %v", err)
				}
				_ = result // 使用结果避免优化
			}
		})
	}
}

// BenchmarkClassificationPerformance 测试分类性能
func BenchmarkClassificationPerformance(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()
	classifier := tdb.Classification()

	// 准备不同类型的测试数据
	testCases := []struct {
		name  string
		title string
		body  string
	}{
		{"Performance", "Memory leak causing performance degradation", "There is a memory leak that affects system performance"},
		{"Security", "Security vulnerability in authentication", "Authentication system has a security vulnerability"},
		{"Feature", "Add new feature for user management", "We need to implement user management features"},
		{"Bug", "Bug in data processing module", "The data processing module is not working correctly"},
		{"Documentation", "Improve API documentation", "We need to improve the API documentation"},
	}

	// 创建测试issues
	issues := make([]*Issue, 100)
	for i := 0; i < 100; i++ {
		testCase := testCases[i%len(testCases)]
		issue := CreateTestIssue()
		issue.Number = 110000 + i
		issue.Title = testCase.title
		issue.Body = testCase.body
		issue.ContentHash = fmt.Sprintf("classification_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create classification test data: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 随机选择一个issue进行分类
		issue := issues[i%len(issues)]
		result, err := classifier.ClassifySingleIssue(issue)
		if err != nil {
			b.Fatalf("Classification failed: %v", err)
		}
		_ = result // 使用结果避免优化
	}
}

// BenchmarkStatisticsPerformance 测试统计性能
func BenchmarkStatisticsPerformance(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 准备大量统计测试数据
	issues := make([]*Issue, 500)
	categories := []string{"performance", "security", "feature", "bug", "documentation"}
	priorities := []string{"high", "medium", "low", "critical"}
	states := []string{"open", "closed"}
	techStacks := []string{"go", "python", "javascript", "java", "rust"}

	for i := 0; i < 500; i++ {
		issue := CreateTestIssue()
		issue.Number = 120000 + i
		issue.Category = categories[i%len(categories)]
		issue.Priority = priorities[i%len(priorities)]
		issue.State = states[i%len(states)]
		issue.Score = float64(i % 100) // 0-99的分数
		issue.TechStack = StringArray{techStacks[i%len(techStacks)]}
		issue.ContentHash = fmt.Sprintf("stats_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create statistics test data: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		stats, err := crud.GetIssueStats()
		if err != nil {
			b.Fatalf("Failed to get issue stats: %v", err)
		}
		_ = stats // 使用结果避免优化
	}
}

// BenchmarkMemoryUsage 测试内存使用
func BenchmarkMemoryUsage(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	// 准备大量数据
	issues := make([]*Issue, 1000)
	for i := 0; i < 1000; i++ {
		issue := CreateTestIssue()
		issue.Number = 130000 + i
		issue.Title = fmt.Sprintf("Memory Usage Test Issue %d with a very long title to test memory consumption and data handling", i)
		issue.Body = fmt.Sprintf("This is a detailed body content for memory usage testing. It contains multiple lines of text to simulate real-world scenarios where issues might have extensive descriptions, error messages, stack traces, and other relevant information that would impact memory usage in the application. Issue number: %d", i)
		issue.Keywords = StringArray{"memory", "test", "performance", "benchmark", "usage", "consumption"}
		issue.TechStack = StringArray{"go", "database", "sqlite", "memory", "optimization"}
		issue.ContentHash = fmt.Sprintf("memory_hash_%d", i)
		issues[i] = issue
	}

	_, err = crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create memory usage test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 大量查询操作，测试内存使用
		result, err := crud.GetAllIssues(1000, 0)
		if err != nil {
			b.Fatalf("Failed to query issues: %v", err)
		}
		
		// 处理结果以确保内存分配
		totalScore := 0.0
		for _, issue := range result {
			totalScore += issue.Score
		}
		_ = totalScore
	}
}

// BenchmarkLargeDataset 大数据集性能测试
func BenchmarkLargeDataset(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	datasetSizes := []int{1000, 5000, 10000}

	for _, size := range datasetSizes {
		b.Run(fmt.Sprintf("DatasetSize_%d", size), func(b *testing.B) {
			log.Printf("Creating dataset with %d issues...", size)
			
			// 分批创建大量数据
			batchSize := 100
			for i := 0; i < size; i += batchSize {
				issues := make([]*Issue, batchSize)
				for j := 0; j < batchSize && i+j < size; j++ {
					issue := CreateTestIssue()
					issue.Number = 140000 + i + j
					issue.Title = fmt.Sprintf("Large Dataset Issue %d", i+j)
					issue.ContentHash = fmt.Sprintf("large_hash_%d", i+j)
					issues[j] = issue
				}
				
				_, err := crud.CreateIssues(issues[:batchSize])
				if err != nil {
					b.Fatalf("Failed to create large dataset batch: %v", err)
				}
			}

			log.Printf("Dataset created, starting benchmarks...")

			// 测试各种操作在大数据集上的性能
			b.Run("Query_All", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := crud.GetAllIssues(100, 0)
					if err != nil {
						b.Fatalf("Failed to query large dataset: %v", err)
					}
					_ = result
				}
			})

			b.Run("Search", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := crud.SearchIssues("Large Dataset", 50, 0)
					if err != nil {
						b.Fatalf("Failed to search large dataset: %v", err)
					}
					_ = result
				}
			})

			b.Run("Advanced_Search", func(b *testing.B) {
				search := &AdvancedSearch{
					Limit:    50,
					Offset:   0,
					SortBy:   "score",
					SortOrder: "DESC",
				}
				
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := crud.SearchIssuesAdvanced(search)
					if err != nil {
						b.Fatalf("Failed to advanced search large dataset: %v", err)
					}
					_ = result
				}
			})

			// 清理数据
			_, err := tdb.GetDB().Exec("DELETE FROM issues WHERE number >= 140000")
			if err != nil {
				b.Logf("Failed to cleanup large dataset: %v", err)
			}
		})
	}
}

// BenchmarkConcurrentTransaction 并发事务性能测试
func BenchmarkConcurrentTransaction(b *testing.B) {
	tdb, err := TestSetup()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		b.Fatalf("Database initialization failed: %v", err)
	}

	transaction := tdb.Transaction()

	b.ResetTimer()
	b.ReportAllocs()

	// 并发事务测试
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
				issueNumber := 150000 + counter
				counter++

				// 模拟复杂的事务操作
				_, err := tx.Exec(`
					INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, issueNumber, fmt.Sprintf("Transaction Issue %d", issueNumber), 
					"Transaction test body", "https://github.com/test/repo/issues/"+fmt.Sprintf("%d", issueNumber),
					"open", "test", "repo", time.Now(), time.Now())
				
				if err != nil {
					return err
				}

				// 模拟更多的数据库操作
				_, err = tx.Exec("UPDATE issues SET score = score + 1 WHERE number = ?", issueNumber-1)
				return err
			})
			
			if err != nil {
				b.Fatalf("Transaction failed: %v", err)
			}
		}
	})
}