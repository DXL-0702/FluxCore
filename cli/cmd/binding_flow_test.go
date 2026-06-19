package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	localconfig "github.com/jaxson/FluxCore/cli/internal/config"
)

func TestInitLinkStatusFlow(t *testing.T) {
	repositoryRoot := newTestGitRepository(t)
	server := newBindingTestServer(t, repositoryRoot)
	defer server.Close()

	withWorkingDir(t, repositoryRoot, func() {
		output, err := executeForTest("init", "--server", server.URL, "--token", "secret-token")
		if err != nil {
			t.Fatalf("init error = %v\noutput:\n%s", err, output)
		}
		for _, expected := range []string{
			"Initialized FluxCore config",
			"Token: configured",
		} {
			if !strings.Contains(output, expected) {
				t.Fatalf("init output does not contain %q\noutput:\n%s", expected, output)
			}
		}
		if strings.Contains(output, "secret-token") {
			t.Fatalf("init output leaked token\noutput:\n%s", output)
		}

		output, err = executeForTest("link", "--project", "FluxCore")
		if err != nil {
			t.Fatalf("link error = %v\noutput:\n%s", err, output)
		}
		for _, expected := range []string{
			"Linked repository",
			"Project ID: 7",
			"Repository ID: 11",
		} {
			if !strings.Contains(output, expected) {
				t.Fatalf("link output does not contain %q\noutput:\n%s", expected, output)
			}
		}
		if strings.Contains(output, "secret-token") {
			t.Fatalf("link output leaked token\noutput:\n%s", output)
		}

		output, err = executeForTest("status")
		if err != nil {
			t.Fatalf("status error = %v\noutput:\n%s", err, output)
		}
		for _, expected := range []string{
			"FluxCore: linked",
			"Server: " + server.URL,
			"Project: FluxCore (7)",
			"Bound repository: " + filepath.Base(repositoryRoot) + " (11)",
			"Remote: git@example.com:DXL-0702/FluxCore.git",
			"Token: configured",
		} {
			if !strings.Contains(output, expected) {
				t.Fatalf("status output does not contain %q\noutput:\n%s", expected, output)
			}
		}
		if strings.Contains(output, "secret-token") {
			t.Fatalf("status output leaked token\noutput:\n%s", output)
		}
	})

	cfg := readTestConfig(t, repositoryRoot)
	if cfg.Project.ID != 7 || cfg.Project.Name != "FluxCore" {
		t.Fatalf("project config = %#v", cfg.Project)
	}
	if cfg.Repository.ID != 11 || cfg.Repository.LocalPath != repositoryRoot {
		t.Fatalf("repository config = %#v", cfg.Repository)
	}

	gitignore, err := os.ReadFile(filepath.Join(repositoryRoot, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if strings.Count(string(gitignore), ".fluxcore/") != 1 {
		t.Fatalf(".gitignore = %q", string(gitignore))
	}
}

func TestInitFailsOutsideGitRepository(t *testing.T) {
	withWorkingDir(t, t.TempDir(), func() {
		_, err := executeForTest("init")
		if err == nil {
			t.Fatal("init error = nil, want error")
		}
		if !strings.Contains(err.Error(), "current directory is not inside a git repository") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestStatusReportsUninitializedRepository(t *testing.T) {
	repositoryRoot := newTestGitRepository(t)

	withWorkingDir(t, repositoryRoot, func() {
		output, err := executeForTest("status")
		if err != nil {
			t.Fatalf("status error = %v\noutput:\n%s", err, output)
		}
		for _, expected := range []string{
			"FluxCore: not initialized",
			"Next step: fluxcore init",
		} {
			if !strings.Contains(output, expected) {
				t.Fatalf("status output does not contain %q\noutput:\n%s", expected, output)
			}
		}
	})
}

func TestLinkRequiresInitializedConfig(t *testing.T) {
	repositoryRoot := newTestGitRepository(t)

	withWorkingDir(t, repositoryRoot, func() {
		_, err := executeForTest("link", "--project", "FluxCore")
		if err == nil {
			t.Fatal("link error = nil, want error")
		}
		if !strings.Contains(err.Error(), "run fluxcore init first") {
			t.Fatalf("error = %v", err)
		}
	})
}

func newBindingTestServer(t *testing.T, repositoryRoot string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Fatalf("Authorization = %q", got)
		}
		writer.Header().Set("Content-Type", "application/json")

		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/projects":
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode project request: %v", err)
			}
			if body["name"] != "FluxCore" {
				t.Fatalf("project name = %q", body["name"])
			}
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"project":{"id":7,"name":"FluxCore","status":"active"}}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/projects/7/repositories":
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode repository request: %v", err)
			}
			if body["name"] != filepath.Base(repositoryRoot) {
				t.Fatalf("repository name = %q", body["name"])
			}
			if body["local_path"] != repositoryRoot {
				t.Fatalf("local_path = %q, want %q", body["local_path"], repositoryRoot)
			}
			if body["remote_url"] != "git@example.com:DXL-0702/FluxCore.git" {
				t.Fatalf("remote_url = %q", body["remote_url"])
			}
			writer.WriteHeader(http.StatusCreated)
			response := map[string]interface{}{
				"repository": map[string]interface{}{
					"id":             11,
					"project_id":     7,
					"name":           filepath.Base(repositoryRoot),
					"local_path":     repositoryRoot,
					"remote_url":     "git@example.com:DXL-0702/FluxCore.git",
					"default_branch": body["default_branch"],
				},
			}
			if err := json.NewEncoder(writer).Encode(response); err != nil {
				t.Fatalf("encode repository response: %v", err)
			}
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
}

func newTestGitRepository(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval temp repository path: %v", err)
	}
	runTestGit(t, root, "init", "-b", "main")
	runTestGit(t, root, "remote", "add", "origin", "git@example.com:DXL-0702/FluxCore.git")
	runTestGit(t, root, "config", "user.email", "test@example.com")
	runTestGit(t, root, "config", "user.name", "FluxCore Test")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Test\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runTestGit(t, root, "add", "README.md")
	runTestGit(t, root, "commit", "-m", "initial commit")
	return root
}

func runTestGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}()

	fn()
}

func readTestConfig(t *testing.T, root string) localconfig.Config {
	t.Helper()

	cfg, err := localconfig.NewStore(root).Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return cfg
}
