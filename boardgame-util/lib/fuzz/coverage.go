package fuzz

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance/corpus"
)

// Coverage, as fabric/TRANSPILE-PROPOSAL.md "Coverage" defines it, per game.
// Three of its four measures are read from documents alone (Coverage): the
// moves applied and refused, the verdicts each declared precondition
// returned, and whether each property a viewer is not always shown was
// sanitized for a non-owner while it held a non-zero value. The fourth, Go
// block coverage of the game package, comes from a profile of a
// `go build -cover` binary (BuildBlockProfile, BlockCoverage).

// Report is one game's coverage.
type Report struct {
	Game       string `json:"game"`
	Documents  int    `json:"documents"`
	Versions   int    `json:"versions"`
	Rejections int    `json:"rejections"`
	// Moves has one row per move type, in move-table order.
	Moves []MoveCoverage `json:"moves"`
	// Predicates has one row per precondition of each declared plan, in
	// move-table order and then the plan's order.
	Predicates []PredicateCoverage `json:"predicates"`
	// Visibility has one row per property some viewer other than the admin
	// was, at some version, not shown the values of.
	Visibility []VisibilityCoverage `json:"visibility"`
	// Blocks is per-function block coverage of the game's Go package, when
	// a profile was given.
	Blocks []FuncCoverage `json:"blocks,omitempty"`
}

// MoveCoverage counts one move type across the documents.
type MoveCoverage struct {
	Name    string `json:"name"`
	IsFixUp bool   `json:"isFixUp"`
	// Applied counts versions this move produced, whatever the trigger.
	Applied int `json:"applied"`
	// Proposed counts the outside proposals of it that were accepted
	// (versions with trigger "propose").
	Proposed int `json:"proposed"`
	// Refused counts the rejections that name it.
	Refused int `json:"refused"`
}

// PredicateCoverage is one precondition of a move's declared plan and the
// verdicts recorded for it, over every viewer of every version.
type PredicateCoverage struct {
	Move    string   `json:"move"`
	Index   int      `json:"index"`
	Name    string   `json:"name"`
	Args    []string `json:"args,omitempty"`
	Pass    int      `json:"pass"`
	Fail    int      `json:"fail"`
	Unknown int      `json:"unknown"`
}

// Missing lists the verdicts never recorded.
func (p PredicateCoverage) Missing() []string {
	var missing []string
	if p.Pass == 0 {
		missing = append(missing, "pass")
	}
	if p.Fail == 0 {
		missing = append(missing, "fail")
	}
	if p.Unknown == 0 {
		missing = append(missing, "unknown")
	}
	return missing
}

// VisibilityCoverage is one property whose values some non-admin viewer was
// not shown. Property is "Game.<name>" or "Players.<name>"; a player
// property is owned by its player, a game property by no one.
type VisibilityCoverage struct {
	Property string `json:"property"`
	// SanitizedNonZero counts (version, viewer) pairs where a viewer other
	// than the owner and the admin was sent the property without its values
	// while it held a non-zero value: the case that shows the policy hides
	// something. Zero is the gap.
	SanitizedNonZero int `json:"sanitizedNonZero"`
	// RevealedNonZero counts the pairs where such a viewer was shown its
	// values while it was non-zero (a group or self policy at work).
	RevealedNonZero int `json:"revealedNonZero"`
	// Facets lists, sorted, every facet set a non-owner was sent.
	Facets []string `json:"facets"`
}

// FuncCoverage is one function of the game package and its blocks.
type FuncCoverage struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Func      string `json:"func"`
	Generated bool   `json:"generated,omitempty"`
	Blocks    int    `json:"blocks"`
	Reached   int    `json:"reached"`
	// Unreached lists each unreached block as "line.col-line.col".
	Unreached []string `json:"unreached,omitempty"`
}

// LoadDocuments reads every <game>/seed-*.json.gz under dir, in name
// order. With game empty it reads every game's.
func LoadDocuments(dir, game string) (map[string][]*conformance.Document, error) {
	pattern := filepath.Join(dir, "*", "seed-*.json.gz")
	if game != "" {
		pattern = filepath.Join(dir, game, "seed-*.json.gz")
	}
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	out := map[string][]*conformance.Document{}
	for _, path := range paths {
		blob, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if blob, err = corpus.Decompress(blob); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		var doc conformance.Document
		if err := json.Unmarshal(blob, &doc); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		name := filepath.Base(filepath.Dir(path))
		out[name] = append(out[name], &doc)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no documents match %s", pattern)
	}
	return out, nil
}

