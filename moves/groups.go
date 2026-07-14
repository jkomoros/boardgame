package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// MoveGroupHistoryItem is a singly-linked list (referred to in various
// comments as a "tape") that is passed to MoveProgressionGroup.Satisfied(). It
// represents a list of all of the moves that have applied so far since
// game.CurrentPhase() last changed.
type MoveGroupHistoryItem struct {
	MoveName string
	Rest     *MoveGroupHistoryItem
}

// MoveProgressionGroup is an object that can be used to define a valid move
// progression. moves.AutoConfigurer().Config() returns objects that fit this
// interface.
type MoveProgressionGroup interface {
	//MoveConfigs should return the full enumeration of contained MoveConfigs
	//within this Group, from left to right and top to bottom. This is used by
	//moves.AddOrderedForPhase to know which MoveConfigs contained within it
	//to install.
	MoveConfigs() []boardgame.MoveConfig

	//Satisfied reads the tape and returns an error if the sequence was not
	//valid (did not match, for example, or the group was configured in an
	//invalid way in general). If it returns a nil error, it should also
	//return the rest of the tape representing the items it did not yet
	//consume. If passed a nil tape, it should immediately return nil, nil. If
	//the top-level MoveProgressionGroup consumes the entire tape and doesn't
	//return an error then the progression is considered valid.
	Satisfied(tape *MoveGroupHistoryItem) (rest *MoveGroupHistoryItem, err error)
}

// StatefulMoveProgressionGroup is an OPTIONAL interface a MoveProgressionGroup
// may additionally implement to gain access to a legal.Context (state,
// proposer, move, chest) while matching the tape — design spec §7's "named
// plumbing change", added to unlock #644 (RepeatFromProp: a repeat count
// resolved dynamically from live state, rather than baked in at config
// time).
//
// This is deliberately an ADDITIVE, optional interface rather than a
// breaking change to MoveProgressionGroup.Satisfied's signature: Go
// interfaces are structural, so widening Satisfied's own signature would
// break every existing implementation, in this package and in any
// out-of-tree consumer (the design spec's ../games repo is a real one —
// see Task 13 forward-context in the runbook). No third-party
// implementation of MoveProgressionGroup was found in this repo or in
// ../games as of this task (grep for "Satisfied(tape" outside moves/ and
// examples/ turns up nothing); this interface stays optional regardless,
// since "none found today" is not a guarantee for tomorrow.
//
// Every compositor group type in this package (serial, parallelCount,
// repeat, repeatFromProp) implements this, using satisfiedDispatch (below)
// to recurse into children — so a state-driven child (RepeatFromProp)
// nested anywhere inside a Serial/Parallel/Repeat tree still receives ctx,
// without every compositor needing its own bespoke know-your-children
// logic. A plain MoveProgressionGroup that does NOT implement this
// interface (any leaf move config, or a hypothetical third-party group)
// keeps working exactly as before: satisfiedDispatch falls back to its
// ordinary, context-free Satisfied.
type StatefulMoveProgressionGroup interface {
	MoveProgressionGroup
	// SatisfiedWithContext is exactly like Satisfied, but additionally
	// receives ctx, the legal.Context the tape is being matched against
	// (built by matchTape from the state/proposer moves.Default.Legal()'s
	// frozen chain and the "inProgression" wrapper predicate are both
	// evaluating against).
	SatisfiedWithContext(tape *MoveGroupHistoryItem, ctx legal.Context) (rest *MoveGroupHistoryItem, err error)
}

// satisfiedDispatch evaluates group against tape, preferring
// StatefulMoveProgressionGroup.SatisfiedWithContext (ctx-aware) when group
// implements it, and falling back to plain MoveProgressionGroup.Satisfied
// otherwise. Every compositor group type in this package uses this to
// recurse into its children, rather than calling child.Satisfied directly,
// which is what makes a state-driven child nested arbitrarily deep still
// receive ctx.
func satisfiedDispatch(group MoveProgressionGroup, tape *MoveGroupHistoryItem, ctx legal.Context) (*MoveGroupHistoryItem, error) {
	if stateful, ok := group.(StatefulMoveProgressionGroup); ok {
		return stateful.SatisfiedWithContext(tape, ctx)
	}
	return group.Satisfied(tape)
}
