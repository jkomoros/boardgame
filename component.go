package boardgame

import (
	"crypto/sha1"
	"fmt"
	"strconv"
)

//A Component represents a movable resource in the game. Cards, dice, meeples,
//resource tokens, etc are all components. Values is a struct that stores the
//specific values for the component.
type Component struct {
	Values SubState
	//The deck we're a part of.
	Deck *Deck
	//The index we are in the deck we're in.
	DeckIndex int
}

//Id returns a semi-stable ID for this component within this game and
//currentState. Within this game, it will only change when the shuffleCount
//for this component changes. Across games the Id for the "same" component
//will be different, in a way that cannot be guessed without access to
//game.SecretSalt. See the package doc for more on semi-stable Ids for
//components, what they can be used for, and when they do (and don't) change.
func (c *Component) Id(s State) string {
	game := s.(*state).game

	input := game.Id() + game.SecretSalt() + c.Deck.Name() + strconv.Itoa(c.DeckIndex)

	hash := sha1.Sum([]byte(input))

	return fmt.Sprintf("%x", hash)
}

func (c *Component) DynamicValues(state State) SubState {

	//TODO: test this

	dynamic := state.DynamicComponentValues()

	values := dynamic[c.Deck.Name()]

	if values == nil {
		return nil
	}

	if len(values) <= c.DeckIndex {
		return nil
	}

	if c.DeckIndex < 0 {
		return c.Deck.Chest().Manager().Delegate().EmptyDynamicComponentValues(c.Deck)
	}

	return values[c.DeckIndex]
}
