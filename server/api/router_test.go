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
		Security: config.SecurityConfig{
			APIToken: "test-token",
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

func TestAPIRouteRejectsMissingToken(t *testing.T) {
	router := NewRouter(testConfig())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil)

	router.ServeHTTP(recorder, request)

	assertUnauthorized(t, recorder)
}

func TestAPIRouteRejectsMalformedAuthorizationHeader(t *testing.T) {
	router := NewRouter(testConfig())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	request.Header.Set("Authorization", "Token test-token")

	router.ServeHTTP(recorder, request)

	assertUnauthorized(t, recorder)
}

func TestAPIRouteRejectsInvalidToken(t *testing.T) {
	router := NewRouter(testConfig())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	request.Header.Set("Authorization", "Bearer wrong-token")

	router.ServeHTTP(recorder, request)

	assertUnauthorized(t, recorder)
}

func TestAPIRouteAllowsValidTokenToReachRouter(t *testing.T) {
	router := NewRouter(testConfig())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	request.Header.Set("Authorization", "Bearer test-token")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestAPIRouteRejectsRequestsWhenTokenIsNotConfigured(t *testing.T) {
	router := NewRouter(config.Config{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	request.Header.Set("Authorization", "Bearer test-token")

	router.ServeHTTP(recorder, request)

	assertUnauthorized(t, recorder)
}

func assertUnauthorized(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if recorder.Header().Get("WWW-Authenticate") != `Bearer realm="fluxcore"` {
		t.Fatalf("WWW-Authenticate = %q, want %q", recorder.Header().Get("WWW-Authenticate"), `Bearer realm="fluxcore"`)
	}

	var response struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Error.Code != "unauthorized" {
		t.Fatalf("error.code = %q, want %q", response.Error.Code, "unauthorized")
	}
	if response.Error.Message != "authentication required" {
		t.Fatalf("error.message = %q, want %q", response.Error.Message, "authentication required")
	}
}

func testConfig() config.Config {
	return config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
		},
		Security: config.SecurityConfig{
			APIToken: "test-token",
		},
	}
}
