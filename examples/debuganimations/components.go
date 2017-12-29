package debuganimations

import (
	"github.com/jkomoros/boardgame"
)

var cardNames []string = []string{
	"🏇",
	"🚴",
	"✋",
	"💘",
	"🎓",
	"🐕",
	"🐄",
	"🐘",
	"🐍",
	"🦀",
	"🍒",
	"🍔",
	"🍭",
}

const cardsDeckName = "cards"
const tokensDeckName = "tokens"

//+autoreader reader
type cardValue struct {
	boardgame.BaseComponentValues
	Type string
}
