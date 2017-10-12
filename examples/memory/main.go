/*

memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame"
	"strconv"
	"strings"
)

//go:generate autoreader

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "memory"
}

func (g *gameDelegate) DisplayName() string {
	return "Memory"
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) ComputedGlobalProperties(state boardgame.State) map[string]interface{} {
	game, _ := concreteStates(state)
	return map[string]interface{}{
		"CurrentPlayerHasCardsToReveal": game.CurrentPlayerHasCardsToReveal(),
		"CardsInGrid":                   game.CardsInGrid(),
	}
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers < 4 && numPlayers > 1
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) GameStateConstructor() boardgame.MutableSubState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	//We want to size the stack based on the size of the deck, so we'll do it
	//ourselves and not use tag-based auto-inflation.
	return &gameState{
		HiddenCards:   cards.NewSizedStack(len(cards.Components())),
		RevealedCards: cards.NewSizedStack(len(cards.Components())),
	}
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.MutablePlayerState {

	return &playerState{
		playerIndex: playerIndex,
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)

	return game.HiddenCards, nil

}

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) {
	game, players := concreteStates(state)

	game.HiddenCards.Shuffle()

	players[0].CardsLeftToReveal = 2
}

func (g *gameDelegate) Diagram(state boardgame.State) string {
	game, players := concreteStates(state)

	var result []string

	result = append(result, "Board")

	for i, c := range game.HiddenCards.Components() {

		//If there's no hidden card in this slot, see if there is a revealed one.
		if c == nil {
			c = game.RevealedCards.ComponentAt(i)
		}

		value := fmt.Sprintf("%2d", i) + ": "

		if c == nil {
			value += "<empty>"
		} else {
			value += c.Values.(*cardValue).Type
		}

		result = append(result, "\t"+value)

	}

	result = append(result, "*****")

	for i, player := range players {
		playerName := "Player " + strconv.Itoa(i)
		if boardgame.PlayerIndex(i) == game.CurrentPlayer {
			playerName += " *CURRENT* " + strconv.Itoa(player.CardsLeftToReveal)
		}
		result = append(result, playerName)
		result = append(result, strconv.Itoa(player.WonCards.NumComponents()))
	}

	return strings.Join(result, "\n")
}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {
	game, players := concreteStates(state)

	if game.HiddenCards.NumComponents() != 0 || game.RevealedCards.NumComponents() != 0 {
		return false, nil
	}

	//If we get to here, the game is over. Who won?
	maxScore := 0

	for _, player := range players {
		score := player.WonCards.NumComponents()
		if score > maxScore {
			maxScore = score
		}
	}

	for i, player := range players {
		score := player.WonCards.NumComponents()

		if score >= maxScore {
			winners = append(winners, boardgame.PlayerIndex(i))
		}
	}

	return true, winners

}

func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(nil)

	cards := boardgame.NewDeck()

	for _, val := range cardNames {
		cards.AddComponentMulti(&cardValue{
			Type: val,
		}, 2)
	}

	cards.SetShadowValues(&cardValue{
		Type: "<hidden>",
	})

	if err := chest.AddDeck(cardsDeckName, cards); err != nil {
		return nil, errors.New("Couldn't add deck: " + err.Error())
	}

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		return nil, errors.New("No manager returned")
	}

	moveTypeConfigs := []*boardgame.MoveTypeConfig{
		&moveRevealCardConfig,
		&moveHideCardsConfig,
		&moveFinishTurnConfig,
		&moveCaptureCardsConfig,
		&moveStartHideCardsTimerConfig,
	}

	if err := manager.BulkAddMoveTypes(moveTypeConfigs); err != nil {
		return nil, errors.New("Couldn't add moves: " + err.Error())
	}

	manager.AddAgent(&Agent{})

	if err := manager.SetUp(); err != nil {
		return nil, errors.New("Couldn't set up manager: " + err.Error())
	}

	return manager, nil
}
