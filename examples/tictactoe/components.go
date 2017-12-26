package tictactoe

//+autoreader
const (
	PhaseBeforeFirstMove = iota
	PhaseAfterFirstMove
)

const (
	X     = "X"
	O     = "O"
	Empty = ""
)

//+autoreader reader
type playerToken struct {
	Value string
}
