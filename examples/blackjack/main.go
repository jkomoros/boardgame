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
	"strings"
)

//go:generate autoreader

const targetScore = 21

const gameDisplayname = "Blackjack"
const gameName = "blackjack"

var computedPropertiesConfig *boardgame.ComputedPropertiesConfig

//computeHandValue is used in our ComputedPropertyConfig.
func computeHandValue(state boardgame.PlayerState) (interface{}, error) {

	playerState := state.(*playerState)

	return playerState.HandValue(), nil

}

func init() {
	computedPropertiesConfig = &boardgame.ComputedPropertiesConfig{
		Player: map[string]boardgame.ComputedPlayerPropertyDefinition{
			"HandValue": boardgame.ComputedPlayerPropertyDefinition{
				Dependencies: []boardgame.StatePropertyRef{
					{
						Group:    boardgame.StateGroupPlayer,
						PropName: "HiddenHand",
					},
					{
						Group:    boardgame.StateGroupPlayer,
						PropName: "VisibleHand",
					},
				},
				PropType: boardgame.TypeInt,
				Compute:  computeHandValue,
			},
		},
	}
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

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) ComputedPropertiesConfig() *boardgame.ComputedPropertiesConfig {
	return computedPropertiesConfig
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {

	game, _ := concreteStates(state)

	card := c.Values.(*playingcards.Card)

	if card.Rank == playingcards.RankJoker {
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

		handValue, _ := state.Computed().Player(boardgame.PlayerIndex(i)).Reader().IntProp("HandValue")

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

	for i, player := range players {
		if player.Busted {
			continue
		}

		handValue, _ := state.Computed().Player(boardgame.PlayerIndex(i)).Reader().IntProp("HandValue")
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
		handValue, _ := state.Computed().Player(boardgame.PlayerIndex(i)).Reader().IntProp("HandValue")
		if handValue == maxScore {
			result = append(result, boardgame.PlayerIndex(i))
		}
	}

	return true, result

}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers > 0 && numPlayers < 7
}

func (g *gameDelegate) EmptyGameState() boardgame.MutableSubState {
	cards := g.Manager().Chest().Deck("cards")

	if cards == nil {
		return nil
	}
	return &gameState{
		DiscardStack:  boardgame.NewGrowableStack(cards, 0),
		DrawStack:     boardgame.NewGrowableStack(cards, 0),
		UnusedCards:   boardgame.NewGrowableStack(cards, 0),
		CurrentPlayer: 0,
	}
}

func (g *gameDelegate) EmptyPlayerState(playerIndex boardgame.PlayerIndex) boardgame.MutablePlayerState {
	cards := g.Manager().Chest().Deck("cards")

	if cards == nil {
		return nil
	}
	return &playerState{
		playerIndex:    playerIndex,
		GotInitialDeal: false,
		HiddenHand:     boardgame.NewGrowableStack(cards, 1),
		VisibleHand:    boardgame.NewGrowableStack(cards, 0),
		Busted:         false,
		Stood:          false,
	}
}

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()
}

var policy *boardgame.StatePolicy

func (g *gameDelegate) StateSanitizationPolicy() *boardgame.StatePolicy {

	if policy == nil {
		policy = &boardgame.StatePolicy{
			Player: map[string]boardgame.GroupPolicy{
				"HiddenHand": boardgame.GroupPolicy{
					boardgame.GroupOther: boardgame.PolicyLen,
				},
			},
			Game: map[string]boardgame.GroupPolicy{
				"DiscardStack": boardgame.GroupPolicy{
					boardgame.GroupAll: boardgame.PolicyLen,
				},
				"DrawStack": boardgame.GroupPolicy{
					boardgame.GroupAll: boardgame.PolicyLen,
				},
			},
		}
	}

	return policy

}

func NewManager(storage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	chest.AddDeck("cards", playingcards.NewDeck(false))

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		panic("No manager returned")
	}

	manager.BulkAddMoveTypes([]*boardgame.MoveTypeConfig{
		&moveCurrentPlayerHitConfig,
		&moveCurrentPlayerStandConfig,
	}, []*boardgame.MoveTypeConfig{
		&moveDealInitialCardConfig,
		&moveRevealHiddenCardConfig,
		&moveShuffleDiscardToDrawConfig,
		&moveAdvanceNextPlayerConfig,
	})

	manager.SetUp()

	return manager
}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := boardgame.NewGame(manager)

	if err := game.SetUp(0, nil); err != nil {
		panic(err)
	}

	return game
}
