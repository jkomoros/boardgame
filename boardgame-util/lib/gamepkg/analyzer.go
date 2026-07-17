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
	Position Position `json:"position,omitempty"`
	Message  string   `json:"message"`
}

// Analysis is the typed result of inspecting one Go package. Candidate is true
// when the package declares any function or method named NewDelegate. A
// candidate is a valid game package only when Diagnostics is empty.
type Analysis struct {
	ImportPath  string       `json:"importPath"`
	Name        string       `json:"name"`
	Dir         string       `json:"dir"`
	Candidate   bool         `json:"candidate"`
	Diagnostics []Diagnostic `json:"diagnostics"`
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
		packages.NeedTypes |
		packages.NeedTypesInfo
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
	for _, loadedPackage := range loaded {
		analyses = append(analyses, analyzePackage(loadedPackage))
	}
	sort.Slice(analyses, func(i, j int) bool {
		if analyses[i].ImportPath != analyses[j].ImportPath {
			return analyses[i].ImportPath < analyses[j].ImportPath
		}
		return analyses[i].Dir < analyses[j].Dir
	})
	return analyses, nil
}

func analyzePackage(pkg *packages.Package) Analysis {
	result := Analysis{
		ImportPath: pkg.PkgPath,
		Name:       pkg.Name,
		Dir:        packageDir(pkg),
	}
	if result.ImportPath == "" {
		result.ImportPath = pkg.ID
	}

	var declarations []*ast.FuncDecl
	for _, file := range pkg.Syntax {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Name.Name == "NewDelegate" {
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
	result.Diagnostics = append(result.Diagnostics, validateNewDelegate(pkg, declarations)...)
	result.Diagnostics = append(result.Diagnostics, validateRandImports(pkg)...)
	result.Diagnostics = normalizeDiagnostics(result.Diagnostics)
	return result
}

func erroredPackageDeclaresNewDelegate(pkg *packages.Package) bool {
	for _, filename := range pkg.GoFiles {
		file, _ := parser.ParseFile(token.NewFileSet(), filename, nil, parser.AllErrors|parser.SkipObjectResolution)
		if file != nil {
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if ok && function.Name.Name == "NewDelegate" {
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
		insideFunctionDeclaration := false
		for {
			_, scanned, literal := lexer.Scan()
			if scanned == token.EOF {
				break
			}
			if scanned == token.FUNC {
				insideFunctionDeclaration = true
				continue
			}
			if insideFunctionDeclaration && scanned == token.IDENT && literal == "NewDelegate" {
				return true
			}
			if scanned == token.LBRACE || scanned == token.SEMICOLON {
				insideFunctionDeclaration = false
			}
		}
	}
	return false
}

func packageDir(pkg *packages.Package) string {
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
			Position: position,
			Message:  message,
		})
	}
	return result
}

func validateNewDelegate(pkg *packages.Package, declarations []*ast.FuncDecl) []Diagnostic {
	if len(declarations) > 1 {
		return []Diagnostic{{
			Position: positionFor(pkg.Fset, declarations[1].Name.Pos()),
			Message:  "NewDelegate is declared more than once",
		}}
	}
	declaration := declarations[0]
	position := positionFor(pkg.Fset, declaration.Name.Pos())
	if declaration.Recv != nil {
		return []Diagnostic{{Position: position, Message: "NewDelegate must be a package-level function, not a method"}}
	}

	object := pkg.Types.Scope().Lookup("NewDelegate")
	function, ok := object.(*types.Func)
	if !ok {
		return []Diagnostic{{Position: position, Message: "NewDelegate could not be type checked as a package-level function"}}
	}
	signature, ok := function.Type().(*types.Signature)
	if !ok {
		return []Diagnostic{{Position: position, Message: "NewDelegate has an invalid function type"}}
	}
	if signature.Params().Len() != 0 {
		return []Diagnostic{{Position: position, Message: "NewDelegate must not accept parameters"}}
	}
	if signature.Results().Len() != 1 {
		return []Diagnostic{{Position: position, Message: "NewDelegate must return exactly one value"}}
	}
	if signature.TypeParams() != nil && signature.TypeParams().Len() != 0 {
		return []Diagnostic{{Position: position, Message: "NewDelegate must not declare type parameters"}}
	}

	delegateInterface := gameDelegateInterface(pkg)
	if delegateInterface == nil {
		return []Diagnostic{{Position: position, Message: "could not load boardgame.GameDelegate for contract validation"}}
	}
	returned := signature.Results().At(0).Type()
	if !types.AssignableTo(returned, delegateInterface) {
		return []Diagnostic{{
			Position: position,
			Message:  fmt.Sprintf("NewDelegate returns %s, which does not implement boardgame.GameDelegate", types.TypeString(returned, qualifier)),
		}}
	}
	return nil
}

func gameDelegateInterface(pkg *packages.Package) *types.Interface {
	boardgamePackage := findImportedPackage(pkg, boardgameImportPath, make(map[string]bool))
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

func findImportedPackage(pkg *packages.Package, importPath string, visited map[string]bool) *packages.Package {
	if pkg == nil || visited[pkg.ID] {
		return nil
	}
	visited[pkg.ID] = true
	if pkg.PkgPath == importPath {
		return pkg
	}
	for _, imported := range pkg.Imports {
		if found := findImportedPackage(imported, importPath, visited); found != nil {
			return found
		}
	}
	return nil
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
		key := fmt.Sprintf("%s:%d:%d:%s", diagnostic.Position.File, diagnostic.Position.Line, diagnostic.Position.Column, diagnostic.Message)
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
		return left.Message < right.Message
	})
	return result
}
