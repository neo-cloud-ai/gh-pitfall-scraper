package database

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

// TestCRUDOperationsQuery 查询操作测试
func TestCRUDOperationsQuery(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	// 填充测试数据
	tr := NewTestRepository(tdb.GetDB())
	if err := tr.SeedTestData(); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	crud := tdb.CRUD()

	t.Run("GetAllIssues", func(t *testing.T) {
		// 测试基本查询
		issues, err := crud.GetAllIssues(10, 0)
		if err != nil {
			t.Fatalf("Failed to get all issues: %v", err)
		}

		if len(issues) == 0 {
			t.Error("Expected at least one issue")
		}

		// 验证排序（应该按创建时间倒序）
		if len(issues) > 1 {
			for i := 0; i < len(issues)-1; i++ {
				if issues[i].CreatedAtDB.Before(issues[i+1].CreatedAtDB) {
					t.Error("Issues are not sorted by created_at_db DESC")
				}
			}
		}

		t.Logf("Retrieved %d issues", len(issues))
	})

	t.Run("GetAllIssues Pagination", func(t *testing.T) {
		// 测试分页
		page1, err := crud.GetAllIssues(2, 0)
		if err != nil {
			t.Fatalf("Failed to get page 1: %v", err)
		}

		page2, err := crud.GetAllIssues(2, 2)
		if err != nil {
			t.Fatalf("Failed to get page 2: %v", err)
		}

		if len(page1) != 2 || len(page2) != 2 {
			t.Errorf("Expected 2 issues per page, got page1: %d, page2: %d", len(page1), len(page2))
		}

		// 验证分页结果不重叠
		for _, issue1 := range page1 {
			for _, issue2 := range page2 {
				if issue1.ID == issue2.ID {
					t.Error("Found duplicate issues across pages")
				}
			}
		}

		// 测试偏移量超出范围
		emptyPage, err := crud.GetAllIssues(10, 1000)
		if err != nil {
			t.Fatalf("Failed to get empty page: %v", err)
		}
		if len(emptyPage) != 0 {
			t.Error("Expected empty result for offset beyond data range")
		}
	})

	t.Run("GetIssuesByRepository", func(t *testing.T) {
		// 测试按仓库查询
		repoIssues, err := crud.GetIssuesByRepository("test", "repo", 10, 0)
		if err != nil {
			t.Fatalf("Failed to get repository issues: %v", err)
		}

		// 验证所有issue都属于指定仓库
		for _, issue := range repoIssues {
			if issue.RepoOwner != "test" || issue.RepoName != "repo" {
				t.Errorf("Issue from wrong repository: %s/%s", issue.RepoOwner, issue.RepoName)
			}
		}

		// 测试不存在的仓库
		emptyIssues, err := crud.GetIssuesByRepository("nonexistent", "repo", 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues for nonexistent repository: %v", err)
		}
		if len(emptyIssues) != 0 {
			t.Error("Expected no issues for nonexistent repository")
		}

		t.Logf("Found %d issues for test/repo", len(repoIssues))
	})

	t.Run("GetIssuesByScore", func(t *testing.T) {
		// 测试按分数范围查询
		scoreIssues, err := crud.GetIssuesByScore(20.0, 30.0, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by score: %v", err)
		}

		// 验证分数在范围内
		for _, issue := range scoreIssues {
			if issue.Score < 20.0 || issue.Score > 30.0 {
				t.Errorf("Issue score %f outside range [20.0, 30.0]", issue.Score)
			}
		}

		// 测试降序排序
		if len(scoreIssues) > 1 {
			for i := 0; i < len(scoreIssues)-1; i++ {
				if scoreIssues[i].Score < scoreIssues[i+1].Score {
					t.Error("Issues are not sorted by score DESC")
				}
			}
		}

		// 测试边界值
		exactMatch, err := crud.GetIssuesByScore(25.0, 25.0, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues with exact score match: %v", err)
		}

		for _, issue := range exactMatch {
			if issue.Score != 25.0 {
				t.Errorf("Expected exact score match 25.0, got %f", issue.Score)
			}
		}

		t.Logf("Found %d issues with score 20-30", len(scoreIssues))
	})

	t.Run("GetIssuesByCategory", func(t *testing.T) {
		// 测试按分类查询
		categoryIssues, err := crud.GetIssuesByCategory("performance", 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by category: %v", err)
		}

		// 验证分类
		for _, issue := range categoryIssues {
			if issue.Category != "performance" {
				t.Errorf("Expected category 'performance', got '%s'", issue.Category)
			}
		}

		// 测试不存在的分类
		emptyIssues, err := crud.GetIssuesByCategory("nonexistent", 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues for nonexistent category: %v", err)
		}
		if len(emptyIssues) != 0 {
			t.Error("Expected no issues for nonexistent category")
		}

		t.Logf("Found %d issues in performance category", len(categoryIssues))
	})

	t.Run("GetIssuesByPriority", func(t *testing.T) {
		// 测试按优先级查询
		priorityIssues, err := crud.GetIssuesByPriority("high", 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by priority: %v", err)
		}

		// 验证优先级
		for _, issue := range priorityIssues {
			if issue.Priority != "high" {
				t.Errorf("Expected priority 'high', got '%s'", issue.Priority)
			}
		}

		t.Logf("Found %d issues with high priority", len(priorityIssues))
	})

	t.Run("GetIssuesByKeywords", func(t *testing.T) {
		// 测试单个关键词
		keywordIssues, err := crud.GetIssuesByKeywords([]string{"memory"}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by keywords: %v", err)
		}

		if len(keywordIssues) == 0 {
			t.Log("No issues found with 'memory' keyword")
		} else {
			t.Logf("Found %d issues with 'memory' keyword", len(keywordIssues))
		}

		// 测试多个关键词（AND条件）
		multiKeywordIssues, err := crud.GetIssuesByKeywords([]string{"performance", "memory"}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by multiple keywords: %v", err)
		}

		t.Logf("Found %d issues with both 'performance' and 'memory' keywords", len(multiKeywordIssues))

		// 测试空关键词列表
		allIssues, err := crud.GetIssuesByKeywords([]string{}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues with empty keywords: %v", err)
		}

		// 应该返回所有issues
		if len(allIssues) == 0 {
			t.Error("Expected issues when using empty keyword list")
		}

		// 测试不存在
		nonexistentIssues, err := crud.GetIssuesByKeywords([]string{"nonexistentkeyword123"}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues for nonexistent keyword: %v", err)
		}
		if len(nonexistentIssues) != 0 {
			t.Error("Expected no issues for nonexistent keyword")
		}
	})

	t.Run("GetIssuesByTechStack", func(t *testing.T) {
		// 测试按技术栈查询
		techStackIssues, err := crud.GetIssuesByTechStack([]string{"go"}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by tech stack: %v", err)
		}

		if len(techStackIssues) == 0 {
			t.Log("No issues found with 'go' tech stack")
		} else {
			t.Logf("Found %d issues with 'go' tech stack", len(techStackIssues))
		}

		// 测试多个技术栈
		multiTechIssues, err := crud.GetIssuesByTechStack([]string{"go", "database"}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues by multiple tech stacks: %v", err)
		}

		t.Logf("Found %d issues with both 'go' and 'database' tech stacks", len(multiTechIssues))

		// 测试空技术栈列表
		allTechIssues, err := crud.GetIssuesByTechStack([]string{}, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get issues with empty tech stack: %v", err)
		}

		if len(allTechIssues) == 0 {
			t.Error("Expected issues when using empty tech stack list")
		}
	})

	t.Run("SearchIssues", func(t *testing.T) {
		// 测试基本搜索
		searchResults, err := crud.SearchIssues("memory", 10, 0)
		if err != nil {
			t.Fatalf("Failed to search issues: %v", err)
		}

		if len(searchResults) == 0 {
			t.Log("No issues found for search term 'memory'")
		} else {
			t.Logf("Found %d issues matching 'memory'", len(searchResults))
		}

		// 测试空搜索词
		allSearchResults, err := crud.SearchIssues("", 10, 0)
		if err != nil {
			t.Fatalf("Failed to search with empty query: %v", err)
		}

		// 应该返回所有issues
		if len(allSearchResults) == 0 {
			t.Error("Expected issues when using empty search query")
		}

		// 测试不存在的搜索词
		noResults, err := crud.SearchIssues("nonexistentterm123", 10, 0)
		if err != nil {
			t.Fatalf("Failed to search for nonexistent term: %v", err)
		}
		if len(noResults) != 0 {
			t.Error("Expected no results for nonexistent search term")
		}

		// 测试部分匹配
		partialResults, err := crud.SearchIssues("cache", 10, 0)
		if err != nil {
			t.Fatalf("Failed to search for 'cache': %v", err)
		}

		t.Logf("Found %d issues matching 'cache'", len(partialResults))
	})

	t.Run("SearchIssuesAdvanced", func(t *testing.T) {
		// 测试高级搜索 - 基本条件
		search := &AdvancedSearch{
			Query:       "memory",
			Categories:  []string{"performance"},
			SortBy:      "score",
			SortOrder:   "DESC",
			Limit:       10,
			Offset:      0,
		}

		results, err := crud.SearchIssuesAdvanced(search)
		if err != nil {
			t.Fatalf("Failed to perform advanced search: %v", err)
		}

		if len(results) == 0 {
			t.Log("No issues found for advanced search")
		} else {
			t.Logf("Found %d issues for advanced search", len(results))
		}

		// 测试高级搜索 - 多个条件
		search2 := &AdvancedSearch{
			Query:       "",
			Categories:  []string{"performance", "security"},
			Priorities:  []string{"high", "critical"},
			States:      []string{"open"},
			MinScore:    func() *float64 { score := 15.0; return &score }(),
			SortBy:      "created_at",
			SortOrder:   "ASC",
			Limit:       5,
			Offset:      0,
		}

		results2, err := crud.SearchIssuesAdvanced(search2)
		if err != nil {
			t.Fatalf("Failed to perform complex advanced search: %v", err)
		}

		t.Logf("Found %d issues for complex advanced search", len(results2))

		// 测试日期范围搜索
		now := time.Now()
		dateFrom := now.AddDate(0, -1, 0) // 一个月前

		search3 := &AdvancedSearch{
			DateFrom:    &dateFrom,
			DateTo:      &now,
			Limit:       10,
			Offset:      0,
		}

		results3, err := crud.SearchIssuesAdvanced(search3)
		if err != nil {
			t.Fatalf("Failed to perform date range search: %v", err)
		}

		t.Logf("Found %d issues in date range", len(results3))

		// 测试排除重复
		search4 := &AdvancedSearch{
			ExcludeDuplicates: true,
			Limit:             10,
			Offset:            0,
		}

		results4, err := crud.SearchIssuesAdvanced(search4)
		if err != nil {
			t.Fatalf("Failed to perform search excluding duplicates: %v", err)
		}

		t.Logf("Found %d non-duplicate issues", len(results4))

		// 测试JSON数组过滤
		search5 := &AdvancedSearch{
			Keywords:    []string{"performance", "bug"},
			TechStacks:  []string{"go"},
			Limit:       10,
			Offset:      0,
		}

		results5, err := crud.SearchIssuesAdvanced(search5)
		if err != nil {
			t.Fatalf("Failed to perform JSON array search: %v", err)
		}

		t.Logf("Found %d issues matching JSON array filters", len(results5))
	})

	t.Run("AdvancedSearch Default Values", func(t *testing.T) {
		// 测试默认搜索条件
		search := DefaultAdvancedSearch()

		if search.SortBy != "score" {
			t.Errorf("Expected default sort by 'score', got '%s'", search.SortBy)
		}

		if search.SortOrder != "DESC" {
			t.Errorf("Expected default sort order 'DESC', got '%s'", search.SortOrder)
		}

		if search.Limit != 100 {
			t.Errorf("Expected default limit 100, got %d", search.Limit)
		}

		if search.Offset != 0 {
			t.Errorf("Expected default offset 0, got %d", search.Offset)
		}

		if !search.ExcludeDuplicates {
			t.Error("Expected default ExcludeDuplicates to be true")
		}
	})

	t.Run("AdvancedSearch Repository Filter", func(t *testing.T) {
		// 测试仓库过滤
		search := &AdvancedSearch{
			Repos:       []string{"test/repo", "golang/go"},
			Limit:       10,
			Offset:      0,
		}

		results, err := crud.SearchIssuesAdvanced(search)
		if err != nil {
			t.Fatalf("Failed to perform repository filter search: %v", err)
		}

		// 验证所有结果都属于指定仓库
		for _, issue := range results {
			found := false
			for _, repo := range search.Repos {
				parts := split(repo, "/")
				if len(parts) == 2 && issue.RepoOwner == parts[0] && issue.RepoName == parts[1] {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Issue from unexpected repository: %s/%s", issue.RepoOwner, issue.RepoName)
			}
		}

		t.Logf("Found %d issues for repository filter", len(results))
	})
}

// TestRepositoryOperations 仓库操作测试
func TestRepositoryOperations(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	t.Run("GetAllRepositories", func(t *testing.T) {
		// 创建测试仓库
		repos := []*Repository{
			CreateTestRepository(),
			{
				Owner:       "golang",
				Name:        "go",
				FullName:    "golang/go",
				Description: "The Go programming language",
				URL:         "https://github.com/golang/go",
				Language:    "Go",
				Stars:       120000,
				Forks:       18000,
				IssuesCount: 5000,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		for _, repo := range repos {
			_, err := crud.CreateRepository(repo)
			if err != nil {
				t.Fatalf("Failed to create repository: %v", err)
			}
		}

		// 测试获取所有仓库
		allRepos, err := crud.GetAllRepositories(10, 0)
		if err != nil {
			t.Fatalf("Failed to get all repositories: %v", err)
		}

		if len(allRepos) < 2 {
			t.Errorf("Expected at least 2 repositories, got %d", len(allRepos))
		}

		// 验证排序（应该按stars降序）
		if len(allRepos) > 1 {
			for i := 0; i < len(allRepos)-1; i++ {
				if allRepos[i].Stars < allRepos[i+1].Stars {
					t.Error("Repositories are not sorted by stars DESC")
				}
			}
		}

		t.Logf("Retrieved %d repositories", len(allRepos))
	})

	t.Run("GetAllRepositories Pagination", func(t *testing.T) {
		// 测试分页
		page1, err := crud.GetAllRepositories(1, 0)
		if err != nil {
			t.Fatalf("Failed to get repository page 1: %v", err)
		}

		page2, err := crud.GetAllRepositories(1, 1)
		if err != nil {
			t.Fatalf("Failed to get repository page 2: %v", err)
		}

		if len(page1) != 1 || len(page2) != 1 {
			t.Errorf("Expected 1 repository per page, got page1: %d, page2: %d", len(page1), len(page2))
		}

		// 验证分页结果不重叠
		if page1[0].ID == page2[0].ID {
			t.Error("Found duplicate repositories across pages")
		}
	})

	t.Run("GetRepositoryByName", func(t *testing.T) {
		// 创建测试仓库
		repo := CreateTestRepository()
		id, err := crud.CreateRepository(repo)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// 测试按名称获取
		retrievedRepo, err := crud.GetRepositoryByName("test", "repo")
		if err != nil {
			t.Fatalf("Failed to get repository by name: %v", err)
		}

		if retrievedRepo == nil {
			t.Error("Expected non-nil repository")
			return
		}

		if retrievedRepo.ID != int(id) {
			t.Errorf("Expected repository ID %d, got %d", id, retrievedRepo.ID)
		}

		if retrievedRepo.Owner != "test" || retrievedRepo.Name != "repo" {
			t.Errorf("Expected repository test/repo, got %s/%s", retrievedRepo.Owner, retrievedRepo.Name)
		}

		// 测试获取不存在的仓库
		_, err = crud.GetRepositoryByName("nonexistent", "repo")
		if err == nil {
			t.Error("Expected error for nonexistent repository")
		}
	})

	t.Run("Repository CRUD Operations", func(t *testing.T) {
		// 测试创建
		repo := CreateTestRepository()
		id, err := crud.CreateRepository(repo)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// 测试获取
		retrievedRepo, err := crud.GetRepository(id)
		if err != nil {
			t.Fatalf("Failed to get repository: %v", err)
		}

		if retrievedRepo.Owner != repo.Owner {
			t.Errorf("Expected owner %s, got %s", repo.Owner, retrievedRepo.Owner)
		}

		// 测试更新
		repo.Stars = 200
		repo.Description = "Updated description"
		err = crud.UpdateRepository(repo)
		if err != nil {
			t.Fatalf("Failed to update repository: %v", err)
		}

		// 验证更新
		updatedRepo, err := crud.GetRepository(id)
		if err != nil {
			t.Fatalf("Failed to get updated repository: %v", err)
		}

		if updatedRepo.Stars != 200 {
			t.Errorf("Expected stars 200, got %d", updatedRepo.Stars)
		}

		if updatedRepo.Description != "Updated description" {
			t.Errorf("Expected updated description, got %s", updatedRepo.Description)
		}

		// 测试删除
		err = crud.DeleteRepository(id)
		if err != nil {
			t.Fatalf("Failed to delete repository: %v", err)
		}

		// 验证删除
		_, err = crud.GetRepository(id)
		if err == nil {
			t.Error("Expected error for deleted repository")
		}
	})
}

// TestStatistics 统计信息测试
func TestStatistics(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	// 填充测试数据
	tr := NewTestRepository(tdb.GetDB())
	if err := tr.SeedTestData(); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	crud := tdb.CRUD()

	t.Run("GetIssueStats", func(t *testing.T) {
		stats, err := crud.GetIssueStats()
		if err != nil {
			t.Fatalf("Failed to get issue stats: %v", err)
		}

		if stats == nil {
			t.Error("Expected non-nil issue stats")
			return
		}

		// 验证基本统计
		if stats.TotalCount <= 0 {
			t.Error("Expected positive total count")
		}

		// 验证分类分布
		if len(stats.ByCategory) == 0 {
			t.Log("No category distribution found")
		} else {
			t.Logf("Category distribution: %+v", stats.ByCategory)
		}

		// 验证优先级分布
		if len(stats.ByPriority) == 0 {
			t.Log("No priority distribution found")
		} else {
			t.Logf("Priority distribution: %+v", stats.ByPriority)
		}

		// 验证技术栈分布
		if len(stats.ByTechStack) == 0 {
			t.Log("No tech stack distribution found")
		} else {
			t.Logf("Tech stack distribution: %+v", stats.ByTechStack)
		}

		// 验证状态分布
		if len(stats.ByState) == 0 {
			t.Log("No state distribution found")
		} else {
			t.Logf("State distribution: %+v", stats.ByState)
		}

		// 验证分数分布
		if len(stats.ScoreDistribution) == 0 {
			t.Log("No score distribution found")
		} else {
			t.Logf("Score distribution: %+v", stats.ScoreDistribution)
		}

		// 验证平均分数
		if stats.AverageScore < 0.0 {
			t.Error("Expected non-negative average score")
		} else {
			t.Logf("Average score: %.2f", stats.AverageScore)
		}

		// 验证热门关键词
		if len(stats.TopKeywords) == 0 {
			t.Log("No top keywords found")
		} else {
			t.Logf("Top keywords: %+v", stats.TopKeywords)
		}

		// 验证日期范围
		if len(stats.DateRange) == 0 {
			t.Log("No date range found")
		} else {
			t.Logf("Date range: %+v", stats.DateRange)
		}
	})

	t.Run("GetRepositoryStats", func(t *testing.T) {
		// 测试特定仓库的统计
		stats, err := crud.GetRepositoryStats("test", "repo")
		if err != nil {
			t.Fatalf("Failed to get repository stats: %v", err)
		}

		if stats == nil {
			t.Error("Expected non-nil repository stats")
			return
		}

		// 验证统计信息
		if stats.TotalCount < 0 {
			t.Error("Expected non-negative total count")
		}

		t.Logf("Repository stats for test/repo: %+v", stats)

		// 测试不存在仓库的统计
		emptyStats, err := crud.GetRepositoryStats("nonexistent", "repo")
		if err != nil {
			t.Fatalf("Failed to get stats for nonexistent repository: %v", err)
		}

		if emptyStats.TotalCount != 0 {
			t.Errorf("Expected 0 count for nonexistent repository, got %d", emptyStats.TotalCount)
		}
	})

	t.Run("Statistics Consistency", func(t *testing.T) {
		// 获取总体统计
		overallStats, err := crud.GetIssueStats()
		if err != nil {
			t.Fatalf("Failed to get overall stats: %v", err)
		}

		// 获取特定仓库统计
		repoStats, err := crud.GetRepositoryStats("test", "repo")
		if err != nil {
			t.Fatalf("Failed to get repo stats: %v", err)
		}

		// 验证总体统计的总数应该大于等于任何单个仓库的统计
		if overallStats.TotalCount < repoStats.TotalCount {
			t.Errorf("Overall count %d should be >= repo count %d", 
				overallStats.TotalCount, repoStats.TotalCount)
		}

		t.Logf("Statistics consistency check passed: overall=%d, test/repo=%d", 
			overallStats.TotalCount, repoStats.TotalCount)
	})
}

// TestMaintenance 维护操作测试
func TestMaintenance(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)

	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	crud := tdb.CRUD()

	t.Run("Cleanup Operation", func(t *testing.T) {
		// 创建老旧的重复数据
		oldIssues := []*Issue{
			{
				Number:      1000,
				Title:       "Old duplicate issue 1",
				Body:        "Old body content",
				URL:         "https://github.com/test/repo/issues/1000",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				IsDuplicate: true,
				CreatedAtDB: time.Now().AddDate(-2, 0, 0), // 2年前
				UpdatedAtDB: time.Now().AddDate(-2, 0, 0),
			},
			{
				Number:      1001,
				Title:       "Old duplicate issue 2",
				Body:        "Another old body",
				URL:         "https://github.com/test/repo/issues/1001",
				State:       "open",
				RepoOwner:   "test",
				RepoName:    "repo",
				IsDuplicate: true,
				CreatedAtDB: time.Now().AddDate(-2, 0, 0),
				UpdatedAtDB: time.Now().AddDate(-2, 0, 0),
			},
		}

		_, err := crud.CreateIssues(oldIssues)
		if err != nil {
			t.Fatalf("Failed to create old issues: %v", err)
		}

		// 记录清理前的数量
		var beforeCount int
		err = tdb.GetDB().QueryRow("SELECT COUNT(*) FROM issues").Scan(&beforeCount)
		if err != nil {
			t.Fatalf("Failed to count issues before cleanup: %v", err)
		}

		// 执行清理
		err = crud.Cleanup()
		if err != nil {
			t.Fatalf("Cleanup failed: %v", err)
		}

		// 记录清理后的数量
		var afterCount int
		err = tdb.GetDB().QueryRow("SELECT COUNT(*) FROM issues").Scan(&afterCount)
		if err != nil {
			t.Fatalf("Failed to count issues after cleanup: %v", err)
		}

		// 验证清理效果
		if afterCount >= beforeCount {
			t.Log("No issues were cleaned up (may be expected if cleanup criteria not met)")
		} else {
			t.Logf("Cleanup removed %d issues", beforeCount-afterCount)
		}
	})

	t.Run("Optimize Operation", func(t *testing.T) {
		// 执行优化
		err := crud.Optimize()
		if err != nil {
			t.Fatalf("Optimization failed: %v", err)
		}

		t.Log("Database optimization completed successfully")
	})
}

// 辅助函数：分割字符串
func split(s, sep string) []string {
	if len(s) == 0 {
		return []string{}
	}
	
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == []byte(sep)[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}