package lint

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
)

// The declarativeness lint's codes continue the series after the preflight's
// (fabric/TRANSPILE-PROPOSAL.md, "The declarativeness lint").
const (
	// CodeNeedsTranslation is a function reaching engine output that the
	// translator would take: it is code, not data.
	CodeNeedsTranslation = "BGLINT0101"
	// CodeRefused is a function the translator would refuse, outside the
	// subset or blocked on an engine feature.
	CodeRefused = "BGLINT0102"
	// CodeNondeterministic is a construct with no single answer in Go: a
	// range over a map, the global math/rand, enum ValueFromString.
	CodeNondeterministic = "BGLINT0103"
	// CodeStaleGenerated is generated TypeScript whose header names a Go
	// source hash the function no longer has. Reported only when a
	// generated bodies file exists.
	CodeStaleGenerated = "BGLINT0104"
)

// ExampleGames are the games under examples/ the proposal's baseline table
// counts, in its row order.
var ExampleGames = []string{"memory", "blackjack", "pig", "tictactoe", "werewolf", "debuganimations", "checkers"}

// DeclarativeReportVersion is DeclarativeReport.Version.
const DeclarativeReportVersion = 1

// DeclarativeOptions controls the declarativeness lint.
type DeclarativeOptions struct {
	// BasePath resolves relative package paths; empty is the current directory.
	BasePath string
	// MaxRefused fails the report when more than this many functions, across
	// all packages, are refused (Verdict "refused": outside the subset or
	// needing an engine feature). Negative means no limit.
	MaxRefused int
}

// Baseline is one row of the proposal's baseline table, measured.
type Baseline struct {
	MoveConfigs               int `json:"moveConfigs"`
	DataOnly                  int `json:"dataOnly"`
	WithGameCode              int `json:"withGameCode"`
	DeclaredPreconditionSpecs int `json:"declaredPreconditionSpecs"`
	Functions                 int `json:"functions"`
	InSubset                  int `json:"inSubset"`
	Outside                   int `json:"outside"`
	NeedEngineFeature         int `json:"needEngineFeature"`
	NotPorted                 int `json:"notPorted"`
}

func (b *Baseline) add(other Baseline) {
	b.MoveConfigs += other.MoveConfigs
	b.DataOnly += other.DataOnly
	b.WithGameCode += other.WithGameCode
	b.DeclaredPreconditionSpecs += other.DeclaredPreconditionSpecs
	b.Functions += other.Functions
	b.InSubset += other.InSubset
	b.Outside += other.Outside
	b.NeedEngineFeature += other.NeedEngineFeature
	b.NotPorted += other.NotPorted
}

// Shares are the report's totals as fractions, "n/d".
type Shares struct {
	// DataOnly is move configs that are data, of all move configs.
	DataOnly string `json:"dataOnly"`
	// DeclaredLegality is move configs whose legality has no imperative
	// Legal override (stock checks, declared specs, and at most a
	// LegalCustom residue), of all move configs.
	DeclaredLegality string `json:"declaredLegality"`
	// InSubset is functions in the subset, of all counted functions.
	InSubset string `json:"inSubset"`
}

// MoveRow is one move config.
type MoveRow struct {
	Type              string   `json:"type"`
	MoveName          string   `json:"moveName,omitempty"`
	MoveNameSuffix    string   `json:"moveNameSuffix,omitempty"`
	ExpandedFrom      string   `json:"expandedFrom,omitempty"`
	Stock             string   `json:"stock"`
	StockPorted       bool     `json:"stockPorted"`
	Data              bool     `json:"data"`
	Methods           []string `json:"methods"`
	PreconditionSpecs int      `json:"preconditionSpecs"`
	ImperativeLegal   bool     `json:"imperativeLegal"`
	LegalCustom       bool     `json:"legalCustom"`
	File              string   `json:"file"`
	Line              int      `json:"line"`
}

// ImperativeLegalRow is a move config whose legality is a Legal override.
type ImperativeLegalRow struct {
	Type  string `json:"type"`
	Move  string `json:"move,omitempty"`
	Legal string `json:"legal"`
	Note  string `json:"note"`
}

// FunctionRow is one function reaching engine output.
type FunctionRow struct {
	Func       string            `json:"func"`
	Role       string            `json:"role"`
	Bucket     string            `json:"bucket"`
	Verdict    string            `json:"verdict"`
	File       string            `json:"file"`
	Line       int               `json:"line"`
	SourceHash string            `json:"sourceHash"`
	Calls      []string          `json:"calls"`
	Refusals   []refusal.Refusal `json:"refusals"`
}

