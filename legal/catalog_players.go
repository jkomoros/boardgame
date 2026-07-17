package legal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

// Template keys for the quantifier/proposer predicates in this file.
const (
	// TemplateAllActivePlayers is the default Fail template key for
	// AllActivePlayers: fired when at least one active (per
	// behaviors.PlayerIsInactive) player's inner predicate does not hold.
	// No bindings.
	TemplateAllActivePlayers = "legal.all_active_players_failed"
	// TemplateProposerTargetInvalid is the default Fail template key for
	// ProposerIsCurrentPlayer when move.TargetPlayerIndex is a special
	// negative PlayerIndex (ObserverPlayerIndex/AdminPlayerIndex/
	// AnyPlayerIndex) rather than a concrete target player. Bindings:
	// "detail", the verbatim legacy error string from
	// moves/current_player.go:48/52 ("The specified target player is not
	// valid"). See ProposerIsCurrentPlayer's doc comment for why the legacy
	// string rides as a binding rather than baked into the template body.
	TemplateProposerTargetInvalid = "legal.proposer_target_invalid"
	// TemplateProposerNotYourTurn is the default Fail template key for
	// ProposerIsCurrentPlayer when move.TargetPlayerIndex doesn't match the
	// game's current player, or doesn't match the proposer. Bindings:
	// "detail", the verbatim legacy error string from
	// moves/current_player.go:56/60 ("it's not your turn").
	TemplateProposerNotYourTurn = "legal.proposer_not_your_turn"
	// TemplateProposerNotMovePlayer is the default Fail template key for
	// ProposerIsPlayerFromMove. No bindings.
	TemplateProposerNotMovePlayer = "legal.proposer_not_move_player"
)

// Any returns a Spec for the "any" compositor: Passes if at least one of
// subs passes (Kleene semantics: Pass beats Unknown beats Fail — see the
// design spec §6 and boardgame's evalLegalAnyKleene). This builder is
// deliberately dumb, per the design spec's "builders stay dumb" rule: it
// does not itself enforce the >= 2 subs requirement or the depth-1
// (no-nested-any) rule. Both are enforced at Spec RESOLUTION time by core
// (boardgame.resolveLegalSpecs / resolveLegalAnySpec), which is the only
// place "any" is actually interpreted — see legal_predicate.go. Any is not
// itself a registered PredicateConstructor (DefaultConstructors() does not,
// and must not, include an entry named "any": core intercepts the name
// directly so it can never be shadowed).
func Any(subs ...Spec) Spec {
	return Spec{Name: "any", Sub: subs}
}

// AllActivePlayers returns a Spec for the "allActivePlayers" quantifier:
// Passes if inner holds for every active player (per
// behaviors.PlayerIsInactive — inactive players are skipped entirely, never
// counted toward pass or fail). inner is stored as the sole element of
// Spec.Sub.
//
// v1 restriction (plan-mandated, enforced at construction — see
// allActivePlayersConstructor): inner must be one of playerBool,
// player-path propAtLeast, player-path propCompare, or an "any" composing
// two or more of those three leaf kinds (matching the design spec §8 acid
// test: AllActivePlayers(Any(PlayerBool("Eliminated"), PlayerBool("Stood")))).
// Any other inner Name — including a nested "any" beneath the top-level
// "any" — is a boot error naming the unsupported predicate.
func AllActivePlayers(inner Spec) Spec {
	return Spec{Name: "allActivePlayers", Sub: []Spec{inner}}
}

// ProposerIsCurrentPlayer returns a Spec for the "proposerIsCurrentPlayer"
// predicate: replicates moves/current_player.go:37-65's TargetPlayerIndex
// checks exactly (the part of CurrentPlayer.Legal beyond its
// Default.Legal() super-call, which is contributed separately by
// Default.ContributedPreconditions per spec §2). See
// proposerIsCurrentPlayerConstructor for the byte-for-byte mapping.
func ProposerIsCurrentPlayer() Spec {
	return Spec{Name: "proposerIsCurrentPlayer"}
}

// ProposerIsPlayerFromMove passes when field is a concrete PlayerIndex move
// property naming the proposer. AdminPlayerIndex bypasses this actor check;
// ObserverPlayerIndex and AnyPlayerIndex are never valid proposing actors.
// This is useful when a move acts for an explicitly selected player without
// requiring that player to be the game's current player.
func ProposerIsPlayerFromMove(field string) Spec {
	return Spec{
		Name:        "proposerIsPlayerFromMove",
		Args:        []string{field},
		AdminPolicy: boardgame.LegalAdminBypass,
	}
}

