package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Database represents the main database interface
type Database struct {
	db              *sql.DB
	crud            CRUDOperations
	deduplication   *DeduplicationService
	classification  *ClassificationService
	transaction     *TransactionManager
	config          DatabaseConfig
	logger          *log.Logger
}

// NewDatabase creates a new database instance
func NewDatabase(config DatabaseConfig) (*Database, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}
	
	// Open database connection
	db, err := sql.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure database connection
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections / 2)
	db.SetConnMaxLifetime(config.Timeout)
	
	// Configure SQLite-specific settings
	if err := configureSQLite(db, config); err != nil {
		return nil, fmt.Errorf("failed to configure SQLite: %w", err)
	}
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	logger := log.New(log.Writer(), "[Database] ", log.LstdFlags)
	
	// Create services
	crud := NewCRUDOperations(db)
	deduplication := NewDeduplicationService(db, DefaultDeduplicationConfig())
	classification := NewClassificationService(db, DefaultClassificationConfig())
	transaction := NewTransactionManager(db, DefaultTransactionConfig())
	
	return &Database{
		db:             db,
		crud:           crud,
		deduplication:  deduplication,
		classification: classification,
		transaction:    transaction,
		config:         config,
		logger:         logger,
	}, nil
}

// configureSQLite configures SQLite-specific settings
func configureSQLite(db *sql.DB, config DatabaseConfig) error {
	// Enable WAL mode for better concurrency
	if config.EnableWAL {
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			return fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}
	
	// Enable foreign key constraints
	if config.EnableForeignKeys {
		if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
			return fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}
	
	// Set busy timeout
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d;", int(config.BusyTimeout.Milliseconds()))); err != nil {
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}
	
	// Set synchronous mode
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	
	// Set cache size
	if _, err := db.Exec("PRAGMA cache_size=10000;"); err != nil {
		return fmt.Errorf("failed to set cache size: %w", err)
	}
	
	// Set temp store to memory
	if _, err := db.Exec("PRAGMA temp_store=memory;"); err != nil {
		return fmt.Errorf("failed to set temp store: %w", err)
	}
	
	return nil
}

// Initialize initializes the database by creating tables and indexes
func (db *Database) Initialize() error {
	db.logger.Println("Initializing database...")
	
	// Create tables
	if err := db.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	// Create indexes
	if err := db.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	// Create triggers
	if err := db.createTriggers(); err != nil {
		return fmt.Errorf("failed to create triggers: %w", err)
	}
	
	db.logger.Println("Database initialization completed")
	return nil
}

// createTables creates all necessary tables
func (db *Database) createTables() error {
	// Issues table
	if _, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			number INTEGER NOT NULL,
			title TEXT NOT NULL,
			body TEXT,
			url TEXT UNIQUE,
			state TEXT DEFAULT 'open',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			comments INTEGER DEFAULT 0,
			reactions INTEGER DEFAULT 0,
			assignee TEXT,
			milestone TEXT,
			repo_owner TEXT NOT NULL,
			repo_name TEXT NOT NULL,
			keywords TEXT DEFAULT '[]',
			score REAL DEFAULT 0.0,
			content_hash TEXT UNIQUE,
			category TEXT,
			priority TEXT,
			tech_stack TEXT DEFAULT '[]',
			labels TEXT DEFAULT '[]',
			is_duplicate BOOLEAN DEFAULT 0,
			duplicate_of INTEGER,
			created_at_db DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at_db DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (duplicate_of) REFERENCES issues(id) ON DELETE SET NULL
		)
	`); err != nil {
		return fmt.Errorf("failed to create issues table: %w", err)
	}
	
	// Repositories table
	if _, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner TEXT NOT NULL,
			name TEXT NOT NULL,
			full_name TEXT NOT NULL UNIQUE,
			description TEXT,
			url TEXT,
			language TEXT,
			stars INTEGER DEFAULT 0,
			forks INTEGER DEFAULT 0,
			issues_count INTEGER DEFAULT 0,
			last_scraped DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create repositories table: %w", err)
	}
	
	// Classification rules table
	if _, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS classification_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			pattern TEXT NOT NULL,
			category TEXT,
			priority TEXT,
			tech_stack TEXT DEFAULT '[]',
			weight REAL DEFAULT 1.0,
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create classification_rules table: %w", err)
	}
	
	// Transaction log table
	if _, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS transaction_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_id TEXT NOT NULL,
			operation_type TEXT NOT NULL,
			table_name TEXT NOT NULL,
			record_id INTEGER,
			data TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			success BOOLEAN DEFAULT 1
		)
	`); err != nil {
		return fmt.Errorf("failed to create transaction_log table: %w", err)
	}
	
	return nil
}

