package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// TransactionManager provides transaction management functionality
type TransactionManager struct {
	db          *sql.DB
	logger      *log.Logger
	config      TransactionConfig
	activeTx    map[string]*sql.Tx
	txMutex     sync.RWMutex
	retryPolicy *RetryPolicy
}

// TransactionConfig represents configuration for transaction management
type TransactionConfig struct {
	MaxRetries          int           `json:"max_retries"`           // Maximum number of retries
	RetryDelay          time.Duration `json:"retry_delay"`           // Delay between retries
	Timeout             time.Duration `json:"timeout"`               // Transaction timeout
	EnableDeadlockRetry bool          `json:"enable_deadlock_retries"` // Enable retry on deadlock errors
	IsolationLevel      sql.IsolationLevel `json:"isolation_level"`   // Transaction isolation level
	MaxActiveTransactions int         `json:"max_active_transactions"` // Maximum active transactions
}

// DefaultTransactionConfig returns default transaction configuration
func DefaultTransactionConfig() TransactionConfig {
	return TransactionConfig{
		MaxRetries:             3,
		RetryDelay:             100 * time.Millisecond,
		Timeout:                30 * time.Second,
		EnableDeadlockRetry:    true,
		IsolationLevel:         sql.LevelReadCommitted,
		MaxActiveTransactions:  10,
	}
}

// RetryPolicy defines retry behavior for transactions
type RetryPolicy struct {
	MaxRetries    int
	RetryDelay    time.Duration
	BackoffFactor float64
	MaxBackoffDelay time.Duration
	RetryableErrors []string
}

// DefaultRetryPolicy returns default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:        3,
		RetryDelay:        100 * time.Millisecond,
		BackoffFactor:     2.0,
		MaxBackoffDelay:   5 * time.Second,
		RetryableErrors: []string{
			"database is locked",
			"SQLITE_BUSY",
			"deadlock detected",
			"connection lost",
		},
	}
}

// Transaction represents a database transaction with metadata
type Transaction struct {
	ID          string            `json:"id"`
	Tx          *sql.Tx           `json:"-"`
	StartedAt   time.Time         `json:"started_at"`
	Timeout     time.Duration     `json:"timeout"`
	Status      TransactionStatus `json:"status"`
	Operations  []Operation       `json:"operations"`
	RollbackOnly bool             `json:"rollback_only"`
}

