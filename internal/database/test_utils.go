package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDatabase provides a test database for unit tests
type TestDatabase struct {
	*Database
	tmpDir string
	dbPath string
}

// NewTestDatabase creates a new test database
func NewTestDatabase() (*TestDatabase, error) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test_db_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	dbPath := filepath.Join(tmpDir, "test.db")
	
	// Create test configuration
	config := DefaultDatabaseConfig()
	config.Path = dbPath
	config.MaxConnections = 1 // Use single connection for tests
	
	// Create database
	db, err := NewDatabase(config)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}
	
	// Initialize database
	if err := db.Initialize(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to initialize test database: %w", err)
	}
	
	return &TestDatabase{
		Database: db,
		tmpDir:   tmpDir,
		dbPath:   dbPath,
	}, nil
}

// Close closes and cleans up the test database
func (tdb *TestDatabase) Close() error {
	if tdb.Database != nil {
		if err := tdb.Database.Close(); err != nil {
			return err
		}
	}
	
	// Clean up temporary files
	return os.RemoveAll(tdb.tmpDir)
}

// CreateTestIssue creates a test issue for testing
func CreateTestIssue() *Issue {
	return &Issue{
		Number:      1,
		Title:       "Test Issue Title",
		Body:        "This is a test issue body with some content for testing purposes.",
		URL:         "https://github.com/test/repo/issues/1",
		State:       "open",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		Comments:    5,
		Reactions:   3,
		Assignee:    "testuser",
		Milestone:   "v1.0",
		RepoOwner:   "test",
		RepoName:    "repo",
		Keywords:    StringArray{"test", "bug", "performance"},
		Score:       15.5,
		ContentHash: "test_hash_123",
		Category:    "bug",
		Priority:    "medium",
		TechStack:   StringArray{"go", "database"},
		Labels:      StringArray{"bug", "performance"},
		IsDuplicate: false,
		CreatedAtDB: time.Now(),
		UpdatedAtDB: time.Now(),
	}
}

// CreateTestRepository creates a test repository for testing
func CreateTestRepository() *Repository {
	return &Repository{
		Owner:       "test",
		Name:        "repo",
		FullName:    "test/repo",
		Description: "A test repository for testing purposes",
		URL:         "https://github.com/test/repo",
		Language:    "Go",
		Stars:       100,
		Forks:       20,
		IssuesCount: 50,
		LastScraped: time.Now().Add(-1 * time.Hour),
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now(),
	}
}

// TestSetup sets up the test environment
func TestSetup() (*TestDatabase, error) {
	return NewTestDatabase()
}

// TestTeardown tears down the test environment
func TestTeardown(tdb *TestDatabase) {
	if tdb != nil {
		if err := tdb.Close(); err != nil {
			log.Printf("Warning: failed to close test database: %v", err)
		}
	}
}

// TestRepository provides test utilities for database operations
type TestRepository struct {
	db       *sql.DB
	issues   []*Issue
	repos    []*Repository
	logger   *log.Logger
}

// NewTestRepository creates a new test repository
func NewTestRepository(db *sql.DB) *TestRepository {
	return &TestRepository{
		db:     db,
		issues: make([]*Issue, 0),
		repos:  make([]*Repository, 0),
		logger: log.New(log.Writer(), "[TestRepo] ", log.LstdFlags),
	}
}

