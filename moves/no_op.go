package moves

import (
	"github.com/jkomoros/boardgame"
)

/*
NoOp is a simple FixUp move that has a defined Apply() method that does
nothing, and whose Legal() is equivalent to base.Legal(). This move is
generally useful for MoveProgressionGroup matching scenarios where a move
that AllowMultipleInProgression() would match too many of itself in a row
greedily, even though the move was in two adjacent groups. A simple example:

	//...
	moves.AddOrderedForPhase(PhaseNormal,
		Serial(
			auto.MustConfig(new(FixUpAllowsMultiple)),
		),
		Serial(
			auto.MustConfig(new(FixUpAllowsMultiple)),
			auto.Mustconfig(new(FixUpAnotherMove)),
		),
	),
	//...

FixUpAllowsMultiple would match all historical moves in the first Serial
group, meaning that the second Serial group could never be matched. moves.NoOp
provides a barrier:

	//...
	moves.AddOrderedForPhase(PhaseNormal,
		Serial(
			auto.MustConfig(new(FixUpAllowsMultiple)),
		),
		Serial(
			auto.MustConfig(new(moves.NoOp)),
			auto.MustConfig(new(FixUpAllowsMultiple)),
			auto.Mustconfig(new(FixUpAnotherMove)),
		),
	),
	//...

FixUpAllowsMultiple will at some point decide itself that it is no longer
legal to apply. At that point, the next move in the progression is NoOp, which
only uses moves.Default.Legal() and is therefore always legal at its point in the
phase. After NoOp applies, the next FixUpAllowsMultiple applies, guaranteeing
it begins matching the second group.

NoOp is also used by AddOrderedForPhase to signal that the lack of a
StartPhase move was intentional.

Note that because every move config installed must have a different name, if
you use multiple moves.NoOp in your game, you will want to override at least
one of their names with WithMoveName or WithMoveNameSuffix.

boardgame:codegen
*/
type NoOp struct {
	FixUp
}

//Apply is a no-op; it makes no change to state and alwasy returns nil.
func (n *NoOp) Apply(state boardgame.State) error {
	return nil
}

//FallbackName returns "No Op"
func (n *NoOp) FallbackName(m *boardgame.GameManager) string {
	return "No Op"
}

//FallbackHelpText returns "A move that does nothing and is primarily used in
//specific move progression situations."
func (n *NoOp) FallbackHelpText() string {
	return "A move that does nothing and is primarily used in specific move progression situations."
}

type isNoOper interface {
	isNoOp() bool
}

//Quick way to check for a no op move
func (n *NoOp) isNoOp() bool {
	return true
}

/*
Done is a simple move that does nothing and whose Legal is equivalent to
moves.Default.Legal(), meaning it is legal purely if the phase matches and it's
the right time in the progression.

It's a non-fix-up equivalent to NoOp. It's used primarily when a player could
decide to make multiple moves, or end early, and has opted to end early.

boardgame:codegen
*/
type Done struct {
	Default
}

//Apply is a no-op; it makes no change to state and alwasy returns nil.
func (d *Done) Apply(state boardgame.State) error {
	return nil
}

//FallbackName returns "Done"
func (d *Done) FallbackName(m *boardgame.GameManager) string {
	return "Done"
}

//FallbackHelpText returns "The player has signaled that they are done
//applying moves in this group and are ready to move on."
func (d *Done) FallbackHelpText() string {
	return "The player has signaled that they are done applying moves in this group and are ready to move on."
}
