package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/jaxson/FluxCore/server/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultMaxIdleConns    = 5
	defaultMaxOpenConns    = 25
	defaultConnMaxLifetime = time.Hour
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dialector, err := dialectorFor(cfg)
	if err != nil {
		return nil, err
	}

	conn, err := gorm.Open(dialector, &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.Type, err)
	}

	if err := configurePool(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func Ping(ctx context.Context, conn *gorm.DB) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("get database handle: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func Close(conn *gorm.DB) error {
	if conn == nil {
		return nil
	}

	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("get database handle: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}

func configurePool(conn *gorm.DB) error {
	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("get database handle for pool config: %w", err)
	}

	sqlDB.SetMaxIdleConns(defaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(defaultConnMaxLifetime)

	return nil
}

func dialectorFor(cfg config.DatabaseConfig) (gorm.Dialector, error) {
	dbType := strings.ToLower(strings.TrimSpace(cfg.Type))

	switch dbType {
	case config.DBTypeSQLite:
		sqlitePath := strings.TrimSpace(cfg.SQLitePath)
		if sqlitePath == "" {
			return nil, fmt.Errorf("SQLite path must not be empty")
		}
		return sqlite.Open(sqlitePath), nil
	case config.DBTypePostgres:
		postgresDSN := strings.TrimSpace(cfg.PostgresDSN)
		if postgresDSN == "" {
			return nil, fmt.Errorf("PostgreSQL DSN must not be empty")
		}
		return postgres.Open(postgresDSN), nil
	default:
		return nil, fmt.Errorf("unsupported database type %q", cfg.Type)
	}
}
