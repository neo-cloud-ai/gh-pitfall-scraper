package database

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// QueryBuilder 查询构建器
type QueryBuilder struct {
	db      *sql.DB
	logger  *log.Logger
	query   string
	params  []interface{}
}

// QueryResult 查询结果
type QueryResult struct {
	Issues      []*Issue         `json:"issues"`
	TotalCount  int              `json:"total_count"`
	HasMore     bool             `json:"has_more"`
	NextOffset  int              `json:"next_offset"`
	QueryTime   time.Duration    `json:"query_time"`
	QueryString string           `json:"query_string"`
}

// SearchCriteria 搜索条件
type SearchCriteria struct {
	// 基础条件
	Query       string      `json:"query"`
	Keywords    []string    `json:"keywords"`
	Categories  []string    `json:"categories"`
	Tags        []string    `json:"tags"`
	Repositories []string   `json:"repositories"`
	States      []string    `json:"states"`
	Authors     []string    `json:"authors"`
	Assignees   []string    `json:"assignees"`
	
	// 时间条件
	DateFrom    *time.Time  `json:"date_from"`
	DateTo      *time.Time  `json:"date_to"`
	AgeMin      *int        `json:"age_min"`
	AgeMax      *int        `json:"age_max"`
	
	// 数值条件
	MinScore    *float64    `json:"min_score"`
	MaxScore    *float64    `json:"max_score"`
	MinSeverity *float64    `json:"min_severity"`
	MaxSeverity *float64    `json:"max_severity"`
	MinComments *int        `json:"min_comments"`
	MaxComments *int        `json:"max_comments"`
	
	// 布尔条件
	IsPitfall   *bool       `json:"is_pitfall"`
	IsDuplicate *bool       `json:"is_duplicate"`
	IsOpen      *bool       `json:"is_open"`
	HasAssignee *bool       `json:"has_assignee"`
	
	// 分页和排序
	Page        int         `json:"page"`
	PageSize    int         `json:"page_size"`
	SortBy      string      `json:"sort_by"`
	SortOrder   string      `json:"sort_order"`
	
	// 高级选项
	ExcludeCategories []string `json:"exclude_categories"`
	ExcludeRepositories []string `json:"exclude_repositories"`
	FuzzyMatch    bool       `json:"fuzzy_match"`
}

// AggregatedQuery 聚合查询
type AggregatedQuery struct {
	GroupBy      string            `json:"group_by"`
	Metrics      []string          `json:"metrics"`
	DateGrouping string            `json:"date_grouping"` // day, week, month, year
	Filters      SearchCriteria    `json:"filters"`
}

// AggregationResult 聚合结果
type AggregationResult struct {
	Groups       []map[string]interface{} `json:"groups"`
	TotalGroups  int                      `json:"total_groups"`
	QueryTime    time.Duration            `json:"query_time"`
}

// TimeSeriesQuery 时间序列查询
type TimeSeriesQuery struct {
	StartDate    time.Time            `json:"start_date"`
	EndDate      time.Time            `json:"end_date"`
	Interval     string               `json:"interval"` // hour, day, week, month
	Metrics      []string             `json:"metrics"`
	Repositories []string             `json:"repositories"`
	Categories   []string             `json:"categories"`
}

// TimeSeriesResult 时间序列结果
type TimeSeriesResult struct {
	Points       []TimeSeriesPoint    `json:"points"`
	TotalPoints  int                  `json:"total_points"`
	QueryTime    time.Duration        `json:"query_time"`
}

// TimeSeriesPoint 时间序列点
type TimeSeriesPoint struct {
	Timestamp   time.Time               `json:"timestamp"`
	Metrics     map[string]float64      `json:"metrics"`
	Repository  string                  `json:"repository,omitempty"`
	Category    string                  `json:"category,omitempty"`
}

// FacetedQuery 分面查询
type FacetedQuery struct {
	BaseCriteria SearchCriteria        `json:"base_criteria"`
	Facets       []string              `json:"facets"` // category, repository, author, tag, etc.
}

// FacetedResult 分面结果
type FacetedResult struct {
	Issues          []*Issue           `json:"issues"`
	Facets          map[string]map[string]int `json:"facets"`
	TotalCount      int                `json:"total_count"`
	QueryTime       time.Duration      `json:"query_time"`
}

