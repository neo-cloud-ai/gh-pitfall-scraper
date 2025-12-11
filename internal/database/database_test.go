package database

import (
	"database/sql"
	"os"
	"testing"
	"time"
)

// TestDatabaseCreation tests database creation and initialization
func TestDatabaseCreation(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	// Test database initialization
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	// Test health check
	if err := tdb.HealthCheck(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

// TestIssueCRUD tests basic CRUD operations for issues
func TestIssueCRUD(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	// Initialize database
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	crud := tdb.CRUD()
	
	// Test Create
	issue := CreateTestIssue()
	id, err := crud.CreateIssue(issue)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero ID")
	}
	
	// Test GetByID
	retrievedIssue, err := crud.GetIssue(id)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}
	if retrievedIssue == nil {
		t.Error("Expected non-nil issue")
		return
	}
	
	// Verify fields
	if retrievedIssue.Number != issue.Number {
		t.Errorf("Expected number %d, got %d", issue.Number, retrievedIssue.Number)
	}
	if retrievedIssue.Title != issue.Title {
		t.Errorf("Expected title %s, got %s", issue.Title, retrievedIssue.Title)
	}
	
	// Test Update
	issue.Title = "Updated Test Issue Title"
	issue.Score = 20.0
	if err := crud.UpdateIssue(issue); err != nil {
		t.Fatalf("Failed to update issue: %v", err)
	}
	
	// Verify update
	updatedIssue, err := crud.GetIssue(id)
	if err != nil {
		t.Fatalf("Failed to get updated issue: %v", err)
	}
	if updatedIssue.Title != "Updated Test Issue Title" {
		t.Errorf("Expected updated title, got %s", updatedIssue.Title)
	}
	if updatedIssue.Score != 20.0 {
		t.Errorf("Expected updated score 20.0, got %f", updatedIssue.Score)
	}
	
	// Test Exists
	exists, err := crud.Exists(id)
	if err != nil {
		t.Fatalf("Failed to check if issue exists: %v", err)
	}
	if !exists {
		t.Error("Expected issue to exist")
	}
	
	// Test Delete
	if err := crud.DeleteIssue(id); err != nil {
		t.Fatalf("Failed to delete issue: %v", err)
	}
	
	// Verify deletion
	_, err = crud.GetIssue(id)
	if err == nil {
		t.Error("Expected error when getting deleted issue")
	}
	
	exists, err = crud.Exists(id)
	if err != nil {
		t.Fatalf("Failed to check if issue exists after deletion: %v", err)
	}
	if exists {
		t.Error("Expected issue to not exist after deletion")
	}
}

// TestIssueBatchOperations tests batch operations
func TestIssueBatchOperations(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	crud := tdb.CRUD()
	
	// Create test issues
	issues := make([]*Issue, 5)
	for i := 0; i < 5; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 100
		issue.Title = fmt.Sprintf("Batch Test Issue %d", i)
		issues[i] = issue
	}
	
	// Test BatchCreate
	ids, err := crud.CreateIssues(issues)
	if err != nil {
		t.Fatalf("Failed to batch create issues: %v", err)
	}
	if len(ids) != 5 {
		t.Errorf("Expected 5 IDs, got %d", len(ids))
	}
	
	// Verify creation
	for i, id := range ids {
		if id == 0 {
			t.Errorf("Expected non-zero ID for issue %d", i)
		}
	}
	
	// Test BatchUpdate
	for i, issue := range issues {
		issue.Score = float64(i) + 50.0
		issue.Category = "updated"
	}
	
	if err := crud.UpdateIssues(issues); err != nil {
		t.Fatalf("Failed to batch update issues: %v", err)
	}
	
	// Verify updates
	for i, issue := range issues {
		retrieved, err := crud.GetIssue(int64(ids[i]))
		if err != nil {
			t.Fatalf("Failed to get updated issue %d: %v", i, err)
		}
		if retrieved.Score != float64(i)+50.0 {
			t.Errorf("Expected score %f for issue %d, got %f", float64(i)+50.0, i, retrieved.Score)
		}
	}
	
	// Test BatchDelete
	if err := crud.DeleteIssues(ids); err != nil {
		t.Fatalf("Failed to batch delete issues: %v", err)
	}
	
	// Verify deletion
	for _, id := range ids {
		_, err := crud.GetIssue(id)
		if err == nil {
			t.Errorf("Expected error when getting deleted issue with ID %d", id)
		}
	}
}

