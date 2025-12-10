package scraper

import (
	"context"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	client *github.Client
}

func NewGithubClient(token string) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return &GithubClient{client: github.NewClient(tc)}
}

func (g *GithubClient) ListIssues(owner, repo string) ([]*github.Issue, error) {
	opts := &github.IssueListByRepoOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	var all []*github.Issue

	for {
		issues, resp, err := g.client.Issues.ListByRepo(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, issues...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return all, nil
}
