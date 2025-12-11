package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
)

// ScrapeRepo scrapes a repository for pitfall issues and persists to database
func ScrapeRepo(client *GithubClient, dbService *DatabaseService, owner, repoName string, keywords []string) ([]PitfallIssue, error) {
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
	var persistMutex sync.Mutex
	var persistCount int
	
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
		
		// Persist to database in background
		go func(issues []PitfallIssue) {
			if dbService != nil && len(issues) > 0 {
				batchSize := 50
				for i := 0; i < len(issues); i += batchSize {
					end := i + batchSize
					if end > len(issues) {
						end = len(issues)
					}
					
					batch := issues[i:end]
					count, err := dbService.PersistIssuesBatch(batch)
					if err != nil {
						log.Printf("Error persisting batch: %v", err)
					} else {
						persistMutex.Lock()
						persistCount += count
						persistMutex.Unlock()
					}
					
					// Small delay to avoid overwhelming the database
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(filteredIssues)
		
		allIssues = append(allIssues, filteredIssues...)
		
		log.Printf("Found %d pitfall issues on page %d", len(filteredIssues), page)
		
		// Small delay to avoid rate limiting
		if page < maxPages {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	log.Printf("Completed scraping %s/%s. Found %d total pitfall issues, persisted %d", 
		owner, repoName, len(allIssues), persistCount)
	
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

// ScrapeMultipleRepos scrapes multiple repositories concurrently with database integration
func ScrapeMultipleRepos(client *GithubClient, dbService *DatabaseService, repositories []Repository, keywords []string) ([]PitfallIssue, error) {
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
			
			issues, err := ScrapeRepo(client, dbService, r.Owner, r.Name, keywords)
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

// DatabaseService provides database operations for scraper
type DatabaseService struct {
	db                    *sql.DB
	crudOps               database.CRUDOperations
	dedupService          *database.DeduplicationService
	classificationService *database.ClassificationService
	logger                *log.Logger
}

// NewDatabaseService creates a new database service
func NewDatabaseService(db *sql.DB) *DatabaseService {
	logger := log.New(log.Writer(), "[DB] ", log.LstdFlags)
	
	return &DatabaseService{
		db:                    db,
		crudOps:               database.NewCRUDOperations(db),
		dedupService:          database.NewDeduplicationService(db, database.DefaultDeduplicationConfig()),
		classificationService: database.NewClassificationService(db, database.DefaultClassificationConfig()),
		logger:                logger,
	}
}

// PersistIssue persists a pitfall issue to database
func (ds *DatabaseService) PersistIssue(issue *PitfallIssue) error {
	// Convert to database model
	dbIssue := ds.convertToDatabaseIssue(issue)
	
	// Check for duplicates first
	similar, similarity, err := ds.dedupService.FindSimilarIssues(&dbIssue, 1)
	if err != nil {
		ds.logger.Printf("Error checking for duplicates: %v", err)
		// Continue without duplicate check
	} else if len(similar) > 0 && similarity > 0.75 {
		ds.logger.Printf("Issue %d is duplicate of issue %d (similarity: %.2f)", 
			issue.Number, similar[0].Number, similarity)
		return fmt.Errorf("duplicate issue detected")
	}
	
	// Auto-classify the issue
	if ds.classificationService != nil {
		_, err := ds.classificationService.ClassifySingleIssue(&dbIssue)
		if err != nil {
			ds.logger.Printf("Error classifying issue: %v", err)
			// Continue without classification
		}
	}
	
	// Persist to database
	_, err = ds.crudOps.CreateIssue(&dbIssue)
	if err != nil {
		return fmt.Errorf("failed to persist issue: %w", err)
	}
	
	ds.logger.Printf("Successfully persisted issue #%d: %s", issue.Number, issue.Title)
	return nil
}

// PersistIssuesBatch persists multiple issues in a single transaction
func (ds *DatabaseService) PersistIssuesBatch(issues []*PitfallIssue) (int, error) {
	if len(issues) == 0 {
		return 0, nil
	}
	
	// Convert to database models
	dbIssues := make([]*database.Issue, len(issues))
	for i, issue := range issues {
		dbIssues[i] = ds.convertToDatabaseIssue(issue)
	}
	
	// Check for duplicates and classify
	for i, dbIssue := range dbIssues {
		// Check for duplicates
		similar, similarity, err := ds.dedupService.FindSimilarIssues(dbIssue, 1)
		if err != nil {
			ds.logger.Printf("Error checking for duplicates: %v", err)
			continue
		}
		
		if len(similar) > 0 && similarity > 0.75 {
			ds.logger.Printf("Issue %d is duplicate, skipping", dbIssue.Number)
			continue
		}
		
		// Auto-classify
		if ds.classificationService != nil {
			_, err := ds.classificationService.ClassifySingleIssue(dbIssue)
			if err != nil {
				ds.logger.Printf("Error classifying issue: %v", err)
			}
		}
	}
	
	// Batch insert
	ids, err := ds.crudOps.CreateIssues(dbIssues)
	if err != nil {
		return 0, fmt.Errorf("failed to batch persist issues: %w", err)
	}
	
	ds.logger.Printf("Successfully persisted %d issues", len(ids))
	return len(ids), nil
}

// convertToDatabaseIssue converts PitfallIssue to database.Issue
func (ds *DatabaseService) convertToDatabaseIssue(issue *PitfallIssue) database.Issue {
	return database.Issue{
		IssueID:       int64(issue.ID),
		Number:        issue.Number,
		Title:         issue.Title,
		Body:          issue.Body,
		URL:           issue.URL,
		State:         issue.State,
		AuthorLogin:   issue.Assignee, // Map assignee to author login
		Labels:        database.JSONSlice{},
		Assignees:     database.JSONSlice{issue.Assignee},
		Milestone:     issue.Milestone,
		Reactions: database.ReactionCount{
			Total: issue.Reactions,
		},
		CreatedAt:      issue.CreatedAt,
		UpdatedAt:      issue.UpdatedAt,
		FirstSeenAt:    issue.CreatedAt,
		LastSeenAt:     issue.UpdatedAt,
		CommentsCount:  issue.Comments,
		Score:          issue.Score,
		URL:            issue.URL,
		HTMLURL:        issue.URL,
		RepoOwner:      issue.RepoOwner,
		RepoName:       issue.RepoName,
		Keywords:       database.JSONSlice(issue.Keywords),
	}
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

// AdvancedScraper provides advanced scraping capabilities with database integration
type AdvancedScraper struct {
	client      *GithubClient
	scorer      *PitfallScorer
	filter      *AdvancedIssueFilter
	options     ScrapingOptions
	dbService   *DatabaseService
}

// NewAdvancedScraper creates a new advanced scraper with database support
func NewAdvancedScraper(client *GithubClient, dbService *DatabaseService, options ScrapingOptions) *AdvancedScraper {
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
		client:    client,
		scorer:    NewPitfallScorer(),
		filter:    advancedFilter,
		options:   options,
		dbService: dbService,
	}
}

// ScrapeWithAdvancedOptions performs advanced scraping with custom options and database persistence
func (as *AdvancedScraper) ScrapeWithAdvancedOptions(owner, repoName string, keywords []string) ([]PitfallIssue, error) {
	log.Printf("Starting advanced scraping of %s/%s", owner, repoName)
	
	var allIssues []PitfallIssue
	var persistCount int
	var wg sync.WaitGroup
	
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
		
		// Persist to database in background
		if as.dbService != nil && len(filteredIssues) > 0 {
			wg.Add(1)
			go func(issues []PitfallIssue) {
				defer wg.Done()
				count, err := as.dbService.PersistIssuesBatch(issues)
				if err != nil {
					log.Printf("Error persisting issues to database: %v", err)
				} else {
					persistCount += count
				}
			}(filteredIssues)
		}
		
		// Rate limiting
		if page < as.options.MaxPages {
			time.Sleep(as.options.RateLimitDelay)
		}
	}
	
	// Wait for all persistence operations to complete
	wg.Wait()
	
	log.Printf("Advanced scraping completed for %s/%s. Found %d issues, persisted %d", 
		owner, repoName, len(allIssues), persistCount)
	
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