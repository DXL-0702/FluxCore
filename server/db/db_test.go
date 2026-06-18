package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaxson/FluxCore/server/config"
)

func TestOpenSQLiteAndPing(t *testing.T) {
	sqlitePath := filepath.Join(t.TempDir(), "fluxcore.db")

	conn, err := Open(config.DatabaseConfig{
		Type:       config.DBTypeSQLite,
		SQLitePath: sqlitePath,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := Close(conn); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	if err := Ping(conn); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	if _, err := os.Stat(sqlitePath); err != nil {
		t.Fatalf("os.Stat(%q) error = %v", sqlitePath, err)
	}
}

func TestOpenRejectsEmptySQLitePath(t *testing.T) {
	_, err := Open(config.DatabaseConfig{
		Type: config.DBTypeSQLite,
	})
	if err == nil {
		t.Fatal("Open() error = nil, want error")
	}
}

func TestOpenRejectsEmptyPostgresDSN(t *testing.T) {
	_, err := Open(config.DatabaseConfig{
		Type: config.DBTypePostgres,
	})
	if err == nil {
		t.Fatal("Open() error = nil, want error")
	}
}

func TestOpenPostgresDoesNotRequireRunningServer(t *testing.T) {
	conn, err := Open(config.DatabaseConfig{
		Type:        config.DBTypePostgres,
		PostgresDSN: "postgres://user:pass@127.0.0.1:1/fluxcore?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := Close(conn); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
}

func TestOpenRejectsUnsupportedDBType(t *testing.T) {
	_, err := Open(config.DatabaseConfig{
		Type: "mysql",
	})
	if err == nil {
		t.Fatal("Open() error = nil, want error")
	}
}

func TestPingRejectsNilConnection(t *testing.T) {
	if err := Ping(nil); err == nil {
		t.Fatal("Ping() error = nil, want error")
	}
}
