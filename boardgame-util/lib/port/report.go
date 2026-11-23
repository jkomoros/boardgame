package port

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// ReportsDir holds the per-game reports, relative to the module root.
const ReportsDir = "fabric/transpile/reports"

// ReportTable is the table of every game's report, relative to the module root.
const ReportTable = "fabric/transpile/REPORT.md"

// ReportVersion is the report format's version.
const ReportVersion = 1

// Report is fabric/transpile/reports/<game>.json
// (fabric/TRANSPILE-PROPOSAL.md, "The per-game report"). A field that
// cannot be produced yet is null, and Nulls says why, by field path.
type Report struct {
	Version int    `json:"version"`
	Game    string `json:"game"`
	Package string `json:"package"`
	// Definition is "current" when the checked-in definition is what the
	// exporter writes now, else "stale".
	Definition string `json:"definition"`

	MovesData     int `json:"movesData"`
	MovesWithCode int `json:"movesWithCode"`

	Functions ReportFunctions `json:"functions"`
	Headline  Headline        `json:"headline"`
	Coverage  *ReportCoverage `json:"coverage"`
	// Replay is the generated game (checked in, else the translator's
	// output) replayed on the TypeScript side against every document.
	Replay *ReportReplay `json:"replay"`
	// Stale are the checked-in blocks whose header names a Go source hash
	// other than the function's current one.
	Stale []StaleBlock `json:"stale"`

	Nulls map[string]string `json:"nulls"`
}

// ReportFunctions are the function counts.
type ReportFunctions struct {
	// Total are the game-written functions that reach engine output (the
	// manifest).
	Total int `json:"total"`
	// Data are the game functions the exporter evaluates on the live
	// manager instead of translating (gamedef StaticReport.NotCounted).
	Data FunctionList `json:"data"`
	// Mechanical are the blocks of the checked-in bodies.gen.ts the
	// translator wrote.
	Mechanical *FunctionList `json:"mechanical"`
	// Agent are the checked-in agent-written blocks.
	Agent FunctionList `json:"agent"`
	// AgentNowMechanical are those of Agent the translator now translates
	// (a dry run: the checked-in text is not changed).
	AgentNowMechanical FunctionList `json:"agentNowMechanical"`
	// EngineGap are the functions refused only for engine gaps, each named
	// with its gaps.
	EngineGap EngineGapList `json:"engineGap"`
	// NotPorted are Diagram and agent code.
	NotPorted FunctionList `json:"notPorted"`
	// Translatable is what a dry run of the translator translates now.
	Translatable int `json:"translatable"`
	// BoundForAgent are refused outside the subset with no agent-written
	// block yet: the agent's queue, callee first.
	BoundForAgent RefusedList `json:"boundForAgent"`
}

// FunctionList is a count and the FnRefs counted, sorted.
type FunctionList struct {
	Count     int      `json:"count"`
	Functions []string `json:"functions"`
}

// EngineGapList are engine-gap functions with what each lacks.
type EngineGapList struct {
	Count     int           `json:"count"`
	Functions []RefusedFunc `json:"functions"`
}

// RefusedList are refused functions with their refusals.
type RefusedList struct {
	Count     int           `json:"count"`
	Functions []RefusedFunc `json:"functions"`
}

// RefusedFunc is one refused function and its refusals as "reason detail".
type RefusedFunc struct {
	Func     string   `json:"func"`
	Refusals []string `json:"refusals"`
}

// Headline is the four-way split of a game's units of porting work (move
// configs that are data only, plus functions reaching engine output) into
// data, mechanical, agent and engine gap.
type Headline struct {
	Units      int   `json:"units"`
	Data       Share `json:"data"`
	Mechanical Share `json:"mechanical"`
	Agent      Share `json:"agent"`
	EngineGap  Share `json:"engineGap"`
}

// Share is a count of the headline's units.
type Share struct {
	Count int `json:"count"`
	// Of is the units.
	Of int `json:"of"`
	// Percent is Count/Of, one decimal.
	Percent string `json:"percent"`
}

// ReportCoverage are the proposal's four measures over the checked-in fuzz
// corpus, and the surviving mutants.
type ReportCoverage struct {
	Documents  int                       `json:"documents"`
	Versions   int                       `json:"versions"`
	Rejections int                       `json:"rejections"`
	Moves      []fuzz.MoveCoverage       `json:"moves"`
	Predicates []fuzz.PredicateCoverage  `json:"predicates"`
	Visibility []fuzz.VisibilityCoverage `json:"visibility"`
	Blocks     *BlockSummary             `json:"blocks"`
	Gaps       fuzz.Gaps                 `json:"gaps"`
	// Mutants are the surviving body mutants per function.
	Mutants *[]Mutant `json:"mutants"`
}

