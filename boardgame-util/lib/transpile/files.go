package transpile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GamesDir is where generated games live, relative to the module root.
const GamesDir = "server/static/src/engine/games"

// DefinitionsDir is where exported definitions live, relative to the module
// root.
const DefinitionsDir = "conformance/definitions"

// Output is the three files of a generated game, by file name.
type Output struct {
	Game  string
	Files map[string][]byte
	// Functions is the translator's answer for every manifest function.
	Functions []Function
	// Agent are the agent-written blocks Generate kept from the existing
	// bodies file, by FnRef. Their Functions entry is still the
	// translator's own answer (a refusal, or a translation not used).
	Agent map[string]Block
}

// FileNames are a generated game's files, in the order they are written.
var FileNames = []string{"bodies.gen.ts", "definition.gen.ts", "index.ts"}

// header is the line before each generated function.
var header = regexp.MustCompile(`^// @generated transpile@(\d+) go:(\S+) (sha256:[0-9a-f]{64}) first:(\S+)$`)

// agentHeader is the line before a function an agent translated and
// `boardgame-util port accept` copied in (fabric/TRANSPILE-PROPOSAL.md,
// "Validation and storage"): the producer's model id, the SHA-256 of the
// brief it was given, the Go function and its source hash, and the commit
// at which the text was first written.
var agentHeader = regexp.MustCompile(`^// @generated agent:(\S+) brief:([0-9a-f]{64}) go:(\S+) (sha256:[0-9a-f]{64}) first:(\S+)$`)

// AgentHeader is the header line of an agent-written block.
func AgentHeader(model, brief, ref, sourceHash, first string) string {
	return fmt.Sprintf("// @generated agent:%s brief:%s go:%s %s first:%s", model, brief, ref, sourceHash, first)
}

// existingBlock is one function block of a bodies file already on disk.
type existingBlock struct {
	first string
	code  string
	// sourceHash is the Go source hash the header names.
	sourceHash string
	// agent and brief are set for an agent-written block: its model id and
	// the brief's SHA-256. version is a translator block's transpile@N.
	agent, brief, version string
}

// Block is one function block of a bodies.gen.ts, as its header names it.
type Block struct {
	Func string
	// Producer is "transpile@<N>" (the N the header names) or
	// "agent:<model id>".
	Producer string
	// Agent is the model id of an agent-written block, "" otherwise.
	Agent string
	// Brief is an agent-written block's brief SHA-256, "" otherwise.
	Brief      string
	SourceHash string
	First      string
	// Code is the function text under the header.
	Code string
}

// Blocks reads every function block of a bodies.gen.ts, by FnRef.
func Blocks(text []byte) map[string]Block {
	out := make(map[string]Block)
	for ref, b := range parseBodies(text) {
		block := Block{Func: ref, Agent: b.agent, Brief: b.brief, SourceHash: b.sourceHash, First: b.first, Code: b.code}
		if b.agent != "" {
			block.Producer = "agent:" + b.agent
		} else {
			block.Producer = "transpile@" + b.version
		}
		out[ref] = block
	}
	return out
}

// WithoutBlock is text with the function block of ref (its header and its
// code) removed, so that the next Generate writes the translator's answer
// for it. `port --adopt` replaces an agent-written block this way.
func WithoutBlock(text []byte, ref string) []byte {
	var out bytes.Buffer
	skipping := false
	for _, line := range strings.SplitAfter(string(text), "\n") {
		bare := strings.TrimSuffix(line, "\n")
		if m := header.FindStringSubmatch(bare); m != nil {
			skipping = m[2] == ref
		} else if m := agentHeader.FindStringSubmatch(bare); m != nil {
			skipping = m[3] == ref
		}
		if !skipping {
			out.WriteString(line)
		} else if bare == "}" {
			skipping = false
		}
	}
	return out.Bytes()
}