// NewQueryBuilder 创建新的查询构建器
func NewQueryBuilder(db *sql.DB) *QueryBuilder {
	return &QueryBuilder{
		db:     db,
		logger: log.New(log.Writer(), "[Query] ", log.LstdFlags),
		query:  "",
		params: make([]interface{}, 0),
	}
}

// SimpleQuery 执行简单查询
func (q *QueryBuilder) SimpleQuery(criteria SearchCriteria) (*QueryResult, error) {
	startTime := time.Now()
	
	// 构建基础查询
	baseQuery := `
		SELECT i.*, r.full_name as repository_name, r.url as repository_url,
			   c.name as category_name, c.description as category_description
		FROM issues i
		LEFT JOIN repositories r ON i.repository_id = r.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE 1=1
	`
	
	var params []interface{}
	paramCount := 0
	
	// 添加WHERE条件
	query, params := q.buildWhereConditions(baseQuery, criteria, &paramCount)
	
	// 添加排序
	orderBy := q.buildOrderBy(criteria)
	query += orderBy
	
	// 添加分页
	offset := (criteria.Page - 1) * criteria.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", criteria.PageSize, offset)
	
	// 执行查询获取数据
	rows, err := q.db.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()
	
	// 扫描结果
	issues, err := q.scanIssuesWithJoins(rows)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan issues")
	}
	
	// 获取总数
	totalCount, err := q.getTotalCount(criteria)
	if err != nil {
		q.logger.Printf("failed to get total count: %v", err)
		totalCount = len(issues) // fallback
	}
	
	// 检查是否有更多数据
	hasMore := len(issues) == criteria.PageSize
	nextOffset := offset + criteria.PageSize
	
	queryTime := time.Since(startTime)
	
	return &QueryResult{
		Issues:     issues,
		TotalCount: totalCount,
		HasMore:    hasMore,
		NextOffset: nextOffset,
		QueryTime:  queryTime,
		QueryString: query,
	}, nil
}

// AggregatedQuery 执行聚合查询
func (q *QueryBuilder) AggregatedQuery(aq AggregatedQuery) (*AggregationResult, error) {
	startTime := time.Now()
	
	// 构建聚合查询
	query, params := q.buildAggregatedQuery(aq)
	
	rows, err := q.db.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute aggregated query")
	}
	defer rows.Close()
	
	var groups []map[string]interface{}
	
	// 扫描聚合结果
	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get column names")
	}
	
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			q.logger.Printf("failed to scan aggregated row: %v", err)
			continue
		}
		
		group := make(map[string]interface{})
		for i, colName := range columns {
			group[colName] = values[i]
		}
		groups = append(groups, group)
	}
	
	queryTime := time.Since(startTime)
	
	return &AggregationResult{
		Groups:      groups,
		TotalGroups: len(groups),
		QueryTime:   queryTime,
	}, nil
}

// TimeSeriesQuery 执行时间序列查询
func (q *QueryBuilder) TimeSeriesQuery(tsq TimeSeriesQuery) (*TimeSeriesResult, error) {
	startTime := time.Now()
	
	// 构建时间序列查询
	query, params := q.buildTimeSeriesQuery(tsq)
	
	rows, err := q.db.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute time series query")
	}
	defer rows.Close()
	
	var points []TimeSeriesPoint
	
	for rows.Next() {
		var timestamp time.Time
		var repository, category sql.NullString
		var newIssues, closedIssues, pitfallIssues, totalComments sql.NullInt64
		var avgSeverity, avgScore sql.NullFloat64
		
		err := rows.Scan(
			&timestamp, &repository, &category,
			&newIssues, &closedIssues, &pitfallIssues, &totalComments,
			&avgSeverity, &avgScore,
		)
		if err != nil {
			q.logger.Printf("failed to scan time series row: %v", err)
			continue
		}
		
		metrics := make(map[string]float64)
		if newIssues.Valid {
			metrics["new_issues"] = float64(newIssues.Int64)
		}
		if closedIssues.Valid {
			metrics["closed_issues"] = float64(closedIssues.Int64)
		}
		if pitfallIssues.Valid {
			metrics["pitfall_issues"] = float64(pitfallIssues.Int64)
		}
		if totalComments.Valid {
			metrics["total_comments"] = float64(totalComments.Int64)
		}
		if avgSeverity.Valid {
			metrics["avg_severity"] = avgSeverity.Float64
		}
		if avgScore.Valid {
			metrics["avg_score"] = avgScore.Float64
		}
		
		point := TimeSeriesPoint{
			Timestamp:   timestamp,
			Metrics:     metrics,
		}
		
		if repository.Valid {
			point.Repository = repository.String
		}
		if category.Valid {
			point.Category = category.String
		}
		
		points = append(points, point)
	}
	
	queryTime := time.Since(startTime)
	
	return &TimeSeriesResult{
		Points:       points,
		TotalPoints:  len(points),
		QueryTime:    queryTime,
	}, nil
}