// BlockSummary is the Go block coverage of the manifest's functions.
type BlockSummary struct {
	Reached   int             `json:"reached"`
	Blocks    int             `json:"blocks"`
	Functions []FunctionCover `json:"functions"`
}

// FunctionCover is one function's blocks.
type FunctionCover struct {
	Func string `json:"func"`
	FunctionBlocks
}

// ReportReplay is the TypeScript replay of the generated game.
type ReportReplay struct {
	// Source is "checked-in" or "translator-output".
	Source    string            `json:"source"`
	Loaded    bool              `json:"loaded"`
	Error     string            `json:"error,omitempty"`
	Documents int               `json:"documents"`
	Passed    int               `json:"passed"`
	First     *ReplayDifference `json:"first,omitempty"`
}

// ReplayDifference locates the first failing document's first difference.
type ReplayDifference struct {
	Document string `json:"document"`
	Version  *int   `json:"version,omitempty"`
	Viewer   *int   `json:"viewer,omitempty"`
	Field    string `json:"field,omitempty"`
	Path     string `json:"path,omitempty"`
	Error    string `json:"error,omitempty"`
}

// StaleBlock is a checked-in block older than its Go function.
type StaleBlock struct {
	Func       string `json:"func"`
	Producer   string `json:"producer"`
	Header     string `json:"header"`
	SourceHash string `json:"sourceHash"`
}

// ReportInputs are what a report is built from beside the game.
type ReportInputs struct {
	// DefinitionCurrent is whether the checked-in definition is the
	// exporter's output now.
	DefinitionCurrent bool
	// Profile is a text block profile of the fuzz corpus ("" when none).
	Profile string
	// Replay is the TypeScript replay (nil when not run, with why).
	Replay       *replayResult
	ReplaySource string
	ReplayWhy    string
}

