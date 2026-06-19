package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/renameio/v2"
)

const (
	directoryMode = 0o700
	fileMode      = 0o600
	gitignoreMode = 0o644
)

var ErrConfigNotFound = errors.New("fluxcore config not found")

type Store struct {
	repositoryRoot string
	now            func() time.Time
}

type InitOptions struct {
	ServerURL    string
	Token        string
	UpdateServer bool
	UpdateToken  bool
}

func NewStore(repositoryRoot string) Store {
	return Store{
		repositoryRoot: repositoryRoot,
		now:            time.Now,
	}
}

func NewStoreWithClock(repositoryRoot string, now func() time.Time) Store {
	if now == nil {
		now = time.Now
	}
	return Store{
		repositoryRoot: repositoryRoot,
		now:            now,
	}
}

func (store Store) Init(options InitOptions) (Config, error) {
	options.ServerURL = strings.TrimSpace(options.ServerURL)
	if options.ServerURL == "" {
		return Config{}, fmt.Errorf("server URL is required")
	}

	now := store.now().UTC()
	path := store.configPath()
	existing, err := store.Load()
	switch {
	case err == nil:
		if options.UpdateServer || strings.TrimSpace(existing.ServerURL) == "" {
			existing.ServerURL = options.ServerURL
		}
		if options.UpdateToken {
			existing.Token = options.Token
		}
		existing.UpdatedAt = now
		if err := store.Save(existing); err != nil {
			return Config{}, err
		}
		if err := store.EnsureGitignored(); err != nil {
			return Config{}, err
		}
		return existing, nil
	case !errors.Is(err, ErrConfigNotFound):
		return Config{}, err
	}

	if err := os.MkdirAll(filepath.Dir(path), directoryMode); err != nil {
		return Config{}, fmt.Errorf("create fluxcore directory: %w", err)
	}

	cfg := New(options.ServerURL, options.Token, now)
	if err := store.Save(cfg); err != nil {
		return Config{}, err
	}
	if err := store.EnsureGitignored(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (store Store) Load() (Config, error) {
	data, err := os.ReadFile(store.configPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, ErrConfigNotFound
		}
		return Config{}, fmt.Errorf("read fluxcore config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse fluxcore config: %w", err)
	}
	return cfg, nil
}

func (store Store) Save(cfg Config) error {
	if strings.TrimSpace(cfg.ServerURL) == "" {
		return fmt.Errorf("server URL is required")
	}

	now := store.now().UTC()
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = now
	}
	cfg.UpdatedAt = now

	if err := os.MkdirAll(store.directoryPath(), directoryMode); err != nil {
		return fmt.Errorf("create fluxcore directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode fluxcore config: %w", err)
	}
	data = append(data, '\n')

	return writeFileAtomic(store.configPath(), data, fileMode)
}

func (store Store) EnsureGitignored() error {
	gitignorePath := filepath.Join(store.repositoryRoot, ".gitignore")
	existing, err := os.ReadFile(gitignorePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read .gitignore: %w", err)
	}

	if hasGitignoreEntry(existing, DirectoryName+"/") {
		return nil
	}

	var buffer bytes.Buffer
	buffer.Write(existing)
	if len(existing) > 0 && !bytes.HasSuffix(existing, []byte("\n")) {
		buffer.WriteByte('\n')
	}
	buffer.WriteString(DirectoryName)
	buffer.WriteString("/\n")

	if err := os.WriteFile(gitignorePath, buffer.Bytes(), gitignoreMode); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}

func (store Store) configPath() string {
	return filepath.Join(store.directoryPath(), FileName)
}

func (store Store) directoryPath() string {
	return filepath.Join(store.repositoryRoot, DirectoryName)
}

func hasGitignoreEntry(data []byte, entry string) bool {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == entry || line == strings.TrimSuffix(entry, "/") {
			return true
		}
	}
	return false
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := renameio.WriteFile(path, data, mode); err != nil {
		return fmt.Errorf("save fluxcore config: %w", err)
	}
	return nil
}
