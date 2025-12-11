package database

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestExporter 测试导出器功能
func TestExporter(t *testing.T) {
	// 创建测试数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// 创建测试表和数据
	createTestTables(t, db)
	insertTestData(t, db)

	// 创建导出器
	exporter := NewExporter(db)

	// 测试JSON导出
	t.Run("JSON Export", func(t *testing.T) {
		filter := ExportFilter{
			IncludeMetadata: true,
		}

		result, err := exporter.ExportIssues(filter, FormatJSON, "test_output.json")
		if err != nil {
			t.Fatalf("JSON export failed: %v", err)
		}

		if result == nil {
			t.Fatal("Export result is nil")
		}

		if result.Format != FormatJSON {
			t.Errorf("Expected format %s, got %s", FormatJSON, result.Format)
		}

		if result.ExportedRecords <= 0 {
			t.Error("Expected positive export count")
		}

		t.Logf("JSON export completed: %d records exported", result.ExportedRecords)
	})

	// 测试CSV导出
	t.Run("CSV Export", func(t *testing.T) {
		filter := ExportFilter{
			IncludeMetadata: true,
		}

		result, err := exporter.ExportIssues(filter, FormatCSV, "test_output.csv")
		if err != nil {
			t.Fatalf("CSV export failed: %v", err)
		}

		if result == nil {
			t.Fatal("Export result is nil")
		}

		if result.Format != FormatCSV {
			t.Errorf("Expected format %s, got %s", FormatCSV, result.Format)
		}

		t.Logf("CSV export completed: %d records exported", result.ExportedRecords)
	})

	// 测试带过滤条件的导出
	t.Run("Filtered Export", func(t *testing.T) {
		now := time.Now()
		dateFrom := now.AddDate(0, -1, 0)
		filter := ExportFilter{
			DateFrom:        &dateFrom,
			IncludeMetadata: true,
		}

		result, err := exporter.ExportIssues(filter, FormatJSON, "test_filtered.json")
		if err != nil {
			t.Fatalf("Filtered export failed: %v", err)
		}

		if result == nil {
			t.Fatal("Export result is nil")
		}

		t.Logf("Filtered export completed: %d records exported", result.ExportedRecords)
	})
}

// TestQueryBuilder 测试查询构建器功能
func TestQueryBuilder(t *testing.T) {
	// 创建测试数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// 创建测试表和数据
	createTestTables(t, db)
	insertTestData(t, db)

	// 创建查询构建器
	qb := NewQueryBuilder(db)

	// 测试简单查询
	t.Run("Simple Query", func(t *testing.T) {
		criteria := DefaultSearchCriteria()
		criteria.Page = 1
		criteria.PageSize = 10

		result, err := qb.SimpleQuery(criteria)
		if err != nil {
			t.Fatalf("Simple query failed: %v", err)
		}

		if result == nil {
			t.Fatal("Query result is nil")
		}

		if result.TotalCount <= 0 {
			t.Error("Expected positive total count")
		}

		if len(result.Issues) == 0 {
			t.Error("Expected at least one issue")
		}

		t.Logf("Simple query completed: %d total records, %d returned", result.TotalCount, len(result.Issues))
	})

	// 测试带条件的查询
	t.Run("Filtered Query", func(t *testing.T) {
		criteria := DefaultSearchCriteria()
		criteria.Query = "test"
		criteria.Page = 1
		criteria.PageSize = 10

		result, err := qb.SimpleQuery(criteria)
		if err != nil {
			t.Fatalf("Filtered query failed: %v", err)
		}

		if result == nil {
			t.Fatal("Query result is nil")
		}

		t.Logf("Filtered query completed: %d records found", len(result.Issues))
	})

	// 测试聚合查询
	t.Run("Aggregated Query", func(t *testing.T) {
		aggCriteria := AggregatedQuery{
			GroupBy: "category",
			Metrics: []string{"count", "avg_score"},
			Filters: DefaultSearchCriteria(),
		}

		result, err := qb.AggregatedQuery(aggCriteria)
		if err != nil {
			t.Fatalf("Aggregated query failed: %v", err)
		}

		if result == nil {
			t.Fatal("Aggregation result is nil")
		}

		if result.TotalGroups <= 0 {
			t.Error("Expected positive group count")
		}

		t.Logf("Aggregated query completed: %d groups", result.TotalGroups)
	})

	// 测试时间序列查询
	t.Run("Time Series Query", func(t *testing.T) {
		endDate := time.Now()
		startDate := endDate.AddDate(0, -1, 0)

		tsq := TimeSeriesQuery{
			StartDate: startDate,
			EndDate:   endDate,
			Interval:  "day",
		}

		result, err := qb.TimeSeriesQuery(tsq)
		if err != nil {
			t.Fatalf("Time series query failed: %v", err)
		}

		if result == nil {
			t.Fatal("Time series result is nil")
		}

		t.Logf("Time series query completed: %d points", result.TotalPoints)
	})

	// 测试分面查询
	t.Run("Faceted Query", func(t *testing.T) {
		fq := FacetedQuery{
			BaseCriteria: DefaultSearchCriteria(),
			Facets:       []string{"category", "state"},
		}

		result, err := qb.FacetedQuery(fq)
		if err != nil {
			t.Fatalf("Faceted query failed: %v", err)
		}

		if result == nil {
			t.Fatal("Faceted result is nil")
		}

		if result.TotalCount <= 0 {
			t.Error("Expected positive total count")
		}

		if len(result.Facets) == 0 {
			t.Error("Expected at least one facet")
		}

		t.Logf("Faceted query completed: %d total records, %d facets", result.TotalCount, len(result.Facets))
	})
}

