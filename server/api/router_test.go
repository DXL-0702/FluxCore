package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaxson/FluxCore/server/config"
)

func TestHealthRoute(t *testing.T) {
	router := NewRouter(config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Service  string `json:"service"`
		Status   string `json:"status"`
		Database struct {
			Type string `json:"type"`
		} `json:"database"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Service != "fluxcore-server" {
		t.Fatalf("service = %q, want %q", response.Service, "fluxcore-server")
	}
	if response.Status != "ok" {
		t.Fatalf("status = %q, want %q", response.Status, "ok")
	}
	if response.Database.Type != "sqlite" {
		t.Fatalf("database.type = %q, want %q", response.Database.Type, "sqlite")
	}
}
