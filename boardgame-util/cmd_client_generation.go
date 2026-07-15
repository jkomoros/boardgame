package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type staleGeneratedClientContractsError struct{ message string }

func (e *staleGeneratedClientContractsError) Error() string { return e.message }
func staleGeneratedClientContracts(message string) error {
	return &staleGeneratedClientContractsError{message: message}
}
func isStaleGeneratedClientContracts(err error) bool {
	var stale *staleGeneratedClientContractsError
	return errors.As(err, &stale)
}

func validateClientExtractionResults(pkgs []*gamepkg.Pkg, resultImports []string, kind string) error {
	actual := make(map[string]int, len(resultImports))
	for _, importPath := range resultImports {
		actual[importPath]++
	}
	var missing, duplicates []string
	seenExpected := make(map[string]bool)
	for _, pkg := range pkgs {
		if pkg == nil || pkg.ClientFolder() == "" || seenExpected[pkg.Import()] {
			continue
		}
		seenExpected[pkg.Import()] = true
		if actual[pkg.Import()] == 0 {
			missing = append(missing, pkg.Import())
		}
	}
	for importPath, count := range actual {
		if count > 1 {
			duplicates = append(duplicates, importPath)
		}
	}
	sort.Strings(missing)
	sort.Strings(duplicates)
	if len(missing) > 0 || len(duplicates) > 0 {
		return fmt.Errorf("%s extractor returned an incomplete or ambiguous client result set (missing: %s; duplicates: %s)", kind, strings.Join(missing, ", "), strings.Join(duplicates, ", "))
	}
	return nil
}
