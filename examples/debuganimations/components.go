package debuganimations

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
	Type string
}
