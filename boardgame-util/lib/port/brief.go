package port

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/subset"
)

// BriefsDir is where briefs are written, relative to the module root:
// <BriefsDir>/<game>/<ts name>.md.
const BriefsDir = "fabric/transpile/briefs"

// ProposalPath is the plan whose Types and Semantics tables the brief
// carries, relative to the module root.
const ProposalPath = "fabric/TRANSPILE-PROPOSAL.md"

// apiModules are the facade and library modules whose exported
// declarations the brief lists, relative to the engine directory.
var apiModules = []string{"go/api.ts", "go/num.ts", "go/str.ts", "go/slice.ts", "go/panic.ts"}

// signatureTypes are the engine types a body's signature or the facade's
// declarations name, shown whole: module and declaration names.
var signatureTypes = []struct{ module, name string }{
	{"rules/state.ts", "RulesState"},
	{"rules/state.ts", "TimerHandle"},
	{"rules/types.ts", "MoveInstance"},
	{"legal/types.ts", "LegalVerdict"},
}

// predicateTypes are what a predicate constructor (role predicate) builds
// and reads beside the facade.
var predicateTypes = []struct{ module, name string }{
	{"rules/types.ts", "PredicateSpec"},
	{"rules/types.ts", "PredicateContext"},
	{"rules/types.ts", "GamePredicate"},
	{"rules/types.ts", "GamePredicateConstructor"},
	{"legal/types.ts", "LegalRead"},
	{"legal/types.ts", "LegalFacet"},
	{"legal/types.ts", "LegalCost"},
	{"legal/types.ts", "PropertyType"},
	{"legal/types.ts", "LegalBindingValue"},
	{"legal/types.ts", "passVerdict"},
	{"legal/types.ts", "failT"},
	{"legal/types.ts", "unknownVerdict"},
}

// mappedFrameworkTypes are the framework types the proposal's Types table
// maps to a TypeScript shape, beside the receivers of the API table's rows:
// the brief names them and points at SEMANTICS instead of quoting their Go.
var mappedFrameworkTypes = map[string]bool{
	"boardgame.Move": true, "boardgame.State": true, "boardgame.ImmutableState": true,
	"boardgame.PlayerIndex": true, "enum.EnumKey": true, "time.Duration": true,
	"boardgame.SubState": true, "boardgame.ImmutableSubState": true,
}

// Brief is one written brief.
type Brief struct {
	Game string
	Func string
	// TSName is the function the agent writes.
	TSName string
	// Path is where the brief is written, relative to the module root.
	Path string
	// Output is where the agent writes its file, relative to the module root.
	Output string
	Text   []byte
	// SHA256 is the lower-case hex SHA-256 of Text: the agent header's brief:.
	SHA256 string
}

// OutputPath is where an agent writes its translation of ref, relative to
// the module root: beside the generated game, so that its imports
// (../../go/*.ts) resolve as they will inside bodies.gen.ts.
func OutputPath(game, ref string) string {
	return transpile.GamesDir + "/" + game + "/" + tsName(ref) + ".ts"
}

// BriefPath is where the brief for ref is written, relative to the module root.
func BriefPath(game, ref string) string {
	return BriefsDir + "/" + game + "/" + tsName(ref) + ".md"
}

