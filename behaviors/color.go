package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

const colorPropertyName = "Color"

/*
PlayerColor is a struct that's designed to be anonymously embedded in your
playerState. It represents the "color" of that player, and its primary use is
convenience methods it exposes for you to use. It is an error if you use it and
don't call ConnectBehavior from within your playerState's FinishStateSetUp (see
the package doc for more).Typically you also embed a ComponentColor in the
ComponentValues of the components that represent tokens or any other item that
are tied to a specific player color. PlayerColor expects there to be an enum
called 'color' that enumerates the valid colors players may be.
*/
type PlayerColor struct {
	container boardgame.SubState
	Color     enum.Val `enum:"color"`
}

//ConnectBehavior stores a reference to the container, which it needs to
//operate.
func (p *PlayerColor) ConnectBehavior(containgSubState boardgame.SubState) {
	p.container = containgSubState
}

//OwnsToken returns whether this player owns the given token. That is, the given
//component has a property named Color that is the same enum as our Color
//property and they are set to the same value.
func (p *PlayerColor) OwnsToken(c boardgame.Component) bool {
	result, _ := p.ownsTokenImpl(c)
	return result
}

//ownsTokenImpl is the main implementation for OwnsToken. It differs in that it
//also returns whether it is possible for the given player to even conceivably
//own the given component (i.e. there is a prop named Color that is of the same
//enum type as ours). If couldOwn is false, then you can safely skip the rest of
//the components in this deck because it's not possible for any to be owned.
func (p *PlayerColor) ownsTokenImpl(c boardgame.Component) (owns, couldOwn bool) {
	color, err := c.Values().Reader().ImmutableEnumProp(colorPropertyName)
	if err != nil {
		//This means that the Values does not hav a property of the right
		//type, so none of them will in this deck, so bail now.
		return false, false
	}
	if color.Enum() != p.Color.Enum() {
		//If the enums are different for the color property then not only are we
		//not owned, but no other ones in this deck can be owned.
		return false, false
	}
	if color.Equals(p.Color) {
		return true, true
	}
	return false, true
}

//TokenSpaceIndex returns the index of the token that it is within its current
//container. Typically the position of a token within a board has semantic
//significance, for example in chutes and ladders which other spaces are
//adjacent, and this method captures that. If it's in a stack that's part of a
//board, it will return the BoardIndex, but if it's in a normal stack it will
//return normal index. If there are multiple types of components that might be
//returned, use TokenSpaceIndexFromDeck instead.
func (p *PlayerColor) TokenSpaceIndex() int {
	return p.spaceIndexForInstance(p.Token())
}

//TokenSpaceIndexFromDeck is like TokenSpaceIndex but when you only want to
//compare against components that are in a specific deck.
func (p *PlayerColor) TokenSpaceIndexFromDeck(deck *boardgame.Deck) int {
	return p.spaceIndexForInstance(p.TokenFromDeck(deck))
}

func (p *PlayerColor) spaceIndexForInstance(instance boardgame.ComponentInstance) int {
	if instance == nil {
		return 0
	}
	stack, slot, err := instance.ContainingStack()
	if err != nil {
		return 0
	}
	if stack.Board() != nil {
		return stack.BoardIndex()
	}
	return slot
}

//TokenFromDeck searches through the given deck to find a ComponentInstance
//whose Color matchs this player's color, returning the first one it finds. May
//return nil if none are found. Typically if only one type of deck has a Color
//property, you don't need to use this, and can instead use Token().
func (p *PlayerColor) TokenFromDeck(deck *boardgame.Deck) boardgame.ComponentInstance {
	if deck == nil {
		return nil
	}
	if deck.Len() == 0 {
		return nil
	}
	//If we don't have a reference to a container then our configuration is
	//invalid so might as well bail now.
	if p.container == nil {
		return nil
	}

	for _, c := range deck.Components() {
		//The token is ours if we own it.
		owns, couldOwn := p.ownsTokenImpl(c)
		if !couldOwn {
			//This token is not owned, an it's also not possible for any other
			//component in this deck to own it so we can bail early
			return nil
		}
		if owns {
			return c.Instance(p.container.State())
		}
	}
	return nil
}

//Token searches through all decks for a component whose Color property matches
//this player's color, and then returns a ComponentInstance for it within this
//state. This may return nil. Typically you can use this instead of TokenFromDeck.
func (p *PlayerColor) Token() boardgame.ComponentInstance {
	if p.container == nil {
		return nil
	}
	chest := p.container.State().Manager().Chest()
	for _, deckName := range chest.DeckNames() {
		deck := chest.Deck(deckName)
		if instance := p.TokenFromDeck(deck); instance != nil {
			return instance
		}
	}
	return nil
}

/*
ComponentColor is a struct designed to be embedded anonymously in any component
structs that have a color that ties them to a given player color, and used in
conjunction with PlayerColor. It doesn't do anything on its own other than
define a property; it's mainly just a convenience so you don't hav to worry
about making sure the token's color property is named correctly. It assumes that
you have an enum called 'color' that enumerates the valid colors players may be.
*/
type ComponentColor struct {
	Color enum.ImmutableVal `enum:"color"`
}