// BuildReport computes the game's report.
func (g *Game) BuildReport(in ReportInputs) (*Report, error) {
	s := g.Pkg.Static
	r := &Report{Version: ReportVersion, Game: g.Name, Package: g.ImportPath, Nulls: map[string]string{}}
	r.Definition = "stale"
	if in.DefinitionCurrent {
		r.Definition = "current"
	}
	dataOnlyPorted := 0
	dataOnlyGap := 0
	for _, m := range s.MoveConfigs {
		if m.DataOnly {
			r.MovesData++
			if m.StockPorted {
				dataOnlyPorted++
			} else {
				dataOnlyGap++
			}
		} else {
			r.MovesWithCode++
		}
	}

	f := &r.Functions
	f.Total = len(s.Functions)
	f.Data = FunctionList{Functions: []string{}}
	for _, ref := range s.NotCounted {
		f.Data.Functions = append(f.Data.Functions, ref.Func)
	}
	f.Data = list(f.Data.Functions)
	f.NotPorted = FunctionList{Functions: []string{}}
	for _, ref := range s.NotPorted {
		f.NotPorted.Functions = append(f.NotPorted.Functions, ref.Func)
	}
	f.NotPorted = list(f.NotPorted.Functions)

	var checkedIn map[string]transpile.Block
	if g.HasGenerated() {
		checkedIn = transpile.Blocks(g.Existing)
	}
	var mechanical, agent, agentNow []string
	for ref, block := range checkedIn {
		if _, ok := g.Entry(ref); !ok {
			continue
		}
		if block.Agent != "" {
			agent = append(agent, ref)
			if fn, ok := g.Function(ref); ok && len(fn.Refusals) == 0 {
				agentNow = append(agentNow, ref)
			}
		} else {
			mechanical = append(mechanical, ref)
		}
	}
	if g.HasGenerated() {
		m := list(mechanical)
		f.Mechanical = &m
	} else {
		r.Nulls["functions.mechanical"] = "no generated game is checked in (" + transpile.GamesDir + "/" + g.Name + "/); functions.translatable is what the translator would write"
	}
	f.Agent = list(agent)
	f.AgentNowMechanical = list(agentNow)

	f.EngineGap = EngineGapList{Functions: []RefusedFunc{}}
	f.BoundForAgent = RefusedList{Functions: []RefusedFunc{}}
	agentBlock := func(ref string) bool {
		b, ok := checkedIn[ref]
		return ok && b.Agent != "" && b.SourceHash == g.byFn[ref].SourceHash
	}
	units := map[string]string{}
	for _, ref := range s.Order {
		entry := g.byFn[ref]
		if len(entry.Refusals) == 0 {
			f.Translatable++
		}
		switch {
		case agentBlock(ref):
			units[ref] = "agent"
		case s.Buckets[ref] == gamedef.BucketEngineGap:
			units[ref] = "engineGap"
			f.EngineGap.Functions = append(f.EngineGap.Functions, refusedFunc(entry))
		case len(entry.Refusals) > 0:
			units[ref] = "agent"
			f.BoundForAgent.Functions = append(f.BoundForAgent.Functions, refusedFunc(entry))
		default:
			units[ref] = "mechanical"
		}
	}
	f.EngineGap.Count = len(f.EngineGap.Functions)
	f.BoundForAgent.Count = len(f.BoundForAgent.Functions)

	h := &r.Headline
	h.Units = r.MovesData + f.Total
	counts := map[string]int{"data": dataOnlyPorted, "engineGap": dataOnlyGap}
	for _, bucket := range units {
		counts[bucket]++
	}
	h.Data, h.Mechanical, h.Agent, h.EngineGap = share(counts["data"], h.Units), share(counts["mechanical"], h.Units), share(counts["agent"], h.Units), share(counts["engineGap"], h.Units)

	// Coverage.
	docs, err := fuzz.LoadDocuments(filepath.Join(g.Root, filepath.FromSlash(fuzz.CorpusDir)), g.Name)
	if err != nil {
		r.Nulls["coverage"] = "no checked-in fuzz corpus (" + fuzz.CorpusDir + "/" + g.Name + "/)"
	} else {
		cov, err := fuzz.Coverage(docs[g.Name])
		if err != nil {
			return nil, err
		}
		c := &ReportCoverage{Documents: cov.Documents, Versions: cov.Versions, Rejections: cov.Rejections, Moves: cov.Moves, Predicates: cov.Predicates, Visibility: cov.Visibility, Gaps: cov.Gaps()}
		if c.Moves == nil {
			c.Moves = []fuzz.MoveCoverage{}
		}
		if c.Predicates == nil {
			c.Predicates = []fuzz.PredicateCoverage{}
		}
		if c.Visibility == nil {
			c.Visibility = []fuzz.VisibilityCoverage{}
		}
		if in.Profile != "" {
			blocks, err := g.manifestBlocks(in.Profile)
			if err != nil {
				return nil, err
			}
			c.Blocks = blocks
		} else {
			r.Nulls["coverage.blocks"] = "no block profile was built for this run"
		}
		r.Nulls["coverage.mutants"] = "no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)"
		r.Coverage = c
	}

	// Replay.
	if in.Replay != nil {
		rr := &ReportReplay{Source: in.ReplaySource, Loaded: in.Replay.Loaded, Error: in.Replay.Error, Documents: len(in.Replay.Documents)}
		for _, d := range in.Replay.Documents {
			if d.Passed {
				rr.Passed++
				continue
			}
			if rr.First == nil {
				rr.First = difference(d)
			}
		}
		r.Replay = rr
	} else {
		r.Nulls["replay"] = in.ReplayWhy
	}

	// Stale.
	if g.HasGenerated() {
		r.Stale = []StaleBlock{}
		for _, ref := range sortedKeys(checkedIn) {
			block := checkedIn[ref]
			entry, ok := g.Entry(ref)
			if !ok {
				r.Stale = append(r.Stale, StaleBlock{Func: ref, Producer: block.Producer, Header: block.SourceHash, SourceHash: "(no longer in the manifest)"})
				continue
			}
			if block.SourceHash != entry.SourceHash {
				r.Stale = append(r.Stale, StaleBlock{Func: ref, Producer: block.Producer, Header: block.SourceHash, SourceHash: entry.SourceHash})
			}
		}
	} else {
		r.Nulls["stale"] = "no generated game is checked in"
	}
	return r, nil
}

func list(refs []string) FunctionList {
	out := append([]string{}, refs...)
	sort.Strings(out)
	return FunctionList{Count: len(out), Functions: out}
}

func refusedFunc(entry gamedef.FunctionEntry) RefusedFunc {
	rf := RefusedFunc{Func: entry.Func, Refusals: []string{}}
	for _, r := range entry.Refusals {
		rf.Refusals = append(rf.Refusals, r.Reason+" "+r.Detail)
	}
	return rf
}

