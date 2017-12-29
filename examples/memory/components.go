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
		cards.AddComponentMulti(&cardValue{
			Type:    val,
			CardSet: cardSetGeneral,
		}, 2)
	}

	for _, val := range foodCards {
		cards.AddComponentMulti(&cardValue{
			Type:    val,
			CardSet: cardSetFoods,
		}, 2)
	}

	for _, val := range animalCards {
		cards.AddComponentMulti(&cardValue{
			Type:    val,
			CardSet: cardSetAnimals,
		}, 2)
	}

	cards.SetShadowValues(&cardValue{
		Type: "<hidden>",
	})

	return cards
}
