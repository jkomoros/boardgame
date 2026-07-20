package legal

import "github.com/jkomoros/boardgame"

// TemplateConfigurer is implemented optionally by a game's GameDelegate to
// extend or override the catalog's default template table
// (DefaultTemplates) with the game's own template keys — e.g. a game
// registering its own predicate (design spec §8's checkers.spaceIsBlack)
// needs a template for the key its Evaluate emits, and any game may want
// friendlier text than a catalog default. This package never calls
// ConfigureLegalTemplates itself: it is consumed via a type-assertion on
// the delegate by a later task's NewGameManager wiring and server-side
// rendering (design spec §6), the same "optional interface, consumed by
// type-assertion, absence = defaults" pattern ConstructorConfigurer uses.
type TemplateConfigurer interface {
	// ConfigureLegalTemplates returns this game's template table: keys are
	// template keys (e.g. "reveal.no_cards_left"), values are the
	// human-readable template body RenderMessage substitutes {bindings}
	// into. A returned table need not be exhaustive — it is overlaid onto
	// DefaultTemplates() by the caller, not used standalone.
	ConfigureLegalTemplates() map[string]string
}

// legalAnyFailedTemplate is this package's own copy of core's unexported
// legalAnyFailedTemplate constant (boardgame/legal_error.go): the "any"
// compositor's implicit default Fail template key. package legal cannot
// reference core's unexported constant directly, so the literal is
// duplicated here. There is no compile-time check that the two copies
// stay equal (core is unexported and unreachable from this package); the
// value is a stable wire-format-adjacent string, not expected to change,
// and TestConformanceCorpus/TestDefaultTemplatesCoversCorpusFailCases would
// catch drift the moment any() becomes reachable through the corpus.
const legalAnyFailedTemplate = "legal.any_failed"

// defaultTemplates is the catalog's own default template table: one
// human-readable entry for every key in defaultTemplateKeys
// (legal/catalog_stack.go, which every catalog file in this package
// appends its own template keys to), plus legalAnyFailedTemplate for the
// core-level "any" compositor. See each Template* constant's doc comment
// (in the catalog_*.go file that declares it) for its bindings.
//
// Two entries deliberately render nothing but a raw "{detail}" binding
// pass-through: TemplateProposerTargetInvalid and
// TemplateProposerNotYourTurn. proposerIsCurrentPlayerConstructor
// (catalog_players.go) attaches the EXACT verbatim legacy string from
// moves/current_player.go ("The specified target player is not valid" /
// "it's not your turn") as each Fail Verdict's "detail" binding
// specifically so that a template of "{detail}" reproduces that string
// byte-for-byte through RenderMessage — see
// TestProposerTemplateRenderingParity below, which asserts exactly that.
var defaultTemplates = map[string]string{
	TemplatePropAtLeast:                   "requires at least {min}, but the current value is {value}",
	TemplatePropCompare:                   "requires the current value ({value}) to be {op} {n}",
	TemplatePlayerBool:                    "requires {prop} to be {want}",
	TemplatePlayerAlreadySubmitted:        "that player has already submitted",
	TemplatePlayerNotSubmitted:            "that player has not submitted",
	TemplatePlayerInactive:                "that player is inactive",
	TemplatePlayerActive:                  "that player is active",
	TemplateSeatNotFilled:                 "that seat is not filled",
	TemplateSeatNotClosed:                 "that seat is not closed",
	TemplatePlayerNotAdmin:                "that player is not a game administrator",
	TemplateComponentMissing:              "there is no component at index {index}",
	TemplateComponentPresentUnexpected:    "there is unexpectedly a component at index {index}",
	TemplateComponentMissingKey:           "there is no component at key {key}",
	TemplateNoComponentToMove:             "there is no component at index {index} to move",
	TemplateMayNotMoveTo:                  "{detail}",
	TemplateMayNotMoveAllTo:               "{detail}",
	TemplateMayNotSwapComponents:          "{detail}",
	TemplateAllActivePlayers:              "not every active player satisfies the required condition",
	TemplateProposerTargetInvalid:         "{detail}",
	TemplateProposerNotYourTurn:           "{detail}",
	TemplateProposerNotMovePlayer:         "the move's player does not match the proposer",
	TemplateNoCardHere:                    "there is no card at that index",
	TemplateAlreadyRevealed:               "that card has already been revealed",
	TemplateComponentPropNotCurrentPlayer: "the component's {prop} does not match the current player's {prop}",
	TemplateInPhase:                       "{detail}",
	TemplateInProgression:                 "{detail}",
	TemplateStackConstraints:              "{detail}",
	TemplateStackCount:                    "requires the stack's count ({count}) to be {op} {n}",
	TemplateStackEmpty:                    "requires the stack to be empty",
	TemplateStackNotEmpty:                 "requires the stack to not be empty",
	TemplatePropEquals:                    "requires the current value ({value}) to equal {want}",
	TemplatePropNotEquals:                 "requires the current value ({value}) to not equal {want}",
	legalAnyFailedTemplate:                "none of the required conditions were satisfied",
}

// DefaultTemplates returns the catalog's default template table: a fresh
// copy (safe for the caller to mutate/overlay) covering every key in
// defaultTemplateKeys plus the core-level "any" compositor's default key.
// A game extends or overrides this via its GameDelegate's optional
// TemplateConfigurer (ConfigureLegalTemplates); the standard overlay
// pattern a caller uses is:
//
//	table := legal.DefaultTemplates()
//	if tc, ok := delegate.(legal.TemplateConfigurer); ok {
//		for k, v := range tc.ConfigureLegalTemplates() {
//			table[k] = v
//		}
//	}
func DefaultTemplates() map[string]string {
	out := make(map[string]string, len(defaultTemplates))
	for k, v := range defaultTemplates {
		out[k] = v
	}
	return out
}

// Errorf lets IMPERATIVE code (LegalCustom bodies, a move's own Legal()
// override) return a structured legality failure: it builds a LegalFail
// Verdict carrying templateKey and bindings, and returns it as an error via
// boardgame.LegalVerdict.Error() — so the result is a concrete
// *boardgame.LegalError, retrievable from the returned error via
// errors.As(err, &target), exactly like a Verdict produced by declarative
// predicate evaluation. bindings may be nil for a template with no
// placeholders.
func Errorf(templateKey string, bindings map[string]boardgame.LegalBindingValue) error {
	v := Verdict{
		Outcome: boardgame.LegalFail,
		Message: &Message{
			Template: templateKey,
			Bindings: bindings,
		},
	}
	return v.Error()
}

// RenderMessage fills m's template (looked up in table by m.Template) with
// m.Bindings, substituting each "{name}" placeholder with the named
// binding's value; a placeholder with no corresponding binding renders as
// the bare placeholder name itself, never panics. A nil m renders as "". If
// table has no entry for m.Template, the bare template key is used as the
// body (so an unregistered key degrades to itself rather than to an empty
// string). This is a thin wrapper around boardgame.RenderLegalMessage,
// which is where the actual algorithm lives — core owns it too because
// boardgame.LegalError.Error() needs the identical rendering behavior and
// cannot import this package (see RenderLegalMessage's doc comment for the
// layering rationale).
func RenderMessage(m *Message, table map[string]string) string {
	return boardgame.RenderLegalMessage(m, table)
}
