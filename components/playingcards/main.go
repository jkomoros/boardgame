/*

playingcards is a convenience package that helps define and work with a set of
american playing cards.

*/
package playingcards

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

//go:generate autoreader

const (
	SuitUnknown = iota
	SuitSpades
	SuitHearts
	SuitClubs
	SuitDiamonds
	SuitJokers
)

var Enums = enum.NewSet()

var SuitEnum = Enums.MustAdd("Suit", map[int]string{
	SuitUnknown:  "\uFFFD",
	SuitSpades:   "\u2660",
	SuitHearts:   "\u2665",
	SuitClubs:    "\u2663",
	SuitDiamonds: "\u2666",
	SuitJokers:   "Jokers",
})

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

var RankEnum = Enums.MustAdd("Rank", map[int]string{
	RankUnknown: "",
	RankAce:     "Ace",
	Rank2:       "2",
	Rank3:       "3",
	Rank4:       "4",
	Rank5:       "5",
	Rank6:       "6",
	Rank7:       "7",
	Rank8:       "8",
	Rank9:       "9",
	Rank10:      "10",
	RankJack:    "Jack",
	RankQueen:   "Queen",
	RankKing:    "King",
	RankJoker:   "Joker",
})

//+autoreader reader
type Card struct {
	Suit enum.Const
	Rank enum.Const
}

func (c *Card) String() string {
	return fmt.Sprintf("%s %s", c.Suit.String(), c.Rank.String())
}

//ValuesToCards is designed to be used with stack.ComponentValues().
func ValuesToCards(in []boardgame.SubState) []*Card {
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
				Suit: SuitEnum.MustNewConst(suit),
				Rank: RankEnum.MustNewConst(rank),
			})
		}
	}

	if withJokers {
		//Add two Jokers
		cards.AddComponentMulti(&Card{
			Suit: SuitEnum.MustNewConst(SuitJokers),
			Rank: RankEnum.MustNewConst(RankJoker),
		}, 2)
	}

	cards.SetShadowValues(&Card{
		Suit: SuitEnum.MustNewConst(SuitUnknown),
		Rank: RankEnum.MustNewConst(RankUnknown),
	})
}
