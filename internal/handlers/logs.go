package handlers

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"deployment-logs/internal/models"
	"deployment-logs/internal/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func ListLogs(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var logs []models.DeploymentLog
		query := db.Model(&models.DeploymentLog{})

		if repo := c.Query("repo"); repo != "" {
			var cfg models.RepositoryConfig
			if err := db.Where("name = ?", repo).First(&cfg).Error; err == nil {
				query = query.Where("repo_config_id = ?", cfg.ID)
			} else {
				return c.JSON([]models.DeploymentLog{})
			}
		}

		if env := c.Query("env"); env != "" {
			query = query.Where("environment = ?", env)
		}

		if err := query.Order("created_at desc").Find(&logs).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(logs)
	}
}

func GetLog(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var log models.DeploymentLog
		if err := db.Preload("LogItems").First(&log, id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		return c.JSON(log)
	}
}

type dateGroup struct {
	Date  string           `json:"date"`
	Items []models.LogItem `json:"items"`
}

type paginatedResponse struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
	TotalItems int         `json:"total_items"`
	Data       []dateGroup `json:"data"`
}

// GetLogsByRepo returns log items for a repo+branch, grouped by date, paginated.
// GET /api/logs/:repo_name?branch=develop&page=1&per_page=20
func GetLogsByRepo(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		repoName := c.Params("repo_name")
		branch := c.Query("branch", "")
		page, _ := strconv.Atoi(c.Query("page", "1"))
		perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}

		// Find repo config
		var repoCfg models.RepositoryConfig
		if err := db.Where("name = ?", repoName).First(&repoCfg).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "repository not found"})
		}

		// Parse branch configs to find date strategy for this environment
		branchConfigs := services.ParseBranchConfigs(repoCfg.Branches)
		useDeployDate := false
		for _, bc := range branchConfigs {
			if bc.Environment == branch {
				useDeployDate = bc.DateStrategy == "deploy_date"
				break
			}
		}

		// Find deployment logs matching repo + branch (environment)
		logQuery := db.Where("repo_config_id = ?", repoCfg.ID)
		if branch != "" {
			logQuery = logQuery.Where("environment = ?", branch)
		}

		var deploymentLogs []models.DeploymentLog
		if err := logQuery.Order("generated_at desc").Find(&deploymentLogs).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if len(deploymentLogs) == 0 {
			return c.JSON(paginatedResponse{
				Page:       page,
				PerPage:    perPage,
				TotalPages: 0,
				TotalItems: 0,
				Data:       []dateGroup{},
			})
		}

		// Build a map from deployment log ID → generated_at date (for deploy_date grouping)
		logDateMap := make(map[uint]string)
		for _, dl := range deploymentLogs {
			logDateMap[dl.ID] = dl.GeneratedAt.Format(time.DateOnly)
		}

		// Collect all log IDs
		logIDs := make([]uint, len(deploymentLogs))
		for i, dl := range deploymentLogs {
			logIDs[i] = dl.ID
		}

		// Fetch all items for these logs
		var allItems []models.LogItem
		if err := db.Where("deployment_log_id IN ?", logIDs).
			Order("merged_at desc").Find(&allItems).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Deduplicate by commit_sha (multiple Generate runs create duplicates)
		seen := make(map[string]bool)
		uniqueItems := make([]models.LogItem, 0, len(allItems))
		for _, item := range allItems {
			if !seen[item.CommitSHA] {
				seen[item.CommitSHA] = true
				uniqueItems = append(uniqueItems, item)
			}
		}
		allItems = uniqueItems

		// Group items by date
		// deploy_date strategy: group by deployment log's generated_at (deploy trigger date)
		// otherwise: group by item's merged_at (actual commit date)
		dateMap := make(map[string][]models.LogItem)
		for _, item := range allItems {
			var key string
			if useDeployDate {
				key = logDateMap[item.DeploymentLogID]
			} else {
				key = item.MergedAt.Format(time.DateOnly)
			}
			dateMap[key] = append(dateMap[key], item)
		}

		// Sort items within each date group by merged_at descending (newest commit first)
		for _, items := range dateMap {
			sort.Slice(items, func(i, j int) bool {
				return items[i].MergedAt.After(items[j].MergedAt)
			})
		}

		// Sort dates descending
		dates := make([]string, 0, len(dateMap))
		for d := range dateMap {
			dates = append(dates, d)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(dates)))

		totalItems := len(allItems)
		totalPages := int(math.Ceil(float64(len(dates)) / float64(perPage)))

		// Paginate by date groups
		start := (page - 1) * perPage
		end := start + perPage
		if start > len(dates) {
			start = len(dates)
		}
		if end > len(dates) {
			end = len(dates)
		}
		pageDates := dates[start:end]

		groups := make([]dateGroup, len(pageDates))
		for i, d := range pageDates {
			groups[i] = dateGroup{
				Date:  d,
				Items: dateMap[d],
			}
		}

		return c.JSON(paginatedResponse{
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
			TotalItems: totalItems,
			Data:       groups,
		})
	}
}

// ClearLogsByRepo deletes all deployment logs (and their items) for a repo + branch.
// DELETE /api/logs/repo/:repo_name?branch=develop
func ClearLogsByRepo(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		repoName := c.Params("repo_name")
		branch := c.Query("branch", "")

		var repoCfg models.RepositoryConfig
		if err := db.Where("name = ?", repoName).First(&repoCfg).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "repository not found"})
		}

		logQuery := db.Where("repo_config_id = ?", repoCfg.ID)
		if branch != "" {
			logQuery = logQuery.Where("environment = ?", branch)
		}

		var deploymentLogs []models.DeploymentLog
		if err := logQuery.Find(&deploymentLogs).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if len(deploymentLogs) == 0 {
			return c.JSON(fiber.Map{"message": "no logs to clear", "deleted": 0})
		}

		logIDs := make([]uint, len(deploymentLogs))
		for i, dl := range deploymentLogs {
			logIDs[i] = dl.ID
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("deployment_log_id IN ?", logIDs).Delete(&models.LogItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", logIDs).Delete(&models.DeploymentLog{}).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		label := repoName
		if branch != "" {
			label = fmt.Sprintf("%s/%s", repoName, branch)
		}
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Cleared %d deployment log(s) for %s", len(deploymentLogs), label),
			"deleted": len(deploymentLogs),
		})
	}
}
