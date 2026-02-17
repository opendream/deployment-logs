package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type PRInfo struct {
	Number      int
	Title       string
	Description string
	MergedAt    time.Time
	CommitSHA   string
	Category    string
}

type GitHubService struct{}

type githubPR struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	MergedAt *string `json:"merged_at"`
	MergeCommitSHA string `json:"merge_commit_sha"`
}

func ParseGitHubOwnerRepo(repoURL string) (string, string, error) {
	url := strings.TrimSuffix(repoURL, ".git")
	url = strings.TrimSuffix(url, "/")

	re := regexp.MustCompile(`github\.com[/:]([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("cannot parse owner/repo from %s", repoURL)
	}
	return matches[1], matches[2], nil
}

func (s *GitHubService) FetchMergedPRs(owner, repo, branch, authToken string, since time.Time) ([]PRInfo, error) {
	var allPRs []PRInfo
	page := 1

	for {
		url := fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls?state=closed&base=%s&sort=updated&direction=desc&per_page=100&page=%d",
			owner, repo, branch, page,
		)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		if authToken != "" {
			req.Header.Set("Authorization", "Bearer "+authToken)
		}
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("github request failed: %w", err)
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
		}

		var prs []githubPR
		err = json.NewDecoder(resp.Body).Decode(&prs)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}

		if len(prs) == 0 {
			break
		}

		reachedOld := false
		for _, pr := range prs {
			if pr.MergedAt == nil {
				continue
			}
			mergedAt, err := time.Parse(time.RFC3339, *pr.MergedAt)
			if err != nil {
				continue
			}
			if mergedAt.Before(since) {
				reachedOld = true
				break
			}
			allPRs = append(allPRs, PRInfo{
				Number:    pr.Number,
				Title:     pr.Title,
				MergedAt:  mergedAt,
				CommitSHA: pr.MergeCommitSHA,
				Category:  ParseCategory(pr.Title),
			})
		}

		if reachedOld || len(prs) < 100 {
			break
		}
		page++
	}

	return allPRs, nil
}

var categoryRe = regexp.MustCompile(`^(feat|fix|hotfix|chore|refactor)(\(.*?\))?:\s*`)

func ParseCategory(title string) string {
	matches := categoryRe.FindStringSubmatch(strings.ToLower(title))
	if len(matches) >= 2 {
		return matches[1]
	}
	return "other"
}
