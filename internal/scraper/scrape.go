package scraper

import (
	"fmt"
	"log"
	"time"
)

// ScrapeRepo scrapes a repository for pitfall issues
func ScrapeRepo(client *GithubClient, owner, repoName string, keywords []string) ([]PitfallIssue, error) {
	log.Printf("Starting to scrape repository: %s/%s", owner, repoName)
	
	// Initialize components
	scorer := NewPitfallScorer()
	filter := NewIssueFilter(FilterOptions{
		MinScore:      10.0,   // Minimum score for high-value issues
		MinComments:   1,      // At least 1 comment indicates discussion
		MinAgeDays:    0,      // No minimum age requirement
		ExcludeLabels: []string{"question", "documentation", "good first issue"},
		IncludeLabels: []string{"bug", "performance", "critical"},
		RequireLabel:  false,
	})
	
	var allIssues []PitfallIssue
	
	// Scrape multiple pages to get comprehensive results
	const maxPages = 5
	const issuesPerPage = 100
	
	for page := 1; page <= maxPages; page++ {
		log.Printf("Fetching page %d of %s/%s...", page, owner, repoName)
		
		issues, err := client.GetIssues(owner, repoName, page, issuesPerPage)
		if err != nil {
			log.Printf("Error fetching issues for page %d: %v", page, err)
			break
		}
		
		if len(issues) == 0 {
			log.Printf("No more issues found on page %d", page)
			break
		}
		
		// Filter and score issues
		filteredIssues := filter.FilterIssues(issues, scorer, keywords)
		
		// Set repository information
		for i := range filteredIssues {
			filteredIssues[i].RepoOwner = owner
			filteredIssues[i].RepoName = repoName
		}
		
		allIssues = append(allIssues, filteredIssues...)
		
		log.Printf("Found %d pitfall issues on page %d", len(filteredIssues), page)
		
		// Small delay to avoid rate limiting
		if page < maxPages {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	log.Printf("Completed scraping %s/%s. Found %d total pitfall issues.", owner, repoName, len(allIssues))
	
	// Sort by score (highest first)
	allIssues = sortIssuesByScore(allIssues)
	
	return allIssues, nil
}

// sortIssuesByScore sorts issues by their pitfall score in descending order
func sortIssuesByScore(issues []PitfallIssue) []PitfallIssue {
	// Simple bubble sort for demonstration (in production, use sort.Slice)
	for i := 0; i < len(issues); i++ {
		for j := i + 1; j < len(issues); j++ {
			if issues[i].Score < issues[j].Score {
				issues[i], issues[j] = issues[j], issues[i]
			}
		}
	}
	return issues
}

// ScrapeMultipleRepos scrapes multiple repositories concurrently
func ScrapeMultipleRepos(client *GithubClient, repositories []Repository, keywords []string) ([]PitfallIssue, error) {
	var allResults []PitfallIssue
	
	// Use semaphore pattern to limit concurrent requests
	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent)
	
	resultsChan := make(chan []PitfallIssue, len(repositories))
	errorChan := make(chan error, len(repositories))
	
	for _, repo := range repositories {
		go func(r Repository) {
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore
			
			issues, err := ScrapeRepo(client, r.Owner, r.Name, keywords)
			if err != nil {
				errorChan <- fmt.Errorf("error scraping %s/%s: %v", r.Owner, r.Name, err)
				return
			}
			
			resultsChan <- issues
		}(repo)
	}
	
	// Collect results
	for i := 0; i < len(repositories); i++ {
		select {
		case issues := <-resultsChan:
			allResults = append(allResults, issues...)
		case err := <-errorChan:
			log.Printf("Repository scraping error: %v", err)
		}
	}
	
	// Sort all results by score
	allResults = sortIssuesByScore(allResults)
	
	return allResults, nil
}

// Repository represents a GitHub repository
type Repository struct {
	Owner string
	Name  string
}

// ScrapingOptions defines options for scraping configuration
type ScrapingOptions struct {
	MaxPages       int
	IssuesPerPage  int
	RateLimitDelay time.Duration
	MinScore       float64
	MinComments    int
	Concurrency    int
}

// DefaultScrapingOptions returns default scraping configuration
func DefaultScrapingOptions() ScrapingOptions {
	return ScrapingOptions{
		MaxPages:       5,
		IssuesPerPage:  100,
		RateLimitDelay: 100 * time.Millisecond,
		MinScore:       10.0,
		MinComments:    1,
		Concurrency:    3,
	}
}

// AdvancedScraper provides advanced scraping capabilities
type AdvancedScraper struct {
	client   *GithubClient
	scorer   *PitfallScraper
	filter   *AdvancedIssueFilter
	options  ScrapingOptions
}

// NewAdvancedScraper creates a new advanced scraper
func NewAdvancedScraper(client *GithubClient, options ScrapingOptions) *AdvancedScraper {
	filterOptions := FilterOptions{
		MinScore:       options.MinScore,
		MinComments:    options.MinComments,
		MinAgeDays:     0,
		ExcludeLabels:  []string{"question", "documentation", "good first issue"},
		IncludeLabels:  []string{"bug", "performance", "critical"},
		RequireLabel:   false,
	}
	
	advancedFilter := NewAdvancedIssueFilter(filterOptions, AdvancedFilterOptions{
		MinReactions:  1,
		HasMilestone:  false,
		HasAssignee:   false,
	})
	
	return &AdvancedScraper{
		client:  client,
		scorer:  NewPitfallScorer(),
		filter:  advancedFilter,
		options: options,
	}
}

// ScrapeWithAdvancedOptions performs advanced scraping with custom options
func (as *AdvancedScraper) ScrapeWithAdvancedOptions(owner, repoName string, keywords []string) ([]PitfallIssue, error) {
	log.Printf("Starting advanced scraping of %s/%s", owner, repoName)
	
	var allIssues []PitfallIssue
	
	for page := 1; page <= as.options.MaxPages; page++ {
		log.Printf("Fetching page %d of %s/%s...", page, owner, repoName)
		
		issues, err := as.client.GetIssues(owner, repoName, page, as.options.IssuesPerPage)
		if err != nil {
			log.Printf("Error fetching issues for page %d: %v", page, err)
			break
		}
		
		if len(issues) == 0 {
			log.Printf("No more issues found on page %d", page)
			break
		}
		
		// Apply advanced filtering
		filteredIssues := as.filter.FilterIssuesWithAdvancedLogic(issues, as.scorer, keywords)
		
		// Set repository information
		for i := range filteredIssues {
			filteredIssues[i].RepoOwner = owner
			filteredIssues[i].RepoName = repoName
		}
		
		allIssues = append(allIssues, filteredIssues...)
		log.Printf("Found %d pitfall issues on page %d", len(filteredIssues), page)
		
		// Rate limiting
		if page < as.options.MaxPages {
			time.Sleep(as.options.RateLimitDelay)
		}
	}
	
	log.Printf("Advanced scraping completed for %s/%s. Found %d issues.", owner, repoName, len(allIssues))
	
	return sortIssuesByScore(allIssues), nil
}

// GetScrapingStats returns statistics about the scraping process
func GetScrapingStats(issues []PitfallIssue) map[string]interface{} {
	if len(issues) == 0 {
		return map[string]interface{}{
			"total_issues":      0,
			"average_score":     0.0,
			"high_value_count":  0,
			"repos_covered":     []string{},
			"top_keywords":      []string{},
		}
	}
	
	// Calculate statistics
	totalScore := 0.0
	highValueCount := 0
	keywordCount := make(map[string]int)
	repoCount := make(map[string]int)
	
	for _, issue := range issues {
		totalScore += issue.Score
		if issue.Score >= 20.0 {
			highValueCount++
		}
		
		// Count keywords
		for _, keyword := range issue.Keywords {
			keywordCount[keyword]++
		}
		
		// Count repositories
		repoKey := fmt.Sprintf("%s/%s", issue.RepoOwner, issue.RepoName)
		repoCount[repoKey]++
	}
	
	// Get top keywords
	topKeywords := getTopKeywords(keywordCount, 10)
	
	// Get covered repositories
	var coveredRepos []string
	for repo := range repoCount {
		coveredRepos = append(coveredRepos, repo)
	}
	
	return map[string]interface{}{
		"total_issues":     len(issues),
		"average_score":    totalScore / float64(len(issues)),
		"high_value_count": highValueCount,
		"repos_covered":    coveredRepos,
		"top_keywords":     topKeywords,
		"score_distribution": map[string]int{
			"very_high": countIssuesByScoreRange(issues, 25.0, 100.0),
			"high":      countIssuesByScoreRange(issues, 20.0, 25.0),
			"medium":    countIssuesByScoreRange(issues, 15.0, 20.0),
			"low":       countIssuesByScoreRange(issues, 10.0, 15.0),
		},
	}
}

// Helper functions
func getTopKeywords(keywordCount map[string]int, limit int) []string {
	type keywordFreq struct {
		Keyword string
		Freq    int
	}
	
	var freqs []keywordFreq
	for keyword, count := range keywordCount {
		freqs = append(freqs, keywordFreq{Keyword: keyword, Freq: count})
	}
	
	// Sort by frequency
	for i := 0; i < len(freqs); i++ {
		for j := i + 1; j < len(freqs); j++ {
			if freqs[i].Freq < freqs[j].Freq {
				freqs[i], freqs[j] = freqs[j], freqs[i]
			}
		}
	}
	
	// Return top keywords
	var topKeywords []string
	for i := 0; i < len(freqs) && i < limit; i++ {
		topKeywords = append(topKeywords, freqs[i].Keyword)
	}
	
	return topKeywords
}

func countIssuesByScoreRange(issues []PitfallIssue, min, max float64) int {
	count := 0
	for _, issue := range issues {
		if issue.Score >= min && issue.Score < max {
			count++
		}
	}
	return count
}