func share(n, of int) Share {
	percent := "0.0"
	if of > 0 {
		percent = fmt.Sprintf("%.1f", float64(n)*100/float64(of))
	}
	return Share{Count: n, Of: of, Percent: percent}
}

func difference(d documentRun) *ReplayDifference {
	out := &ReplayDifference{Document: d.Name}
	if d.Exception != "" {
		out.Error = d.Exception
		return out
	}
	var report struct {
		Version *int   `json:"version"`
		Viewer  *int   `json:"viewer"`
		Field   string `json:"field"`
		Path    string `json:"path"`
	}
	if json.Unmarshal(d.Report, &report) == nil {
		out.Version, out.Viewer, out.Field, out.Path = report.Version, report.Viewer, report.Field, report.Path
	}
	return out
}

// manifestBlocks is the block coverage of every manifest function.
func (g *Game) manifestBlocks(profilePath string) (*BlockSummary, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	blocks, err := fuzz.ParseProfile(f)
	if err != nil {
		return nil, err
	}
	out := &BlockSummary{Functions: []FunctionCover{}}
	for _, ref := range sortedKeys(g.byFn) {
		node, ok := g.src.units[ref]
		if !ok {
			continue
		}
		fb := blocksIn(blocks, g.ImportPath, g.src, node)
		out.Functions = append(out.Functions, FunctionCover{Func: ref, FunctionBlocks: *fb})
		out.Reached += fb.Reached
		out.Blocks += fb.Blocks
	}
	return out, nil
}

// MarshalReport is a report's checked-in bytes.
func MarshalReport(r *Report) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReportPath is a game's report, relative to the module root.
func ReportPath(game string) string { return ReportsDir + "/" + game + ".json" }

// cell is a number, or "n/a" for a null field.
func cell(n *int) string {
	if n == nil {
		return "null"
	}
	return fmt.Sprint(*n)
}

