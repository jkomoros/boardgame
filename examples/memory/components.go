package memory

import (
	"github.com/jkomoros/boardgame"
)

var generalCards []string = []string{
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

var foodCards []string = []string{
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

var animalCards []string = []string{
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

//+autoreader reader
type cardValue struct {
	boardgame.BaseComponentValues
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

	cards.SetShadowValues(&cardValue{
		Type: "<hidden>",
	})

	return cards
}
