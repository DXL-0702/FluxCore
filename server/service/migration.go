package service

import (
	"fmt"

	"github.com/jaxson/FluxCore/server/model"
	"gorm.io/gorm"
)

func Migrate(conn *gorm.DB) error {
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := conn.AutoMigrate(
		&model.Project{},
		&model.Repository{},
		&model.User{},
		&model.Config{},
	); err != nil {
		return fmt.Errorf("auto migrate database: %w", err)
	}

	return nil
}
