package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultServerURL = "http://127.0.0.1:8080"
	envServerURL     = "FLUXCORE_SERVER"
	envAPIToken      = "FLUXCORE_TOKEN"
)

type rootOptions struct {
	server     string
	token      string
	workingDir func() (string, error)
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	return newRootCommandWithOptions(&rootOptions{})
}

func newRootCommandWithOptions(options *rootOptions) *cobra.Command {
	if options == nil {
		options = &rootOptions{}
	}
	applyRootStaticDefaults(options)

	rootCmd := &cobra.Command{
		Use:           "fluxcore",
		Short:         "FluxCore local development context CLI",
		Long:          "FluxCore connects local Git repositories to the FluxCore backend.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			applyRootRuntimeDefaults(cmd, options)
		},
	}

	rootCmd.PersistentFlags().StringVar(&options.server, "server", options.server, "FluxCore server base URL (env "+envServerURL+")")
	rootCmd.PersistentFlags().Var(&secretFlagValue{target: &options.token}, "token", "FluxCore API token (env "+envAPIToken+")")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(newInitCommand(options))
	rootCmd.AddCommand(newLinkCommand(options))
	rootCmd.AddCommand(newStatusCommand(options))

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
	})

	return rootCmd
}

func applyRootStaticDefaults(options *rootOptions) {
	if options.workingDir == nil {
		options.workingDir = os.Getwd
	}
	if strings.TrimSpace(options.server) == "" {
		options.server = envOrDefault(envServerURL, defaultServerURL)
	}
}

func applyRootRuntimeDefaults(cmd *cobra.Command, options *rootOptions) {
	if strings.TrimSpace(options.server) == "" {
		options.server = envOrDefault(envServerURL, defaultServerURL)
	}
	if !flagChanged(cmd, "token") && strings.TrimSpace(options.token) == "" {
		options.token = strings.TrimSpace(os.Getenv(envAPIToken))
	}
}

func envOrDefault(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

type secretFlagValue struct {
	target *string
}

func (value *secretFlagValue) Set(input string) error {
	if value.target != nil {
		*value.target = input
	}
	return nil
}

func (value *secretFlagValue) String() string {
	return ""
}

func (value *secretFlagValue) Type() string {
	return "string"
}

func executeForTest(args ...string) (string, error) {
	return executeForTestWithOptions(&rootOptions{}, args...)
}

func executeForTestInDir(workingDir string, args ...string) (string, error) {
	return executeForTestWithOptions(&rootOptions{
		workingDir: func() (string, error) {
			return workingDir, nil
		},
	}, args...)
}

func executeForTestWithOptions(options *rootOptions, args ...string) (string, error) {
	command := newRootCommandWithOptions(options)
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
