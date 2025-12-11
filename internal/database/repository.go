package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Repository defines the generic repository interface
type Repository interface {
	// Basic CRUD operations
	Create(entity interface{}) (int64, error)
	GetByID(id int64) (interface{}, error)
	Update(entity interface{}) error
	Delete(id int64) error
	
	// Query operations
	FindAll(limit, offset int) ([]interface{}, error)
	FindByFilter(filter interface{}, limit, offset int) ([]interface{}, error)
	CountByFilter(filter interface{}) (int, error)
	
	// Bulk operations
	BulkCreate(entities []interface{}) ([]int64, error)
	BulkUpdate(entities []interface{}) error
	BulkDelete(ids []int64) error
	
	// Utility operations
	Exists(id int64) (bool, error)
	Truncate() error
	GetTableName() string
}

// BaseRepository provides common repository functionality
type BaseRepository struct {
	db        *sql.DB
	tableName string
	logger    *log.Logger
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sql.DB, tableName string) *BaseRepository {
	return &BaseRepository{
		db:        db,
		tableName: tableName,
		logger:    log.New(log.Writer(), fmt.Sprintf("[%s] ", tableName), log.LstdFlags),
	}
}

// IssueRepository is the repository for Issue entities
type IssueRepository struct {
	*BaseRepository
}

// NewIssueRepository creates a new issue repository
func NewIssueRepository(db *sql.DB) *IssueRepository {
	return &IssueRepository{
		BaseRepository: NewBaseRepository(db, "issues"),
	}
}