// playerPathProp validates that path is a "player.X" path and returns X. It
// is used by AllActivePlayers' inner-spec restriction: propAtLeast/
// propCompare are general-purpose (any path kind), but as AllActivePlayers'
// inner they must reference the QUANTIFIED player, i.e. be spelled
// "player.X" (never "game.X" or "move.X" — those wouldn't vary per player
// and don't belong inside a per-player quantifier).
func playerPathProp(path string) (string, error) {
	const prefix = "player."
	if !strings.HasPrefix(path, prefix) {
		return "", fmt.Errorf("path %q must be a player.* path (AllActivePlayers' inner propAtLeast/propCompare only accept player-path args)", path)
	}
	prop := strings.TrimPrefix(path, prefix)
	if prop == "" {
		return "", fmt.Errorf("path %q is missing a property name after %q", path, prefix)
	}
	return prop, nil
}

// allActivePlayersLeaf is what buildAllActivePlayersLeaf/buildAllActivePlayersInner
// compile an inner leaf Spec (playerBool/propAtLeast/propCompare) down to: a
// function that evaluates the leaf directly against one player's
// PropertyReader (bypassing ctx.ResolvePath entirely — see the doc comment
// on allActivePlayersConstructor for why), plus the "players[*].X" Reads it
// implies.
type allActivePlayersLeaf struct {
	eval          func(reader boardgame.PropertyReader) Outcome
	reads         []Read
	requiredTypes map[PropPath]boardgame.PropertyType
}

// buildAllActivePlayersLeaf compiles one of AllActivePlayers' three
// supported leaf inner-spec kinds (playerBool, propAtLeast, propCompare)
// into an allActivePlayersLeaf. Any other Name — including "any", which is
// only accepted ONE level up, by buildAllActivePlayersInner, never as a
// leaf itself — is a boot error. This is what makes the depth-1 restriction
// on the inner "any" automatic: a nested "any" beneath the top-level "any"
// is passed to this function (not back to buildAllActivePlayersInner), and
// falls through to the default case.
func buildAllActivePlayersLeaf(s Spec) (allActivePlayersLeaf, error) {
	switch s.Name {
	case "playerBool":
		// Accepts both playerBool's 1-arg (prop, want=true) and 2-arg
		// (prop, want) spellings — spec §4's negation leaves require
		// AllActivePlayers(PlayerBoolIs("DoneWithPhase", false)) to
		// construct and evaluate per-player. playerBoolWant
		// (catalog_compare.go) is the same parser playerBoolConstructor
		// itself uses, so this inner grammar accepts EXACTLY the same 2nd-
		// arg syntax as the top-level "playerBool" predicate — no
		// independently-drifting copy.
		if len(s.Args) != 1 && len(s.Args) != 2 {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner playerBool requires 1 or 2 args (prop, optional want), got %d", len(s.Args))
		}
		prop := s.Args[0]
		want := true
		if len(s.Args) == 2 {
			w, err := playerBoolWant(s.Args[1])
			if err != nil {
				return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner playerBool: arg 2 (want) %s", err)
			}
			want = w
		}
		read := Read{Path: PropPath("players[*]." + prop), Facet: boardgame.LegalFacetValues}
		return allActivePlayersLeaf{
			reads:         []Read{read},
			requiredTypes: map[PropPath]boardgame.PropertyType{read.Path: boardgame.TypeBool},
			eval: func(reader boardgame.PropertyReader) Outcome {
				if reader == nil {
					return Unknown
				}
				v, err := reader.BoolProp(prop)
				if err != nil {
					return Unknown
				}
				if v == want {
					return Pass
				}
				return Fail
			},
		}, nil
	case "propAtLeast":
		if len(s.Args) != 2 {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propAtLeast requires 2 args (path, n), got %d", len(s.Args))
		}
		prop, err := playerPathProp(s.Args[0])
		if err != nil {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propAtLeast: %w", err)
		}
		n, err := strconv.Atoi(s.Args[1])
		if err != nil {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propAtLeast: arg 2 (n) must be an integer: %w", err)
		}
		read := Read{Path: PropPath("players[*]." + prop), Facet: boardgame.LegalFacetValues}
		return allActivePlayersLeaf{
			reads:         []Read{read},
			requiredTypes: map[PropPath]boardgame.PropertyType{read.Path: boardgame.TypeInt},
			eval: func(reader boardgame.PropertyReader) Outcome {
				if reader == nil {
					return Unknown
				}
				v, err := reader.IntProp(prop)
				if err != nil {
					return Unknown
				}
				if v >= n {
					return Pass
				}
				return Fail
			},
		}, nil
	case "propCompare":
		if len(s.Args) != 3 {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propCompare requires 3 args (path, op, n), got %d", len(s.Args))
		}
		prop, err := playerPathProp(s.Args[0])
		if err != nil {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propCompare: %w", err)
		}
		op := s.Args[1]
		cmp, ok := legalCompareOps[op]
		if !ok {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propCompare: unknown op %q (expected one of ==, !=, <, <=, >, >=)", op)
		}
		n, err := strconv.Atoi(s.Args[2])
		if err != nil {
			return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: inner propCompare: arg 3 (n) must be an integer: %w", err)
		}
		read := Read{Path: PropPath("players[*]." + prop), Facet: boardgame.LegalFacetValues}
		return allActivePlayersLeaf{
			reads:         []Read{read},
			requiredTypes: map[PropPath]boardgame.PropertyType{read.Path: boardgame.TypeInt},
			eval: func(reader boardgame.PropertyReader) Outcome {
				if reader == nil {
					return Unknown
				}
				v, err := reader.IntProp(prop)
				if err != nil {
					return Unknown
				}
				if cmp(v, n) {
					return Pass
				}
				return Fail
			},
		}, nil
	default:
		return allActivePlayersLeaf{}, fmt.Errorf("legal: allActivePlayers: unsupported inner predicate %q (v1 supports playerBool, propAtLeast, and propCompare on player.* paths, optionally composed with any())", s.Name)
	}
}