// createIndexes creates all necessary indexes
func (db *Database) createIndexes() error {
	indexes := []string{
		// Issues table indexes
		"CREATE INDEX IF NOT EXISTS idx_issues_github_id ON issues(number)",
		"CREATE INDEX IF NOT EXISTS idx_issues_repo ON issues(repo_owner, repo_name)",
		"CREATE INDEX IF NOT EXISTS idx_issues_score ON issues(score DESC)",
		"CREATE INDEX IF NOT EXISTS idx_issues_category ON issues(category)",
		"CREATE INDEX IF NOT EXISTS idx_issues_priority ON issues(priority)",
		"CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(state)",
		"CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_issues_updated_at ON issues(updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_issues_duplicate ON issues(is_duplicate, duplicate_of)",
		"CREATE INDEX IF NOT EXISTS idx_issues_content_hash ON issues(content_hash)",
		"CREATE INDEX IF NOT EXISTS idx_issues_keyword_search ON issues(keywords)",
		"CREATE INDEX IF NOT EXISTS idx_issues_tech_stack ON issues(tech_stack)",
		
		// Repositories table indexes
		"CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories(owner)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_name ON repositories(name)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_full_name ON repositories(full_name)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_language ON repositories(language)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_stars ON repositories(stars DESC)",
		"CREATE INDEX IF NOT EXISTS idx_repositories_last_scraped ON repositories(last_scraped)",
		
		// Classification rules indexes
		"CREATE INDEX IF NOT EXISTS idx_classification_rules_category ON classification_rules(category)",
		"CREATE INDEX IF NOT EXISTS idx_classification_rules_priority ON classification_rules(priority)",
		"CREATE INDEX IF NOT EXISTS idx_classification_rules_enabled ON classification_rules(enabled)",
		
		// Transaction log indexes
		"CREATE INDEX IF NOT EXISTS idx_transaction_log_transaction_id ON transaction_log(transaction_id)",
		"CREATE INDEX IF NOT EXISTS idx_transaction_log_timestamp ON transaction_log(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_transaction_log_table ON transaction_log(table_name)",
	}
	
	for _, indexSQL := range indexes {
		if _, err := db.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	
	return nil
}

// createTriggers creates all necessary database triggers
func (db *Database) createTriggers() error {
	triggers := []string{
		// Update timestamp trigger for issues
		`CREATE TRIGGER IF NOT EXISTS update_issues_timestamp 
			AFTER UPDATE ON issues
			FOR EACH ROW
			BEGIN
				UPDATE issues SET updated_at_db = CURRENT_TIMESTAMP WHERE id = NEW.id;
			END`,
		
		// Update timestamp trigger for repositories
		`CREATE TRIGGER IF NOT EXISTS update_repositories_timestamp 
			AFTER UPDATE ON repositories
			FOR EACH ROW
			BEGIN
				UPDATE repositories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
			END`,
		
		// Update timestamp trigger for classification rules
		`CREATE TRIGGER IF NOT EXISTS update_classification_rules_timestamp 
			AFTER UPDATE ON classification_rules
			FOR EACH ROW
			BEGIN
				UPDATE classification_rules SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
			END`,
		
		// Prevent circular duplicates
		`CREATE TRIGGER IF NOT EXISTS prevent_circular_duplicates 
			BEFORE INSERT ON issues
			FOR EACH ROW
			WHEN NEW.duplicate_of IS NOT NULL
			BEGIN
				SELECT 
					CASE 
						WHEN EXISTS (
							SELECT 1 FROM issues i 
							WHERE i.id = NEW.duplicate_of 
							AND i.duplicate_of = NEW.id
						) THEN RAISE(ABORT, 'Circular duplicate reference detected')
					END;
			END`,
		
		// Log transaction operations (simplified)
		`CREATE TRIGGER IF NOT EXISTS log_issue_inserts 
			AFTER INSERT ON issues
			FOR EACH ROW
			BEGIN
				INSERT INTO transaction_log (transaction_id, operation_type, table_name, record_id)
				VALUES ('manual', 'INSERT', 'issues', NEW.id);
			END`,
	}
	
	for _, triggerSQL := range triggers {
		if _, err := db.db.Exec(triggerSQL); err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}
	
	return nil
}

// CRUD returns the CRUD operations interface
func (db *Database) CRUD() CRUDOperations {
	return db.crud
}

// Deduplication returns the deduplication service
func (db *Database) Deduplication() *DeduplicationService {
	return db.deduplication
}

// Classification returns the classification service
func (db *Database) Classification() *ClassificationService {
	return db.classification
}

// Transaction returns the transaction manager
func (db *Database) Transaction() *TransactionManager {
	return db.transaction
}

// GetDB returns the underlying database connection
func (db *Database) GetDB() *sql.DB {
	return db.db
}

// Close closes the database connection
func (db *Database) Close() error {
	db.logger.Println("Closing database connection...")
	return db.db.Close()
}

// HealthCheck performs a health check on the database
func (db *Database) HealthCheck() error {
	if err := db.db.Ping(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	// Check if we can perform basic operations
	if _, err := db.db.Exec("SELECT 1"); err != nil {
		return fmt.Errorf("database health check query failed: %w", err)
	}
	
	return nil
}

// Backup creates a backup of the database
func (db *Database) Backup(backupPath string) error {
	db.logger.Printf("Creating database backup to %s", backupPath)
	
	// SQLite backup using VACUUM INTO
	if _, err := db.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath)); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	db.logger.Println("Database backup completed")
	return nil
}

