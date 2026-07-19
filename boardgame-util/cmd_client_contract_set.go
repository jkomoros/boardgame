package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type generatedClientContract struct {
	path, gameName, kind string
	contents             []byte
}

type generatedClientContractSet struct {
	replacements []generatedClientContract
	deletions    []string
}

func generateCompleteClientContracts(base *boardgameUtil, packages []*gamepkg.Pkg, includeReadOnly bool) (*generatedClientContractSet, error) {
	gameTypes, err := generateGameTypesForPackages(base, packages, includeReadOnly)
	if err != nil {
		return nil, err
	}
	moveNames, err := generateMoveNamesForPackages(base, packages, includeReadOnly)
	if err != nil {
		return nil, err
	}
	moveArgs, err := generateMoveArgsForPackages(base, packages, includeReadOnly)
	if err != nil {
		return nil, err
	}
	boardSpaces, orphans, err := generateBoardSpacesForPackages(packages, includeReadOnly)
	if err != nil {
		return nil, err
	}

	set := &generatedClientContractSet{deletions: append([]string(nil), orphans...)}
	for _, file := range moveNames {
		set.replacements = append(set.replacements, generatedClientContract{
			path: file.path, contents: file.contents, gameName: file.gameName, kind: "move names",
		})
	}
	for _, file := range moveArgs {
		set.replacements = append(set.replacements, generatedClientContract{
			path: file.path, contents: file.contents, gameName: file.gameName, kind: "move inputs",
		})
	}
	for _, file := range boardSpaces {
		set.replacements = append(set.replacements, generatedClientContract{
			path: file.path, contents: file.contents, kind: "board spaces",
		})
	}
	for _, file := range gameTypes {
		set.replacements = append(set.replacements, generatedClientContract{
			path: file.path, contents: file.contents, gameName: file.gameName, kind: "state/renderer types",
		})
	}

	staged := make(map[string][]byte, len(set.replacements))
	for _, file := range set.replacements {
		staged[file.path] = file.contents
	}
	if err := validateGeneratedGameTypesTypeScript(gameTypes, staged); err != nil {
		return nil, err
	}
	if err := set.validatePaths(); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *generatedClientContractSet) validatePaths() error {
	seen := make(map[string]string, len(s.replacements)+len(s.deletions))
	for _, file := range s.replacements {
		if prior, exists := seen[file.path]; exists {
			return fmt.Errorf("duplicate generated client destination %s (%s and %s)", file.path, prior, file.kind)
		}
		seen[file.path] = file.kind
	}
	for _, path := range s.deletions {
		if prior, exists := seen[path]; exists {
			return fmt.Errorf("generated client destination %s is both %s and an orphan deletion", path, prior)
		}
		seen[path] = "orphan deletion"
	}
	return nil
}

func (s *generatedClientContractSet) check() error {
	var stale []string
	for _, file := range s.replacements {
		current, err := os.ReadFile(file.path)
		if err != nil || !bytes.Equal(current, file.contents) {
			stale = append(stale, file.path)
		}
	}
	stale = append(stale, s.deletions...)
	sort.Strings(stale)
	if len(stale) > 0 {
		return staleGeneratedClientContractPaths(stale)
	}
	return nil
}

func (s *generatedClientContractSet) install() error {
	if err := s.validatePaths(); err != nil {
		return err
	}
	files := make(map[string]fileutil.FileSpec, len(s.replacements)+len(s.deletions))
	for _, file := range s.replacements {
		files[file.path] = fileutil.FileSpec{Contents: file.contents, Mode: 0o644, ForceMode: true}
	}
	for _, path := range s.deletions {
		files[path] = fileutil.FileSpec{Delete: true, RequireExisting: true}
	}
	if err := fileutil.WriteFileSetAtomicAbsolute(files, true); err != nil {
		return fmt.Errorf("install generated client contracts: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Installed %d generated client contracts", len(s.replacements))
	if len(s.deletions) > 0 {
		fmt.Fprintf(os.Stderr, " and removed %d orphan(s)", len(s.deletions))
	}
	fmt.Fprintln(os.Stderr)
	return nil
}