type precondition struct {
	Name    string   `json:"name"`
	Args    []string `json:"args"`
	Verdict string   `json:"verdict"`
}

// Coverage reads the three document measures from docs, which must all be
// of one game.
func Coverage(docs []*conformance.Document) (Report, error) {
	var report Report
	if len(docs) == 0 {
		return report, fmt.Errorf("no documents")
	}
	report.Game = docs[0].Game.Name
	report.Documents = len(docs)

	var moveRows []*MoveCoverage
	moves := map[string]*MoveCoverage{}
	moveRow := func(name string) *MoveCoverage {
		if row, ok := moves[name]; ok {
			return row
		}
		row := &MoveCoverage{Name: name}
		moveRows = append(moveRows, row)
		moves[name] = row
		return row
	}
	for _, move := range docs[0].Moves {
		moveRow(move.Name).IsFixUp = move.IsFixUp
	}
	moveOrder := map[string]int{}
	for i, move := range docs[0].Moves {
		moveOrder[move.Name] = i
	}

	type predicateKey struct {
		move  string
		index int
		name  string
		args  string
	}
	predicates := map[predicateKey]*PredicateCoverage{}
	visibility := map[string]*VisibilityCoverage{}
	facetSets := map[string]map[string]bool{}
	// sanitized holds the properties some non-owner, non-admin viewer was
	// not shown the values of at some version: those with a policy other
	// than visible. Only they are reported.
	sanitized := map[string]bool{}

	for _, doc := range docs {
		if doc.Game.Name != report.Game {
			return report, fmt.Errorf("%s is a %s document, not %s", doc.Name, doc.Game.Name, report.Game)
		}
		report.Versions += len(doc.Versions)
		report.Rejections += len(doc.Rejections)
		for _, rejection := range doc.Rejections {
			if rejection.Name != "" {
				moveRow(rejection.Name).Refused++
			}
		}
		for _, version := range doc.Versions {
			if version.Move != nil {
				row := moveRow(version.Move.Name)
				row.Applied++
				if version.Move.Trigger == "propose" {
					row.Proposed++
				}
			}
			raw, err := rawState(version.State)
			if err != nil {
				return report, fmt.Errorf("%s v%d: %w", doc.Name, version.Version, err)
			}
			for _, view := range version.Views {
				for _, legality := range view.Legality {
					if len(legality.Preconditions) == 0 {
						continue
					}
					var list []precondition
					if err := json.Unmarshal(legality.Preconditions, &list); err != nil {
						return report, fmt.Errorf("%s v%d %s: preconditions: %w", doc.Name, version.Version, legality.Name, err)
					}
					for i, p := range list {
						key := predicateKey{legality.Name, i, p.Name, strings.Join(p.Args, "\x00")}
						row := predicates[key]
						if row == nil {
							row = &PredicateCoverage{Move: legality.Name, Index: i, Name: p.Name, Args: p.Args}
							predicates[key] = row
						}
						switch p.Verdict {
						case "pass":
							row.Pass++
						case "fail":
							row.Fail++
						case "unknown":
							row.Unknown++
						default:
							return report, fmt.Errorf("%s v%d %s: unknown verdict %q", doc.Name, version.Version, legality.Name, p.Verdict)
						}
					}
				}
				if view.Viewer == boardgameAdmin {
					continue
				}
				facets, err := viewFacets(view.Game)
				if err != nil {
					return report, fmt.Errorf("%s v%d viewer %d: %w", doc.Name, version.Version, view.Viewer, err)
				}
				for _, f := range facets {
					if f.owner >= 0 && f.owner == int(view.Viewer) {
						continue
					}
					shown := false
					for _, facet := range f.facets {
						if facet == "values" {
							shown = true
						}
					}
					key := f.scope + "." + f.name
					row := visibility[key]
					if row == nil {
						row = &VisibilityCoverage{Property: key}
						visibility[key] = row
						facetSets[key] = map[string]bool{}
					}
					if !shown {
						sanitized[key] = true
					}
					facetSets[key][strings.Join(f.facets, ",")] = true
					if isZero(raw.value(f.scope, f.owner, f.name)) {
						continue
					}
					if shown {
						row.RevealedNonZero++
					} else {
						row.SanitizedNonZero++
					}
				}
			}
		}
	}
	for _, row := range moveRows {
		report.Moves = append(report.Moves, *row)
	}
	for _, row := range predicates {
		report.Predicates = append(report.Predicates, *row)
	}
	sort.SliceStable(report.Predicates, func(i, j int) bool {
		a, b := report.Predicates[i], report.Predicates[j]
		if moveOrder[a.Move] != moveOrder[b.Move] {
			return moveOrder[a.Move] < moveOrder[b.Move]
		}
		if a.Index != b.Index {
			return a.Index < b.Index
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return strings.Join(a.Args, ",") < strings.Join(b.Args, ",")
	})
	for key, row := range visibility {
		if !sanitized[key] {
			continue
		}
		for set := range facetSets[key] {
			row.Facets = append(row.Facets, "["+set+"]")
		}
		sort.Strings(row.Facets)
		report.Visibility = append(report.Visibility, *row)
	}
	sort.Slice(report.Visibility, func(i, j int) bool { return report.Visibility[i].Property < report.Visibility[j].Property })
	return report, nil
}

// boardgameAdmin is boardgame.AdminPlayerIndex, which sees everything.
const boardgameAdmin = -2

type stateJSON struct {
	Game    map[string]json.RawMessage   `json:"Game"`
	Players []map[string]json.RawMessage `json:"Players"`
}

func rawState(blob json.RawMessage) (*stateJSON, error) {
	var state stateJSON
	if err := json.Unmarshal(blob, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *stateJSON) value(scope string, owner int, name string) json.RawMessage {
	if scope == "Game" {
		return s.Game[name]
	}
	if owner < 0 || owner >= len(s.Players) {
		return nil
	}
	return s.Players[owner][name]
}

type propertyFacets struct {
	scope  string
	owner  int
	name   string
	facets []string
}

// viewFacets reads CurrentState.Visibility from a view's game object.
func viewFacets(game json.RawMessage) ([]propertyFacets, error) {
	var parsed struct {
		CurrentState struct {
			Visibility struct {
				Game    map[string][]string   `json:"Game"`
				Players []map[string][]string `json:"Players"`
			} `json:"Visibility"`
		} `json:"CurrentState"`
	}
	if err := json.Unmarshal(game, &parsed); err != nil {
		return nil, err
	}
	var out []propertyFacets
	for name, facets := range parsed.CurrentState.Visibility.Game {
		out = append(out, propertyFacets{scope: "Game", owner: -1, name: name, facets: facets})
	}
	for i, props := range parsed.CurrentState.Visibility.Players {
		for name, facets := range props {
			out = append(out, propertyFacets{scope: "Players", owner: i, name: name, facets: facets})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].scope != out[j].scope {
			return out[i].scope < out[j].scope
		}
		if out[i].owner != out[j].owner {
			return out[i].owner < out[j].owner
		}
		return out[i].name < out[j].name
	})
	return out, nil
}

// isZero reports whether a stored property holds Go's zero value: null,
// false, 0, "", an empty list, a stack with no component (every index
// negative), a timer with no id, or an object of zero values.
func isZero(blob json.RawMessage) bool {
	if len(blob) == 0 {
		return true
	}
	var value interface{}
	if err := json.Unmarshal(blob, &value); err != nil {
		return false
	}
	return zeroValue(value)
}

func zeroValue(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case bool:
		return !v
	case float64:
		return v == 0
	case string:
		return v == ""
	case []interface{}:
		for _, item := range v {
			if !zeroValue(item) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		if indexes, ok := v["Indexes"].([]interface{}); ok {
			for _, index := range indexes {
				if n, ok := index.(float64); ok && n >= 0 {
					return false
				}
			}
			return true
		}
		if isTimer, ok := v["IsTimer"].(bool); ok && isTimer {
			id, _ := v["ID"].(string)
			return id == ""
		}
		for _, item := range v {
			if !zeroValue(item) {
				return false
			}
		}
		return true
	}
	return false
}

// Gaps are the rows the report flags.
type Gaps struct {
	NeverApplied []string `json:"neverApplied"`
	// NeverRefused lists non-fix-up moves no rejection names.
	NeverRefused []string `json:"neverRefused"`
	// Predicates lists preconditions missing a verdict, with the missing ones.
	Predicates []string `json:"predicates"`
	// Visibility lists properties never sanitized for a non-owner while non-zero.
	Visibility []string `json:"visibility"`
	// Functions lists hand-written functions of the game package with an
	// unreached block.
	Functions []string `json:"functions,omitempty"`
}

// Gaps lists what the report flags.
func (r Report) Gaps() Gaps {
	var gaps Gaps
	for _, move := range r.Moves {
		if move.Applied == 0 {
			gaps.NeverApplied = append(gaps.NeverApplied, move.Name)
		}
		if !move.IsFixUp && move.Refused == 0 {
			gaps.NeverRefused = append(gaps.NeverRefused, move.Name)
		}
	}
	for _, p := range r.Predicates {
		if missing := p.Missing(); len(missing) > 0 {
			gaps.Predicates = append(gaps.Predicates, p.label()+": no "+strings.Join(missing, ", "))
		}
	}
	for _, v := range r.Visibility {
		if v.SanitizedNonZero == 0 {
			gaps.Visibility = append(gaps.Visibility, v.Property)
		}
	}
	for _, f := range r.Blocks {
		if !f.Generated && f.Reached < f.Blocks {
			gaps.Functions = append(gaps.Functions, fmt.Sprintf("%s (%s:%d): %d of %d blocks unreached", f.Func, f.File, f.Line, f.Blocks-f.Reached, f.Blocks))
		}
	}
	return gaps
}

func (p PredicateCoverage) label() string {
	return fmt.Sprintf("%s #%d %s(%s)", p.Move, p.Index, p.Name, strings.Join(p.Args, ", "))
}

// Text renders the report as Markdown tables.
func (r Report) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n\n%d documents, %d versions, %d rejections.\n\n", r.Game, r.Documents, r.Versions, r.Rejections)
	b.WriteString("### Moves\n\n| Move | Fix-up | Applied | Proposed | Refused |\n|---|---|---|---|---|\n")
	for _, m := range r.Moves {
		fixUp := ""
		if m.IsFixUp {
			fixUp = "yes"
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d |\n", m.Name, fixUp, m.Applied, m.Proposed, m.Refused)
	}
	b.WriteString("\n### Predicates\n\n")
	if len(r.Predicates) == 0 {
		b.WriteString("No move declares a legal plan.\n")
	} else {
		b.WriteString("| Move | # | Predicate | Pass | Fail | Unknown | Missing |\n|---|---|---|---|---|---|---|\n")
		for _, p := range r.Predicates {
			fmt.Fprintf(&b, "| %s | %d | %s(%s) | %d | %d | %d | %s |\n", p.Move, p.Index, p.Name, strings.Join(p.Args, ", "), p.Pass, p.Fail, p.Unknown, strings.Join(p.Missing(), ", "))
		}
	}
	b.WriteString("\n### Visibility\n\n")
	if len(r.Visibility) == 0 {
		b.WriteString("Every property is shown to every viewer.\n")
	} else {
		b.WriteString("| Property | Sanitized for a non-owner while non-zero | Shown to a non-owner while non-zero | Facets non-owners got |\n|---|---|---|---|\n")
		for _, v := range r.Visibility {
			fmt.Fprintf(&b, "| %s | %d | %d | %s |\n", v.Property, v.SanitizedNonZero, v.RevealedNonZero, strings.Join(v.Facets, " "))
		}
	}
	if r.Blocks != nil {
		b.WriteString("\n### Go blocks (hand-written functions with an unreached block)\n\n")
		total, reached, unreachedFuncs := 0, 0, 0
		var rows []string
		for _, f := range r.Blocks {
			if f.Generated {
				continue
			}
			total += f.Blocks
			reached += f.Reached
			if f.Reached < f.Blocks {
				unreachedFuncs++
				rows = append(rows, fmt.Sprintf("| %s | %s:%d | %d/%d | %s |", f.Func, f.File, f.Line, f.Reached, f.Blocks, strings.Join(f.Unreached, " ")))
			}
		}
		fmt.Fprintf(&b, "%d of %d blocks reached in hand-written files; %d functions with an unreached block.\n\n", reached, total, unreachedFuncs)
		if len(rows) > 0 {
			b.WriteString("| Function | At | Reached | Unreached blocks |\n|---|---|---|---|\n")
			b.WriteString(strings.Join(rows, "\n") + "\n")
		}
	}
	gaps := r.Gaps()
	b.WriteString("\n### Gaps\n\n")
	fmt.Fprintf(&b, "- Never applied: %s\n", listOrNone(gaps.NeverApplied))
	fmt.Fprintf(&b, "- Never refused (non-fix-up): %s\n", listOrNone(gaps.NeverRefused))
	fmt.Fprintf(&b, "- Predicates missing a verdict: %d of %d\n", len(gaps.Predicates), len(r.Predicates))
	fmt.Fprintf(&b, "- Hidden properties never sanitized for a non-owner while non-zero: %s\n", listOrNone(gaps.Visibility))
	if r.Blocks != nil {
		fmt.Fprintf(&b, "- Hand-written functions with an unreached block: %d\n", len(gaps.Functions))
	}
	return b.String()
}

func listOrNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, "; ")
}

