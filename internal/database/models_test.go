package database

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

// TestCustomTime 测试自定义时间类型
func TestCustomTime(t *testing.T) {
	// 测试 Scan 方法
	t.Run("Scan Methods", func(t *testing.T) {
		// 测试 nil 值
		var ct CustomTime
		err := ct.Scan(nil)
		if err != nil {
			t.Errorf("CustomTime Scan(nil) failed: %v", err)
		}
		if !ct.IsZero() {
			t.Error("Expected zero time for nil input")
		}

		// 测试 time.Time 类型
		now := time.Now()
		var ct2 CustomTime
		err = ct2.Scan(now)
		if err != nil {
			t.Errorf("CustomTime Scan(time.Time) failed: %v", err)
		}
		if !ct2.Time.Equal(now) {
			t.Errorf("Expected time %v, got %v", now, ct2.Time)
		}

		// 测试 RFC3339 格式字符串
		rfcTime := "2023-01-01T12:00:00Z"
		var ct3 CustomTime
		err = ct3.Scan(rfcTime)
		if err != nil {
			t.Errorf("CustomTime Scan(RFC3339) failed: %v", err)
		}

		// 测试 SQLite 格式字符串
		sqliteTime := "2023-01-01 12:00:00"
		var ct4 CustomTime
		err = ct4.Scan(sqliteTime)
		if err != nil {
			t.Errorf("CustomTime Scan(SQLite format) failed: %v", err)
		}

		// 测试无效字符串
		var ct5 CustomTime
		err = ct5.Scan("invalid-time")
		if err == nil {
			t.Error("Expected error for invalid time format")
		}
	})

	// 测试 Value 方法
	t.Run("Value Method", func(t *testing.T) {
		// 测试零值
		ct := CustomTime{}
		value, err := ct.Value()
		if err != nil {
			t.Errorf("CustomTime Value() for zero time failed: %v", err)
		}
		if value != nil {
			t.Error("Expected nil value for zero time")
		}

		// 测试非零值
		now := time.Now()
		ct2 := CustomTime{now}
		value2, err := ct2.Value()
		if err != nil {
			t.Errorf("CustomTime Value() for non-zero time failed: %v", err)
		}
		if value2 != now.Format(time.RFC3339) {
			t.Errorf("Expected RFC3339 format, got %v", value2)
		}
	})
}

