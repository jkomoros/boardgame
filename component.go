package boardgame

import (
	"crypto/sha1"
	"fmt"
	"strconv"
)

//A Component represents a movable resource in the game. Cards, dice, meeples,
//resource tokens, etc are all components. Values is a struct that stores the
//specific values for the component.
type Component interface {
	Values() ComponentValues
	Deck() *Deck
	DeckIndex() int
	//ID returns a semi-stable ID for this component within this game and
	//currentState. Within this game, it will only change when the shuffleCount
	//for this component changes. Across games the Id for the "same" component
	//will be different, in a way that cannot be guessed without access to
	//game.SecretSalt. See the package doc for more on semi-stable Ids for
	//components, what they can be used for, and when they do (and don't) change.
	ID(State) string
	//DynamicValues returns the Dynamic Values for this component in the given
	//state. A convenience so you don't have to go find them within the
	//DynamicComponentValues yourself.
	DynamicValues(state State) SubState
	//Equivalent checks whether two components are equivalent--that is,
	//whether they represent the same index in the same deck.
	Equivalent(other Component) bool

	secretMoveCount(s *state) int
	movedSecretly(s *state)
}

type component struct {
	values ComponentValues
	//The deck we're a part of.
	deck *Deck
	//The index we are in the deck we're in.
	deckIndex int
}

func (c *component) Values() ComponentValues {
	if c == nil {
		return nil
	}
	return c.values
}

func (c *component) Deck() *Deck {
	if c == nil {
		return nil
	}
	return c.deck
}

func (c *component) DeckIndex() int {
	if c == nil {
		return -1
	}
	return c.deckIndex
}

func (c *component) Equivalent(other Component) bool {
	if c == nil && other == nil {
		return true
	}
	if c.Deck().Name() != other.Deck().Name() {
		return false
	}
	return c.DeckIndex() == other.DeckIndex()
}

//ComponentValues is the interface that the Values property of a Component
//must implement. BaseComponentValues is designed to be anonymously embedded
//in your component to implement the latter part of the interface. autoreader
//can be used to implement Reader.
type ComponentValues interface {
	Reader
	//ContainingComponent is the component that this ComponentValues is
	//embedded in. It should return the component that was passed to
	//SetContainingComponent.
	ContainingComponent() Component
	//SetContainingComponent is called to let the component values know what
	//its containing component is.
	SetContainingComponent(c Component)
}

//BaseComponentValues is an optional convenience struct designed to be
//embedded anoymously in your component values to implement
//ContainingComponent() and SetContainingComponent() automatically.
type BaseComponentValues struct {
	c Component
}

func (b *BaseComponentValues) ContainingComponent() Component {
	return b.c
}

func (b *BaseComponentValues) SetContainingComponent(c Component) {
	b.c = c
}

func (c *component) ID(s State) string {

	//Shadow components shouldn't get an Id
	if c.Equivalent(c.Deck().GenericComponent()) {
		return ""
	}

	//S should  never be nil in normal circumstances, but if it is, return an
	//obviously-special Id so it doesn't appear to be the actual Id for this
	//component.
	if s == nil {
		return ""
	}

	st := s.(*state)

	input := "insecuredefaultinput"

	//This is only nil in very weird, testing use cases, like
	//blackjack.handvalue.
	if st != nil {
		game := st.game
		input = game.Id() + game.secretSalt
	}

	input += c.Deck().Name() + strconv.Itoa(c.DeckIndex())

	//The id only ever changes when the item has moved secretly.
	input += strconv.Itoa(c.secretMoveCount(st))

	hash := sha1.Sum([]byte(input))

	return fmt.Sprintf("%x", hash)
}

//secretMoveCount returns the secret move count for this component in the
//given state.
func (c *component) secretMoveCount(s *state) int {

	if c == c.Deck().GenericComponent() {
		return 0
	}

	if s == nil {
		return 0
	}

	deckMoveCount, ok := s.secretMoveCount[c.Deck().Name()]

	//No components in that deck have been moved secretly, I guess.
	if !ok {
		return 0
	}

	if c.DeckIndex() >= len(deckMoveCount) {
		//TODO: warn?
		return 0
	}

	if c.DeckIndex() < 0 {
		//This should never happen
		return 0
	}

	return deckMoveCount[c.DeckIndex()]
}

//movedSecretly increments the secretMoveCount for this component.
func (c *component) movedSecretly(s *state) {
	if c.Equivalent(c.Deck().GenericComponent()) {
		return
	}

	if s == nil {
		return
	}

	deckMoveCount, ok := s.secretMoveCount[c.Deck().Name()]

	//We must be the first component in this deck that has been secretly
	//moved. Create the whole int stack for this group.
	if !ok {
		//The zero-value will be fine
		deckMoveCount = make([]int, len(c.Deck().Components()))
		s.secretMoveCount[c.Deck().Name()] = deckMoveCount
	}

	if c.DeckIndex() >= len(deckMoveCount) {
		//TODO: warn?
		return
	}

	if c.DeckIndex() < 0 {
		//This should never happen
		return
	}

	deckMoveCount[c.DeckIndex()]++

}

func (c *component) DynamicValues(state State) SubState {

	//TODO: test this

	dynamic := state.DynamicComponentValues()

	values := dynamic[c.Deck().Name()]

	if values == nil {
		return nil
	}

	if len(values) <= c.DeckIndex() {
		return nil
	}

	if c.DeckIndex() < 0 {
		//TODO: is this the right beahvior now that we have auto-inflation?
		return c.Deck().Chest().Manager().Delegate().DynamicComponentValuesConstructor(c.Deck())
	}

	return values[c.DeckIndex()]
}
