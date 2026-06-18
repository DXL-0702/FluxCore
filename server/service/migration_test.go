package service

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jaxson/FluxCore/server/config"
	"github.com/jaxson/FluxCore/server/db"
	"github.com/jaxson/FluxCore/server/model"
	"gorm.io/gorm"
)

func TestMigrateCreatesPhaseOneTables(t *testing.T) {
	conn := openTestDB(t)

	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	for _, table := range []interface{}{
		&model.Project{},
		&model.Repository{},
		&model.User{},
		&model.Config{},
	} {
		if !conn.Migrator().HasTable(table) {
			t.Fatalf("expected table for %T to exist", table)
		}
	}
}

func TestMigrateSupportsProjectRepositoryAssociation(t *testing.T) {
	conn := openTestDB(t)

	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	project := model.Project{
		Name:        "fluxcore",
		Description: "local-first dev console",
		Status:      model.ProjectStatusActive,
		Repositories: []model.Repository{
			{
				Name:          "FluxCore",
				LocalPath:     "/tmp/fluxcore",
				RemoteURL:     "git@github.com:DXL-0702/FluxCore.git",
				DefaultBranch: "develop",
			},
		},
	}
	if err := conn.Create(&project).Error; err != nil {
		t.Fatalf("Create(Project) error = %v", err)
	}

	var loaded model.Project
	if err := conn.Preload("Repositories").First(&loaded, "name = ?", "fluxcore").Error; err != nil {
		t.Fatalf("First(Project) error = %v", err)
	}

	if len(loaded.Repositories) != 1 {
		t.Fatalf("len(Repositories) = %d, want 1", len(loaded.Repositories))
	}
	if loaded.Repositories[0].ProjectID != loaded.ID {
		t.Fatalf("Repository.ProjectID = %d, want %d", loaded.Repositories[0].ProjectID, loaded.ID)
	}
}

func TestMigrateRejectsNilConnection(t *testing.T) {
	err := Migrate(nil)
	if err == nil {
		t.Fatal("Migrate() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "database connection is nil") {
		t.Fatalf("Migrate() error = %q, want message containing %q", err.Error(), "database connection is nil")
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	conn, err := db.Open(config.DatabaseConfig{
		Type:       config.DBTypeSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "fluxcore.db"),
	})
	if err != nil {
		t.Fatalf("db.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(conn); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})

	return conn
}
