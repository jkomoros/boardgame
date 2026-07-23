package legal

import (
	"testing"

	"github.com/jkomoros/boardgame"
)

// TestDefaultTemplatePlaceholdersCoveredByEmittedBindings pins footgun-batch
// F4's catalog-side invariant: every {placeholder} in every default template
// body (DefaultTemplates()) must be a binding its canonical catalog
// predicate declares it emits with that key (Predicate.EmittedBindings) —
// otherwise the default rendering would show the bare placeholder name
// mid-game. It constructs every catalog predicate through its REAL
// constructor (DefaultConstructors(), the same path NewGameManager
// resolution takes) with a representative Message-free spec, so the pinned
// metadata is what the constructor actually populates, not a re-declared
// copy that could drift.
//
// This is an internal (package legal) test: it consults defaultTemplateKeys
// and legalAnyFailedTemplate directly to prove COMPLETENESS — every default
// key is covered either by a constructed catalog predicate, or by one of the
// two keys whose predicate lives outside this package ("inProgression" in
// package moves, whose constructor-side pin is
// moves.TestInProgressionEmittedBindingsCoverDefaultTemplate; the "any"
// compositor in core, which never attaches bindings, checked directly
// below).
func TestDefaultTemplatePlaceholdersCoveredByEmittedBindings(t *testing.T) {
	// One representative, Message-free spec per catalog constructor. Args
	// only need to satisfy construction-time shape checks (paths are not
	// resolved at construction; the chest is nil so propEquals' typo guard
	// is skipped, as its own doc comment allows for test harnesses).
	canonicalSpecs := map[string]Spec{
		"propAtLeast":                      PropAtLeast("game.Counter", 1),
		"propCompare":                      PropCompare("game.Counter", "==", 1),
		"playerBool":                       PlayerBool("SomeBool"),
		"playerBoolAt":                     PlayerBoolAt(CurrentPlayer(), "SomeBool", true),
		"propEquals":                       PropEquals("game.Counter", "1"),
		"propNotEquals":                    PropNotEquals("game.Counter", "1"),
		"componentPresentAt":               ComponentPresentAt("game.SomeStack", "move.SomeIndex"),
		"componentAbsentAt":                ComponentAbsentAt("game.SomeStack", "move.SomeIndex"),
		"componentPresentAtKey":            ComponentPresentAtKey("game.SomeStack", "move.SomeKey"),
		"mayMoveTo":                        MayMoveTo("game.SrcStack", "game.DstStack", "move.SomeIndex"),
		"mayMoveToSlot":                    MayMoveToSlot("game.SrcStack", "game.DstStack", "move.SomeIndex", "move.SomeSlot"),
		"mayMoveAllTo":                     MayMoveAllTo("game.SrcStack", "game.DstStack"),
		"mayMoveCountTo":                   MayMoveCountTo("game.SrcStack", "game.DstStack", "move.SomeCount"),
		"mayMoveFixedCountTo":              MayMoveFixedCountTo("game.SrcStack", "game.DstStack", 2),
		"maySwapComponents":                MaySwapComponents("game.SrcStack", "move.SomeIndex", "move.OtherIndex"),
		"maySwapComponentsByKey":           MaySwapComponentsByKey("game.SrcStack", "move.SomeKey", "move.OtherKey"),
		"allActivePlayers":                 AllActivePlayers(PlayerBool("SomeBool")),
		"proposerIsCurrentPlayer":          ProposerIsCurrentPlayer(),
		"proposerIsPlayerFromMove":         ProposerIsPlayerFromMove("TargetPlayerIndex"),
		"revealableCardAt":                 RevealableCardAt("game.HiddenStack", "game.VisibleStack", "move.SomeIndex"),
		"componentPropEqualsCurrentPlayer": ComponentPropEqualsCurrentPlayer("game.SomeStack", "move.SomeKey", "Color"),
		"inPhase":                          InPhase(),
		"stackConstraints":                 StackConstraints("SrcStack", "DstStack"),
		"stackCount":                       StackCount("game.SomeStack", "==", 1),
		"stackEmpty":                       StackEmpty("game.SomeStack"),
		"stackNotEmpty":                    StackNotEmpty("game.SomeStack"),
	}

	table := DefaultTemplates()
	covered := make(map[string]bool)

	for _, ctor := range DefaultConstructors() {
		spec, ok := canonicalSpecs[ctor.Name]
		if !ok {
			t.Fatalf("no canonical spec for catalog constructor %q — add one to this test so its default template's placeholders stay pinned to its emitted bindings", ctor.Name)
		}
		pred, err := ctor.Constructor(spec, nil, nil)
		if err != nil {
			t.Fatalf("constructing canonical %q predicate: %v", ctor.Name, err)
		}
		for _, read := range pred.Reads {
			_, fixed := pred.RequiredReadTypes[read.Path]
			allowed, polymorphic := pred.AllowedReadTypes[read.Path]
			if !fixed && (!polymorphic || len(allowed) == 0) {
				t.Errorf("catalog predicate %q read %q has no type contract", ctor.Name, read.Path)
			}
		}
		if pred.EmittedBindings == nil {
			t.Errorf("catalog predicate %q declares no EmittedBindings metadata — every catalog predicate must (footgun-batch F4)", ctor.Name)
			continue
		}
		for _, key := range pred.EmittedTemplates {
			covered[key] = true
			bindings, ok := pred.EmittedBindings[key]
			if !ok {
				t.Errorf("catalog predicate %q emits template key %q but its EmittedBindings has no entry for it", ctor.Name, key)
				continue
			}
			body, ok := table[key]
			if !ok {
				// DefaultTemplates coverage is TestDefaultTemplatesCovers*'s
				// job; don't double-report here.
				continue
			}
			emitted := make(map[string]bool, len(bindings))
			for _, b := range bindings {
				emitted[b] = true
			}
			for _, placeholder := range boardgame.LegalTemplatePlaceholders(body) {
				if !emitted[placeholder] {
					t.Errorf("default template %q (body %q) references placeholder {%s}, which catalog predicate %q does not emit with that key (emits: %v)", key, body, placeholder, ctor.Name, bindings)
				}
			}
		}
	}

	// "inProgression"'s constructor lives in package moves (this package
	// cannot import it — moves imports legal); only its default template
	// BODY lives here. The constructor-side subset pin is
	// moves.TestInProgressionEmittedBindingsCoverDefaultTemplate; mark the
	// key covered so the completeness sweep below stays honest about where
	// its proof lives.
	covered[TemplateInProgression] = true
	// These behavior-aware templates are selected by semantic wrappers around
	// playerBoolAt. They intentionally carry no placeholders.
	for _, key := range []string{
		TemplatePlayerAlreadySubmitted,
		TemplatePlayerNotSubmitted,
		TemplatePlayerInactive,
		TemplatePlayerActive,
		TemplateSeatNotFilled,
		TemplateSeatNotClosed,
		TemplatePlayerNotAdmin,
	} {
		if placeholders := boardgame.LegalTemplatePlaceholders(table[key]); len(placeholders) != 0 {
			t.Errorf("behavior template %q unexpectedly references placeholders %v", key, placeholders)
		}
		covered[key] = true
	}

	// The core "any" compositor (resolved by boardgame.resolveLegalSpecs,
	// never through this registry) attaches NO bindings to its Fail/Unknown
	// Message, so its default template body may not reference any
	// placeholder at all.
	if placeholders := boardgame.LegalTemplatePlaceholders(table[legalAnyFailedTemplate]); len(placeholders) != 0 {
		t.Errorf("default template %q references placeholders %v, but the \"any\" compositor never emits bindings", legalAnyFailedTemplate, placeholders)
	}
	covered[legalAnyFailedTemplate] = true

	// Completeness: every default template key must have been pinned by one
	// of the constructed predicates (or the two out-of-package cases above),
	// so a future catalog predicate cannot add a default key that escapes
	// this test.
	for _, key := range defaultTemplateKeys {
		if !covered[key] {
			t.Errorf("default template key %q was not covered by any constructed catalog predicate's EmittedTemplates — extend this test's canonical specs", key)
		}
	}
}
