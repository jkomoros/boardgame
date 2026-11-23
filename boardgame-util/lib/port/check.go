package port

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// Verdicts of a check.
const (
	VerdictPass         = "PASS"
	VerdictUnproven     = "PASS (unproven)"
	VerdictFail         = "FAIL"
	VerdictPrecondition = "PRECONDITION FAILED"
)

// candidateProducer is the producer a candidate's block carries while it
// is checked (port accept writes the real model id).
const candidateProducer = "candidate"

// CheckOptions configure a check.
type CheckOptions struct {
	Node     Node
	Examples ExampleSource
	// Coverage measures the function's Go blocks over the fuzz corpus
	// (fuzz.BuildBlockProfile, about a minute), unless Profile names a
	// text profile already built.
	Coverage bool
	Profile  string
	// Mutation runs the body mutants of the function.
	Mutation bool
	// Log receives progress lines (nil: none).
	Log io.Writer
}

// CheckResult is everything `port check` found.
type CheckResult struct {
	Func, TSName string
	Brief        *Brief
	Verdict      string
	// Precondition lists what must hold before the function can be checked:
	// every other function of the game translated, the game loading.
	Precondition []string
	Node         nodeCheck
	// Blocks is the function's Go block coverage over the fuzz corpus;
	// nil when not measured.
	Blocks *FunctionBlocks
	// Bodies is the generated game's bodies.gen.ts with the candidate in
	// place (under the producer "candidate").
	Bodies []byte
	// Files are the generated game's three files with the candidate.
	Files map[string][]byte
	// FunctionText is the candidate's function, as it is copied in.
	FunctionText string
}

// nodeCheck is port-check.mjs check's result.
type nodeCheck struct {
	Static *struct {
		OK       bool     `json:"ok"`
		Problems []string `json:"problems"`
	} `json:"static"`
	Types *struct {
		OK          bool     `json:"ok"`
		Diagnostics []string `json:"diagnostics"`
	} `json:"types"`
	Load *struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	} `json:"load"`
	Differential *struct {
		Passed    bool          `json:"passed"`
		Documents int           `json:"documents"`
		Of        int           `json:"of"`
		First     *documentRun  `json:"first"`
		LastCall  *lastCallInfo `json:"lastCall"`
	} `json:"differential"`
	Mutation *MutationResult `json:"mutation"`
}

type documentRun struct {
	Name      string          `json:"name"`
	File      string          `json:"file"`
	Passed    bool            `json:"passed"`
	Exception string          `json:"exception"`
	Report    json.RawMessage `json:"report"`
}

type lastCallInfo struct {
	Calls     int             `json:"calls"`
	Arguments json.RawMessage `json:"arguments"`
	Error     string          `json:"error"`
}

// MutationResult is the body-mutant run of one function.
type MutationResult struct {
	OK                  bool     `json:"ok"`
	Error               string   `json:"error,omitempty"`
	Total               int      `json:"total"`
	Killed              int      `json:"killed"`
	Survivors           []Mutant `json:"survivors"`
	SkippedDeclarations int      `json:"skippedDeclarations"`
}

