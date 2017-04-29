package memory

var cardNames []string = []string{
	"Apple",
	"Orange",
	"Pear",
	"Ball",
	"Hammer",
	"Star",
	"Dog",
	"Cat",
}

const cardsDeckName = "cards"

//+autoreader reader
type cardValue struct {
	Type string
}
