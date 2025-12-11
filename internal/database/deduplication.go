package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// DeduplicationService provides intelligent deduplication functionality
type DeduplicationService struct {
	db          *sql.DB
	logger      *log.Logger
	config      DeduplicationConfig
	similarity  *SimilarityEngine
}

// DeduplicationConfig represents configuration for deduplication
type DeduplicationConfig struct {
	SimilarityThreshold   float64 `json:"similarity_threshold"`   // Minimum similarity score to consider as duplicate (0.0 - 1.0)
	ContentHashWeight     float64 `json:"content_hash_weight"`     // Weight for content hash matching
	TitleWeight          float64 `json:"title_weight"`            // Weight for title similarity
	BodyWeight           float64 `json:"body_weight"`             // Weight for body similarity
	RepoWeight           float64 `json:"repo_weight"`             // Weight for repository matching
	TechStackWeight      float64 `json:"tech_stack_weight"`       // Weight for tech stack similarity
	BatchSize            int     `json:"batch_size"`              // Number of issues to process in one batch
	MaxProcessingTime    time.Duration `json:"max_processing_time"` // Maximum time to spend on deduplication
	EnableParallel       bool    `json:"enable_parallel"`         // Enable parallel processing
	MaxWorkers           int     `json:"max_workers"`             // Maximum number of parallel workers
}

// DefaultDeduplicationConfig returns default configuration
func DefaultDeduplicationConfig() DeduplicationConfig {
	return DeduplicationConfig{
		SimilarityThreshold: 0.75,
		ContentHashWeight:   0.4,
		TitleWeight:         0.3,
		BodyWeight:          0.2,
		RepoWeight:          0.05,
		TechStackWeight:     0.05,
		BatchSize:           100,
		MaxProcessingTime:   10 * time.Minute,
		EnableParallel:      true,
		MaxWorkers:          4,
	}
}

// SimilarityEngine calculates similarity between issues
type SimilarityEngine struct{}

// NewSimilarityEngine creates a new similarity engine
func NewSimilarityEngine() *SimilarityEngine {
	return &SimilarityEngine{}
}

// CalculateSimilarity calculates similarity between two issues
func (se *SimilarityEngine) CalculateSimilarity(issue1, issue2 *Issue) float64 {
	similarity := 0.0
	
	// Content hash similarity (exact match gets high score)
	contentHashSim := calculateContentHashSimilarity(issue1.ContentHash, issue2.ContentHash)
	similarity += contentHashSim * 0.4
	
	// Title similarity
	titleSim := calculateStringSimilarity(issue1.Title, issue2.Title)
	similarity += titleSim * 0.3
	
	// Body similarity
	bodySim := calculateStringSimilarity(issue1.Body, issue2.Body)
	similarity += bodySim * 0.2
	
	// Repository similarity
	repoSim := calculateRepositorySimilarity(issue1.RepoOwner, issue1.RepoName, issue2.RepoOwner, issue2.RepoName)
	similarity += repoSim * 0.05
	
	// Tech stack similarity
	techStackSim := calculateStringArraySimilarity(issue1.TechStack, issue2.TechStack)
	similarity += techStackSim * 0.05
	
	return math.Min(1.0, math.Max(0.0, similarity))
}

// calculateContentHashSimilarity calculates similarity based on content hash
func calculateContentHashSimilarity(hash1, hash2 string) float64 {
	if hash1 == "" || hash2 == "" {
		return 0.0
	}
	if hash1 == hash2 {
		return 1.0
	}
	
	// Calculate Hamming distance between hashes
	hammingDistance := calculateHammingDistance(hash1, hash2)
	maxDistance := math.Max(float64(len(hash1)), float64(len(hash2)))
	
	return 1.0 - (hammingDistance / maxDistance)
}

// calculateHammingDistance calculates the Hamming distance between two strings
func calculateHammingDistance(s1, s2 string) int {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}
	
	distance := 0
	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			distance++
		}
	}
	
	// Add penalty for different lengths
	distance += abs(len(s1) - len(s2))
	
	return distance
}

// calculateStringSimilarity calculates similarity between two strings using multiple algorithms
func calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	// Normalize strings
	norm1 := normalizeString(s1)
	norm2 := normalizeString(s2)
	
	if norm1 == norm2 {
		return 1.0
	}
	
	if len(norm1) == 0 || len(norm2) == 0 {
		return 0.0
	}
	
	// Use multiple similarity algorithms
	levenshteinSim := calculateLevenshteinSimilarity(norm1, norm2)
	jaccardSim := calculateJaccardSimilarity(norm1, norm2)
	
	// Combine similarities with weights
	return (levenshteinSim*0.6 + jaccardSim*0.4)
}

