/*
Package checkers is a simple example of the classic checkers game. It exercises
a grid-like board.
*/
package checkers

// NOTE: legal/conformance_test.go builds fixture states from this game; renaming state properties breaks that suite.

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {

	//If our package name and delegate.Name() don't match, NewGameManager will
	//fail with an error. Given they have to be the same, we might as well
	//just ensure they are actually the same, via a one-time reflection.

	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
}

func (g *gameDelegate) Description() string {
	return "Checkers is the classic game on a grid where players compete to capture opponents' pieces."
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 2
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.Add(
			auto.MustConfig(
				new(moves.SeatPlayer),
			),
		),
		moves.AddOrderedForPhase(phaseSetup,
			auto.MustConfig(
				new(movePlaceToken),
				moves.WithHelpText("Places one token at a time on the board."),
				// Declarative migration (design spec §8, PARTIAL): Legal() is
				// deleted (see moves.go). Only the first of the three original
				// gates — "No more components to place" (UnusedTokens empty) —
				// is declarative, as this StackNotEmpty precondition;
				// FixUpMulti's phase + progression checks are contributed
				// base-first ahead of it. The remaining two gates
				// (MayMoveToSlot against a fixed source index, then
				// spaceIsBlack) stay imperative, in their original order, in
				// the move's LegalCustom. LegalCustom itself opts the move in;
				// this spec exists because it is a natural client-visible gate.
				moves.WithLegalPreconditions(
					legal.StackNotEmpty("game.UnusedTokens").
						WithMessage("checkers.no_more_components"),
				),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithPhaseToStart(phasePlaying, phaseEnum),
			),
		),
		moves.AddForPhase(phasePlaying,
			auto.MustConfig(
				new(moveCrownToken),
				moves.WithHelpText("Crowns tokens that make it to the other end of the board."),
				moves.WithSourceProperty("Spaces"),
			),
			auto.MustConfig(
				new(moves.FinishTurn),
			),
			auto.MustConfig(
				new(moveMoveToken),
				moves.WithHelpText("Moves a token from one place to another"),
				// Declarative migration (design spec §8's checkers acid
				// test): Legal() is deleted (see moves.go); this plan
				// replaces it exactly, in the same order the old imperative
				// chain ran (the phase check and CurrentPlayer's proposer
				// check are contributed automatically ahead of these four,
				// then these four in declaration order, then LegalCustom's
				// capture-graph-walk residue — see moves.go's comment for
				// the full mapping).
				moves.WithLegalPreconditions(
					legal.MaySwapComponentsByKey("game.Spaces", "move.TokenIndexToMove", "move.SpaceIndex"),
					legal.ComponentPresentAtKey("game.Spaces", "move.TokenIndexToMove").
						WithMessage("checkers.no_token_there"),
					legal.ComponentPropEqualsCurrentPlayer("game.Spaces", "move.TokenIndexToMove", "Color").
						WithMessage("checkers.not_your_token"),
					legal.Spec{Name: "checkers.spaceIsBlack", Args: []string{"move.SpaceIndex"}},
				),
			),
		),
	)
}

// ConfigurePredicateConstructors registers checkers' one game-specific
// predicate, "checkers.spaceIsBlack" (design spec §1/§8): the FIRST real use
// of the game-registered predicate extension path (legal.ConstructorConfigurer),
// consumed via type-assertion on this delegate at NewGameManager, the same
// way legal.TemplateConfigurer below is. Delegate constructors overlay the
// built-in catalog, so this method returns only the game-specific addition.
func (g *gameDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {
	return []*legal.PredicateConstructor{{
		Name: "checkers.spaceIsBlack",
		Constructor: func(spec legal.Spec, _ *boardgame.ComponentChest,
			_ func(legal.Spec) (*legal.Predicate, error)) (*legal.Predicate, error) {
			if len(spec.Args) != 1 {
				return nil, fmt.Errorf("checkers.spaceIsBlack: requires exactly 1 arg (the space-index field), got %d", len(spec.Args))
			}
			field := spec.Args[0]

			template := spec.Message
			if template == "" {
				template = "checkers.black_spaces_only"
			}

			return &legal.Predicate{
				Name: "checkers.spaceIsBlack",
				Args: spec.Args,
				Reads: []legal.Read{
					{Path: legal.PropPath(field), Facet: boardgame.LegalFacetValues},
				},
				Cost:             boardgame.LegalCostTrivial,
				EmittedTemplates: []string{template},
				// Recommended (not required) game-registered metadata: the
				// FailT below attaches no bindings, so the template body —
				// default or WithMessage retarget — may not reference any
				// {placeholder}; declaring that here gets the mismatch
				// caught at boot instead of rendering a bare placeholder
				// name mid-game (see legal/doc.go's game-registered
				// predicates section).
				EmittedBindings: map[string][]string{template: nil},
				Evaluate: func(ctx legal.Context) legal.Verdict {
					val, propType, err := ctx.ResolvePath(legal.PropPath(field))
					if err != nil {
						return legal.UnknownVerdict(err.Error())
					}
					if propType != boardgame.TypeEnum {
						return legal.UnknownVerdict("checkers.spaceIsBlack: path " + field + " is not an enum property")
					}
					ev, ok := val.(enum.ImmutableVal)
					if !ok || ev == nil {
						return legal.UnknownVerdict("checkers.spaceIsBlack: path " + field + " resolved to a nil or non-enum value")
					}
					if spaceIsBlack(ev.Value().Int()) {
						return legal.PassVerdict()
					}
					return legal.FailT(template)
				},
			}, nil
		},
	}}
}

// ConfigureLegalTemplates supplies the checkers.* template keys moveMoveToken's
// and movePlaceToken's WithLegalPreconditions plans (and moveMoveToken's LegalCustom
// residue) reference (design spec §8): three override the generic catalog
// defaults with the exact legacy strings from moveMoveToken's pre-migration
// Legal() body (see moves.go's comment for the mapping), one
// ("checkers.black_spaces_only") is the game-registered spaceIsBlack
// predicate's own default template — it has no catalog default to fall back to
// since it isn't part of legal.DefaultTemplates() — and one
// ("checkers.no_more_components") retargets movePlaceToken's StackNotEmpty
// precondition to the exact string its deleted Legal()'s first gate returned.
// See moves.go's moveMoveToken LegalCustom doc comment for the remaining
// graph-walk residue represented by "checkers.illegal_dest".
func (g *gameDelegate) ConfigureLegalTemplates() map[string]string {
	return map[string]string{
		"checkers.no_token_there":     "That space does not have a component in it",
		"checkers.not_your_token":     "that token isn't your token to move",
		"checkers.black_spaces_only":  "you can only move to spaces that are black",
		"checkers.illegal_dest":       "spaceIndex does not represent a legal space for that token to move to",
		"checkers.no_more_components": "No more components to place",
	}
}

func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return boardgame.PropertyCollection{
		"BOARD_SIZE": boardSize,
	}
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() != tokenDeckName {
		return nil
	}
	return new(tokenDynamic)
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game := state.ImmutableGameState().(*gameState)
	if c.Deck().Name() == tokenDeckName {
		return game.UnusedTokens, nil
	}
	return nil, errors.New("Unknown deck")
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	_, players := concreteStates(state)
	for _, p := range players {
		if p.CapturedTokens.NumComponents() >= numTokens {
			return true
		}
	}

	return false
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	return pState.(*playerState).GameScore()
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		tokenDeckName: newTokenDeck(),
	}
}

// NewDelegate is the primary entrypoint of the package, returning a new delegate
// that configures a game of checkers.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
