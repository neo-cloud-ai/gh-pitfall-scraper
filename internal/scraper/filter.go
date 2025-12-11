package scraper

import (
	"strings"
	"time"
)

// IssueFilter provides filtering logic for pitfall issues
type IssueFilter struct {
	minScore      float64
	minComments   int
	minAge        time.Duration
	excludeLabels []string
	includeLabels []string
}

// FilterOptions defines filtering configuration
type FilterOptions struct {
	MinScore       float64
	MinComments    int
	MinAgeDays     int
	ExcludeLabels  []string
	IncludeLabels  []string
	RequireLabel   bool
}

// NewIssueFilter creates a new issue filter with default options
func NewIssueFilter(options FilterOptions) *IssueFilter {
	return &IssueFilter{
		minScore:      options.MinScore,
		minComments:   options.MinComments,
		minAge:        time.Duration(options.MinAgeDays) * 24 * time.Hour,
		excludeLabels: options.ExcludeLabels,
		includeLabels: options.IncludeLabels,
	}
}

// FilterIssues filters issues based on pitfall criteria
func (f *IssueFilter) FilterIssues(issues []Issue, scorer *PitfallScorer, keywords []string) []PitfallIssue {
	var filteredIssues []PitfallIssue
	
	for _, issue := range issues {
		// Skip pull requests (GitHub issues API returns both issues and PRs)
		if issue.Number == 0 || issue.Body == "" {
			continue
		}
		
		// Calculate score
		score := scorer.Score(&issue, keywords)
		
		// Apply filters
		if f.passesFilters(&issue, score, keywords) {
			pitfallIssue := f.convertToPitfallIssue(&issue, score, keywords)
			filteredIssues = append(filteredIssues, pitfallIssue)
		}
	}
	
	return filteredIssues
}

// passesFilters checks if an issue passes all filtering criteria
func (f *IssueFilter) passesFilters(issue *Issue, score float64, keywords []string) bool {
	// Score threshold filter
	if score < f.minScore {
		return false
	}
	
	// Comments threshold filter
	if issue.Comments < f.minComments {
		return false
	}
	
	// Age filter (skip very old issues unless they're high-value)
	age := time.Since(issue.CreatedAt)
	if age > f.minAge && score < f.minScore*1.5 {
		return false
	}
	
	// Label exclusion filter
	for _, excludeLabel := range f.excludeLabels {
		if hasLabel(issue.Labels, excludeLabel) {
			return false
		}
	}
	
	// Label inclusion filter
	if len(f.includeLabels) > 0 {
		hasRequiredLabel := false
		for _, includeLabel := range f.includeLabels {
			if hasLabel(issue.Labels, includeLabel) {
				hasRequiredLabel = true
				break
			}
		}
		if !hasRequiredLabel {
			return false
		}
	}
	
	// Minimum keyword match requirement
	matchingKeywords := countMatchingKeywords(issue, keywords)
	if matchingKeywords < 2 {
		return false
	}
	
	// Skip issues without meaningful content
	if len(strings.TrimSpace(issue.Body)) < 50 {
		return false
	}
	
	return true
}

// convertToPitfallIssue converts a GitHub issue to a pitfall issue
func (f *IssueFilter) convertToPitfallIssue(issue *Issue, score float64, keywords []string) PitfallIssue {
	matchingKeywords := getMatchingKeywords(issue, keywords)
	
	return PitfallIssue{
		ID:         issue.ID,
		Number:     issue.Number,
		Title:      issue.Title,
		URL:        issue.URL,
		State:      issue.State,
		Labels:     issue.Labels,
		CreatedAt:  issue.CreatedAt,
		UpdatedAt:  issue.UpdatedAt,
		Keywords:   matchingKeywords,
		Score:      score,
		Comments:   issue.Comments,
		Reactions:  issue.Reactions.TotalCount,
		Assignee:   issue.Assignee.Login,
		Milestone:  issue.Milestone.Title,
		Body:       issue.Body,
		RepoOwner:  "", // Will be set by caller
		RepoName:   "", // Will be set by caller
	}
}

// hasLabel checks if an issue has a specific label
func hasLabel(labels []Label, labelName string) bool {
	for _, label := range labels {
		if strings.EqualFold(label.Name, labelName) {
			return true
		}
	}
	return false
}