// TestIssueQueryOperations tests query operations
func TestIssueQueryOperations(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	tr := NewTestRepository(tdb.GetDB())
	if err := tr.SeedTestData(); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
	
	crud := tdb.CRUD()
	
	// Test GetAllIssues
	issues, err := crud.GetAllIssues(10, 0)
	if err != nil {
		t.Fatalf("Failed to get all issues: %v", err)
	}
	if len(issues) == 0 {
		t.Error("Expected issues to be returned")
	}
	
	// Test GetIssuesByRepository
	repoIssues, err := crud.GetIssuesByRepository("test", "repo", 10, 0)
	if err != nil {
		t.Fatalf("Failed to get issues by repository: %v", err)
	}
	if len(repoIssues) == 0 {
		t.Error("Expected repository issues to be returned")
	}
	
	// Test GetIssuesByScore
	scoreIssues, err := crud.GetIssuesByScore(20.0, 30.0, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get issues by score: %v", err)
	}
	for _, issue := range scoreIssues {
		if issue.Score < 20.0 || issue.Score > 30.0 {
			t.Errorf("Issue score %f is outside range [20.0, 30.0]", issue.Score)
		}
	}
	
	// Test GetIssuesByCategory
	categoryIssues, err := crud.GetIssuesByCategory("performance", 10, 0)
	if err != nil {
		t.Fatalf("Failed to get issues by category: %v", err)
	}
	for _, issue := range categoryIssues {
		if issue.Category != "performance" {
			t.Errorf("Expected category 'performance', got %s", issue.Category)
		}
	}
	
	// Test GetIssuesByPriority
	priorityIssues, err := crud.GetIssuesByPriority("high", 10, 0)
	if err != nil {
		t.Fatalf("Failed to get issues by priority: %v", err)
	}
	for _, issue := range priorityIssues {
		if issue.Priority != "high" {
			t.Errorf("Expected priority 'high', got %s", issue.Priority)
		}
	}
	
	// Test GetIssuesByKeywords
	keywordIssues, err := crud.GetIssuesByKeywords([]string{"performance", "bug"}, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get issues by keywords: %v", err)
	}
	if len(keywordIssues) == 0 {
		t.Error("Expected issues with keywords to be returned")
	}
	
	// Test SearchIssues
	searchIssues, err := crud.SearchIssues("memory", 10, 0)
	if err != nil {
		t.Fatalf("Failed to search issues: %v", err)
	}
	if len(searchIssues) == 0 {
		t.Error("Expected search results to be returned")
	}
	
	// Test Advanced Search
	advancedSearch := &AdvancedSearch{
		Query:       "memory",
		Categories:  []string{"performance"},
		SortBy:      "score",
		SortOrder:   "DESC",
		Limit:       10,
		Offset:      0,
	}
	
	advancedIssues, err := crud.SearchIssuesAdvanced(advancedSearch)
	if err != nil {
		t.Fatalf("Failed to perform advanced search: %v", err)
	}
	if len(advancedIssues) == 0 {
		t.Error("Expected advanced search results to be returned")
	}
}

// TestRepositoryCRUD tests repository CRUD operations
func TestRepositoryCRUD(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	crud := tdb.CRUD()
	
	// Test Create
	repo := CreateTestRepository()
	id, err := crud.CreateRepository(repo)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero ID")
	}
	
	// Test GetByID
	retrievedRepo, err := crud.GetRepository(id)
	if err != nil {
		t.Fatalf("Failed to get repository: %v", err)
	}
	if retrievedRepo == nil {
		t.Error("Expected non-nil repository")
		return
	}
	
	// Verify fields
	if retrievedRepo.Owner != repo.Owner {
		t.Errorf("Expected owner %s, got %s", repo.Owner, retrievedRepo.Owner)
	}
	if retrievedRepo.Name != repo.Name {
		t.Errorf("Expected name %s, got %s", repo.Name, retrievedRepo.Name)
	}
	
	// Test GetByName
	repoByName, err := crud.GetRepositoryByName("test", "repo")
	if err != nil {
		t.Fatalf("Failed to get repository by name: %v", err)
	}
	if repoByName == nil {
		t.Error("Expected non-nil repository by name")
	}
	
	// Test Update
	repo.Stars = 150
	repo.Description = "Updated test repository"
	if err := crud.UpdateRepository(repo); err != nil {
		t.Fatalf("Failed to update repository: %v", err)
	}
	
	// Verify update
	updatedRepo, err := crud.GetRepository(id)
	if err != nil {
		t.Fatalf("Failed to get updated repository: %v", err)
	}
	if updatedRepo.Stars != 150 {
		t.Errorf("Expected stars 150, got %d", updatedRepo.Stars)
	}
	
	// Test Delete
	if err := crud.DeleteRepository(id); err != nil {
		t.Fatalf("Failed to delete repository: %v", err)
	}
	
	// Verify deletion
	_, err = crud.GetRepository(id)
	if err == nil {
		t.Error("Expected error when getting deleted repository")
	}
}

