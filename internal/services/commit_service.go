package services

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitService struct{}

var conventionalCommitRe = regexp.MustCompile(`^(feat|fix|hotfix|chore|refactor)(\(.*?\))?:`)

// Matches "Merge pull request #123 from ..." (GitHub) or "See merge request !123" (GitLab)
var mergePRRe = regexp.MustCompile(`Merge pull request #(\d+)`)
var mergeMRRe = regexp.MustCompile(`See merge request .*!(\d+)`)

func (s *CommitService) ExtractCommits(repo *git.Repository, branch string, since time.Time) ([]PRInfo, error) {
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, err
	}

	iter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Since: &since,
	})
	if err != nil {
		return nil, err
	}

	var items []PRInfo
	err = iter.ForEach(func(c *object.Commit) error {
		msg := c.Message
		title := strings.Split(msg, "\n")[0]

		// For merge commits, extract title and description from the commit body
		if c.NumParents() > 1 {
			lines := strings.SplitN(msg, "\n", 2)
			title = strings.TrimSpace(lines[0])
			description := ""
			if len(lines) > 1 {
				description = strings.TrimSpace(lines[1])
			}

			if !conventionalCommitRe.MatchString(strings.ToLower(title)) {
				return nil
			}
			items = append(items, PRInfo{
				Title:       title,
				Description: description,
				MergedAt:    c.Author.When,
				CommitSHA:   c.Hash.String(),
				Category:    ParseCategory(title),
			})
			return nil
		}

		// Regular commits
		if !conventionalCommitRe.MatchString(strings.ToLower(title)) {
			return nil
		}
		items = append(items, PRInfo{
			Title:     title,
			MergedAt:  c.Author.When,
			CommitSHA: c.Hash.String(),
			Category:  ParseCategory(title),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

// ExtractMergeCommits reads merge commits from git log to extract PR/MR info.
// Used in PR mode when no API token is available.
// Parses merge commit messages like:
//   - GitHub: "Merge pull request #123 from branch\n\nfeat: add feature\n\nbody text"
//   - GitLab: "Merge branch 'feat/x' into main\n\nfeat: add feature\n\nSee merge request group/project!123"
//   - Squash merges: "feat(scope): title (#123)\n\nbody text"
func (s *CommitService) ExtractMergeCommits(repo *git.Repository, branch string, since time.Time) ([]PRInfo, error) {
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, err
	}

	iter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Since: &since,
	})
	if err != nil {
		return nil, err
	}

	var items []PRInfo
	err = iter.ForEach(func(c *object.Commit) error {
		msg := strings.TrimSpace(c.Message)
		lines := strings.Split(msg, "\n")
		firstLine := strings.TrimSpace(lines[0])

		isMergeCommit := c.NumParents() > 1

		// Try to extract PR info from merge commits or squash-merge commits
		info := parsePRFromCommit(firstLine, lines, isMergeCommit)
		if info == nil {
			return nil
		}

		info.MergedAt = c.Author.When
		info.CommitSHA = c.Hash.String()

		// Only include items that have a PR/MR number
		if info.Number == 0 {
			return nil
		}

		items = append(items, *info)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

// squashPRRe matches squash-merge titles like "feat(scope): title (#123)"
var squashPRRe = regexp.MustCompile(`\(#(\d+)\)\s*$`)

func parsePRFromCommit(firstLine string, allLines []string, isMergeCommit bool) *PRInfo {
	// Collect the body (everything after the first line, trimmed)
	body := ""
	if len(allLines) > 1 {
		body = strings.TrimSpace(strings.Join(allLines[1:], "\n"))
	}

	// Case 1: GitHub merge commit "Merge pull request #123 from branch/name"
	// The actual PR title is in the body (next non-empty line)
	if matches := mergePRRe.FindStringSubmatch(firstLine); len(matches) >= 2 {
		prNum, _ := strconv.Atoi(matches[1])
		title, description := extractTitleAndDescription(body)
		if title == "" {
			return nil
		}
		if !conventionalCommitRe.MatchString(strings.ToLower(title)) {
			return nil
		}
		return &PRInfo{
			Number:      prNum,
			Title:       title,
			Description: description,
			Category:    ParseCategory(title),
		}
	}

	// Case 2: GitLab merge commit — check body for "See merge request ...!123"
	if isMergeCommit {
		if matches := mergeMRRe.FindStringSubmatch(body); len(matches) >= 2 {
			prNum, _ := strconv.Atoi(matches[1])
			// Remove the "See merge request" line from description
			cleanBody := mergeMRRe.ReplaceAllString(body, "")
			title, description := extractTitleAndDescription(cleanBody)
			if title == "" {
				// Fallback: use the merge commit first line if body has no conventional title
				title = firstLine
				description = strings.TrimSpace(cleanBody)
			}
			if !conventionalCommitRe.MatchString(strings.ToLower(title)) {
				return nil
			}
			return &PRInfo{
				Number:      prNum,
				Title:       title,
				Description: description,
				Category:    ParseCategory(title),
			}
		}

		// Case 3: Generic merge commit with conventional title in body
		// e.g. "Merge branch 'feat/x' into develop\n\nfeat: some title\n\nbody"
		title, description := extractTitleAndDescription(body)
		if title != "" && conventionalCommitRe.MatchString(strings.ToLower(title)) {
			return &PRInfo{
				Title:       title,
				Description: description,
				Category:    ParseCategory(title),
			}
		}

		// Case 4: Merge commit where first line itself is conventional
		if conventionalCommitRe.MatchString(strings.ToLower(firstLine)) {
			return &PRInfo{
				Title:       firstLine,
				Description: body,
				Category:    ParseCategory(firstLine),
			}
		}
	}

	// Case 5: Squash merge with PR number "feat: title (#123)"
	if matches := squashPRRe.FindStringSubmatch(firstLine); len(matches) >= 2 {
		if conventionalCommitRe.MatchString(strings.ToLower(firstLine)) {
			prNum, _ := strconv.Atoi(matches[1])
			// Clean title: remove the (#123) suffix
			title := strings.TrimSpace(squashPRRe.ReplaceAllString(firstLine, ""))
			return &PRInfo{
				Number:      prNum,
				Title:       title,
				Description: body,
				Category:    ParseCategory(title),
			}
		}
	}

	return nil
}

// extractTitleAndDescription splits body into first non-empty line (title) and the rest (description)
func extractTitleAndDescription(body string) (string, string) {
	lines := strings.Split(body, "\n")
	title := ""
	descStart := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			title = trimmed
			descStart = i + 1
			break
		}
	}
	if title == "" {
		return "", ""
	}
	description := ""
	if descStart < len(lines) {
		description = strings.TrimSpace(strings.Join(lines[descStart:], "\n"))
	}
	return title, description
}