// FacetedQuery 执行分面查询
func (q *QueryBuilder) FacetedQuery(fq FacetedQuery) (*FacetedResult, error) {
	startTime := time.Now()
	
	// 首先执行基础查询获取Issues
	qr, err := q.SimpleQuery(fq.BaseCriteria)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute base query")
	}
	
	// 计算分面
	facets := make(map[string]map[string]int)
	
	for _, facet := range fq.Facets {
		facets[facet] = make(map[string]int)
		
		switch facet {
		case "category":
			for _, issue := range qr.Issues {
				categoryName := getCategoryName(issue)
				facets[facet][categoryName]++
			}
		case "repository":
			for _, issue := range qr.Issues {
				if issue.Repository != nil {
					facets[facet][issue.Repository.FullName]++
				}
			}
		case "state":
			for _, issue := range qr.Issues {
				facets[facet][issue.State]++
			}
		case "author":
			for _, issue := range qr.Issues {
				facets[facet][issue.AuthorLogin]++
			}
		case "label":
			for _, issue := range qr.Issues {
				for _, label := range issue.Labels {
					facets[facet][label]++
				}
			}
		case "tag":
			for _, issue := range qr.Issues {
				for _, label := range issue.Labels {
					facets[facet][label]++
				}
			}
		}
	}
	
	queryTime := time.Since(startTime)
	
	return &FacetedResult{
		Issues:      qr.Issues,
		Facets:      facets,
		TotalCount:  qr.TotalCount,
		QueryTime:   queryTime,
	}, nil
}