// WarningRow is one nondeterministic construct.
type WarningRow struct {
	Func   string `json:"func"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// OtherFunction is a game function outside the counted ones.
type OtherFunction struct {
	Func string `json:"func"`
	Why  string `json:"why"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// DeclarativePackage is the declarativeness report for one game package.
type DeclarativePackage struct {
	Package  string   `json:"package"`
	Name     string   `json:"name"`
	Baseline Baseline `json:"baseline"`
	Shares   Shares   `json:"shares"`
	// Buckets counts functions by gamedef bucket.
	Buckets         map[string]int       `json:"buckets"`
	Moves           []MoveRow            `json:"moves"`
	ImperativeLegal []ImperativeLegalRow `json:"imperativeLegal"`
	// Functions are callee first: every in-package callee before its callers.
	Functions  []FunctionRow   `json:"functions"`
	Warnings   []WarningRow    `json:"warnings"`
	NotPorted  []OtherFunction `json:"notPorted"`
	NotCounted []OtherFunction `json:"notCounted"`
	Erased     []OtherFunction `json:"erased"`
	Unreached  []OtherFunction `json:"unreached"`
}

// DeclarativeReport is the whole report.
type DeclarativeReport struct {
	Version int `json:"version"`
	// OK is false when a package failed to load or MaxRefused was exceeded.
	OK          bool                 `json:"ok"`
	Packages    []DeclarativePackage `json:"packages"`
	Totals      Baseline             `json:"totals"`
	Refused     int                  `json:"refused"`
	MaxRefused  *int                 `json:"maxRefused,omitempty"`
	Diagnostics []Diagnostic         `json:"diagnostics"`
}

// Declarative runs the declarativeness lint over the game packages and
// patterns in inputs (empty means the current package). It writes nothing.
func Declarative(inputs []string, options DeclarativeOptions) DeclarativeReport {
	if len(inputs) == 0 {
		inputs = []string{"."}
	}
	report := DeclarativeReport{Version: DeclarativeReportVersion, OK: true, Packages: []DeclarativePackage{}, Diagnostics: []Diagnostic{}}
	packages, diagnostics := resolvePackages(inputs, options.BasePath)
	report.Diagnostics = append(report.Diagnostics, diagnostics...)
	if len(diagnostics) > 0 {
		report.OK = false
	}
	for _, pkg := range packages {
		static, err := gamedef.LoadStatic(".", pkg.AbsolutePath())
		if err != nil {
			report.OK = false
			report.Diagnostics = append(report.Diagnostics, Diagnostic{
				Package: pkg.Import(), Code: CodePackage,
				Message:     "could not analyze game package source: " + err.Error(),
				Remediation: "Fix the package's type errors, then rerun boardgame-util lint --declarative.",
			})
			continue
		}
		result, packageDiagnostics := DeclarativeFor(static)
		report.Packages = append(report.Packages, result)
		report.Diagnostics = append(report.Diagnostics, packageDiagnostics...)
		report.Diagnostics = append(report.Diagnostics, staleGeneratedDiagnostics(static)...)
		report.Totals.add(result.Baseline)
		for _, fn := range result.Functions {
			if fn.Verdict == gamedef.VerdictRefused {
				report.Refused++
			}
		}
	}
	if options.MaxRefused >= 0 {
		limit := options.MaxRefused
		report.MaxRefused = &limit
		if report.Refused > limit {
			report.OK = false
		}
	}
	sortDiagnostics(report.Diagnostics)
	return report
}

// DeclarativeFor builds one package's report and diagnostics from its
// static half.
func DeclarativeFor(static *gamedef.StaticReport) (DeclarativePackage, []Diagnostic) {
	result := DeclarativePackage{
		Package: static.Package, Name: static.PackageName,
		Buckets:         map[string]int{gamedef.BucketSubset: 0, gamedef.BucketEngineGap: 0, gamedef.BucketOutside: 0},
		Moves:           []MoveRow{},
		ImperativeLegal: []ImperativeLegalRow{},
		Functions:       []FunctionRow{},
		Warnings:        []WarningRow{},
	}
	var diagnostics []Diagnostic
	declaredLegality := 0
	for _, config := range static.MoveConfigs {
		row := MoveRow{
			Type: config.GoType, MoveName: config.MoveName, MoveNameSuffix: config.MoveNameSuffix,
			ExpandedFrom: config.ExpandedFrom, Stock: config.StockType, StockPorted: config.StockPorted,
			Data: config.DataOnly, Methods: config.Methods, PreconditionSpecs: config.PreconditionSpecs,
			ImperativeLegal: config.Bodies.Legal != "", LegalCustom: config.Bodies.LegalCustom != "",
			File: config.File, Line: config.Line,
		}
		if row.Methods == nil {
			row.Methods = []string{}
		}
		result.Moves = append(result.Moves, row)
		result.Baseline.MoveConfigs++
		if row.Data {
			result.Baseline.DataOnly++
		} else {
			result.Baseline.WithGameCode++
		}
		if config.PreconditionSpecs > 0 {
			result.Baseline.DeclaredPreconditionSpecs += config.PreconditionSpecs
		}
		if row.ImperativeLegal {
			move := config.MoveName
			if move == "" && config.MoveNameSuffix != "" {
				move = "(suffix " + config.MoveNameSuffix + ")"
			}
			result.ImperativeLegal = append(result.ImperativeLegal, ImperativeLegalRow{
				Type: config.GoType, Move: move, Legal: config.Bodies.Legal,
				Note: "legality is an imperative Legal override: its checks are invisible to viewers and to the declared plan",
			})
		} else {
			declaredLegality++
		}
	}

	byFunc := make(map[string]gamedef.FunctionEntry, len(static.Functions))
	for _, entry := range static.Functions {
		byFunc[entry.Func] = entry
	}
	for _, ref := range static.Order {
		entry := byFunc[ref]
		bucket := static.Buckets[ref]
		result.Buckets[bucket]++
		result.Functions = append(result.Functions, FunctionRow{
			Func: entry.Func, Role: entry.Role, Bucket: bucket, Verdict: entry.Verdict,
			File: entry.File, Line: entry.Line, SourceHash: entry.SourceHash,
			Calls: entry.Calls, Refusals: entry.Refusals,
		})
		diagnostics = append(diagnostics, functionDiagnostic(static.Package, entry, bucket))
	}
	result.Baseline.Functions = len(static.Functions)
	result.Baseline.Outside = result.Buckets[gamedef.BucketOutside]
	result.Baseline.NeedEngineFeature = result.Buckets[gamedef.BucketEngineGap]
	result.Baseline.InSubset = result.Baseline.Functions - result.Baseline.Outside
	result.Baseline.NotPorted = len(static.NotPorted)

	for _, finding := range static.Findings {
		result.Warnings = append(result.Warnings, WarningRow{
			Func: finding.Func, Kind: finding.Kind, Detail: finding.Detail,
			File: finding.File, Line: finding.Line, Column: finding.Column,
		})
		diagnostics = append(diagnostics, Diagnostic{
			Package: static.Package, Code: CodeNondeterministic,
			File: finding.File, Line: finding.Line, Column: finding.Column,
			Message:     fmt.Sprintf("%s in %s: %s has no single answer in Go", finding.Kind, finding.Func, finding.Detail),
			Remediation: nondeterminismRemediation[finding.Kind],
		})
	}
	result.NotPorted = otherFunctions(static.NotPorted)
	result.NotCounted = otherFunctions(static.NotCounted)
	result.Erased = otherFunctions(static.Erased)
	result.Unreached = otherFunctions(static.Unreached)
	result.Shares = Shares{
		DataOnly:         fmt.Sprintf("%d/%d", result.Baseline.DataOnly, result.Baseline.MoveConfigs),
		DeclaredLegality: fmt.Sprintf("%d/%d", declaredLegality, result.Baseline.MoveConfigs),
		InSubset:         fmt.Sprintf("%d/%d", result.Baseline.InSubset, result.Baseline.Functions),
	}
	return result, diagnostics
}

var nondeterminismRemediation = map[string]string{
	"map-range":         "Range over a sorted slice of the map's keys instead.",
	"global-rand":       "Use state.Rand() in game logic; code the engine does not run (an agent) cannot be replayed while it draws from math/rand.",
	"value-from-string": "Compare enum keys, not normalized strings.",
}

func otherFunctions(refs []gamedef.FunctionRef) []OtherFunction {
	result := make([]OtherFunction, 0, len(refs))
	for _, ref := range refs {
		result = append(result, OtherFunction{Func: ref.Func, Why: ref.Why, File: ref.File, Line: ref.Line})
	}
	return result
}

func functionDiagnostic(pkg string, entry gamedef.FunctionEntry, bucket string) Diagnostic {
	if entry.Verdict == gamedef.VerdictSubset {
		return Diagnostic{
			Package: pkg, Code: CodeNeedsTranslation, File: entry.File, Line: entry.Line,
			Message: fmt.Sprintf("%s (%s) reaches engine output and needs translation; it is in the subset", entry.Func, entry.Role),
		}
	}
	var reasons []string
	for _, r := range entry.Refusals {
		reasons = append(reasons, r.Reason+" "+r.Detail)
	}
	what := "is outside the subset"
	remediation := "An agent translates this function; or rewrite the Go so each refusal no longer applies."
	if bucket == gamedef.BucketEngineGap {
		what = "needs an engine feature TypeScript lacks"
		remediation = "Port the missing engine feature; the function itself is translatable."
	}
	return Diagnostic{
		Package: pkg, Code: CodeRefused, File: entry.File, Line: entry.Line,
		Message:     fmt.Sprintf("%s (%s) %s: %s", entry.Func, entry.Role, what, strings.Join(reasons, "; ")),
		Remediation: remediation,
	}
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		left, right := diagnostics[i], diagnostics[j]
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
}

// GeneratedBodiesPath is where a generated game's translated bodies live,
// relative to the module root: games/<game>/bodies.gen.ts in the engine.
func GeneratedBodiesPath(gameName string) string {
	return filepath.Join("server", "static", "src", "engine", "games", gameName, "bodies.gen.ts")
}

var generatedHeader = regexp.MustCompile(`^//\s*@generated\b.*\bgo:(\S+)\s+(sha256:[0-9a-f]{64})\b`)

// GeneratedHeader is one generated function's header line.
type GeneratedHeader struct {
	Func       string
	SourceHash string
	Line       int
}

// ParseGeneratedHeaders reads the "// @generated ... go:<FnRef> sha256:<hash>"
// headers of a generated bodies file.
func ParseGeneratedHeaders(r io.Reader) ([]GeneratedHeader, error) {
	var headers []GeneratedHeader
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		if match := generatedHeader.FindStringSubmatch(strings.TrimSpace(scanner.Text())); match != nil {
			headers = append(headers, GeneratedHeader{Func: match[1], SourceHash: match[2], Line: line})
		}
	}
	return headers, scanner.Err()
}