// TestJSONSlice 测试 JSON 切片类型
func TestJSONSlice(t *testing.T) {
	// 测试 Scan 方法
	t.Run("Scan Methods", func(t *testing.T) {
		// 测试 nil 值
		var js JSONSlice
		err := js.Scan(nil)
		if err != nil {
			t.Errorf("JSONSlice Scan(nil) failed: %v", err)
		}
		if js != nil {
			t.Error("Expected nil slice for nil input")
		}

		// 测试有效 JSON
		jsonData := `["item1", "item2", "item3"]`
		var js2 JSONSlice
		err = js2.Scan([]byte(jsonData))
		if err != nil {
			t.Errorf("JSONSlice Scan() failed: %v", err)
		}
		if len(js2) != 3 {
			t.Errorf("Expected 3 items, got %d", len(js2))
		}
		if js2[0] != "item1" || js2[1] != "item2" || js2[2] != "item3" {
			t.Errorf("Expected [item1, item2, item3], got %v", js2)
		}

		// 测试无效 JSON
		var js3 JSONSlice
		err = js3.Scan([]byte("invalid json"))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	// 测试 Value 方法
	t.Run("Value Method", func(t *testing.T) {
		// 测试空切片
		js := JSONSlice{}
		value, err := js.Value()
		if err != nil {
			t.Errorf("JSONSlice Value() for empty slice failed: %v", err)
		}
		if value != nil {
			t.Error("Expected nil value for empty slice")
		}

		// 测试非空切片
		js2 := JSONSlice{"go", "database", "testing"}
		value2, err := js2.Value()
		if err != nil {
			t.Errorf("JSONSlice Value() for non-empty slice failed: %v", err)
		}
		expected, _ := json.Marshal(js2)
		if string(value2.([]byte)) != string(expected) {
			t.Errorf("Expected %s, got %s", string(expected), string(value2.([]byte)))
		}
	})
}

// TestJSONMap 测试 JSON 映射类型
func TestJSONMap(t *testing.T) {
	// 测试 Scan 方法
	t.Run("Scan Methods", func(t *testing.T) {
		// 测试 nil 值
		var jm JSONMap
		err := jm.Scan(nil)
		if err != nil {
			t.Errorf("JSONMap Scan(nil) failed: %v", err)
		}
		if jm != nil {
			t.Error("Expected nil map for nil input")
		}

		// 测试有效 JSON
		jsonData := `{"key1": "value1", "key2": 42, "key3": true}`
		var jm2 JSONMap
		err = jm2.Scan([]byte(jsonData))
		if err != nil {
			t.Errorf("JSONMap Scan() failed: %v", err)
		}
		if len(jm2) != 3 {
			t.Errorf("Expected 3 items, got %d", len(jm2))
		}
		if jm2["key1"] != "value1" || jm2["key2"] != 42 {
			t.Errorf("Expected specific values, got %v", jm2)
		}

		// 测试无效 JSON
		var jm3 JSONMap
		err = jm3.Scan([]byte("invalid json"))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	// 测试 Value 方法
	t.Run("Value Method", func(t *testing.T) {
		// 测试空映射
		jm := JSONMap{}
		value, err := jm.Value()
		if err != nil {
			t.Errorf("JSONMap Value() for empty map failed: %v", err)
		}
		if value != nil {
			t.Error("Expected nil value for empty map")
		}

		// 测试非空映射
		jm2 := JSONMap{"key1": "value1", "key2": 42}
		value2, err := jm2.Value()
		if err != nil {
			t.Errorf("JSONMap Value() for non-empty map failed: %v", err)
		}
		expected, _ := json.Marshal(jm2)
		if string(value2.([]byte)) != string(expected) {
			t.Errorf("Expected %s, got %s", string(expected), string(value2.([]byte)))
		}
	})
}

// TestReactionCount 测试反应统计类型
func TestReactionCount(t *testing.T) {
	// 测试 Scan 方法
	t.Run("Scan Methods", func(t *testing.T) {
		// 测试 nil 值
		var rc ReactionCount
		err := rc.Scan(nil)
		if err != nil {
			t.Errorf("ReactionCount Scan(nil) failed: %v", err)
		}
		if rc.Total != 0 {
			t.Error("Expected zero total for nil input")
		}

		// 测试有效 JSON
		jsonData := `{"+1": 5, "laugh": 3, "heart": 2, "total": 10}`
		var rc2 ReactionCount
		err = rc2.Scan([]byte(jsonData))
		if err != nil {
			t.Errorf("ReactionCount Scan() failed: %v", err)
		}
		if rc2.PlusOne != 5 || rc2 Laugh != 3 || rc2.Total != 10 {
			t.Errorf("Expected specific values, got %+v", rc2)
		}

		// 测试无效 JSON
		var rc3 ReactionCount
		err = rc3.Scan([]byte("invalid json"))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	// 测试 Value 方法
	t.Run("Value Method", func(t *testing.T) {
		rc := ReactionCount{
			PlusOne:  5,
			Laugh:    3,
			Heart:    2,
			Total:    10,
		}
		value, err := rc.Value()
		if err != nil {
			t.Errorf("ReactionCount Value() failed: %v", err)
		}
		expected, _ := json.Marshal(rc)
		if string(value.([]byte)) != string(expected) {
			t.Errorf("Expected %s, got %s", string(expected), string(value.([]byte)))
		}
	})
}

// TestRepository 模型测试
func TestRepository(t *testing.T) {
	repo := CreateTestRepository()

	t.Run("TableName", func(t *testing.T) {
		if repo.TableName() != "repositories" {
			t.Errorf("Expected table name 'repositories', got '%s'", repo.TableName())
		}
	})

	t.Run("String Method", func(t *testing.T) {
		expected := "Repository(test/repo)"
		if repo.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, repo.String())
		}
	})

	t.Run("GetFullName", func(t *testing.T) {
		if repo.GetFullName() != "test/repo" {
			t.Errorf("Expected full name 'test/repo', got '%s'", repo.GetFullName())
		}
	})

	t.Run("IsStale", func(t *testing.T) {
		// 测试新创建的仓库（刚刚爬取）
		repo2 := CreateTestRepository()
		repo2.LastScraped = time.Now()
		if repo2.IsStale() {
			t.Error("Expected fresh repository to not be stale")
		}

		// 测试超过7天的仓库
		repo3 := CreateTestRepository()
		repo3.LastScraped = time.Now().Add(-8 * 24 * time.Hour)
		if !repo3.IsStale() {
			t.Error("Expected old repository to be stale")
		}

		// 测试从未爬取的仓库
		repo4 := &Repository{}
		if !repo4.IsStale() {
			t.Error("Expected never-scraped repository to be stale")
		}
	})
}

// TestIssue 模型测试
func TestIssue(t *testing.T) {
	issue := CreateTestIssue()

	t.Run("TableName", func(t *testing.T) {
		if issue.TableName() != "issues" {
			t.Errorf("Expected table name 'issues', got '%s'", issue.TableName())
		}
	})

	t.Run("String Method", func(t *testing.T) {
		expected := "Issue(#1: Test Issue Title)"
		if issue.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, issue.String())
		}
	})

	t.Run("State Check Methods", func(t *testing.T) {
		// 测试 open 状态
		issue.State = "open"
		if !issue.IsOpen() {
			t.Error("Expected open issue to return true for IsOpen()")
		}
		if issue.IsClosed() {
			t.Error("Expected open issue to return false for IsClosed()")
		}

		// 测试 closed 状态
		issue.State = "closed"
		if issue.IsOpen() {
			t.Error("Expected closed issue to return false for IsOpen()")
		}
		if !issue.IsClosed() {
			t.Error("Expected closed issue to return true for IsClosed()")
		}
	})

	t.Run("Severity Check Methods", func(t *testing.T) {
		// 测试高严重程度
		issue.SeverityScore = 8.5
		if !issue.IsHighSeverity() {
			t.Error("Expected high severity issue to return true")
		}

		// 测试低严重程度
		issue.SeverityScore = 5.0
		if issue.IsHighSeverity() {
			t.Error("Expected low severity issue to return false")
		}
	})

	t.Run("Score Check Methods", func(t *testing.T) {
		// 测试高分
		issue.Score = 9.5
		if !issue.IsHighScore() {
			t.Error("Expected high score issue to return true")
		}

		// 测试低分
		issue.Score = 7.0
		if issue.IsHighScore() {
			t.Error("Expected low score issue to return false")
		}
	})

	t.Run("Label Methods", func(t *testing.T) {
		// 测试包含标签
		issue.Labels = StringArray{"bug", "performance", "critical"}
		if !issue.HasLabels("bug") {
			t.Error("Expected issue to contain 'bug' label")
		}
		if !issue.HasLabels("bug", "performance") {
			t.Error("Expected issue to contain 'bug' and 'performance' labels")
		}
		if issue.HasLabels("nonexistent") {
			t.Error("Expected issue to not contain 'nonexistent' label")
		}

		// 测试空标签列表
		if !issue.HasLabels() {
			t.Error("Expected issue with labels to return true for empty label list")
		}

		// 测试 GetLabelString
		expectedLabels := "bug, performance, critical"
		if issue.GetLabelString() != expectedLabels {
			t.Errorf("Expected '%s', got '%s'", expectedLabels, issue.GetLabelString())
		}
	})

	t.Run("Assignee Methods", func(t *testing.T) {
		issue.Assignees = StringArray{"user1", "user2"}
		expectedAssignees := "user1, user2"
		if issue.GetAssigneeString() != expectedAssignees {
			t.Errorf("Expected '%s', got '%s'", expectedAssignees, issue.GetAssigneeString())
		}
	})

	t.Run("Age Methods", func(t *testing.T) {
		// 测试有创建时间的情况
		issue.CreatedAt = CustomTime{time.Now().Add(-48 * time.Hour)}
		days := issue.GetAgeInDays()
		if days < 1 || days > 2 {
			t.Errorf("Expected age between 1-2 days, got %d", days)
		}

		// 测试零时间的情况
		issue.CreatedAt = CustomTime{}
		days = issue.GetAgeInDays()
		if days != 0 {
			t.Errorf("Expected age 0 for zero time, got %d", days)
		}
	})

	t.Run("Content Hash", func(t *testing.T) {
		// 测试已有 hash 的情况
		issue.ContentHash = "existing_hash"
		if issue.GetContentHash() != "existing_hash" {
			t.Error("Expected existing hash to be returned")
		}

		// 测试生成 hash 的情况
		issue.ContentHash = ""
		hash := issue.GetContentHash()
		if hash == "" {
			t.Error("Expected non-empty generated hash")
		}

		// 测试相同内容生成相同 hash
		issue2 := CreateTestIssue()
		issue2.ContentHash = ""
		if issue.GetContentHash() != issue2.GetContentHash() {
			t.Error("Expected same content to generate same hash")
		}
	})

	t.Run("Similarity", func(t *testing.T) {
		// 测试相同的问题
		issue1 := CreateTestIssue()
		issue2 := CreateTestIssue()
		similarity := issue1.IsSimilar(*issue2)
		if similarity != 1.0 {
			t.Errorf("Expected similarity 1.0 for identical issues, got %f", similarity)
		}

		// 测试完全不同的问题
		issue1.Title = "Memory leak issue"
		issue1.Body = "There is a memory leak"
		issue1.RepoOwner = "owner1"
		issue1.RepoName = "repo1"
		
		issue2.Title = "Add new feature"
		issue2.Body = "Please add a new feature"
		issue2.RepoOwner = "owner2"
		issue2.RepoName = "repo2"

		similarity = issue1.IsSimilar(*issue2)
		if similarity >= 0.8 {
			t.Errorf("Expected low similarity for different issues, got %f", similarity)
		}
	})
}

// TestTimeSeries 模型测试
func TestTimeSeries(t *testing.T) {
	ts := &TimeSeries{
		ID:                 1,
		RepositoryID:       1,
		Date:               time.Now(),
		Year:               2023,
		Month:              1,
		Day:                1,
		NewIssuesCount:     5,
		ClosedIssuesCount:  3,
		ActiveIssuesCount:  2,
		PitfallIssuesCount: 1,
		AvgSeverityScore:   7.5,
		TotalComments:      25,
	}

	t.Run("TableName", func(t *testing.T) {
		if ts.TableName() != "time_series" {
			t.Errorf("Expected table name 'time_series', got '%s'", ts.TableName())
		}
	})

	t.Run("String Method", func(t *testing.T) {
		expected := "TimeSeries(" + ts.Date.Format("2006-01-02") + ")"
		if ts.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, ts.String())
		}
	})

	t.Run("Date Check Methods", func(t *testing.T) {
		// 测试今天的数据
		now := time.Now()
		tsToday := &TimeSeries{Date: now}
		if !tsToday.IsToday() {
			t.Error("Expected today's data to return true")
		}

		// 测试不是今天的数据
		tsYesterday := &TimeSeries{Date: now.Add(-24 * time.Hour)}
		if tsYesterday.IsToday() {
			t.Error("Expected yesterday's data to return false")
		}

		// 测试本周的数据
		tsThisWeek := &TimeSeries{Date: now}
		if !tsThisWeek.IsThisWeek() {
			t.Error("Expected this week's data to return true")
		}

		// 测试本月的数据
		tsThisMonth := &TimeSeries{Date: now}
		if !tsThisMonth.IsThisMonth() {
			t.Error("Expected this month's data to return true")
		}
	})

	t.Run("Activity Score", func(t *testing.T) {
		ts := &TimeSeries{
			NewIssuesCount:     10,
			ClosedIssuesCount:  5,
			PitfallIssuesCount: 2,
		}

		expectedScore := 10.0*1.0 + 5.0*1.2 + 2.0*2.0 // 10 + 6 + 4 = 20
		score := ts.GetActivityScore()
		if score != expectedScore {
			t.Errorf("Expected activity score %f, got %f", expectedScore, score)
		}
	})
}

