package nestedimport

import "github.com/jkomoros/boardgame/moves"

//boardgame:codegen
type choosePlayer struct {
	moves.RecordCurrentPlayerChoice
	Choice int
}
