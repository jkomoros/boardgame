package pairs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// Extracted is one hand pair as server/static/scripts/pairs-extract.mjs
// cuts it from a hand port.
type Extracted struct {
	Game       string `json:"game"`
	Func       string `json:"func"`
	Role       string `json:"role"`
	Name       string `json:"name"`
	Extraction string `json:"extraction"`
	Note       string `json:"note"`
	Accepted   string `json:"accepted"`
}

// NoPair is a manifest function the hand port has no counterpart for.
type NoPair struct {
	Game   string `json:"game"`
	Func   string `json:"func"`
	Role   string `json:"role"`
	Reason string `json:"reason"`
}

// Extraction is pairs-extract.mjs's output.
type Extraction struct {
	Pairs []Extracted `json:"pairs"`
	None  []NoPair    `json:"none"`
}

// Extract runs pairs-extract.mjs.
func Extract(node, root string) (*Extraction, error) {
	out, err := RunScript(node, root, "pairs-extract.mjs", nil)
	if err != nil {
		return nil, err
	}
	var ex Extraction
	if err := json.Unmarshal(out, &ex); err != nil {
		return nil, fmt.Errorf("pairs-extract.mjs output: %w", err)
	}
	return &ex, nil
}

// Manifests loads each game's static report from its Go source.
func Manifests(root string, games []string) (map[string]*gamedef.StaticReport, error) {
	out := map[string]*gamedef.StaticReport{}
	for _, game := range games {
		report, err := gamedef.LoadStatic("./examples/"+game, root)
		if err != nil {
			return nil, err
		}
		out[game] = report
	}
	return out, nil
}

// Build is the store's files for an extraction, by path relative to the
// store directory: each pair's three files and index.ts. previous are the
// pairs on disk, whose measured fuzz fields carry over while what they
// measured is unchanged (the Go source for blocks, accepted.ts for
// mutants).
func Build(root string, ex *Extraction, manifests map[string]*gamedef.StaticReport, previous []Pair) (map[string][]byte, error) {
	old := map[string]Pair{}
	for _, p := range previous {
		old[p.Game+"/"+p.Name] = p
	}
	files := map[string][]byte{}
	seeds := map[string][]int{}
	spans := map[string]map[string]Span{}
	for game, report := range manifests {
		s, err := Spans(report)
		if err != nil {
			return nil, err
		}
		spans[game] = s
		if seeds[game], err = Seeds(root, game); err != nil {
			return nil, err
		}
	}
	var stored []Extracted
	for _, e := range ex.Pairs {
		report := manifests[e.Game]
		if report == nil {
			return nil, fmt.Errorf("%s: no manifest for %s", e.Func, e.Game)
		}
		var entry *gamedef.FunctionEntry
		for i := range report.Functions {
			if report.Functions[i].Func == e.Func {
				entry = &report.Functions[i]
			}
		}
		if entry == nil {
			return nil, fmt.Errorf("%s: not in the manifest of %s", e.Func, e.Game)
		}
		name, err := transpile.Mangle(e.Func)
		if err != nil {
			return nil, err
		}
		if name != e.Name {
			return nil, fmt.Errorf("%s: pairs-extract.mjs mangles it %s, transpile.Mangle %s", e.Func, e.Name, name)
		}
		refusals := entry.Refusals
		if refusals == nil {
			refusals = []transpile.Refusal{}
		}
		meta := Meta{
			Func: e.Func, Role: entry.Role, SourceHash: entry.SourceHash, Producer: ProducerHand, Refusals: refusals,
			Extraction: e.Extraction, Note: e.Note,
			Fuzz: Fuzz{Seeds: seeds[e.Game]},
		}
		if prev, ok := old[e.Game+"/"+name]; ok {
			if prev.Meta.SourceHash == meta.SourceHash {
				meta.Fuzz.BlocksCovered, meta.Fuzz.BlocksTotal, meta.Fuzz.Unreached = prev.Meta.Fuzz.BlocksCovered, prev.Meta.Fuzz.BlocksTotal, prev.Meta.Fuzz.Unreached
			}
			if string(prev.Accepted) == e.Accepted {
				meta.Fuzz.MutantsKilled, meta.Fuzz.MutantsTotal, meta.Fuzz.Survivors = prev.Meta.Fuzz.MutantsKilled, prev.Meta.Fuzz.MutantsTotal, prev.Meta.Fuzz.Survivors
				meta.Fuzz.Why = prev.Meta.Fuzz.Why
			}
		}
		dir := e.Game + "/" + name + "/"
		// Verbatim: exactly the bytes gamedef hashed, with no newline added.
		files[dir+SourceFile] = append([]byte{}, spans[e.Game][e.Func].Source...)
		files[dir+AcceptedFile] = []byte(e.Accepted)
		blob, err := MarshalMeta(meta)
		if err != nil {
			return nil, err
		}
		files[dir+MetaFile] = blob
		stored = append(stored, e)
	}
	files["index.ts"] = []byte(IndexText(stored))
	return files, nil
}