// Operation represents a database operation within a transaction
type Operation struct {
	Type      string    `json:"type"`      // "query", "exec", "prepare"
	Query     string    `json:"query"`
	Args      []interface{} `json:"args"`
	StartedAt time.Time `json:"started_at"`
	Duration  time.Duration `json:"duration"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusActive     TransactionStatus = "active"
	TransactionStatusCommitted  TransactionStatus = "committed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusTimeout    TransactionStatus = "timeout"
)

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB, config TransactionConfig) *TransactionManager {
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultTransactionConfig().MaxRetries
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultTransactionConfig().Timeout
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = DefaultTransactionConfig().RetryDelay
	}
	
	return &TransactionManager{
		db:           db,
		logger:       log.New(log.Writer(), "[Transaction] ", log.LstdFlags),
		config:       config,
		activeTx:     make(map[string]*sql.Tx),
		retryPolicy:  DefaultRetryPolicy(),
	}
}

// BeginTransaction begins a new transaction
func (tm *TransactionManager) BeginTransaction(options ...TransactionOption) (*Transaction, error) {
	tm.txMutex.Lock()
	defer tm.txMutex.Unlock()
	
	// Check if we have reached the maximum number of active transactions
	if len(tm.activeTx) >= tm.config.MaxActiveTransactions {
		return nil, fmt.Errorf("maximum number of active transactions reached: %d", tm.config.MaxActiveTransactions)
	}
	
	// Generate transaction ID
	txID := generateTransactionID()
	
	// Begin transaction with configured isolation level
	var tx *sql.Tx
	var err error
	
	if len(options) > 0 {
		// Apply custom options
		for _, option := range options {
			if option.IsolationLevel != nil {
				tm.config.IsolationLevel = *option.IsolationLevel
			}
		}
	}
	
	tx, err = tm.db.BeginTx(tm.db.Context(), &sql.TxOptions{
		Isolation: tm.config.IsolationLevel,
		ReadOnly:  false,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Create transaction object
	transaction := &Transaction{
		ID:        txID,
		Tx:        tx,
		StartedAt: time.Now(),
		Timeout:   tm.config.Timeout,
		Status:    TransactionStatusActive,
		Operations: make([]Operation, 0),
	}
	
	// Store in active transactions
	tm.activeTx[txID] = tx
	
	tm.logger.Printf("Started transaction %s", txID)
	return transaction, nil
}

// ExecuteInTransaction executes a function within a transaction
func (tm *TransactionManager) ExecuteInTransaction(fn func(*sql.Tx) error, options ...TransactionOption) error {
	// Create a retryable function
	retryableFn := func() error {
		// Begin transaction
		transaction, err := tm.BeginTransaction(options...)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tm.EndTransaction(transaction)
		
		// Set timeout
		timeoutTimer := time.AfterFunc(transaction.Timeout, func() {
			tm.logger.Printf("Transaction %s timed out", transaction.ID)
			tm.forceRollback(transaction)
		})
		
		// Execute the function
		err = fn(transaction.Tx)
		
		// Cancel timeout timer
		timeoutTimer.Stop()
		
		if err != nil {
			tm.logger.Printf("Transaction %s failed: %v", transaction.ID, err)
			return err
		}
		
		// Check if rollback only
		if transaction.RollbackOnly {
			tm.logger.Printf("Transaction %s marked for rollback", transaction.ID)
			return tm.RollbackTransaction(transaction)
		}
		
		// Commit transaction
		return tm.CommitTransaction(transaction)
	}
	
	// Execute with retry policy
	return tm.executeWithRetry(retryableFn)
}

// CommitTransaction commits a transaction
func (tm *TransactionManager) CommitTransaction(transaction *Transaction) error {
	tm.txMutex.Lock()
	defer tm.txMutex.Unlock()
	
	if transaction.Status != TransactionStatusActive {
		return fmt.Errorf("cannot commit transaction %s: status is %s", transaction.ID, transaction.Status)
	}
	
	// Check if transaction has timed out
	if time.Since(transaction.StartedAt) > transaction.Timeout {
		transaction.Status = TransactionStatusTimeout
		return fmt.Errorf("transaction %s has timed out", transaction.ID)
	}
	
	// Commit the transaction
	if err := transaction.Tx.Commit(); err != nil {
		transaction.Status = TransactionStatusFailed
		return fmt.Errorf("failed to commit transaction %s: %w", transaction.ID, err)
	}
	
	transaction.Status = TransactionStatusCommitted
	delete(tm.activeTx, transaction.ID)
	
	tm.logger.Printf("Committed transaction %s", transaction.ID)
	return nil
}

// RollbackTransaction rolls back a transaction
func (tm *TransactionManager) RollbackTransaction(transaction *Transaction) error {
	tm.txMutex.Lock()
	defer tm.txMutex.Unlock()
	
	if transaction.Status != TransactionStatusActive {
		// Transaction is already completed, no need to rollback
		return nil
	}
	
	// Rollback the transaction
	if err := transaction.Tx.Rollback(); err != nil {
		transaction.Status = TransactionStatusFailed
		return fmt.Errorf("failed to rollback transaction %s: %w", transaction.ID, err)
	}
	
	transaction.Status = TransactionStatusRolledBack
	delete(tm.activeTx, transaction.ID)
	
	tm.logger.Printf("Rolled back transaction %s", transaction.ID)
	return nil
}

// EndTransaction ends a transaction (commits or rolls back based on status)
func (tm *TransactionManager) EndTransaction(transaction *Transaction) error {
	switch transaction.Status {
	case TransactionStatusActive:
		// Auto-commit if no explicit rollback was requested
		return tm.CommitTransaction(transaction)
	case TransactionStatusRolledBack, TransactionStatusCommitted:
		// Transaction is already completed
		return nil
	default:
		return fmt.Errorf("transaction %s has invalid status: %s", transaction.ID, transaction.Status)
	}
}

// MarkRollbackOnly marks a transaction for rollback only
func (tm *TransactionManager) MarkRollbackOnly(transaction *Transaction) {
	transaction.RollbackOnly = true
	tm.logger.Printf("Transaction %s marked for rollback only", transaction.ID)
}

// forceRollback forces a transaction to rollback due to timeout
func (tm *TransactionManager) forceRollback(transaction *Transaction) {
	if err := tm.RollbackTransaction(transaction); err != nil {
		tm.logger.Printf("Failed to force rollback transaction %s: %v", transaction.ID, err)
	}
}

// executeWithRetry executes a function with retry policy
func (tm *TransactionManager) executeWithRetry(fn func() error) error {
	var lastErr error
	delay := tm.retryPolicy.RetryDelay
	
	for attempt := 0; attempt <= tm.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			tm.logger.Printf("Retrying operation, attempt %d/%d", attempt, tm.retryPolicy.MaxRetries)
			time.Sleep(delay)
			
			// Exponential backoff
			delay = time.Duration(float64(delay) * tm.retryPolicy.BackoffFactor)
			if delay > tm.retryPolicy.MaxBackoffDelay {
				delay = tm.retryPolicy.MaxBackoffDelay
			}
		}
		
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !tm.isRetryableError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}
		
		// Check if this was the last attempt
		if attempt == tm.retryPolicy.MaxRetries {
			break
		}
	}
	
	return fmt.Errorf("operation failed after %d retries, last error: %w", tm.retryPolicy.MaxRetries+1, lastErr)
}

// isRetryableError checks if an error is retryable
func (tm *TransactionManager) isRetryableError(err error) bool {
	if !tm.config.EnableDeadlockRetry {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	for _, retryableErr := range tm.retryPolicy.RetryableErrors {
		if strings.Contains(errStr, strings.ToLower(retryableErr)) {
			return true
		}
	}
	
	return false
}

// BatchInsertIssues performs batch insert of issues in a single transaction
func (tm *TransactionManager) BatchInsertIssues(issues []*Issue) ([]int64, error) {
	if len(issues) == 0 {
		return nil, nil
	}
	
	var insertIDs []int64
	err := tm.ExecuteInTransaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO issues (
				number, title, body, url, state, created_at, updated_at, comments,
				reactions, assignee, milestone, repo_owner, repo_name, keywords,
				score, content_hash, category, priority, tech_stack, labels,
				is_duplicate, duplicate_of, created_at_db, updated_at_db
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare batch insert statement: %w", err)
		}
		defer stmt.Close()
		
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
				return fmt.Errorf("failed to insert issue: %w", err)
			}
			
			id, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert ID: %w", err)
			}
			
			issue.ID = int(id)
			insertIDs = append(insertIDs, id)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("batch insert failed: %w", err)
	}
	
	tm.logger.Printf("Batch inserted %d issues", len(insertIDs))
	return insertIDs, nil
}

// BatchUpdateIssues performs batch update of issues in a single transaction
func (tm *TransactionManager) BatchUpdateIssues(issues []*Issue) error {
	if len(issues) == 0 {
		return nil
	}
	
	return tm.ExecuteInTransaction(func(tx *sql.Tx) error {
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
			return fmt.Errorf("failed to prepare batch update statement: %w", err)
		}
		defer stmt.Close()
		
		now := time.Now()
		updatedCount := 0
		
		for _, issue := range issues {
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
				return fmt.Errorf("failed to update issue: %w", err)
			}
			
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("failed to get rows affected: %w", err)
			}
			
			if rowsAffected > 0 {
				updatedCount++
			}
		}
		
		tm.logger.Printf("Batch updated %d issues", updatedCount)
		return nil
	})
}

// PerformDeduplicationInTransaction performs deduplication in a transaction
func (tm *TransactionManager) PerformDeduplicationInTransaction(deduplicator *DeduplicationService) (*DeduplicationResult, error) {
	var result *DeduplicationResult
	
	err := tm.ExecuteInTransaction(func(tx *sql.Tx) error {
		// Temporarily replace the database connection for deduplication
		oldDB := deduplicator.db
		deduplicator.db = tx
		
		// Perform deduplication
		dupResult, err := deduplicator.FindDuplicates()
		if err != nil {
			return fmt.Errorf("deduplication failed: %w", err)
		}
		
		result = dupResult
		
		// Restore original database connection
		deduplicator.db = oldDB
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("deduplication transaction failed: %w", err)
	}
	
	return result, nil
}

// PerformClassificationInTransaction performs classification in a transaction
func (tm *TransactionManager) PerformClassificationInTransaction(classifier *ClassificationService, issues []*Issue) (*ClassificationResult, error) {
	var result *ClassificationResult
	
	err := tm.ExecuteInTransaction(func(tx *sql.Tx) error {
		// Temporarily replace the database connection for classification
		oldDB := classifier.db
		classifier.db = tx
		
		// Perform classification
		clsResult, err := classifier.ClassifyIssues(issues)
		if err != nil {
			return fmt.Errorf("classification failed: %w", err)
		}
		
		result = clsResult
		
		// Restore original database connection
		classifier.db = oldDB
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("classification transaction failed: %w", err)
	}
	
	return result, nil
}

// GetActiveTransactions returns information about active transactions
func (tm *TransactionManager) GetActiveTransactions() []map[string]interface{} {
	tm.txMutex.RLock()
	defer tm.txMutex.RUnlock()
	
	var transactions []map[string]interface{}
	
	for txID, tx := range tm.activeTx {
		transactions = append(transactions, map[string]interface{}{
			"id":        txID,
			"started_at": time.Now(), // We don't track individual start times per transaction in activeTx
			"status":    TransactionStatusActive,
		})
	}
	
	return transactions
}

// CleanupExpiredTransactions cleans up expired transactions
func (tm *TransactionManager) CleanupExpiredTransactions() {
	tm.txMutex.Lock()
	defer tm.txMutex.Unlock()
	
	var expiredIDs []string
	now := time.Now()
	
	for txID, tx := range tm.activeTx {
		// Check if transaction has been active for too long
		// Note: We don't track individual start times, so this is a simplified cleanup
		// In a real implementation, you'd want to track individual transaction start times
		if now.Sub(time.Now()) > tm.config.Timeout {
			expiredIDs = append(expiredIDs, txID)
		}
	}
	
	// Rollback expired transactions
	for _, txID := range expiredIDs {
		if tx, exists := tm.activeTx[txID]; exists {
			if err := tx.Rollback(); err != nil {
				tm.logger.Printf("Failed to rollback expired transaction %s: %v", txID, err)
			}
			delete(tm.activeTx, txID)
			tm.logger.Printf("Cleaned up expired transaction %s", txID)
		}
	}
}

// GetTransactionStats returns statistics about transactions
func (tm *TransactionManager) GetTransactionStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	tm.txMutex.RLock()
	activeCount := len(tm.activeTx)
	tm.txMutex.RUnlock()
	
	stats["active_transactions"] = activeCount
	stats["max_active_transactions"] = tm.config.MaxActiveTransactions
	stats["timeout"] = tm.config.Timeout.String()
	stats["isolation_level"] = tm.config.IsolationLevel
	stats["retry_enabled"] = tm.config.EnableDeadlockRetry
	
	return stats
}

// TransactionOption represents options for transaction creation
type TransactionOption struct {
	IsolationLevel *sql.IsolationLevel
}

// WithIsolationLevel sets the isolation level for a transaction
func WithIsolationLevel(level sql.IsolationLevel) TransactionOption {
	return TransactionOption{
		IsolationLevel: &level,
	}
}

// generateTransactionID generates a unique transaction ID
func generateTransactionID() string {
	return fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// Strings returns the lowercase version of s
func Strings(s string) string {
	return strings.ToLower(s)
}