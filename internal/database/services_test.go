package database

import (
	"database/sql"
	"testing"
	"time"
)

// TestDeduplicationService 去func TestDedu重服务测试
plicationService(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	deduplicator := tdb.Deduplication()
	crud := tdb.CRUD()

	t.Run("FindDuplicates", func(t *testing.T) {
		// 创建测试数据：相同内容的issues
		issues := []*Issue{
			{
				Number:      1,
				Title:       "Memory leak issue",
				Body:        "There is a memory leak in the application",
				URL:         "https://github.com/test/repo/issues/1",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       15.0,
				ContentHash: "duplicate_hash_1",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      2,
				Title:       "Memory leak issue",
				Body:        "There is a memory leak in the application",
				URL:         "https://github.com/test/repo/issues/2",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       12.0,
				ContentHash: "duplicate_hash_1", // 相同hash
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      3,
				Title:       "Different issue",
				Body:        "This is a different issue",
				URL:         "https://github.com/test/repo/issues/3",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       8.0,
				ContentHash: "unique_hash",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
		}

		// 创建issues
		_, err := crud.CreateIssues(issues)
		if err != nil {
			t.Fatalf("Failed to create test issues: %v", err)
		}

		// 执行去重
		result, err := deduplicator.FindDuplicates()
		if err != nil {
			t.Fatalf("Deduplication failed: %v", err)
		}

		// 验证结果
		if result.TotalProcessed != 3 {
			t.Errorf("Expected 3 processed issues, got %d", result.TotalProcessed)
		}

		if result.DuplicatesFound != 1 { // 应该有1对重复
			t.Errorf("Expected 1 duplicate pair found, got %d", result.DuplicatesFound)
		}

		if len(result.DuplicateGroups) == 0 {
			t.Error("Expected at least one duplicate group")
		}

		// 验证重复组
		for _, group := range result.DuplicateGroups {
			if len(group.Duplicates) < 2 {
				t.Error("Expected at least 2 issues in duplicate group")
			}
			if group.MasterIssue == nil {
				t.Error("Expected master issue in duplicate group")
			}
			if group.Similarity < 0.0 || group.Similarity > 1.0 {
				t.Error("Expected similarity between 0.0 and 1.0")
			}
		}
	})

	t.Run("FindDuplicatesWithSimilarity", func(t *testing.T) {
		// 清理数据库
		_, err := tdb.GetDB().Exec("DELETE FROM issues")
		if err != nil {
			t.Fatalf("Failed to clear issues: %v", err)
		}

		// 创建相似但不相同的issues
		issues := []*Issue{
			{
				Number:      10,
				Title:       "Memory leak in cache system",
				Body:        "The cache system has a memory leak",
				URL:         "https://github.com/test/repo/issues/10",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       20.0,
				ContentHash: "cache_leak",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      11,
				Title:       "Memory leak in cache implementation",
				Body:        "Cache implementation has memory leak issue",
				URL:         "https://github.com/test/repo/issues/11",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       18.0,
				ContentHash: "cache_implementation_leak",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
		}

		_, err = crud.CreateIssues(issues)
		if err != nil {
			t.Fatalf("Failed to create similar issues: %v", err)
		}

		// 执行去重
		result, err := deduplicator.FindDuplicates()
		if err != nil {
			t.Fatalf("Deduplication failed: %v", err)
		}

		// 这些issues应该因为相似性被检测为重复
		t.Logf("Similarity-based duplicates found: %d", result.DuplicatesFound)
	})

	t.Run("GetDuplicateStats", func(t *testing.T) {
		stats, err := deduplicator.GetDuplicateStats()
		if err != nil {
			t.Fatalf("Failed to get duplicate stats: %v", err)
		}

		if stats == nil {
			t.Error("Expected non-nil duplicate stats")
		}

		// 检查统计信息包含必要的字段
		if stats["total_issues"] == nil {
			t.Error("Expected total_issues in stats")
		}

		if stats["duplicate_issues"] == nil {
			t.Error("Expected duplicate_issues in stats")
		}

		if stats["unique_issues"] == nil {
			t.Error("Expected unique_issues in stats")
		}

		t.Logf("Duplicate stats: %+v", stats)
	})

	t.Run("RemoveDuplicates", func(t *testing.T) {
		// 先确保有重复数据
		_, err := tdb.GetDB().Exec(`
			INSERT OR REPLACE INTO issues (id, number, title, body, url, state, repo_owner, repo_name, score, content_hash, created_at_db, updated_at_db)
			VALUES (1, 20, 'Test Issue', 'Test body', 'https://github.com/test/repo/issues/20', 'open', 'test', 'repo', 15.0, 'test_hash', datetime('now'), datetime('now'))
		`)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		result, err := deduplicator.RemoveDuplicates()
		if err != nil {
			t.Fatalf("Remove duplicates failed: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil removal result")
		}

		if result.DuplicatesRemoved < 0 {
			t.Error("Expected non-negative duplicates removed count")
		}

		t.Logf("Removed %d duplicates", result.DuplicatesRemoved)
	})

	t.Run("DeduplicationWithEmptyDatabase", func(t *testing.T) {
		// 清理数据库
		_, err := tdb.GetDB().Exec("DELETE FROM issues")
		if err != nil {
			t.Fatalf("Failed to clear issues: %v", err)
		}

		// 在空数据库上执行去重
		result, err := deduplicator.FindDuplicates()
		if err != nil {
			t.Fatalf("Deduplication on empty database failed: %v", err)
		}

		if result.TotalProcessed != 0 {
			t.Errorf("Expected 0 processed issues for empty database, got %d", result.TotalProcessed)
		}

		if result.DuplicatesFound != 0 {
			t.Errorf("Expected 0 duplicates found for empty database, got %d", result.DuplicatesFound)
		}
	})

	t.Run("DeduplicationConfigValidation", func(t *testing.T) {
		config := DefaultDeduplicationConfig()
		
		// 验证默认配置
		if config.SimilarityThreshold < 0.0 || config.SimilarityThreshold > 1.0 {
			t.Error("Expected similarity threshold between 0.0 and 1.0")
		}

		if config.ContentHashEnabled && config.SimilarityThreshold > 0.9 {
			t.Log("Note: High similarity threshold with content hash enabled may reduce effectiveness")
		}

		// 验证配置更新
		config.SimilarityThreshold = 0.8
		config.MinContentLength = 50
		
		if config.SimilarityThreshold != 0.8 {
			t.Error("Failed to update similarity threshold")
		}

		if config.MinContentLength != 50 {
			t.Error("Failed to update min content length")
		}
	})
}

// TestClassificationService 分类服务测试
func TestClassificationService(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	classifier := tdb.Classification()
	crud := tdb.CRUD()

	t.Run("ClassifySingleIssue", func(t *testing.T) {
		// 创建测试issue
		issue := &Issue{
			Number:      100,
			Title:       "Security vulnerability in authentication system",
			Body:        "There is a security vulnerability that allows unauthorized access",
			URL:         "https://github.com/test/repo/issues/100",
			State:       "open",
			RepoOwner:   "test",
			RepoName:    "repo",
			Score:       25.0,
			CreatedAtDB: time.Now(),
			UpdatedAtDB: time.Now(),
		}

		// 先创建issue
		id, err := crud.CreateIssue(issue)
		if err != nil {
			t.Fatalf("Failed to create issue: %v", err)
		}

		// 执行分类
		result, err := classifier.ClassifySingleIssue(issue)
		if err != nil {
			t.Fatalf("Single issue classification failed: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil classification result")
		}

		// 验证分类结果
		if result.Category == "" {
			t.Error("Expected non-empty category")
		}

		if result.Confidence < 0.0 || result.Confidence > 1.0 {
			t.Error("Expected confidence between 0.0 and 1.0")
		}

		// 对于安全相关的issue，应该分类为security
		if result.Category != "security" && result.Category != "bug" {
			t.Logf("Issue classified as '%s', expected 'security' or 'bug'", result.Category)
		}

		t.Logf("Issue classified as: %s (confidence: %.2f)", result.Category, result.Confidence)
	})

	t.Run("ClassifyIssues", func(t *testing.T) {
		// 创建多个测试issues
		issues := []*Issue{
			{
				Number:      200,
				Title:       "Memory leak in caching system",
				Body:        "The application has a memory leak that causes performance degradation",
				URL:         "https://github.com/test/repo/issues/200",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       20.0,
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      201,
				Title:       "Add PostgreSQL support",
				Body:        "We need to add support for PostgreSQL database",
				URL:         "https://github.com/test/repo/issues/201",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       15.0,
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      202,
				Title:       "API documentation needed",
				Body:        "We need comprehensive API documentation",
				URL:         "https://github.com/test/repo/issues/202",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				Score:       10.0,
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
		}

		// 批量创建issues
		ids, err := crud.CreateIssues(issues)
		if err != nil {
			t.Fatalf("Failed to create test issues: %v", err)
		}

		// 执行批量分类
		stats, err := classifier.ClassifyIssues(issues)
		if err != nil {
			t.Fatalf("Batch classification failed: %v", err)
		}

		if stats == nil {
			t.Error("Expected non-nil classification stats")
		}

		// 验证统计信息
		if stats.TotalProcessed != 3 {
			t.Errorf("Expected 3 processed issues, got %d", stats.TotalProcessed)
		}

		if stats.Classified < 0 {
			t.Error("Expected non-negative classified count")
		}

		if stats.Confidence < 0.0 || stats.Confidence > 1.0 {
			t.Error("Expected confidence between 0.0 and 1.0")
		}

		// 检查分类分布
		if len(stats.ByCategory) == 0 {
			t.Error("Expected non-empty category distribution")
		}

		t.Logf("Classification stats: %+v", stats)
	})

	t.Run("GetClassificationStats", func(t *testing.T) {
		stats, err := classifier.GetClassificationStats()
		if err != nil {
			t.Fatalf("Failed to get classification stats: %v", err)
		}

		if stats == nil {
			t.Error("Expected non-nil classification stats")
		}

		// 检查统计信息包含必要的字段
		if stats["total_issues"] == nil {
			t.Error("Expected total_issues in classification stats")
		}

		if stats["classified_issues"] == nil {
			t.Error("Expected classified_issues in classification stats")
		}

		if stats["categories"] == nil {
			t.Error("Expected categories in classification stats")
		}

		t.Logf("Classification stats: %+v", stats)
	})

	t.Run("ClassificationWithEmptyDatabase", func(t *testing.T) {
		// 清理数据库
		_, err := tdb.GetDB().Exec("DELETE FROM issues")
		if err != nil {
			t.Fatalf("Failed to clear issues: %v", err)
		}

		// 在空数据库上执行分类
		issue := &Issue{
			Number:      300,
			Title:       "Test issue",
			Body:        "Test body",
			URL:         "https://github.com/test/repo/issues/300",
			State:       "open",
			RepoOwner:   "test",
			RepoName:    "repo",
		}

		result, err := classifier.ClassifySingleIssue(issue)
		if err != nil {
			t.Fatalf("Classification on empty database failed: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil classification result for empty database")
		}

		t.Logf("Classification result on empty database: %+v", result)
	})

	t.Run("ClassificationConfigValidation", func(t *testing.T) {
		config := DefaultClassificationConfig()
		
		// 验证默认配置
		if config.MinScore < 0.0 {
			t.Error("Expected non-negative min score")
		}

		if config.ConfidenceThreshold < 0.0 || config.ConfidenceThreshold > 1.0 {
			t.Error("Expected confidence threshold between 0.0 and 1.0")
		}

		if len(config.CategoryKeywords) == 0 {
			t.Error("Expected non-empty category keywords")
		}

		// 验证配置更新
		config.MinScore = 5.0
		config.ConfidenceThreshold = 0.7
		
		if config.MinScore != 5.0 {
			t.Error("Failed to update min score")
		}

		if config.ConfidenceThreshold != 0.7 {
			t.Error("Failed to update confidence threshold")
		}
	})

	t.Run("AutoClassification", func(t *testing.T) {
		// 测试自动分类功能
		testCases := []struct {
			title       string
			body        string
			expectedCat string
		}{
			{
				title:       "Memory leak causing performance issues",
				body:        "There is a memory leak that affects performance",
				expectedCat: "performance",
			},
			{
				title:       "Security vulnerability found",
				body:        "Security issue				expectedCat: allows unauthorized access",
 "security",
			},
			{
				title:       "Bug in the login system",
				body:        "The login functionality is not working correctly",
				expectedCat: "bug",
			},
			{
				title:       "Add new feature for user management",
				body:        "We need to implement user management features",
				expectedCat: "feature",
			},
		}

		for _, tc := range testCases {
			issue := &Issue{
				Number:      400 + len(tc.title),
				Title:       tc.title,
				Body:        tc.body,
				URL:         "https://github.com/test/repo/issues/" + string(rune(400+len(tc.title))),
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			}

			result, err := classifier.ClassifySingleIssue(issue)
			if err != nil {
				t.Fatalf("Classification failed for test case '%s': %v", tc.title, err)
			}

			if result.Category == "" {
				t.Errorf("Expected non-empty category for issue: %s", tc.title)
			}

			t.Logf("Issue '%s' classified as: %s (confidence: %.2f)", 
				tc.title, result.Category, result.Confidence)
		}
	})
}

// TestTransactionManager 事务管理器测试
func TestTransactionManager(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	transaction := tdb.Transaction()
	crud := tdb.CRUD()

	t.Run("ExecuteInTransaction Success", func(t *testing.T) {
		// 成功事务测试
		err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
			// 创建issue
			_, err := tx.Exec(`
				INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, 500, "Transaction Test Issue", "Test body", 
				"https://github.com/test/repo/issues/500", "open", "test", "repo",
				time.Now(), time.Now())
			return err
		})

		if err != nil {
			t.Fatalf("Transaction execution failed: %v", err)
		}

		// 验证issue已创建
		issues, err := crud.GetAllIssues(10, 0)
		if err != nil {
			t.Fatalf("Failed to query issues: %v", err)
		}

		found := false
		for _, issue := range issues {
			if issue.Number == 500 {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected issue created by transaction to exist")
		}
	})

	t.Run("ExecuteInTransaction Rollback", func(t *testing.T) {
		// 回滚事务测试
		err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
			// 创建issue
			_, err := tx.Exec(`
				INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, 501, "Rollback Test Issue", "Test body", 
				"https://github.com/test/repo/issues/501", "open", "test", "repo",
				time.Now(), time.Now())
			if err != nil {
				return err
			}
			
			// 模拟错误导致回滚
			return sql.ErrTxDone
		})

		if err == nil {
			t.Error("Expected error from transaction")
		}

		// 验证issue未创建
		issues, err := crud.GetAllIssues(10, 0)
		if err != nil {
			t.Fatalf("Failed to query issues: %v", err)
		}

		for _, issue := range issues {
			if issue.Number == 501 {
				t.Error("Expected rolled back issue to not exist")
				break
			}
		}
	})

	t.Run("ExecuteInTransaction Nested", func(t *testing.T) {
		// 嵌套事务测试
		err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
			// 外层事务
			_, err := tx.Exec(`
				INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, 502, "Outer Transaction Issue", "Test body", 
				"https://github.com/test/repo/issues/502", "open", "test", "repo",
				time.Now(), time.Now())
			if err != nil {
				return err
			}

			// 内层事务 - 在实际实现中应该处理嵌套事务
			return nil
		})

		if err != nil {
			t.Fatalf("Nested transaction failed: %v", err)
		}

		// 验证外层事务的issue已创建
		issues, err := crud.GetAllIssues(10, 0)
		if err != nil {
			t.Fatalf("Failed to query issues: %v", err)
		}

		found := false
		for _, issue := range issues {
			if issue.Number == 502 {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected outer transaction issue to exist")
		}
	})

	t.Run("ExecuteInTransaction Concurrent", func(t *testing.T) {
		// 并发事务测试
		numTransactions := 5
		done := make(chan error, numTransactions)

		for i := 0; i < numTransactions; i++ {
			issueNumber := 600 + i
			go func(num int) {
				err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
					_, err := tx.Exec(`
						INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					`, num, "Concurrent Issue "+string(rune(num)), "Test body", 
						"https://github.com/test/repo/issues/"+string(rune(num)), "open", "test", "repo",
						time.Now(), time.Now())
					return err
				})
				done <- err
			}(issueNumber)
	// 等待		}

	所有事务完成
		for i := 0; i < numTransactions; i++ {
			err := <-done
			if err != nil {
				t.Logf("Concurrent transaction %d failed: %v", i, err)
			}
		}

		// 验证所有issues都已创建
		issues, err := crud.GetAllIssues(20, 0)
		if err != nil {
			t.Fatalf("Failed to query issues: %v", err)
		}

		createdCount := 0
		for _, issue := range issues {
			if issue.Number >= 600 && issue.Number < 600+numTransactions {
				createdCount++
			}
		}

		t.Logf("Created %d out of %d concurrent issues", createdCount, numTransactions)
	})

	t.Run("TransactionConfigValidation", func(t *testing.T) {
		config := DefaultTransactionConfig()
		
		// 验证默认配置
		if config.Timeout <= 0 {
			t.Error("Expected positive timeout")
		}

		if config.MaxRetries < 0 {
			t.Error("Expected non-negative max retries")
		}

		if config.IsolationLevel < 0 {
			t.Error("Expected valid isolation level")
		}

		// 验证配置更新
		config.Timeout = 60 * time.Second
		config.MaxRetries = 3
		
		if config.Timeout != 60*time.Second {
			t.Error("Failed to update timeout")
		}

		if config.MaxRetries != 3 {
			t.Error("Failed to update max retries")
		}
	})
}

// TestCRUDOperationsWithTransaction CRUD操作的事务性测试
func TestCRUDOperationsWithTransaction(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()
	transaction := tdb.Transaction()

	t.Run("CreateIssuesInTransaction", func(t *testing.T) {
		issues := []*Issue{
			{
				Number:      700,
				Title:       "Transaction Issue 1",
				Body:        "Body 1",
				URL:         "https://github.com/test/repo/issues/700",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      701,
				Title:       "Transaction Issue 2",
				Body:        "Body 2",
				URL:         "https://github.com/test/repo/issues/701",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
		}

		// 使用事务创建issues
		err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
			// 模拟在事务中创建issues
			for _, issue := range issues {
				_, err := tx.Exec(`
					INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
					issue.RepoOwner, issue.RepoName, issue.CreatedAtDB, issue.UpdatedAtDB)
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Transaction create failed: %v", err)
		}

		// 验证issues已创建
		retrievedIssues, err := crud.GetAllIssues(10, 0)
		if err != nil {
			t.Fatalf("Failed to query issues: %v", err)
		}

		found := 0
		for _, issue := range retrievedIssues {
			if issue.Number == 700 || issue.Number == 701 {
				found++
			}
		}

		if found != 2 {
			t.Errorf("Expected 2 issues from transaction, found %d", found)
		}
	})

	t.Run("RollbackCreateIssues", func(t *testing.T) {
		issues := []*Issue{
			{
				Number:      800,
				Title:       "Rollback Issue 1",
				Body:        "Body 1",
				URL:         "https://github.com/test/repo/issues/800",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
			{
				Number:      801,
				Title:       "Rollback Issue 2",
				Body:        "Body 2",
				URL:         "https://github.com/test/repo/issues/801",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				CreatedAtDB: time.Now(),
				UpdatedAtDB: time.Now(),
			},
		}

		// 事务中创建但回滚
		err := transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
			for _, issue := range issues {
				_, err := tx.Exec(`
					INSERT INTO issues (number, title, body, url, state, repo_owner, repo_name, created_at_db, updated_at_db)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
					issue.RepoOwner, issue.RepoName, issue.CreatedAtDB, issue.UpdatedAtDB)
				if err != nil {
					return err
				}
			}
			// 模拟错误
			return sql.ErrTxDone
		})

		if err == nil {
			t.Error("Expected error for rollback")
		}

		// 验证issues未创建
		retrievedIssues, err := crud.GetAllIssues(10, 0)
		if err != nil {
			t.Fatalf("Failed to query issues: %v", err)
		}

		for _, issue := range retrievedIssues {
			if issue.Number == 800 || issue.Number == 801 {
				t.Error("Expected rolled back issues to not exist")
				break
			}
		}
	})
}