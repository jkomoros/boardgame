/*

	groups is a package of various objects that implement
	moves/interfaces.MoveProgressionGroup, and are thus appropriate for
	passing to moves.AddOrderedForPhase. They can be nested as often as you'd
	like to express the semantics of your move progression.

	They are defined as functions that return anonymous underlying structs so
	that when used in configuration you can avoid needing to wrap your
	children list with []interfaces.MoveProgressionGroup, saving you typing.

		//Example

		//AddOrderedForPhase accepts move configs from auto.Config, or
		//groups.
		moves.AddOrderedForPhase(PhaseNormal,
			//Top level groups are all joined implicitly into a group.Serial.
			auto.MustConfig(new(MoveZero)),
			groups.Serial(
				auto.MustConfig(new(MoveOne)),
				groups.Optional(
					groups.Serial(
						auto.MustConfig(new(MoveTwo)),
						auto.MustConfig(new(MoveThree)),
					),
				),
				groups.ParallelCount(
					count.Any(),
					auto.MustConfig(new(MoveFour)),
					auto.MustConfig(new(MoveFive)),
					groups.Repeat(
						count.AtMost(2),
						groups.Serial(
							auto.MustConfig(new(MoveSix)),
							auto.MustConfig(new(MoveSeven)),
						),
					),
				),
			),
		)

*/
package groups

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
Serial returns a type of move group that represents the sub-groups provided
from top to bottom, in order. It is one of the most basic types of groups.

Its Satisfied walks through each sub-group in turn. It errors if no tape is
read.
*/
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
