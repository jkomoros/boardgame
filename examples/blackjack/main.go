/*

blackjack implements a simple blackjack game. This example is interesting
because it has hidden state.

*/
package blackjack

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/playingcards"
	"strings"
)

const targetScore = 21

const gameDisplayname = "Blackjack"
const gameName = "blackjack"

var computedPropertiesConfig *boardgame.ComputedPropertiesConfig

//computeHandValue is used in our ComputedPropertyConfig.
func computeHandValue(shadow *boardgame.ShadowPlayerState) (interface{}, error) {

	hiddenHand, err := shadow.GrowableStackProp("HiddenHand")

	if err != nil {
		return nil, err
	}

	visibleHand, err := shadow.GrowableStackProp("VisibleHand")

	if err != nil {
		return nil, err
	}

	return handValue(hiddenHand, visibleHand), nil

}

func handValue(hiddenHand *boardgame.GrowableStack, visibleHand *boardgame.GrowableStack) int {
	effectiveHand := append(playingcards.ValuesToCards(hiddenHand.ComponentValues()), playingcards.ValuesToCards(visibleHand.ComponentValues())...)
	var numUnconvertedAces int
	var currentValue int

	for _, card := range effectiveHand {
		switch card.Rank {
		case playingcards.RankAce:
			numUnconvertedAces++
			//We count the ace as 1 now. Later we'll check to see if we can
			//expand any aces.
			currentValue += 1
		case playingcards.RankJack, playingcards.RankQueen, playingcards.RankKing:
			currentValue += 10
		default:
			currentValue += int(card.Rank)
		}
	}

	for numUnconvertedAces > 0 {

		if currentValue >= (targetScore - 10) {
			break
		}

		numUnconvertedAces--
		currentValue += 10
	}

	return currentValue
}

func init() {
	computedPropertiesConfig = &boardgame.ComputedPropertiesConfig{
		PlayerProperties: map[string]boardgame.ComputedPlayerPropertyDefinition{
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

func (g *gameDelegate) ComputedPropertiesConfig() *boardgame.ComputedPropertiesConfig {
	return computedPropertiesConfig
}

func (g *gameDelegate) DistributeComponentToStarterStack(state *boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {

	game, _ := concreteStates(state)

	card := c.Values.(*playingcards.Card)

	if card.Rank == playingcards.RankJoker {
		return game.UnusedCards, nil
	} else {
		return game.DrawStack, nil
	}

	return nil, errors.New("Unexpected control point reached")

}

func (g *gameDelegate) Diagram(state *boardgame.State) string {

	game, players := concreteStates(state)

	var result []string

	result = append(result, fmt.Sprintf("Cards left in deck: %d", game.DrawStack.NumComponents()))

	for i, player := range players {

		playerLine := fmt.Sprintf("Player %d", i)

		if i == game.CurrentPlayer {
			playerLine += "  *CURRENT*"
		}

		result = append(result, playerLine)

		handValue, _ := state.Computed().Player(i).IntProp("HandValue")

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

func (g *gameDelegate) CheckGameFinished(state *boardgame.State) (finished bool, winners []int) {

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

		handValue, _ := state.Computed().Player(i).IntProp("HandValue")
		if handValue > maxScore {
			maxScore = handValue
		}
	}

	//OK, now who got the maxScore?

	var result []int

	for i, player := range players {
		if player.Busted {
			continue
		}
		handValue, _ := state.Computed().Player(i).IntProp("HandValue")
		if handValue == maxScore {
			result = append(result, i)
		}
	}

	return true, result

}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers > 0 && numPlayers < 7
}

func (g *gameDelegate) EmptyGameState() boardgame.GameState {
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

func (g *gameDelegate) EmptyPlayerState(playerIndex int) boardgame.PlayerState {
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

func (g *gameDelegate) FinishSetUp(state *boardgame.State) {
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

	manager.AddPlayerMove(&MoveCurrentPlayerHit{})
	manager.AddPlayerMove(&MoveCurrentPlayerStand{})

	manager.AddFixUpMove(&MoveDealInitialCard{})
	manager.AddFixUpMove(&MoveRevealHiddenCard{})
	manager.AddFixUpMove(&MoveShuffleDiscardToDraw{})
	manager.AddFixUpMove(&MoveAdvanceNextPlayer{})

	manager.SetUp()

	return manager
}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := boardgame.NewGame(manager)

	if err := game.SetUp(0); err != nil {
		panic(err)
	}

	return game
}
