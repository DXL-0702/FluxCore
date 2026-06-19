package git

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestInspectorInspect(t *testing.T) {
	calls := make([][]string, 0)
	runner := func(ctx context.Context, dir string, args ...string) (string, error) {
		calls = append(calls, append([]string{dir}, args...))
		switch strings.Join(args, " ") {
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "config --get remote.origin.url":
			return "git@github.com:DXL-0702/FluxCore.git\n", nil
		case "branch --show-current":
			return "develop\n", nil
		case "symbolic-ref --quiet --short refs/remotes/origin/HEAD":
			return "origin/main\n", nil
		default:
			return "", errors.New("unexpected git command")
		}
	}

	repository, err := NewInspectorWithRunner("/repo/subdir", runner).Inspect(context.Background())
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	if repository.Root != "/repo" {
		t.Fatalf("Root = %q", repository.Root)
	}
	if repository.Name != "repo" {
		t.Fatalf("Name = %q", repository.Name)
	}
	if repository.RemoteURL != "git@github.com:DXL-0702/FluxCore.git" {
		t.Fatalf("RemoteURL = %q", repository.RemoteURL)
	}
	if repository.CurrentBranch != "develop" {
		t.Fatalf("CurrentBranch = %q", repository.CurrentBranch)
	}
	if repository.DefaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q", repository.DefaultBranch)
	}

	expected := [][]string{
		{"/repo/subdir", "rev-parse", "--show-toplevel"},
		{"/repo", "config", "--get", "remote.origin.url"},
		{"/repo", "branch", "--show-current"},
		{"/repo", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"},
	}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("calls = %#v, want %#v", calls, expected)
	}
}

func TestRepositoryRootMapsGitFailureToNotRepository(t *testing.T) {
	underlying := errors.New("not a git repository")
	runner := func(ctx context.Context, dir string, args ...string) (string, error) {
		return "", underlying
	}

	_, err := NewInspectorWithRunner("/tmp", runner).RepositoryRoot(context.Background())
	if !errors.Is(err, ErrNotRepository) {
		t.Fatalf("RepositoryRoot() error = %v, want %v", err, ErrNotRepository)
	}
	if !errors.Is(err, underlying) {
		t.Fatalf("RepositoryRoot() error = %v, want underlying error", err)
	}
}

func TestRepositoryRootReportsMissingGitExecutable(t *testing.T) {
	runner := func(ctx context.Context, dir string, args ...string) (string, error) {
		return "", exec.ErrNotFound
	}

	_, err := NewInspectorWithRunner("/tmp", runner).RepositoryRoot(context.Background())
	if !errors.Is(err, ErrGitNotFound) {
		t.Fatalf("RepositoryRoot() error = %v, want %v", err, ErrGitNotFound)
	}
	if errors.Is(err, ErrNotRepository) {
		t.Fatalf("RepositoryRoot() error = %v, should not be %v", err, ErrNotRepository)
	}
}

func TestRepositoryRootPreservesContextErrors(t *testing.T) {
	for _, expectedErr := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(expectedErr.Error(), func(t *testing.T) {
			runner := func(ctx context.Context, dir string, args ...string) (string, error) {
				return "", expectedErr
			}

			_, err := NewInspectorWithRunner("/tmp", runner).RepositoryRoot(context.Background())
			if !errors.Is(err, expectedErr) {
				t.Fatalf("RepositoryRoot() error = %v, want %v", err, expectedErr)
			}
			if errors.Is(err, ErrNotRepository) {
				t.Fatalf("RepositoryRoot() error = %v, should not be %v", err, ErrNotRepository)
			}
		})
	}
}

func TestRemoteURLFallsBackToFirstConfiguredRemote(t *testing.T) {
	calls := make([][]string, 0)
	runner := func(ctx context.Context, dir string, args ...string) (string, error) {
		calls = append(calls, append([]string{dir}, args...))
		switch strings.Join(args, " ") {
		case "config --get remote.origin.url":
			return "", errors.New("origin missing")
		case "remote":
			return "upstream\ngithub\n", nil
		case "config --get remote.upstream.url":
			return "git@example.com:upstream/repo.git\n", nil
		default:
			return "", errors.New("unexpected git command")
		}
	}

	remoteURL, err := NewInspectorWithRunner("/repo", runner).RemoteURL(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("RemoteURL() error = %v", err)
	}
	if remoteURL != "git@example.com:upstream/repo.git" {
		t.Fatalf("RemoteURL() = %q", remoteURL)
	}

	expected := [][]string{
		{"/repo", "config", "--get", "remote.origin.url"},
		{"/repo", "remote"},
		{"/repo", "config", "--get", "remote.upstream.url"},
	}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("calls = %#v, want %#v", calls, expected)
	}
}

func TestRemoteURLRequiresConfiguredRemote(t *testing.T) {
	underlying := errors.New("missing remote")
	runner := func(ctx context.Context, dir string, args ...string) (string, error) {
		return "", underlying
	}

	_, err := NewInspectorWithRunner("/repo", runner).RemoteURL(context.Background(), "/repo")
	if !errors.Is(err, ErrRemoteMissing) {
		t.Fatalf("RemoteURL() error = %v, want %v", err, ErrRemoteMissing)
	}
	if !errors.Is(err, underlying) {
		t.Fatalf("RemoteURL() error = %v, want underlying error", err)
	}
}

func TestRemoteURLPreservesContextErrors(t *testing.T) {
	for _, expectedErr := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(expectedErr.Error(), func(t *testing.T) {
			runner := func(ctx context.Context, dir string, args ...string) (string, error) {
				return "", expectedErr
			}

			_, err := NewInspectorWithRunner("/repo", runner).RemoteURL(context.Background(), "/repo")
			if !errors.Is(err, expectedErr) {
				t.Fatalf("RemoteURL() error = %v, want %v", err, expectedErr)
			}
			if errors.Is(err, ErrRemoteMissing) {
				t.Fatalf("RemoteURL() error = %v, should not be %v", err, ErrRemoteMissing)
			}
		})
	}
}

func TestDefaultBranchFallsBackToCurrentBranch(t *testing.T) {
	runner := func(ctx context.Context, dir string, args ...string) (string, error) {
		return "", errors.New("origin head missing")
	}

	branch, err := NewInspectorWithRunner("/repo", runner).DefaultBranch(context.Background(), "/repo", "develop")
	if err != nil {
		t.Fatalf("DefaultBranch() error = %v", err)
	}
	if branch != "develop" {
		t.Fatalf("branch = %q, want develop", branch)
	}
}
