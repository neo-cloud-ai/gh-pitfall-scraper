package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
)

// PitfallScorer implements intelligent scoring for pitfall issues
type PitfallScorer struct {
	keywordWeights map[string]float64
	labelWeights   map[string]float64
}

// NewPitfallScorer creates a new pitfall scorer with default weights
func NewPitfallScorer() *PitfallScorer {
	return &PitfallScorer{
		keywordWeights: map[string]float64{
			// Critical performance issues
			"performance regression": 10.0,
			"latency":                8.0,
			"throughput":             8.0,
			"memory leak":            9.0,
			"OOM":                    9.5,
			
			// GPU-specific issues
			"CUDA":                   7.0,
			"kernel":                 6.5,
			"NCCL":                   8.5,
			"memory fragmentation":   7.5,
			
			// System-level issues
			"hang":                   8.0,
			"deadlock":               9.0,
			"distributed training":   8.5,
			
			// Inference-specific
			"kv cache":               7.0,
			"prefill":                6.0,
			"decode":                 6.0,
			
			// General technical issues
			"crash":                  8.0,
			"timeout":                6.5,
			"regression":             7.5,
		},
		labelWeights: map[string]float64{
			"bug":           8.0,
			"performance":   9.0,
			"critical":      10.0,
			"enhancement":   4.0,
			"documentation": 2.0,
			"question":      3.0,
			"help wanted":   5.0,
			"good first issue": 3.0,
		},
	}
}

// Score calculates the pitfall value score for an issue
func (s *PitfallScorer) Score(issue *Issue, keywords []string) float64 {
	var totalScore float64
	
	// 1. Keyword matching score (40% weight)
	keywordScore := s.calculateKeywordScore(issue, keywords)
	totalScore += keywordScore * 0.4
	
	// 2. Label-based score (25% weight)
	labelScore := s.calculateLabelScore(issue.Labels)
	totalScore += labelScore * 0.25
	
	// 3. Engagement score (20% weight)
	engagementScore := s.calculateEngagementScore(issue)
	totalScore += engagementScore * 0.2
	
	// 4. Recency score (15% weight)
	recencyScore := s.calculateRecencyScore(issue)
	totalScore += recencyScore * 0.15
	
	return totalScore
}

// calculateKeywordScore calculates score based on keyword matches
func (s *PitfallScorer) calculateKeywordScore(issue *Issue, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0.0
	}
	
	searchText := strings.ToLower(issue.Title + " " + issue.Body)
	var score float64
	
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		
		// Check exact matches in title (higher weight)
		if strings.Contains(strings.ToLower(issue.Title), keywordLower) {
			weight, exists := s.keywordWeights[keyword]
			if !exists {
				weight = 5.0 // default weight
			}
			score += weight * 2.0 // Title matches are worth double
		}
		
		// Check matches in body
		if strings.Contains(searchText, keywordLower) {
			weight, exists := s.keywordWeights[keyword]
			if !exists {
				weight = 5.0 // default weight
			}
			score += weight
		}
	}
	
	return math.Min(score, 50.0) // Cap at 50 points
}

// calculateLabelScore calculates score based on issue labels
func (s *PitfallScorer) calculateLabelScore(labels []Label) float64 {
	var score float64
	
	for _, label := range labels {
		labelLower := strings.ToLower(label.Name)
		if weight, exists := s.labelWeights[labelLower]; exists {
			score += weight
		}
	}
	
	return math.Min(score, 25.0) // Cap at 25 points
}

// calculateEngagementScore calculates score based on community engagement
func (s *PitfallScorer) calculateEngagementScore(issue *Issue) float64 {
	var score float64
	
	// Comments score (more discussion = higher priority)
	if issue.Comments > 0 {
		score += math.Log(float64(issue.Comments)+1) * 2.0
	}
	
	// Reactions score (community feedback)
	if issue.Reactions.TotalCount > 0 {
		score += math.Sqrt(float64(issue.Reactions.TotalCount)) * 3.0
	}
	
	// Assigned developer indicator
	if issue.Assignee.Login != "" {
		score += 2.0
	}
	
	// Milestone indicator
	if issue.Milestone.Title != "" {
		score += 3.0
	}
	
	return math.Min(score, 20.0) // Cap at 20 points
}

