package port

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseTSModule: the declaration scanner finds every exported name of
// the modules the brief quotes, in source order, and cuts a function at
// its signature.
func TestParseTSModule(t *testing.T) {
	root := moduleRoot(t)
	for _, c := range []struct {
		module string
		want   []string
	}{
		{"go/api.ts", []string{"StackRef", "Comp", "DynamicRef", "Owner", "GraphRef", "ILLEGAL_ENUM_VALUE", "player", "stockApply", "legalError"}},
		{"rules/state.ts", []string{"TimerHandle", "RulesState"}},
		{"rules/types.ts", []string{"MoveInstance", "PredicateSpec", "PredicateContext", "GamePredicate", "GamePredicateConstructor"}},
		{"legal/types.ts", []string{"LegalVerdict", "passVerdict", "failT", "unknownVerdict", "PropertyType", "LegalRead"}},
		{"go/num.ts", []string{"INT_OVERFLOW_PREFIX", "idiv", "imod", "imul"}},
	} {
		text, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(EngineDir), filepath.FromSlash(c.module)))
		if err != nil {
			t.Fatal(err)
		}
		mod := parseTSModule(string(text))
		exported := strings.Count("\n"+string(text), "\nexport ")
		if len(mod.Decls) != exported {
			var names []string
			for _, d := range mod.Decls {
				names = append(names, d.Name)
			}
			t.Errorf("%s: %d declarations found, %d lines start with export: %v", c.module, len(mod.Decls), exported, names)
		}
		for _, name := range c.want {
			d, ok := mod.find(name)
			if !ok {
				t.Errorf("%s: %s not found", c.module, name)
				continue
			}
			if d.Kind == "function" && strings.HasSuffix(strings.TrimSpace(d.Text), "{") {
				t.Errorf("%s: %s keeps its body brace: %q", c.module, name, d.Text)
			}
		}
	}
	owner, _ := parseTSModule(mustRead(t, filepath.Join(root, EngineDir, "go", "api.ts"))).find("Owner")
	if !strings.HasSuffix(owner.Text, "| { readonly move: MoveInstance };") {
		t.Errorf("a multi-line type alias was cut short: %q", owner.Text)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}
