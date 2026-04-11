package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

// ShuffleDiscardIntoDraw is a FixUp move that automatically shuffles the
// discard pile back into the draw pile when the draw pile is empty. It works
// with gameStates that embed [behaviors.DrawDiscardPair], which satisfies the
// [interfaces.DrawDiscardProvider] interface.
//
// Usage is zero-config:
//
//	auto.MustConfig(new(moves.ShuffleDiscardIntoDraw))
//
//boardgame:codegen
type ShuffleDiscardIntoDraw struct {
	FixUp
}

func drawDiscardPair(state boardgame.ImmutableState) (*behaviors.DrawDiscardPair, error) {
	hasDD, ok := state.ImmutableGameState().(behaviors.HasDrawDiscardPair)
	if !ok {
		return nil, errors.New("gameState does not embed behaviors.DrawDiscardPair")
	}
	return hasDD.GetDrawDiscardPair(), nil
}

// Legal returns nil when the draw stack is empty and the discard stack has
// components to shuffle back in.
func (s *ShuffleDiscardIntoDraw) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := s.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	dd, err := drawDiscardPair(state)
	if err != nil {
		return err
	}

	if !dd.NeedsReshuffle() {
		return errors.New("draw stack is not empty or discard stack has no components")
	}

	return nil
}

// Apply moves all components from the discard stack to the draw stack, then
// shuffles the draw stack.
func (s *ShuffleDiscardIntoDraw) Apply(state boardgame.State) error {
	dd, err := drawDiscardPair(state)
	if err != nil {
		return err
	}

	if err := dd.DiscardStack().MoveAllTo(dd.DrawStack()); err != nil {
		return errors.New("couldn't move discard to draw: " + err.Error())
	}

	if err := dd.DrawStack().Shuffle(); err != nil {
		return errors.New("couldn't shuffle draw stack: " + err.Error())
	}

	return nil
}

// ValidConfiguration verifies that the gameState embeds DrawDiscardPair.
func (s *ShuffleDiscardIntoDraw) ValidConfiguration(exampleState boardgame.State) error {
	if _, err := drawDiscardPair(exampleState); err != nil {
		return err
	}
	return s.FixUp.ValidConfiguration(exampleState)
}

// FallbackName returns a descriptive name for the move.
func (s *ShuffleDiscardIntoDraw) FallbackName(m *boardgame.GameManager) string {
	return "Shuffle Discard Into Draw"
}

// FallbackHelpText returns a description of what the move does.
func (s *ShuffleDiscardIntoDraw) FallbackHelpText() string {
	return "When the draw pile is empty, shuffles the discard pile back in."
}
