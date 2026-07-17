// Package lint provides non-destructive preflight checks for boardgame game
// packages. It is the implementation behind boardgame-util lint.
package lint

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/codegen"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

const ReportVersion = 1

const (
	CodePackage   = "BGLINT0001"
	CodeGenerated = "BGLINT0002"
	CodeRuntime   = "BGLINT0003"
)

var (
	legalPathFailure = regexp.MustCompile(`legal path "([^"]+)"`)
	legalSpecFailure = regexp.MustCompile(`legal spec "([^"]+)"`)
)

// Diagnostic is one actionable game-author preflight failure.
type Diagnostic struct {
	Package     string `json:"package,omitempty"`
	Code        string `json:"code"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

// Report is the deterministic result of checking one or more game packages.
type Report struct {
	Version     int          `json:"version"`
	OK          bool         `json:"ok"`
	Packages    []string     `json:"packages"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Options controls optional lint behavior.
type Options struct {
	// BasePath resolves relative exact package paths and is the working
	// directory for go-list patterns. Empty means the current directory.
	BasePath string
	// Fix atomically refreshes deterministic generated Go files before the
	// runtime validation pass.
	Fix bool
}

type generatedFile struct {
	name     string
	contents string
}

// Check validates all exact game packages and go-list patterns in inputs. An
// empty input list means the current package. Patterns such as ./... ignore
// ordinary non-game packages but validate every package that declares a
// top-level NewDelegate function.
func Check(inputs []string, options Options) Report {
	if len(inputs) == 0 {
		inputs = []string{"."}
	}
	packages, diagnostics := resolvePackages(inputs, options.BasePath)
	checkedPackages := make([]string, 0, len(packages))
	for _, pkg := range packages {
		checkedPackages = append(checkedPackages, pkg.Import())
		files, err := expectedGeneratedFiles(pkg.AbsolutePath())
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{
				Package: pkg.Import(), Code: CodeGenerated,
				Message:     "could not generate expected Go contracts: " + err.Error(),
				Remediation: "Fix the reported code-generation error, then rerun boardgame-util lint.",
			})
		} else {
			for _, file := range files {
				diagnostic, stale := checkGeneratedFile(pkg.Import(), pkg.AbsolutePath(), file, options.Fix)
				if stale {
					diagnostics = append(diagnostics, diagnostic)
				}
			}
		}

		calls, _ := sourceLegalCalls(pkg.AbsolutePath())
		if err := validateRuntime(pkg); err != nil {
			diagnostic := Diagnostic{
				Package: pkg.Import(), Code: CodeRuntime,
				Message:     "game manager preflight failed: " + err.Error(),
				Remediation: "Fix the game configuration or compile error reported above, then rerun boardgame-util lint.",
			}
			attachLegalCallPosition(&diagnostic, err.Error(), calls)
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	sort.Slice(diagnostics, func(i, j int) bool {
		left := diagnostics[i]
		right := diagnostics[j]
		if left.Package != right.Package {
			return left.Package < right.Package
		}
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Column != right.Column {
			return left.Column < right.Column
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Message < right.Message
	})
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	return Report{Version: ReportVersion, OK: len(diagnostics) == 0, Packages: checkedPackages, Diagnostics: diagnostics}
}

func attachLegalCallPosition(diagnostic *Diagnostic, runtimeError string, calls []legalCallPath) {
	pathMatch := legalPathFailure.FindStringSubmatch(runtimeError)
	if len(pathMatch) != 2 {
		return
	}
	predicates := make(map[string]bool)
	for _, specMatch := range legalSpecFailure.FindAllStringSubmatch(runtimeError, -1) {
		if len(specMatch) == 2 {
			predicates[specMatch[1]] = true
		}
	}
	var matches []legalCallPath
	for _, call := range calls {
		if call.path != pathMatch[1] || (len(predicates) != 0 && !predicates[call.predicate]) {
			continue
		}
		matches = append(matches, call)
	}
	// A generic runtime location is better than confidently pointing at the
	// wrong move when the same predicate/path literal occurs more than once.
	if len(matches) != 1 {
		return
	}
	diagnostic.File = matches[0].file
	diagnostic.Line = matches[0].line
	diagnostic.Column = matches[0].column
	diagnostic.Remediation = "Fix this literal legal path or its state-property type, then rerun boardgame-util lint."
}

func resolvePackages(inputs []string, basePath string) ([]*gamepkg.Pkg, []Diagnostic) {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	packagesByImport := make(map[string]*gamepkg.Pkg)
	var diagnostics []Diagnostic
	for _, input := range inputs {
		if strings.Contains(input, "...") {
			analyses, err := gamepkg.Analyze([]string{input}, basePath, gamepkg.Options{ReadOnly: true})
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{Code: CodePackage, Package: input, Message: "could not analyze package pattern: " + err.Error(), Remediation: "Fix the package pattern or its Go module errors."})
				continue
			}
			found := false
			for _, analysis := range analyses {
				if !analysis.Candidate {
					continue
				}
				found = true
				if !analysis.ValidGame() {
					diagnostics = append(diagnostics, analysisDiagnostics(analysis)...)
					continue
				}
				pkg, err := analysis.GamePackage()
				if err != nil {
					diagnostics = append(diagnostics, packageDiagnostic(analysis.ImportPath, err))
					continue
				}
				packagesByImport[pkg.Import()] = pkg
			}
			if !found {
				loadFailure := false
				for _, analysis := range analyses {
					if analysis.Dir != "" || len(analysis.Diagnostics) == 0 {
						continue
					}
					loadFailure = true
					diagnostics = append(diagnostics, analysisDiagnostics(analysis)...)
				}
				if !loadFailure {
					diagnostics = append(diagnostics, Diagnostic{Code: CodePackage, Package: input, Message: "pattern did not contain any package declaring NewDelegate", Remediation: "Run lint against a game package or a pattern containing game packages."})
				}
			}
			continue
		}

		pkg, err := gamepkg.NewWithOptions(input, basePath, gamepkg.Options{ReadOnly: true})
		if err != nil {
			diagnostics = append(diagnostics, packageErrorDiagnostics(input, err)...)
			continue
		}
		packagesByImport[pkg.Import()] = pkg
	}
	imports := make([]string, 0, len(packagesByImport))
	for importPath := range packagesByImport {
		imports = append(imports, importPath)
	}
	sort.Strings(imports)
	packages := make([]*gamepkg.Pkg, 0, len(imports))
	for _, importPath := range imports {
		packages = append(packages, packagesByImport[importPath])
	}
	return packages, diagnostics
}

