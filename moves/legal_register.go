package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

/*
This file is the injection point that hands core the universal
declarative-legality catalog it cannot import for itself (design spec §3
layering: core ← legal ← moves). See boardgame/legal_registry.go's package
doc for the full rationale.

Any game that uses declarative legality necessarily imports package moves
(WithPreconditions/WithoutPrecondition live here), and importing moves runs
this init() before NewGameManager, so by boot time core's default registry is
populated. The registration is of the UNIVERSAL catalog only — legal's
built-in predicates plus the one framework predicate ("inProgression") that
must be registered from package moves rather than package legal (see
catalog_framework.go). A game's own predicates and template overrides do NOT
come through here: they ride the per-delegate optional interfaces
(legal.ConstructorConfigurer / legal.TemplateConfigurer), which core overlays
on top of these defaults at boot.
*/

func init() {
	boardgame.RegisterDefaultLegalPredicateConstructors(legal.DefaultConstructors()...)
	boardgame.RegisterDefaultLegalPredicateConstructors(FrameworkConstructors()...)
	boardgame.RegisterDefaultLegalTemplates(legal.DefaultTemplates())
}
