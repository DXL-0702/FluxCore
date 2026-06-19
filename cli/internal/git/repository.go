package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrNotRepository = errors.New("current directory is not inside a git repository")
	ErrRemoteMissing = errors.New("git remote origin is not configured")
	ErrGitNotFound   = errors.New("git executable not found in PATH")
)

type Repository struct {
	Root          string
	Name          string
	RemoteURL     string
	DefaultBranch string
	CurrentBranch string
}

type Inspector struct {
	workingDir string
	run        commandRunner
}

type commandRunner func(context.Context, string, ...string) (string, error)

func NewInspector(workingDir string) Inspector {
	return Inspector{
		workingDir: workingDir,
		run:        runGit,
	}
}

func NewInspectorWithRunner(workingDir string, runner commandRunner) Inspector {
	if runner == nil {
		runner = runGit
	}
	return Inspector{
		workingDir: workingDir,
		run:        runner,
	}
}

func (inspector Inspector) Inspect(ctx context.Context) (Repository, error) {
	root, err := inspector.RepositoryRoot(ctx)
	if err != nil {
		return Repository{}, err
	}

	remoteURL, err := inspector.RemoteURL(ctx, root)
	if err != nil {
		return Repository{}, err
	}

	currentBranch, err := inspector.CurrentBranch(ctx, root)
	if err != nil {
		return Repository{}, err
	}

	defaultBranch, err := inspector.DefaultBranch(ctx, root, currentBranch)
	if err != nil {
		return Repository{}, err
	}

	return Repository{
		Root:          root,
		Name:          filepath.Base(root),
		RemoteURL:     remoteURL,
		DefaultBranch: defaultBranch,
		CurrentBranch: currentBranch,
	}, nil
}

func (inspector Inspector) RepositoryRoot(ctx context.Context) (string, error) {
	root, err := inspector.run(ctx, inspector.workingDir, "rev-parse", "--show-toplevel")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("%w: %w", ErrGitNotFound, err)
		}
		return "", ErrNotRepository
	}

	root = strings.TrimSpace(root)
	if root == "" {
		return "", ErrNotRepository
	}
	return root, nil
}

func (inspector Inspector) RemoteURL(ctx context.Context, root string) (string, error) {
	if remoteURL, err := inspector.remoteURLForName(ctx, root, "origin"); err == nil {
		return remoteURL, nil
	}

	remoteNames, err := inspector.remoteNames(ctx, root)
	if err != nil {
		return "", err
	}
	for _, name := range remoteNames {
		if name == "origin" {
			continue
		}
		remoteURL, err := inspector.remoteURLForName(ctx, root, name)
		if err == nil {
			return remoteURL, nil
		}
	}

	return "", ErrRemoteMissing
}

func (inspector Inspector) CurrentBranch(ctx context.Context, root string) (string, error) {
	branch, err := inspector.run(ctx, root, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("read current branch: %w", err)
	}
	return strings.TrimSpace(branch), nil
}

func (inspector Inspector) DefaultBranch(ctx context.Context, root string, currentBranch string) (string, error) {
	remoteHead, err := inspector.run(ctx, root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	if err == nil {
		branch := strings.TrimSpace(remoteHead)
		if strings.HasPrefix(branch, "origin/") {
			branch = strings.TrimPrefix(branch, "origin/")
		}
		if branch != "" {
			return branch, nil
		}
	}

	if strings.TrimSpace(currentBranch) != "" {
		return currentBranch, nil
	}
	return "main", nil
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), message, err)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return stdout.String(), nil
}

func (inspector Inspector) remoteURLForName(ctx context.Context, root string, name string) (string, error) {
	remoteURL, err := inspector.run(ctx, root, "config", "--get", "remote."+name+".url")
	if err != nil {
		return "", ErrRemoteMissing
	}

	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return "", ErrRemoteMissing
	}
	return remoteURL, nil
}

func (inspector Inspector) remoteNames(ctx context.Context, root string) ([]string, error) {
	rawRemotes, err := inspector.run(ctx, root, "remote")
	if err != nil {
		return nil, ErrRemoteMissing
	}

	names := make([]string, 0)
	for _, line := range strings.Split(rawRemotes, "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, ErrRemoteMissing
	}
	return names, nil
}
