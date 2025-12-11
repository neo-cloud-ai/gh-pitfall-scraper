package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CustomTime 自定义时间类型，适配 SQLite 的 DATETIME
type CustomTime struct {
	time.Time
}

// Scan 实现 sql.Scanner 接口
func (t *CustomTime) Scan(value interface{}) error {
	if value == nil {
		*t = CustomTime{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = CustomTime{v}
	case string:
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			*t = CustomTime{parsed}
		} else if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			*t = CustomTime{parsed}
		} else {
			return fmt.Errorf("无法解析时间: %v", v)
		}
	default:
		return fmt.Errorf("不支持的时间类型: %T", v)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (t CustomTime) Value() (interface{}, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time.Format(time.RFC3339), nil
}

// JSONSlice 通用 JSON 切片类型
type JSONSlice []string

// Scan 实现 sql.Scanner 接口
func (js *JSONSlice) Scan(value interface{}) error {
	if value == nil {
		*js = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var result []string
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*js = result
	return nil
}

// Value 实现 driver.Valuer 接口
func (js JSONSlice) Value() (interface{}, error) {
	if len(js) == 0 {
		return nil, nil
	}
	return json.Marshal(js)
}

// JSONMap 通用 JSON 映射类型
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (jm *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*jm = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*jm = result
	return nil
}

// Value 实现 driver.Valuer 接口
func (jm JSONMap) Value() (interface{}, error) {
	if len(jm) == 0 {
		return nil, nil
	}
	return json.Marshal(jm)
}

// ReactionCount 反应统计数据
type ReactionCount struct {
	PlusOne    int `json:"+1,omitempty"`
	MinusOne   int `json:"-1,omitempty"`
	Laugh      int `json:"laugh,omitempty"`
	Confused   int `json:"confused,omitempty"`
	Heart      int `json:"heart,omitempty"`
	Hooray     int `json:"hooray,omitempty"`
	Rocket     int `json:"rocket,omitempty"`
	Eyes       int `json:"eyes,omitempty"`
	Total      int `json:"total,omitempty"`
}

// Scan 实现 sql.Scanner 接口
func (rc *ReactionCount) Scan(value interface{}) error {
	if value == nil {
		*rc = ReactionCount{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var result ReactionCount
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*rc = result
	return nil
}

// Value 实现 driver.Valuer 接口
func (rc ReactionCount) Value() (interface{}, error) {
	return json.Marshal(rc)
}

// Repository 仓库模型
type Repository struct {
	ID              int           `json:"id" db:"id"`
	Owner           string        `json:"owner" db:"owner"`
	Name            string        `json:"name" db:"name"`
	FullName        string        `json:"full_name" db:"full_name" binding:"required"`
	Description     string        `json:"description" db:"description"`
	URL             string        `json:"url" db:"url" binding:"required"`
	Stars           int           `json:"stars" db:"stars"`
	Forks           int           `json:"forks" db:"forks"`
	IssuesCount     int           `json:"issues_count" db:"issues_count"`
	Language        string        `json:"language" db:"language"`
	CreatedAt       CustomTime    `json:"created_at" db:"created_at"`
	UpdatedAt       CustomTime    `json:"updated_at" db:"updated_at"`
	LastScrapedAt   CustomTime    `json:"last_scraped_at" db:"last_scraped_at"`
	IsActive        bool          `json:"is_active" db:"is_active"`
	Metadata        JSONMap       `json:"metadata" db:"metadata"`
}

// TableName 返回表名
func (Repository) TableName() string {
	return "repositories"
}

// String 返回字符串表示
func (r Repository) String() string {
	return fmt.Sprintf("Repository(%s)", r.FullName)
}

// GetFullName 返回完整仓库名称
func (r Repository) GetFullName() string {
	return r.FullName
}

// IsStale 检查仓库是否需要重新爬取（超过7天未爬取）
func (r Repository) IsStale() bool {
	if r.LastScrapedAt.IsZero() {
		return true
	}
	return time.Since(r.LastScrapedAt.Time) > 7*24*time.Hour
}

// RepositoryStats 仓库统计信息
type RepositoryStats struct {
	Repository
	TotalIssues      int     `json:"total_issues" db:"total_issues"`
	OpenIssues       int     `json:"open_issues" db:"open_issues"`
	ClosedIssues     int     `json:"closed_issues" db:"closed_issues"`
	PitfallIssues    int     `json:"pitfall_issues" db:"pitfall_issues"`
	DuplicateIssues  int     `json:"duplicate_issues" db:"duplicate_issues"`
	AvgSeverityScore float64 `json:"avg_severity_score" db:"avg_severity_score"`
	AvgScore         float64 `json:"avg_score" db:"avg_score"`
	LastIssueUpdate  *CustomTime `json:"last_issue_update" db:"last_issue_update"`
}

// TableName 返回表名
func (RepositoryStats) TableName() string {
	return "repository_stats"
}

// Category 分类模型
type Category struct {
	ID          int           `json:"id" db:"id"`
	Name        string        `json:"name" db:"name" binding:"required"`
	Description string        `json:"description" db:"description"`
	Color       string        `json:"color" db:"color"`
	IsActive    bool          `json:"is_active" db:"is_active"`
	Priority    int           `json:"priority" db:"priority"`
	CreatedAt   CustomTime    `json:"created_at" db:"created_at"`
	UpdatedAt   CustomTime    `json:"updated_at" db:"updated_at"`
}

// TableName 返回表名
func (Category) TableName() string {
	return "categories"
}

// String 返回字符串表示
func (c Category) String() string {
	return fmt.Sprintf("Category(%s)", c.Name)
}

// IsHighPriority 检查是否为高优先级分类
func (c Category) IsHighPriority() bool {
	return c.Priority >= 80
}

// Issue Issue 模型 - 扩展现有结构
type Issue struct {
	ID             int            `json:"id" db:"id"`
	IssueID        int64          `json:"issue_id" db:"issue_id" binding:"required"`
	RepositoryID   int            `json:"repository_id" db:"repository_id" binding:"required"`
	Number         int            `json:"number" db:"number" binding:"required"`
	Title          string         `json:"title" db:"title" binding:"required"`
	Body           string         `json:"body" db:"body"`
	State          string         `json:"state" db:"state" binding:"required,oneof=open closed"`
	AuthorLogin    string         `json:"author_login" db:"author_login" binding:"required"`
	AuthorType     string         `json:"author_type" db:"author_type"`
	Labels         JSONSlice      `json:"labels" db:"labels"`
	Assignees      JSONSlice      `json:"assignees" db:"assignees"`
	Milestone      string         `json:"milestone" db:"milestone"`
	Reactions      ReactionCount  `json:"reactions" db:"reactions"`
	CreatedAt      CustomTime     `json:"created_at" db:"created_at"`
	UpdatedAt      CustomTime     `json:"updated_at" db:"updated_at"`
	ClosedAt       *CustomTime    `json:"closed_at" db:"closed_at"`
	FirstSeenAt    CustomTime     `json:"first_seen_at" db:"first_seen_at"`
	LastSeenAt     CustomTime     `json:"last_seen_at" db:"last_seen_at"`
	IsPitfall      bool           `json:"is_pitfall" db:"is_pitfall"`
	SeverityScore  float64        `json:"severity_score" db:"severity_score"`
	CategoryID     *int           `json:"category_id" db:"category_id"`
	Score          float64        `json:"score" db:"score"`
	URL            string         `json:"url" db:"url" binding:"required"`
	HTMLURL        string         `json:"html_url" db:"html_url" binding:"required"`
	CommentsCount  int            `json:"comments_count" db:"comments_count"`
	IsDuplicate    bool           `json:"is_duplicate" db:"is_duplicate"`
	DuplicateOf    *int64         `json:"duplicate_of" db:"duplicate_of"`
	Metadata       JSONMap        `json:"metadata" db:"metadata"`
	
	// 兼容现有结构
	ReactionsOld   int            `json:"reactions" db:"reactions"`
	CommentsOld    int            `json:"comments" db:"comments"`
	Assignee       string         `json:"assignee" db:"assignee"`
	RepoOwner      string         `json:"repo_owner" db:"repo_owner"`
	RepoName       string         `json:"repo_name" db:"repo_name"`
	Keywords       JSONSlice      `json:"keywords" db:"keywords"`
	ContentHash    string         `json:"content_hash" db:"content_hash"`
	CategoryOld    string         `json:"category" db:"category"`
	Priority       string         `json:"priority" db:"priority"`
	TechStack      JSONSlice      `json:"tech_stack" db:"tech_stack"`
	IsDuplicateOld bool           `json:"is_duplicate" db:"is_duplicate"`
	DuplicateOfOld sql.NullInt64  `json:"duplicate_of" db:"duplicate_of"`
	CreatedAtDB    CustomTime     `json:"created_at_db" db:"created_at_db"`
	UpdatedAtDB    CustomTime     `json:"updated_at_db" db:"updated_at_db"`
}

// TableName 返回表名
func (Issue) TableName() string {
	return "issues"
}

// String 返回字符串表示
func (i Issue) String() string {
	return fmt.Sprintf("Issue(#%d: %s)", i.Number, i.Title)
}

// IsOpen 检查Issue是否处于打开状态
func (i Issue) IsOpen() bool {
	return i.State == "open"
}

// IsClosed 检查Issue是否处于关闭状态
func (i Issue) IsClosed() bool {
	return i.State == "closed"
}

// IsHighSeverity 检查严重程度是否较高
func (i Issue) IsHighSeverity() bool {
	return i.SeverityScore >= 7.0
}

// IsHighScore 检查综合评分是否较高
func (i Issue) IsHighScore() bool {
	return i.Score >= 8.0
}

// HasLabels 检查是否包含指定标签
func (i Issue) HasLabels(labels ...string) bool {
	if len(labels) == 0 {
		return len(i.Labels) > 0
	}
	
	for _, label := range labels {
		for _, issueLabel := range i.Labels {
			if strings.EqualFold(label, issueLabel) {
				return true
			}
		}
	}
	return false
}

// GetLabelString 返回标签的字符串表示
func (i Issue) GetLabelString() string {
	return strings.Join(i.Labels, ", ")
}

// GetAssigneeString 返回指派人的字符串表示
func (i Issue) GetAssigneeString() string {
	return strings.Join(i.Assignees, ", ")
}

// GetAgeInDays 返回Issue存在天数
func (i Issue) GetAgeInDays() int {
	if i.CreatedAt.IsZero() {
		return 0
	}
	return int(time.Since(i.CreatedAt.Time).Hours() / 24)
}

// GetDaysSinceUpdate 返回距离最后更新的天数
func (i Issue) GetDaysSinceUpdate() int {
	if i.UpdatedAt.IsZero() {
		return 0
	}
	return int(time.Since(i.UpdatedAt.Time).Hours() / 24)
}

// GetContentHash 生成内容哈希值
func (i Issue) GetContentHash() string {
	// 如果已有content_hash，直接返回
	if i.ContentHash != "" {
		return i.ContentHash
	}
	
	// 创建规范化内容字符串
	content := strings.Join([]string{
		i.Title,
		i.Body,
		i.RepoOwner,
		i.RepoName,
	}, "\n")
	
	// 简单哈希函数
	hash := 0
	for _, ch := range content {
		hash = hash*31 + int(ch)
	}
	return strconv.Itoa(hash)
}

// IsSimilar 检查与另一个Issue的相似度
func (i Issue) IsSimilar(other Issue) float64 {
	similarity := 0.0
	
	// 标题相似度 (40%权重)
	titleSim := calculateStringSimilarity(i.Title, other.Title)
	similarity += titleSim * 0.4
	
	// 内容相似度 (30%权重)
	bodySim := calculateStringSimilarity(i.Body, other.Body)
	similarity += bodySim * 0.3
	
	// 仓库相似度 (20%权重)
	if i.RepoOwner == other.RepoOwner && i.RepoName == other.RepoName {
		similarity += 0.2
	}
	
	// 技术栈相似度 (10%权重)
	commonTechStacks := len(intersectStringSlices(i.TechStack, other.TechStack))
	totalTechStacks := len(unionStringSlices(i.TechStack, other.TechStack))
	if totalTechStacks > 0 {
		techSim := float64(commonTechStacks) / float64(totalTechStacks)
		similarity += techSim * 0.1
	}
	
	return similarity
}

// TimeSeries 时间序列数据模型
type TimeSeries struct {
	ID                 int           `json:"id" db:"id"`
	RepositoryID       int           `json:"repository_id" db:"repository_id" binding:"required"`
	Date               time.Time     `json:"date" db:"date" binding:"required"`
	Year               int           `json:"year" db:"year"`
	Month              int           `json:"month" db:"month"`
	Day                int           `json:"day" db:"day"`
	WeekOfYear         int           `json:"week_of_year" db:"week_of_year"`
	DayOfWeek          int           `json:"day_of_week" db:"day_of_week"`
	NewIssuesCount     int           `json:"new_issues_count" db:"new_issues_count"`
	ClosedIssuesCount  int           `json:"closed_issues_count" db:"closed_issues_count"`
	ActiveIssuesCount  int           `json:"active_issues_count" db:"active_issues_count"`
	PitfallIssuesCount int           `json:"pitfall_issues_count" db:"pitfall_issues_count"`
	AvgSeverityScore   float64       `json:"avg_severity_score" db:"avg_severity_score"`
	TotalComments      int           `json:"total_comments" db:"total_comments"`
	CreatedAt          CustomTime    `json:"created_at" db:"created_at"`
	UpdatedAt          CustomTime    `json:"updated_at" db:"updated_at"`
	
	// 关联对象（用于 JOIN 查询）
	Repository *Repository `json:"repository,omitempty" db:"-"`
}

// TableName 返回表名
func (TimeSeries) TableName() string {
	return "time_series"
}

// String 返回字符串表示
func (ts TimeSeries) String() string {
	return fmt.Sprintf("TimeSeries(%s)", ts.Date.Format("2006-01-02"))
}

// IsToday 检查是否为今天的数据
func (ts TimeSeries) IsToday() bool {
	now := time.Now()
	return ts.Date.Year() == now.Year() && 
		   ts.Date.Month() == now.Month() && 
		   ts.Date.Day() == now.Day()
}

// IsThisWeek 检查是否为本周的数据
func (ts TimeSeries) IsThisWeek() bool {
	_, week := time.Now().ISOWeek()
	_, tsWeek := ts.Date.ISOWeek()
	return ts.Date.Year() == time.Now().Year() && week == tsWeek
}

// IsThisMonth 检查是否为本月的数据
func (ts TimeSeries) IsThisMonth() bool {
	now := time.Now()
	return ts.Date.Year() == now.Year() && ts.Date.Month() == now.Month()
}

// GetActivityScore 计算活动分数
func (ts TimeSeries) GetActivityScore() float64 {
	// 简单算法：新Issues * 1 + 关闭Issues * 1.2 + 坑点Issues * 2
	score := float64(ts.NewIssuesCount) * 1.0 + 
			 float64(ts.ClosedIssuesCount) * 1.2 + 
			 float64(ts.PitfallIssuesCount) * 2.0
	return score
}

// 兼容现有结构体
type IssueFilter struct {
	ID          sql.NullInt64  `db:"id"`
	RepoOwner   sql.NullString `db:"repo_owner"`
	RepoName    sql.NullString `db:"repo_name"`
	Category    sql.NullString `db:"category"`
	Priority    sql.NullString `db:"priority"`
	MinScore    sql.NullFloat64 `db:"min_score"`
	MaxScore    sql.NullFloat64 `db:"max_score"`
	State       sql.NullString `db:"state"`
	TechStack   JSONSlice      `db:"tech_stack"`
	Keywords    JSONSlice      `db:"keywords"`
	IsDuplicate sql.NullBool   `db:"is_duplicate"`
	DateFrom    sql.NullTime   `db:"date_from"`
	DateTo      sql.NullTime   `db:"date_to"`
}

type IssueSort struct {
	Field     string
	Direction string // "ASC" or "DESC"
}

type IssueStats struct {
	TotalCount      int                    `json:"total_count"`
	ByCategory      map[string]int         `json:"by_category"`
	ByPriority      map[string]int         `json:"by_priority"`
	ByTechStack     map[string]int         `json:"by_tech_stack"`
	ByState         map[string]int         `json:"by_state"`
	ScoreDistribution map[string]int       `json:"score_distribution"`
	AverageScore    float64                `json:"average_score"`
	TopKeywords     map[string]int         `json:"top_keywords"`
	DateRange       map[string]interface{} `json:"date_range"`
}

type DeduplicationResult struct {
	TotalProcessed   int                   `json:"total_processed"`
	DuplicatesFound  int                   `json:"duplicates_found"`
	DuplicatesRemoved int                  `json:"duplicates_removed"`
	UniqueIssues     int                   `json:"unique_issues"`
	DuplicateGroups  []DuplicateGroup      `json:"duplicate_groups"`
}

type DuplicateGroup struct {
	MasterIssue  *Issue          `json:"master_issue"`
	Duplicates   []*Issue        `json:"duplicates"`
	Similarity   float64         `json:"similarity"`
	Reason       string          `json:"reason"`
}

type ClassificationResult struct {
	TotalProcessed int                    `json:"total_processed"`
	Classified     int                    `json:"classified"`
	AutoClassified int                    `json:"auto_classified"`
	ManuallyReviewed int                  `json:"manually_reviewed"`
	Confidence     float64                `json:"confidence"`
	ByCategory     map[string]int         `json:"by_category"`
	ByPriority     map[string]int         `json:"by_priority"`
	ByTechStack    map[string]int         `json:"by_tech_stack"`
}

type DatabaseConfig struct {
	Path            string        `json:"path"`
	MaxConnections  int           `json:"max_connections"`
	Timeout         time.Duration `json:"timeout"`
	EnableWAL       bool          `json:"enable_wal"`
	EnableForeignKeys bool        `json:"enable_foreign_keys"`
	BusyTimeout     time.Duration `json:"busy_timeout"`
}

func (c DatabaseConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("database path is required")
	}
	if c.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}

func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Path:            "./data/issues.db",
		MaxConnections:  10,
		Timeout:         30 * time.Second,
		EnableWAL:       true,
		EnableForeignKeys: true,
		BusyTimeout:     30 * time.Second,
	}
}

// Helper functions for string similarity calculation
func calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple Jaccard similarity
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 && len(words2) == 0 {
		return 1.0
	}

	set1 := make(map[string]bool)
	for _, word := range words1 {
		set1[word] = true
	}

	intersection := 0
	union := len(set1)

	for _, word := range words2 {
		if set1[word] {
			intersection++
		} else {
			union++
		}
	}

	return float64(intersection) / float64(union)
}

func intersectStringSlices(s1, s2 JSONSlice) JSONSlice {
	if len(s1) == 0 || len(s2) == 0 {
		return nil
	}

	result := make(JSONSlice, 0, min(len(s1), len(s2)))
	set := make(map[string]bool)

	for _, item := range s1 {
		set[item] = true
	}

	for _, item := range s2 {
		if set[item] {
			result = append(result, item)
		}
	}

	return result
}

func unionStringSlices(s1, s2 JSONSlice) JSONSlice {
	result := make(JSONSlice, 0, len(s1)+len(s2))
	set := make(map[string]bool)

	for _, item := range s1 {
		if !set[item] {
			result = append(result, item)
			set[item] = true
		}
	}

	for _, item := range s2 {
		if !set[item] {
			result = append(result, item)
			set[item] = true
		}
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}