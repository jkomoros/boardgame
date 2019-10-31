package memory

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

var generalCards = []string{
	"🚴",
	"✋",
	"💘",
	"🎓",
	"🌍",
	"🏖",
	"🏛",
	"⛺",
	"🚑",
	"🚕",
	"⚓",
	"🕰",
	"🌈",
	"🔥",
	"⛄",
	"🎄",
	"🎁",
	"🏆",
	"⚽",
	"🎳",
}

var foodCards = []string{
	"🍒",
	"🍔",
	"🍭",
	"🍇",
	"🍉",
	"🍊",
	"🍌",
	"🍍",
	"🍓",
	"🌽",
	"🥕",
	"🍗",
	"🍕",
	"🍩",
	"🍦",
	"🍺",
	"🌮",
	"🌭",
	"🧀",
	"🥐",
}

var animalCards = []string{
	"🐕",
	"🐄",
	"🐘",
	"🐍",
	"🦀",
	"🏇",
	"🦍",
	"🐈",
	"🐖",
	"🐫",
	"🐁",
	"🐿",
	"🦇",
	"🐓",
	"🦅",
	"🦉",
	"🐋",
	"🦑",
	"🐝",
	"🐡",
}

const cardsDeckName = "cards"

//boardgame:codegen reader
type cardValue struct {
	base.ComponentValues
	Type    string
	CardSet string
}

func newDeck() *boardgame.Deck {
	cards := boardgame.NewDeck()

	for _, val := range generalCards {
		for i := 0; i < 2; i++ {
			cards.AddComponent(&cardValue{
				Type:    val,
				CardSet: cardSetGeneral,
			})
		}
	}

	for _, val := range foodCards {
		for i := 0; i < 2; i++ {
			cards.AddComponent(&cardValue{
				Type:    val,
				CardSet: cardSetFoods,
			})
		}
	}

	for _, val := range animalCards {
		for i := 0; i < 2; i++ {
			cards.AddComponent(&cardValue{
				Type:    val,
				CardSet: cardSetAnimals,
			})
		}
	}

	cards.SetGenericValues(&cardValue{
		Type: "<hidden>",
	})

	return cards
}
