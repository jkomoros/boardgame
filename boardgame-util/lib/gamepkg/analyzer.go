package gamepkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const boardgameImportPath = "github.com/jkomoros/boardgame"

var positionedPackageError = regexp.MustCompile(`^(.+):([0-9]+):([0-9]+):\s*(.+)$`)

// Position identifies the source location of an analysis diagnostic.
type Position struct {
	File   string `json:"file,omitempty"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
}

// Diagnostic is one package loading or game-contract problem.
type Diagnostic struct {
	Kind     DiagnosticKind `json:"kind"`
	Position Position       `json:"position,omitempty"`
	Message  string         `json:"message"`
}

// DiagnosticKind is a stable machine-readable category for package analysis.
type DiagnosticKind string

const (
	DiagnosticLoad       DiagnosticKind = "load"
	DiagnosticContract   DiagnosticKind = "contract"
	DiagnosticRandomness DiagnosticKind = "randomness"
)

// Analysis is the typed result of inspecting one Go package. Candidate is true
// when the package declares a package-level function named NewDelegate. A
// candidate is a valid game package only when Diagnostics is empty.
type Analysis struct {
	ImportPath  string       `json:"importPath"`
	Name        string       `json:"name"`
	Dir         string       `json:"dir"`
	Candidate   bool         `json:"candidate"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	loaded      bool
}

// InvalidGamePackageError preserves the complete typed analysis when a legacy
// constructor is asked to load an ordinary or malformed game package. Callers
// that need source-aware diagnostics can use errors.As to inspect Analysis.
type InvalidGamePackageError struct {
	Analysis Analysis
}

func (e *InvalidGamePackageError) Error() string {
	if !e.Analysis.Candidate && len(e.Analysis.Diagnostics) == 0 {
		return e.Analysis.Dir + " was not a valid game package: couldn't find NewDelegate"
	}
	return e.Analysis.Dir + " was not a valid game package: " + formatDiagnostics(e.Analysis.Diagnostics)
}

// ValidGame reports whether the package declares a correctly typed
// NewDelegate and passed all package-safety checks.
func (a Analysis) ValidGame() bool {
	return a.Candidate && len(a.Diagnostics) == 0
}

// GamePackage converts a successful analysis into the package handle used by
// boardgame-util commands. It does not reload or re-type-check the package.
func (a Analysis) GamePackage() (*Pkg, error) {
	if !a.loaded || !a.ValidGame() || a.ImportPath == "" || a.Name == "" || !filepath.IsAbs(a.Dir) {
		return nil, fmt.Errorf("analysis is not a validated package result")
	}
	if info, err := os.Stat(a.Dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("analysis package directory is invalid: %q", a.Dir)
	}
	pkg, _, err := newPkgFromAnalysis(a, "")
	return pkg, err
}

