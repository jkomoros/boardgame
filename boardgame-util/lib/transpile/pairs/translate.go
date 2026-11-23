package pairs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// Translation is the translator's answer for one stored pair's Go source,
// as a module that can stand in for the pair's accepted.ts.
type Translation struct {
	Game string
	Func string
	Name string
	// Accepted reports whether the translator accepted the function and
	// every in-package callee it reaches.
	Accepted bool
	// Refusals are the function's (or, prefixed "callee ", a callee's).
	Refusals []transpile.Refusal
	// Code is the function's own declaration, as the translator wrote it.
	Code string
	// Module is a self-contained TypeScript module: the facade imports (by
	// absolute path), every callee's declaration, then the function, the
	// only export. Empty unless Accepted.
	Module string
	// Callees are the in-package functions bundled, callee first.
	Callees []string
}

// Rewrite edits a function's translation before it is bundled: a test hook
// for a deliberately wrong rule.
type Rewrite func(fn *transpile.Function)

// Translate translates every function of one game and returns the
// translation of each FnRef in refs.
func Translate(root, game string, refs []string, rewrite Rewrite) (map[string]Translation, error) {
	pkg, err := transpile.Load(fmtGamePackage(game), root)
	if err != nil {
		return nil, err
	}
	functions := pkg.Translate()
	byFunc := map[string]*transpile.Function{}
	order := map[string]int{}
	for i := range functions {
		fn := &functions[i]
		if rewrite != nil {
			rewrite(fn)
		}
		byFunc[fn.Func] = fn
		order[fn.Func] = i
	}
	calls := map[string][]string{}
	for _, entry := range pkg.Static.Functions {
		calls[entry.Func] = entry.Calls
	}
	engine := EngineDir(root)
	out := map[string]Translation{}
	for _, ref := range refs {
		fn := byFunc[ref]
		if fn == nil {
			return nil, fmt.Errorf("%s: the translator has no function %s", game, ref)
		}
		t := Translation{Game: game, Func: ref, Name: fn.Name, Code: fn.Code, Refusals: fn.Refusals}
		// Every callee reached, transitively.
		seen := map[string]bool{ref: true}
		queue := append([]string{}, calls[ref]...)
		for len(queue) > 0 {
			callee := queue[0]
			queue = queue[1:]
			if seen[callee] {
				continue
			}
			seen[callee] = true
			t.Callees = append(t.Callees, callee)
			queue = append(queue, calls[callee]...)
		}
		sort.Slice(t.Callees, func(i, j int) bool { return order[t.Callees[i]] < order[t.Callees[j]] })
		t.Accepted = len(fn.Refusals) == 0
		for _, callee := range t.Callees {
			c := byFunc[callee]
			if c == nil {
				t.Accepted = false
				t.Refusals = append(t.Refusals, transpile.Refusal{Func: callee, Reason: "callee", Detail: "not in the translator's output"})
				continue
			}
			if len(c.Refusals) > 0 {
				t.Accepted = false
				for _, r := range c.Refusals {
					r.Detail = "callee " + callee + ": " + r.Detail
					t.Refusals = append(t.Refusals, r)
				}
			}
		}
		if t.Accepted {
			bundle := []*transpile.Function{}
			for _, callee := range t.Callees {
				bundle = append(bundle, byFunc[callee])
			}
			bundle = append(bundle, fn)
			t.Module = module(engine, bundle)
		}
		out[ref] = t
	}
	return out, nil
}

func fmtGamePackage(game string) string {
	return "github.com/jkomoros/boardgame/examples/" + game
}

// typesOnly is the translator's prefix for a type-only import module.
const typesOnly = "type:"

// module renders a bundle: the last function exported, the rest not.
func module(engine string, bundle []*transpile.Function) string {
	values := map[string]map[string]bool{}
	types := map[string]map[string]bool{}
	for _, fn := range bundle {
		for mod, names := range fn.Imports {
			target := values
			if strings.HasPrefix(mod, typesOnly) {
				target, mod = types, strings.TrimPrefix(mod, typesOnly)
			}
			if target[mod] == nil {
				target[mod] = map[string]bool{}
			}
			for _, name := range names {
				target[mod][name] = true
			}
		}
	}
	var b strings.Builder
	last := bundle[len(bundle)-1]
	fmt.Fprintf(&b, "// The translator's output (transpile@%d) for %s and the in-package functions it calls,\n// written for a rule proof (boardgame-util/lib/transpile/pairs). Not checked in.\n\n", transpile.Version, last.Func)
	b.WriteString(importBlock(engine, values, false))
	b.WriteString(importBlock(engine, types, true))
	for i, fn := range bundle {
		code := fn.Code
		if i < len(bundle)-1 {
			code = strings.Replace(code, "export function ", "function ", 1)
		}
		fmt.Fprintf(&b, "\n// go:%s %s\n%s", fn.Func, fn.SourceHash, code)
		if !strings.HasSuffix(code, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func importBlock(engine string, imports map[string]map[string]bool, typeOnly bool) string {
	mods := make([]string, 0, len(imports))
	for mod := range imports {
		mods = append(mods, mod)
	}
	sort.Strings(mods)
	var b strings.Builder
	for _, mod := range mods {
		names := make([]string, 0, len(imports[mod]))
		for name := range imports[mod] {
			names = append(names, name)
		}
		sort.Strings(names)
		keyword := "import"
		if typeOnly {
			keyword = "import type"
		}
		fmt.Fprintf(&b, "%s { %s } from '%s';\n", keyword, strings.Join(names, ", "), filepath.ToSlash(filepath.Join(engine, filepath.FromSlash(mod))))
	}
	return b.String()
}

// ExportedText is the text of the exported function of an accepted.ts,
// from its `export function` line to the closing brace at column 0.
func ExportedText(accepted []byte, name string) string {
	text := string(accepted)
	start := strings.Index(text, "export function "+name+"(")
	if start < 0 {
		return ""
	}
	end := strings.Index(text[start:], "\n}")
	if end < 0 {
		return text[start:]
	}
	return text[start : start+end+2]
}

// TextuallyEqual reports whether the translator's declaration and the
// pair's exported function are the same text, whitespace aside.
func TextuallyEqual(code string, accepted []byte, name string) bool {
	return strings.Join(strings.Fields(code), " ") == strings.Join(strings.Fields(ExportedText(accepted, name)), " ")
}
