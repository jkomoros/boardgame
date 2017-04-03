/*

memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"github.com/jkomoros/boardgame"
	"strconv"
	"strings"
)

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

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers < 4 && numPlayers > 1
}

func (g *gameDelegate) EmptyGameState() boardgame.MutableGameState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	return &gameState{
		CurrentPlayer: 0,
		HiddenCards:   boardgame.NewSizedStack(cards, len(cards.Components())),
		RevealedCards: boardgame.NewSizedStack(cards, len(cards.Components())),
	}
}

func (g *gameDelegate) EmptyPlayerState(playerIndex int) boardgame.MutablePlayerState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	return &playerState{
		playerIndex: playerIndex,
		WonCards:    boardgame.NewGrowableStack(cards, 0),
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

		value := "<empty>"

		if c != nil {
			value = c.Values.(*cardValue).Type
		}

		result = append(result, "\t"+value)

	}

	result = append(result, "*****")

	for i, player := range players {
		playerName := "Player" + strconv.Itoa(i)
		if i == game.CurrentPlayer {
			playerName += " *CURRENT* " + strconv.Itoa(player.CardsLeftToReveal)
		}
		result = append(result, playerName)
		result = append(result, strconv.Itoa(player.WonCards.NumComponents()))
	}

	return strings.Join(result, "\n")
}

var policy *boardgame.StatePolicy

func (g *gameDelegate) StateSanitizationPolicy() *boardgame.StatePolicy {

	if policy == nil {
		policy = &boardgame.StatePolicy{
			Game: map[string]boardgame.GroupPolicy{
				"HiddenCards": boardgame.GroupPolicy{
					boardgame.GroupAll: boardgame.PolicyLen,
				},
			},
		}
	}

	return policy

}

func NewManager(storage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	cards := boardgame.NewDeck()

	for _, val := range cardNames {
		cards.AddComponentMulti(&cardValue{
			Type: val,
		}, 2)
	}

	cards.SetShadowValues(&cardValue{
		Type: "<hidden>",
	})

	chest.AddDeck(cardsDeckName, cards)

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		panic("No manager returned")
	}

	manager.AddFixUpMove(&MoveAdvanceNextPlayer{})
	manager.AddPlayerMove(&MoveRevealCard{})
	manager.AddPlayerMove(&MoveHideCards{})

	manager.SetUp()

	return manager
}