// ---------------------------------------------------------------------------
// Go blocks

// ProfileBlock is one block of a text coverage profile (`go tool covdata
// textfmt`, the `go test -coverprofile` format).
type ProfileBlock struct {
	File                string
	StartLine, StartCol int
	EndLine, EndCol     int
	Statements, Count   int
}

// ParseProfile reads a text coverage profile. A block listed more than once
// (several counter files) keeps its largest count.
func ParseProfile(r io.Reader) ([]ProfileBlock, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	index := map[string]int{}
	var blocks []ProfileBlock
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "mode:") {
			continue
		}
		// file:startLine.startCol,endLine.endCol statements count
		colon := strings.LastIndex(text, ":")
		if colon < 0 {
			return nil, fmt.Errorf("profile line %d: %q", line, text)
		}
		var b ProfileBlock
		b.File = text[:colon]
		if _, err := fmt.Sscanf(text[colon+1:], "%d.%d,%d.%d %d %d", &b.StartLine, &b.StartCol, &b.EndLine, &b.EndCol, &b.Statements, &b.Count); err != nil {
			return nil, fmt.Errorf("profile line %d: %q: %w", line, text, err)
		}
		key := text[:strings.LastIndex(text, " ")]
		key = key[:strings.LastIndex(key, " ")]
		if i, ok := index[key]; ok {
			if b.Count > blocks[i].Count {
				blocks[i].Count = b.Count
			}
			continue
		}
		index[key] = len(blocks)
		blocks = append(blocks, b)
	}
	return blocks, scanner.Err()
}