// StaleGenerated compares generated headers with the functions' current
// source hashes: a header is stale when its function's hash differs or the
// function no longer reaches engine output.
func StaleGenerated(pkg, file string, headers []GeneratedHeader, functions []gamedef.FunctionEntry) []Diagnostic {
	current := make(map[string]string, len(functions))
	for _, fn := range functions {
		current[fn.Func] = fn.SourceHash
	}
	var diagnostics []Diagnostic
	for _, header := range headers {
		hash, ok := current[header.Func]
		if ok && hash == header.SourceHash {
			continue
		}
		message := fmt.Sprintf("generated TypeScript for %s was translated from Go source %s; the source is now %s", header.Func, header.SourceHash, hash)
		if !ok {
			message = fmt.Sprintf("generated TypeScript for %s names a function that no longer reaches engine output", header.Func)
		}
		diagnostics = append(diagnostics, Diagnostic{
			Package: pkg, Code: CodeStaleGenerated, File: file, Line: header.Line,
			Message:     message,
			Remediation: "Regenerate the game's TypeScript bodies (boardgame-util transpile) and review the change.",
		})
	}
	return diagnostics
}

// staleGeneratedDiagnostics checks the game's generated bodies file, when
// one exists under the module root.
func staleGeneratedDiagnostics(static *gamedef.StaticReport) []Diagnostic {
	root := moduleRoot(static.Dir)
	if root == "" {
		return nil
	}
	path := filepath.Join(root, GeneratedBodiesPath(static.PackageName))
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	headers, err := ParseGeneratedHeaders(file)
	if err != nil {
		return []Diagnostic{{Package: static.Package, Code: CodeStaleGenerated, File: path, Message: "could not read generated bodies: " + err.Error()}}
	}
	return StaleGenerated(static.Package, path, headers, static.Functions)
}

