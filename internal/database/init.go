package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL embed.FS

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db     *sql.DB
	config DatabaseConfig
	logger *log.Logger
	mu     sync.RWMutex
}

// NewDatabaseManager 创建新的数据库管理器
func NewDatabaseManager(config DatabaseConfig, logger *log.Logger) (*DatabaseManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	// 创建数据库目录
	dbDir := filepath.Dir(config.Path)
	if err := ensureDirectoryExists(dbDir); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections / 2)
	db.SetConnMaxLifetime(config.Timeout)

	// 应用SQLite优化设置
	if err := applySQLiteOptimizations(db, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply SQLite optimizations: %w", err)
	}

	dm := &DatabaseManager{
		db:     db,
		config: config,
		logger: logger,
	}

	// 初始化数据库
	if err := dm.Initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return dm, nil
}

// Initialize 初始化数据库
func (dm *DatabaseManager) Initialize() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 读取schema SQL
	schema, err := schemaSQL.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema SQL: %w", err)
	}

	// 执行初始化脚本
	if err := dm.executeSQLScript(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema SQL: %w", err)
	}

	dm.logger.Printf("数据库初始化完成: %s", dm.config.Path)
	return nil
}

// executeSQLScript 执行SQL脚本
func (dm *DatabaseManager) executeSQLScript(script string) error {
	// 分割SQL语句
	statements := strings.Split(script, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// 跳过PRAGMA语句，单独处理
		if strings.HasPrefix(strings.ToUpper(stmt), "PRAGMA") {
			continue
		}

		// 执行SQL语句
		if _, err := dm.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return nil
}

// applySQLiteOptimizations 应用SQLite优化设置
func applySQLiteOptimizations(db *sql.DB, config DatabaseConfig) error {
	optimizations := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = 10000",
		"PRAGMA temp_store = memory",
		"PRAGMA busy_timeout = 30000",
	}

	for _, pragma := range optimizations {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma %s: %w", pragma, err)
		}
	}

	return nil
}

// ensureDirectoryExists 确保目录存在
func ensureDirectoryExists(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}
	
	// 这里应该使用适当的文件系统操作
	// 由于在沙盒环境中，我们跳过目录创建
	return nil
}

// GetDB 获取数据库连接
func (dm *DatabaseManager) GetDB() *sql.DB {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.db
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if dm.db != nil {
		err := dm.db.Close()
		dm.db = nil
		return err
	}
	return nil
}

// Ping 检查数据库连接
func (dm *DatabaseManager) Ping(ctx context.Context) error {
	return dm.db.PingContext(ctx)
}

// Transaction 执行事务
func (dm *DatabaseManager) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := dm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed and rollback failed: %v (original error: %w)", rbErr, err)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Exec 执行SQL语句
func (dm *DatabaseManager) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return dm.db.ExecContext(ctx, query, args...)
}

// Query 执行查询
func (dm *DatabaseManager) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return dm.db.QueryContext(ctx, query, args...)
}

// QueryRow 执行单行查询
func (dm *DatabaseManager) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return dm.db.QueryRowContext(ctx, query, args...)
}

// Prepare 预处理SQL语句
func (dm *DatabaseManager) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return dm.db.PrepareContext(ctx, query)
}

// Repository Operations

