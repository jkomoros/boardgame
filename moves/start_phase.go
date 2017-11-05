package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
	"strconv"
)

//phaseToEnterer should be implemented by moves that embed moves.StartPhase to
//configure which phase to enter. It's a private interface because StartPhase
//already has a base PhaseToEnter, and to keep the number of interfaces
//smaller.
type phaseToEnterer interface {
	PhaseToEnter(currentPhase int) int
}

//StartPhase is a simple move, often used in game SetUp phases, to advance to
//the next phase, as returned by the embedding move's PhaseToEnter(). If
//BeforeLeavePhase or BeforeEnterPhase are defined they will be called at the
//appropriate time. In many cases you don't even need to define your own
//struct, but can just get a MoveTypeConfig by calling
//NewStartPhaseMoveConfig.
//
//+autoreader
type StartPhase struct {
	Base
	phaseToStart int
}

func (s *StartPhase) ValidConfiguration(exampleState boardgame.MutableState) error {
	embeddingMove := s.TopLevelStruct()

	if _, ok := embeddingMove.(phaseToEnterer); !ok {
		return errors.New("The embedding move does not implement PhaseToEnterer")
	}

	if _, ok := exampleState.GameState().(moveinterfaces.CurrentPhaseSetter); !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	return nil
}

//PhaseToEnter uses the Phase provided via StartPhaseMoveConfig constructor
//(or 0 if NewStartPhaseMoveConfig wasn't used). If you want a different
//behavior, override PhaseToEnter in your embedding move.
func (s *StartPhase) PhaseToEnter(currentPhase int) int {
	return s.phaseToStart
}

//Apply call BeforeLeavePhase() (if it exists), then BeforeEnterPhase() (if it
//exists),then SetCurrentPhase to the phase index returned by PhaseToEnter
//from this move type.
func (s *StartPhase) Apply(state boardgame.MutableState) error {

	phaseEnterer, ok := s.TopLevelStruct().(phaseToEnterer)

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

//NewStartPhaseMoveConfig returns a MoveConfig object configured so that you
//don't need to even define your own StartPhase embedding move but can just
//use this one directly.
func NewStartPhaseMoveConfig(manager *boardgame.GameManager, legalPhases []int, phaseToStart int) *boardgame.MoveTypeConfig {

	phaseEnum := manager.Delegate().PhaseEnum()

	phaseToStartName := strconv.Itoa(phaseToStart)

	if phaseEnum != nil {
		phaseToStartName = phaseEnum.String(phaseToStart)
	}

	return &boardgame.MoveTypeConfig{
		Name:     "Start Phase " + phaseToStartName,
		HelpText: "Enters phase " + phaseToStartName,
		MoveConstructor: func() boardgame.Move {
			result := new(StartPhase)
			result.phaseToStart = phaseToStart
			return result
		},
		IsFixUp:     true,
		LegalPhases: legalPhases,
	}

}
