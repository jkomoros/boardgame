package boardgame

import (
	"strconv"

	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
)

/*
This file holds the small set of helpers extracted from moves.Default's
frozen imperative Legal() chain so that BOTH that chain (moves/default.go)
and the "inPhase"/"stackConstraints" wrapper predicates in package legal
(legal/catalog_framework.go) can call exactly one implementation — see the
design spec §2/§3 and the Task 7 brief's "extract-and-share" instruction.
The extraction lives in core (this package) rather than in package moves
because package legal cannot import package moves (moves already imports
legal; see the design spec §3's layering diagram), so core is the only
package both moves and legal can share code through.

The THIRD frozen-chain check, legalMoveInProgression (move-tape matching),
is deliberately NOT extracted here. Its "inProgression" wrapper predicate
lives in package moves, not package legal, because it fundamentally needs
moves.MoveProgressionGroup (the interface driving Satisfied/tape matching),
which lives in package moves and cannot be referenced from core or legal
without moving that type — out of scope for this task. Since both the
frozen chain's legalMoveInProgression and the inProgression predicate live
in the SAME package (moves), they share moves/default.go's existing
historicalMovesSincePhaseTransition/matchTape functions directly, with no
core detour needed. See moves/catalog_framework.go's doc comment for the
full rationale, and the Task 7 report's "layering decision" section.

Every helper here preserves its moves/default.go original's behavior and
error strings byte-for-byte: this is literally the same code, relocated and
parameterized instead of reading straight from a *moves.Default's config
bag. moves/default.go's legalInPhase/legalStackConstraints methods now call
these directly.
*/

// legalCurrentPhaseInfo extracts both the ImmutableVal and EnumKey from the
// delegate's CurrentPhase. If CurrentPhase returns nil (no phase configured),
// returns (nil, 0). This is core's copy of moves/default.go's unexported
// currentPhaseInfo, duplicated (not shared) because moves/default.go's own
// copy is also used by legalMoveInProgression, which stays entirely in
// package moves (see this file's doc comment). Both copies are one-line
// wrappers around GameDelegate.CurrentPhase and are trivially kept in sync.
func legalCurrentPhaseInfo(state ImmutableState) (enum.ImmutableVal, enum.EnumKey) {
	val := state.Manager().Delegate().CurrentPhase(state)
	if val == nil {
		return nil, 0
	}
	return val, val.Value()
}

// LegalInPhaseCheck reports whether state's current phase (per the game's
// delegate) is one of legalPhases, walking TreeEnum ancestors exactly like
// moves.Default.Legal() always has. A zero-length legalPhases is legal in
// every phase. This is the byte-for-byte extraction of
// moves/default.go's legalInPhase method (see that file's Legal doc comment
// for the full behavioral description); moves.Default.legalInPhase now
// calls this directly, and legal's "inPhase" wrapper predicate
// (legal/catalog_framework.go) calls it too, so the two observably agree by
// construction, not by convention.
func LegalInPhaseCheck(state ImmutableState, legalPhases []enum.EnumKey) error {

	if len(legalPhases) == 0 {
		//If PhaseEnum is a TreeEnum, this is basically equivalent to the
		//legalPhases being []int{0}.
		return nil
	}

	currentPhaseVal, currentPhase := legalCurrentPhaseInfo(state)

	var treeEnum enum.TreeEnum
	if currentPhaseVal != nil {
		if e := currentPhaseVal.Enum(); e != nil {
			treeEnum = e.TreeEnum()
		}
	}

	//totalCurrentPhases is all of the current phases we could be considered
	//to be in. Defaults to an []EnumKey with just the current phase.
	totalCurrentPhases := []enum.EnumKey{currentPhase}

	if treeEnum != nil {
		//If PhaseEnum is a tree, then the phase we're in for this purpose is
		//all ancestor phases.
		totalCurrentPhases = treeEnum.Ancestors(currentPhase)
	}

	for _, phase := range legalPhases {
		for _, candidateCurrentPhase := range totalCurrentPhases {
			if phase == candidateCurrentPhase {
				return nil
			}
		}
	}

	phaseName := strconv.Itoa(currentPhase.Int())

	if currentPhaseVal != nil {
		phaseName = currentPhaseVal.String()
	}

	return errors.New("Move is not legal in phase " + phaseName)
}

// LegalStackConstraintsCheck reports whether the first component of the
// srcName stack (read from state's GameState) would be accepted by the
// dstName stack (per ImmutableComponentInstance.MayMoveTo), giving early
// feedback at Legal() time. Either stack property missing, unreadable, or
// empty is treated as "nothing to check" (nil error) — this is the
// byte-for-byte extraction of moves/default.go's legalStackConstraints
// method; moves.Default.legalStackConstraints now calls this directly, and
// legal's "stackConstraints" wrapper predicate (legal/catalog_framework.go)
// calls it too.
func LegalStackConstraintsCheck(state ImmutableState, srcName, dstName string) error {

	reader := state.ImmutableGameState().Reader()

	srcStack, err := reader.ImmutableStackProp(srcName)
	if err != nil || srcStack == nil {
		return nil
	}
	dstStack, err := reader.ImmutableStackProp(dstName)
	if err != nil || dstStack == nil {
		return nil
	}

	first := srcStack.ImmutableFirst()
	if first == nil {
		return nil
	}

	return first.MayMoveTo(dstStack)
}