// WithAgentBlock is text with ref's block replaced by (or, when text has
// none, followed by) an agent-written block: AgentHeader's line and code,
// which must be one function declaration ending in a line "}". The next
// Generate keeps it in its place in callee-first order.
func WithAgentBlock(text []byte, ref, model, brief, sourceHash, first, code string) []byte {
	out := WithoutBlock(text, ref)
	if !strings.HasSuffix(code, "\n") {
		code += "\n"
	}
	return append(out, []byte("\n"+AgentHeader(model, brief, ref, sourceHash, first)+"\n"+code)...)
}

// agentImportable is every name an agent-written block may use from
// outside bodies.gen.ts: the emitter's facade names, and what a predicate
// constructor (role predicate, rules/data.ts) needs beside them. Generate
// imports each one an agent block uses.
func agentImportable(name string) (string, bool) {
	if module, ok := facadeNames[name]; ok {
		return module, true
	}
	module, ok := agentExtraNames[name]
	return module, ok
}

// AgentImportable is every name an agent-written block may use from
// outside bodies.gen.ts, with its module relative to the engine directory
// ("go/api.ts"); a module prefixed "type:" exports a type only.
func AgentImportable() map[string]string {
	out := make(map[string]string, len(facadeNames)+len(agentExtraNames))
	for name, module := range facadeNames {
		out[name] = module
	}
	for name, module := range agentExtraNames {
		out[name] = module
	}
	return out
}

var agentExtraNames = map[string]string{
	"passVerdict": modLegal, "failT": modLegal, "unknownVerdict": modLegal,
	"LegalRead": typesOnly + modLegal, "PropertyType": typesOnly + modLegal, "LegalCost": typesOnly + modLegal,
	"PredicateSpec": typesOnly + modTypes, "PredicateContext": typesOnly + modTypes, "GamePredicate": typesOnly + modTypes,
}

var tsIdentifier = regexp.MustCompile(`[A-Za-z_$][A-Za-z0-9_$]*`)

// agentImports are the modules and names an agent block's code uses:
// every identifier outside comments and string literals that is not a
// property name (after '.') and that agentImportable knows.
func agentImports(code string) map[string][]string {
	set := make(map[string]map[string]bool)
	text := blankCommentsAndStrings(code)
	for _, loc := range tsIdentifier.FindAllStringIndex(text, -1) {
		if loc[0] > 0 && (text[loc[0]-1] == '.' || isIdentByte(text[loc[0]-1])) {
			continue
		}
		name := text[loc[0]:loc[1]]
		module, ok := agentImportable(name)
		if !ok {
			continue
		}
		if set[module] == nil {
			set[module] = make(map[string]bool)
		}
		set[module][name] = true
	}
	return sortedImports(set)
}