// Mutant is one surviving mutant.
type Mutant struct {
	ID     int    `json:"id"`
	Kind   string `json:"kind"`
	Line   int    `json:"line"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// FunctionBlocks is one function's Go block coverage.
type FunctionBlocks struct {
	Blocks    int      `json:"blocks"`
	Reached   int      `json:"reached"`
	Unreached []string `json:"unreached"`
}

// candidateFunction is the candidate's function declaration: from
// "export function <name>(" to the first line after it that is exactly
// "}", which a block of bodies.gen.ts must end with.
func candidateFunction(candidate []byte, name string) (string, error) {
	text := string(candidate)
	start := strings.Index(text, "export function "+name+"(")
	if start < 0 {
		return "", fmt.Errorf("the file declares no `export function %s(`", name)
	}
	lines := strings.SplitAfter(text[start:], "\n")
	var out strings.Builder
	for _, line := range lines {
		out.WriteString(line)
		if strings.TrimRight(line, "\r\n") == "}" {
			s := out.String()
			if !strings.HasSuffix(s, "\n") {
				s += "\n"
			}
			return s, nil
		}
	}
	return "", fmt.Errorf("the function %s does not end with a line holding only `}`", name)
}

// WithCandidate is the generated game's files with code as ref's block
// under producer: Generate over the checked-in bodies (or the translator's
// own, when none is checked in) with that block in place.
func (g *Game) WithCandidate(ref, code, producer, brief string) (*transpile.Output, error) {
	entry, ok := g.Entry(ref)
	if !ok {
		return nil, fmt.Errorf("port: %s is not in %s's manifest", ref, g.Name)
	}
	existing := g.Existing
	if existing == nil {
		existing = g.Output.Files["bodies.gen.ts"]
	}
	first := g.Commit
	if block, ok := transpile.Blocks(existing)[ref]; ok && block.Agent != "" && block.Code == code {
		first = block.First
	}
	return g.Pkg.Generate(g.Definition, transpile.WithAgentBlock(existing, ref, producer, brief, entry.SourceHash, first, code), g.Commit)
}

// Check runs `port check` on candidate, a file holding a translation of
// ref: static, types, differential, then coverage and mutation.
func (g *Game) Check(ref string, candidate []byte, opts CheckOptions) (*CheckResult, error) {
	logf := func(format string, args ...any) {
		if opts.Log != nil {
			fmt.Fprintf(opts.Log, format+"\n", args...)
		}
	}
	brief, err := g.WriteBrief(ref, opts.Examples)
	if err != nil {
		return nil, err
	}
	entry, _ := g.Entry(ref)
	result := &CheckResult{Func: ref, TSName: brief.TSName, Brief: brief}
	signature, err := g.Signature(ref)
	if err != nil {
		return nil, err
	}
	code, err := candidateFunction(candidate, brief.TSName)
	if err != nil {
		// The static step names the problem; assemble nothing.
		code = ""
	}
	result.FunctionText = code
	files := map[string][]byte{}
	if code != "" {
		out, err := g.WithCandidate(ref, code, candidateProducer, brief.SHA256)
		if err != nil {
			return nil, err
		}
		files = out.Files
		result.Files = files
		result.Bodies = out.Files["bodies.gen.ts"]
		for _, fn := range out.Functions {
			if fn.Func == ref {
				continue
			}
			if _, agent := out.Agent[fn.Func]; agent || len(fn.Refusals) == 0 {
				continue
			}
			result.Precondition = append(result.Precondition, fmt.Sprintf("%s has no translation yet (%s %s); every other function of %s must be translated first", fn.Func, fn.Refusals[0].Reason, fn.Refusals[0].Detail, g.Name))
		}
	}
	docs, err := g.Documents()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		result.Precondition = append(result.Precondition, fmt.Sprintf("%s has no recordings (no conformance/fuzz/%s/ corpus and no conformance documents), so nothing can be compared", g.Name, g.Name))
	}
	var callees []string
	for _, c := range entry.Calls {
		callees = append(callees, tsName(c))
	}
	if callees == nil {
		callees = []string{}
	}
	job := map[string]any{
		"game":              g.Name,
		"tsName":            brief.TSName,
		"expectedSignature": strings.TrimSuffix(signature, " {"),
		"outputPath":        strings.TrimPrefix(brief.Output, EngineDir+"/"),
		"candidate":         string(candidate),
		"importable":        transpile.AgentImportable(),
		"callees":           callees,
		"virtual":           g.Virtual(files),
		"docs":              docs,
		"mutation":          opts.Mutation && len(result.Precondition) == 0,
	}
	if len(result.Precondition) > 0 {
		// The static and type checks still run; the replay would only
		// report the missing functions again.
		job["docs"] = []string{}
	}
	if code == "" {
		job["virtual"] = map[string]string{}
	}
	logf("port check %s: static, types, differential over %d documents%s", ref, len(docs), map[bool]string{true: ", mutation", false: ""}[opts.Mutation])
	if err := opts.Node.Run("check", job, &result.Node); err != nil {
		return nil, err
	}
	n := result.Node
	switch {
	case n.Static == nil || !n.Static.OK, n.Types == nil || !n.Types.OK:
		result.Verdict = VerdictFail
		return result, nil
	case len(result.Precondition) > 0:
		result.Verdict = VerdictPrecondition
		return result, nil
	case n.Load == nil || !n.Load.OK:
		if n.Load != nil {
			result.Precondition = append(result.Precondition, "the generated game does not load: "+n.Load.Error)
		}
		result.Verdict = VerdictPrecondition
		return result, nil
	case n.Differential == nil || !n.Differential.Passed:
		result.Verdict = VerdictFail
		return result, nil
	}
	result.Verdict = VerdictPass
	if opts.Coverage || opts.Profile != "" {
		profile := opts.Profile
		if profile == "" {
			work, err := os.MkdirTemp("", "boardgame-port-coverage-")
			if err != nil {
				return nil, err
			}
			defer os.RemoveAll(work)
			logf("port check %s: building the coverage profile of the fuzz corpus", ref)
			if profile, err = fuzz.BuildBlockProfile(g.Root, work, "", opts.Log); err != nil {
				return nil, err
			}
		}
		blocks, err := g.FunctionBlocks(profile, ref)
		if err != nil {
			return nil, err
		}
		result.Blocks = blocks
		if blocks.Reached < blocks.Blocks {
			result.Verdict = VerdictUnproven
		}
	}
	return result, nil
}

// FunctionBlocks is ref's Go block coverage in a text profile: every block
// inside the function's source span.
func (g *Game) FunctionBlocks(profilePath, ref string) (*FunctionBlocks, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	blocks, err := fuzz.ParseProfile(f)
	if err != nil {
		return nil, err
	}
	node, ok := g.src.units[ref]
	if !ok {
		return nil, fmt.Errorf("port: no source for %s", ref)
	}
	return blocksIn(blocks, g.ImportPath, g.src, node), nil
}

func blocksIn(blocks []fuzz.ProfileBlock, importPath string, src *source, node ast.Node) *FunctionBlocks {
	start, end := src.fset.Position(node.Pos()), src.fset.Position(node.End())
	file := filepath.Base(start.Filename)
	out := &FunctionBlocks{Unreached: []string{}}
	for _, b := range blocks {
		if filepath.ToSlash(filepath.Dir(b.File)) != importPath || filepath.Base(b.File) != file {
			continue
		}
		after := b.StartLine > start.Line || b.StartLine == start.Line && b.StartCol >= start.Column
		before := b.EndLine < end.Line || b.EndLine == end.Line && b.EndCol <= end.Column
		if !after || !before {
			continue
		}
		out.Blocks++
		if b.Count > 0 {
			out.Reached++
		} else {
			out.Unreached = append(out.Unreached, fmt.Sprintf("%s:%d.%d-%d.%d", file, b.StartLine, b.StartCol, b.EndLine, b.EndCol))
		}
	}
	sort.Strings(out.Unreached)
	return out
}

// Text is the check's report as `port check` prints it.
func (r *CheckResult) Text() string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	w("port check %s (%s)", r.Func, r.TSName)
	w("  brief: %s (sha256 %s)", r.Brief.Path, r.Brief.SHA256)
	n := r.Node
	if n.Static != nil {
		if n.Static.OK {
			w("  static: ok (one exported function, signature, imports, sandbox and purity rules)")
		} else {
			w("  static: %d problems", len(n.Static.Problems))
			for _, p := range n.Static.Problems {
				w("    %s", p)
			}
		}
	}
	if n.Types != nil {
		if n.Types.OK {
			w("  types: ok (tsconfig.sandbox.json, the file and the generated game with it in place)")
		} else {
			w("  types: %d errors", len(n.Types.Diagnostics))
			for _, d := range n.Types.Diagnostics {
				w("    %s", d)
			}
		}
	}
	for _, p := range r.Precondition {
		w("  precondition: %s", p)
	}
	if n.Load != nil && n.Load.OK {
		w("  the generated game loads: ok")
	}
	if d := n.Differential; d != nil && len(r.Precondition) == 0 {
		if d.Passed {
			w("  differential: %d of %d documents pass (the fuzz corpus and the conformance documents)", d.Documents, d.Of)
		} else {
			w("  differential: document %d of %d differs", d.Documents, d.Of)
		}
	}
	if r.Blocks != nil {
		w("  coverage: %d of %d Go blocks of the function reached over the fuzz corpus", r.Blocks.Reached, r.Blocks.Blocks)
		for _, u := range r.Blocks.Unreached {
			w("    unreached: %s", u)
		}
	}
	if m := n.Mutation; m != nil {
		if !m.OK {
			w("  mutation: not run: %s", m.Error)
		} else {
			w("  mutation: %d of %d mutants killed (%d declarations not deleted)", m.Killed, m.Total, m.SkippedDeclarations)
			for _, s := range m.Survivors {
				w("    survivor #%d %s at line %d of the function: `%s` -> `%s`", s.ID, s.Kind, s.Line, s.Before, s.After)
			}
		}
	}
	if d := n.Differential; d != nil && !d.Passed && d.First != nil {
		w("%s: the first difference", VerdictFail)
		writeDifference(&b, d.First, d.LastCall)
		return b.String()
	}
	w("%s", r.Verdict)
	return b.String()
}

// writeDifference writes a divergence report field by field.
func writeDifference(b *strings.Builder, doc *documentRun, last *lastCallInfo) {
	w := func(format string, args ...any) { fmt.Fprintf(b, format+"\n", args...) }
	w("  document: %s", doc.Name)
	if doc.Exception != "" {
		w("  the replay threw: %s", doc.Exception)
		return
	}
	var report map[string]json.RawMessage
	if err := json.Unmarshal(doc.Report, &report); err != nil {
		w("  report: %s", string(doc.Report))
		return
	}
	field := func(label, key string) {
		if raw, ok := report[key]; ok {
			w("  %s: %s", label, compactJSON(raw))
		}
	}
	field("version", "version")
	field("viewer", "viewer")
	field("field", "field")
	field("path", "path")
	field("Go recorded", "want")
	field("produced", "got")
	field("move", "move")
	field("last outside input", "lastOutsideInput")
	field("bodies executed", "bodiesExecuted")
	if last != nil {
		switch {
		case last.Error != "":
			w("  arguments the function was last called with: unknown (%s)", last.Error)
		case last.Calls == 0:
			w("  arguments the function was last called with: none (it was not called up to that version)")
		default:
			w("  arguments the function was last called with (call %d up to that version): %s", last.Calls, compactJSON(last.Arguments))
		}
	}
	field("repro", "repro")
}

// replayResult is port-check.mjs replay's result.
type replayResult struct {
	Loaded    bool          `json:"loaded"`
	Error     string        `json:"error"`
	Documents []documentRun `json:"documents"`
}

// Replay replays every document of the game on files (the generated
// game's three files; nil for the checked-in ones).
func (g *Game) Replay(node Node, files map[string][]byte) (*replayResult, error) {
	docs, err := g.Documents()
	if err != nil {
		return nil, err
	}
	var result replayResult
	if err := node.Run("replay", map[string]any{"game": g.Name, "virtual": g.Virtual(files), "docs": docs}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Accept checks candidate and, when it passes (proven or not), copies its
// function into the generated game's bodies.gen.ts under the agent header
// (producer is the model id) and stores the pair through pairs (nil: no
// store). It returns the check and whether the pair was stored.
func (g *Game) Accept(ref string, candidate []byte, producer string, opts CheckOptions, pairs PairWriter) (*CheckResult, bool, error) {
	if producer == "" || strings.ContainsAny(producer, " \t\n") {
		return nil, false, fmt.Errorf("port: --producer must be a model id without spaces")
	}
	result, err := g.Check(ref, candidate, opts)
	if err != nil {
		return nil, false, err
	}
	if result.Verdict != VerdictPass && result.Verdict != VerdictUnproven {
		return result, false, nil
	}
	out, err := g.WithCandidate(ref, result.FunctionText, producer, result.Brief.SHA256)
	if err != nil {
		return nil, false, err
	}
	if err := g.writeGenerated(out.Files); err != nil {
		return nil, false, err
	}
	stored := false
	if pairs != nil {
		entry, _ := g.Entry(ref)
		pair := Pair{Game: g.Name, Func: ref, Role: entry.Role, TSName: result.TSName, SourceHash: entry.SourceHash, Producer: "agent:" + producer,
			Brief: result.Brief.SHA256, Refusals: entry.Refusals, GoSource: g.src.unitText(ref), TypeScript: result.FunctionText}
		docs, _ := g.Documents()
		for _, d := range docs {
			if strings.Contains(d, "/fuzz/") {
				pair.Seeds = append(pair.Seeds, strings.TrimSuffix(filepath.Base(d), ".json.gz"))
			}
		}
		if result.Blocks != nil {
			pair.BlocksCovered, pair.BlocksTotal = &result.Blocks.Reached, &result.Blocks.Blocks
		}
		if m := result.Node.Mutation; m != nil && m.OK {
			pair.MutantsKilled, pair.MutantsTotal = &m.Killed, &m.Total
		}
		if err := pairs.StorePair(pair); err == nil {
			stored = true
		} else if err != ErrNoPairStore {
			return result, false, err
		}
	}
	return result, stored, nil
}

// writeGenerated writes the generated game's three files.
func (g *Game) writeGenerated(files map[string][]byte) error {
	if err := os.MkdirAll(g.GeneratedDir(), 0755); err != nil {
		return err
	}
	for _, name := range transpile.FileNames {
		if err := os.WriteFile(g.GeneratedPath(name), files[name], 0644); err != nil {
			return err
		}
	}
	g.Existing = files["bodies.gen.ts"]
	return nil
}

// AdoptResult is what `port --adopt` found.
type AdoptResult struct {
	Func string
	// Problems are the checks that failed; empty when adopted.
	Problems []string
	Replay   *replayResult
	Blocks   *FunctionBlocks
	Adopted  bool
}

// Adopt replaces ref's agent-written block with the translator's output,
// after the proposal's three checks ("Turning a refusal reason into a
// rule"): the translator now accepts the Go source; with its output in
// place, the game's fuzz corpus and conformance documents pass; and the
// function's recorded block coverage is complete.
func (g *Game) Adopt(ref string, node Node, profile string, log io.Writer) (*AdoptResult, error) {
	result := &AdoptResult{Func: ref}
	if _, ok := g.Output.Agent[ref]; !ok {
		return nil, fmt.Errorf("port: %s has no agent-written block in %s", ref, g.GeneratedPath("bodies.gen.ts"))
	}
	fn, ok := g.Function(ref)
	if !ok {
		return nil, fmt.Errorf("port: %s is not in %s's manifest", ref, g.Name)
	}
	if len(fn.Refusals) > 0 {
		result.Problems = append(result.Problems, fmt.Sprintf("the translator still refuses %s: %s %s", ref, fn.Refusals[0].Reason, fn.Refusals[0].Detail))
		return result, nil
	}
	out, err := g.Pkg.Generate(g.Definition, transpile.WithoutBlock(g.Existing, ref), g.Commit)
	if err != nil {
		return nil, err
	}
	replay, err := g.Replay(node, out.Files)
	if err != nil {
		return nil, err
	}
	result.Replay = replay
	if !replay.Loaded {
		result.Problems = append(result.Problems, "the generated game with the translator's text does not load: "+replay.Error)
	}
	for _, d := range replay.Documents {
		if !d.Passed {
			result.Problems = append(result.Problems, "with the translator's text, "+d.Name+" differs")
			break
		}
	}
	if profile == "" {
		work, err := os.MkdirTemp("", "boardgame-port-coverage-")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(work)
		if profile, err = fuzz.BuildBlockProfile(g.Root, work, "", log); err != nil {
			return nil, err
		}
	}
	blocks, err := g.FunctionBlocks(profile, ref)
	if err != nil {
		return nil, err
	}
	result.Blocks = blocks
	if blocks.Reached < blocks.Blocks {
		result.Problems = append(result.Problems, fmt.Sprintf("%d of %d Go blocks of %s are unreached by the fuzz corpus, so the corpus is no proof", blocks.Blocks-blocks.Reached, blocks.Blocks, ref))
	}
	if len(result.Problems) > 0 {
		return result, nil
	}
	if err := g.writeGenerated(out.Files); err != nil {
		return nil, err
	}
	result.Adopted = true
	return result, nil
}

// Pair is one accepted translation, for the pair store
// (fabric/TRANSPILE-PROPOSAL.md, "Stored pairs").
type Pair struct {
	Game, Func, Role, TSName, SourceHash, Producer, Brief   string
	Refusals                                                []transpile.Refusal
	GoSource, TypeScript                                    string
	Seeds                                                   []string
	BlocksCovered, BlocksTotal, MutantsKilled, MutantsTotal *int
}

// PairWriter stores an accepted pair.
type PairWriter interface {
	StorePair(Pair) error
}

// ErrNoPairStore is StorePair's answer when there is no store to write to.
var ErrNoPairStore = fmt.Errorf("port: no pair store (%s does not exist)", PairsDir)

// FilePairWriter writes the proposal's layout,
// <PairsDir>/<game>/<ts name>/{source.go,accepted.ts,meta.json}, when
// PairsDir exists (the pair store's own workstream creates it).
type FilePairWriter struct{ Root string }

// StorePair implements PairWriter.
func (w FilePairWriter) StorePair(p Pair) error {
	base := filepath.Join(w.Root, filepath.FromSlash(PairsDir))
	if info, err := os.Stat(base); err != nil || !info.IsDir() {
		return ErrNoPairStore
	}
	dir := filepath.Join(base, p.Game, p.TSName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	refusals := p.Refusals
	if refusals == nil {
		refusals = []transpile.Refusal{}
	}
	meta := map[string]any{
		"func": p.Func, "role": p.Role, "sourceHash": p.SourceHash, "producer": p.Producer, "brief": p.Brief,
		"refusals": refusals,
		"fuzz": map[string]any{"seeds": p.Seeds, "blocksCovered": p.BlocksCovered, "blocksTotal": p.BlocksTotal,
			"mutantsKilled": p.MutantsKilled, "mutantsTotal": p.MutantsTotal},
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(meta); err != nil {
		return err
	}
	files := map[string][]byte{"source.go": []byte(p.GoSource + "\n"), "accepted.ts": []byte(p.TypeScript), "meta.json": buf.Bytes()}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
