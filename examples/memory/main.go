/*

memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/auto"
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

func (g *gameDelegate) Description() string {
	return "Players flip over two cards at a time and keep any matches they find"
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 6
}

func (g *gameDelegate) ComputedGlobalProperties(state boardgame.State) boardgame.PropertyCollection {
	game, _ := concreteStates(state)
	return boardgame.PropertyCollection{
		"CurrentPlayerHasCardsToReveal": game.CurrentPlayerHasCardsToReveal(),
	}
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
	if err := game.VisibleCards.SetSize(game.NumCards); err != nil {
		return errors.New("Couldn't set up revealed cards: " + err.Error())
	}

	return nil
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c boardgame.Component) (boardgame.Stack, error) {
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
		firstCard := game.UnusedCards.ComponentAt(0).Values().(*cardValue)

		//Now find its pair. If we keep it, we'll also keep its pair. If we
		//move it to scratch, we'll also move its pair to scratch.
		var pairCardIndex int

		for i := 1; i < game.UnusedCards.Len(); i++ {
			candidateCard := game.UnusedCards.ComponentAt(i).Values().(*cardValue)

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

		//Doing the pair card first means that its index doesn't have to be
		//modified down by 1
		if useCard {
			game.UnusedCards.MoveComponent(pairCardIndex, game.HiddenCards, boardgame.NextSlotIndex)
			game.UnusedCards.MoveComponent(0, game.HiddenCards, boardgame.NextSlotIndex)
		} else {
			game.UnusedCards.MutableComponentAt(pairCardIndex).MoveToEnd()
			game.UnusedCards.MutableFirst().MoveToEnd()
		}

	}

	game.HiddenCards.Shuffle()

	players[0].CardsLeftToReveal = 2

	return nil
}

func (g *gameDelegate) Diagram(state boardgame.State) string {
	game, players := concreteStates(state)

	var result []string

	result = append(result, "Board")

	for i, c := range game.Cards.Components() {

		value := fmt.Sprintf("%2d", i) + ": "

		if c == nil {
			value += "<empty>"
		} else {
			value += c.Values().(*cardValue).Type
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

func (g *gameDelegate) GameEndConditionMet(state boardgame.State) bool {
	game, _ := concreteStates(state)

	if game.Cards.NumComponents() > 0 {
		return false
	}

	return true
}

func (g *gameDelegate) PlayerScore(pState boardgame.PlayerState) int {
	player := pState.(*playerState)

	return player.WonCards.NumComponents()
}

func (g *gameDelegate) ConfigureAgents() []boardgame.Agent {
	return []boardgame.Agent{
		&Agent{},
	}
}

var revealCardMoveName string

func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {

	revealCardConfig := auto.MustConfig(
		new(MoveRevealCard),
		moves.WithHelpText("Reveals the card at the specified location"),
	)

	//Save this name so agent can use it and we don't have to worry about
	//string constants that change.
	revealCardMoveName = revealCardConfig.Name

	return boardgame.NewMoveTypeConfigBundle().AddMoves(
		revealCardConfig,
		auto.MustConfig(
			new(MoveHideCards),
			moves.WithHelpText("After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
		),
		auto.MustConfig(
			new(MoveCaptureCards),
			moves.WithHelpText("If two cards are showing and they are the same type, capture them to the current player's hand."),
		),
		auto.MustConfig(
			new(MoveStartHideCardsTimer),
			moves.WithHelpText("If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them."),
		),
	)
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		cardsDeckName: newDeck(),
	}
}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
