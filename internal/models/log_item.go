package models

import "time"

type LogItem struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	DeploymentLogID uint      `gorm:"not null" json:"deployment_log_id"`
	Title           string    `gorm:"not null" json:"title"`
	PRNumber        *int      `json:"pr_number,omitempty"`
	CommitSHA       string    `gorm:"not null" json:"commit_sha"`
	MergedAt        time.Time `json:"merged_at"`
	Category        string    `json:"category"`
	Description     string    `json:"description,omitempty"`
}
