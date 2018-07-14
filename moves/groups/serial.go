/*

	groups is a package of various objects that implement
	moves/interfaces.MoveProgressionGroup, and are thus appropriate for
	passing to moves.AddOrderedForPhase. They are defined as functions that
	return anonymous underlying structs so that when used in configuration you
	can avoid needing to wrap your children list with
	[]interfaces.MoveProgressionGroup.

*/
package groups

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Serial returns a type of move group that represents the sub-groups provided
//from top to bottom, in order. It is one of the most basic types of groups.
func Serial(children ...interfaces.MoveProgressionGroup) interfaces.MoveProgressionGroup {
	return serial(children)
}

type serial []interfaces.MoveProgressionGroup

func (s serial) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range s {
		result = append(result, group.MoveConfigs()...)
	}
	return result
}

//Satisfied walks through each sub-group in turn. It errors if no tape is read.
func (s serial) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {

	tapeHead := tape

	for _, group := range s {

		if tapeHead == nil {
			return nil, nil
		}

		err, rest := group.Satisfied(tapeHead)
		if err != nil {
			return err, tape
		}

		tapeHead = rest

	}

	return nil, tapeHead

}
