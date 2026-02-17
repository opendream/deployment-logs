package models

import "time"

type RepositoryConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	RepoURL       string    `gorm:"not null" json:"repo_url"`
	Provider      string    `gorm:"not null;default:github" json:"provider"`
	DefaultBranch string    `gorm:"not null;default:main" json:"default_branch"`
	SSHKey        string    `gorm:"type:text" json:"ssh_key,omitempty"`
	Branches      string    `gorm:"type:text" json:"branches"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