// countMatchingKeywords counts how many keywords match an issue
func countMatchingKeywords(issue *Issue, keywords []string) int {
	searchText := strings.ToLower(issue.Title + " " + issue.Body)
	count := 0
	
	for _, keyword := range keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			count++
		}
	}
	
	return count
}

// getMatchingKeywords returns keywords that match an issue
func getMatchingKeywords(issue *Issue, keywords []string) []string {
	var matchingKeywords []string
	searchText := strings.ToLower(issue.Title + " " + issue.Body)
	
	for _, keyword := range keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			matchingKeywords = append(matchingKeywords, keyword)
		}
	}
	
	return matchingKeywords
}

// AdvancedFilterOptions provides additional filtering capabilities
type AdvancedFilterOptions struct {
	TitlePatterns   []string
	BodyPatterns    []string
	AuthorBlacklist []string
	AuthorWhitelist []string
	MinReactions    int
	HasMilestone    bool
	HasAssignee     bool
}

// AdvancedIssueFilter provides sophisticated filtering capabilities
type AdvancedIssueFilter struct {
	basicFilter   *IssueFilter
	advancedOpts  AdvancedFilterOptions
}

// NewAdvancedIssueFilter creates a new advanced issue filter
func NewAdvancedIssueFilter(basicOpts FilterOptions, advancedOpts AdvancedFilterOptions) *AdvancedIssueFilter {
	return &AdvancedIssueFilter{
		basicFilter:  NewIssueFilter(basicOpts),
		advancedOpts: advancedOpts,
	}
}

// FilterIssuesWithAdvancedLogic applies advanced filtering logic
func (af *AdvancedIssueFilter) FilterIssuesWithAdvancedLogic(issues []Issue, scorer *PitfallScorer, keywords []string) []PitfallIssue {
	var filteredIssues []PitfallIssue
	
	for _, issue := range issues {
		// Apply basic filters first
		if !af.basicFilter.passesFilters(&issue, scorer.Score(&issue, keywords), keywords) {
			continue
		}
		
		// Apply advanced filters
		if af.passesAdvancedFilters(&issue) {
			score := scorer.Score(&issue, keywords)
			pitfallIssue := af.basicFilter.convertToPitfallIssue(&issue, score, keywords)
			pitfallIssue.RepoOwner = "" // Will be set by caller
			pitfallIssue.RepoName = ""  // Will be set by caller
			filteredIssues = append(filteredIssues, pitfallIssue)
		}
	}
	
	return filteredIssues
}

// passesAdvancedFilters applies advanced filtering logic
func (af *AdvancedIssueFilter) passesAdvancedFilters(issue *Issue) bool {
	// Title pattern matching
	if len(af.advancedOpts.TitlePatterns) > 0 {
		if !matchesAnyPattern(issue.Title, af.advancedOpts.TitlePatterns) {
			return false
		}
	}
	
	// Body pattern matching
	if len(af.advancedOpts.BodyPatterns) > 0 {
		if !matchesAnyPattern(issue.Body, af.advancedOpts.BodyPatterns) {
			return false
		}
	}
	
	// Author whitelist/blacklist
	if len(af.advancedOpts.AuthorWhitelist) > 0 {
		if !containsString(af.advancedOpts.AuthorWhitelist, issue.Assignee.Login) {
			return false
		}
	}
	
	if len(af.advancedOpts.AuthorBlacklist) > 0 {
		if containsString(af.advancedOpts.AuthorBlacklist, issue.Assignee.Login) {
			return false
		}
	}
	
	// Minimum reactions
	if af.advancedOpts.MinReactions > 0 && issue.Reactions.TotalCount < af.advancedOpts.MinReactions {
		return false
	}
	
	// Milestone requirement
	if af.advancedOpts.HasMilestone && issue.Milestone.Title == "" {
		return false
	}
	
	// Assignee requirement
	if af.advancedOpts.HasAssignee && issue.Assignee.Login == "" {
		return false
	}
	
	return true
}

// matchesAnyPattern checks if text matches any of the given patterns
func matchesAnyPattern(text string, patterns []string) bool {
	lowerText := strings.ToLower(text)
	for _, pattern := range patterns {
		if strings.Contains(lowerText, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// containsString checks if a slice contains a specific string
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}