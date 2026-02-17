package services

import (
	"fmt"
	"time"

	"deployment-logs/internal/config"
	"deployment-logs/internal/models"

	"gorm.io/gorm"
)

type LogGenerator struct {
	Git    *GitService
	GitHub *GitHubService
	GitLab *GitLabService
	Commit *CommitService
}

func NewLogGenerator() *LogGenerator {
	return &LogGenerator{
		Git:    &GitService{},
		GitHub: &GitHubService{},
		GitLab: &GitLabService{},
		Commit: &CommitService{},
	}
}

func Generate(db *gorm.DB, cfg *config.Config, repoName string, branch string) (*models.DeploymentLog, error) {
	return NewLogGenerator().Generate(db, cfg, repoName, branch)
}

func (g *LogGenerator) Generate(db *gorm.DB, cfg *config.Config, repoName string, branch string) (*models.DeploymentLog, error) {
	// 1. Find RepositoryConfig
	var repoCfg models.RepositoryConfig
	if err := db.Where("name = ?", repoName).First(&repoCfg).Error; err != nil {
		return nil, fmt.Errorf("repository config not found: %w", err)
	}

	// 2. Match branch against patterns to find environment + date strategy
	branchConfigs := ParseBranchConfigs(repoCfg.Branches)
	matched := MatchBranch(branchConfigs, branch)

	environment := branch // default: environment = branch name
	if matched != nil {
		environment = matched.Environment
	}

	// 3. Clone or pull repo, checking out the target branch
	repo, err := g.Git.CloneOrPull(repoCfg.RepoURL, repoCfg.SSHKey, cfg.ReposDir, repoCfg.Name, branch)
	if err != nil {
		return nil, fmt.Errorf("git clone/pull failed: %w", err)
	}

	// 4. Determine "since" from last deployment log
	var lastLog models.DeploymentLog
	since := time.Now().AddDate(0, 0, -30)
	if err := db.Where("repo_config_id = ? AND environment = ?", repoCfg.ID, environment).
		Order("generated_at desc").First(&lastLog).Error; err == nil {
		since = lastLog.GeneratedAt
	}

	now := time.Now()

	// 5. Extract PR info from merge commits in git log for the target branch
	allItems, err := g.Commit.ExtractMergeCommits(repo, branch, since)
	if err != nil {
		return nil, fmt.Errorf("fetch failed for branch %s: %w", branch, err)
	}

	// 6. Create DeploymentLog + LogItems
	deploymentLog := models.DeploymentLog{
		RepoConfigID: repoCfg.ID,
		Environment:  environment,
		GeneratedAt:  now,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&deploymentLog).Error; err != nil {
			return err
		}

		for _, item := range allItems {
			prNum := item.Number
			logItem := models.LogItem{
				DeploymentLogID: deploymentLog.ID,
				Title:           item.Title,
				Description:     item.Description,
				CommitSHA:       item.CommitSHA,
				MergedAt:        item.MergedAt,
				Category:        item.Category,
			}
			if prNum > 0 {
				logItem.PRNumber = &prNum
			}
			if err := tx.Create(&logItem).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	// 7. Reload with items
	db.Preload("LogItems").First(&deploymentLog, deploymentLog.ID)
	return &deploymentLog, nil
}