// TestReportGenerator 测试报告生成器功能
func TestReportGenerator(t *testing.T) {
	// 创建测试数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// 创建测试表和数据
	createTestTables(t, db)
	insertTestData(t, db)

	// 创建报告生成器
	rg := NewReportGenerator(db)

	// 测试报告生成
	t.Run("HTML Report Generation", func(t *testing.T) {
		config := ReportConfig{
			Title:       "Test Report",
			Description: "This is a test report",
			OutputPath:  "test_report.html",
			Format:      "html",
			Parameters: map[string]interface{}{
				"test": true,
			},
		}

		result, err := rg.GenerateReport(config)
		if err != nil {
			t.Fatalf("Report generation failed: %v", err)
		}

		if result == nil {
			t.Fatal("Report result is nil")
		}

		if result.OutputPath == "" {
			t.Error("Expected non-empty output path")
		}

		t.Logf("HTML report generated: %s", result.OutputPath)
	})

	// 测试JSON报告生成
	t.Run("JSON Report Generation", func(t *testing.T) {
		config := ReportConfig{
			Title:       "Test JSON Report",
			Description: "This is a test JSON report",
			OutputPath:  "test_report.json",
			Format:      "json",
			Parameters: map[string]interface{}{
				"test": true,
			},
		}

		result, err := rg.GenerateReport(config)
		if err != nil {
			t.Fatalf("JSON report generation failed: %v", err)
		}

		if result == nil {
			t.Fatal("JSON report result is nil")
		}

		t.Logf("JSON report generated: %s", result.OutputPath)
	})
}

// TestExportFilter 验证导出过滤器
func TestExportFilter(t *testing.T) {
	// 测试默认值
	t.Run("Default Values", func(t *testing.T) {
		filter := ExportFilter{}
		
		// 验证默认值为零值
		if filter.IncludeMetadata != false {
			t.Error("Expected default IncludeMetadata to be false")
		}
	})

	// 测试设置值
	t.Run("Set Values", func(t *testing.T) {
		now := time.Now()
		dateFrom := now.AddDate(0, -1, 0)
		
		filter := ExportFilter{
			DateFrom:        &dateFrom,
			Categories:      []string{"bug", "enhancement"},
			MinScore:        func() *float64 { score := 5.0; return &score }(),
			IncludeMetadata: true,
		}

		if filter.DateFrom == nil {
			t.Error("Expected DateFrom to be set")
		}

		if len(filter.Categories) != 2 {
			t.Errorf("Expected 2 categories, got %d", len(filter.Categories))
		}

		if filter.MinScore == nil || *filter.MinScore != 5.0 {
			t.Error("Expected MinScore to be 5.0")
		}

		if !filter.IncludeMetadata {
			t.Error("Expected IncludeMetadata to be true")
		}
	})
}

// TestSearchCriteria 验证搜索条件
func TestSearchCriteria(t *testing.T) {
	t.Run("Default Search Criteria", func(t *testing.T) {
		criteria := DefaultSearchCriteria()
		
		// 验证默认值
		if criteria.Page != 1 {
			t.Errorf("Expected default page to be 1, got %d", criteria.Page)
		}
		
		if criteria.PageSize != 100 {
			t.Errorf("Expected default page size to be 100, got %d", criteria.PageSize)
		}
		
		if criteria.SortBy != "score" {
			t.Errorf("Expected default sort field to be 'score', got %s", criteria.SortBy)
		}
		
		if criteria.SortOrder != "DESC" {
			t.Errorf("Expected default sort order to be 'DESC', got %s", criteria.SortOrder)
		}
	})

	t.Run("Search Criteria Validation", func(t *testing.T) {
		criteria := DefaultSearchCriteria()
		criteria.Page = 0 // 无效页码
		
		err := ValidateSearchCriteria(criteria)
		if err == nil {
			t.Error("Expected validation error for invalid page")
		}

		criteria.Page = 1
		criteria.SortBy = "invalid_field" // 无效排序字段
		
		err = ValidateSearchCriteria(criteria)
		if err == nil {
			t.Error("Expected validation error for invalid sort field")
		}
	})
}

