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

func (g *gameDelegate) ComputedGlobalProperties(state boardgame.State) boardgame.PropertyCollection {
	game, _ := concreteStates(state)
	return boardgame.PropertyCollection{
		"CurrentPlayerHasCardsToReveal": game.CurrentPlayerHasCardsToReveal(),
		"CardsInGrid":                   game.CardsInGrid(),
	}
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers < 4 && numPlayers > 1
}

const (
	configKeyNumCards = "numcards"
	configKeyCardSet  = "cardset"
)

const (
	numCardsSmall  = "small"
	numCardsMedium = "medium"
	numCardsLarge  = "large"
)

const (
	cardSetAll     = "all"
	cardSetFoods   = "foods"
	cardSetAnimals = "animals"
	cardSetGeneral = "general"
)

func (g *gameDelegate) Configs() map[string][]string {
	return map[string][]string{
		configKeyCardSet:  {cardSetAll, cardSetFoods, cardSetAnimals, cardSetGeneral},
		configKeyNumCards: {numCardsMedium, numCardsSmall, numCardsLarge},
	}
}

func (g *gameDelegate) ConfigKeyDisplay(key string) (displayName, description string) {
	switch key {
	case configKeyCardSet:
		return "Card Set", "Which theme of cards to use"
	case configKeyNumCards:
		return "Number of Cards", "How many cards to use? Larger numbers are more difficult."
	}
	return "", ""
}

func (g *gameDelegate) ConfigValueDisplay(key, val string) (displayName, description string) {
	switch key {
	case configKeyCardSet:
		switch val {
		case cardSetAll:
			return "All Cards", "All cards mixed together"
		case cardSetAnimals:
			return "Animals", "Animal cards"
		case cardSetFoods:
			return "Foods", "Food cards"
		case cardSetGeneral:
			return "General", "Random cards with no particular theme"
		}
	case configKeyNumCards:
		switch val {
		case numCardsSmall:
			return "Small", "An easy game"
		case numCardsMedium:
			return "Medium", "A default difficulty game"
		case numCardsLarge:
			return "Large", "A challenging game"
		}
	}
	return "", ""
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {

	return &playerState{
		playerIndex: playerIndex,
	}
}

func (g *gameDelegate) BeginSetUp(state boardgame.MutableState, config boardgame.GameConfig) error {
	game, _ := concreteStates(state)

	game.CardSet = config[configKeyCardSet]

	if game.CardSet == "" {
		game.CardSet = cardSetAll
	}

	switch config[configKeyNumCards] {
	case numCardsSmall:
		game.NumCards = 10
	case numCardsMedium:
		game.NumCards = 20
	case numCardsLarge:
		game.NumCards = 40
	default:
		game.NumCards = 20
	}

	if err := game.HiddenCards.SetSize(game.NumCards); err != nil {
		return errors.New("Couldn't set up hidden cards: " + err.Error())
	}
	if err := game.RevealedCards.SetSize(game.NumCards); err != nil {
		return errors.New("Couldn't set up revealed cards: " + err.Error())
	}

	return nil
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)

	//For now, shunt all cards to UnusedCards. In FinishSetup we'll construct
	//the deck based on config.
	return game.UnusedCards, nil

}

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	//First, shuffle unused cards to ensure a different set of cards that
	//adhere to config each time.

	game.UnusedCards.Shuffle()

	//Now, go assemble the deck by going through each component from the
	//front, seeing if it matches. If it does, put it in the HiddenCards array
	//and find its match and also put it in the HiddenCards. If it doesn't,
	//put it in the UnusedCardsScratch (along with its pair) to get it out of
	//the way.

	for game.HiddenCards.NumComponents() < game.NumCards {

		//The card to match.
		firstCard := game.UnusedCards.ComponentAt(0).Values.(*cardValue)

		//Now find its pair. If we keep it, we'll also keep its pair. If we
		//move it to scratch, we'll also move its pair to scratch.
		var pairCardIndex int

		for i := 1; i < game.UnusedCards.Len(); i++ {
			candidateCard := game.UnusedCards.ComponentAt(i).Values.(*cardValue)

			if candidateCard.Type == firstCard.Type {
				pairCardIndex = i
				break
			}
		}

		if pairCardIndex == 0 {
			//Uh oh, couldn't find the pair...

			return errors.New("Unexpectedly unable to find the pair when sorting cards to include.")
		}

		useCard := false

		if game.CardSet == cardSetAll {
			useCard = true
		} else if game.CardSet == firstCard.CardSet {
			useCard = true
		}

		targetStack := game.UnusedCardsScratch

		if useCard {
			targetStack = game.HiddenCards
		}

		game.UnusedCards.MoveComponent(0, targetStack, boardgame.NextSlotIndex)
		//Index of the pair card has moved down one since we just popped off
		//the front item.
		pairCardIndex -= 1
		game.UnusedCards.MoveComponent(pairCardIndex, targetStack, boardgame.NextSlotIndex)

	}

	//Scratch shouldn't be full outside of this method; move everything back
	//over there.
	game.UnusedCardsScratch.MoveAllTo(game.UnusedCards)

	game.HiddenCards.Shuffle()

	players[0].CardsLeftToReveal = 2

	return nil
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

	if err := chest.AddDeck(cardsDeckName, newDeck()); err != nil {
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
