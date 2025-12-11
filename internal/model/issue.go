package model

import "time"

// Issue represents a GitHub issue with scoring information
type Issue struct {
	ID          int       `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	URL         string    `json:"url"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Labels      []Label   `json:"labels"`
	Comments    int       `json:"comments"`
	Reactions   int       `json:"reactions"`
	
	// Scoring information
	Score       float64   `json:"score"`
	ScoreReason []string  `json:"score_reason"`
	
	// Repository information
	Repository  string    `json:"repository"`
}

// Label represents a GitHub label
type Label struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}