// WriteBrief composes the brief for ref (a function of the game the
// translator refused, or any manifest function), every section from the
// game's source, its definition, the engine's sources, the proposal and
// the example source; nothing in it is written by hand.
func (g *Game) WriteBrief(ref string, examples ExampleSource) (*Brief, error) {
	entry, ok := g.Entry(ref)
	if !ok {
		return nil, fmt.Errorf("port: %s is not a function of %s's manifest", ref, g.Name)
	}
	if _, ok := g.src.units[ref]; !ok {
		return nil, fmt.Errorf("port: no source for %s", ref)
	}
	name := tsName(ref)
	if name == "" {
		return nil, fmt.Errorf("port: %s has no TypeScript name", ref)
	}
	signature, err := g.Signature(ref)
	if err != nil {
		return nil, err
	}
	engine := filepath.Join(g.Root, filepath.FromSlash(EngineDir))
	modules := map[string]tsModule{}
	module := func(rel string) (tsModule, error) {
		if m, ok := modules[rel]; ok {
			return m, nil
		}
		text, err := os.ReadFile(filepath.Join(engine, filepath.FromSlash(rel)))
		if err != nil {
			return tsModule{}, err
		}
		m := parseTSModule(string(text))
		modules[rel] = m
		return m, nil
	}
	proposal, err := os.ReadFile(filepath.Join(g.Root, filepath.FromSlash(ProposalPath)))
	if err != nil {
		return nil, err
	}
	dataTS, err := os.ReadFile(filepath.Join(engine, "rules", "data.ts"))
	if err != nil {
		return nil, err
	}
	convention := callingConvention(string(dataTS))

	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }
	section := func(title string) { w("\n%s\n", title) }

	b.WriteString(`Translate one Go function to TypeScript. Its behavior must be identical,
including behavior that looks like a bug: do not fix, simplify or guard
anything the Go does not guard. If the Go would panic, the TypeScript must
throw the same text (see "Panics").
`)

	section("FUNCTION")
	w("  %s, role %s, from examples/%s/%s:%d at Go source hash %s\n", ref, entry.Role, g.Name, entry.File, entry.Line, entry.SourceHash)

	section("GO SOURCE")
	w("%s\n", indentLines(g.src.unitText(ref), "  "))

	section("WHY THE TRANSLATOR REFUSED IT")
	if len(entry.Refusals) == 0 {
		w("  Nothing: the translator accepts it (this brief is for a function translated mechanically).\n")
	}
	for _, r := range entry.Refusals {
		w("  %s %s, at %s:%d:%d\n", r.Reason, r.Detail, r.File, r.Line, r.Column)
		if line := g.src.lineText(r.File, r.Line); line != "" {
			w("    %s\n", line)
		}
	}

	section("TYPES IT TOUCHES")
	g.writeTypes(&b, ref)

	section("FUNCTIONS IT CALLS (already translated; call them, do not rewrite them)")
	g.writeCallees(&b, ref, entry.Calls)

	section("THE TYPESCRIPT API YOU MAY USE")
	if err := g.writeAPI(&b, ref, entry.Role, module, convention); err != nil {
		return nil, err
	}

	section("SEMANTICS")
	w("  From %s, \"Types\" and \"Semantics that differ\", verbatim.\n\n", ProposalPath)
	for _, heading := range []string{"### Types", "### Semantics that differ"} {
		text := markdownSection(string(proposal), heading)
		if text == "" {
			return nil, fmt.Errorf("port: %s has no section %q", ProposalPath, heading)
		}
		w("%s\n\n", indentLines(text, "  "))
	}

	section("EXAMPLES")
	g.writeExamples(&b, ref, entry.Role, entry.Refusals, examples)

	section("OUTPUT")
	output := OutputPath(g.Name, ref)
	w("  One file, %s, containing exactly:\n", output)
	w("    %s ... }\n", strings.TrimSuffix(signature, " {"))
	w("  That is the body calling convention of rules/data.ts for role %s:\n", entry.Role)
	for _, row := range convention.rows(entry.Role) {
		w("    %s\n", row)
	}
	if detail := convention.details[entry.Role]; detail != "" {
		w("%s\n", indentLines(detail, "    "))
	}
	w(`  Imports only from ../../go/*.ts, with the .ts extension, of the names
  listed under "Names a translation may import" above (the types and the
  verdict functions there come from ../../rules/*.ts and ../../legal/types.ts),
  plus, from ./bodies.gen.ts, the functions listed under "FUNCTIONS IT
  CALLS". Nothing else: no default or namespace import, no renaming.
  No classes, no module-scope let, no objects of functions, no double casts,
  no Math.random, Date, BigInt, async, or anything outside the engine.
`)

	section("HOW TO CHECK")
	w("  boardgame-util port check %s --file %s\n", ref, output)
	b.WriteString(`  It type-checks your file, runs the sandbox and purity checks on it, and
  replays the game's fuzz corpus with your function in place. It prints
  PASS, or the first difference: version, viewer, field, the JSON path, what
  Go recorded and what you produced, and the arguments your function was
  last called with.
`)

	text := []byte(b.String())
	sum := sha256.Sum256(text)
	return &Brief{Game: g.Name, Func: ref, TSName: name, Path: BriefPath(g.Name, ref), Output: output, Text: text, SHA256: hex.EncodeToString(sum[:])}, nil
}

// roleSignatures are the TypeScript signatures of the roles the emitter has
// no rule for (it refuses their Go signatures), from rules/data.ts's body
// calling convention: the parameters and the result after the name.
var roleSignatures = map[string]string{
	"predicate": "(spec: PredicateSpec): GamePredicate | string",
	"counter":   "(currentCount: number, length: number): string | null",
}

