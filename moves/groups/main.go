/*

	groups is a package of various objects that implement
	moves/interfaces.MoveProgressionGroup, and are thus appropriate for
	passing to moves.AddOrderedForPhase.

*/
package groups

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Serial is a type of move group that represents the moves provided from top
//to bottom, in order. It is one of the most basic types of groups.
type Serial []interfaces.MoveProgressionGroup

func (s Serial) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range s {
		result = append(result, group.MoveConfigs()...)
	}
	return result
}

//Satisfied walks through each sub-group in turn. It errors if no tape is read.
func (s Serial) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {

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