// buildWhereConditions 构建WHERE条件
func (q *QueryBuilder) buildWhereConditions(baseQuery string, criteria SearchCriteria, paramCount *int) (string, []interface{}) {
	query := baseQuery
	var params []interface{}
	
	// 文本搜索
	if criteria.Query != "" {
		*paramCount++
		if criteria.FuzzyMatch {
			query += fmt.Sprintf(" AND (i.title ILIKE $%d OR i.body ILIKE $%d)", *paramCount, *paramCount+1)
			searchTerm := "%" + strings.ToLower(criteria.Query) + "%"
			params = append(params, searchTerm, searchTerm)
			*paramCount++
		} else {
			query += fmt.Sprintf(" AND (LOWER(i.title) LIKE LOWER($%d) OR LOWER(i.body) LIKE LOWER($%d))", *paramCount, *paramCount+1)
			searchTerm := "%" + criteria.Query + "%"
			params = append(params, searchTerm, searchTerm)
			*paramCount++
		}
	}
	
	// 关键词搜索
	if len(criteria.Keywords) > 0 {
		for _, keyword := range criteria.Keywords {
			*paramCount++
			query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM json_each_text(i.labels) WHERE value = $%d)", *paramCount)
			params = append(params, keyword)
		}
	}
	
	// 分类搜索
	if len(criteria.Categories) > 0 {
		*paramCount++
		placeholders := make([]string, len(criteria.Categories))
		for i := range criteria.Categories {
			placeholders[i] = fmt.Sprintf("$%d", *paramCount+i)
		}
		query += fmt.Sprintf(" AND c.name = ANY(ARRAY[%s])", strings.Join(placeholders, ","))
		params = append(params, interfaceSlice(criteria.Categories)...)
		*paramCount += len(criteria.Categories)
	}
	
	// 排除分类
	if len(criteria.ExcludeCategories) > 0 {
		*paramCount++
		placeholders := make([]string, len(criteria.ExcludeCategories))
		for i := range criteria.ExcludeCategories {
			placeholders[i] = fmt.Sprintf("$%d", *paramCount+i)
		}
		query += fmt.Sprintf(" AND c.name != ALL(ARRAY[%s])", strings.Join(placeholders, ","))
		params = append(params, interfaceSlice(criteria.ExcludeCategories)...)
		*paramCount += len(criteria.ExcludeCategories)
	}
	
	// 仓库搜索
	if len(criteria.Repositories) > 0 {
		for _, repo := range criteria.Repositories {
			*paramCount++
			parts := strings.Split(repo, "/")
			if len(parts) == 2 {
				query += fmt.Sprintf(" AND (r.owner = $%d AND r.name = $%d)", *paramCount, *paramCount+1)
				params = append(params, parts[0], parts[1])
				*paramCount++
			}
		}
	}
	
	// 状态搜索
	if len(criteria.States) > 0 {
		*paramCount++
		query += fmt.Sprintf(" AND i.state = ANY(ARRAY[%s])", fmt.Sprintf("$%d", *paramCount))
		params = append(params, interfaceSlice(criteria.States)...)
	}
	
	// 作者搜索
	if len(criteria.Authors) > 0 {
		*paramCount++
		query += fmt.Sprintf(" AND i.author_login = ANY(ARRAY[%s])", fmt.Sprintf("$%d", *paramCount))
		params = append(params, interfaceSlice(criteria.Authors)...)
	}
	
	// 指派人搜索
	if len(criteria.Assignees) > 0 {
		for _, assignee := range criteria.Assignees {
			*paramCount++
			query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM json_each_text(i.assignees) WHERE value = $%d)", *paramCount)
			params = append(params, assignee)
		}
	}
	
	// 时间范围
	if criteria.DateFrom != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.created_at >= $%d", *paramCount)
		params = append(params, *criteria.DateFrom)
	}
	
	if criteria.DateTo != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.created_at <= $%d", *paramCount)
		params = append(params, *criteria.DateTo)
	}
	
	// 数值范围
	if criteria.MinScore != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.score >= $%d", *paramCount)
		params = append(params, *criteria.MinScore)
	}
	
	if criteria.MaxScore != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.score <= $%d", *paramCount)
		params = append(params, *criteria.MaxScore)
	}
	
	if criteria.MinSeverity != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.severity_score >= $%d", *paramCount)
		params = append(params, *criteria.MinSeverity)
	}
	
	if criteria.MaxSeverity != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.severity_score <= $%d", *paramCount)
		params = append(params, *criteria.MaxSeverity)
	}
	
	if criteria.MinComments != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.comments_count >= $%d", *paramCount)
		params = append(params, *criteria.MinComments)
	}
	
	if criteria.MaxComments != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.comments_count <= $%d", *paramCount)
		params = append(params, *criteria.MaxComments)
	}
	
	// 布尔条件
	if criteria.IsPitfall != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.is_pitfall = $%d", *paramCount)
		params = append(params, *criteria.IsPitfall)
	}
	
	if criteria.IsDuplicate != nil {
		*paramCount++
		query += fmt.Sprintf(" AND i.is_duplicate = $%d", *paramCount)
		params = append(params, *criteria.IsDuplicate)
	}
	
	if criteria.IsOpen != nil {
		if *criteria.IsOpen {
			query += " AND i.state = 'open'"
		} else {
			query += " AND i.state = 'closed'"
		}
	}
	
	if criteria.HasAssignee != nil {
		if *criteria.HasAssignee {
			query += " AND i.assignees IS NOT NULL AND json_array_length(i.assignees) > 0"
		} else {
			query += " AND (i.assignees IS NULL OR json_array_length(i.assignees) = 0)"
		}
	}
	
	return query, params
}

// buildOrderBy 构建ORDER BY子句
func (q *QueryBuilder) buildOrderBy(criteria SearchCriteria) string {
	sortField := "created_at"
	sortOrder := "DESC"
	
	switch strings.ToLower(criteria.SortBy) {
	case "score":
		sortField = "score"
	case "severity":
		sortField = "severity_score"
	case "created":
		sortField = "created_at"
	case "updated":
		sortField = "updated_at"
	case "comments":
		sortField = "comments_count"
	case "title":
		sortField = "title"
	case "number":
		sortField = "number"
	}
	
	if strings.ToUpper(criteria.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}
	
	return fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
}

