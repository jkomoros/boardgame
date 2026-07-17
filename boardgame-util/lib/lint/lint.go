// Package lint provides non-destructive preflight checks for boardgame game
// packages. It is the implementation behind boardgame-util lint.
package lint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
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

// Diagnostic is one actionable game-author preflight failure.
type Diagnostic struct {
	Package     string `json:"package,omitempty"`
	Code        string `json:"code"`
	File        string `json:"file,omitempty"`
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

		if err := validateRuntime(pkg); err != nil {
			diagnostics = append(diagnostics, Diagnostic{
				Package: pkg.Import(), Code: CodeRuntime,
				Message:     "game manager preflight failed: " + err.Error(),
				Remediation: "Fix the game configuration or compile error reported above, then rerun boardgame-util lint.",
			})
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

func resolvePackages(inputs []string, basePath string) ([]*gamepkg.Pkg, []Diagnostic) {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	packagesByImport := make(map[string]*gamepkg.Pkg)
	var diagnostics []Diagnostic
	for _, input := range inputs {
		if strings.Contains(input, "...") {
			dirs, err := expandPattern(input, basePath)
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{Code: CodePackage, Package: input, Message: err.Error(), Remediation: "Fix the package pattern or its Go module errors."})
				continue
			}
			found := false
			for _, dir := range dirs {
				candidate, err := declaresNewDelegate(dir)
				if err != nil {
					diagnostics = append(diagnostics, Diagnostic{Code: CodePackage, Package: dir, Message: "could not inspect package: " + err.Error()})
					continue
				}
				if !candidate {
					continue
				}
				found = true
				pkg, err := gamepkg.NewFromPathWithOptions(dir, "", gamepkg.Options{ReadOnly: true})
				if err != nil {
					diagnostics = append(diagnostics, packageDiagnostic(dir, err))
					continue
				}
				packagesByImport[pkg.Import()] = pkg
			}
			if !found {
				diagnostics = append(diagnostics, Diagnostic{Code: CodePackage, Package: input, Message: "pattern did not contain any package declaring NewDelegate", Remediation: "Run lint against a game package or a pattern containing game packages."})
			}
			continue
		}

		pkg, err := gamepkg.NewWithOptions(input, basePath, gamepkg.Options{ReadOnly: true})
		if err != nil {
			diagnostics = append(diagnostics, packageDiagnostic(input, err))
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

func packageDiagnostic(input string, err error) Diagnostic {
	return Diagnostic{
		Package: input, Code: CodePackage,
		Message:     "could not load game package: " + err.Error(),
		Remediation: "Ensure the package compiles, exports func NewDelegate() boardgame.GameDelegate, and uses state.Rand() for game randomness.",
	}
}

func expandPattern(pattern, basePath string) ([]string, error) {
	cmd := exec.Command("go", "list", "-mod=readonly", "-f", "{{.Dir}}", pattern)
	cmd.Dir = basePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go list %s failed: %w: %s", pattern, err, strings.TrimSpace(string(output)))
	}
	seen := make(map[string]bool)
	var result []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		dir := strings.TrimSpace(line)
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		result = append(result, dir)
	}
	sort.Strings(result)
	return result, nil
}

func declaresNewDelegate(dir string) (bool, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return false, err
	}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if ok && function.Recv == nil && function.Name.Name == "NewDelegate" {
					return true, nil
				}
			}
		}
	}
	return false, nil
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