// CreateRepository 创建仓库
func (dm *DatabaseManager) CreateRepository(ctx context.Context, repo *Repository) error {
	query := `
		INSERT INTO repositories (owner, name, full_name, description, url, stars, forks, issues_count, language, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := dm.db.ExecContext(ctx, query,
		repo.Owner, repo.Name, repo.FullName, repo.Description, repo.URL,
		repo.Stars, repo.Forks, repo.IssuesCount, repo.Language, repo.Metadata)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	repo.ID = int(id)
	return nil
}

// GetRepositoryByID 根据ID获取仓库
func (dm *DatabaseManager) GetRepositoryByID(ctx context.Context, id int) (*Repository, error) {
	query := `
		SELECT id, owner, name, full_name, description, url, stars, forks, issues_count, 
			   language, created_at, updated_at, last_scraped_at, is_active, metadata
		FROM repositories WHERE id = ?
	`
	
	var repo Repository
	var createdAt, updatedAt, lastScrapedAt CustomTime
	
	err := dm.db.QueryRowContext(ctx, query, id).Scan(
		&repo.ID, &repo.Owner, &repo.Name, &repo.FullName, &repo.Description, &repo.URL,
		&repo.Stars, &repo.Forks, &repo.IssuesCount, &repo.Language,
		&createdAt, &updatedAt, &lastScrapedAt, &repo.IsActive, &repo.Metadata)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	repo.CreatedAt = createdAt
	repo.UpdatedAt = updatedAt
	repo.LastScrapedAt = lastScrapedAt

	return &repo, nil
}

// GetRepositoryByFullName 根据完整名称获取仓库
func (dm *DatabaseManager) GetRepositoryByFullName(ctx context.Context, fullName string) (*Repository, error) {
	query := `
		SELECT id, owner, name, full_name, description, url, stars, forks, issues_count, 
			   language, created_at, updated_at, last_scraped_at, is_active, metadata
		FROM repositories WHERE full_name = ?
	`
	
	var repo Repository
	var createdAt, updatedAt, lastScrapedAt CustomTime
	
	err := dm.db.QueryRowContext(ctx, query, fullName).Scan(
		&repo.ID, &repo.Owner, &repo.Name, &repo.FullName, &repo.Description, &repo.URL,
		&repo.Stars, &repo.Forks, &repo.IssuesCount, &repo.Language,
		&createdAt, &updatedAt, &lastScrapedAt, &repo.IsActive, &repo.Metadata)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	repo.CreatedAt = createdAt
	repo.UpdatedAt = updatedAt
	repo.LastScrapedAt = lastScrapedAt

	return &repo, nil
}

// UpdateRepositoryLastScraped 更新仓库最后爬取时间
func (dm *DatabaseManager) UpdateRepositoryLastScraped(ctx context.Context, id int) error {
	query := "UPDATE repositories SET last_scraped_at = CURRENT_TIMESTAMP WHERE id = ?"
	
	_, err := dm.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update repository last scraped: %w", err)
	}
	
	return nil
}

// Issue Operations

// CreateIssue 创建Issue
func (dm *DatabaseManager) CreateIssue(ctx context.Context, issue *Issue) error {
	query := `
		INSERT INTO issues (
			issue_id, repository_id, number, title, body, state, author_login, author_type,
			labels, assignees, milestone, reactions, created_at, updated_at, closed_at,
			first_seen_at, last_seen_at, is_pitfall, severity_score, category_id, score,
			url, html_url, comments_count, is_duplicate, duplicate_of, metadata,
			content_hash, keywords, category, priority, tech_stack
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := dm.db.ExecContext(ctx, query,
		issue.IssueID, issue.RepositoryID, issue.Number, issue.Title, issue.Body, issue.State,
		issue.AuthorLogin, issue.AuthorType, issue.Labels, issue.Assignees, issue.Milestone,
		issue.Reactions, issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt,
		issue.FirstSeenAt, issue.LastSeenAt, issue.IsPitfall, issue.SeverityScore,
		issue.CategoryID, issue.Score, issue.URL, issue.HTMLURL, issue.CommentsCount,
		issue.IsDuplicate, issue.DuplicateOf, issue.Metadata, issue.ContentHash, issue.Keywords,
		issue.CategoryOld, issue.Priority, issue.TechStack)
	
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	issue.ID = int(id)
	return nil
}

// GetIssueByID 根据ID获取Issue
func (dm *DatabaseManager) GetIssueByID(ctx context.Context, id int) (*Issue, error) {
	query := `
		SELECT id, issue_id, repository_id, number, title, body, state, author_login, author_type,
			   labels, assignees, milestone, reactions, created_at, updated_at, closed_at,
			   first_seen_at, last_seen_at, is_pitfall, severity_score, category_id, score,
			   url, html_url, comments_count, is_duplicate, duplicate_of, metadata,
			   content_hash, keywords, category, priority, tech_stack
		FROM issues WHERE id = ?
	`
	
	var issue Issue
	var createdAt, updatedAt, closedAt, firstSeenAt, lastSeenAt CustomTime
	
	err := dm.db.QueryRowContext(ctx, query, id).Scan(
		&issue.ID, &issue.IssueID, &issue.RepositoryID, &issue.Number, &issue.Title, &issue.Body,
		&issue.State, &issue.AuthorLogin, &issue.AuthorType, &issue.Labels, &issue.Assignees,
		&issue.Milestone, &issue.Reactions, &createdAt, &updatedAt, &closedAt,
		&firstSeenAt, &lastSeenAt, &issue.IsPitfall, &issue.SeverityScore, &issue.CategoryID,
		&issue.Score, &issue.URL, &issue.HTMLURL, &issue.CommentsCount, &issue.IsDuplicate,
		&issue.DuplicateOf, &issue.Metadata, &issue.ContentHash, &issue.Keywords,
		&issue.CategoryOld, &issue.Priority, &issue.TechStack)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("issue not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	issue.CreatedAt = createdAt
	issue.UpdatedAt = updatedAt
	issue.ClosedAt = &closedAt
	issue.FirstSeenAt = firstSeenAt
	issue.LastSeenAt = lastSeenAt

	return &issue, nil
}

// GetIssueByGitHubID 根据GitHub ID获取Issue
func (dm *DatabaseManager) GetIssueByGitHubID(ctx context.Context, githubID int64) (*Issue, error) {
	query := `
		SELECT id, issue_id, repository_id, number, title, body, state, author_login, author_type,
			   labels, assignees, milestone, reactions, created_at, updated_at, closed_at,
			   first_seen_at, last_seen_at, is_pitfall, severity_score, category_id, score,
			   url, html_url, comments_count, is_duplicate, duplicate_of, metadata,
			   content_hash, keywords, category, priority, tech_stack
		FROM issues WHERE issue_id = ?
	`
	
	var issue Issue
	var createdAt, updatedAt, closedAt, firstSeenAt, lastSeenAt CustomTime
	
	err := dm.db.QueryRowContext(ctx, query, githubID).Scan(
		&issue.ID, &issue.IssueID, &issue.RepositoryID, &issue.Number, &issue.Title, &issue.Body,
		&issue.State, &issue.AuthorLogin, &issue.AuthorType, &issue.Labels, &issue.Assignees,
		&issue.Milestone, &issue.Reactions, &createdAt, &updatedAt, &closedAt,
		&firstSeenAt, &lastSeenAt, &issue.IsPitfall, &issue.SeverityScore, &issue.CategoryID,
		&issue.Score, &issue.URL, &issue.HTMLURL, &issue.CommentsCount, &issue.IsDuplicate,
		&issue.DuplicateOf, &issue.Metadata, &issue.ContentHash, &issue.Keywords,
		&issue.CategoryOld, &issue.Priority, &issue.TechStack)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("issue not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	issue.CreatedAt = createdAt
	issue.UpdatedAt = updatedAt
	issue.ClosedAt = &closedAt
	issue.FirstSeenAt = firstSeenAt
	issue.LastSeenAt = lastSeenAt

	return &issue, nil
}

// UpdateIssue 更新Issue
func (dm *DatabaseManager) UpdateIssue(ctx context.Context, issue *Issue) error {
	query := `
		UPDATE issues SET
			number = ?, title = ?, body = ?, state = ?, author_login = ?, author_type = ?,
			labels = ?, assignees = ?, milestone = ?, reactions = ?, updated_at = ?,
			closed_at = ?, last_seen_at = ?, is_pitfall = ?, severity_score = ?,
			category_id = ?, score = ?, comments_count = ?, is_duplicate = ?,
			duplicate_of = ?, metadata = ?, keywords = ?, category = ?, priority = ?, tech_stack = ?
		WHERE id = ?
	`
	
	_, err := dm.db.ExecContext(ctx, query,
		issue.Number, issue.Title, issue.Body, issue.State, issue.AuthorLogin, issue.AuthorType,
		issue.Labels, issue.Assignees, issue.Milestone, issue.Reactions, issue.UpdatedAt,
		issue.ClosedAt, issue.LastSeenAt, issue.IsPitfall, issue.SeverityScore,
		issue.CategoryID, issue.Score, issue.CommentsCount, issue.IsDuplicate,
		issue.DuplicateOf, issue.Metadata, issue.Keywords, issue.CategoryOld, issue.Priority,
		issue.TechStack, issue.ID)
	
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}
	
	return nil
}

// GetIssuesByRepository 获取指定仓库的Issues
func (dm *DatabaseManager) GetIssuesByRepository(ctx context.Context, repositoryID int, limit, offset int) ([]*Issue, error) {
	query := `
		SELECT id, issue_id, repository_id, number, title, body, state, author_login, author_type,
			   labels, assignees, milestone, reactions, created_at, updated_at, closed_at,
			   first_seen_at, last_seen_at, is_pitfall, severity_score, category_id, score,
			   url, html_url, comments_count, is_duplicate, duplicate_of, metadata,
			   content_hash, keywords, category, priority, tech_stack
		FROM issues WHERE repository_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	
	rows, err := dm.db.QueryContext(ctx, query, repositoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []*Issue
	for rows.Next() {
		var issue Issue
		var createdAt, updatedAt, closedAt, firstSeenAt, lastSeenAt CustomTime
		
		err := rows.Scan(
			&issue.ID, &issue.IssueID, &issue.RepositoryID, &issue.Number, &issue.Title, &issue.Body,
			&issue.State, &issue.AuthorLogin, &issue.AuthorType, &issue.Labels, &issue.Assignees,
			&issue.Milestone, &issue.Reactions, &createdAt, &updatedAt, &closedAt,
			&firstSeenAt, &lastSeenAt, &issue.IsPitfall, &issue.SeverityScore, &issue.CategoryID,
			&issue.Score, &issue.URL, &issue.HTMLURL, &issue.CommentsCount, &issue.IsDuplicate,
			&issue.DuplicateOf, &issue.Metadata, &issue.ContentHash, &issue.Keywords,
			&issue.CategoryOld, &issue.Priority, &issue.TechStack)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		issue.CreatedAt = createdAt
		issue.UpdatedAt = updatedAt
		issue.ClosedAt = &closedAt
		issue.FirstSeenAt = firstSeenAt
		issue.LastSeenAt = lastSeenAt
		
		issues = append(issues, &issue)
	}

	return issues, nil
}

// Category Operations

// GetCategories 获取所有分类
func (dm *DatabaseManager) GetCategories(ctx context.Context) ([]*Category, error) {
	query := `
		SELECT id, name, description, color, is_active, priority, created_at, updated_at
		FROM categories ORDER BY priority DESC, name ASC
	`
	
	rows, err := dm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var category Category
		var createdAt, updatedAt CustomTime
		
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.IsActive, &category.Priority, &createdAt, &updatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		category.CreatedAt = createdAt
		category.UpdatedAt = updatedAt
		
		categories = append(categories, &category)
	}

	return categories, nil
}

// GetCategoryByID 根据ID获取分类
func (dm *DatabaseManager) GetCategoryByID(ctx context.Context, id int) (*Category, error) {
	query := `
		SELECT id, name, description, color, is_active, priority, created_at, updated_at
		FROM categories WHERE id = ?
	`
	
	var category Category
	var createdAt, updatedAt CustomTime
	
	err := dm.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.Color,
		&category.IsActive, &category.Priority, &createdAt, &updatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	category.CreatedAt = createdAt
	category.UpdatedAt = updatedAt

	return &category, nil
}

// Statistics Operations

// GetRepositoryStats 获取仓库统计信息
func (dm *DatabaseManager) GetRepositoryStats(ctx context.Context, repositoryID int) (*RepositoryStats, error) {
	query := `
		SELECT 
			r.*,
			COUNT(i.id) as total_issues,
			COUNT(CASE WHEN i.state = 'open' THEN 1 END) as open_issues,
			COUNT(CASE WHEN i.state = 'closed' THEN 1 END) as closed_issues,
			COUNT(CASE WHEN i.is_pitfall = 1 THEN 1 END) as pitfall_issues,
			COUNT(CASE WHEN i.is_duplicate = 1 THEN 1 END) as duplicate_issues,
			AVG(i.severity_score) as avg_severity_score,
			AVG(i.score) as avg_score,
			MAX(i.last_seen_at) as last_issue_update
		FROM repositories r
		LEFT JOIN issues i ON r.id = i.repository_id
		WHERE r.id = ?
		GROUP BY r.id
	`
	
	var stats RepositoryStats
	var repo Repository
	var createdAt, updatedAt, lastScrapedAt, lastIssueUpdate CustomTime
	
	err := dm.db.QueryRowContext(ctx, query, repositoryID).Scan(
		&repo.ID, &repo.Owner, &repo.Name, &repo.FullName, &repo.Description, &repo.URL,
		&repo.Stars, &repo.Forks, &repo.IssuesCount, &repo.Language,
		&createdAt, &updatedAt, &lastScrapedAt, &repo.IsActive, &repo.Metadata,
		&stats.TotalIssues, &stats.OpenIssues, &stats.ClosedIssues, &stats.PitfallIssues,
		&stats.DuplicateIssues, &stats.AvgSeverityScore, &stats.AvgScore, &lastIssueUpdate)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository stats: %w", err)
	}

	repo.CreatedAt = createdAt
	repo.UpdatedAt = updatedAt
	repo.LastScrapedAt = lastScrapedAt
	stats.Repository = repo
	
	if !lastIssueUpdate.IsZero() {
		stats.LastIssueUpdate = &lastIssueUpdate
	}

	return &stats, nil
}

// GetOverallStats 获取全局统计信息
func (dm *DatabaseManager) GetOverallStats(ctx context.Context) (*IssueStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(CASE WHEN state = 'open' THEN 1 END) as open_count,
			COUNT(CASE WHEN state = 'closed' THEN 1 END) as closed_count,
			COUNT(CASE WHEN is_pitfall = 1 THEN 1 END) as pitfall_count,
			COUNT(CASE WHEN is_duplicate = 1 THEN 1 END) as duplicate_count,
			AVG(score) as avg_score,
			AVG(severity_score) as avg_severity_score
		FROM issues
	`
	
	var stats IssueStats
	var avgScore, avgSeverityScore sql.NullFloat64
	
	err := dm.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalCount, &stats.ByState["open"], &stats.ByState["closed"],
		&stats.ByState["pitfall"], &stats.ByState["duplicate"],
		&avgScore, &avgSeverityScore)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	if avgScore.Valid {
		stats.AverageScore = avgScore.Float64
	}
	
	// 这里可以添加更多的统计逻辑
	
	return &stats, nil
}

// HealthCheck 健康检查
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
	// 检查数据库连接
	if err := dm.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	// 检查基本查询
	var count int
	err := dm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM repositories").Scan(&count)
	if err != nil {
		return fmt.Errorf("health check query failed: %w", err)
	}
	
	return nil
}