// buildAggregatedQuery 构建聚合查询
func (q *QueryBuilder) buildAggregatedQuery(aq AggregatedQuery) (string, []interface{}) {
	var query string
	var params []interface{}
	
	switch aq.GroupBy {
	case "category":
		query = `
			SELECT c.name as category, COUNT(*) as count, 
				   AVG(i.score) as avg_score, AVG(i.severity_score) as avg_severity
			FROM issues i
			LEFT JOIN categories c ON i.category_id = c.id
			WHERE 1=1
		`
		// 添加过滤条件
		_, params = q.buildWhereConditions(query, aq.Filters, &struct{ paramCount int }{0})
	case "repository":
		query = `
			SELECT r.full_name as repository, r.language as language,
				   COUNT(*) as count, AVG(i.score) as avg_score
			FROM issues i
			LEFT JOIN repositories r ON i.repository_id = r.id
			WHERE 1=1
		`
		_, params = q.buildWhereConditions(query, aq.Filters, &struct{ paramCount int }{0})
	case "author":
		query = `
			SELECT i.author_login as author, COUNT(*) as count, 
				   AVG(i.score) as avg_score
			FROM issues i
			WHERE 1=1
		`
		_, params = q.buildWhereConditions(query, aq.Filters, &struct{ paramCount int }{0})
	default:
		query = `
			SELECT 'all' as group_name, COUNT(*) as count,
				   AVG(i.score) as avg_score, AVG(i.severity_score) as avg_severity
			FROM issues i
			WHERE 1=1
		`
		_, params = q.buildWhereConditions(query, aq.Filters, &struct{ paramCount int }{0})
	}
	
	query += " GROUP BY " + aq.GroupBy
	query += " ORDER BY count DESC"
	
	return query, params
}

// buildTimeSeriesQuery 构建时间序列查询
func (q *QueryBuilder) buildTimeSeriesQuery(tsq TimeSeriesQuery) (string, []interface{}) {
	var dateFormat string
	var interval string
	
	switch tsq.Interval {
	case "hour":
		dateFormat = "YYYY-MM-DD HH24:00:00"
		interval = "1 hour"
	case "day":
		dateFormat = "YYYY-MM-DD"
		interval = "1 day"
	case "week":
		dateFormat = "YYYY-'W'WW"
		interval = "1 week"
	case "month":
		dateFormat = "YYYY-MM"
		interval = "1 month"
	case "year":
		dateFormat = "YYYY"
		interval = "1 year"
	default:
		dateFormat = "YYYY-MM-DD"
		interval = "1 day"
	}
	
	query := fmt.Sprintf(`
		SELECT 
			DATE_TRUNC('%s', i.created_at) as timestamp,
			r.full_name as repository,
			c.name as category,
			COUNT(*) as new_issues,
			SUM(CASE WHEN i.state = 'closed' THEN 1 ELSE 0 END) as closed_issues,
			SUM(CASE WHEN i.is_pitfall = true THEN 1 ELSE 0 END) as pitfall_issues,
			SUM(i.comments_count) as total_comments,
			AVG(i.severity_score) as avg_severity,
			AVG(i.score) as avg_score
		FROM issues i
		LEFT JOIN repositories r ON i.repository_id = r.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.created_at >= $1 AND i.created_at <= $2
	`, tsq.Interval)
	
	var params []interface{} = []interface{}{tsq.StartDate, tsq.EndDate}
	
	// 添加过滤条件
	if len(tsq.Repositories) > 0 {
		placeholders := make([]string, len(tsq.Repositories))
		for i := range tsq.Repositories {
			placeholders[i] = fmt.Sprintf("$%d", len(params)+1)
			params = append(params, tsq.Repositories[i])
		}
		query += fmt.Sprintf(" AND r.full_name = ANY(ARRAY[%s])", strings.Join(placeholders, ","))
	}
	
	if len(tsq.Categories) > 0 {
		placeholders := make([]string, len(tsq.Categories))
		for i := range tsq.Categories {
			placeholders[i] = fmt.Sprintf("$%d", len(params)+1)
			params = append(params, tsq.Categories[i])
		}
		query += fmt.Sprintf(" AND c.name = ANY(ARRAY[%s])", strings.Join(placeholders, ","))
	}
	
	query += fmt.Sprintf(" GROUP BY DATE_TRUNC('%s', i.created_at), r.full_name, c.name", tsq.Interval)
	query += " ORDER BY timestamp DESC"
	
	return query, params
}

