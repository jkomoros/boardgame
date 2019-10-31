/*

Package playingcards is a convenience package that helps define and work with a
set of american playing cards.

*/
package playingcards

import (
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
)

//go:generate boardgame-util codegen

//boardgame:codegen
const (
	//display:"\uFFFD"
	SuitUnknown = iota
	//display:"\u2660"
	SuitSpades
	//display:"\u2665"
	SuitHearts
	//display:"\u2663"
	SuitClubs
	//display:"\u2666"
	SuitDiamonds
	//dislay:"Jokers"
	SuitJokers
)

//boardgame:codegen
const (
	RankUnknown = iota
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
	RankJoker
)

//Card is a struct, ready for use in your own package without modification, that
//encodes the notion of a typical playing card.
//boardgame:codegen reader
type Card struct {
	base.ComponentValues
	Suit enum.Val
	Rank enum.Val
}

func (c *Card) String() string {
	return fmt.Sprintf("%s %s", c.Suit.String(), c.Rank.String())
}

//NewDeckMulti is like NewDeck, but returns count normal decks together, in
//canonical order. Useful for e.g. casino games where there might be four
//decks shuffled together for the draw stack.
func NewDeckMulti(count int, withJokers bool) *boardgame.Deck {

	if count < 1 {
		count = 1
	}

	cards := boardgame.NewDeck()

	for i := 0; i < count; i++ {
		deckCanonicalOrder(cards, withJokers)
	}

	return cards

}

//NewDeck returns a new deck of playing cards with or without Jokers in a
//canonical, stable order, ready for being added to a chest.
func NewDeck(withJokers bool) *boardgame.Deck {
	cards := boardgame.NewDeck()

	deckCanonicalOrder(cards, withJokers)

	return cards
}

func deckCanonicalOrder(cards *boardgame.Deck, withJokers bool) {
	ranks := []int{RankAce, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10, RankJack, RankQueen, RankKing}
	suits := []int{SuitSpades, SuitHearts, SuitClubs, SuitDiamonds}

	for _, suit := range suits {
		for _, rank := range ranks {
			cards.AddComponent(&Card{
				Suit: SuitEnum.MustNewVal(suit),
				Rank: RankEnum.MustNewVal(rank),
			})
		}
	}

	if withJokers {
		//Add two Jokers
		for i := 0; i < 2; i++ {
			cards.AddComponent(&Card{
				Suit: SuitEnum.MustNewVal(SuitJokers),
				Rank: RankEnum.MustNewVal(RankJoker),
			})
		}
	}

	cards.SetGenericValues(&Card{
		Suit: SuitEnum.MustNewVal(SuitUnknown),
		Rank: RankEnum.MustNewVal(RankUnknown),
	})
}