// ModulePath is the repository's Go module.
const ModulePath = "github.com/jkomoros/boardgame"

// GamePackage is the import path of a game's Go package.
func GamePackage(game string) string {
	return ModulePath + "/examples/" + game
}

// BlockCoverage assigns the profile's blocks of one package (files directly
// in pkgDir, imported as importPath) to the functions that contain them.
func BlockCoverage(blocks []ProfileBlock, importPath, pkgDir string) ([]FuncCoverage, error) {
	byFile := map[string][]ProfileBlock{}
	for _, b := range blocks {
		if filepath.ToSlash(filepath.Dir(b.File)) != importPath {
			continue
		}
		byFile[filepath.Base(b.File)] = append(byFile[filepath.Base(b.File)], b)
	}
	if len(byFile) == 0 {
		return nil, fmt.Errorf("the profile has no block of %s", importPath)
	}
	files := make([]string, 0, len(byFile))
	for name := range byFile {
		files = append(files, name)
	}
	sort.Strings(files)
	var out []FuncCoverage
	fset := token.NewFileSet()
	for _, name := range files {
		path := filepath.Join(pkgDir, name)
		parsed, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		generated := strings.HasPrefix(name, "auto_") || ast.IsGenerated(parsed)
		type span struct {
			fn         *FuncCoverage
			start, end token.Position
		}
		var spans []span
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			start, end := fset.Position(fn.Pos()), fset.Position(fn.End())
			spans = append(spans, span{&FuncCoverage{File: name, Line: start.Line, Func: funcName(fn), Generated: generated}, start, end})
		}
		for _, b := range byFile[name] {
			for _, s := range spans {
				if !notBefore(b.StartLine, b.StartCol, s.start.Line, s.start.Column) || !notBefore(s.end.Line, s.end.Column, b.EndLine, b.EndCol) {
					continue
				}
				s.fn.Blocks++
				if b.Count > 0 {
					s.fn.Reached++
				} else {
					s.fn.Unreached = append(s.fn.Unreached, fmt.Sprintf("%d.%d-%d.%d", b.StartLine, b.StartCol, b.EndLine, b.EndCol))
				}
				break
			}
		}
		for _, s := range spans {
			if s.fn.Blocks > 0 {
				out = append(out, *s.fn)
			}
		}
	}
	return out, nil
}

