package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientContractSetCheckReportsCompleteDeterministicStaleSet(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "current.ts")
	stale := filepath.Join(dir, "stale.ts")
	orphan := filepath.Join(dir, "orphan.ts")
	for path, contents := range map[string]string{current: "current", stale: "old", orphan: "orphan"} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	set := &generatedClientContractSet{
		replacements: []generatedClientContract{
			{path: stale, contents: []byte("new")},
			{path: current, contents: []byte("current")},
		},
		deletions: []string{orphan},
	}
	err := set.check()
	if err == nil {
		t.Fatal("check succeeded for stale contracts")
	}
	wantOrder := stale + ", " + orphan
	if stale > orphan {
		wantOrder = orphan + ", " + stale
	}
	if !strings.Contains(err.Error(), wantOrder) {
		t.Fatalf("error = %q, want sorted stale paths %q", err, wantOrder)
	}
	contents, readErr := os.ReadFile(orphan)
	if readErr != nil || string(contents) != "orphan" {
		t.Fatalf("check mutated orphan: contents=%q err=%v", contents, readErr)
	}
}

func TestClientContractSetStagingFailureMutatesNothing(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "a-existing.ts")
	orphan := filepath.Join(dir, "b-orphan.ts")
	for path, contents := range map[string]string{existing: "old", orphan: "orphan"} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	set := &generatedClientContractSet{
		replacements: []generatedClientContract{
			{path: existing, contents: []byte("new")},
			{path: filepath.Join(orphan, "z.ts"), contents: []byte("new")},
		},
		deletions: []string{orphan},
	}
	if err := set.install(); err == nil {
		t.Fatal("install succeeded despite missing staging directory")
	}
	for path, want := range map[string]string{existing: "old", orphan: "orphan"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s after staging failure = %q, %v; want %q", path, got, err, want)
		}
	}
}

func TestClientContractSetInstallsCompleteReplacementAndDeletion(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing.ts")
	created := filepath.Join(dir, "created.ts")
	orphan := filepath.Join(dir, "orphan.ts")
	for path, contents := range map[string]string{existing: "old", orphan: "orphan"} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	set := &generatedClientContractSet{
		replacements: []generatedClientContract{
			{path: existing, contents: []byte("new")},
			{path: created, contents: []byte("created")},
		},
		deletions: []string{orphan},
	}
	if err := set.install(); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{existing: "new", created: "created"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v; want %q", path, got, err, want)
		}
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("orphan still exists or stat failed unexpectedly: %v", err)
	}
}

func TestClientContractSetRejectsOverlappingMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contract.ts")
	set := &generatedClientContractSet{
		replacements: []generatedClientContract{{path: path, contents: []byte("new")}},
		deletions:    []string{path},
	}
	if err := set.install(); err == nil || !strings.Contains(err.Error(), "both") {
		t.Fatalf("install error = %v, want overlap rejection", err)
	}
}
