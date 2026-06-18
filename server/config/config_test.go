package config

import "testing"

func TestLoadFromLookupDefaults(t *testing.T) {
	cfg, err := LoadFromLookup(mapLookup(nil))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.Server.Port != DefaultPort {
		t.Fatalf("Server.Port = %q, want %q", cfg.Server.Port, DefaultPort)
	}
	if cfg.Server.Address != ":8080" {
		t.Fatalf("Server.Address = %q, want %q", cfg.Server.Address, ":8080")
	}
	if cfg.Database.Type != DefaultDBType {
		t.Fatalf("Database.Type = %q, want %q", cfg.Database.Type, DefaultDBType)
	}
	if cfg.Database.SQLitePath != DefaultSQLitePath {
		t.Fatalf("Database.SQLitePath = %q, want %q", cfg.Database.SQLitePath, DefaultSQLitePath)
	}
}

func TestLoadFromLookupHostPortAddress(t *testing.T) {
	cfg, err := LoadFromLookup(mapLookup(map[string]string{
		"PORT": "127.0.0.1:18080",
	}))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.Server.Address != "127.0.0.1:18080" {
		t.Fatalf("Server.Address = %q, want %q", cfg.Server.Address, "127.0.0.1:18080")
	}
}

func TestLoadFromLookupInvalidDBType(t *testing.T) {
	_, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_TYPE": "mysql",
	}))
	if err == nil {
		t.Fatal("LoadFromLookup() error = nil, want error")
	}
}

func TestLoadFromLookupPostgresRequiresDSN(t *testing.T) {
	_, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_TYPE": "postgres",
	}))
	if err == nil {
		t.Fatal("LoadFromLookup() error = nil, want error")
	}
}

func TestLoadFromLookupPostgres(t *testing.T) {
	cfg, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_TYPE":      "postgres",
		"POSTGRES_DSN": "postgres://user:pass@localhost:5432/fluxcore",
		"API_TOKEN":    " test-token ",
	}))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.Database.Type != "postgres" {
		t.Fatalf("Database.Type = %q, want %q", cfg.Database.Type, "postgres")
	}
	if cfg.Database.PostgresDSN == "" {
		t.Fatal("Database.PostgresDSN is empty")
	}
	if cfg.Security.APIToken != "test-token" {
		t.Fatalf("Security.APIToken = %q, want %q", cfg.Security.APIToken, "test-token")
	}
}

func TestOptionalEnvWhitespaceIsAbsent(t *testing.T) {
	value, ok := optionalEnv(mapLookup(map[string]string{
		"API_TOKEN": "   ",
	}), "API_TOKEN")
	if ok {
		t.Fatal("optionalEnv() ok = true, want false")
	}
	if value != "" {
		t.Fatalf("optionalEnv() value = %q, want empty", value)
	}
}

func mapLookup(values map[string]string) LookupEnv {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
