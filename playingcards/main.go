/*

playingcards is a convenience package that helps define and work with a set of
american playing cards.

*/
package playingcards

import (
	"fmt"
	"github.com/jkomoros/boardgame"
)

type Suit string

const (
	SuitSpades   Suit = "\u2660"
	SuitHearts        = "\u2665"
	SuitClubs         = "\u2663"
	SuitDiamonds      = "\u2666"
	SuitJokers        = "Jokers"
)

type Rank int

const (
	RankJoker Rank = iota
	RankAce
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJack
	RankQueen
	RankKing
)

type Card struct {
	Suit Suit
	Rank Rank
}

func (c *Card) Props() []string {
	return boardgame.PropertyReaderPropsImpl(c)
}

func (c *Card) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(c, name)
}

func (c *Card) String() string {
	return fmt.Sprintf("%s %d", c.Suit, c.Rank)
}

//ValuesToCards is designed to be used with stack.ComponentValues().
func ValuesToCards(in []boardgame.PropertyReader) []*Card {
	result := make([]*Card, len(in))
	for i := 0; i < len(in); i++ {
		c := in[i]
		if c == nil {
			result[i] = nil
			continue
		}
		result[i] = c.(*Card)
	}
	return result
}

//NewDeck returns a new deck of playing cards with or without Jokers, ready
//for being added to a chest.
func NewDeck(withJokers bool) *boardgame.Deck {
	cards := &boardgame.Deck{}

	ranks := []Rank{RankAce, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10, RankJack, RankQueen, RankKing}
	suits := []Suit{SuitClubs, SuitDiamonds, SuitHearts, SuitSpades}

	for _, rank := range ranks {
		for _, suit := range suits {
			cards.AddComponent(&Card{
				Suit: suit,
				Rank: rank,
			})
		}
	}

	if withJokers {
		//Add two Jokers
		cards.AddComponentMulti(&Card{
			Suit: SuitJokers,
			Rank: RankJoker,
		}, 2)
	}

	return cards
}
