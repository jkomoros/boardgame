package groups

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/count"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Optional returns a MoveProgressionGroup that matches the provided group
//either 0 or 1 times. Equivalent to Repeat() with a count of Between(0, 1).
func Optional(group interfaces.MoveProgressionGroup) interfaces.MoveProgressionGroup {
	return Repeat(count.Between(0, 1), group)
}

//Repeat returns a MoveProgressionGroup that repeats the provided group the
//number of times count is looking for, in serial. Assumes that the
//ValidCounter has a single range of legal count values, where before it they
//are illegal and after it they are legal, and will read as many times from
//the tape as it can within that legal range. The first count value passed is
//1. It is conceptually equivalent to duplicating a given group within a
//parent groups.Serial count times.
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

	count := 1

	//we assume that there is precisely one continguous bound that is legal.
	//We want to go up until we enter the lower bound, then any error we run
	//into within that bound is OK (just return last known good tape position
	//and ignore the group that errored), and then when we reach the upper
	//limit we end.
	lowerBoundReached := false

	if err := r.Count(0, 1); err == nil {
		lowerBoundReached = true
	}

	for {

		//If we ever reach the tape end without having found an erro then it's
		//legal.
		if tapeHead == nil {
			return nil, nil
		}

		err, rest := r.Child.Satisfied(tapeHead)

		if err != nil {
			if lowerBoundReached {
				//We're in over-time, so errors are not a big deal, just
				//return the last known good state.
				return nil, tapeHead
			}
			//Otherwise, we haven't yet gotten the smallest legal amount so we
			//should stop.
			return err, nil
		}

		if lowerBoundReached {
			//As soon as we find the first non-nil count afer we've passed the
			//lower limit we're done.
			if err := r.Count(count, 1); err != nil {
				break
			}
		} else {
			//Is this the transition into the lower legal bound?
			if err := r.Count(count, 1); err == nil {
				lowerBoundReached = true
			}
		}

		count++
		tapeHead = rest

	}

	return nil, tapeHead

}
