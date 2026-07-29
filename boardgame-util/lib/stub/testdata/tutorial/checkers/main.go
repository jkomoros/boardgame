package checkers

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/moves"
)

/*

Call the code generation for readers and enums here, so "go generate" will generate code correctly.

*/
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

func (g *gameDelegate) DisplayName() string {
	return "CHECKERS!!!"
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		//Seating is legal in both phases, so a player can join before the
		//game starts OR drop in after it has. The pairing with an activation
		//move is load-bearing: SeatPlayer marks everyone it seats as
		//Inactive, and an inactive player can never be the CurrentPlayer, so
		//a player seated in a phase with no activation move would be a silent
		//permanent spectator.
		//
		//The two phases need DIFFERENT activation moves. In phaseSetUp,
		//DefaultRoundSetup's ActivateInactivePlayer activates everyone --
		//correct there, because no seat has been closed yet, and
		//WaitForEnoughPlayers holds the game there until enough players
		//arrive. In phaseNormal, ActivateInactivePlayer would be WRONG: it
		//would re-activate the empty seats DefaultRoundSetup's
		//InactivateEmptySeat just closed, and the game would start waiting on
		//seats nobody is sitting in. ActivateFilledSeat activates only the
		//seats a real player is in, which is exactly what drop-in joining
		//needs.
		moves.AddForPhase(
			phaseSetUp,
			auto.MustConfig(new(moves.SeatPlayer)),
		),
		moves.AddForPhase(
			phaseNormal,
			auto.MustConfig(new(moves.SeatPlayer),
				moves.WithMoveNameSuffix("Mid Game"),
			),
			auto.MustConfig(new(moves.ActivateFilledSeat)),
		),
		moves.AddOrderedForPhase(
			phaseSetUp,
			moves.DefaultRoundSetup(auto),
			//This move will keep on applying itself in round robin fashion
			//until all of the cards are dealt.
			auto.MustConfig(new(moves.DealComponentsUntilPlayerCountReached),
				moves.WithGameProperty("DrawStack"),
				moves.WithPlayerProperty("Hand"),
				moves.WithTargetCount(2),
			),
			//Because we used AddOrderedForPhase, this next move won't apply
			//until the move before it is done applying.
			auto.MustConfig(new(moves.StartPhase),
				moves.WithPhaseToStart(phaseNormal, phaseEnum),
				moves.WithHelpText("Move to the normal play phase."),
			),
		),

		moves.AddForPhase(
			phaseNormal,
			auto.MustConfig(new(moveDrawCard),
				moves.WithHelpText("Draw a card from the deck when it's your turn"),
			),
			//FinishTurn will advance to the next player automatically, when
			//playerState.TurnDone() is true.
			auto.MustConfig(new(moves.FinishTurn)),
		),
	)

}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		exampleCardDeckName: newExampleCardDeck(),
	}
}

func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {

	//ConfigureConstants isn't needed very often. It's useful to ensure a
	//constant value is available client-side, or if you want to use the value
	//in a struct tag.

	return boardgame.PropertyCollection{
		"numCards": numCards,
	}
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() == exampleCardDeckName {
		return new(exampleCardDynamicValues)
	}
	return nil
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	game := state.ImmutableGameState().(*gameState)
	if c.Deck().Name() == exampleCardDeckName {
		return game.DrawStack, nil
	}
	return nil, errors.New("Unknown deck: " + c.Deck().Name())

}

func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {

	//This is the only time that config is passed in, so we need to interpret
	//it now and set it as a property in GameState.
	targetCardsLeftVal := variant[variantKeyTargetCardsLeft]

	var targetCardsLeft int

	switch targetCardsLeftVal {
	case variantTargetCardsLeftShort:
		targetCardsLeft = 2
	case variantTargetCardsLeftDefault:
		targetCardsLeft = 0
	default:
		//This shouldn't happen because NewGame checks that given configs are
		//legal before passing to this method.
		return errors.New("Unknown value for " + variantKeyTargetCardsLeft + ": " + targetCardsLeftVal)
	}

	game := state.GameState().(*gameState)
	game.TargetCardsLeft = targetCardsLeft

	return nil

}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game := state.GameState().(*gameState)
	return game.DrawStack.Shuffle()
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	//base.GameDelegate's CheckGameFinished checks this method and if true
	//looks at the score to see who won.

	//In this example, the game is over once all of the cards are gone.
	return state.ImmutableGameState().(*gameState).CardsDone()
}

// values for the variant setup
const (
	variantKeyTargetCardsLeft = "target_cards_left"
)

const (
	variantTargetCardsLeftDefault = "default"
	variantTargetCardsLeftShort   = "short"
)

func (g *gameDelegate) Variants() boardgame.VariantConfig {

	//variants are the legal configuration options that will be show in the
	//new game dialog.

	return boardgame.VariantConfig{
		variantKeyTargetCardsLeft: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				//Can skip DisplayName because that will be set automatically correctly
				Description: "Whether or not the target cards left is the default",
			},
			Default: variantTargetCardsLeftDefault,
			Values: map[string]*boardgame.VariantDisplayInfo{
				variantTargetCardsLeftShort: {
					Description: "A short game that ends when 2 cards are left",
				},
				variantTargetCardsLeftDefault: {
					Description: "A normal-length game that ends when no cards are left",
				},
			},
		},
	}
}

func (g *gameDelegate) ConfigureComputedProperties() []boardgame.ComputedProperty {
	return []boardgame.ComputedProperty{
		boardgame.GlobalComputedBool("CardsDone", func(state boardgame.ImmutableState) bool {
			return state.ImmutableGameState().(*gameState).CardsDone()
		}),
		boardgame.PlayerComputedInt("RoundScore", func(player boardgame.ImmutableSubState) int {
			return player.(*playerState).GameScore()
		}),
	}
}

// NewDelegate is the primary entrypoint to the package. It implements a
// boardgame.GameDelegate that configures a game of CHECKERS!!!
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