func analysisDiagnostics(analysis gamepkg.Analysis) []Diagnostic {
	result := make([]Diagnostic, 0, len(analysis.Diagnostics))
	for _, problem := range analysis.Diagnostics {
		result = append(result, Diagnostic{
			Package:     analysis.ImportPath,
			Code:        CodePackage,
			File:        problem.Position.File,
			Line:        problem.Position.Line,
			Column:      problem.Position.Column,
			Message:     problem.Message,
			Remediation: "Fix the typed package or NewDelegate contract error, then rerun boardgame-util lint.",
		})
	}
	return result
}

func packageDiagnostic(input string, err error) Diagnostic {
	return Diagnostic{
		Package: input, Code: CodePackage,
		Message:     "could not load game package: " + err.Error(),
		Remediation: "Ensure the package compiles, exports func NewDelegate() boardgame.GameDelegate, and uses state.Rand() for game randomness.",
	}
}

func packageErrorDiagnostics(input string, err error) []Diagnostic {
	var invalid *gamepkg.InvalidGamePackageError
	if !errors.As(err, &invalid) || len(invalid.Analysis.Diagnostics) == 0 {
		return []Diagnostic{packageDiagnostic(input, err)}
	}
	result := make([]Diagnostic, 0, len(invalid.Analysis.Diagnostics))
	for _, problem := range invalid.Analysis.Diagnostics {
		result = append(result, Diagnostic{
			Package:     invalid.Analysis.ImportPath,
			Code:        CodePackage,
			File:        problem.Position.File,
			Line:        problem.Position.Line,
			Column:      problem.Position.Column,
			Message:     problem.Message,
			Remediation: "Fix the typed package or NewDelegate contract error, then rerun boardgame-util lint.",
		})
	}
	return result
}

