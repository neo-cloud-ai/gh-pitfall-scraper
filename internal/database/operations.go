package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// CRUDOperations defines the CRUD operations interface
type CRUDOperations interface {
	// Issue operations
	CreateIssue(issue *Issue) (int64, error)
	GetIssue(id int64) (*Issue, error)
	UpdateIssue(issue *Issue) error
	DeleteIssue(id int64) error
	
	// Batch operations
	CreateIssues(issues []*Issue) ([]int64, error)
	UpdateIssues(issues []*Issue) error
	DeleteIssues(ids []int64) error
	
	// Query operations
	GetAllIssues(limit, offset int) ([]*Issue, error)
	GetIssuesByRepository(owner, name string, limit, offset int) ([]*Issue, error)
	GetIssuesByScore(minScore, maxScore float64, limit, offset int) ([]*Issue, error)
	GetIssuesByCategory(category string, limit, offset int) ([]*Issue, error)
	GetIssuesByPriority(priority string, limit, offset int) ([]*Issue, error)
	GetIssuesByKeywords(keywords []string, limit, offset int) ([]*Issue, error)
	GetIssuesByTechStack(techStack []string, limit, offset int) ([]*Issue, error)
	
	// Search operations
	SearchIssues(query string, limit, offset int) ([]*Issue, error)
	SearchIssuesAdvanced(search *AdvancedSearch) ([]*Issue, error)
	
	// Repository operations
	CreateRepository(repo *Repository) (int64, error)
	GetRepository(id int64) (*Repository, error)
	UpdateRepository(repo *Repository) error
	DeleteRepository(id int64) error
	GetAllRepositories(limit, offset int) ([]*Repository, error)
	GetRepositoryByName(owner, name string) (*Repository, error)
	
	// Statistics
	GetIssueStats() (*IssueStats, error)
	GetRepositoryStats(owner, name string) (*IssueStats, error)
	
	// Maintenance
	Cleanup() error
	Optimize() error
}

// AdvancedSearch represents advanced search criteria
type AdvancedSearch struct {
	Query       string      `json:"query"`
	Keywords    []string    `json:"keywords"`
	Categories  []string    `json:"categories"`
	Priorities  []string    `json:"priorities"`
	TechStacks  []string    `json:"tech_stacks"`
	Repos       []string    `json:"repos"` // Format: "owner/name"
	States      []string    `json:"states"`
	MinScore    *float64    `json:"min_score"`
	MaxScore    *float64    `json:"max_score"`
	DateFrom    *time.Time  `json:"date_from"`
	DateTo      *time.Time  `json:"date_to"`
	SortBy      string      `json:"sort_by"` // "score", "created_at", "updated_at", "comments", "reactions"
	SortOrder   string      `json:"sort_order"` // "ASC", "DESC"
	Limit       int         `json:"limit"`
	Offset      int         `json:"offset"`
	ExcludeDuplicates bool   `json:"exclude_duplicates"`
}

// DefaultAdvancedSearch returns default search configuration
func DefaultAdvancedSearch() *AdvancedSearch {
	return &AdvancedSearch{
		SortBy:         "score",
		SortOrder:      "DESC",
		Limit:          100,
		Offset:         0,
		ExcludeDuplicates: true,
	}
}

// CRUDOperationsImpl implements the CRUD operations
type CRUDOperationsImpl struct {
	db      *sql.DB
	logger  *log.Logger
}

// NewCRUDOperations creates a new CRUD operations instance
func NewCRUDOperations(db *sql.DB) CRUDOperations {
	return &CRUDOperationsImpl{
		db:     db,
		logger: log.New(log.Writer(), "[CRUD] ", log.LstdFlags),
	}
}

