package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/count"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Optional returns a MoveProgressionGroup that matches the provided group
//either 0 or 1 times. Equivalent to Repeat() with a count of Between(0, 1).
func Optional(group MoveProgressionGroup) MoveProgressionGroup {
	return Repeat(count.Between(0, 1), group)
}

//Repeat returns a MoveProgressionGroup that repeats the provided group the
//number of times count is looking for, in serial. Assumes that the
//ValidCounter has a single range of legal count values, where before it they
//are illegal, during the range they are legal, and after it they are illegal
//agin, and will read as many times from the tape as it can within that legal
//range. All ValidCounters in the count package satisfy this. It is
//conceptually equivalent to duplicating a given group within a parent
//groups.Serial count times.
func Repeat(count interfaces.ValidCounter, group MoveProgressionGroup) MoveProgressionGroup {
	return repeat{
		count,
		group,
	}
}

type repeat struct {
	Count interfaces.ValidCounter
	Child MoveProgressionGroup
}

func (r repeat) MoveConfigs() []boardgame.MoveConfig {
	return r.Child.MoveConfigs()
}

func (r repeat) Satisfied(tape *MoveGroupHistoryItem) (error, *MoveGroupHistoryItem) {

	tapeHead := tape

	//we assume that there is precisely one continguous bound that is legal.
	//We want to go up until we enter the lower bound, then any error we run
	//into within that bound is OK (just return last known good tape position
	//and ignore the group that errored), and then when we reach the upper
	//limit we end.
	lowerBoundReached := false

	//Check if we start within the lower bound (for example, a count.AtMost()
	//will start within the legal lower bound.z)
	if err := r.Count(0, 1); err == nil {
		lowerBoundReached = true
	}

	//The count happens after the group has been consumed each time, so by the
	//time we look at this the first time it will have already been one group.
	count := 1

	for {

		//If we ever reach the tape end without having found an error then it's
		//legal.
		if tapeHead == nil {
			return nil, nil
		}

		err, rest := r.Child.Satisfied(tapeHead)

		if err != nil {
			if lowerBoundReached {
				//We're between the lower and upper bound of legal counts, so
				//errors are not a big deal, just return the last known good
				//state.
				return nil, tapeHead
			}
			//Otherwise, we haven't yet gotten the smallest legal amount so we
			//should stop.
			return err, nil
		}

		boundErr := r.Count(count, 1)

		if lowerBoundReached {
			//As soon as we find the first non-nil count afer we've passed the
			//lower bound we're done, because we've passed outside of the
			//legal bound.
			if boundErr != nil {
				break
			}
		} else {
			//Is this the transition into the lower legal bound?
			if boundErr == nil {
				lowerBoundReached = true
			}
		}

		count++
		tapeHead = rest

	}

	return nil, tapeHead

}
