package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/model"
	"github.com/neo-cloud-ai/gh-pitfall-scraper/internal/scraper"
)

func TestScorer(t *testing.T) {
	scorer := scraper.NewScorer()
	
	// Test case 1: High score issue
	highScoreIssue := model.Issue{
		ID:        1,
		Title:     "Performance regression in GPU memory usage after v0.4.0",
		Body:      "After upgrading to v0.4.0, we're seeing significant memory usage increase. GPU OOM issues with large batch sizes.",
		State:     "open",
		Comments:  15,
		Reactions: 8,
		Labels: []model.Label{
			{Name: "bug", Color: "d73a4a"},
			{Name: "performance", Color: "fbca04"},
		},
		CreatedAt: time.Now().AddDate(0, 0, -10),
		UpdatedAt: time.Now().AddDate(0, 0, -2),
	}
	
	score, reasons := scorer.ScoreIssue(&highScoreIssue)
	
	fmt.Printf("高价值问题测试:\n")
	fmt.Printf("评分: %.1f\n", score)
	fmt.Printf("评分理由:\n")
	for _, reason := range reasons {
		fmt.Printf("  - %s\n", reason)
	}
	
	if score < 60 {
		t.Errorf("Expected high score for performance regression issue, got %.1f", score)
	}
	
	// Test case 2: Low score issue
	lowScoreIssue := model.Issue{
		ID:        2,
		Title:     "Feature request: add new API endpoint",
		Body:      "Please add a new API endpoint for better integration.",
		State:     "open",
		Comments:  2,
		Reactions: 1,
		Labels: []model.Label{
			{Name: "enhancement", Color: "a2eeef"},
		},
		CreatedAt: time.Now().AddDate(0, 0, -10),
		UpdatedAt: time.Now().AddDate(0, 0, -2),
	}
	
	score, reasons = scorer.ScoreIssue(&lowScoreIssue)
	
	fmt.Printf("\n低价值问题测试:\n")
	fmt.Printf("评分: %.1f\n", score)
	fmt.Printf("评分理由:\n")
	for _, reason := range reasons {
		fmt.Printf("  - %s\n", reason)
	}
	
	if score > 30 {
		t.Errorf("Expected low score for feature request, got %.1f", score)
	}
}

func TestFilter(t *testing.T) {
	filterConfig := scraper.FilterConfig{
		MinScore:      20.0,
		RequiredState: "open",
		MaxIssues:     10,
	}
	
	filter := scraper.NewFilter(filterConfig)
	scorer := scraper.NewScorer()
	
	issues := []model.Issue{
		{
			ID:        1,
			Title:     "Performance regression in GPU memory usage",
			Body:      "GPU OOM issues with large batch sizes",
			State:     "open",
			CreatedAt: time.Now().AddDate(0, 0, -10),
			UpdatedAt: time.Now().AddDate(0, 0, -2),
		},
		{
			ID:        2,
			Title:     "Feature request: add new API endpoint",
			Body:      "Please add a new API endpoint",
			State:     "open",
			CreatedAt: time.Now().AddDate(0, 0, -10),
			UpdatedAt: time.Now().AddDate(0, 0, -2),
		},
		{
			ID:        3,
			Title:     "CUDA kernel crash",
			Body:      "CUDA kernel crash when using flash attention",
			State:     "closed",
			CreatedAt: time.Now().AddDate(0, 0, -10),
			UpdatedAt: time.Now().AddDate(0, 0, -2),
		},
	}
	
	filtered := filter.FilterIssues(issues, scorer)
	
	fmt.Printf("过滤测试:\n")
	fmt.Printf("原始问题数: %d\n", len(issues))
	fmt.Printf("过滤后问题数: %d\n", len(filtered))
	
	// Should only include open issues
	for _, issue := range filtered {
		if issue.State != "open" {
			t.Errorf("Filtered issue should be open, got %s", issue.State)
		}
	}
}

func TestCategorization(t *testing.T) {
	filterConfig := scraper.FilterConfig{
		MinScore:      0.0,
		RequiredState: "all",
		MaxIssues:     100,
	}
	
	filter := scraper.NewFilter(filterConfig)
	scorer := scraper.NewScorer()
	
	issues := []model.Issue{
		{
			ID:        1,
			Title:     "Performance regression in GPU memory usage",
			Body:      "GPU OOM issues with large batch sizes",
			State:     "open",
			CreatedAt: time.Now().AddDate(0, 0, -10),
			UpdatedAt: time.Now().AddDate(0, 0, -2),
		},
		{
			ID:        2,
			Title:     "Distributed training hang",
			Body:      "NCCL deadlock during multi-node training",
			State:     "open",
			CreatedAt: time.Now().AddDate(0, 0, -10),
			UpdatedAt: time.Now().AddDate(0, 0, -2),
		},
		{
			ID:        3,
			Title:     "Memory leak in inference",
			Body:      "Memory usage keeps increasing during inference",
			State:     "open",
			CreatedAt: time.Now().AddDate(0, 0, -10),
			UpdatedAt: time.Now().AddDate(0, 0, -2),
		},
	}
	
	// Calculate scores first
	for i := range issues {
		score, _ := scorer.ScoreIssue(&issues[i])
		issues[i].Score = score
	}
	
	categories := filter.CategorizeIssues(issues)
	
	fmt.Printf("分类测试:\n")
	for category, catIssues := range categories {
		fmt.Printf("  %s: %d 个问题\n", category, len(catIssues))
	}
	
	// Verify categorization
	if len(categories["gpu_memory"]) == 0 {
		t.Error("Expected GPU memory issues to be categorized")
	}
	
	if len(categories["distributed"]) == 0 {
		t.Error("Expected distributed training issues to be categorized")
	}
}