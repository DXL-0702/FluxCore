package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

const defaultServerURL = "http://127.0.0.1:8080"

type rootOptions struct {
	server string
	token  string
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	return newRootCommandWithOptions(&rootOptions{})
}

func newRootCommandWithOptions(options *rootOptions) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "fluxcore",
		Short:         "FluxCore local development context CLI",
		Long:          "FluxCore connects local Git repositories to the FluxCore backend.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVar(&options.server, "server", defaultServerURL, "FluxCore server base URL")
	rootCmd.PersistentFlags().StringVar(&options.token, "token", "", "FluxCore API token")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(newInitCommand(options))
	rootCmd.AddCommand(newLinkCommand(options))
	rootCmd.AddCommand(newStatusCommand())

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
	})

	return rootCmd
}

func executeForTest(args ...string) (string, error) {
	command := newRootCommand()
	output := newTestOutput()
	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs(args)

	err := command.Execute()
	return output.String(), err
}

type testOutput struct {
	data []byte
}

func newTestOutput() *testOutput {
	return &testOutput{}
}

func (out *testOutput) Write(data []byte) (int, error) {
	out.data = append(out.data, data...)
	return len(data), nil
}

func (out *testOutput) String() string {
	return string(out.data)
}

var _ io.Writer = (*testOutput)(nil)