// 辅助函数：创建测试表
func createTestTables(t *testing.T, db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS repositories (
			id INTEGER PRIMARY KEY,
			owner TEXT,
			name TEXT,
			full_name TEXT,
			description TEXT,
			url TEXT,
			stars INTEGER,
			forks INTEGER,
			language TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			description TEXT,
			color TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS issues (
			id INTEGER PRIMARY KEY,
			issue_id INTEGER,
			repository_id INTEGER,
			number INTEGER,
			title TEXT,
			body TEXT,
			state TEXT,
			author_login TEXT,
			labels TEXT,
			assignees TEXT,
			milestone TEXT,
			reactions TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			closed_at DATETIME,
			first_seen_at DATETIME,
			last_seen_at DATETIME,
			is_pitfall BOOLEAN,
			severity_score REAL,
			category_id INTEGER,
			score REAL,
			url TEXT,
			html_url TEXT,
			comments_count INTEGER,
			is_duplicate BOOLEAN,
			duplicate_of INTEGER,
			metadata TEXT
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}
	}
}

// 辅助函数：插入测试数据
func insertTestData(t *testing.T, db *sql.DB) {
	// 插入测试仓库
	_, err := db.Exec(`
		INSERT INTO repositories (id, owner, name, full_name, description, url, stars, forks, language, created_at, updated_at)
		VALUES (1, 'test', 'repo1', 'test/repo1', 'Test repository', 'https://github.com/test/repo1', 100, 50, 'Go', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test repository: %v", err)
	}

	// 插入测试分类
	_, err = db.Exec(`
		INSERT INTO categories (id, name, description, color, created_at, updated_at)
		VALUES (1, 'bug', 'Bug reports', '#ff0000', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}

	// 插入测试Issues
	testIssues := []struct {
		title  string
		state  string
		score  float64
		category int
	}{
		{"Test issue 1", "open", 8.5, 1},
		{"Test issue 2", "closed", 6.2, 1},
		{"Test bug report", "open", 9.1, 1},
	}

	for i, issue := range testIssues {
		_, err = db.Exec(`
			INSERT INTO issues (
				id, issue_id, repository_id, number, title, body, state, author_login,
				created_at, updated_at, is_pitfall, severity_score, category_id, score,
				url, html_url, comments_count, is_duplicate, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			i+1, int64(i+1), 1, i+1, issue.title, "Test body content", issue.state, "testuser",
			time.Now(), time.Now(), true, issue.score, issue.category, issue.score,
			"https://github.com/test/repo1/issues/"+string(rune(i+1)),
			"https://github.com/test/repo1/issues/"+string(rune(i+1)),
			i, false, "{}",
		)
		if err != nil {
			t.Fatalf("Failed to insert test issue %d: %v", i+1, err)
		}
	}
}

// TestHelperFunctions 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	t.Run("Interface Slice Conversion", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		result := interfaceSlice(input)
		
		if len(result) != len(input) {
			t.Errorf("Expected length %d, got %d", len(input), len(result))
		}
		
		for i, v := range input {
			if result[i] != v {
				t.Errorf("Expected %s at index %d, got %v", v, i, result[i])
			}
		}
	})

	t.Run("String Slice Operations", func(t *testing.T) {
		s1 := JSONSlice{"a", "b", "c"}
		s2 := JSONSlice{"b", "c", "d"}
		
		intersection := intersectStringSlices(s1, s2)
		expectedIntersection := JSONSlice{"b", "c"}
		
		if !reflect.DeepEqual(intersection, expectedIntersection) {
			t.Errorf("Expected intersection %v, got %v", expectedIntersection, intersection)
		}
		
		union := unionStringSlices(s1, s2)
		expectedUnion := JSONSlice{"a", "b", "c", "d"}
		
		if !reflect.DeepEqual(union, expectedUnion) {
			t.Errorf("Expected union %v, got %v", expectedUnion, union)
		}
	})
}