package scraper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/database"
)

// PitfallIssue represents a high-value issue with pitfall information
type PitfallIssue struct {
	ID          int    `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	State       string `json:"state"`
	Labels      []Label `json:"labels"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Pitfall-specific information
	Keywords   []string `json:"keywords"`   // Matching keywords
	Score      float64  `json:"score"`      // Pitfall value score
	Comments   int      `json:"comments"`   // Number of comments
	Reactions  int      `json:"reactions"`  // Number of reactions
	Assignee   string   `json:"assignee"`   // Assigned person
	Milestone  string   `json:"milestone"`  // Milestone if any
	
	// Issue body for detailed analysis
	Body       string   `json:"body"`
	
	// Repository information
	RepoOwner  string   `json:"repo_owner"`
	RepoName   string   `json:"repo_name"`
}

// Label represents a GitHub issue label
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// GithubClient handles GitHub API interactions with database integration
type GithubClient struct {
	token          string
	baseURL        string
	client         *http.Client
	dbService      *DatabaseService
	retryConfig    RetryConfig
}

// DatabaseIntegration provides database integration functionality
type DatabaseIntegration struct {
	dedupService       *database.DeduplicationService
	classificationService *database.ClassificationService
	crudOperations     database.CRUDOperations
	logger             *log.Logger
}

// RetryConfig configuration for API retry logic
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		BackoffFactor: 2.0,
	}
}

// NewGithubClient creates a new GitHub API client
func NewGithubClient(token string) *GithubClient {
	return &GithubClient{
		token:     token,
		baseURL:   "https://api.github.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		retryConfig: DefaultRetryConfig(),
	}
}

// NewGithubClientWithDB creates a new GitHub API client with database integration
func NewGithubClientWithDB(token string, dbService *DatabaseService) *GithubClient {
	return &GithubClient{
		token:     token,
		baseURL:   "https://api.github.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		dbService:   dbService,
		retryConfig: DefaultRetryConfig(),
	}
}

// makeRequest performs an authenticated GitHub API request
func (c *GithubClient) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	
	// Set authentication header
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gh-pitfall-scraper")
	
	return c.client.Do(req)
}

// GetIssues retrieves issues for a repository
func (c *GithubClient) GetIssues(owner, repo string, page, perPage int) ([]Issue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues?state=open&per_page=%d&page=%d&sort=updated&direction=desc",
		owner, repo, perPage, page)
	
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}
	
	return issues, nil
}

// Issue represents a GitHub issue
type Issue struct {
	ID          int       `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Comments    int       `json:"comments"`
	Reactions   Reactions `json:"reactions"`
	Labels      []Label   `json:"labels"`
	Assignee    User      `json:"assignee"`
	Milestone   Milestone `json:"milestone"`
	URL         string    `json:"html_url"`
}

// Reactions represents GitHub reactions
type Reactions struct {
	TotalCount int `json:"total_count"`
}

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

// Milestone represents a GitHub milestone
type Milestone struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueOn       time.Time `json:"due_on"`
}

// GetIssuesFromDatabase retrieves issues from database with filtering
func (c *GithubClient) GetIssuesFromDatabase(filters DatabaseFilter, limit, offset int) ([]*database.Issue, error) {
	if c.dbService == nil {
		return nil, fmt.Errorf("database service not initialized")
	}
	
	search := &database.AdvancedSearch{
		Query:       filters.Query,
		Keywords:    filters.Keywords,
		Categories:  filters.Categories,
		Priorities:  filters.Priorities,
		TechStacks:  filters.TechStacks,
		Repos:       filters.Repos,
		States:      filters.States,
		MinScore:    filters.MinScore,
		MaxScore:    filters.MaxScore,
		DateFrom:    filters.DateFrom,
		DateTo:      filters.DateTo,
		SortBy:      "score",
		SortOrder:   "DESC",
		Limit:       limit,
		Offset:      offset,
		ExcludeDuplicates: true,
	}
	
	return c.dbService.crudOps.SearchIssuesAdvanced(search)
}

// GetRecentIssuesFromDatabase retrieves recent issues from database
func (c *GithubClient) GetRecentIssuesFromDatabase(limit int) ([]*database.Issue, error) {
	return c.dbService.crudOps.GetAllIssues(limit, 0)
}

// GetIssuesByRepositoryFromDatabase retrieves issues by repository from database
func (c *GithubClient) GetIssuesByRepositoryFromDatabase(owner, name string, limit, offset int) ([]*database.Issue, error) {
	return c.dbService.crudOps.GetIssuesByRepository(owner, name, limit, offset)
}

// GetHighScoreIssuesFromDatabase retrieves high-score issues from database
func (c *GithubClient) GetHighScoreIssuesFromDatabase(minScore float64, limit int) ([]*database.Issue, error) {
	return c.dbService.crudOps.GetIssuesByScore(minScore, 100.0, limit, 0)
}

// GetDuplicateStatsFromDatabase retrieves duplicate statistics from database
func (c *GithubClient) GetDuplicateStatsFromDatabase() (map[string]interface{}, error) {
	return c.dbService.dedupService.GetDuplicateStats()
}

// GetClassificationStatsFromDatabase retrieves classification statistics from database
func (c *GithubClient) GetClassificationStatsFromDatabase() (map[string]interface{}, error) {
	return c.dbService.classificationService.GetClassificationStats()
}

// RunDeduplication runs duplicate detection on database issues
func (c *GithubClient) RunDeduplication() (*database.DeduplicationResult, error) {
	return c.dbService.dedupService.FindDuplicates()
}

// RunClassification runs classification on database issues
func (c *GithubClient) RunClassification(issues []*database.Issue) (*database.ClassificationResult, error) {
	return c.dbService.classificationService.ClassifyIssues(issues)
}

// DatabaseFilter represents filters for database queries
type DatabaseFilter struct {
	Query       string    `json:"query"`
	Keywords    []string  `json:"keywords"`
	Categories  []string  `json:"categories"`
	Priorities  []string  `json:"priorities"`
	TechStacks  []string  `json:"tech_stacks"`
	Repos       []string  `json:"repos"`
	States      []string  `json:"states"`
	MinScore    *float64  `json:"min_score"`
	MaxScore    *float64  `json:"max_score"`
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
}

// DefaultDatabaseFilter returns default database filter
func DefaultDatabaseFilter() DatabaseFilter {
	return DatabaseFilter{
		States: []string{"open"},
	}
}