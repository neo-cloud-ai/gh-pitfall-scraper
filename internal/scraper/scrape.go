package scraper

import (
	"fmt"
)

type PitfallIssue struct {
	Repo     string `json:"repo"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Score    int    `json:"score"`
	Comments int    `json:"comments"`
}

func ScrapeRepo(
	client *GithubClient,
	owner, repo string,
	keywords []string,
) ([]PitfallIssue, error) {

	issues, err := client.ListIssues(owner, repo)
	if err != nil {
		return nil, err
	}

	results := []PitfallIssue{}

	for _, i := range issues {
		score := scoreIssue(i, keywords)
		if score >= 15 { // score threshold，可调
			results = append(results, PitfallIssue{
				Repo:     fmt.Sprintf("%s/%s", owner, repo),
				Title:    i.GetTitle(),
				URL:      i.GetHTMLURL(),
				Score:    score,
				Comments: i.GetComments(),
			})
		}
	}

	return results, nil
}