func moduleRoot(dir string) string {
	for dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

// MarshalDeclarativePackage is one package's report as indented JSON with a
// trailing newline: the form of the checked-in goldens
// (testdata/declarative/<game>.json).
func MarshalDeclarativePackage(pkg DeclarativePackage) ([]byte, error) {
	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// WriteDeclarativeText writes the report for a person.
func WriteDeclarativeText(w io.Writer, report DeclarativeReport) {
	for _, pkg := range report.Packages {
		b := pkg.Baseline
		fmt.Fprintf(w, "%s (%s)\n", pkg.Name, pkg.Package)
		fmt.Fprintf(w, "  move configs %d, data only %d, with game code %d, declared precondition specs %d\n",
			b.MoveConfigs, b.DataOnly, b.WithGameCode, b.DeclaredPreconditionSpecs)
		fmt.Fprintf(w, "  functions reaching engine output %d: in the subset %d, outside it %d, needing an engine feature %d; not ported %d\n",
			b.Functions, b.InSubset, b.Outside, b.NeedEngineFeature, b.NotPorted)
		fmt.Fprintf(w, "  shares: data %s, declared legality %s, in subset %s\n", pkg.Shares.DataOnly, pkg.Shares.DeclaredLegality, pkg.Shares.InSubset)
		fmt.Fprintln(w, "  moves:")
		for _, move := range pkg.Moves {
			label := move.Type
			if move.MoveName != "" {
				label += " \"" + move.MoveName + "\""
			}
			if move.MoveNameSuffix != "" {
				label += " +\"" + move.MoveNameSuffix + "\""
			}
			if move.ExpandedFrom != "" {
				label += " (from " + move.ExpandedFrom + ")"
			}
			what := "data"
			if !move.Data {
				what = strings.Join(move.Methods, ", ")
			}
			stock := move.Stock
			if stock == "" {
				stock = "Default"
			}
			if !move.StockPorted {
				stock += ", not ported"
			}
			fmt.Fprintf(w, "    %s [%s] %s; %d specs\n", label, stock, what, move.PreconditionSpecs)
		}
		if len(pkg.ImperativeLegal) > 0 {
			fmt.Fprintln(w, "  imperative Legal overrides (checks invisible to viewers):")
			for _, row := range pkg.ImperativeLegal {
				fmt.Fprintf(w, "    %s %s\n", row.Type, row.Legal)
			}
		}
		fmt.Fprintln(w, "  functions (callee first):")
		for _, fn := range pkg.Functions {
			fmt.Fprintf(w, "    %s %s [%s] %s:%d\n", fn.Bucket, fn.Func, fn.Role, fn.File, fn.Line)
			for _, r := range fn.Refusals {
				fmt.Fprintf(w, "      %s %s at %s:%d:%d\n", r.Reason, r.Detail, r.File, r.Line, r.Column)
			}
		}
		for _, warning := range pkg.Warnings {
			fmt.Fprintf(w, "  warning: %s %s in %s at %s:%d:%d\n", warning.Kind, warning.Detail, warning.Func, warning.File, warning.Line, warning.Column)
		}
		if len(pkg.NotPorted) > 0 {
			var names []string
			for _, fn := range pkg.NotPorted {
				names = append(names, fn.Func)
			}
			fmt.Fprintf(w, "  not ported: %s\n", strings.Join(names, ", "))
		}
		if len(pkg.Unreached) > 0 {
			var names []string
			for _, fn := range pkg.Unreached {
				names = append(names, fn.Func)
			}
			fmt.Fprintf(w, "  unreached: %s\n", strings.Join(names, ", "))
		}
	}
	t := report.Totals
	fmt.Fprintf(w, "totals: move configs %d, data only %d, with game code %d, declared precondition specs %d, functions %d, in the subset %d, outside it %d, needing an engine feature %d, not ported %d\n",
		t.MoveConfigs, t.DataOnly, t.WithGameCode, t.DeclaredPreconditionSpecs, t.Functions, t.InSubset, t.Outside, t.NeedEngineFeature, t.NotPorted)
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == CodeNeedsTranslation || diagnostic.Code == CodeRefused {
			continue // listed above
		}
		fmt.Fprintf(w, "%s %s: %s\n", diagnostic.Code, diagnostic.Package, diagnostic.Message)
	}
	if report.MaxRefused != nil {
		fmt.Fprintf(w, "refused functions: %d (limit %d)\n", report.Refused, *report.MaxRefused)
	}
}
