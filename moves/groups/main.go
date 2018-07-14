/*

	groups is a package of various objects that implement
	moves/interfaces.MoveProgressionGroup, and are thus appropriate for
	passing to moves.AddOrderedForPhase.

*/
package groups

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Serial is a type of move group that represents the sub-groups provided from top
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

//Parallel is a type of move group that requires all sub-groups to be present,
//but in any order. It is one of the most basic types of groups. If you want
//parallel semantics but don't want to require matching all groups, see
//ParallelCount.
type Parallel []interfaces.MoveProgressionGroup

func (p Parallel) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range p {
		result = append(result, group.MoveConfigs()...)
	}
	return result
}

//Satisfied goes through each item in turn, seeing if any of them can consume
//items off of the front of the tape without erroring. It continues going
//through until all are met, or no more un-triggered items can consume another.
func (p Parallel) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {

	tapeHead := tape

	numItems := len(p)
	//Keep track of items that have matched, by index into self.
	matchedItems := make(map[int]bool, numItems)

	//Continue until all items have been matched.
	for len(matchedItems) < numItems {

		if tapeHead == nil {
			return nil, nil
		}

		matchedIndex := -1
		var rest *interfaces.MoveGroupHistoryItem

		for i, group := range p {
			//Skip items that have already been matched
			if matchedItems[i] {
				continue
			}

			var err error

			err, rest = group.Satisfied(tapeHead)

			if err != nil {
				//That one didn't work
				continue
			}

			matchedIndex = i
			break

		}

		if matchedIndex == -1 {
			//Didn't find any matches
			return errors.New("No more items match, but tape still left."), nil
		}

		//Mark that this item has been matched.
		matchedItems[matchedIndex] = true
		tapeHead = rest

	}

	return nil, tapeHead

}
