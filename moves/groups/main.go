/*

	groups is a package of various objects that implement
	moves/interfaces.MoveProgressionGroup, and are thus appropriate for
	passing to moves.AddOrderedForPhase.

*/
package groups

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/count"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"math"
	"sort"
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
//ParallelCount. The base Parallel is equivalent to ParallelCount with a Count
//of count.All().
type Parallel []interfaces.MoveProgressionGroup

func (p Parallel) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range p {
		result = append(result, group.MoveConfigs()...)
	}
	return result
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

//Satisfied goes through each item in turn, seeing if any of them can consume
//items off of the front of the tape without erroring. It continues going
//through until all are met, or no more un-triggered items can consume
//another. If at any point more than one item could match at the given point
//in the tape, it chooses the match that consumes the most tape.
func (p Parallel) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {
	return parallelSatisfiedHelper(p, count.All(), tape)
}

//ParallelCount is a version of Parallel, but where the target count is given
//by Count. The length argument to Count will be the number of Groups who are
//children. See moves/count package for many options for this.
type ParallelCount struct {
	Children []interfaces.MoveProgressionGroup
	Count    interfaces.ValidCounter
}

func (p ParallelCount) MoveConfigs() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, group := range p.Children {
		result = append(result, group.MoveConfigs()...)
	}
	return result
}

func (p ParallelCount) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {
	return parallelSatisfiedHelper(p.Children, p.Count, tape)
}

func parallelSatisfiedHelper(children []interfaces.MoveProgressionGroup, counter interfaces.ValidCounter, tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {
	tapeHead := tape

	//Keep track of items that have matched, by index into self.
	matchedItems := make(map[int]bool, len(children))

	//Continue until all items have been matched.
	for {

		if err := counter(len(matchedItems), len(children)); err == nil {
			break
		}

		if tapeHead == nil {
			return nil, nil
		}

		//Keep track of all matches found this run through so we can pick the
		//longest one.
		var matches []*matchInfo

		for i, group := range children {
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
			return errors.New("No more items match, but tape still left."), nil
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