// calculateLevenshteinSimilarity calculates similarity using Levenshtein distance
func calculateLevenshteinSimilarity(s1, s2 string) float64 {
	dp := make([][]int, len(s1)+1)
	for i := range dp {
		dp[i] = make([]int, len(s2)+1)
	}
	
	for i := 0; i <= len(s1); i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		dp[0][j] = j
	}
	
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = min(
					dp[i-1][j]+1,    // deletion
					dp[i][j-1]+1,    // insertion
					dp[i-1][j-1]+1,  // substitution
				)
			}
		}
	}
	
	maxLen := float64(max(len(s1), len(s2)))
	return 1.0 - (float64(dp[len(s1)][len(s2)]) / maxLen)
}

// calculateJaccardSimilarity calculates similarity using Jaccard index
func calculateJaccardSimilarity(s1, s2 string) float64 {
	words1 := strings.Fields(strings.ToLower(s1))
	words2 := strings.Fields(strings.ToLower(s2))
	
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

// calculateRepositorySimilarity calculates similarity based on repository
func calculateRepositorySimilarity(owner1, name1, owner2, name2 string) float64 {
	if owner1 == owner2 && name1 == name2 {
		return 1.0
	}
	if owner1 == owner2 || name1 == name2 {
		return 0.5
	}
	return 0.0
}

// calculateStringArraySimilarity calculates similarity between string arrays
func calculateStringArraySimilarity(arr1, arr2 StringArray) float64 {
	if len(arr1) == 0 && len(arr2) == 0 {
		return 1.0
	}
	
	set1 := make(map[string]bool)
	for _, item := range arr1 {
		set1[item] = true
	}
	
	intersection := 0
	union := len(set1)
	
	for _, item := range arr2 {
		if set1[item] {
			intersection++
		} else {
			union++
		}
	}
	
	return float64(intersection) / float64(union)
}

// normalizeString normalizes a string for comparison
func normalizeString(s string) string {
	// Convert to lowercase
	normalized := strings.ToLower(s)
	
	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")
	
	// Remove punctuation (optional, but can help with matching)
	normalized = removePunctuation(normalized)
	
	return normalized
}

// removePunctuation removes punctuation from a string
func removePunctuation(s string) string {
	var result strings.Builder
	for _, r := range s {
		if isAlphaNum(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return result.String()
}

// isAlphaNum checks if a rune is alphanumeric
func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// abs returns absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// max returns maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewDeduplicationService creates a new deduplication service
func NewDeduplicationService(db *sql.DB, config DeduplicationConfig) *DeduplicationService {
	if config.BatchSize == 0 {
		config.BatchSize = DefaultDeduplicationConfig().BatchSize
	}
	if config.MaxWorkers == 0 {
		config.MaxWorkers = DefaultDeduplicationConfig().MaxWorkers
	}
	
	return &DeduplicationService{
		db:         db,
		logger:     log.New(log.Writer(), "[Deduplication] ", log.LstdFlags),
		config:     config,
		similarity: NewSimilarityEngine(),
	}
}

// FindDuplicates finds duplicate issues in the database
func (ds *DeduplicationService) FindDuplicates() (*DeduplicationResult, error) {
	ds.logger.Println("Starting duplicate detection...")
	
	startTime := time.Now()
	
	// Get all issues that haven't been processed yet
	issues, err := ds.getUnprocessedIssues()
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed issues: %w", err)
	}
	
	if len(issues) == 0 {
		ds.logger.Println("No issues to process for deduplication")
		return &DeduplicationResult{
			TotalProcessed:  0,
			DuplicatesFound: 0,
			DuplicatesRemoved: 0,
			UniqueIssues:   0,
			DuplicateGroups: []DuplicateGroup{},
		}, nil
	}
	
	ds.logger.Printf("Processing %d issues for duplicate detection", len(issues))
	
	// Process issues in batches
	var allDuplicateGroups []DuplicateGroup
	var mutex sync.Mutex
	var wg sync.WaitGroup
	
	if ds.config.EnableParallel {
		// Process in parallel
		batchSize := ds.config.BatchSize
		numBatches := (len(issues) + batchSize - 1) / batchSize
		
		for i := 0; i < numBatches; i++ {
			start := i * batchSize
			end := min(start+batchSize, len(issues))
			batch := issues[start:end]
			
			wg.Add(1)
			go func(b []*Issue) {
				defer wg.Done()
				groups := ds.processBatch(b)
				mutex.Lock()
				allDuplicateGroups = append(allDuplicateGroups, groups...)
				mutex.Unlock()
			}(batch)
		}
		
		wg.Wait()
	} else {
		// Process sequentially
		allDuplicateGroups = ds.processBatch(issues)
	}
	
	// Update database with duplicate information
	duplicatesFound, duplicatesRemoved, err := ds.updateDuplicateFlags(allDuplicateGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to update duplicate flags: %w", err)
	}
	
	processingTime := time.Since(startTime)
	ds.logger.Printf("Duplicate detection completed in %v", processingTime)
	
	result := &DeduplicationResult{
		TotalProcessed:   len(issues),
		DuplicatesFound:  duplicatesFound,
		DuplicatesRemoved: duplicatesRemoved,
		UniqueIssues:     len(issues) - duplicatesFound,
		DuplicateGroups:  allDuplicateGroups,
	}
	
	ds.logger.Printf("Found %d duplicate groups, %d total duplicates", len(allDuplicateGroups), duplicatesFound)
	return result, nil
}

// processBatch processes a batch of issues for duplicate detection
func (ds *DeduplicationService) processBatch(issues []*Issue) []DuplicateGroup {
	var duplicateGroups []DuplicateGroup
	
	// Sort issues by score (highest first) to prioritize keeping high-score issues
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Score > issues[j].Score
	})
	
	processed := make(map[int]bool)
	
	for i, issue := range issues {
		if processed[i] {
			continue
		}
		
		var duplicates []*Issue
		highestSimilarity := 0.0
		
		// Compare with all other unprocessed issues
		for j := i + 1; j < len(issues); j++ {
			if processed[j] {
				continue
			}
			
			otherIssue := issues[j]
			similarity := ds.similarity.CalculateSimilarity(issue, otherIssue)
			
			if similarity >= ds.config.SimilarityThreshold {
				duplicates = append(duplicates, otherIssue)
				processed[j] = true
				
				if similarity > highestSimilarity {
					highestSimilarity = similarity
				}
			}
		}
		
		// If duplicates found, create a duplicate group
		if len(duplicates) > 0 {
			group := DuplicateGroup{
				MasterIssue: issue,
				Duplicates:  duplicates,
				Similarity:  highestSimilarity,
				Reason:      ds.getDuplicateReason(issue, duplicates),
			}
			duplicateGroups = append(duplicateGroups, group)
			processed[i] = true
		}
	}
	
	return duplicateGroups
}

