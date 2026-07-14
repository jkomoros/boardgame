package boardgame

import "sync"

/*
This file is the layering bridge for declarative-legality boot assembly.

NewGameManager must resolve a move type's declared preconditions against a
registry of LegalPredicateConstructors, and render their failure templates
against a default template table. Both the default constructors
(legal.DefaultConstructors + the "inProgression" constructor that must live
in package moves — see moves/catalog_framework.go) and the default templates
(legal.DefaultTemplates) are defined in packages this package cannot import:
package legal imports package boardgame, and package moves imports both, so
the dependency arrows point core ← legal ← moves and never the reverse (see
the design spec §3's layering diagram).

Rather than a new GameDelegate method (which the spec §1 explicitly rejects —
it would be a compile break for every existing delegate) or having package
base provide the defaults (base cannot import package moves either: moves
imports base, so that would be an import cycle), package moves injects the
universal defaults into this process-wide registry from an init() — the same
decoupling idiom Go's image and database/sql packages use for format/driver
registration (see moves/legal_register.go). The registration is of the
UNIVERSAL catalog only (identical for every game in the process); a game's
own predicates (checkers.spaceIsBlack, design spec §8) and template overrides
ride the per-delegate optional interfaces (legal.ConstructorConfigurer /
legal.TemplateConfigurer), consumed by type-assertion at boot and overlaid on
top of these defaults — so nothing game-specific ever lands in this global.
*/

var legalDefaultRegistryMu sync.RWMutex
var legalDefaultConstructors = map[string]*LegalPredicateConstructor{}
var legalDefaultTemplates = map[string]string{}

// RegisterDefaultLegalPredicateConstructors adds ctors to the process-wide
// default LegalPredicateConstructor registry NewGameManager consults when a
// delegate does not override it via the optional
// ConfigurePredicateConstructors interface (design spec §1). It is called
// once, from package moves' init(), with the merged universal catalog
// (legal.DefaultConstructors() plus moves.FrameworkConstructors()). A later
// registration of the same Name wins (last write). A nil or unnamed
// constructor is ignored. This is process-global by design: the universal
// catalog is identical for every game; game-specific predicates never come
// through here (see this file's package doc).
func RegisterDefaultLegalPredicateConstructors(ctors ...*LegalPredicateConstructor) {
	legalDefaultRegistryMu.Lock()
	defer legalDefaultRegistryMu.Unlock()
	for _, ctor := range ctors {
		if ctor == nil || ctor.Name == "" {
			continue
		}
		legalDefaultConstructors[ctor.Name] = ctor
	}
}

// RegisterDefaultLegalTemplates adds the given template-key→body entries to
// the process-wide default template table NewGameManager merges under a
// delegate's own ConfigureLegalTemplates (design spec §6). Called once from
// package moves' init() with legal.DefaultTemplates(). A later registration
// of the same key wins.
func RegisterDefaultLegalTemplates(templates map[string]string) {
	legalDefaultRegistryMu.Lock()
	defer legalDefaultRegistryMu.Unlock()
	for k, v := range templates {
		legalDefaultTemplates[k] = v
	}
}

// legalRegistrySnapshot returns fresh copies of the registered default
// constructor registry and template table, safe for the caller to overlay a
// delegate's own entries onto without mutating the process-global originals.
func legalRegistrySnapshot() (map[string]*LegalPredicateConstructor, map[string]string) {
	legalDefaultRegistryMu.RLock()
	defer legalDefaultRegistryMu.RUnlock()

	registry := make(map[string]*LegalPredicateConstructor, len(legalDefaultConstructors))
	for k, v := range legalDefaultConstructors {
		registry[k] = v
	}
	templates := make(map[string]string, len(legalDefaultTemplates))
	for k, v := range legalDefaultTemplates {
		templates[k] = v
	}
	return registry, templates
}