// TableMarkdown is fabric/transpile/REPORT.md for the reports, in
// ExampleGames order.
func TableMarkdown(reports []*Report) []byte {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }
	w("# Transpile pipeline: the per-game report\n\n")
	w("Written by `boardgame-util port --all --dry-run` from `%s/<game>.json`; do not edit by hand. `port --all --dry-run --check` fails when this file or a report is stale. Fields are those of fabric/TRANSPILE-PROPOSAL.md, \"The per-game report\". A field that cannot be produced yet is `null` in the JSON, with its reason under `nulls`, and \"null\" here.\n\n", ReportsDir)

	w("## Headline: data, mechanical, agent, engine gap\n\n")
	w("| Game | Units | Data | Mechanical | Agent | Engine gap |\n|---|---|---|---|---|---|\n")
	totals := map[string]int{}
	for _, r := range reports {
		h := r.Headline
		w("| %s | %d | %s | %s | %s | %s |\n", r.Game, h.Units, shareCell(h.Data), shareCell(h.Mechanical), shareCell(h.Agent), shareCell(h.EngineGap))
		totals["units"] += h.Units
		totals["data"] += h.Data.Count
		totals["mechanical"] += h.Mechanical.Count
		totals["agent"] += h.Agent.Count
		totals["engineGap"] += h.EngineGap.Count
	}
	u := totals["units"]
	w("| **Total** | **%d** | **%s** | **%s** | **%s** | **%s** |\n\n", u, shareCell(share(totals["data"], u)), shareCell(share(totals["mechanical"], u)), shareCell(share(totals["agent"], u)), shareCell(share(totals["engineGap"], u)))
	w(`How each bucket is counted. The units of a game are its move configs that
are data only plus its functions that reach engine output (the manifest of
boardgame-util/lib/gamedef, the declarativeness lint's "functions reaching
engine output"). A move config with game code is not a unit of its own: its
configuration is exported as data like any other, and its code is counted as
the functions it holds, so counting the config too would count that code
twice. Every unit is in exactly one bucket:

- **Data**: a data-only move config (its type is a stock move, or a game type
  that embeds one and declares no method) whose stock type the engine has
  (MoveConfigStatic.StockPorted).
- **Mechanical**: a function the translator translates now (a dry run of
  lib/transpile, verdict subset) that has no checked-in agent-written block.
- **Agent**: a function with a checked-in agent-written block whose Go source
  hash is current, or a function the translator refuses for a reason other
  than an engine gap, which is bound for an agent (functions.boundForAgent).
- **Engine gap**: a function refused only for engine gaps (gamedef bucket
  engineGap: every refusal is engine-gap, or callee-refused naming such a
  function), and a data-only move config whose stock type the engine lacks.

The headline is a dry run: it does not need a generated game to be checked
in. functions.mechanical below counts only what is checked in.

`)

	w("## Per game\n\n")
	w("| Game | movesData | movesWithCode | functions | data | mechanical | translatable | agent | agentNowMechanical | engineGap | boundForAgent | notPorted | stale | definition | TS replay |\n")
	w("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range reports {
		f := r.Functions
		var mech *int
		if f.Mechanical != nil {
			mech = &f.Mechanical.Count
		}
		stale := "null"
		if r.Stale != nil {
			stale = fmt.Sprint(len(r.Stale))
		}
		replay := "null"
		if r.Replay != nil {
			if !r.Replay.Loaded {
				replay = "does not load (" + r.Replay.Source + ")"
			} else {
				replay = fmt.Sprintf("%d of %d pass (%s)", r.Replay.Passed, r.Replay.Documents, r.Replay.Source)
			}
		}
		w("| %s | %d | %d | %d | %d | %s | %d | %d | %d | %d | %d | %d | %s | %s | %s |\n", r.Game, r.MovesData, r.MovesWithCode, f.Total, f.Data.Count, cell(mech), f.Translatable, f.Agent.Count, f.AgentNowMechanical.Count, f.EngineGap.Count, f.BoundForAgent.Count, f.NotPorted.Count, stale, r.Definition, replay)
	}
	w("\nThe one number to watch per game is agent − agentNowMechanical; it should fall.\n\n")

	w("## Coverage\n\n")
	w("| Game | Fuzz documents | Versions | Rejections | Moves never applied | Predicates missing a verdict | Hidden props never sanitized while non-zero | Go blocks reached | Surviving mutants |\n")
	w("|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range reports {
		c := r.Coverage
		if c == nil {
			w("| %s | null | null | null | null | null | null | null | null |\n", r.Game)
			continue
		}
		blocks := "null"
		if c.Blocks != nil {
			blocks = fmt.Sprintf("%d of %d", c.Blocks.Reached, c.Blocks.Blocks)
		}
		mutants := "null"
		if c.Mutants != nil {
			mutants = fmt.Sprint(len(*c.Mutants))
		}
		w("| %s | %d | %d | %d | %s | %d of %d | %s | %s | %s |\n", r.Game, c.Documents, c.Versions, c.Rejections, listCell(c.Gaps.NeverApplied), len(c.Gaps.Predicates), len(c.Predicates), listCell(c.Gaps.Visibility), blocks, mutants)
	}

	w("\n## Named functions\n\n")
	for _, r := range reports {
		var lines []string
		for _, fn := range r.Functions.EngineGap.Functions {
			lines = append(lines, "- engine gap: `"+fn.Func+"`: "+strings.Join(fn.Refusals, "; "))
		}
		for _, fn := range r.Functions.BoundForAgent.Functions {
			lines = append(lines, "- bound for an agent: `"+fn.Func+"`: "+strings.Join(fn.Refusals, "; "))
		}
		for _, fn := range r.Functions.Agent.Functions {
			lines = append(lines, "- agent-written: `"+fn+"`")
		}
		for _, s := range r.Stale {
			lines = append(lines, "- stale: `"+s.Func+"` ("+s.Producer+")")
		}
		if r.Replay != nil && r.Replay.First != nil {
			d := r.Replay.First
			text := d.Document
			if d.Error != "" {
				text += ": threw " + d.Error
			} else {
				text += fmt.Sprintf(": %s %s", d.Field, d.Path)
				if d.Version != nil {
					text += fmt.Sprintf(" at version %d", *d.Version)
				}
			}
			lines = append(lines, "- first replay difference: "+text)
		}
		if r.Replay != nil && !r.Replay.Loaded {
			lines = append(lines, "- the generated game does not load: "+r.Replay.Error)
		}
		for _, key := range sortedKeys(r.Nulls) {
			lines = append(lines, "- null `"+key+"`: "+r.Nulls[key])
		}
		if len(lines) == 0 {
			continue
		}
		w("### %s\n\n%s\n\n", r.Game, strings.Join(lines, "\n"))
	}
	return []byte(b.String())
}

func shareCell(s Share) string { return fmt.Sprintf("%d (%s%%)", s.Count, s.Percent) }

func listCell(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, "; ")
}