// TestCategory 模型测试
func TestCategory(t *testing.T) {
	category := &Category{
		ID:          1,
		Name:        "Performance",
		Description: "Performance related issues",
		Color:       "#ff0000",
		IsActive:    true,
		Priority:    85,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("TableName", func(t *testing.T) {
		if category.TableName() != "categories" {
			t.Errorf("Expected table name 'categories', got '%s'", category.TableName())
		}
	})

	t.Run("String Method", func(t *testing.T) {
		expected := "Category(Performance)"
		if category.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, category.String())
		}
	})

	t.Run("IsHighPriority", func(t *testing.T) {
		// 测试高优先级
		category.Priority = 85
		if !category.IsHighPriority() {
			t.Error("Expected priority 85 to be high priority")
		}

		// 测试低优先级
		category.Priority = 75
		if category.IsHighPriority() {
			t.Error("Expected priority 75 to not be high priority")
		}

		// 测试边界值
		category.Priority = 80
		if !category.IsHighPriority() {
			t.Error("Expected priority 80 to be high priority (boundary)")
		}

		category.Priority = 79
		if category.IsHighPriority() {
			t.Error("Expected priority 79 to not be high priority (boundary)")
		}
	})
}

// TestDatabaseConfig 配置测试
func TestDatabaseConfig(t *testing.T) {
	t.Run("Validate Valid Config", func(t *testing.T) {
		config := DefaultDatabaseConfig()
		config.Path = "/valid/path.db"
		config.MaxConnections = 10
		config.Timeout = 30 * time.Second

		if err := config.Validate(); err != nil {
			t.Errorf("Valid config should not produce error: %v", err)
		}
	})

	t.Run("Validate Invalid Config", func(t *testing.T) {
		// 测试空路径
		config := DatabaseConfig{
			Path: "",
		}
		if err := config.Validate(); err == nil {
			t.Error("Empty path should produce validation error")
		}

		// 测试无效连接数
		config = DatabaseConfig{
			Path:           "/valid/path.db",
			MaxConnections: 0,
		}
		if err := == nil {
			t.Error("Zero max connections should produce config.Validate(); err	}

		// 测试无效超时
		config = validation error")
		Path:   "/valid/path.db",
			Timeout: 0,
		}
		if err := config.Validate(); err == nil {
			t.Error("Zero error")
		}
	})

	t timeout should produce validation DatabaseConfig{
		.Run("Default Database Config", func(t *testing.T) {
		config := DefaultDatabaseConfig()

		if config.Path == "" {
			t.Error("Default path should not be empty")
		}
		if config.MaxConnections <= 0 {
			t.Error("Default max connections should be positive")
		}
		if config.Timeout <= 0 be positive")
		}
		if("Default timeout should {
			t.Error !config.EnableWAL {
			t.Error("Default WAL mode should be enabled")
		}
.EnableForeignKeys {
			t.Error("Default foreign keys should be enabled")
		}
	})
		if !config}

// TestIssueStats 统计信息测试
func TestIssueStats(t *testing.T) {
	stats := &IssueStats{
		TotalCount:        100,
		ByCategory:        map[string]int{"bug": 30, "feature": 20},
		ByPriority:        map[string]int{"high": 15, "medium": 25},
		ByTechStack:       map[string]int{"go": 40, "js": 35State:           map},
		By[string]int{"open": 60, "closed": 40},
		ScoreDistribution: map[string]int": 20,{"high "medium": ": 30},
		AverageScore:      7.5,
	50, "low	TopKeywords:       map[string]int{"memory": 15, "performance": 25},
		DateRange:         map[string]interface{}{"min": "2023- "max": "01-01",2023-12-31"},
	}

	t.Run("TotalCount", func(t *testing.T) {
		if stats.TotalCount != 100 {
			t.Errorf("Expected total count d", stats.TotalCount)
		}
	})

	t.Run("Category Distribution", func(t *testing.T) {
		if stats.ByCategory["bug"]100, got % != 30 {
			t.Errorf("Expected bug count 30, got %d", stats.ByCategory["bug"])
		}
		if stats.ByCategory["feature"] != 20 {
			t.Errorf("Expected feature count 20, got %d", stats.ByCategory["feature"])
		}
	})

	t.Run("Average Score", func(t *testing.T) {
		if stats.AverageScore != 7.5 {
			t.Errorf("Expected average score 7.5, got %f", stats.AverageScore)
		}
	})
}

// TestHelperFunctions 辅助函数测试
func TestHelperFunctions(t *testing.T) {
	t.Run("String Similar(t *testing.T) {
		// 测试完全相同的ity Calculation", func字符串
		sim := calculateStringSimilarity("hello world")
		if world", "hello sim != 1.0 {
			t.Errorf("Expected similarity 1.0 for identical strings, got %f", sim)
		}

		// 测试完全不同
		sim = calculateStringSimilarity("hello", "world")
		if sim > 0.5 {
			t.Errorf("Expected low similarity for different strings, got %f", sim)
		}

		// 测试包含相同单词
		Similarity("hellosim = calculateString world test", "hello world example")
		if sim < 0.3 {
			t.Errorf("Expected medium similarity for partially matching strings, got %f", sim)
		}

		// 测试空字符串
		sim = calculateStringSimilarity("", "test")
		if sim != 0.0 {
			t.Errorf("Expected similarity 0.0 for empty string, got %f", sim)
		}

		// 测试大小写不敏感
		sim = calculateStringSimilarity("HELLO", "hello")
		if sim != 1.0 {
			t.Errorf("Expected similarity 1.0 for case-insensitive match, got %f", sim)
		}
	})

	t.Run("String Slice Intersection", func(t *testing.T) {
		s1 := JSONSlice{"a", "b", "c", "d"}
		s2 := JSONSlice{"b", "c", "e", "f"}
		result := intersectStringSlices(s1, s2)
		expected := JSONSlice{"b", "c"}

		if len(result) != len(expected) {
			t.Errorf("Expected intersection length %d, got %d", len(expected), len(result))
		}

		for _, item := range expected {
			found := false
			for _, r := range result {
				if r == item {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected item '%s' in intersection", item)
			}
		}

		// 测试空切片
		result = intersectStringSlices(JSONSlice{}, s2)
		if len(result) != 0 {
			t.Error("Expected empty intersection for empty first slice")
		}

		result = intersectStringSlices(s1, JSONSlice{})
		if len(result) != 0 {
			t.Error("Expected empty intersection for empty second slice")
		}
	})

	t.Run("String Slice Union", func(t *testing.T) {
		s1 := JSONSlice{"a", "b", "c"}
		s2 := JSONSlice{"b", "c", "d", "e"}
		result := unionStringSlices(s1, s2)
		expected := JSONSlice{"a", "b", "c", "d", "e"}

		if len(result) != len(expected) {
			t.Errorf("Expected union length %d, got %d", len(expected), len(result))
		}

		// 检查去重
		seen := make(map[string]bool)
		for _, item := range result {
			if seen[item] {
				t.Errorf("Duplicate item '%s' in union result", item)
			}
			seen[item] = true
		}

		// 检查包含所有元素
		for _, item := range expected {
			found := false
			for _, r := range result {
				if r == item {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected item '%s' in union result", item)
			}
		}

		// 测试空切片
		result = unionStringSlices(JSONSlice{}, s2)
		if len(result) != len(s2) {
			t.Error("Expected union with empty slice to return original slice")
		}
	})

	t.Run("Min Function", func(t *testing.T) {
		if min(1, 2) != 1 {
			t.Error("Expected min(1, 2) to return 1")
		}
		if min(2, 1) != 1 {
			t.Error("Expected min(2, 1) to return 1")
		}
		if min(5, 5) != 5 {
			t.Error("Expected min(5, 5) to return 5")
		}
		if min(-1, 1) != -1 {
			t.Error("Expected min(-1, 1) to return -1")
		}
	})
}