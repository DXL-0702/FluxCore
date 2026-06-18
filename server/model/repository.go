package model

import "time"

type Repository struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `gorm:"not null;uniqueIndex:idx_project_repository_remote;index:idx_project_repository_name,unique" json:"project_id"`
	Project       Project   `json:"-"`
	Name          string    `gorm:"not null;size:120;uniqueIndex:idx_project_repository_name" json:"name"`
	LocalPath     string    `gorm:"not null;uniqueIndex;size:500" json:"local_path"`
	RemoteURL     string    `gorm:"not null;size:500;uniqueIndex:idx_project_repository_remote" json:"remote_url"`
	DefaultBranch string    `gorm:"not null;size:120;default:main" json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
