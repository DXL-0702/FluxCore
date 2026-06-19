package cmd

import (
	"strings"
	"testing"
)

func TestRootHelpIncludesGlobalFlags(t *testing.T) {
	t.Setenv(envServerURL, "")
	t.Setenv(envAPIToken, "")

	output, err := executeForTest("--help")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, expected := range []string{
		"FluxCore connects local Git repositories to the FluxCore backend.",
		"--server",
		"--token",
		envServerURL,
		envAPIToken,
		defaultServerURL,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("help output does not contain %q\noutput:\n%s", expected, output)
		}
	}
}

func TestRootCommandRunsHelpByDefault(t *testing.T) {
	output, err := executeForTest()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("output does not contain usage\noutput:\n%s", output)
	}
}

func TestRootFlagsPopulateOptions(t *testing.T) {
	options := &rootOptions{}
	command := newRootCommandWithOptions(options)
	output := newTestOutput()
	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs([]string{
		"--server", "http://127.0.0.1:9090",
		"--token", "test-token",
		"--help",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if options.server != "http://127.0.0.1:9090" {
		t.Fatalf("server = %q, want %q", options.server, "http://127.0.0.1:9090")
	}
	if options.token != "test-token" {
		t.Fatalf("token = %q, want %q", options.token, "test-token")
	}
}

func TestRootUsesEnvironmentDefaults(t *testing.T) {
	t.Setenv(envServerURL, "http://127.0.0.1:9090")
	t.Setenv(envAPIToken, "env-token")

	options := &rootOptions{}
	output, err := executeForTestWithOptions(options)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if options.server != "http://127.0.0.1:9090" {
		t.Fatalf("server = %q, want %q", options.server, "http://127.0.0.1:9090")
	}
	if options.token != "env-token" {
		t.Fatalf("token = %q, want %q", options.token, "env-token")
	}
	if strings.Contains(output, "env-token") {
		t.Fatalf("help output leaked token\noutput:\n%s", output)
	}
}

func TestRootHelpDoesNotLeakEnvironmentToken(t *testing.T) {
	t.Setenv(envAPIToken, "env-secret-token")

	output, err := executeForTest("--help")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if strings.Contains(output, "env-secret-token") {
		t.Fatalf("help output leaked token\noutput:\n%s", output)
	}
}
