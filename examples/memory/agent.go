package memory

import (
	"github.com/jkomoros/boardgame"
	"log"
	"math/rand"
)

type Agent struct{}

type agentCardInfo struct {
	Value string
	Index int
}

type agentState struct {
	LastCards      []agentCardInfo
	MemoryLength   int
	NextCardToFlip int
}

func (a *Agent) Name() string {
	return "ai"
}

func (a *Agent) DisplayName() string {
	return "Robby the Robot"
}

func (a *Agent) SetUpForGame(game *boardgame.Game, player boardgame.PlayerIndex) []byte {
	//TODO: do something
	return nil
}

func (a *Agent) ProposeMove(game *boardgame.Game, player boardgame.PlayerIndex, agentState []byte) (move boardgame.Move, newState []byte) {
	//TODO: do something
	return nil, nil
}

//CullInvalidCards removes any remembered cards that no longer exist.
func (a *agentState) CullInvalidCards(gameState *gameState) {
	i := 0
	for i < len(a.LastCards) {
		card := a.LastCards[i]
		if c := gameState.HiddenCards.ComponentAt(card.Index); c != nil {
			//This card is still legit.
			i++
			continue
		}
		a.LastCards = append(a.LastCards[:i], a.LastCards[i+1:]...)
		//DON'T increment i; the next index is now i
	}
}

//CardSeen is called when a card is visible. If will return true if that was
//new information, or false if not.
func (a *agentState) CardSeen(value string, index int) bool {

	//Is this card currently in the known set of cards?
	for _, card := range a.LastCards {
		if card.Value == value && card.Index == index {
			return false
		}
	}

	//Add it to the list.

	info := agentCardInfo{
		Value: value,
		Index: index,
	}

	a.LastCards = append([]agentCardInfo{info}, a.LastCards...)

	if len(a.LastCards) > a.MemoryLength {
		//Trim it if there are more cards than we can remember at once.
		a.LastCards = a.LastCards[:a.MemoryLength]
	}

	return true

}

//FlippedCard should be -1 if no cards are currently flipped.
func (a *agentState) FirstCardToFlip(gameState *gameState) int {
	//In our memory is there a pair?

	a.CullInvalidCards(gameState)

	seenValues := make(map[string]bool)
	valueToFlip := ""

	for _, card := range a.LastCards {
		if seenValues[card.Value] {
			//We saw two of the same cards, that's the one we want!
			valueToFlip = card.Value
			break
		}
		seenValues[card.Value] = true
	}

	if valueToFlip != "" {
		//Find the cards and return them.
		for _, card := range a.LastCards {
			if card.Value == valueToFlip {
				return card.Index
			}
		}
	}

	//Meh, we don't know which one to flip, flip any cards that haven't been
	//seen in memory and are not empty.
	return a.PickRandomCard(gameState)

}

func (a *agentState) PickRandomCard(gameState *gameState) int {
	for {
		index := rand.Intn(gameState.HiddenCards.Len())

		//Make sure that index actually is for a card that exists.
		if c := gameState.HiddenCards.ComponentAt(index); c == nil {
			continue
		}

		//Make sure that index isn't one we're already aware of.
		ok := true

		for _, card := range a.LastCards {
			if card.Index == index {
				ok = false
				break
			}
		}

		if !ok {
			continue
		}

		//OK, this sems like a good index.
		return index
	}

	return -1
}

func (a *agentState) SecondCardToFlip(gameState *gameState) int {
	a.CullInvalidCards(gameState)

	flippedCard := ""

	for _, c := range gameState.RevealedCards.Components() {
		if c == nil {
			continue
		}
		flippedCard = c.Values.(*cardValue).Type
	}

	if flippedCard == "" {
		log.Println("We were told to flip the second card but there wasn't already one flipped")
		return -1
	}

	//Is the flippedCard one where we know where the other one is?
	for _, card := range a.LastCards {
		if card.Value == flippedCard {
			return card.Index
		}
	}

	//Otherwise, just pick a random card
	return a.PickRandomCard(gameState)
}
