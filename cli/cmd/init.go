package cmd

import (
	"context"
	"fmt"
	"time"

	localconfig "github.com/jaxson/FluxCore/cli/internal/config"
	localgit "github.com/jaxson/FluxCore/cli/internal/git"
	"github.com/spf13/cobra"
)

const commandTimeout = 10 * time.Second

func newInitCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize FluxCore configuration in the current Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), commandTimeout)
			defer cancel()

			workingDir, err := options.workingDir()
			if err != nil {
				return fmt.Errorf("read working directory: %w", err)
			}

			root, err := localgit.NewInspector(workingDir).RepositoryRoot(ctx)
			if err != nil {
				return err
			}

			cfg, err := localconfig.NewStore(root).Init(localconfig.InitOptions{
				ServerURL:    options.server,
				Token:        options.token,
				UpdateServer: flagChanged(cmd, "server"),
				UpdateToken:  flagChanged(cmd, "token"),
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized FluxCore config at %s/%s\n", localconfig.DirectoryName, localconfig.FileName)
			fmt.Fprintf(cmd.OutOrStdout(), "Server: %s\n", cfg.ServerURL)
			if cfg.Token != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Token: configured")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Token: not configured")
			}
			return nil
		},
	}
}