// getDuplicateReason generates a human-readable reason for why issues are considered duplicates
func (ds *DeduplicationService) getDuplicateReason(master *Issue, duplicates []*Issue) string {
	var reasons []string
	
	// Check content hash similarity
	if len(duplicates) > 0 {
		other := duplicates[0]
		if master.ContentHash == other.ContentHash {
			reasons = append(reasons, "identical content")
		}
	}
	
	// Check title similarity
	titleSimilarities := make([]float64, len(duplicates))
	for i, dup := range duplicates {
		titleSimilarities[i] = calculateStringSimilarity(master.Title, dup.Title)
	}
	avgTitleSim := average(titleSimilarities)
	
	if avgTitleSim > 0.8 {
		reasons = append(reasons, "very similar titles")
	} else if avgTitleSim > 0.6 {
		reasons = append(reasons, "similar titles")
	}
	
	// Check repository
	if len(duplicates) > 0 {
		sameRepo := true
		for _, dup := range duplicates {
			if master.RepoOwner != dup.RepoOwner || master.RepoName != dup.RepoName {
				sameRepo = false
				break
			}
		}
		if sameRepo {
			reasons = append(reasons, "same repository")
		}
	}
	
	if len(reasons) == 0 {
		reasons = append(reasons, "high similarity score")
	}
	
	return strings.Join(reasons, ", ")
}