// CreateIssue creates a new issue
func (c *CRUDOperationsImpl) CreateIssue(issue *Issue) (int64, error) {
	if issue == nil {
		return 0, fmt.Errorf("issue cannot be nil")
	}
	
	// Set timestamps
	now := time.Now()
	issue.CreatedAtDB = now
	issue.UpdatedAtDB = now
	
	// Generate content hash if not set
	if issue.ContentHash == "" {
		issue.ContentHash = issue.GetContentHash()
	}
	
	query := `
		INSERT INTO issues (
			number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := c.db.Exec(query,
		issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
		issue.CreatedAt, issue.UpdatedAt, issue.Comments, issue.Reactions,
		issue.Assignee, issue.Milestone, issue.RepoOwner, issue.RepoName,
		issue.Keywords, issue.Score, issue.ContentHash, issue.Category,
		issue.Priority, issue.TechStack, issue.Labels, issue.IsDuplicate,
		issue.DuplicateOf, issue.CreatedAtDB, issue.UpdatedAtDB,
	)
	if err != nil {
		c.logger.Printf("Error creating issue: %v", err)
		return 0, fmt.Errorf("failed to create issue: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	
	issue.ID = int(id)
	c.logger.Printf("Created issue with ID: %d, Title: %s", id, issue.Title)
	return id, nil
}

// GetIssue retrieves an issue by ID
func (c *CRUDOperationsImpl) GetIssue(id int64) (*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues WHERE id = ?
	`
	
	var issue Issue
	var keywords, techStack, labels StringArray
	
	err := c.db.QueryRow(query, id).Scan(
		&issue.ID, &issue.Number, &issue.Title, &issue.Body, &issue.URL,
		&issue.State, &issue.CreatedAt, &issue.UpdatedAt, &issue.Comments,
		&issue.Reactions, &issue.Assignee, &issue.Milestone, &issue.RepoOwner,
		&issue.RepoName, &keywords, &issue.Score, &issue.ContentHash,
		&issue.Category, &issue.Priority, &techStack, &labels,
		&issue.IsDuplicate, &issue.DuplicateOf, &issue.CreatedAtDB,
		&issue.UpdatedAtDB,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("issue not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query issue: %w", err)
	}
	
	issue.Keywords = keywords
	issue.TechStack = techStack
	issue.Labels = labels
	
	return &issue, nil
}

// UpdateIssue updates an existing issue
func (c *CRUDOperationsImpl) UpdateIssue(issue *Issue) error {
	if issue == nil {
		return fmt.Errorf("issue cannot be nil")
	}
	
	issue.UpdatedAtDB = time.Now()
	
	query := `
		UPDATE issues SET 
			number = ?, title = ?, body = ?, url = ?, state = ?, created_at = ?,
			updated_at = ?, comments = ?, reactions = ?, assignee = ?, milestone = ?,
			repo_owner = ?, repo_name = ?, keywords = ?, score = ?, content_hash = ?,
			category = ?, priority = ?, tech_stack = ?, labels = ?, is_duplicate = ?,
			duplicate_of = ?, updated_at_db = ?
		WHERE id = ?
	`
	
	result, err := c.db.Exec(query,
		issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
		issue.CreatedAt, issue.UpdatedAt, issue.Comments, issue.Reactions,
		issue.Assignee, issue.Milestone, issue.RepoOwner, issue.RepoName,
		issue.Keywords, issue.Score, issue.ContentHash, issue.Category,
		issue.Priority, issue.TechStack, issue.Labels, issue.IsDuplicate,
		issue.DuplicateOf, issue.UpdatedAtDB, issue.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("issue not found")
	}
	
	c.logger.Printf("Updated issue with ID: %d", issue.ID)
	return nil
}

// DeleteIssue deletes an issue by ID
func (c *CRUDOperationsImpl) DeleteIssue(id int64) error {
	query := "DELETE FROM issues WHERE id = ?"
	
	result, err := c.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("issue not found")
	}
	
	c.logger.Printf("Deleted issue with ID: %d", id)
	return nil
}