func expectedGeneratedFiles(dir string) ([]generatedFile, error) {
	reader, readerTest, err := codegen.ProcessReaders(dir)
	if err != nil {
		return nil, err
	}
	enums, err := codegen.ProcessEnums(dir)
	if err != nil {
		return nil, err
	}
	return []generatedFile{
		{name: "auto_enum.go", contents: enums},
		{name: "auto_reader.go", contents: reader},
		{name: "auto_reader_test.go", contents: readerTest},
	}, nil
}

func checkGeneratedFile(packageName, packageDir string, expected generatedFile, fix bool) (Diagnostic, bool) {
	path := filepath.Join(packageDir, expected.name)
	actual, err := os.ReadFile(path)
	missing := os.IsNotExist(err)
	if err != nil && !missing {
		return generatedDiagnostic(packageName, packageDir, expected.name, "could not read generated file: "+err.Error()), true
	}
	if (missing && expected.contents == "") || (!missing && bytes.Equal(actual, []byte(expected.contents))) {
		return Diagnostic{}, false
	}

	if fix {
		if !missing && !generatedFileOwned(actual) {
			return generatedDiagnostic(packageName, packageDir, expected.name, "refusing to replace a file that is not marked as boardgame-generated"), true
		}
		if expected.contents == "" {
			if err := os.Remove(path); err == nil || os.IsNotExist(err) {
				return Diagnostic{}, false
			} else {
				return generatedDiagnostic(packageName, packageDir, expected.name, "could not remove orphaned generated file: "+err.Error()), true
			}
		}
		if err := atomicWrite(path, []byte(expected.contents)); err == nil {
			return Diagnostic{}, false
		} else {
			return generatedDiagnostic(packageName, packageDir, expected.name, "could not refresh generated file: "+err.Error()), true
		}
	}

	message := "generated file is stale"
	if missing {
		message = "generated file is missing"
	} else if expected.contents == "" {
		message = "generated file is orphaned"
	}
	return generatedDiagnostic(packageName, packageDir, expected.name, message), true
}

func generatedFileOwned(contents []byte) bool {
	return bytes.Contains(contents, []byte("generated by the codegen package via 'boardgame-util codegen'")) &&
		bytes.Contains(contents, []byte("DO NOT EDIT"))
}

func generatedDiagnostic(packageName, packageDir, name, message string) Diagnostic {
	return Diagnostic{
		Package: packageName, Code: CodeGenerated, File: filepath.Join(packageDir, name),
		Message:     message,
		Remediation: "Run boardgame-util lint --fix " + strconv.Quote(packageDir) + " and commit the generated result.",
	}
}

func atomicWrite(path string, contents []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".boardgame-lint-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(contents); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func validateRuntime(pkg *gamepkg.Pkg) error {
	moduleCmd := exec.Command("go", "env", "GOMOD")
	moduleCmd.Dir = pkg.AbsolutePath()
	moduleOutput, err := moduleCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not locate module: %w: %s", err, strings.TrimSpace(string(moduleOutput)))
	}
	goMod := strings.TrimSpace(string(moduleOutput))
	if goMod == "" || goMod == os.DevNull {
		return fmt.Errorf("package is not inside a Go module")
	}
	tempDir, err := os.MkdirTemp(filepath.Dir(goMod), ".boardgame-lint-*")
	if err != nil {
		return fmt.Errorf("could not create temporary preflight package: %w", err)
	}
	defer os.RemoveAll(tempDir)

	source := fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/jkomoros/boardgame"
	game %s
	"github.com/jkomoros/boardgame/storage/memory"
)

func main() {
	defer func() {
		if value := recover(); value != nil {
			fmt.Fprintln(os.Stderr, value)
			os.Exit(1)
		}
	}()
	if _, err := boardgame.NewGameManager(game.NewDelegate(), memory.NewStorageManager()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, strconv.Quote(pkg.Import()))
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(source), 0644); err != nil {
		return fmt.Errorf("could not write temporary preflight package: %w", err)
	}
	cmd := exec.Command("go", "run", "-mod=readonly", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("%s", message)
	}
	return nil
}