// Signature is the declaration line a translation of ref must have
// ("export function <name>(...): <result> {"): the emitter's, by the body
// calling convention, or for a role the emitter has no rule for, that
// role's row of the convention.
func (g *Game) Signature(ref string) (string, error) {
	signature, refused := g.Pkg.Signature(ref)
	if refused == nil {
		return signature, nil
	}
	entry, _ := g.Entry(ref)
	if tail, ok := roleSignatures[entry.Role]; ok {
		return "export function " + tsName(ref) + tail + " {", nil
	}
	return "", fmt.Errorf("port: %s's signature is outside the subset: %s %s", ref, refused.Reason, refused.Detail)
}

// writeTypes writes the types a function names, the PropDefs of the state
// properties it reads or writes, the package variables and constants it
// reads, and the game's moves.
func (g *Game) writeTypes(b *strings.Builder, ref string) {
	w := func(format string, args ...any) { fmt.Fprintf(b, format, args...) }
	n := g.src.collect(ref)
	covered := apiTableReceivers()
	var game, framework []*types.TypeName
	for _, obj := range n.types {
		if obj.Pkg() == g.src.pkg.Types {
			game = append(game, obj)
		} else {
			framework = append(framework, obj)
		}
	}
	w("  The game's types it names, verbatim:\n\n")
	if len(game) == 0 {
		w("    (none)\n")
	}
	for _, obj := range game {
		if text := g.src.declText(obj); text != "" {
			w("%s\n\n", indentLines(text, "    "))
		}
	}
	w("  Framework types it names:\n\n")
	if len(framework) == 0 {
		w("    (none)\n")
	}
	for _, obj := range framework {
		qualified := obj.Pkg().Name() + "." + obj.Name()
		switch {
		case mappedFrameworkTypes[qualified]:
			w("    %s (%s): its TypeScript shape is in the Types table under SEMANTICS.\n", qualified, obj.Pkg().Path())
		case covered[qualified]:
			w("    %s (%s): its methods are the API table's rows (THE TYPESCRIPT API YOU MAY USE).\n", qualified, obj.Pkg().Path())
		default:
			text := g.src.declText(obj)
			if text == "" {
				w("    %s (%s): declaration not found.\n", qualified, obj.Pkg().Path())
				continue
			}
			w("    %s (%s), verbatim:\n\n%s\n\n", qualified, obj.Pkg().Path(), indentLines(text, "      "))
		}
	}
	if len(framework) > 0 {
		b.WriteByte('\n')
	}

	w("  State properties it reads or writes, with their PropDef from the exported definition:\n\n")
	props := g.propDefs(n.selection)
	if len(props) == 0 {
		w("    (none)\n")
	}
	for _, p := range props {
		w("    %s: %s\n", p[0], p[1])
	}

	if len(n.vars) > 0 {
		w("\n  Package variables it reads, verbatim, and every assignment to each:\n\n")
		for _, v := range n.vars {
			w("%s\n", indentLines(g.src.declText(v), "    "))
			for _, a := range g.src.assignments(v) {
				w("      assigned in %s at %s:%d:\n%s\n", a.in, a.file, a.line, indentLines(a.text, "        "))
				for _, from := range a.from {
					w("        where, in that function:\n%s\n", indentLines(from, "          "))
				}
			}
			b.WriteByte('\n')
		}
	}
	if len(n.consts) > 0 {
		w("\n  Package constants it reads, verbatim (the translator folds a constant to its value):\n\n")
		for _, c := range n.consts {
			w("%s\n      = %s\n", indentLines(g.src.declText(c), "    "), c.Val().ExactString())
		}
	}
	if len(n.frameworkConsts) > 0 {
		w("\n  Framework constants it reads, with their values:\n\n")
		for _, c := range n.frameworkConsts {
			w("    %s.%s (%s) = %s\n", c.Pkg().Name(), c.Name(), types.TypeString(c.Type(), func(p *types.Package) string { return p.Name() }), c.Val().ExactString())
		}
	}

	var constructs []string
	for _, p := range g.def.Predicates {
		if p.Construct == ref {
			constructs = append(constructs, p.Name)
		}
	}
	if len(constructs) > 0 {
		w("\n  The exported definition's predicates this function constructs (GameData.Predicates[i].Construct names it):\n\n")
		for _, name := range constructs {
			w("    %q: the loader calls the translation with a spec whose name is %q, once per plan spec naming it\n", name, name)
		}
		w("    Go returns a list of constructors; the translation is the Constructor of the entry with that Name.\n")
	}

	w("\n  The game's moves, from the exported definition (name: Go type, fix-up, stock type):\n\n")
	for _, m := range g.def.Moves {
		stock := "none (moves.Default)"
		if m.Stock != nil {
			stock = "moves." + m.Stock.Type
		}
		w("    %q: %s, fix-up %v, stock %s\n", m.Name, m.GoType, m.IsFixUp, stock)
	}
}

