package models

import "time"

type DeploymentLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RepoConfigID uint      `gorm:"not null" json:"repo_config_id"`
	Environment  string    `gorm:"not null" json:"environment"`
	GeneratedAt  time.Time `gorm:"not null" json:"generated_at"`
	CreatedAt    time.Time `json:"created_at"`
	LogItems     []LogItem `gorm:"foreignKey:DeploymentLogID" json:"log_items,omitempty"`
}
