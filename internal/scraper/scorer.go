package scraper

import "github.com/google/go-github/v55/github"

func scoreIssue(i *github.Issue, keywords []string) int {
	score := 0

	// 💥 keyword match
	if matchKeywords(i.GetTitle(), keywords) {
		score += 10
	}
	if matchKeywords(i.GetBody(), keywords) {
		score += 20
	}

	// 👍 reactions
	score += i.GetReactions().GetTotalCount()

	// 💬 comment count
	score += i.GetComments()

	// 🏷️ label bonus
	for _, l := range i.Labels {
		if l.GetName() == "bug" || l.GetName() == "performance" {
			score += 10
		}
	}

	return score
}
