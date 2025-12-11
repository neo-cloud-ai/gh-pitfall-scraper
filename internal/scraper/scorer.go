package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/model"
)

// Scorer implements intelligent scoring for issues
type Scorer struct {
	// AI/ML specific keywords
	keywords []string
	
	// High-value issue patterns
	patterns []*regexp.Regexp
	
	// Priority labels that indicate high value
	priorityLabels []string
}

// NewScorer creates a new issue scorer
func NewScorer() *Scorer {
	return &Scorer{
		keywords: []string{
			// Performance issues
			"performance regression", "performance", "slow", "speed", "optimization",
			"throughput", "latency", "bottleneck", "memory leak",
			
			// GPU and CUDA issues
			"gpu", "cuda", "oom", "out of memory", "memory fragmentation",
			"cuda kernel", "kernel crash", "nvidia",
			
			// Distributed training
			"nccl", "deadlock", "hanging", "distributed", "multi-gpu",
			"multi-node", "all-reduce", "gradient sync",
			
			// Model serving
			"kv cache", "prefill", "decode", "inference", "serving",
			"flashattention", "flashdecoding", "flash attention",
			
			// Common bug patterns
			"crash", "error", "exception", "bug", "failure", "timeout",
			"hang", "freeze", "stuck", "loop", "infinite",
		},
		
		patterns: []*regexp.Regexp{
			// Performance regression patterns
			regexp.MustCompile(`(?i)(performance\s+(regression|degradation|drop))`),
			regexp.MustCompile(`(?i)(slow(er)?\s+(than|vs|compared))`),
			regexp.MustCompile(`(?i)(speed\s+(drop|decrease|regression))`),
			
			// GPU/CUDA patterns
			regexp.MustCompile(`(?i)(gpu\s+(out\s+of\s+memory|oom))`),
			regexp.MustCompile(`(?i)(cuda\s+(error|crash|failed|timeout))`),
			regexp.MustCompile(`(?i)(memory\s+(leak|fragmentation|overflow))`),
			
			// Distributed training patterns
			regexp.MustCompile(`(?i)(nccl\s+(error|timeout|deadlock))`),
			regexp.MustCompile(`(?i)(distributed\s+(training|sync|hang))`),
			
			// Model serving patterns
			regexp.MustCompile(`(?i)(kv\s+cache\s+(error|issue|problem))`),
			regexp.MustCompile(`(?i)(flash\s*(attention|decoding)\s+(bug|error|issue))`),
		},
		
		priorityLabels: []string{
			"bug", "critical", "performance", "enhancement", "feature",
			"documentation", "good first issue", "help wanted",
			"high priority", "urgent", "regression",
		},
	}
}

// ScoreIssue calculates a score for an issue based on multiple factors
func (s *Scorer) ScoreIssue(issue *model.Issue) (float64, []string) {
	var score float64
	var reasons []string
	
	// 1. Keyword matching (30 points max)
	keywordScore := s.scoreKeywords(issue)
	score += keywordScore
	if keywordScore > 0 {
		reasons = append(reasons, fmt.Sprintf("关键词匹配: %.1f分", keywordScore))
	}
	
	// 2. Pattern matching (25 points max)
	patternScore := s.scorePatterns(issue)
	score += patternScore
	if patternScore > 0 {
		reasons = append(reasons, fmt.Sprintf("模式匹配: %.1f分", patternScore))
	}
	
	// 3. Label scoring (20 points max)
	labelScore := s.scoreLabels(issue)
	score += labelScore
	if labelScore > 0 {
		reasons = append(reasons, fmt.Sprintf("标签匹配: %.1f分", labelScore))
	}
	
	// 4. Status scoring (10 points max)
	statusScore := s.scoreStatus(issue)
	score += statusScore
	if statusScore > 0 {
		reasons = append(reasons, fmt.Sprintf("状态评分: %.1f分", statusScore))
	}
	
	// 5. Activity scoring (15 points max)
	activityScore := s.scoreActivity(issue)
	score += activityScore
	if activityScore > 0 {
		reasons = append(reasons, fmt.Sprintf("活跃度评分: %.1f分", activityScore))
	}
	
	return score, reasons
}

// scoreKeywords scores based on keyword matches
func (s *Scorer) scoreKeywords(issue *model.Issue) float64 {
	var score float64
	text := strings.ToLower(issue.Title + " " + issue.Body)
	
	// High-value keywords
	highValueKeywords := []string{
		"performance regression", "gpu oom", "cuda kernel crash",
		"nccl deadlock", "memory leak", "distributed hanging",
		"kv cache error", "flashattention bug",
	}
	
	for _, keyword := range highValueKeywords {
		if strings.Contains(text, keyword) {
			score += 3.0
		}
	}
	
	// Medium-value keywords
	for _, keyword := range s.keywords {
		if strings.Contains(text, keyword) {
			score += 1.0
		}
	}
	
	// Cap at 30 points
	if score > 30 {
		score = 30
	}
	
	return score
}

// scorePatterns scores based on regex patterns
func (s *Scorer) scorePatterns(issue *model.Issue) float64 {
	var score float64
	text := strings.ToLower(issue.Title + " " + issue.Body)
	
	for _, pattern := range s.patterns {
		if pattern.MatchString(text) {
			score += 5.0
		}
	}
	
	// Cap at 25 points
	if score > 25 {
		score = 25
	}
	
	return score
}

// scoreLabels scores based on issue labels
func (s *Scorer) scoreLabels(issue *model.Issue) float64 {
	var score float64
	
	for _, label := range issue.Labels {
		labelName := strings.ToLower(label.Name)
		
		switch {
		case containsSlice(s.priorityLabels, labelName):
			score += 2.0
		case strings.Contains(labelName, "bug") || strings.Contains(labelName, "error"):
			score += 1.5
		case strings.Contains(labelName, "performance"):
			score += 1.5
		case strings.Contains(labelName, "critical") || strings.Contains(labelName, "urgent"):
			score += 2.0
		}
	}
	
	// Cap at 20 points
	if score > 20 {
		score = 20
	}
	
	return score
}

// scoreStatus scores based on issue status
func (s *Scorer) scoreStatus(issue *model.Issue) float64 {
	var score float64
	
	// Open issues are more valuable (ongoing problems)
	if issue.State == "open" {
		score += 10.0
	}
	
	return score
}

// scoreActivity scores based on comments and reactions
func (s *Scorer) scoreActivity(issue *model.Issue) float64 {
	var score float64
	
	// Comments indicate community interest
	if issue.Comments > 0 {
		score += float64(issue.Comments) * 0.5
	}
	
	// Reactions indicate community engagement
	if issue.Reactions > 0 {
		score += float64(issue.Reactions) * 1.0
	}
	
	// Cap at 15 points
	if score > 15 {
		score = 15
	}
	
	return score
}

// containsSlice checks if a slice contains a specific string
func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
