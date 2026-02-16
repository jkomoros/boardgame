/*

Package debuganimations is a very simple debug "game" designed to allow us to
exercise component animations very directly and purely, in order to build and
debug that system.

*/
package debuganimations

import (
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
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

func (g *gameDelegate) DisplayName() string {
	return "Animations Debugger"
}

func (g *gameDelegate) Description() string {
	return "A game type designed to test all of the stack animations in one place"
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 2
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)

	if c.Deck().Name() == tokensDeckName {

		if game.TokensTo.NumComponents() < 9 {
			return game.TokensTo, nil
		}

		if game.SanitizedTokensFrom.NumComponents() < 10 {
			return game.SanitizedTokensFrom, nil
		}

		if game.SanitizedTokensTo.NumComponents() < 9 {
			return game.SanitizedTokensTo, nil
		}

		return game.TokensFrom, nil
	}

	if game.FirstShortStack.NumComponents() < 1 {
		return game.FirstShortStack, nil
	}

	if game.SecondShortStack.NumComponents() < 1 {
		return game.SecondShortStack, nil
	}

	if game.DiscardStack.NumComponents() < 2 {
		return game.DiscardStack, nil
	}

	if game.HiddenCard.NumComponents() < 1 {
		return game.HiddenCard, nil
	}

	if game.FanStack.NumComponents() < 6 {
		return game.FanStack, nil
	}

	if game.FanDiscard.NumComponents() < 3 {
		return game.FanDiscard, nil
	}

	if game.VisibleStack.NumComponents() < 5 {
		return game.VisibleStack, nil
	}

	if game.HiddenStack.NumComponents() < 4 {
		return game.HiddenStack, nil
	}

	if game.AllVisibleStack.NumComponents() < 4 {
		return game.AllVisibleStack, nil
	}

	return game.DrawStack, nil

}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil

}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	return "Not implemented"
}

func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {
	//This debug game is never finished
	return false, nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Add(
		auto.MustConfig(new(moveMoveCardBetweenShortStacks)),
		auto.MustConfig(new(moveMoveCardBetweenDrawAndDiscardStacks)),
		auto.MustConfig(new(moveFlipHiddenCard),
			moves.WithMoveName("Flip Card Between Hidden and Revealed")),
		auto.MustConfig(new(moveMoveCardBetweenFanStacks),
			moves.WithMoveName("Move Fan Card"),
		),
		auto.MustConfig(new(moveVisibleShuffleCards),
			moves.WithMoveName("Visible Shuffle"),
		),
		auto.MustConfig(new(moveShuffleCards),
			moves.WithMoveName("Shuffle"),
		),
		auto.MustConfig(new(moveMoveBetweenHidden)),
		auto.MustConfig(new(moveMoveToken)),
		auto.MustConfig(new(moveMoveTokenSanitized)),
		auto.MustConfig(new(moveStartMoveAllComponentsToHidden)),
		auto.MustConfig(new(moveStartMoveAllComponentsToVisible)),
		auto.MustConfig(new(moveShuffleHidden),
			moves.WithMoveName("Shuffle Hidden"),
		),
	)
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {

	cards := boardgame.NewDeck()

	for _, val := range cardNames {
		for i := 0; i < 3; i++ {
			cards.AddComponent(&cardValue{
				Type: val,
			})
		}
	}

	cards.SetGenericValues(&cardValue{
		Type: "<hidden>",
	})

	tokens := boardgame.NewDeck()

	for i := 0; i < 38; i++ {
		tokens.AddComponent(nil)
	}

	return map[string]*boardgame.Deck{
		cardsDeckName:  cards,
		tokensDeckName: tokens,
	}
}

//NewDelegate is the primary entrypoint of this package. It returns a new
//delegate that configures a game of debuganimations.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