// notBefore reports whether line.col is at or after atLine.atCol.
func notBefore(line, col, atLine, atCol int) bool {
	return line > atLine || line == atLine && col >= atCol
}

func funcName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	recv := fn.Recv.List[0].Type
	prefix := ""
	if star, ok := recv.(*ast.StarExpr); ok {
		prefix, recv = "*", star.X
	}
	if index, ok := recv.(*ast.IndexExpr); ok {
		recv = index.X
	}
	if ident, ok := recv.(*ast.Ident); ok {
		return "(" + prefix + ident.Name + ")." + fn.Name.Name
	}
	return fn.Name.Name
}

// BuildBlockProfile measures Go block coverage of the example games over the
// checked-in fuzz corpus. In workDir it builds boardgame-util with
// `go build -cover`, instrumenting the command's own package and every
// example game (a binary whose main package is not instrumented writes no
// coverage data), runs `fuzz corpus --check` with GOCOVERDIR set (the
// corpus regenerated in memory: exactly the runs the documents record), and
// converts the counters with `go tool covdata textfmt`. It returns the
// profile's path. goCmd is the go command to run ("go" when empty); it
// inherits the environment, so under ./scripts/go-local the shared GOCACHE
// is used.
func BuildBlockProfile(root, workDir, goCmd string, log io.Writer) (string, error) {
	if goCmd == "" {
		goCmd = "go"
	}
	bin := filepath.Join(workDir, "boardgame-util-cover")
	coverDir := filepath.Join(workDir, "covdata")
	profile := filepath.Join(workDir, "fuzz-corpus.coverprofile")
	if err := os.RemoveAll(coverDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(coverDir, 0755); err != nil {
		return "", err
	}
	run := func(name string, env []string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), env...)
		var stderr bytes.Buffer
		cmd.Stdout = io.Discard
		cmd.Stderr = &stderr
		if log != nil {
			fmt.Fprintf(log, "+ %s %s\n", name, strings.Join(args, " "))
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, stderr.String())
		}
		return nil
	}
	coverPkg := ModulePath + "/boardgame-util," + ModulePath + "/examples/..."
	if err := run(goCmd, nil, "build", "-cover", "-coverpkg="+coverPkg, "-o", bin, "./boardgame-util"); err != nil {
		return "", err
	}
	if err := run(bin, []string{"GOCOVERDIR=" + coverDir}, "fuzz", "corpus", "--check", "--root", root); err != nil {
		return "", err
	}
	if err := run(goCmd, nil, "tool", "covdata", "textfmt", "-i="+coverDir, "-o", profile); err != nil {
		return "", err
	}
	return profile, nil
}

// ReadBlockCoverage parses a profile file and returns one game's functions.
func ReadBlockCoverage(profilePath, root, game string) ([]FuncCoverage, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	blocks, err := ParseProfile(f)
	if err != nil {
		return nil, err
	}
	return BlockCoverage(blocks, GamePackage(game), filepath.Join(root, "examples", game))
}
