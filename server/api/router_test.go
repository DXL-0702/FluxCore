package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jaxson/FluxCore/server/config"
	"github.com/jaxson/FluxCore/server/db"
	"github.com/jaxson/FluxCore/server/service"
	"gorm.io/gorm"
)

func TestHealthRoute(t *testing.T) {
	router, _ := newTestRouter(t)

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
	decodeResponse(t, recorder, &response)

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
	router, _ := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects", "", nil)

	assertUnauthorized(t, recorder)
}

func TestAPIRouteRejectsMalformedAuthorizationHeader(t *testing.T) {
	router, _ := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects", "Token test-token", nil)

	assertUnauthorized(t, recorder)
}

func TestAPIRouteRejectsInvalidToken(t *testing.T) {
	router, _ := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects", "Bearer wrong-token", nil)

	assertUnauthorized(t, recorder)
}

func TestAPIRouteAllowsValidToken(t *testing.T) {
	router, _ := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects", testAuthorizationHeader(), nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestUnknownAPIRouteRequiresAuthenticationBeforeNotFound(t *testing.T) {
	router, _ := newTestRouter(t)

	missingToken := performJSONRequest(router, http.MethodGet, "/api/unknown", "", nil)
	assertUnauthorized(t, missingToken)

	validToken := performJSONRequest(router, http.MethodGet, "/api/unknown", testAuthorizationHeader(), nil)
	if validToken.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", validToken.Code, http.StatusNotFound)
	}
}

func TestAPIRouteRejectsRequestsWhenTokenIsNotConfigured(t *testing.T) {
	_, conn := newTestRouter(t)
	router := NewRouter(config.Config{}, conn)

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects", testAuthorizationHeader(), nil)

	assertUnauthorized(t, recorder)
}

func TestCreateProject(t *testing.T) {
	router, _ := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodPost, "/api/projects", testAuthorizationHeader(), map[string]string{
		"name":        "FluxCore",
		"description": "local-first dev console",
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Project struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
		} `json:"project"`
	}
	decodeResponse(t, recorder, &response)

	if response.Project.ID == 0 {
		t.Fatal("project.id = 0, want generated id")
	}
	if response.Project.Name != "FluxCore" {
		t.Fatalf("project.name = %q, want %q", response.Project.Name, "FluxCore")
	}
	if response.Project.Description != "local-first dev console" {
		t.Fatalf("project.description = %q, want %q", response.Project.Description, "local-first dev console")
	}
	if response.Project.Status != "active" {
		t.Fatalf("project.status = %q, want %q", response.Project.Status, "active")
	}
}

func TestListProjects(t *testing.T) {
	router, _ := newTestRouter(t)
	createProject(t, router, "FluxCore")

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects", testAuthorizationHeader(), nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Projects []struct {
			Name string `json:"name"`
		} `json:"projects"`
	}
	decodeResponse(t, recorder, &response)

	if len(response.Projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(response.Projects))
	}
	if response.Projects[0].Name != "FluxCore" {
		t.Fatalf("projects[0].name = %q, want %q", response.Projects[0].Name, "FluxCore")
	}
}

