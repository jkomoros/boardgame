package groups

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Repeat returns a MoveProgressionGroup that repeats the provided group the
//number of times count is looking for, in serial. It is conceptually
//equivalent to duplicating a given group within a parent groups.Serial count
//times.
func Repeat(count interfaces.ValidCounter, group interfaces.MoveProgressionGroup) interfaces.MoveProgressionGroup {
	return repeat{
		count,
		group,
	}
}

type repeat struct {
	Count interfaces.ValidCounter
	Child interfaces.MoveProgressionGroup
}

func (r repeat) MoveConfigs() []boardgame.MoveConfig {
	return r.Child.MoveConfigs()
}

func (r repeat) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {

	tapeHead := tape

	count := 0

	for {

		if tapeHead == nil {
			return nil, nil
		}

		if err := r.Count(count, 1); err == nil {
			break
		}

		err, rest := r.Child.Satisfied(tapeHead)

		if err != nil {
			return err, nil
		}

		count++
		tapeHead = rest

	}

	return nil, tapeHead

}
