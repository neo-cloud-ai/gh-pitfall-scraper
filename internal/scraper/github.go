package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

// GithubClient handles GitHub API interactions
type GithubClient struct {
	token     string
	baseURL   string
	client    *http.Client
}

// NewGithubClient creates a new GitHub API client
func NewGithubClient(token string) *GithubClient {
	return &GithubClient{
		token:   token,
		baseURL: "https://api.github.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
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