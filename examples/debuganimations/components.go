package debuganimations

import (
	"github.com/jkomoros/boardgame"
)

//boardgame:codegen
const (
	Phase = iota
	PhaseNormal
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
	"🍺",
	"🌮",
	"🌭",
	"🧀",
	"🥐",
}

const cardsDeckName = "cards"
const tokensDeckName = "tokens"

//boardgame:codegen reader
type cardValue struct {
	boardgame.BaseComponentValues
	Type string
}