// Analyze loads and type-checks every package matched by patterns. It honors
// build tags and module configuration exactly as the Go tool does. Ordinary
// packages are returned with Candidate false, allowing callers such as lint to
// distinguish them from malformed game packages without source heuristics.
func Analyze(patterns []string, basePath string, options Options) ([]Analysis, error) {
	if len(patterns) == 0 {
		patterns = []string{"."}
	}
	if basePath == "" {
		var err error
		basePath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedSyntax |
		packages.NeedTypes
	config := &packages.Config{Mode: mode, Dir: basePath, Tests: false}
	if options.ReadOnly {
		config.BuildFlags = append(config.BuildFlags, "-mod=readonly")
	}
	loaded, err := packages.Load(config, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load Go packages: %w", err)
	}
	if len(loaded) == 0 {
		return nil, fmt.Errorf("package pattern matched no packages: %s", strings.Join(patterns, ", "))
	}

	analyses := make([]Analysis, 0, len(loaded))
	seenPackages := make(map[string]bool)
	for _, loadedPackage := range loaded {
		key := loadedPackage.PkgPath + "\x00" + packageDir(loadedPackage)
		if seenPackages[key] {
			continue
		}
		seenPackages[key] = true
		delegateInterface := gameDelegateInterface(loadedPackage)
		if delegateInterface == nil && hasTopLevelNewDelegate(loadedPackage) {
			if reloadedPackage, reloadedInterface, reloadErr := reloadWithBoardgame(config, loadedPackage.PkgPath); reloadErr == nil {
				loadedPackage = reloadedPackage
				delegateInterface = reloadedInterface
			}
		}
		analyses = append(analyses, analyzePackage(loadedPackage, delegateInterface))
	}
	sort.Slice(analyses, func(i, j int) bool {
		if analyses[i].ImportPath != analyses[j].ImportPath {
			return analyses[i].ImportPath < analyses[j].ImportPath
		}
		return analyses[i].Dir < analyses[j].Dir
	})
	return analyses, nil
}

func analyzePackage(pkg *packages.Package, delegateInterface *types.Interface) Analysis {
	result := Analysis{
		ImportPath: pkg.PkgPath,
		Name:       pkg.Name,
		Dir:        packageDir(pkg),
		loaded:     true,
	}
	if result.ImportPath == "" {
		result.ImportPath = pkg.ID
	}

	var declarations []*ast.FuncDecl
	for _, file := range pkg.Syntax {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Recv == nil && function.Name.Name == "NewDelegate" {
				declarations = append(declarations, function)
			}
		}
	}
	recoveredCandidate := false
	if len(declarations) == 0 && len(pkg.Errors) > 0 {
		recoveredCandidate = erroredPackageDeclaresNewDelegate(pkg)
	}
	result.Candidate = len(declarations) > 0 || recoveredCandidate
	if !result.Candidate {
		result.Diagnostics = normalizeDiagnostics(packageDiagnostics(pkg))
		return result
	}

	result.Diagnostics = append(result.Diagnostics, packageDiagnostics(pkg)...)
	if recoveredCandidate {
		result.Diagnostics = normalizeDiagnostics(result.Diagnostics)
		return result
	}
	result.Diagnostics = append(result.Diagnostics, validateNewDelegate(pkg, declarations, delegateInterface)...)
	result.Diagnostics = append(result.Diagnostics, validateRandImports(pkg)...)
	result.Diagnostics = normalizeDiagnostics(result.Diagnostics)
	return result
}

func hasTopLevelNewDelegate(pkg *packages.Package) bool {
	for _, file := range pkg.Syntax {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Recv == nil && function.Name.Name == "NewDelegate" {
				return true
			}
		}
	}
	return false
}

func reloadWithBoardgame(config *packages.Config, importPath string) (*packages.Package, *types.Interface, error) {
	loaded, err := packages.Load(config, importPath, boardgameImportPath)
	if err != nil {
		return nil, nil, err
	}
	var target *packages.Package
	var boardgamePackage *packages.Package
	for _, pkg := range loaded {
		switch pkg.PkgPath {
		case importPath:
			target = pkg
		case boardgameImportPath:
			boardgamePackage = pkg
		}
	}
	if target == nil || boardgamePackage == nil || boardgamePackage.Types == nil {
		return nil, nil, fmt.Errorf("could not load game and boardgame packages together")
	}
	object := boardgamePackage.Types.Scope().Lookup("GameDelegate")
	if object == nil {
		return nil, nil, fmt.Errorf("boardgame.GameDelegate was not found")
	}
	delegateInterface, ok := object.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, nil, fmt.Errorf("boardgame.GameDelegate is not an interface")
	}
	return target, delegateInterface, nil
}

