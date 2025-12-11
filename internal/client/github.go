package client

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"github.com/google/go-github/v67/github"
)

// GitHubClient wraps the GitHub API client
type GitHubClient struct {
	client *github.Client
	token  string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string) *GitHubClient {
	client := github.NewClient(nil)
	if token != "" {
		client = github.NewClient(nil).WithAuthToken(token)
	}
	
	return &GitHubClient{
		client: client,
		token:  token,
	}
}

// GetIssues retrieves issues from a repository
func (c *GitHubClient) GetIssues(ctx context.Context, owner, repo string, state string, maxIssues int) ([]*github.Issue, error) {
	var allIssues []*github.Issue
	page := 1
	perPage := 100
	
	for {
		issues, resp, err := c.client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
			State:       state,
			Sort:        "updated",
			Direction:   "desc",
			ListOptions: github.ListOptions{Page: page, PerPage: perPage},
		})
		
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}
		
		if len(issues) == 0 {
			break
		}
		
		allIssues = append(allIssues, issues...)
		
		if len(allIssues) >= maxIssues {
			break
		}
		
		if resp.NextPage == 0 {
			break
		}
		
		page = resp.NextPage
		
		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}
	
	log.Printf("Retrieved %d issues from %s/%s", len(allIssues), owner, repo)
	return allIssues, nil
}

// GetIssueComments retrieves comments for an issue
func (c *GitHubClient) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*github.IssueComment, error) {
	comments, _, err := c.client.Issues.ListComments(ctx, owner, repo, issueNumber, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments for issue %d: %w", issueNumber, err)
	}
	
	return comments, nil
}

// GetIssueReactions retrieves reactions for an issue
func (c *GitHubClient) GetIssueReactions(ctx context.Context, owner, repo string, issueNumber int) ([]*github.Reaction, error) {
	reactions, _, err := c.client.Reactions.ListIssueReactions(ctx, owner, repo, issueNumber, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions for issue %d: %w", issueNumber, err)
	}
	
	return reactions, nil
}

// GetRepoInfo retrieves repository information
func (c *GitHubClient) GetRepoInfo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repoInfo, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository info: %w", err)
	}
	
	return repoInfo, nil
}