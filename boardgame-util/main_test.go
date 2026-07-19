package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMainImplReturnsSuccessForHelp(t *testing.T) {
	stdout, _ := captureCommandOutput(t)
	if code := mainImpl([]string{"boardgame-util", "help"}); code != 0 {
		t.Fatalf("mainImpl(help) = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("help output did not contain usage: %q", stdout.String())
	}
}

func TestMainImplReturnsFailureForMissingCommand(t *testing.T) {
	_, stderr := captureCommandOutput(t)
	if code := mainImpl([]string{"boardgame-util"}); code != 1 {
		t.Fatalf("mainImpl(no command) = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "COMMAND is required") {
		t.Fatalf("error output did not explain missing command: %q", stderr.String())
	}
}

func TestMainImplReturnsFailureForInvalidOption(t *testing.T) {
	_, stderr := captureCommandOutput(t)
	if code := mainImpl([]string{"boardgame-util", "--definitely-invalid"}); code != 1 {
		t.Fatalf("mainImpl(invalid option) = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "definitely-invalid") {
		t.Fatalf("error output did not identify invalid option: %q", stderr.String())
	}
}

func captureCommandOutput(t *testing.T) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	oldStdout, oldStderr := commandStdout, commandStderr
	commandStdout, commandStderr = stdout, stderr
	t.Cleanup(func() {
		commandStdout, commandStderr = oldStdout, oldStderr
	})
	return stdout, stderr
}
