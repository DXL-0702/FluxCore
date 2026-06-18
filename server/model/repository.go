package model

import "time"

type Repository struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `gorm:"not null;index;uniqueIndex:idx_project_repository_remote" json:"project_id"`
	Project       Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Name          string    `gorm:"not null;size:120" json:"name"`
	LocalPath     string    `gorm:"not null;uniqueIndex;size:500" json:"local_path"`
	RemoteURL     string    `gorm:"not null;size:500;uniqueIndex:idx_project_repository_remote" json:"remote_url"`
	DefaultBranch string    `gorm:"not null;size:120;default:main" json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