// TestTransactionManager tests transaction management
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
	
	// Test ExecuteInTransaction with success
	err = transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO issues (number, title, repo_owner, repo_name) VALUES (?, ?, ?, ?)",
			999, "Transaction Test Issue", "test", "repo")
		return err
	})
	if err != nil {
		t.Fatalf("Transaction execution failed: %v", err)
	}
	
	// Verify the issue was created
	var count int
	err = tdb.GetDB().QueryRow("SELECT COUNT(*) FROM issues WHERE number = 999").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify transaction: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 issue, got %d", count)
	}
	
	// Test ExecuteInTransaction with rollback
	err = transaction.ExecuteInTransaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO issues (number, title, repo_owner, repo_name) VALUES (?, ?, ?, ?)",
			998, "Rollback Test Issue", "test", "repo")
		if err != nil {
			return err
		}
		// Simulate an error to trigger rollback
		return sql.ErrTxDone
	})
	if err == nil {
		t.Error("Expected error from transaction")
	}
	
	// Verify the issue was not created (rollback)
	err = tdb.GetDB().QueryRow("SELECT COUNT(*) FROM issues WHERE number = 998").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify rollback: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 issues after rollback, got %d", count)
	}
}

// TestDeduplicationService tests deduplication functionality
func TestDeduplicationService(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	// Create test issues with similar content
	issues := make([]*Issue, 3)
	for i := 0; i < 3; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 300
		issue.Title = "Similar Memory Leak Issue" // Same title for all
		issue.ContentHash = "same_content_hash"  // Same content hash
		issues[i] = issue
	}
	
	crud := tdb.CRUD()
	_, err = crud.CreateIssues(issues)
	if err != nil {
		t.Fatalf("Failed to create test issues: %v", err)
	}
	
	deduplicator := tdb.Deduplication()
	
	// Test FindDuplicates
	result, err := deduplicator.FindDuplicates()
	if err != nil {
		t.Fatalf("Deduplication failed: %v", err)
	}
	
	if result.TotalProcessed != 3 {
		t.Errorf("Expected 3 processed issues, got %d", result.TotalProcessed)
	}
	
	// Test GetDuplicateStats
	stats, err := deduplicator.GetDuplicateStats()
	if err != nil {
		t.Fatalf("Failed to get duplicate stats: %v", err)
	}
	
	if stats["total_duplicates"] == nil {
		t.Error("Expected duplicate count in stats")
	}
}

// TestClassificationService tests classification functionality
func TestClassificationService(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	// Create test issues
	issues := make([]*Issue, 3)
	
	// Bug issue
	issues[0] = CreateTestIssue()
	issues[0].Number = 400
	issues[0].Title = "Bug in the authentication system"
	issues[0].Body = "There is a bug causing the authentication to fail"
	
	// Performance issue
	issues[1] = CreateTestIssue()
	issues[1].Number = 401
	issues[1].Title = "Memory leak causing performance issues"
	issues[1].Body = "The application has a memory leak that affects performance"
	
	// Security issue
	issues[2] = CreateTestIssue()
	issues[2].Number = 402
	issues[2].Title = "Security vulnerability in login"
	issues[2].Body = "There is a security vulnerability in the login module"
	
	crud := tdb.CRUD()
	_, err = crud.CreateIssues(issues)
	if err != nil {
		t.Fatalf("Failed to create test issues: %v", err)
	}
	
	classifier := tdb.Classification()
	
	// Test ClassifySingleIssue
	result, err := classifier.ClassifySingleIssue(issues[0])
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}
	
	if result == nil {
		t.Error("Expected classification result")
	}
	
	// Test ClassifyIssues
	stats, err := classifier.ClassifyIssues(issues)
	if err != nil {
		t.Fatalf("Batch classification failed: %v", err)
	}
	
	if stats.TotalProcessed != 3 {
		t.Errorf("Expected 3 processed issues, got %d", stats.TotalProcessed)
	}
	
	// Test GetClassificationStats
	clsStats, err := classifier.GetClassificationStats()
	if err != nil {
		t.Fatalf("Failed to get classification stats: %v", err)
	}
	
	if clsStats["categories"] == nil {
		t.Error("Expected categories in classification stats")
	}
}

