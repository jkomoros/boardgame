package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallGeneratedMoveArgsCheckNeverWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "_move_args.ts")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	err := installGeneratedMoveArgs([]generatedMoveArgsFile{{path: path, contents: []byte("new"), gameName: "test"}}, true)
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("error = %v, want stale diagnostic", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "old" {
		t.Fatalf("--check wrote destination: %q", got)
	}
}

func TestInstallGeneratedMoveArgsPreparationFailureLeavesAllOldFiles(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a", "_move_args.ts")
	if err := os.Mkdir(filepath.Dir(first), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(first, []byte("old-a"), 0644); err != nil {
		t.Fatal(err)
	}
	invalidParent := filepath.Join(dir, "z-not-a-directory")
	if err := os.WriteFile(invalidParent, []byte("sentinel"), 0644); err != nil {
		t.Fatal(err)
	}
	second := filepath.Join(invalidParent, "_move_args.ts")

	err := installGeneratedMoveArgs([]generatedMoveArgsFile{
		{path: first, contents: []byte("new-a"), gameName: "a"},
		{path: second, contents: []byte("new-z"), gameName: "z"},
	}, false)
	if err == nil {
		t.Fatal("install succeeded despite invalid destination parent")
	}
	got, readErr := os.ReadFile(first)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "old-a" {
		t.Fatalf("first destination changed after later preparation failure: %q", got)
	}
	matches, globErr := filepath.Glob(filepath.Join(filepath.Dir(first), ".move-args-*"))
	if globErr != nil {
		t.Fatal(globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("staged files leaked: %v", matches)
	}
}

func TestInstallGeneratedMoveArgsReplacesPreparedFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "_move_args.ts")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installGeneratedMoveArgs([]generatedMoveArgsFile{{path: path, contents: []byte("new"), gameName: "test"}}, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("destination = %q, want new", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Fatalf("permissions = %o, want 0644", info.Mode().Perm())
	}
}

func TestValidateGeneratedMoveArgsTypeScriptReportsCompilerFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	dir := t.TempDir()
	compiler := filepath.Join(dir, "failing-tsc")
	if err := os.WriteFile(compiler, []byte("#!/bin/sh\necho deliberate compiler failure >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BOARDGAME_TSC", compiler)
	err := validateGeneratedMoveArgsTypeScript([]generatedMoveArgsFile{{
		contents: []byte("export type Valid = string;\n"),
		gameName: "test",
	}})
	if err == nil || !strings.Contains(err.Error(), "deliberate compiler failure") {
		t.Fatalf("error = %v, want compiler output", err)
	}
}

func TestMoveArgsTypeScriptCompilerPrefersCreatorWorkspace(t *testing.T) {
	dir := t.TempDir()
	client := filepath.Join(dir, "client")
	compilerName := "tsc"
	if runtime.GOOS == "windows" {
		compilerName = "tsc.cmd"
	}
	compiler := filepath.Join(dir, "node_modules", ".bin", compilerName)
	if err := os.MkdirAll(filepath.Dir(compiler), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(compiler, []byte("workspace compiler"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BOARDGAME_TSC", "")
	got, err := moveArgsTypeScriptCompiler([]generatedMoveArgsFile{{path: filepath.Join(client, "_move_args.ts")}})
	if err != nil {
		t.Fatal(err)
	}
	if got != compiler {
		t.Fatalf("compiler = %q, want creator-local %q", got, compiler)
	}
}
