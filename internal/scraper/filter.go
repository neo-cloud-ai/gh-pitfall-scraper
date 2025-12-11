package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
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

// DatabaseFilterService provides database-based filtering functionality
type DatabaseFilterService struct {
	db           *sql.DB
	crudOps      database.CRUDOperations
	logger       *log.Logger
}

// NewDatabaseFilterService creates a new database filter service
func NewDatabaseFilterService(db *sql.DB) *DatabaseFilterService {
	return &DatabaseFilterService{
		db:      db,
		crudOps: database.NewCRUDOperations(db),
		logger:  log.New(log.Writer(), "[DB-Filter] ", log.LstdFlags),
	}
}

// DatabaseFilterResult represents the result of database filtering
type DatabaseFilterResult struct {
	Issues        []*database.Issue `json:"issues"`
	TotalCount    int               `json:"total_count"`
	FilteredCount int               `json:"filtered_count"`
	FilterApplied string            `json:"filter_applied"`
	ProcessingTime time.Duration    `json:"processing_time"`
}

// FilterFromDatabase filters issues from database using advanced criteria
func (dfs *DatabaseFilterService) FilterFromDatabase(criteria DatabaseFilterCriteria) (*DatabaseFilterResult, error) {
	startTime := time.Now()
	
	search := &database.AdvancedSearch{
		Query:         criteria.Query,
		Keywords:      criteria.Keywords,
		Categories:    criteria.Categories,
		Priorities:    criteria.Priorities,
		TechStacks:    criteria.TechStacks,
		Repos:         criteria.Repos,
		States:        criteria.States,
		MinScore:      criteria.MinScore,
		MaxScore:      criteria.MaxScore,
		DateFrom:      criteria.DateFrom,
		DateTo:        criteria.DateTo,
		SortBy:        criteria.SortBy,
		SortOrder:     criteria.SortOrder,
		Limit:         criteria.Limit,
		Offset:        criteria.Offset,
		ExcludeDuplicates: criteria.ExcludeDuplicates,
	}
	
	issues, err := dfs.crudOps.SearchIssuesAdvanced(search)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}
	
	// Apply additional real-time filtering if specified
	if criteria.RealTimeFilter != nil {
		issues = dfs.applyRealTimeFilter(issues, criteria.RealTimeFilter)
	}
	
	processingTime := time.Since(startTime)
	
	result := &DatabaseFilterResult{
		Issues:         issues,
		TotalCount:     len(issues),
		FilteredCount:  len(issues),
		FilterApplied:  criteria.String(),
		ProcessingTime: processingTime,
	}
	
	dfs.logger.Printf("Database filtering completed: %d issues found in %v", len(issues), processingTime)
	return result, nil
}

// GetFilteredStatsFromDatabase gets statistics for filtered issues
func (dfs *DatabaseFilterService) GetFilteredStatsFromDatabase(criteria DatabaseFilterCriteria) (map[string]interface{}, error) {
	result, err := dfs.FilterFromDatabase(criteria)
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]interface{})
	stats["total_issues"] = len(result.Issues)
	stats["average_score"] = dfs.calculateAverageScore(result.Issues)
	stats["categories"] = dfs.groupByCategory(result.Issues)
	stats["priorities"] = dfs.groupByPriority(result.Issues)
	stats["repos"] = dfs.groupByRepository(result.Issues)
	stats["score_distribution"] = dfs.calculateScoreDistribution(result.Issues)
	
	return stats, nil
}

// DatabaseFilterCriteria represents criteria for database filtering
type DatabaseFilterCriteria struct {
	Query             string    `json:"query"`
	Keywords          []string  `json:"keywords"`
	Categories        []string  `json:"categories"`
	Priorities        []string  `json:"priorities"`
	TechStacks        []string  `json:"tech_stacks"`
	Repos             []string  `json:"repos"`
	States            []string  `json:"states"`
	MinScore          *float64  `json:"min_score"`
	MaxScore          *float64  `json:"max_score"`
	DateFrom          *time.Time `json:"date_from"`
	DateTo            *time.Time `json:"date_to"`
	SortBy            string    `json:"sort_by"`
	SortOrder         string    `json:"sort_order"`
	Limit             int       `json:"limit"`
	Offset            int       `json:"offset"`
	ExcludeDuplicates bool      `json:"exclude_duplicates"`
	RealTimeFilter    *RealTimeFilter `json:"-"`
}

// RealTimeFilter represents real-time filtering criteria
type RealTimeFilter struct {
	MinKeywords      int     `json:"min_keywords"`
	RequireAllLabels bool    `json:"require_all_labels"`
	ExcludePatterns  []string `json:"exclude_patterns"`
	IncludePatterns  []string `json:"include_patterns"`
}

