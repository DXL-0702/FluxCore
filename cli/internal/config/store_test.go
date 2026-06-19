package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreInitCreatesConfigAndGitignore(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(root, func() time.Time { return now })

	cfg, err := store.Init(InitOptions{
		ServerURL:    "http://127.0.0.1:8080",
		Token:        "secret-token",
		UpdateServer: true,
		UpdateToken:  true,
	})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if cfg.ServerURL != "http://127.0.0.1:8080" {
		t.Fatalf("ServerURL = %q", cfg.ServerURL)
	}
	if cfg.Token != "secret-token" {
		t.Fatalf("Token = %q", cfg.Token)
	}

	configInfo, err := os.Stat(filepath.Join(root, DirectoryName, FileName))
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if mode := configInfo.Mode().Perm(); mode != fileMode {
		t.Fatalf("config mode = %v, want %v", mode, fileMode)
	}

	directoryInfo, err := os.Stat(filepath.Join(root, DirectoryName))
	if err != nil {
		t.Fatalf("stat config directory: %v", err)
	}
	if mode := directoryInfo.Mode().Perm(); mode != directoryMode {
		t.Fatalf("directory mode = %v, want %v", mode, directoryMode)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(gitignore) != ".fluxcore/\n" {
		t.Fatalf(".gitignore = %q", string(gitignore))
	}

	entries, err := os.ReadDir(filepath.Join(root, DirectoryName))
	if err != nil {
		t.Fatalf("read config directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != FileName {
		t.Fatalf("unexpected files in config directory: %v", entries)
	}
}

func TestStoreInitIsIdempotentWithoutExplicitOverrides(t *testing.T) {
	root := t.TempDir()
	times := []time.Time{
		time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 19, 11, 0, 0, 0, time.UTC),
	}
	index := 0
	store := NewStoreWithClock(root, func() time.Time {
		current := times[index]
		if index < len(times)-1 {
			index++
		}
		return current
	})

	first, err := store.Init(InitOptions{
		ServerURL:    "http://custom.local:8080",
		Token:        "first-token",
		UpdateServer: true,
		UpdateToken:  true,
	})
	if err != nil {
		t.Fatalf("first Init() error = %v", err)
	}

	second, err := store.Init(InitOptions{
		ServerURL: "http://127.0.0.1:8080",
		Token:     "",
	})
	if err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if second.ServerURL != first.ServerURL {
		t.Fatalf("ServerURL = %q, want %q", second.ServerURL, first.ServerURL)
	}
	if second.Token != first.Token {
		t.Fatalf("Token = %q, want %q", second.Token, first.Token)
	}
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("CreatedAt = %v, want %v", second.CreatedAt, first.CreatedAt)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("UpdatedAt = %v, want after %v", second.UpdatedAt, first.UpdatedAt)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if count := strings.Count(string(gitignore), ".fluxcore/"); count != 1 {
		t.Fatalf(".gitignore contains .fluxcore/ %d times", count)
	}
}

func TestStoreInitAllowsExplicitOverrides(t *testing.T) {
	root := t.TempDir()
	store := NewStoreWithClock(root, func() time.Time {
		return time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	})

	if _, err := store.Init(InitOptions{
		ServerURL:    "http://old.local:8080",
		Token:        "old-token",
		UpdateServer: true,
		UpdateToken:  true,
	}); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}

	cfg, err := store.Init(InitOptions{
		ServerURL:    "http://new.local:8080",
		Token:        "",
		UpdateServer: true,
		UpdateToken:  true,
	})
	if err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if cfg.ServerURL != "http://new.local:8080" {
		t.Fatalf("ServerURL = %q", cfg.ServerURL)
	}
	if cfg.Token != "" {
		t.Fatalf("Token = %q, want empty", cfg.Token)
	}
}

func TestStoreInitFillsMissingTokenWithoutExplicitOverride(t *testing.T) {
	root := t.TempDir()
	store := NewStoreWithClock(root, func() time.Time {
		return time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	})

	if _, err := store.Init(InitOptions{
		ServerURL:    "http://127.0.0.1:8080",
		Token:        "",
		UpdateServer: true,
		UpdateToken:  true,
	}); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}

	cfg, err := store.Init(InitOptions{
		ServerURL: "http://127.0.0.1:8080",
		Token:     "env-token",
	})
	if err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if cfg.Token != "env-token" {
		t.Fatalf("Token = %q, want %q", cfg.Token, "env-token")
	}
}

func TestStoreInitDoesNotReplaceExistingTokenWithoutExplicitOverride(t *testing.T) {
	root := t.TempDir()
	store := NewStoreWithClock(root, func() time.Time {
		return time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	})

	if _, err := store.Init(InitOptions{
		ServerURL:    "http://127.0.0.1:8080",
		Token:        "existing-token",
		UpdateServer: true,
		UpdateToken:  true,
	}); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}

	cfg, err := store.Init(InitOptions{
		ServerURL: "http://127.0.0.1:8080",
		Token:     "env-token",
	})
	if err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if cfg.Token != "existing-token" {
		t.Fatalf("Token = %q, want %q", cfg.Token, "existing-token")
	}
}

func TestStoreLoadMissingConfig(t *testing.T) {
	_, err := NewStore(t.TempDir()).Load()
	if err != ErrConfigNotFound {
		t.Fatalf("Load() error = %v, want %v", err, ErrConfigNotFound)
	}
}