func erroredPackageDeclaresNewDelegate(pkg *packages.Package) bool {
	for _, filename := range pkg.GoFiles {
		file, _ := parser.ParseFile(token.NewFileSet(), filename, nil, parser.AllErrors|parser.SkipObjectResolution)
		if file != nil {
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if ok && function.Recv == nil && function.Name.Name == "NewDelegate" {
					return true
				}
			}
		}
		// Severe syntax errors may prevent parser recovery from retaining later
		// declarations. The Go scanner still distinguishes identifiers from
		// comments and strings, so use it only as a last-resort classifier.
		contents, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		var lexer scanner.Scanner
		lexer.Init(token.NewFileSet().AddFile(filename, -1, len(contents)), contents, nil, 0)
		awaitingFunctionName := false
		for {
			_, scanned, literal := lexer.Scan()
			if scanned == token.EOF {
				break
			}
			if scanned == token.FUNC {
				awaitingFunctionName = true
				continue
			}
			if !awaitingFunctionName {
				continue
			}
			if scanned == token.IDENT {
				if literal == "NewDelegate" {
					return true
				}
				awaitingFunctionName = false
				continue
			}
			// A receiver begins with '(', so this is a method. Any other token
			// also means the declaration is too malformed to identify safely.
			awaitingFunctionName = false
		}
	}
	return false
}

func packageDir(pkg *packages.Package) string {
	if pkg.Dir != "" {
		return pkg.Dir
	}
	files := pkg.CompiledGoFiles
	if len(files) == 0 {
		files = pkg.GoFiles
	}
	if len(files) == 0 {
		return ""
	}
	return filepath.Dir(files[0])
}

func packageDiagnostics(pkg *packages.Package) []Diagnostic {
	result := make([]Diagnostic, 0, len(pkg.Errors))
	for _, problem := range pkg.Errors {
		position := parsePosition(problem.Pos)
		message := problem.Msg
		if position.File == "" {
			for _, line := range strings.Split(problem.Msg, "\n") {
				matches := positionedPackageError.FindStringSubmatch(line)
				if matches == nil {
					continue
				}
				position = Position{File: matches[1]}
				position.Line, _ = strconv.Atoi(matches[2])
				position.Column, _ = strconv.Atoi(matches[3])
				if !filepath.IsAbs(position.File) && packageDir(pkg) != "" {
					packageFile := filepath.Join(packageDir(pkg), filepath.Base(position.File))
					if _, err := os.Stat(packageFile); err == nil {
						position.File = packageFile
					}
				}
				message = matches[4]
				break
			}
		}
		result = append(result, Diagnostic{
			Kind:     DiagnosticLoad,
			Position: position,
			Message:  message,
		})
	}
	return result
}

func validateNewDelegate(pkg *packages.Package, declarations []*ast.FuncDecl, delegateInterface *types.Interface) []Diagnostic {
	if len(declarations) > 1 {
		return []Diagnostic{{
			Kind:     DiagnosticContract,
			Position: positionFor(pkg.Fset, declarations[1].Name.Pos()),
			Message:  "NewDelegate is declared more than once",
		}}
	}
	declaration := declarations[0]
	position := positionFor(pkg.Fset, declaration.Name.Pos())
	object := pkg.Types.Scope().Lookup("NewDelegate")
	function, ok := object.(*types.Func)
	if !ok {
		return []Diagnostic{{Kind: DiagnosticContract, Position: position, Message: "NewDelegate could not be type checked as a package-level function"}}
	}
	signature, ok := function.Type().(*types.Signature)
	if !ok {
		return []Diagnostic{{Kind: DiagnosticContract, Position: position, Message: "NewDelegate has an invalid function type"}}
	}
	if signature.Params().Len() != 0 {
		return []Diagnostic{{Kind: DiagnosticContract, Position: position, Message: "NewDelegate must not accept parameters"}}
	}
	if signature.Results().Len() != 1 {
		return []Diagnostic{{Kind: DiagnosticContract, Position: position, Message: "NewDelegate must return exactly one value"}}
	}
	if signature.TypeParams() != nil && signature.TypeParams().Len() != 0 {
		return []Diagnostic{{Kind: DiagnosticContract, Position: position, Message: "NewDelegate must not declare type parameters"}}
	}

	if delegateInterface == nil {
		return []Diagnostic{{Kind: DiagnosticContract, Position: position, Message: "could not load boardgame.GameDelegate for contract validation"}}
	}
	returned := signature.Results().At(0).Type()
	if !types.AssignableTo(returned, delegateInterface) {
		return []Diagnostic{{
			Kind:     DiagnosticContract,
			Position: position,
			Message:  fmt.Sprintf("NewDelegate returns %s, which does not implement boardgame.GameDelegate", types.TypeString(returned, qualifier)),
		}}
	}
	return nil
}

