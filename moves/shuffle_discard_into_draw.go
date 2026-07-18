package moves

import (
	"errors"
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

// ShuffleDiscardIntoDraw is a FixUp move that automatically shuffles the
// discard pile back into the draw pile when the draw pile is empty. It works
// with gameStates that embed [behaviors.DrawDiscardPair], which satisfies the
// [behaviors.HasDrawDiscardPair] interface.
//
// Usage is zero-config:
//
//	auto.MustConfig(new(moves.ShuffleDiscardIntoDraw))
//
//boardgame:codegen
type ShuffleDiscardIntoDraw struct {
	FixUp
}

const configPropDrawDiscardPairField = fullyQualifiedPackageName + "DrawDiscardPairField"

// WithDrawDiscardPairField selects a named, direct-value DrawDiscardPair field
// on gameState. Omit it for the common single anonymously embedded pair.
func WithDrawDiscardPairField(fieldName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropDrawDiscardPairField] = fieldName
	}
}

func (s *ShuffleDiscardIntoDraw) drawDiscardPair(state boardgame.ImmutableState) (*behaviors.DrawDiscardPair, error) {
	fieldName, configured, err := configuredString(s.CustomConfiguration(), configPropDrawDiscardPairField, "WithDrawDiscardPairField")
	if err != nil {
		return nil, err
	}
	if configured {
		value, err := namedBehaviorField(state, fieldName, new(behaviors.DrawDiscardPair))
		if err != nil {
			return nil, err
		}
		return value.(*behaviors.DrawDiscardPair), nil
	}
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

	dd, err := s.drawDiscardPair(state)
	if err != nil {
		return err
	}

	if !dd.NeedsReshuffle() {
		return errors.New("draw stack is not empty or discard stack has no components")
	}
	if err := dd.DiscardStack().MayMoveAllTo(dd.DrawStack()); err != nil {
		return fmt.Errorf("discard cannot be moved into draw stack: %w", err)
	}
	return nil
}

// Apply moves all components from the discard stack to the draw stack, then
// shuffles the draw stack.
func (s *ShuffleDiscardIntoDraw) Apply(state boardgame.State) error {
	dd, err := s.drawDiscardPair(state)
	if err != nil {
		return err
	}

	if err := dd.DiscardStack().MoveAllTo(dd.DrawStack()); err != nil {
		return fmt.Errorf("couldn't move discard to draw: %w", err)
	}

	if err := dd.DrawStack().Shuffle(); err != nil {
		return fmt.Errorf("couldn't shuffle draw stack: %w", err)
	}

	return nil
}

// ValidConfiguration verifies that the gameState embeds DrawDiscardPair.
func (s *ShuffleDiscardIntoDraw) ValidConfiguration(exampleState boardgame.State) error {
	pair, err := s.drawDiscardPair(exampleState)
	if err != nil {
		return err
	}
	if pair == nil {
		return errors.New("draw/discard provider returned nil")
	}
	if err := pair.ValidConfiguration(exampleState); err != nil {
		return fmt.Errorf("ShuffleDiscardIntoDraw: %w", err)
	}
	return s.FixUp.ValidConfiguration(exampleState)
}

// FallbackName returns a descriptive name for the move.
func (s *ShuffleDiscardIntoDraw) FallbackName(m *boardgame.GameManager) string {
	if field, configured, err := configuredString(s.CustomConfiguration(), configPropDrawDiscardPairField, "WithDrawDiscardPairField"); configured && err == nil {
		return "Shuffle " + titleCaseToWords(field) + " Discard Into Draw"
	}
	return "Shuffle Discard Into Draw"
}

// FallbackHelpText returns a description of what the move does.
func (s *ShuffleDiscardIntoDraw) FallbackHelpText() string {
	return "When the draw pile is empty, shuffles the discard pile back in."
}
