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
//number of times count is looking for, in serial. Note that count.AtMost and
//count.AtLeast will work as expected, with Repeat matching as many moves as
//it can while staying within those bounds. It is conceptually equivalent to
//duplicating a given group within a parent groups.Serial count times.
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

	//Keep track of if we start off with the count being an err. That means
	//that we're likely dealing with an AtLeast.
	startedOnErr := false
	if err := r.Count(0, 1); err != nil {
		startedOnErr = true
	}

	//overtime is a signal that we are working with an AtLeast counter, and
	//found at least one non-nil err, so when we find an err that's ok, just
	//return the last position of the tape-head; otherwise keep consuming as
	//much as we can.
	overtime := false

	for {

		if tapeHead == nil {
			return nil, nil
		}

		if err := r.Count(count, 1); err == nil {

			if startedOnErr {
				//We're in AtLeast mode, and
				overtime = true
			}

			//Break if this count is legal, until the next one isn't (which
			//means we're the last loop through that's legal).

			if err := r.Count(count+1, 1); err != nil {
				break
			}
		}

		err, rest := r.Child.Satisfied(tapeHead)

		if err != nil {
			if overtime {
				//We're in over-time, so errors are not a big deal, just
				//return the last known good state.
				return nil, tapeHead
			}
			return err, nil
		}

		count++
		tapeHead = rest

	}

	return nil, tapeHead

}
