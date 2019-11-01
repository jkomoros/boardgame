package debuganimations

import (
	"github.com/jkomoros/boardgame/base"
)

//boardgame:codegen
const (
	phase = iota
	phaseNormal
)

var cardNames = []string{
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
	base.ComponentValues
	Type string
}
