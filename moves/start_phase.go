package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

//StartPhase is a simple move, often used in game SetUp phases, to advance to
//the next phase, as returned by the embedding move's PhaseToEnter(). If
//BeforeLeavePhase or BeforeEnterPhase are defined they will be called at the
//appropriate time.
type StartPhase struct {
	Base
}

func (s *StartPhase) ValidConfiguration(exampleState boardgame.MutableState) error {
	embeddingMove := s.TopLevelStruct()

	if _, ok := embeddingMove.(moveinterfaces.PhaseToEnterer); !ok {
		return errors.New("The embedding move does not implement PhaseToEnterer")
	}

	if _, ok := exampleState.GameState().(moveinterfaces.CurrentPhaseSetter); !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	return nil
}

func (s *StartPhase) Apply(state boardgame.MutableState) error {
	embeddingMove := s.Info().Type().NewMove(state)

	phaseEnterer, ok := embeddingMove.(moveinterfaces.PhaseToEnterer)

	if !ok {
		return errors.New("The embedding move does not implement PhaseToEnterer")
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	phaseToEnter := phaseEnterer.PhaseToEnter(currentPhase)

	phaseSetter, ok := state.GameState().(moveinterfaces.CurrentPhaseSetter)

	if !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	beforeLeaver, ok := state.GameState().(moveinterfaces.BeforeLeavePhaser)

	if ok {
		if err := beforeLeaver.BeforeLeavePhase(currentPhase, state); err != nil {
			return errors.New("Before Leave Phase errored: " + err.Error())
		}
	}

	beforeEnterer, ok := state.GameState().(moveinterfaces.BeforeEnterPhaser)

	if ok {
		if err := beforeEnterer.BeforeEnterPhase(phaseToEnter, state); err != nil {
			return errors.New("Before Enter Phase errored: " + err.Error())
		}
	}

	phaseSetter.SetCurrentPhase(phaseToEnter)

	return nil
}