// calculateRecencyScore calculates score based on issue freshness
func (s *PitfallScorer) calculateRecencyScore(issue *Issue) float64 {
	daysSinceUpdate := time.Since(issue.UpdatedAt).Hours() / 24.0
	
	// Recent issues get higher scores
	if daysSinceUpdate < 7 {
		return 15.0
	} else if daysSinceUpdate < 30 {
		return 10.0
	} else if daysSinceUpdate < 90 {
		return 5.0
	} else {
		return 2.0
	}
}

// GetMatchingKeywords returns keywords that match an issue
func (s *PitfallScorer) GetMatchingKeywords(issue *Issue, keywords []string) []string {
	var matchingKeywords []string
	searchText := strings.ToLower(issue.Title + " " + issue.Body)
	
	for _, keyword := range keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			matchingKeywords = append(matchingKeywords, keyword)
		}
	}
	
	return matchingKeywords
}

// IsHighValueIssue determines if an issue meets high-value criteria
func (s *PitfallScorer) IsHighValueIssue(score float64) bool {
	return score >= 15.0 // Threshold for high-value issues
}

// DatabaseScorer extends PitfallScorer with database storage capabilities
type DatabaseScorer struct {
	*PitfallScorer
	db         *sql.DB
	crudOps    database.CRUDOperations
	logger     *log.Logger
	storeMutex sync.Mutex
}

// NewDatabaseScorer creates a new database-enabled scorer
func NewDatabaseScorer(db *sql.DB) *DatabaseScorer {
	return &DatabaseScorer{
		PitfallScorer: NewPitfallScorer(),
		db:           db,
		crudOps:      database.NewCRUDOperations(db),
		logger:       log.New(log.Writer(), "[DB-Scorer] ", log.LstdFlags),
	}
}

// ScoreAndStore scores an issue and stores it to database
func (ds *DatabaseScorer) ScoreAndStore(issue *Issue, keywords []string) (float64, error) {
	// Calculate score
	score := ds.Score(issue, keywords)
	
	// Convert to database model
	dbIssue := ds.convertToDatabaseIssue(issue, score, keywords)
	
	// Store to database
	_, err := ds.crudOps.CreateIssue(&dbIssue)
	if err != nil {
		return score, fmt.Errorf("failed to store scored issue: %w", err)
	}
	
	ds.logger.Printf("Stored issue #%d with score %.2f", issue.Number, score)
	return score, nil
}

// ScoreAndStoreBatch scores and stores multiple issues in batch
func (ds *DatabaseScorer) ScoreAndStoreBatch(issues []*Issue, keywords []string) ([]float64, error) {
	if len(issues) == 0 {
		return nil, nil
	}
	
	ds.storeMutex.Lock()
	defer ds.storeMutex.Unlock()
	
	// Convert all issues to database models
	dbIssues := make([]*database.Issue, len(issues))
	scores := make([]float64, len(issues))
	
	for i, issue := range issues {
		scores[i] = ds.Score(issue, keywords)
		dbIssues[i] = ds.convertToDatabaseIssue(issue, scores[i], keywords)
	}
	
	// Batch insert
	ids, err := ds.crudOps.CreateIssues(dbIssues)
	if err != nil {
		return scores, fmt.Errorf("failed to batch store scored issues: %w", err)
	}
	
	ds.logger.Printf("Stored %d scored issues with IDs: %v", len(ids), ids)
	return scores, nil
}

// UpdateIssueScore updates an existing issue's score in database
func (ds *DatabaseScorer) UpdateIssueScore(issueID int64, issue *Issue, keywords []string) error {
	// Calculate new score
	newScore := ds.Score(issue, keywords)
	
	// Update in database
	_, err := ds.crudOps.UpdateIssue(&database.Issue{
		ID:    int(issueID),
		Score: newScore,
	})
	
	if err != nil {
		return fmt.Errorf("failed to update issue score: %w", err)
	}
	
	ds.logger.Printf("Updated score for issue #%d to %.2f", issue.Number, newScore)
	return nil
}

// GetTopScoredIssues retrieves top scored issues from database
func (ds *DatabaseScorer) GetTopScoredIssues(limit int) ([]*database.Issue, error) {
	return ds.crudOps.GetIssuesByScore(20.0, 100.0, limit, 0) // High-value issues
}