// String returns string representation of filter criteria
func (c *DatabaseFilterCriteria) String() string {
	parts := []string{}
	
	if c.Query != "" {
		parts = append(parts, fmt.Sprintf("query:%s", c.Query))
	}
	if len(c.Keywords) > 0 {
		parts = append(parts, fmt.Sprintf("keywords:%v", c.Keywords))
	}
	if len(c.Categories) > 0 {
		parts = append(parts, fmt.Sprintf("categories:%v", c.Categories))
	}
	if len(c.Priorities) > 0 {
		parts = append(parts, fmt.Sprintf("priorities:%v", c.Priorities))
	}
	if len(c.TechStacks) > 0 {
		parts = append(parts, fmt.Sprintf("tech_stacks:%v", c.TechStacks))
	}
	if len(c.Repos) > 0 {
		parts = append(parts, fmt.Sprintf("repos:%v", c.Repos))
	}
	if len(c.States) > 0 {
		parts = append(parts, fmt.Sprintf("states:%v", c.States))
	}
	if c.MinScore != nil {
		parts = append(parts, fmt.Sprintf("min_score:%.1f", *c.MinScore))
	}
	if c.MaxScore != nil {
		parts = append(parts, fmt.Sprintf("max_score:%.1f", *c.MaxScore))
	}
	
	return strings.Join(parts, ",")
}

// DefaultDatabaseFilterCriteria returns default filter criteria
func DefaultDatabaseFilterCriteria() DatabaseFilterCriteria {
	return DatabaseFilterCriteria{
		States:         []string{"open"},
		SortBy:         "score",
		SortOrder:      "DESC",
		Limit:          100,
		Offset:         0,
		ExcludeDuplicates: true,
	}
}

// applyRealTimeFilter applies real-time filtering to issues
func (dfs *DatabaseFilterService) applyRealTimeFilter(issues []*database.Issue, filter *RealTimeFilter) []*database.Issue {
	var filtered []*database.Issue
	
	for _, issue := range issues {
		if dfs.passesRealTimeFilter(issue, filter) {
			filtered = append(filtered, issue)
		}
	}
	
	return filtered
}

// passesRealTimeFilter checks if an issue passes real-time filter
func (dfs *DatabaseFilterService) passesRealTimeFilter(issue *database.Issue, filter *RealTimeFilter) bool {
	// Minimum keywords check
	if filter.MinKeywords > 0 && len(issue.Keywords) < filter.MinKeywords {
		return false
	}
	
	// Pattern exclusion
	for _, pattern := range filter.ExcludePatterns {
		if strings.Contains(strings.ToLower(issue.Title), strings.ToLower(pattern)) ||
		   strings.Contains(strings.ToLower(issue.Body), strings.ToLower(pattern)) {
			return false
		}
	}
	
	// Pattern inclusion
	if len(filter.IncludePatterns) > 0 {
		hasMatch := false
		for _, pattern := range filter.IncludePatterns {
			if strings.Contains(strings.ToLower(issue.Title), strings.ToLower(pattern)) ||
			   strings.Contains(strings.ToLower(issue.Body), strings.ToLower(pattern)) {
				hasMatch = true
				break
			}
		}
		if !hasMatch {
			return false
		}
	}
	
	return true
}

// Helper methods for statistics
func (dfs *DatabaseFilterService) calculateAverageScore(issues []*database.Issue) float64 {
	if len(issues) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, issue := range issues {
		total += issue.Score
	}
	return total / float64(len(issues))
}

func (dfs *DatabaseFilterService) groupByCategory(issues []*database.Issue) map[string]int {
	groups := make(map[string]int)
	for _, issue := range issues {
		category := issue.Category
		if category == "" {
			category = "uncategorized"
		}
		groups[category]++
	}
	return groups
}

func (dfs *DatabaseFilterService) groupByPriority(issues []*database.Issue) map[string]int {
	groups := make(map[string]int)
	for _, issue := range issues {
		priority := issue.Priority
		if priority == "" {
			priority = "unassigned"
		}
		groups[priority]++
	}
	return groups
}

func (dfs *DatabaseFilterService) groupByRepository(issues []*database.Issue) map[string]int {
	groups := make(map[string]int)
	for _, issue := range issues {
		repoKey := fmt.Sprintf("%s/%s", issue.RepoOwner, issue.RepoName)
		groups[repoKey]++
	}
	return groups
}

func (dfs *DatabaseFilterService) calculateScoreDistribution(issues []*database.Issue) map[string]int {
	distribution := map[string]int{
		"very_high": 0, // >= 25
		"high":      0, // 20-25
		"medium":    0, // 15-20
		"low":       0, // 10-15
		"very_low":  0, // < 10
	}
	
	for _, issue := range issues {
		switch {
		case issue.Score >= 25:
			distribution["very_high"]++
		case issue.Score >= 20:
			distribution["high"]++
		case issue.Score >= 15:
			distribution["medium"]++
		case issue.Score >= 10:
			distribution["low"]++
		default:
			distribution["very_low"]++
		}
	}
	
	return distribution
}