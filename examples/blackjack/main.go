/*

blackjack implements a simple blackjack game. This example is interesting
because it has hidden state.

*/
package blackjack

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/moves"
	"strings"
	"time"
)

//go:generate boardgame-util codegen

const targetScore = 21

const gameDisplayname = "Blackjack"
const gameName = "blackjack"

//computeHandValue is used in our ComputedPropertyConfig.
func computeHandValue(state boardgame.PlayerState) (interface{}, error) {

	playerState := state.(*playerState)

	return playerState.HandValue(), nil

}

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return gameName
}

func (g *gameDelegate) DisplayName() string {
	return gameDisplayname
}

func (g *gameDelegate) Description() string {
	return "Players draw cards trying to get as close to 21 as possible without going over"
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 7
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 4
}

func (g *gameDelegate) ComputedPlayerProperties(player boardgame.ImmutablePlayerState) boardgame.PropertyCollection {

	p := player.(*playerState)

	return boardgame.PropertyCollection{
		"HandValue": p.HandValue(),
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	game, _ := concreteStates(state)

	card := c.Values().(*playingcards.Card)

	if card.Rank.Value() == playingcards.RankJoker {
		return game.UnusedCards, nil
	}

	return game.DrawStack, nil

}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {

	game, players := concreteStates(state)

	var result []string

	result = append(result, fmt.Sprintf("Cards left in deck: %d", game.DrawStack.NumComponents()))

	for i, player := range players {

		playerLine := fmt.Sprintf("Player %d", i)

		if boardgame.PlayerIndex(i) == game.CurrentPlayer {
			playerLine += "  *CURRENT*"
		}

		result = append(result, playerLine)

		handValue := player.HandValue()

		statusLine := fmt.Sprintf("\tValue: %d", handValue)

		if player.Busted {
			statusLine += " BUSTED"
		}

		if player.Stood {
			statusLine += " STOOD"
		}

		result = append(result, statusLine)

		result = append(result, "\tCards:")

		for _, c := range player.HiddenHand.Components() {
			result = append(result, "\t\t"+c.Values().(*playingcards.Card).String())
		}

		for _, c := range player.VisibleHand.Components() {
			result = append(result, "\t\t"+c.Values().(*playingcards.Card).String())
		}

		result = append(result, "")
	}

	return strings.Join(result, "\n")
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutablePlayerState) int {
	player := pState.(*playerState)

	if player.Busted {
		return 0
	}

	return player.HandValue()
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	_, players := concreteStates(state)

	for _, player := range players {
		if !player.Busted && !player.Stood {
			return false
		}
	}

	return true
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: playerIndex,
	}
}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.Add(
			auto.MustConfig(
				new(MoveShuffleDiscardToDraw),
				moves.WithHelpText("When the draw deck is empty, shuffles the discard deck into draw deck."),
			),
		),
		moves.AddForPhase(PhaseNormalPlay,
			auto.MustConfig(
				new(MoveCurrentPlayerHit),
				moves.WithHelpText("The current player hits, drawing a card."),
			),
			auto.MustConfig(
				new(MoveCurrentPlayerStand),
				moves.WithHelpText("If the current player no longer wants to draw cards, they can stand."),
			),
			auto.MustConfig(
				new(MoveRevealHiddenCard),
				moves.WithHelpText("Reveals the hidden card in the user's hand"),
				moves.WithIsFixUp(true),
			),
			auto.MustConfig(
				new(moves.FinishTurn),
				moves.WithHelpText("When the current player has either busted or decided to stand, we advance to next player."),
			),
		),
		moves.AddOrderedForPhase(PhaseInitialDeal,
			auto.MustConfig(
				new(moves.DealCountComponents),
				moves.WithMoveName("Deal Initial Hidden Card"),
				moves.WithHelpText("Deals a hidden card to each player"),
				moves.WithGameProperty("DrawStack"),
				moves.WithPlayerProperty("HiddenHand"),
				moves.WithPostAnimationDelay(time.Duration(750)*time.Millisecond),
			),
			auto.MustConfig(
				new(moves.DealCountComponents),
				moves.WithMoveName("Deal Initial Visible Card"),
				moves.WithHelpText("Deals a visible card to each player"),
				moves.WithGameProperty("DrawStack"),
				moves.WithPlayerProperty("VisibleHand"),
				moves.WithPostAnimationDelay(time.Duration(750)*time.Millisecond),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithPhaseToStart(PhaseNormalPlay, PhaseEnum),
			),
		),
	)

}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		"cards": playingcards.NewDeck(false),
	}
}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
