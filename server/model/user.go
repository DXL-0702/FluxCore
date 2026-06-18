package model

import "time"

type UserRole string

const (
	UserRoleOwner UserRole = "owner"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex;size:120" json:"name"`
	Role      UserRole  `gorm:"not null;size:40;default:'owner'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
