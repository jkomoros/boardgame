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

func (a *agentState) CardsToFlip(gameState *gameState) (one, two int) {
	//In our memory is there a pair?

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
		one = -1
		two = -1
		for _, card := range a.LastCards {
			if card.Value == valueToFlip {
				if one == -1 {
					one = card.Index
				} else {
					two = card.Index
					return
				}
			}
		}
		//If we got to here something weird happened.
		log.Println("We thought we found two cards with same value in memory, but I guess we didnt")
		//Reset one and two and just return random cards, below
		one = -1
		two = -1
	}

	//Meh, we don't know which one to flip, flip any cards that haven't been
	//seen in memory and are not empty.

	for one == -1 && two == -1 {
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
		if one == -1 {
			one = index
		} else {
			two = index
		}
	}

	return

}
