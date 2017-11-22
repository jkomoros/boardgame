/*

blackjack implements a simple blackjack game. This example is interesting
because it has hidden state.

*/
package blackjack

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/moves"
	"strings"
)

//go:generate autoreader

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

func (g *gameDelegate) ComputedPlayerProperties(player boardgame.PlayerState) boardgame.PropertyCollection {

	p := player.(*playerState)

	return boardgame.PropertyCollection{
		"HandValue": p.HandValue(),
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {

	game, _ := concreteStates(state)

	card := c.Values.(*playingcards.Card)

	if card.Rank.Value() == playingcards.RankJoker {
		return game.UnusedCards, nil
	}

	return game.DrawStack, nil

}

func (g *gameDelegate) Diagram(state boardgame.State) string {

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

		for _, card := range playingcards.ValuesToCards(player.HiddenHand.ComponentValues()) {
			result = append(result, "\t\t"+card.String())
		}

		for _, card := range playingcards.ValuesToCards(player.VisibleHand.ComponentValues()) {
			result = append(result, "\t\t"+card.String())
		}

		result = append(result, "")
	}

	return strings.Join(result, "\n")
}

func (g *gameDelegate) PlayerScore(pState boardgame.PlayerState) int {
	player := pState.(*playerState)

	if player.Busted {
		return 0
	}

	return player.HandValue()
}

func (g *gameDelegate) GameEndConditionMent(state boardgame.State) bool {
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

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil
}

func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {

	return boardgame.NewMoveTypeConfigBundle().AddMoves(
		&moveShuffleDiscardToDrawConfig,
	).AddMovesForPhase(PhaseNormalPlay,
		&moveCurrentPlayerHitConfig,
		&moveCurrentPlayerStandConfig,
		&moveRevealHiddenCardConfig,
		&moveFinishTurnConfig,
	).AddOrderedMovesForPhase(PhaseInitialDeal,
		&moveDealInitialHiddenCardConfig,
		&moveDealInitialVisibleCardConfig,
		moves.NewStartPhaseConfig(g.Manager(), PhaseNormalPlay, nil),
	)

}

func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(Enums)

	if err := chest.AddDeck("cards", playingcards.NewDeck(false)); err != nil {
		return nil, errors.New("Couldn't add deck: " + err.Error())
	}

	delegate := &gameDelegate{}

	return boardgame.NewGameManager(delegate, chest, storage)

}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := manager.NewGame()

	if err := game.SetUp(0, nil, nil); err != nil {
		panic(err)
	}

	return game
}
