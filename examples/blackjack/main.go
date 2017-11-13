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

func (g *gameDelegate) DefaultNumPlayers() int {
	return 4
}

func (g *gameDelegate) CurrentPhase(state boardgame.State) int {
	game, _ := concreteStates(state)
	return game.Phase.Value()
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
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
	} else {
		return game.DrawStack, nil
	}

	return nil, errors.New("Unexpected control point reached")

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

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {

	_, players := concreteStates(state)

	for _, player := range players {
		if !player.Busted && !player.Stood {
			return false, nil
		}
	}

	//OK, everyone has either busted or Stood. So who won?

	maxScore := 0

	for _, player := range players {
		if player.Busted {
			continue
		}

		handValue := player.HandValue()
		if handValue > maxScore {
			maxScore = handValue
		}
	}

	//OK, now who got the maxScore?

	var result []boardgame.PlayerIndex

	for i, player := range players {
		if player.Busted {
			continue
		}
		handValue := player.HandValue()
		if handValue == maxScore {
			result = append(result, boardgame.PlayerIndex(i))
		}
	}

	return true, result

}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers > 0 && numPlayers < 7
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

func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(Enums)

	if err := chest.AddDeck("cards", playingcards.NewDeck(false)); err != nil {
		return nil, errors.New("Couldn't add deck: " + err.Error())
	}

	delegate := &gameDelegate{}

	manager := boardgame.NewGameManager(delegate, chest, storage)

	if manager == nil {
		return nil, errors.New("No manager returned")
	}

	err := manager.AddMoves(
		&moveShuffleDiscardToDrawConfig,
	)

	if err != nil {
		return nil, errors.New("Couldn't install general moves: " + err.Error())
	}

	err = manager.AddMovesForPhase(PhaseNormalPlay,
		&moveCurrentPlayerHitConfig,
		&moveCurrentPlayerStandConfig,
		&moveRevealHiddenCardConfig,
		&moveFinishTurnConfig,
	)

	if err != nil {
		return nil, errors.New("Couldn't install normal phase moves: " + err.Error())
	}

	manager.AddOrderedMovesForPhase(PhaseInitialDeal,
		&moveDealInitialHiddenCardConfig,
		&moveDealInitialVisibleCardConfig,
		moves.NewStartPhaseMoveConfig(manager, PhaseNormalPlay, nil),
	)

	if err != nil {
		return nil, errors.New("Couldn't install initial deal moves: " + err.Error())
	}

	if err := manager.SetUp(); err != nil {
		return nil, errors.New("Couldn't set up manager: " + err.Error())
	}

	return manager, nil
}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := manager.NewGame()

	if err := game.SetUp(0, nil, nil); err != nil {
		panic(err)
	}

	return game
}
