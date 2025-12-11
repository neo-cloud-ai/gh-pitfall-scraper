package scraper

import (
	"testing"
	"time"
)

// TestPitfallScorer tests the scoring functionality
func TestPitfallScorer(t *testing.T) {
	scorer := NewPitfallScorer()
	
	// Test case 1: High-value performance regression issue
	highValueIssue := Issue{
		ID:        1,
		Number:    123,
		Title:     "Performance regression in vLLM inference",
		Body:      "The inference latency has increased significantly after the latest update. CUDA kernel performance is degraded.",
		State:     "open",
		CreatedAt: time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
		UpdatedAt: time.Now().Add(-1 * 24 * time.Hour), // 1 day ago
		Comments:  15,
		Reactions: Reactions{TotalCount: 25},
		Labels: []Label{
			{Name: "bug", Color: "d73a4a"},
			{Name: "performance", Color: "fbca04"},
		},
		Assignee: User{Login: "developer1"},
		Milestone: Milestone{Title: "v1.0.0"},
	}
	
	keywords := []string{"performance", "regression", "CUDA", "latency"}
	score := scorer.Score(&highValueIssue, keywords)
	
	t.Logf("High-value issue score: %.2f", score)
	if score < 20.0 {
		t.Errorf("Expected high-value issue to score above 20.0, got %.2f", score)
	}
	
	// Test case 2: Low-value documentation issue
	lowValueIssue := Issue{
		ID:        2,
		Number:    124,
		Title:     "Add documentation for new feature",
		Body:      "We need to add documentation for the new feature.",
		State:     "open",
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
		UpdatedAt: time.Now().Add(-25 * 24 * time.Hour), // 25 days ago
		Comments:  0,
		Reactions: Reactions{TotalCount: 1},
		Labels: []Label{
			{Name: "documentation", Color: "0075ca"},
		},
	}
	
	score = scorer.Score(&lowValueIssue, keywords)
	
	t.Logf("Low-value issue score: %.2f", score)
	if score > 10.0 {
		t.Errorf("Expected low-value issue to score below 10.0, got %.2f", score)
	}
}

// TestIssueFilter tests the filtering functionality
func TestIssueFilter(t *testing.T) {
	scorer := NewPitfallScorer()
	filter := NewIssueFilter(FilterOptions{
		MinScore:      10.0,
		MinComments:   1,
		MinAgeDays:    0,
		ExcludeLabels: []string{"documentation"},
		IncludeLabels: []string{"bug"},
		RequireLabel:  false,
	})
	
	// Create test issues
	issues := []Issue{
		{
			ID:        1,
			Number:    123,
			Title:     "Performance regression in GPU memory usage",
			Body:      "The GPU memory usage has increased significantly. OOM errors occur during training.",
			State:     "open",
			CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
			Comments:  10,
			Reactions: Reactions{TotalCount: 20},
			Labels: []Label{
				{Name: "bug", Color: "d73a4a"},
				{Name: "performance", Color: "fbca04"},
			},
		},
		{
			ID:        2,
			Number:    124,
			Title:     "Add documentation for new feature",
			Body:      "We need better documentation.",
			State:     "open",
			CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
			Comments:  2,
			Reactions: Reactions{TotalCount: 3},
			Labels: []Label{
				{Name: "documentation", Color: "0075ca"},
			},
		},
		{
			ID:        3,
			Number:    125,
			Title:     "CUDA kernel crash in distributed training",
			Body:      "The NCCL communication hangs during multi-node training.",
			State:     "open",
			CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
			Comments:  5,
			Reactions: Reactions{TotalCount: 15},
			Labels: []Label{
				{Name: "bug", Color: "d73a4a"},
			},
		},
	}
	
	keywords := []string{"performance", "CUDA", "OOM", "memory", "hang"}
	
	filtered := filter.FilterIssues(issues, scorer, keywords)
	
	t.Logf("Filtered %d issues out of %d total", len(filtered), len(issues))
	
	// Should filter out the documentation issue
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered issues, got %d", len(filtered))
	}
	
	// Check that all filtered issues have required labels
	for _, issue := range filtered {
		hasBugLabel := false
		for _, label := range issue.Labels {
			if label.Name == "bug" {
				hasBugLabel = true
				break
			}
		}
		if !hasBugLabel {
			t.Errorf("Filtered issue %d missing required 'bug' label", issue.ID)
		}
	}
}

