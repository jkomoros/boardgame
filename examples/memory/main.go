/*

Package memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {

	//If our package name and delegate.Name() don't match, NewGameManager will
	//fail with an error. Given they have to be the same, we might as well
	//just ensure they are actually the same, via a one-time reflection.

	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
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

func (g *gameDelegate) ComputedGlobalProperties(state boardgame.ImmutableState) boardgame.PropertyCollection {
	game, _ := concreteStates(state)
	return boardgame.PropertyCollection{
		"CurrentPlayerHasCardsToReveal": game.CurrentPlayerHasCardsToReveal(),
	}
}

const (
	variantKeyNumCards = "numcards"
	variantKeyCardSet  = "cardset"
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

func (g *gameDelegate) Variants() boardgame.VariantConfig {

	return boardgame.VariantConfig{
		variantKeyCardSet: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				DisplayName: "Card Set",
				Description: "Which theme of cards to use",
			},
			Default: cardSetAll,
			Values: map[string]*boardgame.VariantDisplayInfo{
				cardSetAll: {
					DisplayName: "All Cards",
					Description: "All cards mixed together",
				},
				cardSetFoods: {
					Description: "Food cards",
				},
				cardSetAnimals: {
					Description: "Animal cards",
				},
				cardSetGeneral: {
					Description: "Random cards with no particular theme",
				},
			},
		},
		variantKeyNumCards: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				DisplayName: "Number of Cards",
				Description: "How many cards to use? Larger numbers are more difficult.",
			},
			Default: numCardsMedium,
			Values: map[string]*boardgame.VariantDisplayInfo{
				numCardsMedium: {
					Description: "A default difficulty game",
				},
				numCardsSmall: {
					Description: "An easy game",
				},
				numCardsLarge: {
					Description: "A challenging game",
				},
			},
		},
	}
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	game, _ := concreteStates(state)

	game.CardSet = variant[variantKeyCardSet]

	switch variant[variantKeyNumCards] {
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

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)

	//For now, shunt all cards to UnusedCards. In FinishSetup we'll construct
	//the deck based on config.
	return game.UnusedCards, nil

}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
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

			return errors.New("unexpectedly unable to find the pair when sorting cards to include")
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
			if err := game.UnusedCards.ComponentAt(pairCardIndex).MoveToNextSlot(game.HiddenCards); err != nil {
				return errors.New("Couldn't move pair card to other slot: " + err.Error())
			}
			if err := game.UnusedCards.First().MoveToNextSlot(game.HiddenCards); err != nil {
				return errors.New("Couldn't move first card to other slot: " + err.Error())
			}
		} else {
			if err := game.UnusedCards.ComponentAt(pairCardIndex).SlideToLastSlot(); err != nil {
				return errors.New("Couldn't move pair card to end: " + err.Error())
			}
			if err := game.UnusedCards.First().SlideToLastSlot(); err != nil {
				return errors.New("Couldn't move first card to end: " + err.Error())
			}
		}

	}

	game.HiddenCards.Shuffle()

	players[0].CardsLeftToReveal = 2

	return nil
}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	game, players := concreteStates(state)

	var result []string

	result = append(result, "Board")

	for i, c := range game.Cards.ImmutableComponents() {

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

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, _ := concreteStates(state)

	if game.Cards.NumComponents() > 0 {
		return false
	}

	return true
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	player := pState.(*playerState)

	return player.WonCards.NumComponents()
}

func (g *gameDelegate) ConfigureAgents() []boardgame.Agent {
	return []boardgame.Agent{
		&Agent{},
	}
}

var revealCardMoveName string
var hideCardMoveName string

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	revealCardConfig := auto.MustConfig(
		new(moveRevealCard),
		moves.WithHelpText("Reveals the card at the specified location"),
	)

	hideCardConfig := auto.MustConfig(
		new(moveHideCards),
		moves.WithHelpText("After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."),
	)

	//Save this name so agent can use it and we don't have to worry about
	//string constants that change.
	revealCardMoveName = revealCardConfig.Name()
	hideCardMoveName = hideCardConfig.Name()

	return moves.Add(
		revealCardConfig,
		hideCardConfig,
		auto.MustConfig(
			new(moves.FinishTurn),
		),
		auto.MustConfig(
			new(moveCaptureCards),
			moves.WithHelpText("If two cards are showing and they are the same type, capture them to the current player's hand."),
		),
		auto.MustConfig(
			new(moveStartHideCardsTimer),
			moves.WithHelpText("If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them."),
		),
	)
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		cardsDeckName: newDeck(),
	}
}

//NewDelegate is the primary entrypoint to this package. It returns a
//GameDelegate that configures a memory game.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
