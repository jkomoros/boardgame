package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// TokenAdvancer is required for moves that embed AdvanceToken. It provides the
// LocationBehavior to advance and the logic for computing the next position.
// This interface is defined here (rather than in moves/interfaces) to avoid an
// import cycle with the behaviors package.
type TokenAdvancer interface {
	AdvancableLocation(state boardgame.ImmutableState) *behaviors.LocationBehavior
	NextAdvanceIndex(state boardgame.ImmutableState, currentIndex enum.ImmutableVal) enum.EnumKey
}

/*
AdvanceToken is a framework-provided FixUp for deterministic NPC or token
movement. The embedding move must implement TokenAdvancer to provide the
LocationBehavior and the logic for computing the next position.

Optionally, the embedding move may implement:
  - interfaces.AdvanceCondition: gate whether advancement should happen
  - interfaces.PostAdvanceHandler: run side effects after advancement

boardgame:codegen
*/
type AdvanceToken struct {
	FixUp
}

// Legal checks the base FixUp legality and then the optional AdvanceCondition.
func (a *AdvanceToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := a.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	if condition, ok := a.Info().ConcreteMove().(interfaces.AdvanceCondition); ok {
		if err := condition.ShouldAdvance(state); err != nil {
			return err
		}
	}

	_, _, _, err := a.advanceDestination(state)
	return err
}

func (a *AdvanceToken) advanceDestination(state boardgame.ImmutableState) (*behaviors.LocationBehavior, enum.ImmutableVal, enum.ImmutableVal, error) {
	advancer, ok := a.Info().ConcreteMove().(TokenAdvancer)
	if !ok {
		return nil, nil, nil, errors.New("AdvanceToken: embedding move must implement TokenAdvancer")
	}

	behavior := advancer.AdvancableLocation(state)
	if behavior == nil {
		return nil, nil, nil, errors.New("AdvanceToken: AdvancableLocation returned nil")
	}

	locationEnum := behavior.LocationEnum()
	if locationEnum == nil {
		return nil, nil, nil, errors.New("AdvanceToken: LocationBehavior has no graph connected")
	}

	currentVal := behavior.LocationIndex()
	if currentVal == nil {
		return nil, nil, nil, errors.New("AdvanceToken: no component found in location stack")
	}

	nextIndex := advancer.NextAdvanceIndex(state, currentVal)
	nextVal, err := locationEnum.NewImmutableVal(nextIndex)
	if err != nil {
		return nil, nil, nil, errors.New("AdvanceToken: invalid next index: " + err.Error())
	}
	if err := behavior.MayMoveTo(nextIndex.Int()); err != nil {
		return nil, nil, nil, errors.New("AdvanceToken: cannot move to next index: " + err.Error())
	}

	return behavior, currentVal, nextVal, nil
}

// Apply advances the token using the TokenAdvancer and calls PostAdvanceHandler
// if implemented.
func (a *AdvanceToken) Apply(state boardgame.State) error {

	behavior, currentVal, nextVal, err := a.advanceDestination(state)
	if err != nil {
		return err
	}

	if err := behavior.MoveTo(nextVal.Value().Int()); err != nil {
		return err
	}

	if handler, ok := a.Info().ConcreteMove().(interfaces.PostAdvanceHandler); ok {
		if err := handler.AfterAdvance(state, currentVal, nextVal); err != nil {
			return err
		}
	}

	return nil
}

// FallbackName returns "Advance Token"
func (a *AdvanceToken) FallbackName(m *boardgame.GameManager) string {
	return "Advance Token"
}

// FallbackHelpText returns a description of the FixUp.
func (a *AdvanceToken) FallbackHelpText() string {
	return "Advance a token to its next position deterministically."
}
