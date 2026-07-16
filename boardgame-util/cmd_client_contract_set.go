package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

type clientContractMutation struct {
	path, tempPath, backupPath string
	contents                   []byte
	delete                     bool
	hadOriginal, installed     bool
}

var renameGeneratedClientContract = os.Rename
var restoreGeneratedClientContract = os.Rename

func (s *generatedClientContractSet) install() error {
	if err := s.validatePaths(); err != nil {
		return err
	}
	mutations := make([]clientContractMutation, 0, len(s.replacements)+len(s.deletions))
	for _, file := range s.replacements {
		mutations = append(mutations, clientContractMutation{path: file.path, contents: file.contents})
	}
	for _, path := range s.deletions {
		mutations = append(mutations, clientContractMutation{path: path, delete: true})
	}
	sort.Slice(mutations, func(i, j int) bool { return mutations[i].path < mutations[j].path })

	cleanupTemps := func() {
		for _, mutation := range mutations {
			if mutation.tempPath != "" {
				_ = os.Remove(mutation.tempPath)
			}
		}
	}
	defer cleanupTemps()

	// Inspect every destination and stage every replacement before mutating a
	// single checked-in contract.
	for i := range mutations {
		info, err := os.Lstat(mutations[i].path)
		if err == nil {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("refusing to mutate non-file generated destination %s", mutations[i].path)
			}
			mutations[i].hadOriginal = true
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("couldn't inspect generated destination %s: %w", mutations[i].path, err)
		} else if mutations[i].delete {
			return fmt.Errorf("orphan generated destination disappeared before installation: %s", mutations[i].path)
		}
		if mutations[i].delete {
			placeholder, err := os.CreateTemp(filepath.Dir(mutations[i].path), ".client-contract-delete-*")
			if err != nil {
				return fmt.Errorf("couldn't prepare orphan backup for %s: %w", mutations[i].path, err)
			}
			mutations[i].tempPath = placeholder.Name()
			if err := placeholder.Close(); err != nil {
				return fmt.Errorf("couldn't close orphan backup placeholder for %s: %w", mutations[i].path, err)
			}
			if err := os.Remove(mutations[i].tempPath); err != nil {
				return fmt.Errorf("couldn't prepare orphan backup path for %s: %w", mutations[i].path, err)
			}
			continue
		}
		file, err := os.CreateTemp(filepath.Dir(mutations[i].path), ".client-contract-*")
		if err != nil {
			return fmt.Errorf("couldn't stage %s: %w", mutations[i].path, err)
		}
		mutations[i].tempPath = file.Name()
		if err := file.Chmod(0o644); err != nil {
			_ = file.Close()
			return fmt.Errorf("couldn't set staged permissions for %s: %w", mutations[i].path, err)
		}
		if _, err := file.Write(mutations[i].contents); err != nil {
			_ = file.Close()
			return fmt.Errorf("couldn't stage %s: %w", mutations[i].path, err)
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			return fmt.Errorf("couldn't sync staged contract %s: %w", mutations[i].path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("couldn't close staged contract %s: %w", mutations[i].path, err)
		}
	}

	rollback := func(last int) error {
		var first error
		for i := last; i >= 0; i-- {
			if mutations[i].installed && !mutations[i].delete {
				if err := os.Remove(mutations[i].path); err != nil && !os.IsNotExist(err) && first == nil {
					first = err
				}
			}
			if mutations[i].backupPath != "" {
				if err := restoreGeneratedClientContract(mutations[i].backupPath, mutations[i].path); err != nil {
					if first == nil {
						first = err
					}
				} else {
					mutations[i].backupPath = ""
				}
			}
		}
		return first
	}

	for i := range mutations {
		if mutations[i].hadOriginal {
			mutations[i].backupPath = mutations[i].tempPath + ".backup"
			if err := renameGeneratedClientContract(mutations[i].path, mutations[i].backupPath); err != nil {
				return fmt.Errorf("couldn't preserve prior generated contract %s: %w (rollback: %v)", mutations[i].path, err, rollback(i-1))
			}
		}
		if !mutations[i].delete {
			if err := renameGeneratedClientContract(mutations[i].tempPath, mutations[i].path); err != nil {
				currentRestore := error(nil)
				if mutations[i].backupPath != "" {
					currentRestore = restoreGeneratedClientContract(mutations[i].backupPath, mutations[i].path)
					if currentRestore == nil {
						mutations[i].backupPath = ""
					}
				}
				return fmt.Errorf("couldn't install generated contract %s: %w (current restore: %v; prior rollback: %v; backup: %s)", mutations[i].path, err, currentRestore, rollback(i-1), mutations[i].backupPath)
			}
			mutations[i].tempPath = ""
		}
		mutations[i].installed = true
	}

	for i := range mutations {
		if mutations[i].backupPath != "" {
			if err := os.Remove(mutations[i].backupPath); err != nil {
				return fmt.Errorf("installed complete client generation but couldn't remove backup for %s: %w", mutations[i].path, err)
			}
			mutations[i].backupPath = ""
		}
	}
	fmt.Fprintf(os.Stderr, "Installed %d generated client contracts", len(s.replacements))
	if len(s.deletions) > 0 {
		fmt.Fprintf(os.Stderr, " and removed %d orphan(s)", len(s.deletions))
	}
	fmt.Fprintln(os.Stderr)
	return nil
}
