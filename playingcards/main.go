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

func (c *Card) Reader() boardgame.PropertyReader {
	return boardgame.NewDefaultReader(c)
}

func (c *Card) String() string {
	return fmt.Sprintf("%s %d", c.Suit, c.Rank)
}

//ValuesToCards is designed to be used with stack.ComponentValues().
func ValuesToCards(in []boardgame.ComponentValues) []*Card {
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

//NewDeckMulti is like NewDeck, but returns count normal decks together, in
//canonical order. Useful for e.g. casino games where there might be four
//decks shuffled together for the draw stack.
func NewDeckMulti(count int, withJokers bool) *boardgame.Deck {

	if count < 1 {
		count = 1
	}

	cards := &boardgame.Deck{}

	for i := 0; i < count; i++ {
		deckCanonicalOrder(cards, withJokers)
	}

	return cards

}

//NewDeck returns a new deck of playing cards with or without Jokers in a
//canonical, stable order, ready for being added to a chest.
func NewDeck(withJokers bool) *boardgame.Deck {
	cards := &boardgame.Deck{}

	deckCanonicalOrder(cards, withJokers)

	return cards
}

func deckCanonicalOrder(cards *boardgame.Deck, withJokers bool) {
	ranks := []Rank{RankAce, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10, RankJack, RankQueen, RankKing}
	suits := []Suit{SuitSpades, SuitHearts, SuitClubs, SuitDiamonds}

	for _, suit := range suits {
		for _, rank := range ranks {
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
}
