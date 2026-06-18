package model

import "time"

type Config struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"not null;uniqueIndex;size:120" json:"key"`
	Value     string    `gorm:"size:1000" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
