package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
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

//StartPhase is a simple move that, when it's its turn in the phase move
//progression, will set the current phase of the game to the given value. When
//you use this, you almost always want ot use moves.AutoConfig, and make sure
//to pass the moves.WithPhaseToStart config object, so that the move has
//enough information to know which phase to enter.
//
//boardgame:codegen
type StartPhase struct {
	FixUp
}

func (s *StartPhase) ValidConfiguration(exampleState boardgame.State) error {

	if err := s.FixUp.ValidConfiguration(exampleState); err != nil {
		return err
	}

	embeddingMove := s.TopLevelStruct()

	phaseStarter, ok := embeddingMove.(phaseToStarter)

	if !ok {
		return errors.New("The embedding move does not have PhaseToStart()")
	}

	delegate := exampleState.Game().Manager().Delegate()

	phaseToStart := phaseStarter.PhaseToStart(delegate.CurrentPhase(exampleState))

	if phaseStarter.PhaseToStart(phaseToStart) < 0 {
		return errors.New("Phase to start returned a negative value, which signals an error. Did you call WithPhaseToStart?")
	}

	if _, ok := exampleState.GameState().(interfaces.CurrentPhaseSetter); !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	phaseEnum := delegate.PhaseEnum()

	if phaseEnum == nil {
		return nil
	}

	treeEnum := phaseEnum.TreeEnum()

	if treeEnum == nil {
		return nil
	}

	if !treeEnum.IsLeaf(phaseToStart) {
		return errors.New("PhaseEnum() returns a TreeEnum, and the phase to start is not a Leaf node.")
	}

	return nil
}

//PhaseToStart uses the Phase provided via StartPhaseMoveConfig constructor
//(or 0 if NewStartPhaseConfig wasn't used). If you want a different behavior,
//override PhaseToStart in your embedding move.
func (s *StartPhase) PhaseToStart(currentPhase int) int {
	config := s.CustomConfiguration()
	val, ok := config[configPropStartPhase]
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
func (s *StartPhase) Apply(state boardgame.State) error {

	phaseEnterer, ok := s.TopLevelStruct().(phaseToStarter)

	if !ok {
		return errors.New("The embedding move does not have PhaseToStart()")
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	phaseToEnter := phaseEnterer.PhaseToStart(currentPhase)

	phaseSetter, ok := state.GameState().(interfaces.CurrentPhaseSetter)

	if !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	beforeLeaver, ok := state.GameState().(interfaces.BeforeLeavePhaser)

	if ok {
		if err := beforeLeaver.BeforeLeavePhase(currentPhase, state); err != nil {
			return errors.New("Before Leave Phase errored: " + err.Error())
		}
	}

	beforeEnterer, ok := state.GameState().(interfaces.BeforeEnterPhaser)

	if ok {
		if err := beforeEnterer.BeforeEnterPhase(phaseToEnter, state); err != nil {
			return errors.New("Before Enter Phase errored: " + err.Error())
		}
	}

	phaseSetter.SetCurrentPhase(phaseToEnter)

	return nil
}

//FallbackName returns "Start Phase PHASENAME" where PHASENAME is the
//string value of the phase to start that was passed via WithPhaseToStart, or
//the int value if no enum was passed.
func (s *StartPhase) FallbackName(m *boardgame.GameManager) string {

	return "Start Phase " + s.phaseStringValue()
}

//FallbackHelpText returns "Enters phase PHASENAME" where PHASENAME is the
//string value of the phase to start that was passed via WithPhaseToStart, or
//the int value if no enum was passed.
func (s *StartPhase) FallbackHelpText() string {
	return "Enters phase " + s.phaseStringValue()
}

func (s *StartPhase) phaseStringValue() string {
	config := s.CustomConfiguration()

	var phaseEnum enum.Enum

	val, ok := config[configPropStartPhaseEnum]

	if ok {
		phaseEnum, _ = val.(enum.Enum)
	}

	val, ok = config[configPropStartPhase]

	if !ok {
		return "InvalidPhase"
	}

	intVal, ok := val.(int)

	if !ok {
		return "InvalidPhase"
	}

	if phaseEnum != nil {
		return phaseEnum.String(intVal)
	}

	return strconv.Itoa(intVal)

}