func TestCreateProjectRejectsInvalidPayload(t *testing.T) {
	router, _ := newTestRouter(t)

	cases := []struct {
		name string
		body string
	}{
		{
			name: "empty name",
			body: `{"name":"   "}`,
		},
		{
			name: "unknown field",
			body: `{"name":"FluxCore","unexpected":true}`,
		},
		{
			name: "unsupported status",
			body: `{"name":"FluxCore","status":"archived"}`,
		},
		{
			name: "oversized name",
			body: fmt.Sprintf(`{"name":%q}`, strings.Repeat("x", 121)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRawRequest(router, http.MethodPost, "/api/projects", testAuthorizationHeader(), tc.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
		})
	}
}

func TestCreateProjectCountsUTF8Characters(t *testing.T) {
	router, _ := newTestRouter(t)

	validName := strings.Repeat("项", 120)
	recorder := performJSONRequest(router, http.MethodPost, "/api/projects", testAuthorizationHeader(), map[string]string{
		"name": validName,
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	invalidName := strings.Repeat("项", 121)
	recorder = performJSONRequest(router, http.MethodPost, "/api/projects", testAuthorizationHeader(), map[string]string{
		"name": invalidName,
	})
	assertAPIError(t, recorder, http.StatusBadRequest, "invalid_request")
}

func TestCreateProjectRejectsDuplicateName(t *testing.T) {
	router, _ := newTestRouter(t)
	createProject(t, router, "FluxCore")

	recorder := performJSONRequest(router, http.MethodPost, "/api/projects", testAuthorizationHeader(), map[string]string{
		"name": "FluxCore",
	})

	assertAPIError(t, recorder, http.StatusConflict, "conflict")
}

func TestCreateRepository(t *testing.T) {
	router, _ := newTestRouter(t)
	projectID := createProject(t, router, "FluxCore")

	recorder := performJSONRequest(router, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", projectID), testAuthorizationHeader(), map[string]string{
		"name":           "server",
		"local_path":     "/tmp/fluxcore",
		"remote_url":     "git@example.com:team/fluxcore.git",
		"default_branch": "develop",
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Repository struct {
			ID            uint   `json:"id"`
			ProjectID     uint   `json:"project_id"`
			Name          string `json:"name"`
			LocalPath     string `json:"local_path"`
			RemoteURL     string `json:"remote_url"`
			DefaultBranch string `json:"default_branch"`
		} `json:"repository"`
	}
	decodeResponse(t, recorder, &response)

	if response.Repository.ID == 0 {
		t.Fatal("repository.id = 0, want generated id")
	}
	if response.Repository.ProjectID != projectID {
		t.Fatalf("repository.project_id = %d, want %d", response.Repository.ProjectID, projectID)
	}
	if response.Repository.DefaultBranch != "develop" {
		t.Fatalf("repository.default_branch = %q, want %q", response.Repository.DefaultBranch, "develop")
	}
}

func TestListRepositories(t *testing.T) {
	router, _ := newTestRouter(t)
	projectID := createProject(t, router, "FluxCore")
	createRepository(t, router, projectID, "server", "/tmp/fluxcore-server", "git@example.com:team/server.git")

	recorder := performJSONRequest(router, http.MethodGet, fmt.Sprintf("/api/projects/%d/repositories", projectID), testAuthorizationHeader(), nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Repositories []struct {
			Name string `json:"name"`
		} `json:"repositories"`
	}
	decodeResponse(t, recorder, &response)

	if len(response.Repositories) != 1 {
		t.Fatalf("len(repositories) = %d, want 1", len(response.Repositories))
	}
	if response.Repositories[0].Name != "server" {
		t.Fatalf("repositories[0].name = %q, want %q", response.Repositories[0].Name, "server")
	}
}

func TestRepositoryRoutesRejectMissingProject(t *testing.T) {
	router, _ := newTestRouter(t)

	createRecorder := performJSONRequest(router, http.MethodPost, "/api/projects/999/repositories", testAuthorizationHeader(), map[string]string{
		"name":       "server",
		"local_path": "/tmp/fluxcore-server",
		"remote_url": "git@example.com:team/server.git",
	})
	assertAPIError(t, createRecorder, http.StatusNotFound, "not_found")

	listRecorder := performJSONRequest(router, http.MethodGet, "/api/projects/999/repositories", testAuthorizationHeader(), nil)
	assertAPIError(t, listRecorder, http.StatusNotFound, "not_found")
}

func TestCreateRepositoryRejectsInvalidProjectID(t *testing.T) {
	router, _ := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodGet, "/api/projects/not-a-number/repositories", testAuthorizationHeader(), nil)

	assertAPIError(t, recorder, http.StatusBadRequest, "invalid_request")
}

func TestCreateRepositoryRejectsDuplicateIdentity(t *testing.T) {
	router, _ := newTestRouter(t)
	projectID := createProject(t, router, "FluxCore")
	otherProjectID := createProject(t, router, "Other")
	createRepository(t, router, projectID, "server", "/tmp/fluxcore-server", "git@example.com:team/server.git")

	duplicateName := performJSONRequest(router, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", projectID), testAuthorizationHeader(), map[string]string{
		"name":       "server",
		"local_path": "/tmp/fluxcore-server-copy",
		"remote_url": "git@example.com:team/server-copy.git",
	})
	assertAPIError(t, duplicateName, http.StatusConflict, "conflict")

	duplicateRemote := performJSONRequest(router, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", projectID), testAuthorizationHeader(), map[string]string{
		"name":       "server-copy",
		"local_path": "/tmp/fluxcore-server-copy",
		"remote_url": "git@example.com:team/server.git",
	})
	assertAPIError(t, duplicateRemote, http.StatusConflict, "conflict")

	duplicateLocalPath := performJSONRequest(router, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", otherProjectID), testAuthorizationHeader(), map[string]string{
		"name":       "server",
		"local_path": "/tmp/fluxcore-server",
		"remote_url": "git@example.com:team/server.git",
	})
	assertAPIError(t, duplicateLocalPath, http.StatusConflict, "conflict")
}

