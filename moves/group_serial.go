package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
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
	return s.satisfied(tape, func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
		return group.Satisfied(tapeHead)
	})
}

// SatisfiedWithContext implements StatefulMoveProgressionGroup: identical to
// Satisfied, except each child is evaluated via satisfiedDispatch(child,
// tapeHead, ctx) instead of child.Satisfied(tapeHead) directly, so a
// state-driven child (e.g. RepeatFromProp) nested anywhere inside this
// Serial still receives ctx. See groups.go's StatefulMoveProgressionGroup
// doc comment.
func (s serial) SatisfiedWithContext(tape *MoveGroupHistoryItem, ctx legal.Context) (*MoveGroupHistoryItem, error) {
	return s.satisfied(tape, func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
		return satisfiedDispatch(group, tapeHead, ctx)
	})
}

// satisfied is the shared walk both Satisfied and SatisfiedWithContext use;
// evalChild is how each child group is asked whether it's satisfied
// (context-free for Satisfied, context-aware for SatisfiedWithContext).
func (s serial) satisfied(tape *MoveGroupHistoryItem, evalChild func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error)) (*MoveGroupHistoryItem, error) {

	tapeHead := tape

	for _, group := range s {

		if tapeHead == nil {
			return nil, nil
		}

		rest, err := evalChild(group, tapeHead)
		if err != nil {
			return tape, err
		}

		tapeHead = rest

	}

	return tapeHead, nil

}
