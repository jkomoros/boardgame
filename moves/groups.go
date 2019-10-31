package moves

import (
	"github.com/jkomoros/boardgame"
)

//MoveGroupHistoryItem is a singly-linked list (referred to in various
//comments as a "tape") that is passed to MoveProgressionGroup.Satisfied(). It
//represents a list of all of the moves that have applied so far since
//game.CurrentPhase() last changed.
type MoveGroupHistoryItem struct {
	MoveName string
	Rest     *MoveGroupHistoryItem
}

//MoveProgressionGroup is an object that can be used to define a valid move
//progression. moves.AutoConfigurer().Config() returns objects that fit this
//interface.
type MoveProgressionGroup interface {
	//MoveConfigs should return the full enumeration of contained MoveConfigs
	//within this Group, from left to right and top to bottom. This is used by
	//moves.AddOrderedForPhase to know which MoveConfigs contained within it
	//to install.
	MoveConfigs() []boardgame.MoveConfig

	//Satisfied reads the tape and returns an error if the sequence was not
	//valid (did not match, for example, or the group was configured in an
	//invalid way in general). If it returns a nil error, it should also
	//return the rest of the tape representing the items it did not yet
	//consume. If passed a nil tape, it should immediately return nil, nil. If
	//the top-level MoveProgressionGroup consumes the entire tape and doesn't
	//return an error then the progression is considered valid.
	Satisfied(tape *MoveGroupHistoryItem) (rest *MoveGroupHistoryItem, err error)
}
