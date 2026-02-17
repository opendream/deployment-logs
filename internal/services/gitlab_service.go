package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type GitLabService struct{}

type gitlabMR struct {
	IID            int    `json:"iid"`
	Title          string `json:"title"`
	MergedAt       *string `json:"merged_at"`
	MergeCommitSHA *string `json:"merge_commit_sha"`
	SHA            string `json:"sha"`
}

func ParseGitLabProjectPath(repoURL string) (string, string, error) {
	u := strings.TrimSuffix(repoURL, ".git")
	u = strings.TrimSuffix(u, "/")

	re := regexp.MustCompile(`https?://([^/]+)/(.+)`)
	matches := re.FindStringSubmatch(u)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("cannot parse project path from %s", repoURL)
	}
	baseURL := "https://" + matches[1]
	projectPath := matches[2]
	return projectPath, baseURL, nil
}

func (s *GitLabService) FetchMergedMRs(projectPath, branch, authToken, baseURL string, since time.Time) ([]PRInfo, error) {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	var allPRs []PRInfo
	page := 1

	encodedPath := url.PathEscape(projectPath)

	for {
		apiURL := fmt.Sprintf(
			"%s/api/v4/projects/%s/merge_requests?state=merged&target_branch=%s&order_by=updated_at&sort=desc&per_page=100&page=%d",
			baseURL, encodedPath, url.QueryEscape(branch), page,
		)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
		if authToken != "" {
			req.Header.Set("PRIVATE-TOKEN", authToken)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("gitlab request failed: %w", err)
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("gitlab API returned status %d", resp.StatusCode)
		}

		var mrs []gitlabMR
		err = json.NewDecoder(resp.Body).Decode(&mrs)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}

		if len(mrs) == 0 {
			break
		}

		reachedOld := false
		for _, mr := range mrs {
			if mr.MergedAt == nil {
				continue
			}
			mergedAt, err := time.Parse(time.RFC3339, *mr.MergedAt)
			if err != nil {
				// Try alternate format
				mergedAt, err = time.Parse("2006-01-02T15:04:05.000Z", *mr.MergedAt)
				if err != nil {
					continue
				}
			}
			if mergedAt.Before(since) {
				reachedOld = true
				break
			}
			commitSHA := mr.SHA
			if mr.MergeCommitSHA != nil {
				commitSHA = *mr.MergeCommitSHA
			}
			allPRs = append(allPRs, PRInfo{
				Number:    mr.IID,
				Title:     mr.Title,
				MergedAt:  mergedAt,
				CommitSHA: commitSHA,
				Category:  ParseCategory(mr.Title),
			})
		}

		if reachedOld || len(mrs) < 100 {
			break
		}
		page++
	}

	return allPRs, nil
}
