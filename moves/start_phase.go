package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//CurrentPhaseSetter should be implemented by you gameState to set the
//CurrentPhase. Must be implemented if you use the StartPhase move type.
type CurrentPhaseSetter interface {
	SetCurrentPhase(int)
}

//PhaseToEnterer should be implemented by moves that embed moves.StartPhase to
//configure which phase to enter.
type PhaseToEnterer interface {
	PhaseToEnter(currentPhase int) int
}

//BeforeLeavePhaser is an interface to implement on GameState if you want to
//do some action on state before leaving the given phase.
type BeforeLeavePhaser interface {
	BeforeLeavePhase(phase int, state boardgame.MutableState) error
}

//BeforeEnterPhaser is an interface to implement on GameState if you want to
//do some action on state just before entering the givenn state.
type BeforeEnterPhaser interface {
	BeforeEnterPhase(phase int, state boardgame.MutableState) error
}

//StartPhase is a simple move, often used in game SetUp phases, to advance to
//the next phase, as returned by the embedding move's PhaseToEnter(). If
//BeforeLeavePhase or BeforeEnterPhase are defined they will be called at the
//appropriate time.
type StartPhase struct {
	Base
}

func (s *StartPhase) ValidConfiguration(exampleState boardgame.MutableState) error {
	embeddingMove := s.TopLevelStruct()

	if _, ok := embeddingMove.(PhaseToEnterer); !ok {
		return errors.New("The embedding move does not implement PhaseToEnterer")
	}

	if _, ok := exampleState.GameState().(CurrentPhaseSetter); !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	return nil
}

func (s *StartPhase) Apply(state boardgame.MutableState) error {
	embeddingMove := s.Info().Type().NewMove(state)

	phaseEnterer, ok := embeddingMove.(PhaseToEnterer)

	if !ok {
		return errors.New("The embedding move does not implement PhaseToEnterer")
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	phaseToEnter := phaseEnterer.PhaseToEnter(currentPhase)

	phaseSetter, ok := state.GameState().(CurrentPhaseSetter)

	if !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	beforeLeaver, ok := state.GameState().(BeforeLeavePhaser)

	if ok {
		if err := beforeLeaver.BeforeLeavePhase(currentPhase, state); err != nil {
			return errors.New("Before Leave Phase errored: " + err.Error())
		}
	}

	beforeEnterer, ok := state.GameState().(BeforeEnterPhaser)

	if ok {
		if err := beforeEnterer.BeforeEnterPhase(phaseToEnter, state); err != nil {
			return errors.New("Before Enter Phase errored: " + err.Error())
		}
	}

	phaseSetter.SetCurrentPhase(phaseToEnter)

	return nil
}
