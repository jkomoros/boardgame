package behaviors

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jkomoros/boardgame"
)

/*
DrawDiscardPair tracks a draw stack and a discard stack on a gameState. It is a
[Connectable], [boardgame.TagConfigurable] behavior that is automatically
connected by the framework.

Configure it with struct tags pointing to existing Stack fields on the same
struct:

	type gameState struct {
	    base.SubState
	    behaviors.DrawDiscardPair `draw:"DrawStack" discard:"DiscardStack"`
	    DrawStack    boardgame.Stack `stack:"cards" sanitize:"len"`
	    DiscardStack boardgame.Stack `stack:"cards"`
	}

The companion move [moves.ShuffleDiscardIntoDraw] works with this behavior to
automatically reshuffle the discard pile into the draw pile when the draw pile
empties:

	auto.MustConfig(new(moves.ShuffleDiscardIntoDraw))

For multiple pairs, use direct named behavior fields and select each one from
its companion move. The field option also derives a distinct move name:

	type gameState struct {
	    base.SubState
	    TrainCards behaviors.DrawDiscardPair `draw:"TrainDraw" discard:"TrainDiscard"`
	    TicketCards behaviors.DrawDiscardPair `draw:"TicketDraw" discard:"TicketDiscard"`
	    // stack fields omitted
	}

	auto.MustConfig(new(moves.ShuffleDiscardIntoDraw), moves.WithDrawDiscardPairField("TrainCards"))
	auto.MustConfig(new(moves.ShuffleDiscardIntoDraw), moves.WithDrawDiscardPairField("TicketCards"))

You can also skip the tags and wire everything in FinishStateSetUp:

	func (g *gameState) FinishStateSetUp() {
	    g.DrawDiscardPair.ConnectDrawStack(g.DrawStack)
	    g.DrawDiscardPair.ConnectDiscardStack(g.DiscardStack)
	}
*/
type DrawDiscardPair struct {
	container    boardgame.SubState
	drawStack    boardgame.Stack
	discardStack boardgame.Stack
}

const drawStructTag = "draw"
const discardStructTag = "discard"

// ConnectBehavior stores a reference to the containing SubState.
func (d *DrawDiscardPair) ConnectBehavior(containingSubState boardgame.SubState) {
	d.container = containingSubState
}

// ConfigureFromTags reads the "draw" and "discard" struct tags from the
// embedding site and connects the corresponding Stack fields.
func (d *DrawDiscardPair) ConfigureFromTags(tags reflect.StructTag, containingSubState boardgame.SubState) error {
	if name := tags.Get(drawStructTag); name != "" {
		stack, err := lookupStackField(containingSubState, name)
		if err != nil {
			return fmt.Errorf("DrawDiscardPair: draw tag: %w", err)
		}
		d.drawStack = stack
	}
	if name := tags.Get(discardStructTag); name != "" {
		stack, err := lookupStackField(containingSubState, name)
		if err != nil {
			return fmt.Errorf("DrawDiscardPair: discard tag: %w", err)
		}
		d.discardStack = stack
	}
	return nil
}

// ConnectDrawStack manually sets the draw stack. Use this in FinishStateSetUp
// if you prefer imperative configuration over struct tags.
func (d *DrawDiscardPair) ConnectDrawStack(stack boardgame.Stack) {
	d.drawStack = stack
}

// ConnectDiscardStack manually sets the discard stack.
func (d *DrawDiscardPair) ConnectDiscardStack(stack boardgame.Stack) {
	d.discardStack = stack
}

// ValidConfiguration returns an error if the behavior hasn't been properly
// connected.
func (d *DrawDiscardPair) ValidConfiguration(example boardgame.State) error {
	if d.container == nil {
		return errors.New("DrawDiscardPair: ConnectBehavior hasn't been called. See the behaviors package doc for more on initializing Connectable behaviors")
	}
	if d.drawStack == nil {
		return errors.New("DrawDiscardPair: draw stack not connected. Use the `draw` struct tag or call ConnectDrawStack in FinishStateSetUp")
	}
	if d.discardStack == nil {
		return errors.New("DrawDiscardPair: discard stack not connected. Use the `discard` struct tag or call ConnectDiscardStack in FinishStateSetUp")
	}
	return validateAttachedStackPair(example, d.container, "DrawDiscardPair", "draw", d.drawStack, "discard", d.discardStack)
}

// DrawStack returns the connected draw stack.
func (d *DrawDiscardPair) DrawStack() boardgame.Stack {
	return d.drawStack
}

// DiscardStack returns the connected discard stack.
func (d *DrawDiscardPair) DiscardStack() boardgame.Stack {
	return d.discardStack
}

// DrawIsEmpty returns true if the draw stack has no components.
func (d *DrawDiscardPair) DrawIsEmpty() bool {
	return d.drawStack != nil && d.drawStack.NumComponents() == 0
}

// NeedsReshuffle returns true if the draw stack is empty and the discard stack
// has components to shuffle back in.
func (d *DrawDiscardPair) NeedsReshuffle() bool {
	return d.DrawIsEmpty() && d.discardStack != nil && d.discardStack.NumComponents() > 0
}

// GetDrawDiscardPair returns itself, satisfying [HasDrawDiscardPair].
func (d *DrawDiscardPair) GetDrawDiscardPair() *DrawDiscardPair {
	return d
}

// HasDrawDiscardPair is implemented by any SubState that embeds a
// DrawDiscardPair. It allows companion moves like [moves.ShuffleDiscardIntoDraw]
// to find the behavior via type assertion.
type HasDrawDiscardPair interface {
	GetDrawDiscardPair() *DrawDiscardPair
}
