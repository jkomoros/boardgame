package moves

import (
	"github.com/jkomoros/boardgame"
)

/*
Serial returns a type of move group that represents the sub-groups provided
from top to bottom, in order. It is one of the most basic types of groups.

Its Satisfied walks through each sub-group in turn. It errors if no tape is
read.
*/
func Serial(children ...MoveProgressionGroup) MoveProgressionGroup {
	return serial(children)
}

type serial []MoveProgressionGroup

func (s serial) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range s {
		result = append(result, group.MoveConfigs()...)
	}
	return result
}

func (s serial) Satisfied(tape *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {

	tapeHead := tape

	for _, group := range s {

		if tapeHead == nil {
			return nil, nil
		}

		rest, err := group.Satisfied(tapeHead)
		if err != nil {
			return tape, err
		}

		tapeHead = rest

	}

	return tapeHead, nil

}