func isIdentByte(b byte) bool {
	return b == '_' || b == '$' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// blankCommentsAndStrings replaces every comment and string or template
// literal of TypeScript text with spaces, keeping offsets.
func blankCommentsAndStrings(text string) string {
	out := []byte(text)
	for i := 0; i < len(out); i++ {
		switch {
		case out[i] == '/' && i+1 < len(out) && out[i+1] == '/':
			for ; i < len(out) && out[i] != '\n'; i++ {
				out[i] = ' '
			}
		case out[i] == '/' && i+1 < len(out) && out[i+1] == '*':
			for ; i < len(out) && !(out[i] == '*' && i+1 < len(out) && out[i+1] == '/'); i++ {
				out[i] = ' '
			}
			if i+1 < len(out) {
				out[i], out[i+1] = ' ', ' '
				i++
			}
		case out[i] == '\'' || out[i] == '"' || out[i] == '`':
			quote := out[i]
			out[i] = ' '
			for i++; i < len(out) && out[i] != quote; i++ {
				if out[i] == '\\' && i+1 < len(out) {
					out[i] = ' '
					i++
				}
				if out[i] != '\n' {
					out[i] = ' '
				}
			}
			if i < len(out) {
				out[i] = ' '
			}
		}
	}
	return string(out)
}

// parseBodies reads the function blocks of an existing bodies.gen.ts: per
// FnRef, the commit its header names and the code under it.
func parseBodies(text []byte) map[string]existingBlock {
	blocks := make(map[string]existingBlock)
	scanner := bufio.NewScanner(bytes.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	ref := ""
	var current existingBlock
	var code strings.Builder
	flush := func() {
		if ref != "" {
			current.code = code.String()
			blocks[ref] = current
		}
		ref, current = "", existingBlock{}
		code.Reset()
	}
	for scanner.Scan() {
		line := scanner.Text()
		if m := header.FindStringSubmatch(line); m != nil {
			flush()
			ref, current = m[2], existingBlock{version: m[1], first: m[4], sourceHash: m[3]}
			continue
		}
		if m := agentHeader.FindStringSubmatch(line); m != nil {
			flush()
			ref, current = m[3], existingBlock{first: m[5], sourceHash: m[4], agent: m[1], brief: m[2]}
			continue
		}
		if ref == "" {
			continue
		}
		code.WriteString(line)
		code.WriteByte('\n')
		if line == "}" {
			flush()
		}
	}
	flush()
	return blocks
}

// Generate builds the generated game's files. definitionJSON is the
// game's conformance/definitions/<game>.json; existing is the current
// bodies.gen.ts (nil when absent), whose first: commits are kept for every
// function whose code is unchanged; commit is the commit recorded for a
// function written or changed now.
//
// An agent-written block of existing (its header is AgentHeader's) is kept
// verbatim, header included, for as long as its function's Go source hash
// is unchanged, whether or not the translator can now translate the
// function: the checked-in text changes only when someone adopts the
// translator's output (`port --adopt`, which removes the block first, see
// WithoutBlock). Its imports are recovered from the names it uses
// (agentImportable), and `bodies` resolves it like any other block. An
// agent block whose hash no longer matches its function's is dropped, and
// the function is translated or refused afresh.
func (p *Package) Generate(definitionJSON, existing []byte, commit string) (*Output, error) {
	functions := p.Translate()
	game := p.Name()
	old := parseBodies(existing)

	imports := make(map[string]map[string]bool)
	var blocks strings.Builder
	var cases []string
	var refused []Function
	agents := make(map[string]Block)
	for _, fn := range functions {
		if block, ok := old[fn.Func]; ok && block.agent != "" && block.sourceHash == fn.SourceHash && fn.Name != "" {
			for module, names := range agentImports(block.code) {
				if imports[module] == nil {
					imports[module] = make(map[string]bool)
				}
				for _, name := range names {
					imports[module][name] = true
				}
			}
			fmt.Fprintf(&blocks, "\n%s\n%s", AgentHeader(block.agent, block.brief, fn.Func, block.sourceHash, block.first), block.code)
			agents[fn.Func] = Block{Func: fn.Func, Producer: "agent:" + block.agent, Agent: block.agent, Brief: block.brief, SourceHash: block.sourceHash, First: block.first, Code: block.code}
			if fn.Role != "helper" {
				cases = append(cases, fmt.Sprintf("\t\tcase %s:\n\t\t\treturn %s;\n", jsString(fn.Func), fn.Name))
			}
			continue
		}
		if len(fn.Refusals) > 0 {
			refused = append(refused, fn)
			continue
		}
		for module, names := range fn.Imports {
			if imports[module] == nil {
				imports[module] = make(map[string]bool)
			}
			for _, name := range names {
				imports[module][name] = true
			}
		}
		first := commit
		if block, ok := old[fn.Func]; ok && block.code == fn.Code {
			first = block.first
		}
		fmt.Fprintf(&blocks, "\n// @generated transpile@%d go:%s %s first:%s\n%s", Version, fn.Func, fn.SourceHash, first, fn.Code)
		if fn.Role != "helper" {
			cases = append(cases, fmt.Sprintf("\t\tcase %s:\n\t\t\treturn %s;\n", jsString(fn.Func), fn.Name))
		}
	}
	sort.Strings(cases)

	var out bytes.Buffer
	fmt.Fprintf(&out, "// Code generated by boardgame-util transpile from %s. DO NOT EDIT.\n", p.Static.Package)
	out.WriteString("//\n")
	out.WriteString("// Each function below is one Go function of the game, translated mechanically\n")
	out.WriteString("// (fabric/TRANSPILE-PROPOSAL.md, \"The translator\") against the Go-shaped\n")
	out.WriteString("// facade in ../../go/. Its header names the Go function, the SHA-256 of its\n")
	out.WriteString("// gofmt-ed source, and the commit at which this text was first written.\n")
	out.WriteString("// `bodies` resolves the definition's FnRefs (rules/data.ts, the body calling\n")
	out.WriteString("// convention).\n")
	if len(refused) > 0 {
		out.WriteString("//\n// Refused, so not translated (the loader cannot build the game until each\n// has a body):\n")
		for _, fn := range refused {
			fmt.Fprintf(&out, "//   %s\n", fn.Func)
			for _, r := range fn.Refusals {
				fmt.Fprintf(&out, "//     %s %s (%s:%d)\n", r.Reason, r.Detail, r.File, r.Line)
			}
		}
	}
	out.WriteString("\n")
	out.WriteString(importLines(imports))
	out.WriteString(blocks.String())
	out.WriteString("\n/** The resolver gameFromData calls for every FnRef the definition names. */\n")
	out.WriteString("export function bodies(ref: string): ((...args: never[]) => unknown) | undefined {\n")
	if len(cases) > 0 {
		out.WriteString("\tswitch (ref) {\n")
		for _, c := range cases {
			out.WriteString(c)
		}
		out.WriteString("\t}\n")
	} else {
		out.WriteString("\tvoid ref;\n")
	}
	out.WriteString("\treturn undefined;\n}\n")

	definition, err := definitionFile(game, definitionJSON)
	if err != nil {
		return nil, err
	}
	index := fmt.Sprintf(`// Code generated by boardgame-util transpile from %s. DO NOT EDIT.

import { gameFromData } from '../../rules/data.ts';
import type { GameSpec } from '../../rules/types.ts';
import { bodies } from './bodies.gen.ts';
import { definition } from './definition.gen.ts';

/** %s, built from its exported definition and its translated Go functions. */
export function %sGenerated(): GameSpec {
	return gameFromData(definition(), bodies);
}
`, p.Static.Package, game, game)
	return &Output{
		Game:      game,
		Functions: functions,
		Agent:     agents,
		Files: map[string][]byte{
			"bodies.gen.ts":     out.Bytes(),
			"definition.gen.ts": definition,
			"index.ts":          []byte(index),
		},
	}, nil
}

func importLines(imports map[string]map[string]bool) string {
	type group struct{ values, types []string }
	modules := make(map[string]*group)
	for module, names := range imports {
		path := strings.TrimPrefix(module, typesOnly)
		g := modules[path]
		if g == nil {
			g = &group{}
			modules[path] = g
		}
		for name := range names {
			if strings.HasPrefix(module, typesOnly) {
				g.types = append(g.types, name)
			} else {
				g.values = append(g.values, name)
			}
		}
	}
	paths := make([]string, 0, len(modules))
	for path := range modules {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var out strings.Builder
	for _, path := range paths {
		g := modules[path]
		sort.Strings(g.values)
		sort.Strings(g.types)
		if len(g.values) > 0 {
			fmt.Fprintf(&out, "import { %s } from '../../%s';\n", strings.Join(g.values, ", "), path)
		}
		if len(g.types) > 0 {
			fmt.Fprintf(&out, "import type { %s } from '../../%s';\n", strings.Join(g.types, ", "), path)
		}
	}
	return out.String()
}

// definitionFile is definition.gen.ts: a function returning the GameData
// literal (no JSON import).
func definitionFile(game string, definitionJSON []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(definitionJSON))
	decoder.UseNumber()
	var literal strings.Builder
	if err := writeLiteral(&literal, decoder, 1); err != nil {
		return nil, fmt.Errorf("transpile: %s's definition is not JSON: %w", game, err)
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "// Code generated by boardgame-util transpile from %s/%s.json. DO NOT EDIT.\n\n", DefinitionsDir, game)
	out.WriteString("import type { GameData } from '../../rules/data.ts';\n\n")
	fmt.Fprintf(&out, "/** %s's exported definition (boardgame-util gamedef export), as a literal. */\n", game)
	out.WriteString("export function definition(): GameData {\n\treturn ")
	out.WriteString(literal.String())
	out.WriteString(";\n}\n")
	return out.Bytes(), nil
}

// writeLiteral writes the next JSON value of decoder as a TypeScript
// literal, tab-indented, object keys in the document's order.
func writeLiteral(out *strings.Builder, decoder *json.Decoder, depth int) error {
	indent := strings.Repeat("\t", depth)
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch x := token.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		out.WriteString(strconv.FormatBool(x))
	case json.Number:
		out.WriteString(x.String())
	case string:
		out.WriteString(jsString(x))
	case json.Delim:
		closing := "]"
		if x == '{' {
			closing = "}"
		}
		if !decoder.More() {
			if _, err := decoder.Token(); err != nil {
				return err
			}
			out.WriteString(string(x) + closing)
			return nil
		}
		out.WriteString(string(x) + "\n")
		for decoder.More() {
			out.WriteString(indent + "\t")
			if x == '{' {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("object key %v", key)
				}
				out.WriteString(propertyName(name) + ": ")
			}
			if err := writeLiteral(out, decoder, depth+1); err != nil {
				return err
			}
			out.WriteString(",\n")
		}
		if _, err := decoder.Token(); err != nil {
			return err
		}
		out.WriteString(indent + closing)
	}
	return nil
}