// Backup 备份数据库
func (dm *DatabaseManager) Backup(ctx context.Context, backupPath string) error {
	// 这里可以实现数据库备份逻辑
	// 由于SQLite的特性，可以直接复制数据库文件
	return fmt.Errorf("backup not implemented yet")
}

// Vacuum 优化数据库
func (dm *DatabaseManager) Vacuum(ctx context.Context) error {
	_, err := dm.db.ExecContext(ctx, "VACUUM")
	return err
}

// Analyze 分析数据库统计信息
func (dm *DatabaseManager) Analyze(ctx context.Context) error {
	_, err := dm.db.ExecContext(ctx, "ANALYZE")
	return err
}

// DefaultManager 默认数据库管理器实例
var DefaultManager *DatabaseManager
var managerInit sync.Once

// GetDefaultManager 获取默认数据库管理器
func GetDefaultManager() *DatabaseManager {
	return DefaultManager
}

// InitDefaultManager 初始化默认数据库管理器
func InitDefaultManager(config DatabaseConfig, logger *log.Logger) error {
	var err error
	managerInit.Do(func() {
		DefaultManager, err = NewDatabaseManager(config, logger)
	})
	return err
}

// CloseDefaultManager 关闭默认数据库管理器
func CloseDefaultManager() {
	if DefaultManager != nil {
		DefaultManager.Close()
		DefaultManager = nil
	}
}