// buildAllActivePlayersInner compiles AllActivePlayers' inner Spec (either a
// single leaf, or an "any" composing two or more leaves) into a per-player
// evaluator and its declared Reads. See buildAllActivePlayersLeaf for the
// leaf kinds and the depth-1 enforcement.
func buildAllActivePlayersInner(inner Spec) (func(reader boardgame.PropertyReader) Outcome, []Read, map[PropPath]boardgame.PropertyType, error) {
	if inner.Name != legalAnyName {
		leaf, err := buildAllActivePlayersLeaf(inner)
		if err != nil {
			return nil, nil, nil, err
		}
		return leaf.eval, leaf.reads, leaf.requiredTypes, nil
	}

	if len(inner.Sub) < 2 {
		return nil, nil, nil, fmt.Errorf("legal: allActivePlayers: inner %q compositor requires at least 2 sub-specs, got %d", legalAnyName, len(inner.Sub))
	}

	leaves := make([]allActivePlayersLeaf, 0, len(inner.Sub))
	seen := make(map[Read]bool)
	var reads []Read
	requiredTypes := make(map[PropPath]boardgame.PropertyType)
	for _, sub := range inner.Sub {
		leaf, err := buildAllActivePlayersLeaf(sub)
		if err != nil {
			return nil, nil, nil, err
		}
		leaves = append(leaves, leaf)
		for _, r := range leaf.reads {
			if !seen[r] {
				seen[r] = true
				reads = append(reads, r)
			}
		}
		for path, expected := range leaf.requiredTypes {
			if prior, exists := requiredTypes[path]; exists && prior != expected {
				return nil, nil, nil, fmt.Errorf("legal: allActivePlayers: inner predicates require conflicting types for %q (%v and %v)", path, prior, expected)
			}
			requiredTypes[path] = expected
		}
	}

	eval := func(reader boardgame.PropertyReader) Outcome {
		sawUnknown := false
		for _, leaf := range leaves {
			switch leaf.eval(reader) {
			case Pass:
				return Pass
			case Unknown:
				sawUnknown = true
			}
		}
		if sawUnknown {
			return Unknown
		}
		return Fail
	}
	return eval, reads, requiredTypes, nil
}

// legalAnyName is this package's copy of core's reserved "any" compositor
// name, used only to recognize AllActivePlayers' inner "any" (a distinct,
// hand-interpreted mini-composition — see allActivePlayersConstructor's doc
// comment — not core's generic any-compositor resolution). Kept as its own
// constant (rather than importing boardgame's unexported
// legalAnyCompositorName, which isn't exported anyway) so the literal
// "any" appears exactly once in this file.
const legalAnyName = "any"

