package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultPort       = "8080"
	DefaultDBType     = "sqlite"
	DefaultSQLitePath = "fluxcore.db"
)

type LookupEnv func(string) (string, bool)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port    string
	Address string
}

type DatabaseConfig struct {
	Type        string
	SQLitePath  string
	PostgresDSN string
}

type SecurityConfig struct {
	APIToken string
}

func Load() (Config, error) {
	return LoadFromLookup(os.LookupEnv)
}

func LoadFromLookup(lookup LookupEnv) (Config, error) {
	if lookup == nil {
		return Config{}, fmt.Errorf("environment lookup function is required")
	}

	port, err := envOrDefault(lookup, "PORT", DefaultPort)
	if err != nil {
		return Config{}, err
	}

	address, err := normalizeAddress(port)
	if err != nil {
		return Config{}, err
	}

	dbType, err := envOrDefault(lookup, "DB_TYPE", DefaultDBType)
	if err != nil {
		return Config{}, err
	}
	dbType = strings.ToLower(dbType)

	database := DatabaseConfig{
		Type: dbType,
	}

	switch dbType {
	case "sqlite":
		sqlitePath, err := envOrDefault(lookup, "SQLITE_PATH", DefaultSQLitePath)
		if err != nil {
			return Config{}, err
		}
		database.SQLitePath = sqlitePath
	case "postgres":
		postgresDSN, err := requiredEnv(lookup, "POSTGRES_DSN")
		if err != nil {
			return Config{}, err
		}
		database.PostgresDSN = postgresDSN
	default:
		return Config{}, fmt.Errorf("DB_TYPE must be one of sqlite or postgres, got %q", dbType)
	}

	apiToken, _ := optionalEnv(lookup, "API_TOKEN")

	return Config{
		Server: ServerConfig{
			Port:    port,
			Address: address,
		},
		Database: database,
		Security: SecurityConfig{
			APIToken: apiToken,
		},
	}, nil
}

func normalizeAddress(port string) (string, error) {
	port = strings.TrimSpace(port)
	if port == "" {
		return "", fmt.Errorf("PORT must not be empty")
	}

	if strings.Contains(port, ":") {
		return port, nil
	}

	return ":" + port, nil
}

func envOrDefault(lookup LookupEnv, key string, fallback string) (string, error) {
	value, ok := lookup(key)
	if !ok {
		return fallback, nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be empty", key)
	}

	return value, nil
}

func requiredEnv(lookup LookupEnv, key string) (string, error) {
	value, ok := lookup(key)
	if !ok {
		return "", fmt.Errorf("%s is required", key)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be empty", key)
	}

	return value, nil
}

func optionalEnv(lookup LookupEnv, key string) (string, bool) {
	value, ok := lookup(key)
	if !ok {
		return "", false
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return value, true
}