// MarshalMeta is meta.json's text.
func MarshalMeta(meta Meta) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(meta); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// IndexText is the store's index.ts: every accepted function, by game and
// FnRef, for games/pairs.test.ts.
func IndexText(pairs []Extracted) string {
	sorted := append([]Extracted{}, pairs...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Game != sorted[j].Game {
			return sorted[i].Game < sorted[j].Game
		}
		return sorted[i].Name < sorted[j].Name
	})
	var b strings.Builder
	b.WriteString("// Code generated by boardgame-util/lib/transpile/pairs/testdata/gen. DO NOT EDIT.\n")
	b.WriteString("//\n// Every stored pair's accepted function, by game and FnRef, for\n")
	b.WriteString("// server/static/src/engine/games/pairs.testkit.test.ts. tsconfig.json beside this\n")
	b.WriteString("// file type-checks the store under the engine's sandbox options.\n\n")
	for _, p := range sorted {
		fmt.Fprintf(&b, "import { %s } from './%s/%s/accepted.ts';\n", p.Name, p.Game, p.Name)
	}
	b.WriteString("\nexport type AcceptedBody = (...args: never[]) => unknown;\n")
	b.WriteString("\n/** The stored accepted function for a FnRef of a game, or undefined. */\n")
	b.WriteString("export function accepted(game: string, ref: string): AcceptedBody | undefined {\n")
	b.WriteString("\tswitch (game + ' ' + ref) {\n")
	for _, p := range sorted {
		fmt.Fprintf(&b, "\t\tcase %s:\n\t\t\treturn %s;\n", jsString(p.Game+" "+p.Func), p.Name)
	}
	b.WriteString("\t}\n\treturn undefined;\n}\n")
	b.WriteString("\n/** Every FnRef of a game the store holds a pair for, sorted. */\n")
	b.WriteString("export function pairRefs(game: string): string[] {\n\tswitch (game) {\n")
	byGame := map[string][]string{}
	var games []string
	for _, p := range sorted {
		if byGame[p.Game] == nil {
			games = append(games, p.Game)
		}
		byGame[p.Game] = append(byGame[p.Game], p.Func)
	}
	for _, game := range games {
		refs := byGame[game]
		sort.Strings(refs)
		quoted := make([]string, len(refs))
		for i, r := range refs {
			quoted[i] = jsString(r)
		}
		fmt.Fprintf(&b, "\t\tcase %s:\n\t\t\treturn [%s];\n", jsString(game), strings.Join(quoted, ", "))
	}
	b.WriteString("\t}\n\treturn [];\n}\n")
	return b.String()
}

func jsString(s string) string {
	blob, _ := json.Marshal(s)
	return "'" + strings.ReplaceAll(strings.ReplaceAll(string(blob[1:len(blob)-1]), `\"`, `"`), "'", `\'`) + "'"
}

// WriteStore writes files under the store and removes hand pairs the
// extraction no longer has. It returns the paths written and removed.
func WriteStore(root string, files map[string][]byte, previous []Pair) (written, removed []string, err error) {
	store := filepath.Join(root, filepath.FromSlash(StoreDir))
	keep := map[string]bool{}
	for rel := range files {
		keep[filepath.Dir(rel)] = true
	}
	for _, p := range previous {
		if p.Meta.Producer != ProducerHand || keep[p.Game+"/"+p.Name] {
			continue
		}
		if err := os.RemoveAll(p.Dir); err != nil {
			return nil, nil, err
		}
		removed = append(removed, p.Game+"/"+p.Name)
	}
	paths := make([]string, 0, len(files))
	for rel := range files {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	for _, rel := range paths {
		full := filepath.Join(store, filepath.FromSlash(rel))
		if existing, err := os.ReadFile(full); err == nil && bytes.Equal(existing, files[rel]) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return nil, nil, err
		}
		if err := os.WriteFile(full, files[rel], 0644); err != nil {
			return nil, nil, err
		}
		written = append(written, rel)
	}
	return written, removed, nil
}

// UpdateMeta rewrites one pair's meta.json through f.
func UpdateMeta(p Pair, f func(*Meta)) error {
	meta := p.Meta
	f(&meta)
	blob, err := MarshalMeta(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.Dir, MetaFile), blob, 0644)
}
