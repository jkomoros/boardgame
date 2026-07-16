package static

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const ClientCheckSchemaVersion = 1
const ClientCheckVersion = ClientCheckSchemaVersion

type ClientCheckPackage struct {
	ImportPath   string
	Name         string
	ClientFolder string
}

type ClientDiagnostic struct {
	Source      string `json:"source"`
	Code        string `json:"code"`
	Severity    string `json:"severity"`
	Package     string `json:"package,omitempty"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

type ClientCheckResult struct {
	Version     int                `json:"version"`
	OK          bool               `json:"ok"`
	Diagnostics []ClientDiagnostic `json:"diagnostics"`
}

type ClientCheckReport = ClientCheckResult

func NewClientCheckResult(diagnostics []ClientDiagnostic) ClientCheckResult {
	diagnostics = append([]ClientDiagnostic{}, diagnostics...)
	sort.SliceStable(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Column != b.Column {
			return a.Column < b.Column
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	return ClientCheckResult{Version: ClientCheckSchemaVersion, OK: len(diagnostics) == 0, Diagnostics: diagnostics}
}

func (r ClientCheckResult) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(r)
}

func (r ClientCheckResult) WriteHuman(w io.Writer) error {
	if r.OK {
		_, err := fmt.Fprintln(w, "Client checks passed")
		return err
	}
	for _, diagnostic := range r.Diagnostics {
		location := diagnostic.File
		if diagnostic.Line > 0 {
			location = fmt.Sprintf("%s:%d", location, diagnostic.Line)
			if diagnostic.Column > 0 {
				location = fmt.Sprintf("%s:%d", location, diagnostic.Column)
			}
		}
		if location != "" {
			location += ": "
		}
		if _, err := fmt.Fprintf(w, "%s%s %s: %s\n", location, diagnostic.Severity, diagnostic.Code, diagnostic.Message); err != nil {
			return err
		}
		if diagnostic.Remediation != "" {
			if _, err := fmt.Fprintf(w, "  Fix: %s\n", diagnostic.Remediation); err != nil {
				return err
			}
		}
	}
	return nil
}

type typescriptReport struct {
	Version     int `json:"version"`
	Diagnostics []struct {
		Source   string `json:"source"`
		Code     string `json:"code"`
		Category string `json:"category"`
		Message  string `json:"message"`
		File     string `json:"file,omitempty"`
		Line     int    `json:"line,omitempty"`
		Column   int    `json:"column,omitempty"`
	} `json:"diagnostics"`
	InfrastructureError string `json:"infrastructureError,omitempty"`
}

// CheckClient runs the strict framework authoring project once, then checks
// each game client in its own assembled namespace. Package isolation prevents
// two games with the same Go package name from shadowing one another.
func CheckClient(frameworkStaticDir string, packages []ClientCheckPackage) (ClientCheckReport, error) {
	report := NewClientCheckResult(nil)
	runner := filepath.Join(frameworkStaticDir, "scripts", "check-client.mjs")
	compiler := filepath.Join(frameworkStaticDir, "node_modules", "typescript", "bin", "tsc")
	for _, required := range []string{runner, compiler, filepath.Join(frameworkStaticDir, "tsconfig.authoring.json")} {
		if _, err := os.Stat(required); err != nil {
			return report, fmt.Errorf("BGCLIENT0001: required framework client tool %s is unavailable: %w", required, err)
		}
	}

	diagnostics, err := runTypeScriptCheck(runner, filepath.Join(frameworkStaticDir, "tsconfig.authoring.json"), "framework")
	if err != nil {
		return report, err
	}
	report.Diagnostics = append(report.Diagnostics, diagnostics...)

	workDir, err := os.MkdirTemp("", "boardgame-check-client-")
	if err != nil {
		return report, fmt.Errorf("BGCLIENT0001: couldn't create client-check workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	frameworkDeclarations := filepath.Join(workDir, "game-src", "src")
	if err := os.MkdirAll(frameworkDeclarations, 0700); err != nil {
		return report, fmt.Errorf("BGCLIENT0001: couldn't stage framework declarations: %w", err)
	}
	declarationCommand := exec.Command("node", compiler,
		"--project", filepath.Join(frameworkStaticDir, "tsconfig.json"),
		"--declaration", "--emitDeclarationOnly", "--declarationMap", "false", "--sourceMap", "false",
		"--outDir", frameworkDeclarations, "--rootDir", filepath.Join(frameworkStaticDir, "src"), "--pretty", "false")
	if output, commandErr := declarationCommand.CombinedOutput(); commandErr != nil {
		return report, fmt.Errorf("BGCLIENT0001: couldn't emit framework declarations: %w: %s", commandErr, strings.TrimSpace(string(output)))
	}
	if err := os.Symlink(filepath.Join(frameworkStaticDir, "node_modules"), filepath.Join(workDir, "node_modules")); err != nil {
		return report, fmt.Errorf("BGCLIENT0001: couldn't expose pinned framework dependencies: %w", err)
	}

	packages = append([]ClientCheckPackage(nil), packages...)
	sort.Slice(packages, func(i, j int) bool { return packages[i].ImportPath < packages[j].ImportPath })
	seenImports := make(map[string]bool)
	for _, game := range packages {
		if game.ClientFolder == "" {
			continue
		}
		if seenImports[game.ImportPath] {
			continue
		}
		seenImports[game.ImportPath] = true
		gameDir := filepath.Join(workDir, "game-src", packageDirectoryName(game))
		if err := os.MkdirAll(gameDir, 0700); err != nil {
			return report, fmt.Errorf("BGCLIENT0001: couldn't stage %s: %w", game.ImportPath, err)
		}
		if err := os.Symlink(game.ClientFolder, filepath.Join(gameDir, "client")); err != nil {
			return report, fmt.Errorf("BGCLIENT0001: couldn't link client for %s: %w", game.ImportPath, err)
		}
		projectPath := filepath.Join(gameDir, "tsconfig.json")
		project := map[string]interface{}{
			"extends":         filepath.Join(frameworkStaticDir, "tsconfig.authoring.json"),
			"compilerOptions": map[string]interface{}{"rootDir": "../..", "noEmit": true},
			"include":         []string{"client/**/*.ts"},
			"exclude":         []string{"client/**/*.test.ts"},
		}
		contents, marshalErr := json.MarshalIndent(project, "", "  ")
		if marshalErr != nil {
			return report, fmt.Errorf("BGCLIENT0001: couldn't encode project for %s: %w", game.ImportPath, marshalErr)
		}
		if err := os.WriteFile(projectPath, append(contents, '\n'), 0600); err != nil {
			return report, fmt.Errorf("BGCLIENT0001: couldn't write project for %s: %w", game.ImportPath, err)
		}
		diagnostics, err := runTypeScriptCheck(runner, projectPath, game.ImportPath, "--creator-policy", "--lit")
		if err != nil {
			return report, err
		}
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
	}
	return NewClientCheckResult(report.Diagnostics), nil
}

func runTypeScriptCheck(runner, project, packageName string, options ...string) ([]ClientDiagnostic, error) {
	arguments := []string{runner, "--project", project}
	arguments = append(arguments, options...)
	command := exec.Command("node", arguments...)
	output, commandErr := command.CombinedOutput()
	var exitErr *exec.ExitError
	if commandErr != nil && (!errors.As(commandErr, &exitErr) || (exitErr.ExitCode() != 1 && exitErr.ExitCode() != 2)) {
		return nil, fmt.Errorf("BGCLIENT0001: couldn't run TypeScript check for %s: %w: %s", packageName, commandErr, strings.TrimSpace(string(output)))
	}
	var result typescriptReport
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("BGCLIENT0001: invalid checker output for %s: %w: %s", packageName, err, strings.TrimSpace(string(output)))
	}
	if result.Version != ClientCheckVersion {
		return nil, fmt.Errorf("BGCLIENT0001: unsupported checker result version %d for %s", result.Version, packageName)
	}
	if result.InfrastructureError != "" {
		return nil, fmt.Errorf("BGCLIENT0001: TypeScript checker infrastructure failed for %s: %s", packageName, result.InfrastructureError)
	}
	diagnostics := make([]ClientDiagnostic, 0, len(result.Diagnostics))
	for _, item := range result.Diagnostics {
		diagnostics = append(diagnostics, ClientDiagnostic{Package: packageName, Source: item.Source, Code: item.Code, Severity: item.Category, Message: item.Message, File: item.File, Line: item.Line, Column: item.Column})
	}
	return diagnostics, nil
}

func packageDirectoryName(game ClientCheckPackage) string {
	var name strings.Builder
	for _, character := range game.Name {
		if unicode.IsLetter(character) || unicode.IsDigit(character) || character == '-' || character == '_' {
			name.WriteRune(character)
		} else {
			name.WriteByte('-')
		}
	}
	if name.Len() == 0 {
		name.WriteString("game")
	}
	digest := sha256.Sum256([]byte(game.ImportPath))
	return name.String() + "-" + hex.EncodeToString(digest[:4])
}