func gameDelegateInterface(pkg *packages.Package) *types.Interface {
	boardgamePackage := pkg.Imports[boardgameImportPath]
	if boardgamePackage == nil || boardgamePackage.Types == nil {
		return nil
	}
	object := boardgamePackage.Types.Scope().Lookup("GameDelegate")
	if object == nil {
		return nil
	}
	interfaceType, _ := object.Type().Underlying().(*types.Interface)
	return interfaceType
}

func qualifier(pkg *types.Package) string {
	if pkg == nil {
		return ""
	}
	return pkg.Name()
}

func validateRandImports(pkg *packages.Package) []Diagnostic {
	var result []Diagnostic
	for _, file := range pkg.Syntax {
		for _, imported := range file.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil || path != "math/rand" || importHasMagicComment(imported) {
				continue
			}
			result = append(result, Diagnostic{
				Kind:     DiagnosticRandomness,
				Position: positionFor(pkg.Fset, imported.Pos()),
				Message: "math/rand is unsafe in game logic; use state.Rand() for deterministic games, or add " +
					RandMagicComment + " to this import when the use is intentionally unrelated to game logic",
			})
		}
	}
	return result
}

func importHasMagicComment(imported *ast.ImportSpec) bool {
	for _, group := range []*ast.CommentGroup{imported.Doc, imported.Comment} {
		if group == nil {
			continue
		}
		for _, comment := range group.List {
			if strings.Contains(comment.Text, RandMagicComment) {
				return true
			}
		}
	}
	return false
}

func positionFor(fileSet *token.FileSet, pos token.Pos) Position {
	if fileSet == nil || !pos.IsValid() {
		return Position{}
	}
	position := fileSet.Position(pos)
	return Position{File: position.Filename, Line: position.Line, Column: position.Column}
}

func parsePosition(value string) Position {
	if value == "" || value == "-" {
		return Position{}
	}
	parts := strings.Split(value, ":")
	if len(parts) < 2 {
		return Position{File: value}
	}
	last := len(parts) - 1
	lastNumber, lastErr := strconv.Atoi(parts[last])
	if lastErr != nil {
		return Position{File: value}
	}
	if last >= 2 {
		if line, err := strconv.Atoi(parts[last-1]); err == nil {
			return Position{File: strings.Join(parts[:last-1], ":"), Line: line, Column: lastNumber}
		}
	}
	return Position{File: strings.Join(parts[:last], ":"), Line: lastNumber}
}

func normalizeDiagnostics(diagnostics []Diagnostic) []Diagnostic {
	if len(diagnostics) == 0 {
		return []Diagnostic{}
	}
	seen := make(map[string]bool)
	result := make([]Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		key := fmt.Sprintf("%s:%s:%d:%d:%s", diagnostic.Kind, diagnostic.Position.File, diagnostic.Position.Line, diagnostic.Position.Column, diagnostic.Message)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, diagnostic)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Position.File != right.Position.File {
			return left.Position.File < right.Position.File
		}
		if left.Position.Line != right.Position.Line {
			return left.Position.Line < right.Position.Line
		}
		if left.Position.Column != right.Position.Column {
			return left.Position.Column < right.Position.Column
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		return left.Message < right.Message
	})
	return result
}