// Restore restores the database from a backup
func (db *Database) Restore(backupPath string) error {
	db.logger.Printf("Restoring database from %s", backupPath)
	
	// Close current connection
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	
	// Copy backup file to main database file
	if err := copyFile(backupPath, db.config.Path); err != nil {
		return fmt.Errorf("failed to copy backup file: %w", err)
	}
	
	// Reopen database
	newDB, err := sql.Open("sqlite3", db.config.Path)
	if err != nil {
		return fmt.Errorf("failed to reopen database: %w", err)
	}
	
	// Update database connection
	db.db = newDB
	
	// Recreate services with new connection
	db.crud = NewCRUDOperations(db.db)
	db.deduplication = NewDeduplicationService(db.db, DefaultDeduplicationConfig())
	db.classification = NewClassificationService(db.db, DefaultClassificationConfig())
	db.transaction = NewTransactionManager(db.db, DefaultTransactionConfig())
	
	db.logger.Println("Database restore completed")
	return nil
}

// Optimize performs database optimization
func (db *Database) Optimize() error {
	db.logger.Println("Starting database optimization...")
	
	optimizations := []string{
		"VACUUM",
		"ANALYZE",
		"PRAGMA optimize",
	}
	
	for _, optimization := range optimizations {
		if _, err := db.db.Exec(optimization); err != nil {
			db.logger.Printf("Warning: %s failed: %v", optimization, err)
		}
	}
	
	db.logger.Println("Database optimization completed")
	return nil
}

// GetStats returns database statistics
func (db *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get table counts
	tables := []string{"issues", "repositories", "classification_rules", "transaction_log"}
	
	for _, table := range tables {
		var count int
		err := db.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get count for table %s: %w", table, err)
		}
		stats[table+"_count"] = count
	}
	
	// Get database size
	var dbSize int64
	err := db.db.QueryRow("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").Scan(&dbSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	stats["database_size_bytes"] = dbSize
	stats["database_size_mb"] = float64(dbSize) / (1024 * 1024)
	
	// Get free pages
	var freePages int
	err = db.db.QueryRow("SELECT freelist_count FROM pragma_freelist_count").Scan(&freePages)
	if err != nil {
		return nil, fmt.Errorf("failed to get free pages: %w", err)
	}
	stats["free_pages"] = freePages
	
	// Get WAL mode status
	var walMode string
	err = db.db.QueryRow("PRAGMA journal_mode").Scan(&walMode)
	if err == nil {
		stats["journal_mode"] = walMode
	}
	
	// Get foreign key status
	var foreignKeys int
	err = db.db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err == nil {
		stats["foreign_keys_enabled"] = foreignKeys == 1
	}
	
	return stats, nil
}

// copyFile copies a file (simplified implementation)
func copyFile(src, dst string) error {
	// This is a simplified implementation
	// In production, you'd want to use proper file copying with error handling
	return fmt.Errorf("file copying not implemented in this example")
}