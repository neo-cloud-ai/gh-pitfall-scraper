package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

// Example demonstrating database-integrated scraping
func main() {
	// Initialize database
	db, err := initDatabase("./data/example.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	// Initialize database services
	dbService := scraper.NewDatabaseService(db)
	githubClient := scraper.NewGithubClientWithDB("your_github_token", dbService)
	databaseScorer := scraper.NewDatabaseScorer(db)
	
	// Define repositories and keywords
	repositories := []scraper.Repository{
		{Owner: "vllm-project", Name: "vllm"},
		{Owner: "sgl-project", Name: "sglang"},
	}
	
	keywords := []string{
		"performance", "regression", "latency", "throughput",
		"OOM", "memory leak", "CUDA", "kernel", "NCCL",
		"hang", "deadlock", "kv cache",
	}
	
	fmt.Println("Starting database-integrated scraping...")
	
	// Scrape repositories with database persistence
	issues, err := scraper.ScrapeMultipleRepos(githubClient, dbService, repositories, keywords)
	if err != nil {
		log.Printf("Scraping error: %v", err)
	}
	
	fmt.Printf("Scraping completed. Found %d issues.\n", len(issues))
	
	// Perform deduplication
	fmt.Println("Running deduplication...")
	dedupResult, err := githubClient.RunDeduplication()
	if err != nil {
		log.Printf("Deduplication error: %v", err)
	} else {
		fmt.Printf("Deduplication completed: %d duplicates found\n", dedupResult.DuplicatesFound)
	}
	
	// Run classification
	fmt.Println("Running classification...")
	dbIssues := convertToDatabaseIssues(issues)
	classResult, err := githubClient.RunClassification(dbIssues)
	if err != nil {
		log.Printf("Classification error: %v", err)
	} else {
		fmt.Printf("Classification completed: %d issues classified\n", classResult.Classified)
	}
	
	// Query and filter from database
	fmt.Println("Querying database for high-score issues...")
	filter := scraper.DefaultDatabaseFilterCriteria()
	filter.MinScore = &[]float64{20.0}[0] // High-score threshold
	filter.Limit = 50
	
	filterResult, err := queryDatabaseIssues(dbService, filter)
	if err != nil {
		log.Printf("Database query error: %v", err)
	} else {
		fmt.Printf("Found %d high-score issues in database\n", filterResult.FilteredCount)
	}
	
	// Get statistics
	fmt.Println("Getting statistics...")
	stats, err := getScrapingStatistics(dbService, githubClient)
	if err != nil {
		log.Printf("Statistics error: %v", err)
	} else {
		fmt.Printf("Statistics: %+v\n", stats)
	}
	
	fmt.Println("Database integration example completed.")
}

// initDatabase initializes the database with proper schema
func initDatabase(path string) (*sql.DB, error) {
	// Initialize database connection
	dbConfig := database.DefaultDatabaseConfig()
	dbConfig.Path = path
	
	db, err := database.NewSQLiteDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	
	// Initialize schema
	if err := database.InitializeSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	return db, nil
}

// queryDatabaseIssues demonstrates database querying and filtering
func queryDatabaseIssues(dbService *scraper.DatabaseService, criteria scraper.DatabaseFilterCriteria) (*scraper.DatabaseFilterResult, error) {
	// Note: This would require implementing the DatabaseFilterService
	// For now, we'll use the CRUD operations directly
	
	// Get high-score issues
	issues, err := dbService.CRUDOperations.GetIssuesByScore(20.0, 100.0, 50, 0)
	if err != nil {
		return nil, err
	}
	
	return &scraper.DatabaseFilterResult{
		Issues:         issues,
		TotalCount:     len(issues),
		FilteredCount:  len(issues),
		FilterApplied:  "high_score",
		ProcessingTime: 0,
	}, nil
}

// getScrapingStatistics retrieves comprehensive statistics
func getScrapingStatistics(dbService *scraper.DatabaseService, client *scraper.GithubClient) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get general issue statistics
	issueStats, err := dbService.CRUDOperations.GetIssueStats()
	if err != nil {
		return nil, err
	}
	stats["issues"] = issueStats
	
	// Get duplicate statistics
	duplicateStats, err := client.GetDuplicateStatsFromDatabase()
	if err != nil {
		log.Printf("Failed to get duplicate stats: %v", err)
	} else {
		stats["duplicates"] = duplicateStats
	}
	
	// Get classification statistics
	classificationStats, err := client.GetClassificationStatsFromDatabase()
	if err != nil {
		log.Printf("Failed to get classification stats: %v", err)
	} else {
		stats["classification"] = classificationStats
	}
	
	// Add timestamp
	stats["retrieved_at"] = time.Now()
	
	return stats, nil
}

// convertToDatabaseIssues converts pitfall issues to database issues
func convertToDatabaseIssues(issues []scraper.PitfallIssue) []*database.Issue {
	dbIssues := make([]*database.Issue, len(issues))
	
	for i, issue := range issues {
		dbIssues[i] = &database.Issue{
			IssueID:      int64(issue.ID),
			Number:       issue.Number,
			Title:        issue.Title,
			Body:         issue.Body,
			URL:          issue.URL,
			State:        issue.State,
			AuthorLogin:  issue.Assignee,
			Labels:       database.JSONSlice{},
			Assignees:    database.JSONSlice{issue.Assignee},
			Milestone:    issue.Milestone,
			Reactions: database.ReactionCount{
				Total: issue.Reactions,
			},
			CreatedAt:     issue.CreatedAt,
			UpdatedAt:     issue.UpdatedAt,
			FirstSeenAt:   issue.CreatedAt,
			LastSeenAt:    issue.UpdatedAt,
			CommentsCount: issue.Comments,
			Score:         issue.Score,
			URL:           issue.URL,
			HTMLURL:       issue.URL,
			RepoOwner:     issue.RepoOwner,
			RepoName:      issue.RepoName,
			Keywords:      database.JSONSlice(issue.Keywords),
		}
	}
	
	return dbIssues
}

// Example of advanced scraping with database integration
func advancedScrapingExample() {
	// This function demonstrates advanced scraping capabilities
	fmt.Println("Advanced scraping example:")
	
	// Configure advanced options
	options := scraper.DefaultScrapingOptions()
	options.MaxPages = 3
	options.IssuesPerPage = 50
	options.Concurrency = 2
	
	// Initialize advanced scraper with database
	// Note: This would require a database instance
	// advancedScraper := scraper.NewAdvancedScraper(client, dbService, options)
	
	fmt.Println("Advanced scraping configured with database integration")
}