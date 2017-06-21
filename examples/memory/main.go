/*

memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"strconv"
	"strings"
)

//go:generate autoreader

var computedPropertiesConfig *boardgame.ComputedPropertiesConfig

//computeCurrentPlayerHasCardsToReveal is used in our ComputedPropertyConfig.
func computeCurrentPlayerHasCardsToReveal(state boardgame.State) (interface{}, error) {

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	return p.CardsLeftToReveal > 0, nil

}

func init() {
	computedPropertiesConfig = &boardgame.ComputedPropertiesConfig{
		Global: map[string]boardgame.ComputedGlobalPropertyDefinition{
			"CurrentPlayerHasCardsToReveal": {
				Dependencies: []boardgame.StatePropertyRef{
					{
						Group:    boardgame.StateGroupGame,
						PropName: "CurrentPlayer",
					},
					{
						Group:    boardgame.StateGroupPlayer,
						PropName: "CardsLeftToReveal",
					},
				},
				PropType: boardgame.TypeBool,
				Compute:  computeCurrentPlayerHasCardsToReveal,
			},
			"CardsInGrid": {
				Dependencies: []boardgame.StatePropertyRef{
					{
						Group:    boardgame.StateGroupGame,
						PropName: "HiddenCards",
					},
					{
						Group:    boardgame.StateGroupGame,
						PropName: "RevealedCards",
					},
				},
				PropType: boardgame.TypeInt,
				Compute: func(state boardgame.State) (interface{}, error) {
					game, _ := concreteStates(state)
					return game.CardsInGrid(), nil
				},
			},
		},
	}
}

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

func (g *gameDelegate) ComputedPropertiesConfig() *boardgame.ComputedPropertiesConfig {
	return computedPropertiesConfig
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers < 4 && numPlayers > 1
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) EmptyGameState() boardgame.MutableSubState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	return &gameState{
		CurrentPlayer:  0,
		HiddenCards:    boardgame.NewSizedStack(cards, len(cards.Components())),
		RevealedCards:  boardgame.NewSizedStack(cards, len(cards.Components())),
		HideCardsTimer: boardgame.NewTimer(),
	}
}

func (g *gameDelegate) EmptyPlayerState(playerIndex boardgame.PlayerIndex) boardgame.MutablePlayerState {

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

var policy *boardgame.StatePolicy

func (g *gameDelegate) StateSanitizationPolicy() *boardgame.StatePolicy {

	if policy == nil {
		policy = &boardgame.StatePolicy{
			Game: map[string]boardgame.GroupPolicy{
				"HiddenCards": {
					boardgame.GroupAll: boardgame.PolicyOrder,
				},
			},
		}
	}

	return policy

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

	moveTypeConfigs := []*boardgame.MoveTypeConfig{
		&moveRevealCardConfig,
		&moveHideCardsConfig,
		&moveAdvanceNextPlayerConfig,
		&moveCaptureCardsConfig,
		&moveStartHideCardsTimerConfig,
	}

	if err := manager.BulkAddMoveTypes(moveTypeConfigs); err != nil {
		panic("Couldn't add moves: " + err.Error())
	}

	manager.AddAgent(&Agent{})

	manager.SetUp()

	return manager
}
