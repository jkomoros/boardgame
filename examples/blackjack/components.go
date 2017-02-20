package blackjack

import (
	"fmt"
	"github.com/jkomoros/boardgame"
)

type Suit string

const (
	SuitSpades   Suit = "Spades"
	SuitHearts        = "Hearts"
	SuitClubs         = "Clubs"
	SuitDiamonds      = "Diamonds"
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

//Designed to be used with stack.ComponentValues()
func cards(in []boardgame.PropertyReader) []*Card {
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
