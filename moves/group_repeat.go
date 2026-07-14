package moves

import (
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// Optional returns a MoveProgressionGroup that matches the provided group
// either 0 or 1 times. Equivalent to Repeat() with a count of Between(0, 1).
func Optional(group MoveProgressionGroup) MoveProgressionGroup {
	return Repeat(CountBetween(0, 1), group)
}

// Repeat returns a MoveProgressionGroup that repeats the provided group the
// number of times count is looking for, in serial. Assumes that the
// ValidCounter has a single range of legal count values, where before it they
// are illegal, during the range they are legal, and after it they are illegal
// agin, and will read as many times from the tape as it can within that legal
// range. All ValidCounter methods in this package satisfy this. It is
// conceptually equivalent to duplicating a given group within a parent Serial
// count times.
func Repeat(count ValidCounter, group MoveProgressionGroup) MoveProgressionGroup {
	return repeat{
		count,
		group,
	}
}

type repeat struct {
	Count ValidCounter
	Child MoveProgressionGroup
}

func (r repeat) MoveConfigs() []boardgame.MoveConfig {
	return r.Child.MoveConfigs()
}

func (r repeat) Satisfied(tape *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
	return repeatSatisfied(r.Count, r.Child, tape, func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
		return group.Satisfied(tapeHead)
	})
}

// SatisfiedWithContext implements StatefulMoveProgressionGroup: identical to
// Satisfied, except the child is evaluated via satisfiedDispatch(r.Child,
// tapeHead, ctx) instead of r.Child.Satisfied(tapeHead) directly, so a
// state-driven child (e.g. RepeatFromProp) nested inside this Repeat still
// receives ctx. See groups.go's StatefulMoveProgressionGroup doc comment.
func (r repeat) SatisfiedWithContext(tape *MoveGroupHistoryItem, ctx legal.Context) (*MoveGroupHistoryItem, error) {
	return repeatSatisfied(r.Count, r.Child, tape, func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
		return satisfiedDispatch(group, tapeHead, ctx)
	})
}

// repeatSatisfied is the shared walk both repeat.Satisfied and
// repeat.SatisfiedWithContext (and repeatFromProp.SatisfiedWithContext,
// which resolves its dynamic count and then delegates here) use; evalChild
// is how child is asked whether it's satisfied at a given tape position
// (context-free for Satisfied, context-aware for SatisfiedWithContext).
func repeatSatisfied(count ValidCounter, child MoveProgressionGroup, tape *MoveGroupHistoryItem, evalChild func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error)) (*MoveGroupHistoryItem, error) {

	tapeHead := tape

	//we assume that there is precisely one continguous bound that is legal.
	//We want to go up until we enter the lower bound, then any error we run
	//into within that bound is OK (just return last known good tape position
	//and ignore the group that errored), and then when we reach the upper
	//limit we end.
	lowerBoundReached := false

	//Check if we start within the lower bound (for example, a count.AtMost()
	//will start within the legal lower bound.z)
	if err := count(0, 1); err == nil {
		lowerBoundReached = true
	}

	//The count happens after the group has been consumed each time, so by the
	//time we look at this the first time it will have already been one group.
	iteration := 1

	for {

		//If we ever reach the tape end without having found an error then it's
		//legal.
		if tapeHead == nil {
			return nil, nil
		}

		rest, err := evalChild(child, tapeHead)

		if err != nil {
			if lowerBoundReached {
				//We're between the lower and upper bound of legal counts, so
				//errors are not a big deal, just return the last known good
				//state.
				return tapeHead, nil
			}
			//Otherwise, we haven't yet gotten the smallest legal amount so we
			//should stop.
			return nil, err
		}

		boundErr := count(iteration, 1)

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

		iteration++
		tapeHead = rest

	}

	return tapeHead, nil

}

// RepeatFromProp returns a MoveProgressionGroup that repeats group a number
// of times resolved DYNAMICALLY from live state at match time (design spec
// §7, #644), rather than a count baked in at config time like
// Repeat(CountExactly(n), group). path is a legal path-grammar string (see
// legal.Context.ResolvePath — typically "game.SomeIntProp") naming an int
// property whose value is the target repeat count for THIS evaluation.
//
// Because the count is state-dependent, resolving it requires a
// legal.Context — RepeatFromProp implements StatefulMoveProgressionGroup
// (see groups.go), and matchTape (default.go) always builds and passes one.
// Its plain Satisfied (no context) is a backward-compatibility stub: it
// cannot resolve path without state, so it fails closed with a descriptive
// error naming path, rather than silently treating the count as 0 — a
// caller that invokes Satisfied directly (bypassing matchTape) gets a loud
// error, not a wrong answer.
func RepeatFromProp(path string, group MoveProgressionGroup) MoveProgressionGroup {
	return repeatFromProp{path: path, child: group}
}

type repeatFromProp struct {
	path  string
	child MoveProgressionGroup
}

func (r repeatFromProp) MoveConfigs() []boardgame.MoveConfig {
	return r.child.MoveConfigs()
}

func (r repeatFromProp) Satisfied(tape *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
	return tape, fmt.Errorf("moves: RepeatFromProp(%q): count requires a legal.Context to resolve (Satisfied was called directly, not through matchTape/legalMoveInProgression)", r.path)
}

func (r repeatFromProp) SatisfiedWithContext(tape *MoveGroupHistoryItem, ctx legal.Context) (*MoveGroupHistoryItem, error) {
	n, err := resolveRepeatFromPropCount(r.path, ctx)
	if err != nil {
		return tape, fmt.Errorf("moves: RepeatFromProp(%q): %w", r.path, err)
	}
	return repeatSatisfied(CountExactly(n), r.child, tape, func(group MoveProgressionGroup, tapeHead *MoveGroupHistoryItem) (*MoveGroupHistoryItem, error) {
		return satisfiedDispatch(group, tapeHead, ctx)
	})
}

// resolveRepeatFromPropCount resolves path (a legal path-grammar string) as
// an int against ctx, via ctx.ResolvePath. This is moves' own small copy of
// the same resolve-an-int-path pattern package legal's catalog predicates
// use (legal/path_resolve.go's resolveIntPath) — that helper is unexported
// to package legal and not worth exporting for this one call site, so it's
// duplicated here rather than reached into.
func resolveRepeatFromPropCount(path string, ctx legal.Context) (int, error) {
	val, propType, err := ctx.ResolvePath(legal.PropPath(path))
	if err != nil {
		return 0, err
	}
	if propType != boardgame.TypeInt {
		return 0, fmt.Errorf("path %q is not an int property", path)
	}
	i, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("path %q resolved to an int-typed property but its value was not an int (%T)", path, val)
	}
	return i, nil
}