// allActivePlayersConstructor returns the registry entry for
// "allActivePlayers".
//
// Per-player resolution mechanics (why this constructor never calls the
// resolve callback it's handed): ctx.ResolvePath's "player.X" grammar
// always resolves against state.ImmutableCurrentPlayer() (see
// legal_path.go's pathPlayer case) — there is no way to ask it to resolve
// "player.X" against a DIFFERENT, explicitly-named player. The design
// brief's own considered alternative (a playerOverride field on
// LegalContext) was rejected as a new core field mid-plan. So instead of
// resolving inner into a generic *Predicate and calling its Evaluate with
// some doctored Context, this constructor hand-compiles inner's restricted
// grammar (buildAllActivePlayersInner) into a function that reads a
// SPECIFIC player's boardgame.PropertyReader directly — bypassing
// ctx.ResolvePath and the "current player" concept entirely — and the
// top-level Evaluate below calls that function once per active player's
// own Reader(). This is also why inner is restricted to three leaf kinds in
// v1 (plus a one-level "any" over them): each kind's evaluation is
// hand-written here against a raw PropertyReader, not derived generically
// from an arbitrary resolved Predicate.
func allActivePlayersConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "allActivePlayers",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Sub) != 1 {
				return nil, fmt.Errorf("legal: allActivePlayers requires exactly 1 inner spec (in Sub), got %d", len(spec.Sub))
			}

			innerEval, reads, requiredTypes, err := buildAllActivePlayersInner(spec.Sub[0])
			if err != nil {
				return nil, err
			}

			template := spec.Message
			if template == "" {
				template = TemplateAllActivePlayers
			}

			return &Predicate{
				Name: "allActivePlayers",
				// The wire entry does not yet carry the quantified inner spec.
				ClientEvaluable:   false,
				Reads:             reads,
				RequiredReadTypes: requiredTypes,
				Cost:              boardgame.LegalCostModerate,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: nil},
				Evaluate: func(ctx Context) Verdict {
					if ctx.State == nil {
						return UnknownVerdict("legal: allActivePlayers: state was nil")
					}
					sawUnknown := false
					for _, p := range ctx.State.ImmutablePlayerStates() {
						if p == nil {
							continue
						}
						if behaviors.PlayerIsInactive(p) {
							continue
						}
						switch innerEval(p.Reader()) {
						case Fail:
							return FailT(template)
						case Unknown:
							sawUnknown = true
						}
					}
					if sawUnknown {
						return UnknownVerdict("legal: allActivePlayers: at least one active player's inner predicate was unknown")
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}

// proposerIsCurrentPlayerConstructor returns the registry entry for
// "proposerIsCurrentPlayer".
//
// This replicates moves/current_player.go:37-65's TargetPlayerIndex checks
// exactly — same order, same conditions (via boardgame.PlayerIndex's own
// EnsureValid/Valid/Equivalent, not reimplemented), same two distinct
// legacy error strings. What it deliberately does NOT replicate is the
// leading "if err := c.Default.Legal(state, proposer); err != nil {...}"
// super-call: per the design spec §2, Default's checks (phase/progression/
// stack constraints) are contributed separately by
// Default.ContributedPreconditions, and CurrentPlayer's chain is that list
// PLUS this predicate — so duplicating Default's checks here would run them
// twice in the assembled plan.
//
// String-parity approach (see the design spec §6 and this file's Template*
// constants): rather than waiting on Task 6's template-rendering pass, each
// Fail Verdict carries a "detail" binding holding the EXACT legacy string
// verbatim ("The specified target player is not valid" / "it's not your
// turn"). This makes the string-parity property directly assertable today
// (Verdict.Message.Bindings["detail"].S) without depending on
// legal.DefaultTemplates existing yet. It also hands Task 6 a trivial path
// to byte-for-byte parity: default template body "{{.detail}}" (a
// pass-through interpolation) reproduces the legacy string exactly for any
// renderer that supports the binding, while still leaving room for a
// friendlier, non-verbatim template to be swapped in later without
// touching this file.
func proposerIsCurrentPlayerConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "proposerIsCurrentPlayer",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 0 {
				return nil, fmt.Errorf("legal: proposerIsCurrentPlayer takes no args, got %d", len(spec.Args))
			}

			targetInvalidTemplate := spec.Message
			if targetInvalidTemplate == "" {
				targetInvalidTemplate = TemplateProposerTargetInvalid
			}
			notYourTurnTemplate := spec.Message
			if notYourTurnTemplate == "" {
				notYourTurnTemplate = TemplateProposerNotYourTurn
			}

			return &Predicate{
				Name:            "proposerIsCurrentPlayer",
				ClientEvaluable: true,
				Reads: []Read{
					// FIELD-DEPENDENT (spec §4): this Read is the reason
					// proposerIsCurrentPlayer belongs in a plan's
					// fieldDependent bucket, not fieldIndependent.
					{Path: PropPath("move.TargetPlayerIndex"), Facet: boardgame.LegalFacetValues},
					// The game current-player read, declared by convention
					// on "game.CurrentPlayer" — the property
					// base.GameDelegate.CurrentPlayerIndex reads by
					// default, and the one behaviors.CurrentPlayerBehavior
					// provides (every in-repo game using moves.CurrentPlayer
					// today embeds it). Evaluate below never reads this
					// path directly: it calls ctx.State.CurrentPlayerIndex(),
					// which is delegate-correct even for a delegate that
					// overrides CurrentPlayerIndex non-conventionally. This
					// Read is declared for client-evaluability/documentation
					// purposes to reflect what the value depends on in the
					// common case; a delegate that overrides
					// CurrentPlayerIndex without backing it by a
					// "CurrentPlayer" game-state property would fail
					// boot-time path validation for this specific declared
					// Read — a known v1 limitation of this convention-based
					// declaration, not exercised by any in-repo game today.
					{Path: PropPath("game.CurrentPlayer"), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath("move.TargetPlayerIndex"): boardgame.TypePlayerIndex,
					PropPath("game.CurrentPlayer"):     boardgame.TypePlayerIndex,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{targetInvalidTemplate, notYourTurnTemplate},
				// Both branches emit exactly {"detail"}, so a Spec.Message
				// override collapsing the two keys onto one still guarantees
				// "detail" (the intersection is the same set).
				EmittedBindings: map[string][]string{
					targetInvalidTemplate: {"detail"},
					notYourTurnTemplate:   {"detail"},
				},
				Evaluate: func(ctx Context) Verdict {
					if ctx.State == nil {
						return UnknownVerdict("legal: proposerIsCurrentPlayer: state was nil")
					}

					rawTarget, err := resolvePlayerIndexPath("move.TargetPlayerIndex", ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}

					currentPlayer := ctx.State.CurrentPlayerIndex()
					targetPlayerIndex := rawTarget.EnsureValid(ctx.State)

					if !targetPlayerIndex.Valid(ctx.State) {
						return FailT(targetInvalidTemplate, map[string]BindingValue{
							"detail": String("The specified target player is not valid"),
						})
					}

					if targetPlayerIndex < 0 {
						return FailT(targetInvalidTemplate, map[string]BindingValue{
							"detail": String("The specified target player is not valid"),
						})
					}

					if !targetPlayerIndex.Equivalent(currentPlayer) {
						return FailT(notYourTurnTemplate, map[string]BindingValue{
							"detail": String("it's not your turn"),
						})
					}

					if !targetPlayerIndex.Equivalent(ctx.ProposerPlayerIndex) {
						return FailT(notYourTurnTemplate, map[string]BindingValue{
							"detail": String("it's not your turn"),
						})
					}

					return PassVerdict()
				},
			}, nil
		},
	}
}

func proposerIsPlayerFromMoveConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "proposerIsPlayerFromMove",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 1 || spec.Args[0] == "" {
				return nil, fmt.Errorf("legal: proposerIsPlayerFromMove requires 1 non-empty arg (move PlayerIndex field), got %v", spec.Args)
			}
			path := "move." + spec.Args[0]
			template := spec.Message
			if template == "" {
				template = TemplateProposerNotMovePlayer
			}
			return &Predicate{
				Name:              "proposerIsPlayerFromMove",
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypePlayerIndex},
				Cost:              boardgame.LegalCostTrivial,
				ClientEvaluable:   true,
				UsesProposer:      true,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: nil},
				Evaluate: func(ctx Context) Verdict {
					if ctx.State == nil {
						return UnknownVerdict("legal: proposerIsPlayerFromMove: state was nil")
					}
					target, err := resolvePlayerIndexPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					target = target.EnsureValid(ctx.State)
					if target < 0 || !target.Valid(ctx.State) || ctx.ProposerPlayerIndex < 0 || !target.Equivalent(ctx.ProposerPlayerIndex) {
						return FailT(template)
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}
