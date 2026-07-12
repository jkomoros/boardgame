package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// TestInProgressionEmittedBindingsCoverDefaultTemplate is the
// moves-package half of footgun-batch F4's catalog pin
// (legal.TestDefaultTemplatePlaceholdersCoveredByEmittedBindings covers
// every predicate whose constructor lives in package legal): "inProgression"
// is the one framework predicate registered from THIS package (see
// catalog_framework.go's layering doc comment), so its default template's
// {placeholders} are pinned against its constructor's EmittedBindings here,
// where the constructor is reachable.
func TestInProgressionEmittedBindingsCoverDefaultTemplate(t *testing.T) {
	ctor := inProgressionConstructor()
	pred, err := ctor.Constructor(inProgressionSpec("Some Move"), nil, nil)
	if err != nil {
		t.Fatalf("constructing canonical inProgression predicate: %v", err)
	}
	if pred.EmittedBindings == nil {
		t.Fatal("inProgression declares no EmittedBindings metadata — every catalog predicate must (footgun-batch F4)")
	}

	table := legal.DefaultTemplates()
	for _, key := range pred.EmittedTemplates {
		bindings, ok := pred.EmittedBindings[key]
		if !ok {
			t.Errorf("inProgression emits template key %q but its EmittedBindings has no entry for it", key)
			continue
		}
		body, ok := table[key]
		if !ok {
			t.Errorf("inProgression's default template key %q is missing from legal.DefaultTemplates()", key)
			continue
		}
		emitted := make(map[string]bool, len(bindings))
		for _, b := range bindings {
			emitted[b] = true
		}
		for _, placeholder := range boardgame.LegalTemplatePlaceholders(body) {
			if !emitted[placeholder] {
				t.Errorf("default template %q (body %q) references placeholder {%s}, which inProgression does not emit with that key (emits: %v)", key, body, placeholder, bindings)
			}
		}
	}
}
