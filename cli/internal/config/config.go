package config

import "time"

const (
	DirectoryName = ".fluxcore"
	FileName      = "config.json"
)

type Config struct {
	ServerURL  string     `json:"server_url"`
	Token      string     `json:"token"`
	Project    Project    `json:"project"`
	Repository Repository `json:"repository"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Project struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Repository struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	LocalPath     string `json:"local_path"`
	RemoteURL     string `json:"remote_url"`
	DefaultBranch string `json:"default_branch"`
}

func New(serverURL string, token string, now time.Time) Config {
	now = now.UTC()
	return Config{
		ServerURL: serverURL,
		Token:     token,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (cfg Config) IsLinked() bool {
	return cfg.Project.ID != 0 && cfg.Repository.ID != 0
}
