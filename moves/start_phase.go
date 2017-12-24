package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
	"strconv"
)

const startPhaseConfigName = "__moves.StartPhaseConfigProp"

//phaseToStarter should be implemented by moves that embed moves.StartPhase to
//configure which phase to enter. It's a private interface because StartPhase
//already has a base PhaseToStart, and to keep the number of interfaces
//smaller.
type phaseToStarter interface {
	PhaseToStart(currentPhase int) int
}

//StartPhase is a simple move, often used in game SetUp phases, to advance to
//the next phase, as returned by the embedding move's PhaseToStart(). If
//BeforeLeavePhase or BeforeEnterPhase are defined they will be called at the
//appropriate time. In many cases you don't even need to define your own
//struct, but can just get a MoveTypeConfig by calling NewStartPhaseConfig.
//+autoreader
type StartPhase struct {
	Base
}

func (s *StartPhase) ValidConfiguration(exampleState boardgame.MutableState) error {
	embeddingMove := s.TopLevelStruct()

	phaseStarter, ok := embeddingMove.(phaseToStarter)

	if !ok {
		return errors.New("The embedding move does not have PhaseToStart()")
	}

	if phaseStarter.PhaseToStart(exampleState.Game().Manager().Delegate().CurrentPhase(exampleState)) < 0 {
		return errors.New("Phase to start returned a negative value, which signals an error")
	}

	if _, ok := exampleState.GameState().(moveinterfaces.CurrentPhaseSetter); !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	return nil
}

//PhaseToStart uses the Phase provided via StartPhaseMoveConfig constructor
//(or 0 if NewStartPhaseConfig wasn't used). If you want a different behavior,
//override PhaseToStart in your embedding move.
func (s *StartPhase) PhaseToStart(currentPhase int) int {
	config := s.Info().Type().CustomConfiguration()
	val, ok := config[configNameStartPhase]
	if !ok {
		return -1
	}
	intVal, ok := val.(int)
	if !ok {
		return -1
	}
	return intVal
}

//Apply call BeforeLeavePhase() (if it exists), then BeforeEnterPhase() (if it
//exists),then SetCurrentPhase to the phase index returned by PhaseToStart
//from this move type.
func (s *StartPhase) Apply(state boardgame.MutableState) error {

	phaseEnterer, ok := s.TopLevelStruct().(phaseToStarter)

	if !ok {
		return errors.New("The embedding move does not have PhaseToStart()")
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	phaseToEnter := phaseEnterer.PhaseToStart(currentPhase)

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

//MoveTypeName returns a constant. For StartPhase it's better to use
//NewStartPhaseConfig instead.
func (s *StartPhase) MoveTypeFallbackName(manager *boardgame.GameManager) string {
	return "Start Phase"
}

//MoveTypeName returns a constant. For StartPhase it's better to use
//NewStartPhaseConfig instead.
func (s *StartPhase) MoveTypeFallbackHelpText(manager *boardgame.GameManager) string {
	return "Enters a phase"
}

func (s *StartPhase) MoveTypeIsFixUp(manager *boardgame.GameManager) bool {
	return true
}

//NewStartPhaseConfig returns a MoveConfig object configured so that you
//don't need to even define your own StartPhase embedding move but can just
//use this one directly.
func NewStartPhaseConfig(manager *boardgame.GameManager, phaseToStart int, legalPhases []int) *boardgame.MoveTypeConfig {

	phaseEnum := manager.Delegate().PhaseEnum()

	phaseToStartName := strconv.Itoa(phaseToStart)

	if phaseEnum != nil {
		phaseToStartName = phaseEnum.String(phaseToStart)
	}

	return &boardgame.MoveTypeConfig{
		Name:     "Start Phase " + phaseToStartName,
		HelpText: "Enters phase " + phaseToStartName,
		MoveConstructor: func() boardgame.Move {
			return new(StartPhase)
		},
		CustomConfiguration: boardgame.PropertyCollection{
			configNameStartPhase: phaseToStart,
		},
		IsFixUp:     true,
		LegalPhases: legalPhases,
	}

}
