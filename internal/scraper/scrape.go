package scraper

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/client"
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/model"
	"github.com/google/go-github/v67/github"
)

// Scraper handles the main scraping logic
type Scraper struct {
	githubClient *client.GitHubClient
	filter       *Filter
	scorer       *Scorer
}

// Config represents scraper configuration
type Config struct {
	GitHubToken  string            `yaml:"github_token"`
	Repositories []RepositoryConfig `yaml:"repositories"`
	Filter       FilterConfig      `yaml:"filter"`
	Output       OutputConfig      `yaml:"output"`
}

// RepositoryConfig represents repository scraping configuration
type RepositoryConfig struct {
	Name      string   `yaml:"name"`
	Enabled   bool     `yaml:"enabled"`
	Keywords  []string `yaml:"keywords"`
	MinScore  float64  `yaml:"min_score"`
	MaxIssues int      `yaml:"max_issues"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Format     string `yaml:"format"`
	OutputDir  string `yaml:"output_dir"`
	SortBy     string `yaml:"sort_by"`
	IncludeRaw bool   `yaml:"include_raw"`
}

// NewScraper creates a new scraper instance
func NewScraper(config Config) *Scraper {
	scraper := &Scraper{
		githubClient: client.NewGitHubClient(config.GitHubToken),
		filter:       NewFilter(config.Filter),
		scorer:       NewScorer(),
	}
	
	return scraper
}

// ScrapeRepositories scrapes issues from configured repositories
func (s *Scraper) ScrapeRepositories(ctx context.Context, config Config) (map[string][]model.Issue, error) {
	allIssues := make(map[string][]model.Issue)
	
	log.Printf("Starting to scrape %d repositories...", len(config.Repositories))
	
	for i, repoConfig := range config.Repositories {
		if !repoConfig.Enabled {
			log.Printf("Skipping disabled repository: %s", repoConfig.Name)
			continue
		}
		
		log.Printf("Scraping repository %d/%d: %s", i+1, len(config.Repositories), repoConfig.Name)
		
		issues, err := s.scrapeRepository(ctx, repoConfig)
		if err != nil {
			log.Printf("Error scraping %s: %v", repoConfig.Name, err)
			continue
		}
		
		allIssues[repoConfig.Name] = issues
		log.Printf("Successfully scraped %d issues from %s", len(issues), repoConfig.Name)
		
		// Rate limiting between repositories
		if i < len(config.Repositories)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	return allIssues, nil
}

// scrapeRepository scrapes issues from a single repository
func (s *Scraper) scrapeRepository(ctx context.Context, repoConfig RepositoryConfig) ([]model.Issue, error) {
	// Parse repository name (format: owner/repo)
	parts := parseRepoName(repoConfig.Name)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name format: %s (expected owner/repo)", repoConfig.Name)
	}
	
	owner, repo := parts[0], parts[1]
	
	// Fetch issues
	githubIssues, err := s.githubClient.GetIssues(ctx, owner, repo, "all", repoConfig.MaxIssues)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	
	var issues []model.Issue
	
	for _, ghIssue := range githubIssues {
		issue := s.convertGitHubIssue(ghIssue, repoConfig.Name)
		issues = append(issues, issue)
		
		// Rate limiting between issues
		time.Sleep(50 * time.Millisecond)
	}
	
	return issues, nil
}

// convertGitHubIssue converts GitHub API issue to our model
func (s *Scraper) convertGitHubIssue(ghIssue *github.Issue, repoName string) model.Issue {
	// Extract labels
	var labels []model.Label
	if ghIssue.Labels != nil {
		for _, label := range ghIssue.Labels {
			name := ""
			if label.Name != nil {
				name = *label.Name
			}
			description := ""
			if label.Description != nil {
				description = *label.Description
			}
			color := ""
			if label.Color != nil {
				color = *label.Color
			}
			labels = append(labels, model.Label{
				Name:        name,
				Description: description,
				Color:       color,
			})
		}
	}
	
	// Count comments and reactions (these would be fetched separately in a real implementation)
	// For now, we'll use placeholder values
	comments := 0
	if ghIssue.Comments != nil {
		comments = *ghIssue.Comments
	}
	reactions := 0 // Would be fetched via separate API call
	
	// Safely extract string values
	title := ""
	if ghIssue.Title != nil {
		title = *ghIssue.Title
	}
	
	body := ""
	if ghIssue.Body != nil {
		body = *ghIssue.Body
	}
	
	url := ""
	if ghIssue.URL != nil {
		url = *ghIssue.URL
	}
	
	state := ""
	if ghIssue.State != nil {
		state = *ghIssue.State
	}
	
	createdAt := time.Time{}
	if ghIssue.CreatedAt != nil {
		createdAt = ghIssue.CreatedAt.Time
	}
	
	updatedAt := time.Time{}
	if ghIssue.UpdatedAt != nil {
		updatedAt = ghIssue.UpdatedAt.Time
	}
	
	return model.Issue{
		ID:          int(ghIssue.GetID()),
		Number:      ghIssue.GetNumber(),
		Title:       title,
		Body:        body,
		URL:         url,
		State:       state,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Labels:      labels,
		Comments:    comments,
		Reactions:   reactions,
		Repository:  repoName,
		Score:       0, // Will be calculated later
		ScoreReason: []string{},
	}
}

// FilterAndScoreIssues filters and scores all collected issues
func (s *Scraper) FilterAndScoreIssues(allIssues map[string][]model.Issue, config Config) map[string][]model.Issue {
	filteredIssues := make(map[string][]model.Issue)
	
	for repoName, issues := range allIssues {
		// Apply filtering and scoring
		filtered := s.filter.FilterIssues(issues, s.scorer)
		filteredIssues[repoName] = filtered
		
		log.Printf("Repository %s: %d issues filtered from %d total", 
			repoName, len(filtered), len(issues))
	}
	
	return filteredIssues
}

// GetStatistics returns scraping statistics
func (s *Scraper) GetStatistics(allIssues, filteredIssues map[string][]model.Issue) map[string]interface{} {
	stats := make(map[string]interface{})
	
	// Overall statistics
	totalIssues := 0
	filteredTotal := 0
	repoStats := make(map[string]interface{})
	
	for repoName, issues := range allIssues {
		totalIssues += len(issues)
		repoStats[repoName] = map[string]interface{}{
			"total_issues":     len(issues),
			"filtered_issues":  len(filteredIssues[repoName]),
			"filter_rate":      float64(len(filteredIssues[repoName])) / float64(len(issues)) * 100,
		}
		filteredTotal += len(filteredIssues[repoName])
	}
	
	stats["total_repositories"] = len(allIssues)
	stats["total_issues"] = totalIssues
	stats["filtered_issues"] = filteredTotal
	stats["overall_filter_rate"] = float64(filteredTotal) / float64(totalIssues) * 100
	stats["repository_stats"] = repoStats
	
	return stats
}

// Helper function to parse repository name
func parseRepoName(repoName string) []string {
	// Simple split by "/"
	var parts []string
	current := ""
	
	for _, char := range repoName {
		if char == '/' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}