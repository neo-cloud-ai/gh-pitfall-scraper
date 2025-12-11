package scraper

import (
	"strings"
	"time"
	
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/model"
)

// Filter filters issues based on various criteria
type Filter struct {
	MinScore      float64
	MinAge        time.Duration
	MaxAge        time.Duration
	RequiredState string
	MaxIssues     int
}

// FilterConfig represents filtering configuration
type FilterConfig struct {
	MinScore      float64         `yaml:"min_score"`
	MinAge        string          `yaml:"min_age"`
	MaxAge        string          `yaml:"max_age"`
	RequiredState string          `yaml:"required_state"`
	MaxIssues     int             `yaml:"max_issues"`
}

// NewFilter creates a new issue filter
func NewFilter(config FilterConfig) *Filter {
	filter := &Filter{
		MinScore:      config.MinScore,
		RequiredState: config.RequiredState,
		MaxIssues:     config.MaxIssues,
	}
	
	// Parse age durations
	if config.MinAge != "" {
		if duration, err := time.ParseDuration(config.MinAge); err == nil {
			filter.MinAge = duration
		}
	}
	
	if config.MaxAge != "" {
		if duration, err := time.ParseDuration(config.MaxAge); err == nil {
			filter.MaxAge = duration
		}
	}
	
	return filter
}

// FilterIssues filters and ranks issues based on criteria
func (f *Filter) FilterIssues(issues []model.Issue, scorer *Scorer) []model.Issue {
	var filtered []model.Issue
	
	for _, issue := range issues {
		if f.shouldInclude(issue) {
			// Calculate score
			score, reasons := scorer.ScoreIssue(&issue)
			issue.Score = score
			issue.ScoreReason = reasons
			
			// Apply minimum score filter
			if score >= f.MinScore {
				filtered = append(filtered, issue)
			}
		}
	}
	
	// Sort by score (descending)
	f.sortByScore(filtered)
	
	// Apply maximum issues limit
	if f.MaxIssues > 0 && len(filtered) > f.MaxIssues {
		filtered = filtered[:f.MaxIssues]
	}
	
	return filtered
}

// shouldInclude determines if an issue should be included
func (f *Filter) shouldInclude(issue model.Issue) bool {
	// Check state
	if f.RequiredState != "" && issue.State != f.RequiredState {
		return false
	}
	
	// Check age
	age := time.Since(issue.CreatedAt)
	
	if f.MinAge > 0 && age < f.MinAge {
		return false
	}
	
	if f.MaxAge > 0 && age > f.MaxAge {
		return false
	}
	
	return true
}

// sortByScore sorts issues by score in descending order
func (f *Filter) sortByScore(issues []model.Issue) {
	// Simple bubble sort for now (can be optimized)
	for i := 0; i < len(issues); i++ {
		for j := i + 1; j < len(issues); j++ {
			if issues[i].Score < issues[j].Score {
				issues[i], issues[j] = issues[j], issues[i]
			}
		}
	}
}

// FilterByRepository filters issues for specific repositories
func (f *Filter) FilterByRepository(allIssues map[string][]model.Issue, repos []model.Repository) map[string][]model.Issue {
	result := make(map[string][]model.Issue)
	scorer := NewScorer()
	
	for _, repo := range repos {
		if repo.Enabled {
			issues := allIssues[repo.FullName]
			filtered := f.FilterIssues(issues, scorer)
			result[repo.FullName] = filtered
		}
	}
	
	return result
}

// GetHighValueIssues returns issues with score above threshold
func (f *Filter) GetHighValueIssues(issues []model.Issue, threshold float64) []model.Issue {
	var highValue []model.Issue
	
	for _, issue := range issues {
		if issue.Score >= threshold {
			highValue = append(highValue, issue)
		}
	}
	
	return highValue
}

// CategorizeIssues categorizes issues by type
func (f *Filter) CategorizeIssues(issues []model.Issue) map[string][]model.Issue {
	categories := make(map[string][]model.Issue)
	
	// Define categories based on keywords and patterns
	categoryRules := map[string][]string{
		"performance":     {"performance", "speed", "slow", "optimization", "throughput", "latency"},
		"gpu_memory":      {"gpu", "cuda", "oom", "memory", "fragmentation"},
		"distributed":     {"distributed", "nccl", "multi-gpu", "multi-node", "deadlock"},
		"model_serving":   {"inference", "serving", "kv cache", "prefill", "decode"},
		"crashes":         {"crash", "error", "exception", "kernel", "timeout"},
		"memory_issues":   {"memory leak", "leak", "overflow", "allocation"},
	}
	
	for _, issue := range issues {
		text := issue.Title + " " + issue.Body
		text = strings.ToLower(text)
		
		assigned := false
		for category, keywords := range categoryRules {
			for _, keyword := range keywords {
				if contains(text, keyword) {
					categories[category] = append(categories[category], issue)
					assigned = true
					break
				}
			}
			if assigned {
				break
			}
		}
		
		// If no category assigned, put in "other"
		if !assigned {
			categories["other"] = append(categories["other"], issue)
		}
	}
	
	return categories
}

// Helper functions

func contains(text, keyword string) bool {
	return len(text) >= len(keyword) && 
		   (text == keyword || 
		    len(text) > len(keyword) && (
		    	strings.Contains(text, keyword) ||
		    	hasWordBoundary(text, keyword)))
}

func hasWordBoundary(text, keyword string) bool {
	for i := 0; i <= len(text)-len(keyword); i++ {
		if text[i:i+len(keyword)] == keyword {
			// Check word boundaries
			beforeOK := i == 0 || !isAlphaNum(text[i-1])
			afterOK := i+len(keyword) == len(text) || !isAlphaNum(text[i+len(keyword)])
			if beforeOK && afterOK {
				return true
			}
		}
	}
	return false
}

func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || 
		   (b >= 'A' && b <= 'Z') || 
		   (b >= '0' && b <= '9') || 
		   b == '_'
}