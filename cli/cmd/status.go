package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	localconfig "github.com/jaxson/FluxCore/cli/internal/config"
	localgit "github.com/jaxson/FluxCore/cli/internal/git"
	"github.com/spf13/cobra"
)

func newStatusCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show FluxCore binding status for the current Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), commandTimeout)
			defer cancel()

			workingDir, err := options.workingDir()
			if err != nil {
				return fmt.Errorf("read working directory: %w", err)
			}

			inspector := localgit.NewInspector(workingDir)
			root, err := inspector.RepositoryRoot(ctx)
			if err != nil {
				return err
			}

			cfg, err := localconfig.NewStore(root).Load()
			if err != nil {
				if errors.Is(err, localconfig.ErrConfigNotFound) {
					fmt.Fprintln(cmd.OutOrStdout(), "FluxCore: not initialized")
					fmt.Fprintf(cmd.OutOrStdout(), "Repository: %s\n", root)
					fmt.Fprintln(cmd.OutOrStdout(), "Next step: fluxcore init")
					return nil
				}
				return err
			}

			currentBranch, _ := inspector.CurrentBranch(ctx, root)

			if cfg.IsLinked() {
				fmt.Fprintln(cmd.OutOrStdout(), "FluxCore: linked")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "FluxCore: initialized, not linked")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Server: %s\n", cfg.ServerURL)
			fmt.Fprintf(cmd.OutOrStdout(), "Repository: %s\n", root)
			if strings.TrimSpace(currentBranch) != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Branch: %s\n", currentBranch)
			}
			if cfg.Project.ID != 0 || cfg.Project.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%d)\n", emptyFallback(cfg.Project.Name, "unknown"), cfg.Project.ID)
			}
			if cfg.Repository.ID != 0 || cfg.Repository.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Bound repository: %s (%d)\n", emptyFallback(cfg.Repository.Name, "unknown"), cfg.Repository.ID)
			}
			if cfg.Repository.RemoteURL != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Remote: %s\n", cfg.Repository.RemoteURL)
			}
			if cfg.Repository.DefaultBranch != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Default branch: %s\n", cfg.Repository.DefaultBranch)
			}
			if cfg.Token != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Token: configured")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Token: not configured")
			}
			return nil
		},
	}
}

func emptyFallback(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