// TestGithubClient tests GitHub client functionality (mocked)
func TestGithubClient(t *testing.T) {
	// This is a basic test since we can't make real API calls
	client := NewGithubClient("test_token")
	
	if client == nil {
		t.Error("Failed to create GitHub client")
	}
	
	if client.token != "test_token" {
		t.Error("GitHub client token not set correctly")
	}
	
	if client.baseURL != "https://api.github.com" {
		t.Error("GitHub client base URL not set correctly")
	}
}

// TestScrapingStats tests the statistics functionality
func TestScrapingStats(t *testing.T) {
	issues := []PitfallIssue{
		{
			ID:         1,
			Number:     123,
			Title:      "High-value issue",
			Score:      25.0,
			Keywords:   []string{"performance", "CUDA"},
			RepoOwner:  "test",
			RepoName:   "repo1",
		},
		{
			ID:         2,
			Number:     124,
			Title:      "Medium-value issue",
			Score:      17.5,
			Keywords:   []string{"memory", "OOM"},
			RepoOwner:  "test",
			RepoName:   "repo2",
		},
		{
			ID:         3,
			Number:     125,
			Title:      "Low-value issue",
			Score:      12.0,
			Keywords:   []string{"hang"},
			RepoOwner:  "test",
			RepoName:   "repo1",
		},
	}
	
	stats := GetScrapingStats(issues)
	
	if stats["total_issues"] != 3 {
		t.Errorf("Expected 3 total issues, got %v", stats["total_issues"])
	}
	
	if stats["high_value_count"] != 1 {
		t.Errorf("Expected 1 high-value issue, got %v", stats["high_value_count"])
	}
	
	if stats["average_score"] != 18.17 {
		t.Errorf("Expected average score ~18.17, got %v", stats["average_score"])
	}
}

// TestAdvancedFiltering tests advanced filtering capabilities
func TestAdvancedFiltering(t *testing.T) {
	basicOpts := FilterOptions{
		MinScore:      5.0,
		MinComments:   0,
		MinAgeDays:    0,
		ExcludeLabels: []string{},
		IncludeLabels: []string{},
		RequireLabel:  false,
	}
	
	advancedOpts := AdvancedFilterOptions{
		MinReactions: 5,
		HasMilestone: false,
		HasAssignee:  false,
	}
	
	advancedFilter := NewAdvancedIssueFilter(basicOpts, advancedOpts)
	
	issues := []Issue{
		{
			ID:        1,
			Number:    123,
			Title:     "Issue with high reactions",
			Body:      "This issue has good community engagement.",
			State:     "open",
			CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
			Comments:  10,
			Reactions: Reactions{TotalCount: 10},
			Labels: []Label{
				{Name: "bug", Color: "d73a4a"},
			},
		},
		{
			ID:        2,
			Number:    124,
			Title:     "Issue with low reactions",
			Body:      "This issue has poor community engagement.",
			State:     "open",
			CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
			Comments:  2,
			Reactions: Reactions{TotalCount: 2},
			Labels: []Label{
				{Name: "bug", Color: "d73a4a"},
			},
		},
	}
	
	scorer := NewPitfallScorer()
	filtered := advancedFilter.FilterIssuesWithAdvancedLogic(issues, scorer, []string{"bug"})
	
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered issue (high reactions), got %d", len(filtered))
	}
	
	if filtered[0].Reactions != 10 {
		t.Errorf("Expected filtered issue to have 10 reactions, got %d", filtered[0].Reactions)
	}
}