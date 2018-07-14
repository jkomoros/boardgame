package groups

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/count"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"math"
	"sort"
)

/*
Parallel is a type of move group that requires all sub-groups to be present,
but in any order. It is one of the most basic types of groups. If you want
parallel semantics but don't want to require matching all groups, see
ParallelCount. The base Parallel is equivalent to ParallelCount with a Count
of count.All().

Its Satisfied goes through each item in turn, seeing if any of them can consume
items off of the front of the tape without erroring. It continues going
through until all are met, or no more un-triggered items can consume
another. If at any point more than one item could match at the given point
in the tape, it chooses the match that consumes the most tape.
*/
func Parallel(children ...interfaces.MoveProgressionGroup) interfaces.MoveProgressionGroup {
	return ParallelCount(count.All(), children...)
}

//matchInfo reflects a match that was found while doing a run through of
//groups, for example in a Parallel group.
type matchInfo struct {
	//The index within the containing group of the group that mached
	index int
	//The tape head result, if we choose to use this one
	tapeHead *interfaces.MoveGroupHistoryItem
	//The length of the match from tapehead to this new tapeHead; longer
	//matches are better.
	length int
}

//tapeLength returns the length between from and to, if they're in the same
//tape. If to cannot be reached from from, math.MaxInt64 is returned.
func tapeLength(from, to *interfaces.MoveGroupHistoryItem) int {

	count := 0

	for from != to && from != nil {
		count++
		from = from.Rest
	}

	if from == nil {
		//Fell off end without a match
		return math.MaxInt64
	}

	return count

}

//ParallelCount is a version of Parallel, but where the target count is given
//by Count. The length argument to Count will be the number of Groups who are
//children. See moves/count package for many options for this.
func ParallelCount(count interfaces.ValidCounter, children ...interfaces.MoveProgressionGroup) interfaces.MoveProgressionGroup {
	return &parallelCount{
		children,
		count,
	}
}

type parallelCount struct {
	Children []interfaces.MoveProgressionGroup
	Count    interfaces.ValidCounter
}

func (p parallelCount) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range p.Children {
		result = append(result, group.MoveConfigs()...)
	}
	return result
}

func (p parallelCount) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {
	tapeHead := tape

	//Keep track of items that have matched, by index into self.
	matchedItems := make(map[int]bool, len(p.Children))

	//Continue until all items have been matched.
	for {

		if err := p.Count(len(matchedItems), len(p.Children)); err == nil {
			break
		}

		if tapeHead == nil {
			return nil, nil
		}

		//Keep track of all matches found this run through so we can pick the
		//longest one.
		var matches []*matchInfo

		for i, group := range p.Children {
			//Skip items that have already been matched
			if matchedItems[i] {
				continue
			}

			err, rest := group.Satisfied(tapeHead)

			if err != nil {
				//That one didn't work
				continue
			}

			//Found a potential match.
			match := &matchInfo{
				index:    i,
				tapeHead: rest,
				length:   tapeLength(tapeHead, rest),
			}

			matches = append(matches, match)

		}

		if len(matches) == 0 {
			//Didn't find any matches
			return errors.New("No more items match, but tape still left and count not yet reached."), nil
		}

		//Select the match to use, based on the length

		sort.Slice(matches, func(i, j int) bool {
			return matches[i].length > matches[j].length
		})

		selectedMatch := matches[0]

		//Mark that this item has been matched.
		matchedItems[selectedMatch.index] = true
		tapeHead = selectedMatch.tapeHead

	}

	return nil, tapeHead
}
