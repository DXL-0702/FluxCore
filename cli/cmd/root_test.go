package cmd

import (
	"strings"
	"testing"
)

func TestRootHelpIncludesGlobalFlags(t *testing.T) {
	output, err := executeForTest("--help")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, expected := range []string{
		"FluxCore connects local Git repositories to the FluxCore backend.",
		"--server",
		"--token",
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