// Create creates a new issue
func (r *IssueRepository) Create(entity interface{}) (int64, error) {
	issue, ok := entity.(*Issue)
	if !ok {
		return 0, fmt.Errorf("invalid entity type, expected *Issue")
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
	
	result, err := r.db.Exec(query,
		issue.Number, issue.Title, issue.Body, issue.URL, issue.State,
		issue.CreatedAt, issue.UpdatedAt, issue.Comments, issue.Reactions,
		issue.Assignee, issue.Milestone, issue.RepoOwner, issue.RepoName,
		issue.Keywords, issue.Score, issue.ContentHash, issue.Category,
		issue.Priority, issue.TechStack, issue.Labels, issue.IsDuplicate,
		issue.DuplicateOf, issue.CreatedAtDB, issue.UpdatedAtDB,
	)
	if err != nil {
		r.logger.Printf("Error creating issue: %v", err)
		return 0, fmt.Errorf("failed to create issue: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	
	issue.ID = int(id)
	r.logger.Printf("Created issue with ID: %d", id)
	return id, nil
}

// GetByID retrieves an issue by ID
func (r *IssueRepository) GetByID(id int64) (interface{}, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues WHERE id = ?
	`
	
	var issue Issue
	var keywords, techStack, labels StringArray
	
	err := r.db.QueryRow(query, id).Scan(
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

// Update updates an existing issue
func (r *IssueRepository) Update(entity interface{}) error {
	issue, ok := entity.(*Issue)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *Issue")
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
	
	_, err := r.db.Exec(query,
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
	
	r.logger.Printf("Updated issue with ID: %d", issue.ID)
	return nil
}

// Delete deletes an issue by ID
func (r *IssueRepository) Delete(id int64) error {
	query := "DELETE FROM issues WHERE id = ?"
	
	result, err := r.db.Exec(query, id)
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
	
	r.logger.Printf("Deleted issue with ID: %d", id)
	return nil
}

// FindAll retrieves all issues with pagination
func (r *IssueRepository) FindAll(limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		ORDER BY created_at_db DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.Query(query, limit, offset)
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
	
	return convertIssuesToInterface(issues), nil
}

// FindByFilter finds issues by filter criteria
func (r *IssueRepository) FindByFilter(filter interface{}, limit, offset int) ([]interface{}, error) {
	issueFilter, ok := filter.(*IssueFilter)
	if !ok {
		return nil, fmt.Errorf("invalid filter type, expected *IssueFilter")
	}
	
	query, params := r.buildFilterQuery(issueFilter, limit, offset)
	
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues with filter: %w", err)
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
	
	return convertIssuesToInterface(issues), nil
}

// CountByFilter counts issues by filter criteria
func (r *IssueRepository) CountByFilter(filter interface{}) (int, error) {
	issueFilter, ok := filter.(*IssueFilter)
	if !ok {
		return 0, fmt.Errorf("invalid filter type, expected *IssueFilter")
	}
	
	query, params := r.buildCountFilterQuery(issueFilter)
	
	var count int
	err := r.db.QueryRow(query, params...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count issues: %w", err)
	}
	
	return count, nil
}

// BulkCreate creates multiple issues
func (r *IssueRepository) BulkCreate(entities []interface{}) ([]int64, error) {
	if len(entities) == 0 {
		return nil, nil
	}
	
	issues := make([]*Issue, 0, len(entities))
	for _, entity := range entities {
		if issue, ok := entity.(*Issue); ok {
			issues = append(issues, issue)
		} else {
			return nil, fmt.Errorf("invalid entity type, expected *Issue")
		}
	}
	
	// Start transaction
	tx, err := r.db.Begin()
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
	
	r.logger.Printf("Bulk created %d issues", len(ids))
	return ids, nil
}

// BulkUpdate updates multiple issues
func (r *IssueRepository) BulkUpdate(entities []interface{}) error {
	if len(entities) == 0 {
		return nil
	}
	
	issues := make([]*Issue, 0, len(entities))
	for _, entity := range entities {
		if issue, ok := entity.(*Issue); ok {
			issues = append(issues, issue)
		} else {
			return fmt.Errorf("invalid entity type, expected *Issue")
		}
	}
	
	// Start transaction
	tx, err := r.db.Begin()
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
	
	for _, issue := range issues {
		issue.UpdatedAtDB = now
		
		_, err := stmt.Exec(
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
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	r.logger.Printf("Bulk updated %d issues", len(issues))
	return nil
}

// BulkDelete deletes multiple issues by ID
func (r *IssueRepository) BulkDelete(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	stmt, err := tx.Prepare("DELETE FROM issues WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	for _, id := range ids {
		_, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete issue: %w", err)
		}
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	r.logger.Printf("Bulk deleted %d issues", len(ids))
	return nil
}

// Exists checks if an issue exists by ID
func (r *IssueRepository) Exists(id int64) (bool, error) {
	query := "SELECT COUNT(*) FROM issues WHERE id = ?"
	var count int
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if issue exists: %w", err)
	}
	return count > 0, nil
}

// Truncate removes all issues from the table
func (r *IssueRepository) Truncate() error {
	query := "DELETE FROM issues"
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to truncate issues table: %w", err)
	}
	
	r.logger.Println("Truncated issues table")
	return nil
}

// GetTableName returns the table name
func (r *IssueRepository) GetTableName() string {
	return "issues"
}

// RepositoryRepository is the repository for Repository entities
type RepositoryRepository struct {
	*BaseRepository
}

// NewRepositoryRepository creates a new repository repository
func NewRepositoryRepository(db *sql.DB) *RepositoryRepository {
	return &RepositoryRepository{
		BaseRepository: NewBaseRepository(db, "repositories"),
	}
}

// Helper functions
func convertIssuesToInterface(issues []*Issue) []interface{} {
	result := make([]interface{}, len(issues))
	for i, issue := range issues {
		result[i] = issue
	}
	return result
}

// buildFilterQuery builds the SQL query and parameters for filtering
func (r *IssueRepository) buildFilterQuery(filter *IssueFilter, limit, offset int) (string, []interface{}) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues WHERE 1=1
	`
	
	var params []interface{}
	
	// Add filter conditions
	if filter.ID.Valid {
		query += " AND id = ?"
		params = append(params, filter.ID.Int64)
	}
	
	if filter.RepoOwner.Valid {
		query += " AND repo_owner = ?"
		params = append(params, filter.RepoOwner.String)
	}
	
	if filter.RepoName.Valid {
		query += " AND repo_name = ?"
		params = append(params, filter.RepoName.String)
	}
	
	if filter.Category.Valid {
		query += " AND category = ?"
		params = append(params, filter.Category.String)
	}
	
	if filter.Priority.Valid {
		query += " AND priority = ?"
		params = append(params, filter.Priority.String)
	}
	
	if filter.State.Valid {
		query += " AND state = ?"
		params = append(params, filter.State.String)
	}
	
	if filter.IsDuplicate.Valid {
		query += " AND is_duplicate = ?"
		params = append(params, filter.IsDuplicate.Bool)
	}
	
	if filter.MinScore.Valid {
		query += " AND score >= ?"
		params = append(params, filter.MinScore.Float64)
	}
	
	if filter.MaxScore.Valid {
		query += " AND score <= ?"
		params = append(params, filter.MaxScore.Float64)
	}
	
	if filter.DateFrom.Valid {
		query += " AND created_at >= ?"
		params = append(params, filter.DateFrom.Time)
	}
	
	if filter.DateTo.Valid {
		query += " AND created_at <= ?"
		params = append(params, filter.DateTo.Time)
	}
	
	// Add JSON array filters (keywords, tech_stack)
	if len(filter.Keywords) > 0 {
		for _, keyword := range filter.Keywords {
			query += " AND JSON_CONTAINS(keywords, ?)"
			params = append(params, `"`+keyword+`"`)
		}
	}
	
	if len(filter.TechStack) > 0 {
		for _, tech := range filter.TechStack {
			query += " AND JSON_CONTAINS(tech_stack, ?)"
			params = append(params, `"`+tech+`"`)
		}
	}
	
	// Add ordering and pagination
	query += " ORDER BY created_at_db DESC LIMIT ? OFFSET ?"
	params = append(params, limit, offset)
	
	return query, params
}

// buildCountFilterQuery builds the count query for filtering
func (r *IssueRepository) buildCountFilterQuery(filter *IssueFilter) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM issues WHERE 1=1"
	
	var params []interface{}
	
	// Add the same filter conditions as in buildFilterQuery
	if filter.ID.Valid {
		query += " AND id = ?"
		params = append(params, filter.ID.Int64)
	}
	
	if filter.RepoOwner.Valid {
		query += " AND repo_owner = ?"
		params = append(params, filter.RepoOwner.String)
	}
	
	if filter.RepoName.Valid {
		query += " AND repo_name = ?"
		params = append(params, filter.RepoName.String)
	}
	
	if filter.Category.Valid {
		query += " AND category = ?"
		params = append(params, filter.Category.String)
	}
	
	if filter.Priority.Valid {
		query += " AND priority = ?"
		params = append(params, filter.Priority.String)
	}
	
	if filter.State.Valid {
		query += " AND state = ?"
		params = append(params, filter.State.String)
	}
	
	if filter.IsDuplicate.Valid {
		query += " AND is_duplicate = ?"
		params = append(params, filter.IsDuplicate.Bool)
	}
	
	if filter.MinScore.Valid {
		query += " AND score >= ?"
		params = append(params, filter.MinScore.Float64)
	}
	
	if filter.MaxScore.Valid {
		query += " AND score <= ?"
		params = append(params, filter.MaxScore.Float64)
	}
	
	if filter.DateFrom.Valid {
		query += " AND created_at >= ?"
		params = append(params, filter.DateFrom.Time)
	}
	
	if filter.DateTo.Valid {
		query += " AND created_at <= ?"
		params = append(params, filter.DateTo.Time)
	}
	
	if len(filter.Keywords) > 0 {
		for _, keyword := range filter.Keywords {
			query += " AND JSON_CONTAINS(keywords, ?)"
			params = append(params, `"`+keyword+`"`)
		}
	}
	
	if len(filter.TechStack) > 0 {
		for _, tech := range filter.TechStack {
			query += " AND JSON_CONTAINS(tech_stack, ?)"
			params = append(params, `"`+tech+`"`)
		}
	}
	
	return query, params
}