// GetScoreStatistics returns scoring statistics from database
func (ds *DatabaseScorer) GetScoreStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get high-score issues count
	highScoreIssues, err := ds.crudOps.GetIssuesByScore(20.0, 100.0, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get high-score issues: %w", err)
	}
	
	// Get medium-score issues count
	mediumScoreIssues, err := ds.crudOps.GetIssuesByScore(10.0, 20.0, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get medium-score issues: %w", err)
	}
	
	// Get low-score issues count
	lowScoreIssues, err := ds.crudOps.GetIssuesByScore(0.0, 10.0, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get low-score issues: %w", err)
	}
	
	stats["high_score_count"] = len(highScoreIssues)
	stats["medium_score_count"] = len(mediumScoreIssues)
	stats["low_score_count"] = len(lowScoreIssues)
	stats["total_issues"] = len(highScoreIssues) + len(mediumScoreIssues) + len(lowScoreIssues)
	
	// Calculate average scores
	if len(highScoreIssues) > 0 {
		stats["avg_high_score"] = ds.calculateAverageScore(highScoreIssues)
	}
	if len(mediumScoreIssues) > 0 {
		stats["avg_medium_score"] = ds.calculateAverageScore(mediumScoreIssues)
	}
	if len(lowScoreIssues) > 0 {
		stats["avg_low_score"] = ds.calculateAverageScore(lowScoreIssues)
	}
	
	return stats, nil
}

// RescoreIssues resores all issues in database using current scoring algorithm
func (ds *DatabaseScorer) RescoreIssues(limit int) (int, error) {
	// Get issues to rescore
	issues, err := ds.crudOps.GetAllIssues(limit, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get issues for rescoring: %w", err)
	}
	
	rescoredCount := 0
	
	// Convert back to internal Issue format for rescoring
	for _, dbIssue := range issues {
		// Convert to internal Issue format
		issue := &Issue{
			ID:          dbIssue.ID,
			Number:      dbIssue.Number,
			Title:       dbIssue.Title,
			Body:        dbIssue.Body,
			State:       dbIssue.State,
			CreatedAt:   dbIssue.CreatedAt.Time,
			UpdatedAt:   dbIssue.UpdatedAt.Time,
			Comments:    dbIssue.CommentsCount,
			Labels:      ds.convertFromDatabaseLabels(dbIssue.Labels),
		}
		
		// Recalculate score
		newScore := ds.Score(issue, []string(dbIssue.Keywords))
		
		// Update in database
		dbIssue.Score = newScore
		err = ds.crudOps.UpdateIssue(dbIssue)
		if err != nil {
			ds.logger.Printf("Failed to update score for issue #%d: %v", issue.Number, err)
			continue
		}
		
		rescoredCount++
	}
	
	ds.logger.Printf("Rescored %d issues", rescoredCount)
	return rescoredCount, nil
}

// convertToDatabaseIssue converts Issue to database.Issue with score
func (ds *DatabaseScorer) convertToDatabaseIssue(issue *Issue, score float64, keywords []string) database.Issue {
	return database.Issue{
		IssueID:       int64(issue.ID),
		Number:        issue.Number,
		Title:         issue.Title,
		Body:          issue.Body,
		URL:           issue.URL,
		State:         issue.State,
		AuthorLogin:   issue.Assignee.Login,
		Labels:        database.JSONSlice{},
		Assignees:     database.JSONSlice{issue.Assignee.Login},
		Milestone:     issue.Milestone.Title,
		Reactions: database.ReactionCount{
			Total: issue.Reactions.TotalCount,
		},
		CreatedAt:      issue.CreatedAt,
		UpdatedAt:      issue.UpdatedAt,
		FirstSeenAt:    issue.CreatedAt,
		LastSeenAt:     issue.UpdatedAt,
		CommentsCount:  issue.Comments,
		Score:          score,
		URL:            issue.URL,
		HTMLURL:        issue.URL,
		Keywords:       database.JSONSlice(keywords),
	}
}

// convertFromDatabaseLabels converts database labels to internal format
func (ds *DatabaseScorer) convertFromDatabaseLabels(labels database.JSONSlice) []Label {
	// This is a simplified conversion - in practice, you'd need proper mapping
	return []Label{}
}

// calculateAverageScore calculates average score for a slice of issues
func (ds *DatabaseScorer) calculateAverageScore(issues []*database.Issue) float64 {
	if len(issues) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, issue := range issues {
		total += issue.Score
	}
	return total / float64(len(issues))
}

// BatchScoringResult represents the result of batch scoring operation
type BatchScoringResult struct {
	TotalIssues    int           `json:"total_issues"`
	Processed      int           `json:"processed"`
	Failed         int           `json:"failed"`
	AverageScore   float64       `json:"average_score"`
	ProcessingTime time.Duration `json:"processing_time"`
	Errors         []string      `json:"errors"`
}