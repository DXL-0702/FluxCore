package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jaxson/FluxCore/cli/internal/api"
	localconfig "github.com/jaxson/FluxCore/cli/internal/config"
	localgit "github.com/jaxson/FluxCore/cli/internal/git"
	"github.com/spf13/cobra"
)

type linkOptions struct {
	project string
}

func newLinkCommand(root *rootOptions) *cobra.Command {
	options := &linkOptions{}
	command := &cobra.Command{
		Use:   "link",
		Short: "Bind the current Git repository to a FluxCore project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLink(cmd, root, options)
		},
	}

	command.Flags().StringVar(&options.project, "project", "", "FluxCore project name")
	if err := command.MarkFlagRequired("project"); err != nil {
		panic(err)
	}

	return command
}

func runLink(cmd *cobra.Command, rootOptions *rootOptions, options *linkOptions) error {
	projectName := strings.TrimSpace(options.project)
	if projectName == "" {
		return fmt.Errorf("project name is required")
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), commandTimeout)
	defer cancel()

	workingDir, err := rootOptions.workingDir()
	if err != nil {
		return fmt.Errorf("read working directory: %w", err)
	}

	repository, err := localgit.NewInspector(workingDir).Inspect(ctx)
	if err != nil {
		return err
	}

	store := localconfig.NewStore(repository.Root)
	cfg, err := store.Load()
	if err != nil {
		if errors.Is(err, localconfig.ErrConfigNotFound) {
			return fmt.Errorf("fluxcore config not found; run fluxcore init first")
		}
		return err
	}

	serverURL := cfg.ServerURL
	if flagChanged(cmd, "server") || strings.TrimSpace(serverURL) == "" {
		serverURL = rootOptions.server
	}
	token := cfg.Token
	if flagChanged(cmd, "token") {
		token = rootOptions.token
	}

	client, err := api.NewClient(serverURL, token)
	if err != nil {
		return err
	}

	project, err := client.CreateOrGetProject(ctx, api.CreateProjectInput{Name: projectName})
	if err != nil {
		return err
	}

	linkedRepository, err := client.CreateOrGetRepository(ctx, project.ID, api.CreateRepositoryInput{
		Name:          repository.Name,
		LocalPath:     repository.Root,
		RemoteURL:     repository.RemoteURL,
		DefaultBranch: repository.DefaultBranch,
	})
	if err != nil {
		return err
	}

	cfg.ServerURL = serverURL
	cfg.Token = token
	cfg.Project = localconfig.Project{
		ID:   project.ID,
		Name: project.Name,
	}
	cfg.Repository = localconfig.Repository{
		ID:            linkedRepository.ID,
		Name:          linkedRepository.Name,
		LocalPath:     repository.Root,
		RemoteURL:     repository.RemoteURL,
		DefaultBranch: linkedRepository.DefaultBranch,
	}
	if err := store.Save(cfg); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Linked repository %s to project %s\n", cfg.Repository.Name, cfg.Project.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Repository ID: %d\n", cfg.Repository.ID)
	fmt.Fprintf(cmd.OutOrStdout(), "Project ID: %d\n", cfg.Project.ID)
	return nil
}

func flagChanged(cmd *cobra.Command, name string) bool {
	flag := cmd.Flag(name)
	return flag != nil && flag.Changed
}