// CreateIssues creates multiple issues in a transaction
func (c *CRUDOperationsImpl) CreateIssues(issues []*Issue) ([]int64, error) {
	if len(issues) == 0 {
		return nil, nil
	}
	
	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	stmt, err := tx.Prepare(`
		INSERT INTO issues (
			number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	var ids []int64
	now := time.Now()
	
	for _, issue := range issues {
		if issue == nil {
			tx.Rollback()
			return nil, fmt.Errorf("issue cannot be nil")
		}
		
		// Set timestamps
		issue.CreatedAtDB = now
		issue.UpdatedAtDB = now
		
		// Generate content hash if not set
		if issue.ContentHash == "" {
			issue.ContentHash = issue.GetContentHash()
		}
		
		result, err := stmt.Exec(
			issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
			issue.CreatedAt, issue.UpdatedAt, issue.Comments, issue.Reactions,
			issue.Assignee, issue.Milestone, issue.RepoOwner, issue.RepoName,
			issue.Keywords, issue.Score, issue.ContentHash, issue.Category,
			issue.Priority, issue.TechStack, issue.Labels, issue.IsDuplicate,
			issue.DuplicateOf, issue.CreatedAtDB, issue.UpdatedAtDB,
		)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to insert issue: %w", err)
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}
		
		issue.ID = int(id)
		ids = append(ids, id)
	}
	
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	c.logger.Printf("Created %d issues successfully", len(ids))
	return ids, nil
}

// UpdateIssues updates multiple issues
func (c *CRUDOperationsImpl) UpdateIssues(issues []*Issue) error {
	if len(issues) == 0 {
		return nil
	}
	
	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	stmt, err := tx.Prepare(`
		UPDATE issues SET 
			number = ?, title = ?, body = ?, url = ?, state = ?, created_at = ?,
			updated_at = ?, comments = ?, reactions = ?, assignee = ?, milestone = ?,
			repo_owner = ?, repo_name = ?, keywords = ?, score = ?, content_hash = ?,
			category = ?, priority = ?, tech_stack = ?, labels = ?, is_duplicate = ?,
			duplicate_of = ?, updated_at_db = ?
		WHERE id = ?
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	now := time.Now()
	updatedCount := 0
	
	for _, issue := range issues {
		if issue == nil {
			tx.Rollback()
			return fmt.Errorf("issue cannot be nil")
		}
		
		issue.UpdatedAtDB = now
		
		result, err := stmt.Exec(
			issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
			issue.CreatedAt, issue.UpdatedAt, issue.Comments, issue.Reactions,
			issue.Assignee, issue.Milestone, issue.RepoOwner, issue.RepoName,
			issue.Keywords, issue.Score, issue.ContentHash, issue.Category,
			issue.Priority, issue.TechStack, issue.Labels, issue.IsDuplicate,
			issue.DuplicateOf, issue.UpdatedAtDB, issue.ID,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update issue: %w", err)
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		
		if rowsAffected > 0 {
			updatedCount++
		}
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	c.logger.Printf("Updated %d issues successfully", updatedCount)
	return nil
}

// DeleteIssues deletes multiple issues
func (c *CRUDOperationsImpl) DeleteIssues(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	
	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	stmt, err := tx.Prepare("DELETE FROM issues WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	deletedCount := 0
	
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete issue: %w", err)
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		
		if rowsAffected > 0 {
			deletedCount++
		}
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	c.logger.Printf("Deleted %d issues successfully", deletedCount)
	return nil
}

// GetAllIssues retrieves all issues with pagination
func (c *CRUDOperationsImpl) GetAllIssues(limit, offset int) ([]*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		ORDER BY created_at_db DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := c.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()
	
	var issues []*Issue
	for rows.Next() {
		var issue Issue
		var keywords, techStack, labels StringArray
		
		err := rows.Scan(
			&issue.ID, &issue.Number, &issue.Title, &issue.Body, &issue.URL,
			&issue.State, &issue.CreatedAt, &issue.UpdatedAt, &issue.Comments,
			&issue.Reactions, &issue.Assignee, &issue.Milestone, &issue.RepoOwner,
			&issue.RepoName, &keywords, &issue.Score, &issue.ContentHash,
			&issue.Category, &issue.Priority, &techStack, &labels,
			&issue.IsDuplicate, &issue.DuplicateOf, &issue.CreatedAtDB,
			&issue.UpdatedAtDB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		
		issue.Keywords = keywords
		issue.TechStack = techStack
		issue.Labels = labels
		issues = append(issues, &issue)
	}
	
	return issues, nil
}

// GetIssuesByRepository retrieves issues for a specific repository
func (c *CRUDOperationsImpl) GetIssuesByRepository(owner, name string, limit, offset int) ([]*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE repo_owner = ? AND repo_name = ?
		ORDER BY score DESC, created_at_db DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := c.db.Query(query, owner, name, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query repository issues: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// GetIssuesByScore retrieves issues within a score range
func (c *CRUDOperationsImpl) GetIssuesByScore(minScore, maxScore float64, limit, offset int) ([]*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE score >= ? AND score <= ?
		ORDER BY score DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := c.db.Query(query, minScore, maxScore, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues by score: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// GetIssuesByCategory retrieves issues by category
func (c *CRUDOperationsImpl) GetIssuesByCategory(category string, limit, offset int) ([]*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE category = ?
		ORDER BY score DESC, created_at_db DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := c.db.Query(query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues by category: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// GetIssuesByPriority retrieves issues by priority
func (c *CRUDOperationsImpl) GetIssuesByPriority(priority string, limit, offset int) ([]*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE priority = ?
		ORDER BY score DESC, created_at_db DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := c.db.Query(query, priority, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues by priority: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// GetIssuesByKeywords retrieves issues containing specific keywords
func (c *CRUDOperationsImpl) GetIssuesByKeywords(keywords []string, limit, offset int) ([]*Issue, error) {
	if len(keywords) == 0 {
		return c.GetAllIssues(limit, offset)
	}
	
	// Build query with keyword conditions
	conditions := make([]string, len(keywords))
	params := make([]interface{}, len(keywords))
	
	for i, keyword := range keywords {
		conditions[i] = "JSON_CONTAINS(keywords, ?)"
		params[i] = `"` + keyword + `"`
	}
	
	query := fmt.Sprintf(`
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE %s
		ORDER BY score DESC, created_at_db DESC
		LIMIT ? OFFSET ?
	`, strings.Join(conditions, " AND "))
	
	params = append(params, limit, offset)
	
	rows, err := c.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues by keywords: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// GetIssuesByTechStack retrieves issues by technology stack
func (c *CRUDOperationsImpl) GetIssuesByTechStack(techStack []string, limit, offset int) ([]*Issue, error) {
	if len(techStack) == 0 {
		return c.GetAllIssues(limit, offset)
	}
	
	// Build query with tech stack conditions
	conditions := make([]string, len(techStack))
	params := make([]interface{}, len(techStack))
	
	for i, tech := range techStack {
		conditions[i] = "JSON_CONTAINS(tech_stack, ?)"
		params[i] = `"` + tech + `"`
	}
	
	query := fmt.Sprintf(`
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE %s
		ORDER BY score DESC, created_at_db DESC
		LIMIT ? OFFSET ?
	`, strings.Join(conditions, " AND "))
	
	params = append(params, limit, offset)
	
	rows, err := c.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues by tech stack: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// SearchIssues performs a simple text search in titles and bodies
func (c *CRUDOperationsImpl) SearchIssues(query string, limit, offset int) ([]*Issue, error) {
	if query == "" {
		return c.GetAllIssues(limit, offset)
	}
	
	searchQuery := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE title LIKE ? OR body LIKE ?
		ORDER BY score DESC, created_at_db DESC
		LIMIT ? OFFSET ?
	`
	
	likeQuery := "%" + query + "%"
	
	rows, err := c.db.Query(searchQuery, likeQuery, likeQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// SearchIssuesAdvanced performs advanced search with multiple criteria
func (c *CRUDOperationsImpl) SearchIssuesAdvanced(search *AdvancedSearch) ([]*Issue, error) {
	if search == nil {
		search = DefaultAdvancedSearch()
	}
	
	query, params := c.buildAdvancedSearchQuery(search)
	
	rows, err := c.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to perform advanced search: %w", err)
	}
	defer rows.Close()
	
	return c.scanIssues(rows)
}

// scanIssues helper function to scan issue rows
func (c *CRUDOperationsImpl) scanIssues(rows *sql.Rows) ([]*Issue, error) {
	var issues []*Issue
	for rows.Next() {
		var issue Issue
		var keywords, techStack, labels StringArray
		
		err := rows.Scan(
			&issue.ID, &issue.Number, &issue.Title, &issue.Body, &issue.URL,
			&issue.State, &issue.CreatedAt, &issue.UpdatedAt, &issue.Comments,
			&issue.Reactions, &issue.Assignee, &issue.Milestone, &issue.RepoOwner,
			&issue.RepoName, &keywords, &issue.Score, &issue.ContentHash,
			&issue.Category, &issue.Priority, &techStack, &labels,
			&issue.IsDuplicate, &issue.DuplicateOf, &issue.CreatedAtDB,
			&issue.UpdatedAtDB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		
		issue.Keywords = keywords
		issue.TechStack = techStack
		issue.Labels = labels
		issues = append(issues, &issue)
	}
	
	return issues, nil
}

// buildAdvancedSearchQuery builds the SQL query for advanced search
func (c *CRUDOperationsImpl) buildAdvancedSearchQuery(search *AdvancedSearch) (string, []interface{}) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues WHERE 1=1
	`
	
	var params []interface{}
	
	// Add text search condition
	if search.Query != "" {
		query += " AND (title LIKE ? OR body LIKE ?)"
		likeQuery := "%" + search.Query + "%"
		params = append(params, likeQuery, likeQuery)
	}
	
	// Add category filter
	if len(search.Categories) > 0 {
		categoryPlaceholders := make([]string, len(search.Categories))
		for i := range search.Categories {
			categoryPlaceholders[i] = "?"
		}
		query += fmt.Sprintf(" AND category IN (%s)", strings.Join(categoryPlaceholders, ","))
		params = append(params, interfaceSlice(search.Categories)...)
	}
	
	// Add priority filter
	if len(search.Priorities) > 0 {
		priorityPlaceholders := make([]string, len(search.Priorities))
		for i := range search.Priorities {
			priorityPlaceholders[i] = "?"
		}
		query += fmt.Sprintf(" AND priority IN (%s)", strings.Join(priorityPlaceholders, ","))
		params = append(params, interfaceSlice(search.Priorities)...)
	}
	
	// Add state filter
	if len(search.States) > 0 {
		statePlaceholders := make([]string, len(search.States))
		for i := range search.States {
			statePlaceholders[i] = "?"
		}
		query += fmt.Sprintf(" AND state IN (%s)", strings.Join(statePlaceholders, ","))
		params = append(params, interfaceSlice(search.States)...)
	}
	
	// Add repository filter
	if len(search.Repos) > 0 {
		for _, repo := range search.Repos {
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) == 2 {
				query += " AND (repo_owner = ? AND repo_name = ?)"
				params = append(params, parts[0], parts[1])
			}
		}
	}
	
	// Add score range filter
	if search.MinScore != nil {
		query += " AND score >= ?"
		params = append(params, *search.MinScore)
	}
	
	if search.MaxScore != nil {
		query += " AND score <= ?"
		params = append(params, *search.MaxScore)
	}
	
	// Add date range filter
	if search.DateFrom != nil {
		query += " AND created_at >= ?"
		params = append(params, *search.DateFrom)
	}
	
	if search.DateTo != nil {
		query += " AND created_at <= ?"
		params = append(params, *search.DateTo)
	}
	
	// Add duplicate filter
	if search.ExcludeDuplicates {
		query += " AND is_duplicate = 0"
	}
	
	// Add JSON array filters
	if len(search.Keywords) > 0 {
		for _, keyword := range search.Keywords {
			query += " AND JSON_CONTAINS(keywords, ?)"
			params = append(params, `"`+keyword+`"`)
		}
	}
	
	if len(search.TechStacks) > 0 {
		for _, tech := range search.TechStacks {
			query += " AND JSON_CONTAINS(tech_stack, ?)"
			params = append(params, `"`+tech+`"`)
		}
	}
	
	// Add sorting
	sortField := getSortField(search.SortBy)
	sortOrder := strings.ToUpper(search.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
	
	// Add pagination
	query += " LIMIT ? OFFSET ?"
	params = append(params, search.Limit, search.Offset)
	
	return query, params
}

// getSortField maps sort field names to database columns
func getSortField(field string) string {
	switch strings.ToLower(field) {
	case "score":
		return "score"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	case "comments":
		return "comments"
	case "reactions":
		return "reactions"
	case "title":
		return "title"
	default:
		return "created_at_db"
	}
}

// interfaceSlice converts a slice of strings to a slice of interfaces
func interfaceSlice(s []string) []interface{} {
	interfaces := make([]interface{}, len(s))
	for i, v := range s {
		interfaces[i] = v
	}
	return interfaces
}

// Repository operations
func (c *CRUDOperationsImpl) CreateRepository(repo *Repository) (int64, error) {
	if repo == nil {
		return 0, fmt.Errorf("repository cannot be nil")
	}
	
	now := time.Now()
	repo.CreatedAt = now
	repo.UpdatedAt = now
	repo.FullName = repo.GetFullName()
	
	query := `
		INSERT INTO repositories (
			owner, name, full_name, description, url, language,
			stars, forks, issues_count, last_scraped, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := c.db.Exec(query,
		repo.Owner, repo.Name, repo.FullName, repo.Description, repo.URL,
		repo.Language, repo.Stars, repo.Forks, repo.IssuesCount,
		repo.LastScraped, repo.CreatedAt, repo.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create repository: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	
	repo.ID = int(id)
	return id, nil
}

func (c *CRUDOperationsImpl) GetRepository(id int64) (*Repository, error) {
	query := `
		SELECT id, owner, name, full_name, description, url, language,
			stars, forks, issues_count, last_scraped, created_at, updated_at
		FROM repositories WHERE id = ?
	`
	
	var repo Repository
	err := c.db.QueryRow(query, id).Scan(
		&repo.ID, &repo.Owner, &repo.Name, &repo.FullName, &repo.Description,
		&repo.URL, &repo.Language, &repo.Stars, &repo.Forks, &repo.IssuesCount,
		&repo.LastScraped, &repo.CreatedAt, &repo.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query repository: %w", err)
	}
	
	return &repo, nil
}

func (c *CRUDOperationsImpl) UpdateRepository(repo *Repository) error {
	if repo == nil {
		return fmt.Errorf("repository cannot be nil")
	}
	
	repo.UpdatedAt = time.Now()
	repo.FullName = repo.GetFullName()
	
	query := `
		UPDATE repositories SET 
			owner = ?, name = ?, full_name = ?, description = ?, url = ?,
			language = ?, stars = ?, forks = ?, issues_count = ?,
			last_scraped = ?, updated_at = ?
		WHERE id = ?
	`
	
	_, err := c.db.Exec(query,
		repo.Owner, repo.Name, repo.FullName, repo.Description, repo.URL,
		repo.Language, repo.Stars, repo.Forks, repo.IssuesCount,
		repo.LastScraped, repo.UpdatedAt, repo.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}
	
	return nil
}

func (c *CRUDOperationsImpl) DeleteRepository(id int64) error {
	query := "DELETE FROM repositories WHERE id = ?"
	
	result, err := c.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("repository not found")
	}
	
	return nil
}

func (c *CRUDOperationsImpl) GetAllRepositories(limit, offset int) ([]*Repository, error) {
	query := `
		SELECT id, owner, name, full_name, description, url, language,
			stars, forks, issues_count, last_scraped, created_at, updated_at
		FROM repositories
		ORDER BY stars DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := c.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query repositories: %w", err)
	}
	defer rows.Close()
	
	var repos []*Repository
	for rows.Next() {
		var repo Repository
		
		err := rows.Scan(
			&repo.ID, &repo.Owner, &repo.Name, &repo.FullName, &repo.Description,
			&repo.URL, &repo.Language, &repo.Stars, &repo.Forks, &repo.IssuesCount,
			&repo.LastScraped, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repository: %w", err)
		}
		
		repos = append(repos, &repo)
	}
	
	return repos, nil
}

func (c *CRUDOperationsImpl) GetRepositoryByName(owner, name string) (*Repository, error) {
	query := `
		SELECT id, owner, name, full_name, description, url, language,
			stars, forks, issues_count, last_scraped, created_at, updated_at
		FROM repositories
		WHERE owner = ? AND name = ?
	`
	
	var repo Repository
	err := c.db.QueryRow(query, owner, name).Scan(
		&repo.ID, &repo.Owner, &repo.Name, &repo.FullName, &repo.Description,
		&repo.URL, &repo.Language, &repo.Stars, &repo.Forks, &repo.IssuesCount,
		&repo.LastScraped, &repo.CreatedAt, &repo.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query repository: %w", err)
	}
	
	return &repo, nil
}

// Statistics methods
func (c *CRUDOperationsImpl) GetIssueStats() (*IssueStats, error) {
	stats := &IssueStats{
		ByCategory:      make(map[string]int),
		ByPriority:      make(map[string]int),
		ByTechStack:     make(map[string]int),
		ByState:         make(map[string]int),
		ScoreDistribution: make(map[string]int),
		TopKeywords:     make(map[string]int),
		DateRange:       make(map[string]interface{}),
	}
	
	// Get total count
	var totalCount int
	err := c.db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.TotalCount = totalCount
	
	// Get statistics by category
	rows, err := c.db.Query("SELECT category, COUNT(*) FROM issues GROUP BY category")
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			continue
		}
		stats.ByCategory[category] = count
	}
	
	// Get statistics by priority
	rows, err = c.db.Query("SELECT priority, COUNT(*) FROM issues GROUP BY priority")
	if err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			continue
		}
		stats.ByPriority[priority] = count
	}
	
	// Get statistics by state
	rows, err = c.db.Query("SELECT state, COUNT(*) FROM issues GROUP BY state")
	if err != nil {
		return nil, fmt.Errorf("failed to get state stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			continue
		}
		stats.ByState[state] = count
	}
	
	// Get score distribution
	rows, err = c.db.Query(`
		SELECT 
			CASE 
				WHEN score >= 25 THEN 'very_high'
				WHEN score >= 20 THEN 'high'
				WHEN score >= 15 THEN 'medium'
				WHEN score >= 10 THEN 'low'
				ELSE 'very_low'
			END as score_range,
			COUNT(*) as count
		FROM issues 
		GROUP BY score_range
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get score distribution: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var rangeName string
		var count int
		if err := rows.Scan(&rangeName, &count); err != nil {
			continue
		}
		stats.ScoreDistribution[rangeName] = count
	}
	
	// Get average score
	var avgScore float64
	err = c.db.QueryRow("SELECT AVG(score) FROM issues").Scan(&avgScore)
	if err == nil {
		stats.AverageScore = avgScore
	}
	
	return stats, nil
}

func (c *CRUDOperationsImpl) GetRepositoryStats(owner, name string) (*IssueStats, error) {
	stats := &IssueStats{
		ByCategory:      make(map[string]int),
		ByPriority:      make(map[string]int),
		ByTechStack:     make(map[string]int),
		ByState:         make(map[string]int),
		ScoreDistribution: make(map[string]int),
		TopKeywords:     make(map[string]int),
		DateRange:       make(map[string]interface{}),
	}
	
	// Get total count for this repository
	var totalCount int
	err := c.db.QueryRow("SELECT COUNT(*) FROM issues WHERE repo_owner = ? AND repo_name = ?", owner, name).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.TotalCount = totalCount
	
	// Get statistics by category for this repository
	rows, err := c.db.Query("SELECT category, COUNT(*) FROM issues WHERE repo_owner = ? AND repo_name = ? GROUP BY category", owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			continue
		}
		stats.ByCategory[category] = count
	}
	
	// Get statistics by priority for this repository
	rows, err = c.db.Query("SELECT priority, COUNT(*) FROM issues WHERE repo_owner = ? AND repo_name = ? GROUP BY priority", owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			continue
		}
		stats.ByPriority[priority] = count
	}
	
	// Similar logic for other stats...
	
	return stats, nil
}

// Maintenance methods
func (c *CRUDOperationsImpl) Cleanup() error {
	c.logger.Println("Starting database cleanup...")
	
	// Clean up old records
	cutoffDate := time.Now().AddDate(0, 0, -365) // 1 year ago
	result, err := c.db.Exec("DELETE FROM issues WHERE created_at_db < ? AND is_duplicate = 1", cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to clean up old duplicate records: %w", err)
	}
	
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get deleted rows count: %w", err)
	}
	
	c.logger.Printf("Cleaned up %d old duplicate records", deletedRows)
	return nil
}

func (c *CRUDOperationsImpl) Optimize() error {
	c.logger.Println("Starting database optimization...")
	
	// Analyze tables
	_, err := c.db.Exec("ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to analyze database: %w", err)
	}
	
	// Vacuum database
	_, err = c.db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}
	
	c.logger.Println("Database optimization completed")
	return nil
}