// scanIssuesWithJoins 扫描带关联的Issues
func (q *QueryBuilder) scanIssuesWithJoins(rows *sql.Rows) ([]*Issue, error) {
	var issues []*Issue
	
	for rows.Next() {
		var issue Issue
		var labels, assignees JSONSlice
		var reactions ReactionCount
		var metadata JSONMap
		var repo Repository
		var categoryName, categoryDescription sql.NullString
		
		err := rows.Scan(
			&issue.ID, &issue.Number, &issue.Title, &issue.Body, &issue.URL, &issue.State,
			&issue.AuthorLogin, &labels, &assignees, &issue.Milestone, &reactions,
			&issue.CreatedAt, &issue.UpdatedAt, &issue.ClosedAt, &issue.FirstSeenAt,
			&issue.LastSeenAt, &issue.IsPitfall, &issue.SeverityScore, &issue.CategoryID,
			&issue.Score, &issue.HTMLURL, &issue.CommentsCount, &issue.IsDuplicate,
			&issue.DuplicateOf, &metadata, &repo.FullName, &repo.URL,
			&categoryName, &categoryDescription,
		)
		
		if err != nil {
			q.logger.Printf("failed to scan issue row: %v", err)
			continue
		}
		
		issue.Labels = labels
		issue.Assignees = assignees
		issue.Reactions = reactions
		issue.Metadata = metadata
		issue.Repository = &repo
		
		if categoryName.Valid && categoryName.String != "" {
			issue.Category = &Category{
				ID:          *issue.CategoryID,
				Name:        categoryName.String,
				Description: categoryDescription.String,
			}
		}
		
		issues = append(issues, &issue)
	}
	
	return issues, nil
}

// getTotalCount 获取总数
func (q *QueryBuilder) getTotalCount(criteria SearchCriteria) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM issues i
		LEFT JOIN repositories r ON i.repository_id = r.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE 1=1
	`
	
	var params []interface{}
	paramCount := 0
	
	// 添加过滤条件
	query, params = q.buildWhereConditions(query, criteria, &paramCount)
	
	var total int
	err := q.db.QueryRow(query, params...).Scan(&total)
	if err != nil {
		return 0, err
	}
	
	return total, nil
}

// DefaultSearchCriteria 返回默认搜索条件
func DefaultSearchCriteria() SearchCriteria {
	return SearchCriteria{
		Page:        1,
		PageSize:    100,
		SortBy:      "score",
		SortOrder:   "DESC",
		FuzzyMatch:  false,
	}
}

// ValidateSearchCriteria 验证搜索条件
func ValidateSearchCriteria(criteria SearchCriteria) error {
	if criteria.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	
	if criteria.PageSize < 1 || criteria.PageSize > 1000 {
		return fmt.Errorf("page_size must be between 1 and 1000")
	}
	
	validSortFields := map[string]bool{
		"score":     true,
		"severity":  true,
		"created":   true,
		"updated":   true,
		"comments":  true,
		"title":     true,
		"number":    true,
	}
	
	if !validSortFields[criteria.SortBy] {
		return fmt.Errorf("invalid sort field: %s", criteria.SortBy)
	}
	
	if criteria.SortOrder != "ASC" && criteria.SortOrder != "DESC" {
		return fmt.Errorf("sort order must be ASC or DESC")
	}
	
	return nil
}

// interfaceSlice 转换字符串切片为接口切片
func interfaceSlice(s []string) []interface{} {
	interfaces := make([]interface{}, len(s))
	for i, v := range s {
		interfaces[i] = v
	}
	return interfaces
}

// GetCategoryName 获取分类名称
func getCategoryName(issue *Issue) string {
	if issue.Category != nil {
		return issue.Category.Name
	}
	if issue.CategoryID != nil {
		return fmt.Sprintf("Category ID: %d", *issue.CategoryID)
	}
	return ""
}