// updateDuplicateFlags updates the database with duplicate information
func (ds *DeduplicationService) updateDuplicateFlags(groups []DuplicateGroup) (int, int, error) {
	tx, err := ds.db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	duplicatesFound := 0
	duplicatesRemoved := 0
	
	for _, group := range groups {
		master := group.MasterIssue
		
		// Mark duplicates as duplicates
		for _, duplicate := range group.Duplicates {
			_, err := tx.Exec(
				"UPDATE issues SET is_duplicate = 1, duplicate_of = ?, updated_at_db = ? WHERE id = ?",
				master.ID, time.Now(), duplicate.ID,
			)
			if err != nil {
				tx.Rollback()
				return 0, 0, fmt.Errorf("failed to update duplicate flag: %w", err)
			}
			duplicatesFound++
		}
		
		// Optionally remove duplicates
		if ds.config.MaxProcessingTime > 0 && duplicatesFound > 1000 {
			// For large datasets, remove duplicates after marking them
			for _, duplicate := range group.Duplicates {
				_, err := tx.Exec("DELETE FROM issues WHERE id = ?", duplicate.ID)
				if err != nil {
					tx.Rollback()
					return 0, 0, fmt.Errorf("failed to delete duplicate: %w", err)
				}
				duplicatesRemoved++
			}
		}
	}
	
	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return duplicatesFound, duplicatesRemoved, nil
}

