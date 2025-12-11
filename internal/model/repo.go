package model

// Repository represents a GitHub repository to scrape
type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	
	// Scraping configuration
	Enabled    bool     `json:"enabled"`
	Keywords   []string `json:"keywords"`
	MinScore   float64  `json:"min_score"`
	MaxIssues  int      `json:"max_issues"`
	
	// Statistics
	IssuesScraped int     `json:"issues_scraped"`
	HighValueCount int    `json:"high_value_count"`
}