// TestDatabaseOptimization tests database optimization features
func TestDatabaseOptimization(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	// Test health check
	if err := tdb.HealthCheck(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	
	// Test GetStats
	stats, err := tdb.GetStats()
	if err != nil {
		t.Fatalf("Failed to get database stats: %v", err)
	}
	
	if stats["issues_count"] == nil {
		t.Error("Expected issues count in stats")
	}
	
	if stats["repositories_count"] == nil {
		t.Error("Expected repositories count in stats")
	}
	
	if stats["database_size_mb"] == nil {
		t.Error("Expected database size in stats")
	}
	
	// Test Optimize
	if err := tdb.Optimize(); err != nil {
		t.Fatalf("Database optimization failed: %v", err)
	}
}

// TestStringArray tests StringArray functionality
func TestStringArray(t *testing.T) {
	// Test Value method
	arr := StringArray{"go", "database", "testing"}
	value, err := arr.Value()
	if err != nil {
		t.Fatalf("StringArray Value failed: %v", err)
	}
	
	// Test Scan method
	var scannedArr StringArray
	err = scannedArr.Scan(value)
	if err != nil {
		t.Fatalf("StringArray Scan failed: %v", err)
	}
	
	if len(scannedArr) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(scannedArr))
	}
	
	if scannedArr[0] != "go" || scannedArr[1] != "database" || scannedArr[2] != "testing" {
		t.Errorf("Expected [go, database, testing], got %v", scannedArr)
	}
}

// TestSimilarityEngine tests similarity calculation
func TestSimilarityEngine(t *testing.T) {
	engine := NewSimilarityEngine()
	
	// Test identical issues
	issue1 := CreateTestIssue()
	issue2 := CreateTestIssue()
	
	similarity := engine.CalculateSimilarity(issue1, issue2)
	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical issues, got %f", similarity)
	}
	
	// Test completely different issues
	issue1.Title = "Memory leak in cache"
	issue1.Body = "There is a memory leak"
	issue2.Title = "Add new feature for user authentication"
	issue2.Body = "We need to add authentication"
	
	similarity = engine.CalculateSimilarity(issue1, issue2)
	if similarity >= 0.5 {
		t.Errorf("Expected low similarity for different issues, got %f", similarity)
	}
}

// TestIssueComparison tests issue comparison functionality
func TestIssueComparison(t *testing.T) {
	issue1 := CreateTestIssue()
	issue2 := CreateTestIssue()
	
	// Test identical issues
	similarity := issue1.IsSimilar(issue2)
	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical issues, got %f", similarity)
	}
	
	// Test different issues
	issue2.Title = "Completely different issue title"
	issue2.Body = "Different body content"
	
	similarity = issue1.IsSimilar(issue2)
	if similarity >= 0.8 {
		t.Errorf("Expected low similarity for very different issues, got %f", similarity)
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	// Test valid config
	validConfig := DefaultDatabaseConfig()
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should not produce error: %v", err)
	}
	
	// Test invalid config
	invalidConfig := DatabaseConfig{
		Path: "", // Invalid: empty path
	}
	if err := invalidConfig.Validate(); err == nil {
		t.Error("Invalid config should produce error")
	}
}

// TestCleanup tests cleanup functionality
func TestCleanup(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	crud := tdb.CRUD()
	
	// Create some issues
	issues := make([]*Issue, 3)
	for i := 0; i < 3; i++ {
		issue := CreateTestIssue()
		issue.Number = i + 500
		issue.CreatedAtDB = time.Now().Add(-2 * 365 * 24 * time.Hour) // 2 years ago
		issues[i] = issue
	}
	
	_, err = crud.CreateIssues(issues)
	if err != nil {
		t.Fatalf("Failed to create test issues: %v", err)
	}
	
	// Test Cleanup
	if err := crud.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// TestFileOperations tests file operations
func TestFileOperations(t *testing.T) {
	tdb, err := TestSetup()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer TestTeardown(tdb)
	
	if err := tdb.Initialize(); err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}
	
	// Create a backup
	backupPath := tdb.tmpDir + "/backup.db"
	if err := tdb.Backup(backupPath); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}
	
	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file should exist")
	}
}