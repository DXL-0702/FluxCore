package model

import "time"

type ProjectStatus string

const (
	ProjectStatusActive ProjectStatus = "active"
)

type Project struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	Name         string        `gorm:"not null;uniqueIndex;size:120" json:"name"`
	Description  string        `gorm:"size:500" json:"description"`
	Status       ProjectStatus `gorm:"not null;size:40;default:active" json:"status"`
	Repositories []Repository  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"repositories,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
