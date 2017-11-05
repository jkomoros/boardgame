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

//StartPhase is a simple move, often used in game SetUp phases, to advance to
//the next phase, as returned by the embedding move's PhaseToEnter().
type StartPhase struct {
	Base
}

func (s *StartPhase) ValidConfiguration(exampleState boardgame.MutableState) error {
	embeddingMove := s.Info().Type().NewMove(exampleState)

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

	phaseSetter.SetCurrentPhase(phaseToEnter)

	return nil
}
