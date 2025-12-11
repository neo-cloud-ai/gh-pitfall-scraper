package scraper

import (
	"math"
	"strings"
	"time"
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