// propDefs are "<owner>.<field>" and the PropDef JSON for each field use.
func (g *Game) propDefs(uses []fieldUse) [][2]string {
	var out [][2]string
	seen := map[string]bool{}
	add := func(label string, raw json.RawMessage) {
		if seen[label] {
			return
		}
		seen[label] = true
		out = append(out, [2]string{label, compactJSON(raw)})
	}
	for _, use := range uses {
		typeName := g.Name + "." + use.recv.Obj().Name()
		switch {
		case g.src.gameState != nil && use.recv.Obj() == g.src.gameState.Obj():
			if raw, ok := g.def.Definition.Game.Props[use.field]; ok {
				add("game."+use.field, raw)
			}
		case g.src.playerState != nil && use.recv.Obj() == g.src.playerState.Obj():
			if raw, ok := g.def.Definition.Player.Props[use.field]; ok {
				add("player."+use.field, raw)
			}
		default:
			matched := false
			for _, m := range g.def.Moves {
				if m.GoType != typeName {
					continue
				}
				matched = true
				if raw, ok := m.Fields[use.field]; ok {
					add(fmt.Sprintf("move %q field %s", m.Name, use.field), raw)
				} else {
					add(fmt.Sprintf("move %q field %s", m.Name, use.field), json.RawMessage(`"a field its base (`+m.Base+`) supplies; not in the definition's Fields"`))
				}
			}
			if matched {
				continue
			}
			for _, deck := range sortedKeys(g.def.Definition.DynamicComponentValues) {
				if raw, ok := g.def.Definition.DynamicComponentValues[deck].Props[use.field]; ok {
					add(fmt.Sprintf("%s (dynamic values of deck %q) .%s", typeName, deck, use.field), raw)
				}
			}
			for _, deck := range sortedKeys(g.def.Definition.Decks) {
				if raw, ok := g.def.Definition.Decks[deck].ComponentProps[use.field]; ok {
					add(fmt.Sprintf("%s (component values of deck %q) .%s", typeName, deck, use.field), raw)
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

func compactJSON(raw json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// writeCallees writes each in-package callee's Go and TypeScript
// signatures, and says which calls are erased.
func (g *Game) writeCallees(b *strings.Builder, ref string, calls []string) {
	w := func(format string, args ...any) { fmt.Fprintf(b, format, args...) }
	if len(calls) == 0 {
		w("  (none)\n")
	}
	for _, callee := range calls {
		w("  %s\n", callee)
		w("    Go:         %s\n", oneLine(g.src.signatureText(callee)))
		if ts := g.calleeSignature(callee); ts != "" {
			w("    TypeScript: %s\n", ts)
			w("    import { %s } from './bodies.gen.ts';\n", tsName(callee))
		} else {
			w("    TypeScript: not translated yet; it must be translated first (port check reports it as a precondition)\n")
		}
	}
	var erased []string
	for _, e := range g.Pkg.Static.Erased {
		if strings.Contains(g.src.unitText(ref), shortName(e.Func)+"(") {
			erased = append(erased, e.Func)
		}
	}
	if len(erased) > 0 {
		w("\n  Erased (no TypeScript function; see the Types table under SEMANTICS):\n")
		for _, e := range erased {
			w("    %s: its results are the game handle (property access through state.g, gameStack(state, ...)) and the players (playerIndexes(state), a player is its index)\n", e)
		}
	}
}

// calleeSignature is the TypeScript declaration line of a callee's block.
func (g *Game) calleeSignature(ref string) string {
	code := ""
	if block, ok := g.Output.Agent[ref]; ok {
		code = block.Code
	} else if fn, ok := g.Function(ref); ok && len(fn.Refusals) == 0 {
		code = fn.Code
	}
	if code == "" {
		return ""
	}
	line := strings.SplitN(code, "\n", 2)[0]
	return strings.TrimSuffix(strings.TrimSpace(line), " {")
}

func shortName(ref string) string {
	if i := strings.LastIndex(ref, "."); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// apiTableReceivers are the framework types whose methods the API table
// lists, as "pkg.Type".
func apiTableReceivers() map[string]bool {
	out := map[string]bool{}
	for _, entry := range subset.Table {
		for _, name := range entry.Go {
			if !strings.HasPrefix(name, "(") {
				continue
			}
			end := strings.Index(name, ")")
			if end < 0 {
				continue
			}
			out[strings.TrimPrefix(name[1:end], "*")] = true
		}
	}
	return out
}

// writeAPI writes the facade's declarations, the types the signatures
// name, a predicate's types, and the API table's rows for the function's
// Go calls.
func (g *Game) writeAPI(b *strings.Builder, ref, role string, module func(string) (tsModule, error), convention callingConventionText) error {
	w := func(format string, args ...any) { fmt.Fprintf(b, format, args...) }

	w("  Names a translation may import (module relative to src/engine/; \"type\" marks a type-only export):\n\n")
	importable := transpile.AgentImportable()
	byModule := map[string][]string{}
	for name, mod := range importable {
		byModule[mod] = append(byModule[mod], name)
	}
	for _, mod := range sortedKeys(byModule) {
		names := byModule[mod]
		sort.Strings(names)
		label := mod
		if strings.HasPrefix(mod, "type:") {
			label = strings.TrimPrefix(mod, "type:") + " (type)"
		}
		w("    %s: %s\n", label, strings.Join(names, ", "))
	}

	w("\n  The API table's rows for each Go call this function makes outside its package (lib/transpile/subset):\n\n")
	calls := g.src.goCalls(ref, subset.QualifiedName)
	if len(calls) == 0 {
		w("    (none)\n")
	}
	for _, call := range calls {
		entry, ok := subset.Lookup(call)
		if !ok {
			w("    %s: not in the API table\n", call)
			continue
		}
		gap := ""
		if entry.Gap() {
			gap = "; ENGINE GAP: " + subset.EngineGaps[entry.Facade]
		}
		w("    %s: %s (%s; %s)%s\n", call, entry.Facade, entry.Kind, entry.Source, gap)
	}
	for _, m := range g.def.Moves {
		if m.Stock == nil || !strings.HasPrefix(ref, g.Name+".(*"+strings.TrimPrefix(m.GoType, g.Name+".")+").") {
			continue
		}
		w("\n    The move embeds moves.%s. Its parts (the API table's m.<Stock>.X rows):\n", m.Stock.Type)
		for _, part := range sortedKeys(subset.StockParts) {
			w("      m.%s.%s: %s\n", m.Stock.Type, part, subset.StockParts[part])
		}
		break
	}

	w("\n  The types the signatures name, verbatim:\n\n")
	for _, t := range signatureTypes {
		mod, err := module(t.module)
		if err != nil {
			return err
		}
		if d, ok := mod.find(t.name); ok {
			w("    // %s\n", t.module)
			writeDecl(b, d, "    ")
			b.WriteByte('\n')
		}
	}

	if role == "predicate" {
		w("  A predicate constructor's types and verdicts (role predicate), verbatim:\n\n")
		for _, t := range predicateTypes {
			mod, err := module(t.module)
			if err != nil {
				return err
			}
			if d, ok := mod.find(t.name); ok {
				w("    // %s\n", t.module)
				writeDecl(b, d, "    ")
				b.WriteByte('\n')
			}
		}
	}

	for _, rel := range apiModules {
		mod, err := module(rel)
		if err != nil {
			return err
		}
		w("  %s\n\n", rel)
		if mod.Header != "" {
			w("%s\n\n", indentLines(mod.Header, "    "))
		}
		for _, d := range mod.Decls {
			writeDecl(b, d, "    ")
			b.WriteByte('\n')
		}
	}
	libDir := filepath.Join(g.Root, filepath.FromSlash(EngineDir), "lib")
	entries, err := os.ReadDir(libDir)
	if err != nil {
		return err
	}
	var libs []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ts") && !strings.HasSuffix(e.Name(), ".test.ts") {
			libs = append(libs, "lib/"+e.Name())
		}
	}
	sort.Strings(libs)
	w("  The engine's libraries (lib/*.ts). A translation reaches them through the\n  facade above (diceRoll, rangeValue, neighborsByKey, ...); their own names\n  are not importable (see OUTPUT). For reference:\n\n")
	for _, rel := range libs {
		mod, err := module(rel)
		if err != nil {
			return err
		}
		w("  %s\n\n", rel)
		if mod.Header != "" {
			w("%s\n\n", indentLines(mod.Header, "    "))
		}
		for _, d := range mod.Decls {
			writeDecl(b, d, "    ")
			b.WriteByte('\n')
		}
	}
	return nil
}

// writeExamples writes three stored pairs, nearest first.
func (g *Game) writeExamples(b *strings.Builder, ref, role string, refusals []transpile.Refusal, source ExampleSource) {
	w := func(format string, args ...any) { fmt.Fprintf(b, format, args...) }
	if source == nil {
		w("  (no example source)\n")
		return
	}
	all, err := source.Examples()
	if err != nil {
		w("  (the examples could not be read: %v)\n", err)
		return
	}
	var reasons []string
	for _, r := range refusals {
		reasons = append(reasons, r.Reason)
	}
	chosen := chooseExamples(all, g.Name, role, reasons, goCallNames(g.src.unitText(ref)), 3)
	w("  From %s, never from %s itself; nearest refusal reason first, then role, then Go calls shared.\n", source.Name(), g.Name)
	if len(chosen) == 0 {
		w("\n  (none stored)\n")
	}
	for i, ex := range chosen {
		reasons := "none (translated mechanically)"
		if len(ex.Reasons) > 0 {
			reasons = strings.Join(ex.Reasons, ", ")
		}
		w("\n  Example %d: %s (game %s, role %s, producer %s, refusals %s)\n\n", i+1, ex.Func, ex.Game, ex.Role, ex.Producer, reasons)
		w("    Go:\n\n%s\n\n", indentLines(ex.GoSource, "      "))
		w("    TypeScript:\n\n%s\n", indentLines(strings.TrimRight(ex.TypeScript, "\n"), "      "))
	}
}

// markdownSection is the body of a markdown section (after its heading
// line, up to the next heading of the same or a higher level), trimmed.
func markdownSection(text, heading string) string {
	lines := strings.Split(text, "\n")
	level := strings.Count(strings.SplitN(heading, " ", 2)[0], "#")
	for i, line := range lines {
		if strings.TrimSpace(line) != heading {
			continue
		}
		var body []string
		for _, l := range lines[i+1:] {
			if strings.HasPrefix(l, "#") {
				hashes := len(l) - len(strings.TrimLeft(l, "#"))
				if hashes <= level && strings.HasPrefix(l[hashes:], " ") {
					break
				}
			}
			body = append(body, l)
		}
		return strings.TrimSpace(strings.Join(body, "\n"))
	}
	return ""
}

// callingConventionText is rules/data.ts's body calling convention: the
// role table's rows and the per-role details.
type callingConventionText struct {
	table   []string
	details map[string]string
}

func (c callingConventionText) rows(role string) []string {
	var out []string
	for _, row := range c.table {
		cells := strings.Split(row, "|")
		if len(cells) < 2 {
			continue
		}
		first := strings.TrimSpace(cells[1])
		if first == role || strings.HasPrefix(first, role+" (") {
			out = append(out, row)
		}
	}
	if len(out) == 0 {
		out = append(out, "(rules/data.ts has no row for role "+role+")")
	}
	return out
}

// callingConvention reads the "## Body calling convention" comment of
// rules/data.ts.
func callingConvention(dataTS string) callingConventionText {
	c := callingConventionText{details: map[string]string{}}
	var comment []string
	in := false
	for _, line := range strings.Split(dataTS, "\n") {
		if !strings.HasPrefix(line, "//") {
			if in {
				break
			}
			continue
		}
		text := strings.TrimPrefix(strings.TrimPrefix(line, "//"), " ")
		if strings.HasPrefix(text, "## Body calling convention") {
			in = true
			continue
		}
		if in && strings.HasPrefix(text, "## ") {
			break
		}
		if in {
			comment = append(comment, text)
		}
	}
	var current string
	var detail []string
	flush := func() {
		if current != "" {
			c.details[current] = strings.TrimSpace(strings.Join(detail, "\n"))
		}
		current, detail = "", nil
	}
	for _, text := range comment {
		if strings.HasPrefix(text, "| ") && !strings.HasPrefix(text, "| Role") {
			c.table = append(c.table, text)
			continue
		}
		if strings.HasPrefix(text, "- ") {
			flush()
			if colon := strings.Index(text, ":"); colon > 2 && !strings.Contains(text[2:colon], " ") {
				current = text[2:colon]
				detail = []string{text}
			}
			continue
		}
		if current != "" {
			if strings.TrimSpace(text) == "" {
				flush()
				continue
			}
			detail = append(detail, text)
		}
	}
	flush()
	return c
}