var plainKey = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func propertyName(k string) string {
	if plainKey.MatchString(k) {
		return k
	}
	return jsString(k)
}

// Signature is the TypeScript declaration line the emitter writes for ref
// ("export function <name>(<params>): <result> {"), by the body calling
// convention, without translating the body: what a caller translated by
// the generic rule will pass, and what an agent's translation of a refused
// function must declare. It refuses only when the signature itself is
// outside the subset. It is emitter.function's head, not its body.
func (p *Package) Signature(ref string) (line string, refused *Refusal) {
	node, ok := p.units[ref]
	if !ok {
		return "", &Refusal{Func: ref, Reason: ReasonUnknownCall, Detail: "emitter: no syntax for " + ref}
	}
	e := &emitter{p: p, ref: ref, role: p.roles[ref], node: node, locals: make(map[types.Object]*local), used: make(map[string]bool), imports: make(map[string]map[string]bool)}
	defer func() {
		if r := recover(); r != nil {
			rp, ok := r.(refusalPanic)
			if !ok {
				panic(r)
			}
			line, refused = "", &rp.r
		}
	}()
	var recv *ast.FieldList
	var ftype *ast.FuncType
	switch n := node.(type) {
	case *ast.FuncDecl:
		recv, ftype, e.body = n.Recv, n.Type, n.Body
		e.sig = p.info.Defs[n.Name].(*types.Func).Type().(*types.Signature)
	case *ast.FuncLit:
		ftype, e.body = n.Type, n.Body
		e.sig = p.info.TypeOf(n).(*types.Signature)
	}
	for name := range map[string]bool{"state": true, "base": true, "deck": true, "index": true} {
		e.used[name] = true
	}
	var goParams []types.Object
	if recv != nil && len(recv.List) > 0 {
		goParams = append(goParams, e.fieldObjects(recv)...)
	}
	goParams = append(goParams, e.fieldObjects(ftype.Params)...)
	e.results = e.sig.Results()
	name, err := Mangle(ref)
	if err != nil {
		return "", &Refusal{Func: ref, Reason: ReasonUnknownType, Detail: "emitter: " + err.Error()}
	}
	params, ret := e.signature(goParams)
	var list []string
	for _, prm := range params {
		list = append(list, prm.name+": "+prm.ts)
	}
	return fmt.Sprintf("export function %s(%s): %s {", name, strings.Join(list, ", "), ret), nil
}

// ModuleRoot is the directory holding go.mod above dir.
func ModuleRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for d := abs; ; {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", fmt.Errorf("transpile: no go.mod above %s", dir)
		}
		d = parent
	}
}