// getUnprocessedIssues retrieves issues that haven't been processed for deduplication
func (ds *DeduplicationService) getUnprocessedIssues() ([]*Issue, error) {
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE (is_duplicate IS NULL OR is_duplicate = 0)
		ORDER BY score DESC, created_at_db DESC
		LIMIT ?
	`
	
	rows, err := ds.db.Query(query, ds.config.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed issues: %w", err)
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
		
		// Generate content hash if missing
		if issue.ContentHash == "" {
			issue.ContentHash = generateContentHash(&issue)
		}
		
		issues = append(issues, &issue)
	}
	
	return issues, nil
}

// generateContentHash generates a hash for issue content
func generateContentHash(issue *Issue) string {
	// Combine all relevant content
	content := fmt.Sprintf("%s\n%s\n%s\n%s",
		issue.Title,
		issue.Body,
		strings.Join(issue.Keywords, ","),
		strings.Join(issue.TechStack, ","),
	)
	
	// Create SHA256 hash
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// FindSimilarIssues finds issues similar to a given issue
func (ds *DeduplicationService) FindSimilarIssues(issue *Issue, limit int) ([]*Issue, float64, error) {
	// Get candidate issues (same repository, similar score range)
	candidates, err := ds.getSimilarCandidates(issue)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get similar candidates: %w", err)
	}
	
	// Calculate similarities and sort
	type similarityResult struct {
		issue      *Issue
		similarity float64
	}
	
	var results []similarityResult
	for _, candidate := range candidates {
		if candidate.ID == issue.ID {
			continue // Skip the same issue
		}
		
		similarity := ds.similarity.CalculateSimilarity(issue, candidate)
		if similarity >= ds.config.SimilarityThreshold {
			results = append(results, similarityResult{
				issue:      candidate,
				similarity: similarity,
			})
		}
	}
	
	// Sort by similarity (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})
	
	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}
	
	// Extract issues
	similarIssues := make([]*Issue, len(results))
	avgSimilarity := 0.0
	
	for i, result := range results {
		similarIssues[i] = result.issue
		avgSimilarity += result.similarity
	}
	
	if len(results) > 0 {
		avgSimilarity /= float64(len(results))
	}
	
	return similarIssues, avgSimilarity, nil
}

// getSimilarCandidates retrieves candidate issues for similarity comparison
func (ds *DeduplicationService) getSimilarCandidates(issue *Issue) ([]*Issue, error) {
	// Get issues from the same repository with similar scores
	query := `
		SELECT id, number, title, body, url, state, created_at, updated_at, comments,
			reactions, assignee, milestone, repo_owner, repo_name, keywords,
			score, content_hash, category, priority, tech_stack, labels,
			is_duplicate, duplicate_of, created_at_db, updated_at_db
		FROM issues
		WHERE repo_owner = ? AND repo_name = ?
		AND score >= ? AND score <= ?
		ORDER BY score DESC, created_at_db DESC
		LIMIT 1000
	`
	
	scoreRange := issue.Score * 0.3 // 30% score range
	minScore := issue.Score - scoreRange
	maxScore := issue.Score + scoreRange
	
	rows, err := ds.db.Query(query, issue.RepoOwner, issue.RepoName, minScore, maxScore)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar candidates: %w", err)
	}
	defer rows.Close()
	
	var candidates []*Issue
	for rows.Next() {
		var candidate Issue
		var keywords, techStack, labels StringArray
		
		err := rows.Scan(
			&candidate.ID, &candidate.Number, &candidate.Title, &candidate.Body, &candidate.URL,
			&candidate.State, &candidate.CreatedAt, &candidate.UpdatedAt, &candidate.Comments,
			&candidate.Reactions, &candidate.Assignee, &candidate.Milestone, &candidate.RepoOwner,
			&candidate.RepoName, &keywords, &candidate.Score, &candidate.ContentHash,
			&candidate.Category, &candidate.Priority, &techStack, &labels,
			&candidate.IsDuplicate, &candidate.DuplicateOf, &candidate.CreatedAtDB,
			&candidate.UpdatedAtDB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candidate: %w", err)
		}
		
		candidate.Keywords = keywords
		candidate.TechStack = techStack
		candidate.Labels = labels
		
		candidates = append(candidates, &candidate)
	}
	
	return candidates, nil
}

// RemoveDuplicates removes duplicate issues from the database
func (ds *DeduplicationService) RemoveDuplicates() (int, error) {
	ds.logger.Println("Starting duplicate removal...")
	
	// Get all marked duplicates
	query := `
		SELECT id FROM issues 
		WHERE is_duplicate = 1 AND duplicate_of IS NOT NULL
		ORDER BY score ASC, created_at_db ASC
	`
	
	rows, err := ds.db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to query duplicates: %w", err)
	}
	defer rows.Close()
	
	var duplicateIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("failed to scan duplicate ID: %w", err)
		}
		duplicateIDs = append(duplicateIDs, id)
	}
	
	if len(duplicateIDs) == 0 {
		ds.logger.Println("No duplicates found to remove")
		return 0, nil
	}
	
	// Delete duplicates in batches
	batchSize := 100
	deletedCount := 0
	
	for i := 0; i < len(duplicateIDs); i += batchSize {
		end := min(i+batchSize, len(duplicateIDs))
		batch := duplicateIDs[i:end]
		
		placeholders := strings.Repeat("?,", len(batch)-1) + "?"
		query := fmt.Sprintf("DELETE FROM issues WHERE id IN (%s)", placeholders)
		
		result, err := ds.db.Exec(query, interfaceSlice(batch)...)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to delete duplicate batch: %w", err)
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return deletedCount, fmt.Errorf("failed to get rows affected: %w", err)
		}
		
		deletedCount += int(rowsAffected)
	}
	
	ds.logger.Printf("Removed %d duplicate issues", deletedCount)
	return deletedCount, nil
}

// GetDuplicateStats returns statistics about duplicates in the database
func (ds *DeduplicationService) GetDuplicateStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get total duplicate count
	var totalDuplicates int
	err := ds.db.QueryRow("SELECT COUNT(*) FROM issues WHERE is_duplicate = 1").Scan(&totalDuplicates)
	if err != nil {
		return nil, fmt.Errorf("failed to get duplicate count: %w", err)
	}
	stats["total_duplicates"] = totalDuplicates
	
	// Get unique issues count
	var uniqueIssues int
	err = ds.db.QueryRow("SELECT COUNT(*) FROM issues WHERE is_duplicate = 0 OR is_duplicate IS NULL").Scan(&uniqueIssues)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique issues count: %w", err)
	}
	stats["unique_issues"] = uniqueIssues
	
	// Get duplicates by repository
	rows, err := ds.db.Query(`
		SELECT repo_owner, repo_name, COUNT(*) as duplicate_count
		FROM issues 
		WHERE is_duplicate = 1
		GROUP BY repo_owner, repo_name
		ORDER BY duplicate_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get duplicates by repository: %w", err)
	}
	defer rows.Close()
	
	var duplicatesByRepo []map[string]interface{}
	for rows.Next() {
		var repoOwner, repoName string
		var count int
		if err := rows.Scan(&repoOwner, &repoName, &count); err != nil {
			continue
		}
		duplicatesByRepo = append(duplicatesByRepo, map[string]interface{}{
			"repository": fmt.Sprintf("%s/%s", repoOwner, repoName),
			"count":      count,
		})
	}
	stats["duplicates_by_repository"] = duplicatesByRepo
	
	// Calculate duplicate rate
	if uniqueIssues > 0 {
		duplicateRate := float64(totalDuplicates) / float64(uniqueIssues+totalDuplicates)
		stats["duplicate_rate"] = duplicateRate
	}
	
	return stats, nil
}

// average calculates the average of a slice of floats
func average(slice []float64) float64 {
	if len(slice) == 0 {
		return 0.0
	}
	
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum / float64(len(slice))
}