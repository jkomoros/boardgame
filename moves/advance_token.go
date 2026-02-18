package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//TokenAdvancer is required for moves that embed AdvanceToken. It provides the
//LocationBehavior to advance and the logic for computing the next position.
//This interface is defined here (rather than in moves/interfaces) to avoid an
//import cycle with the behaviors package.
type TokenAdvancer interface {
	AdvancableLocation(state boardgame.State) *behaviors.LocationBehavior
	NextAdvanceIndex(state boardgame.ImmutableState, currentIndex int) int
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

//Legal checks the base FixUp legality and then the optional AdvanceCondition.
func (a *AdvanceToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := a.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	if _, ok := a.TopLevelStruct().(TokenAdvancer); !ok {
		return errors.New("AdvanceToken: embedding move must implement TokenAdvancer")
	}

	if condition, ok := a.TopLevelStruct().(interfaces.AdvanceCondition); ok {
		if err := condition.ShouldAdvance(state); err != nil {
			return err
		}
	}

	return nil
}

//Apply advances the token using the TokenAdvancer and calls PostAdvanceHandler
//if implemented.
func (a *AdvanceToken) Apply(state boardgame.State) error {

	advancer, ok := a.TopLevelStruct().(TokenAdvancer)
	if !ok {
		return errors.New("AdvanceToken: embedding move must implement TokenAdvancer")
	}

	behavior := advancer.AdvancableLocation(state)
	if behavior == nil {
		return errors.New("AdvanceToken: AdvancableLocation returned nil")
	}

	currentIndex := behavior.LocationIndex()
	nextIndex := advancer.NextAdvanceIndex(state, currentIndex)

	if err := behavior.MoveTo(nextIndex); err != nil {
		return err
	}

	if handler, ok := a.TopLevelStruct().(interfaces.PostAdvanceHandler); ok {
		if err := handler.AfterAdvance(state, currentIndex, nextIndex); err != nil {
			return err
		}
	}

	return nil
}

//FallbackName returns "Advance Token"
func (a *AdvanceToken) FallbackName(m *boardgame.GameManager) string {
	return "Advance Token"
}

//FallbackHelpText returns a description of the FixUp.
func (a *AdvanceToken) FallbackHelpText() string {
	return "Advance a token to its next position deterministically."
}
