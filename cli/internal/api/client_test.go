package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateProjectSendsAuthenticatedJSONRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s", request.Method)
			http.Error(writer, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if request.URL.Path != "/api/projects" {
			t.Errorf("path = %s", request.URL.Path)
			http.NotFound(writer, request)
			return
		}
		if got := request.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Errorf("Authorization = %q", got)
			http.Error(writer, "unexpected authorization header", http.StatusUnauthorized)
			return
		}
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
			http.Error(writer, "unexpected content type", http.StatusUnsupportedMediaType)
			return
		}

		var body map[string]string
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
			http.Error(writer, "invalid request body", http.StatusBadRequest)
			return
		}
		if body["name"] != "FluxCore" {
			t.Errorf("name = %q", body["name"])
			http.Error(writer, "unexpected project name", http.StatusBadRequest)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"project":{"id":7,"name":"FluxCore","status":"active"}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret-token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	project, err := client.CreateProject(context.Background(), CreateProjectInput{Name: "FluxCore"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if project.ID != 7 || project.Name != "FluxCore" {
		t.Fatalf("project = %#v", project)
	}
}

func TestCreateOrGetProjectUsesListOnConflict(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCount++
		writer.Header().Set("Content-Type", "application/json")

		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/projects":
			writer.WriteHeader(http.StatusConflict)
			_, _ = writer.Write([]byte(`{"error":{"code":"conflict","message":"project name already exists"}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/projects":
			_, _ = writer.Write([]byte(`{"projects":[{"id":9,"name":"FluxCore","status":"active"}]}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			http.NotFound(writer, request)
			return
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret-token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	project, err := client.CreateOrGetProject(context.Background(), CreateProjectInput{Name: "FluxCore"})
	if err != nil {
		t.Fatalf("CreateOrGetProject() error = %v", err)
	}
	if project.ID != 9 {
		t.Fatalf("project ID = %d, want 9", project.ID)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
}

func TestCreateOrGetProjectReturnsCreateConflictWhenListMisses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/projects":
			writer.WriteHeader(http.StatusConflict)
			_, _ = writer.Write([]byte(`{"error":{"code":"conflict","message":"project name already exists"}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/projects":
			_, _ = writer.Write([]byte(`{"projects":[]}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			http.NotFound(writer, request)
			return
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret-token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.CreateOrGetProject(context.Background(), CreateProjectInput{Name: "FluxCore"})
	if err == nil {
		t.Fatal("CreateOrGetProject() error = nil, want error")
	}
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("CreateOrGetProject() error = %v, want ErrConflict", err)
	}
	if !strings.Contains(err.Error(), "project name already exists") {
		t.Fatalf("error = %q, want original conflict message", err.Error())
	}
	if strings.Contains(err.Error(), "could not be found") {
		t.Fatalf("error = %q, want original conflict error", err.Error())
	}
}

func TestCreateOrGetRepositoryReusesSameRemoteURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/projects/9/repositories":
			writer.WriteHeader(http.StatusConflict)
			_, _ = writer.Write([]byte(`{"error":{"code":"conflict","message":"repository conflicts"}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/projects/9/repositories":
			_, _ = writer.Write([]byte(`{"repositories":[{"id":11,"project_id":9,"name":"FluxCore","local_path":"/old/repo","remote_url":"git@example.com:repo.git","default_branch":"main"}]}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			http.NotFound(writer, request)
			return
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret-token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	repository, err := client.CreateOrGetRepository(context.Background(), 9, CreateRepositoryInput{
		Name:          "FluxCore",
		LocalPath:     "/repo",
		RemoteURL:     "git@example.com:repo.git",
		DefaultBranch: "main",
	})
	if err != nil {
		t.Fatalf("CreateOrGetRepository() error = %v", err)
	}
	if repository.ID != 11 {
		t.Fatalf("repository ID = %d, want 11", repository.ID)
	}
}

func TestCreateOrGetRepositoryDoesNotReuseDifferentRemoteURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/projects/9/repositories":
			writer.WriteHeader(http.StatusConflict)
			_, _ = writer.Write([]byte(`{"error":{"code":"conflict","message":"repository local_path already exists"}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/projects/9/repositories":
			_, _ = writer.Write([]byte(`{"repositories":[{"id":11,"project_id":9,"name":"FluxCore","local_path":"/repo","remote_url":"git@example.com:other.git","default_branch":"main"}]}`))
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			http.NotFound(writer, request)
			return
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret-token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.CreateOrGetRepository(context.Background(), 9, CreateRepositoryInput{
		Name:          "FluxCore",
		LocalPath:     "/repo",
		RemoteURL:     "git@example.com:repo.git",
		DefaultBranch: "main",
	})
	if err == nil {
		t.Fatal("CreateOrGetRepository() error = nil, want error")
	}
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("CreateOrGetRepository() error = %v, want ErrConflict", err)
	}
	if !strings.Contains(err.Error(), "repository local_path already exists") {
		t.Fatalf("error = %q, want original conflict message", err.Error())
	}
	if strings.Contains(err.Error(), "could not be found") {
		t.Fatalf("error = %q, want original conflict error", err.Error())
	}
}

func TestParseErrorWrapsConflict(t *testing.T) {
	err := parseError(http.StatusConflict, []byte(`{"error":{"code":"conflict","message":"exists"}}`))
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("parseError() = %v, want ErrConflict", err)
	}
}

func TestNewClientValidatesServerURLAndToken(t *testing.T) {
	for _, tc := range []struct {
		name    string
		url     string
		token   string
		wantErr string
	}{
		{name: "empty url", url: "", token: "token", wantErr: "server URL is required"},
		{name: "unsupported scheme", url: "ftp://example.com", token: "token", wantErr: "server URL must use http or https"},
		{name: "missing host", url: "http:///api", token: "token", wantErr: "server URL host is required"},
		{name: "empty token", url: "http://127.0.0.1:8080", token: "", wantErr: "API token is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.url, tc.token)
			if err == nil {
				t.Fatal("NewClient() error = nil")
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}