func TestCreateRepositoryAllowsSameNameAndRemoteAcrossProjects(t *testing.T) {
	router, _ := newTestRouter(t)
	projectID := createProject(t, router, "FluxCore")
	otherProjectID := createProject(t, router, "Other")
	createRepository(t, router, projectID, "server", "/tmp/fluxcore-server", "git@example.com:team/server.git")

	recorder := performJSONRequest(router, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", otherProjectID), testAuthorizationHeader(), map[string]string{
		"name":       "server",
		"local_path": "/tmp/other-server",
		"remote_url": "git@example.com:team/server.git",
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
}

func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
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

	if err := service.Migrate(conn); err != nil {
		t.Fatalf("service.Migrate() error = %v", err)
	}

	return NewRouter(testConfig(), conn), conn
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

func createProject(t *testing.T, router *gin.Engine, name string) uint {
	t.Helper()

	recorder := performJSONRequest(router, http.MethodPost, "/api/projects", testAuthorizationHeader(), map[string]string{
		"name": name,
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create project status code = %d, want %d, body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Project struct {
			ID uint `json:"id"`
		} `json:"project"`
	}
	decodeResponse(t, recorder, &response)
	if response.Project.ID == 0 {
		t.Fatal("project.id = 0, want generated id")
	}

	return response.Project.ID
}

func createRepository(t *testing.T, router *gin.Engine, projectID uint, name string, localPath string, remoteURL string) uint {
	t.Helper()

	recorder := performJSONRequest(router, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", projectID), testAuthorizationHeader(), map[string]string{
		"name":       name,
		"local_path": localPath,
		"remote_url": remoteURL,
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create repository status code = %d, want %d, body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Repository struct {
			ID uint `json:"id"`
		} `json:"repository"`
	}
	decodeResponse(t, recorder, &response)
	if response.Repository.ID == 0 {
		t.Fatal("repository.id = 0, want generated id")
	}

	return response.Repository.ID
}

func performJSONRequest(router *gin.Engine, method string, path string, authorization string, body interface{}) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}

	return performRequest(router, method, path, authorization, payload)
}

func performRawRequest(router *gin.Engine, method string, path string, authorization string, body string) *httptest.ResponseRecorder {
	return performRequest(router, method, path, authorization, []byte(body))
}

func performRequest(router *gin.Engine, method string, path string, authorization string, body []byte) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	router.ServeHTTP(recorder, request)

	return recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, destination interface{}) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), destination); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; body = %s", err, recorder.Body.String())
	}
}

func assertUnauthorized(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	assertAPIError(t, recorder, http.StatusUnauthorized, "unauthorized")
	if recorder.Header().Get("WWW-Authenticate") != `Bearer realm="fluxcore"` {
		t.Fatalf("WWW-Authenticate = %q, want %q", recorder.Header().Get("WWW-Authenticate"), `Bearer realm="fluxcore"`)
	}
}

func assertAPIError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}

	var response struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeResponse(t, recorder, &response)
	if response.Error.Code != wantCode {
		t.Fatalf("error.code = %q, want %q", response.Error.Code, wantCode)
	}
	if response.Error.Message == "" {
		t.Fatal("error.message is empty")
	}
}

func testAuthorizationHeader() string {
	return "Bearer test-token"
}