// SeedTestData seeds the database with test data
func (tr *TestRepository) SeedTestData() error {
	// Create test repositories
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
			LastScraped: time.Now().Add(-30 * time.Minute),
			CreatedAt:   time.Now().Add(-10 * 365 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			Owner:       "facebook",
			Name:        "react",
			FullName:    "facebook/react",
			Description: "A declarative, efficient, and flexible JavaScript library for building user interfaces",
			URL:         "https://github.com/facebook/react",
			Language:    "JavaScript",
			Stars:       220000,
			Forks:       45000,
			IssuesCount: 8000,
			LastScraped: time.Now().Add(-1 * time.Hour),
			CreatedAt:   time.Now().Add(-15 * 365 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}
	
	// Insert repositories
	for _, repo := range repos {
		stmt, err := tr.db.Prepare(`
			INSERT INTO repositories (
				owner, name, full_name, description, url, language,
				stars, forks, issues_count, last_scraped, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare repository insert: %w", err)
		}
		
		result, err := stmt.Exec(
			repo.Owner, repo.Name, repo.FullName, repo.Description, repo.URL,
			repo.Language, repo.Stars, repo.Forks, repo.IssuesCount,
			repo.LastScraped, repo.CreatedAt, repo.UpdatedAt,
		)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("failed to insert repository: %w", err)
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			stmt.Close()
			return fmt.Errorf("failed to get repository ID: %w", err)
		}
		
		repo.ID = int(id)
		tr.repos = append(tr.repos, repo)
		stmt.Close()
	}
	
	// Create test issues
	issues := []*Issue{
		CreateTestIssue(),
		{
			Number:      2,
			Title:       "Memory leak in cache implementation",
			Body:        "There is a memory leak in the cache implementation that causes memory usage to grow over time.",
			URL:         "https://github.com/test/repo/issues/2",
			State:       "open",
			CreatedAt:   time.Now().Add(-2 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			Comments:    8,
			Reactions:   12,
			Assignee:    "developer",
			Milestone:   "v1.1",
			RepoOwner:   "test",
			RepoName:    "repo",
			Keywords:    StringArray{"memory", "leak", "cache", "performance"},
			Score:       25.0,
			ContentHash: "memory_leak_hash",
			Category:    "performance",
			Priority:    "high",
			TechStack:   StringArray{"go", "memory"},
			Labels:      StringArray{"performance", "bug"},
			IsDuplicate: false,
			CreatedAtDB: time.Now(),
			UpdatedAtDB: time.Now(),
		},
		{
			Number:      3,
			Title:       "Add support for PostgreSQL database",
			Body:        "We need to add support for PostgreSQL database to make the application more versatile.",
			URL:         "https://github.com/test/repo/issues/3",
			State:       "open",
			CreatedAt:   time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
			Comments:    3,
			Reactions:   7,
			Assignee:    "",
			Milestone:   "v1.2",
			RepoOwner:   "test",
			RepoName:    "repo",
			Keywords:    StringArray{"database", "postgresql", "feature"},
			Score:       18.0,
			ContentHash: "postgresql_feature_hash",
			Category:    "feature",
			Priority:    "medium",
			TechStack:   StringArray{"go", "database", "postgresql"},
			Labels:      StringArray{"feature", "database"},
			IsDuplicate: false,
			CreatedAtDB: time.Now(),
			UpdatedAtDB: time.Now(),
		},
		{
			Number:      4,
			Title:       "Security vulnerability in authentication",
			Body:        "There is a security vulnerability in the authentication module that could allow unauthorized access.",
			URL:         "https://github.com/test/repo/issues/4",
			State:       "closed",
			CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
			Comments:    15,
			Reactions:   25,
			Assignee:    "security",
			Milestone:   "v1.0.1",
			RepoOwner:   "test",
			RepoName:    "repo",
			Keywords:    StringArray{"security", "vulnerability", "authentication"},
			Score:       30.0,
			ContentHash: "security_vuln_hash",
			Category:    "security",
			Priority:    "critical",
			TechStack:   StringArray{"go", "security"},
			Labels:      StringArray{"security", "critical"},
			IsDuplicate: false,
			CreatedAtDB: time.Now(),
			UpdatedAtDB: time.Now(),
		},
		{
			Number:      5,
			Title:       "Slow response time on large datasets",
			Body:        "The application becomes very slow when processing large datasets, especially with more than 10,000 records.",
			URL:         "https://github.com/test/repo/issues/5",
			State:       "open",
			CreatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-3 * time.Hour),
			Comments:    6,
			Reactions:   9,
			Assignee:    "performance",
			Milestone:   "v1.1",
			RepoOwner:   "test",
			RepoName:    "repo",
			Keywords:    StringArray{"performance", "slow", "dataset", "optimization"},
			Score:       22.0,
			ContentHash: "performance_issue_hash",
			Category:    "performance",
			Priority:    "high",
			TechStack:   StringArray{"go", "database", "performance"},
			Labels:      StringArray{"performance", "optimization"},
			IsDuplicate: false,
			CreatedAtDB: time.Now(),
			UpdatedAtDB: time.Now(),
		},
		{
			Number:      6,
			Title:       "Add comprehensive API documentation",
			Body:        "We need to add comprehensive API documentation including examples and usage patterns.",
			URL:         "https://github.com/test/repo/issues/6",
			State:       "open",
			CreatedAt:   time.Now().Add(-1 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * time.Minute),
			Comments:    2,
			Reactions:   4,
			Assignee:    "docs",
			Milestone:   "v1.2",
			RepoOwner:   "test",
			RepoName:    "repo",
			Keywords:    StringArray{"documentation", "api", "docs"},
			Score:       12.0,
			ContentHash: "docs_issue_hash",
			Category:    "documentation",
			Priority:    "low",
			TechStack:   StringArray{"documentation"},
			Labels:      StringArray{"documentation"},
			IsDuplicate: false,
			CreatedAtDB: time.Now(),
			UpdatedAtDB: time.Now(),
		},
	}
	
	// Insert issues
	for _, issue := range issues {
		stmt, err := tr.db.Prepare(`
			INSERT INTO issues (
				number, title, body, url, state, created_at, updated_at, comments,
				reactions, assignee, milestone, repo_owner, repo_name, keywords,
				score, content_hash, category, priority, tech_stack, labels,
				is_duplicate, duplicate_of, created_at_db, updated_at_db
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare issue insert: %w", err)
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
			stmt.Close()
			return fmt.Errorf("failed to insert issue: %w", err)
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			stmt.Close()
			return fmt.Errorf("failed to get issue ID: %w", err)
		}
		
		issue.ID = int(id)
		tr.issues = append(tr.issues, issue)
		stmt.Close()
	}
	
	tr.logger.Printf("Seeded test data: %d repositories, %d issues", len(tr.repos), len(tr.issues))
	return nil
}

// GetIssues returns the seeded test issues
func (tr *TestRepository) GetIssues() []*Issue {
	return tr.issues
}

// GetRepositories returns the seeded test repositories
func (tr *TestRepository) GetRepositories() []*Repository {
	return tr.repos
}

// GetIssueByNumber returns an issue by its number
func (tr *TestRepository) GetIssueByNumber(number int) *Issue {
	for _, issue := range tr.issues {
		if issue.Number == number {
			return issue
		}
	}
	return nil
}

// GetRepositoryByName returns a repository by its name
func (tr *TestRepository) GetRepositoryByName(owner, name string) *Repository {
	for _, repo := range tr.repos {
		if repo.Owner == owner && repo.Name == name {
			return repo
		}
	}
	return nil
}

// AssertIssueCount asserts the number of issues in the database
func (tr *TestRepository) AssertIssueCount(t *testing.T, expected int) {
	var actual int
	err := tr.db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&actual)
	if err != nil {
		t.Fatalf("Failed to count issues: %v", err)
	}
	if actual != expected {
		t.Errorf("Expected %d issues, got %d", expected, actual)
	}
}

// AssertRepositoryCount asserts the number of repositories in the database
func (tr *TestRepository) AssertRepositoryCount(t *testing.T, expected int) {
	var actual int
	err := tr.db.QueryRow("SELECT COUNT(*) FROM repositories").Scan(&actual)
	if err != nil {
		t.Fatalf("Failed to count repositories: %v", err)
	}
	if actual != expected {
		t.Errorf("Expected %d repositories, got %d", expected, actual)
	}
}

// ClearTestData clears all test data from the database
func (tr *TestRepository) ClearTestData() error {
	_, err := tr.db.Exec("DELETE FROM issues")
	if err != nil {
		return fmt.Errorf("failed to clear issues: %w", err)
	}
	
	_, err = tr.db.Exec("DELETE FROM repositories")
	if err != nil {
		return fmt.Errorf("failed to clear repositories: %w", err)
	}
	
	tr.issues = tr.issues[:0]
	tr.repos = tr.repos[:0]
	
	tr.logger.Println("Cleared test data")
	return nil
}

// BenchmarkCreateIssue benchmarks the CreateIssue operation
func BenchmarkCreateIssue(b *testing.B, tdb *TestDatabase) {
	issue := CreateTestIssue()
	crud := tdb.CRUD()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		issue.Number = i + 1000 // Ensure unique number
		_, err := crud.CreateIssue(issue)
		if err != nil {
			b.Fatalf("Failed to create issue: %v", err)
		}
	}
}

// BenchmarkQueryIssues benchmarks querying issues
func BenchmarkQueryIssues(b *testing.B, tdb *TestDatabase) {
	crud := tdb.CRUD()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		issues, err := crud.GetAllIssues(100, 0)
		if err != nil {
			b.Fatalf("Failed to query issues: %v", err)
		}
		_ = issues // Use the result to avoid optimization
	}
}

// BenchmarkDeduplication benchmarks the deduplication process
func BenchmarkDeduplication(b *testing.B, tdb *TestDatabase) {
	// Create a large number of test issues
	issues := make([]*Issue, 1000)
	for i := 0; i < 1000; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 2000
		issue.Title = fmt.Sprintf("Test Issue %d with similar content", i%10)
		issue.ContentHash = fmt.Sprintf("hash_%d", i%10) // Create duplicates
		issues[i] = issue
	}
	
	crud := tdb.CRUD()
	_, err := crud.CreateIssues(issues)
	if err != nil {
		b.Fatalf("Failed to create test issues: %v", err)
	}
	
	deduplicator := tdb.Deduplication()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := deduplicator.FindDuplicates()
		if err != nil {
			b.Fatalf("Failed to perform deduplication: %v", err)
		}
	}
}