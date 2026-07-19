package moves

import (
	"errors"
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// phaseToStarter should be implemented by moves that embed moves.StartPhase to
// configure which phase to enter. It's a private interface because StartPhase
// already has a base PhaseToStart, and to keep the number of interfaces
// smaller.
type phaseToStarter interface {
	PhaseToStart(currentPhase enum.EnumKey) (enum.EnumKey, error)
}

// StartPhase is a simple move that, when it's its turn in the phase move
// progression, will set the current phase of the game to the given value. When
// you use this, you almost always want ot use moves.AutoConfig, and make sure
// to pass the moves.WithPhaseToStart config object, so that the move has
// enough information to know which phase to enter.
//
//boardgame:codegen
type StartPhase struct {
	FixUp
}

// ValidConfiguration checks that the embedding move implements PhaseToStart
// which returns a non-negative value, and that GameState implements
// interfaces.CurrentPhaseStarter, and that PhaseEnum exists and if it's a
// TreeEnum, that the phaseToStart is a leaf enum value.
func (s *StartPhase) ValidConfiguration(exampleState boardgame.State) error {

	if err := s.FixUp.ValidConfiguration(exampleState); err != nil {
		return err
	}

	embeddingMove := s.Info().ConcreteMove()

	phaseStarter, ok := embeddingMove.(phaseToStarter)

	if !ok {
		return errors.New("The embedding move does not have PhaseToStart()")
	}

	delegate := exampleState.Manager().Delegate()

	_, currentPhaseKey := currentPhaseInfo(exampleState)
	phaseToStart, err := phaseStarter.PhaseToStart(currentPhaseKey)

	if err != nil {
		return errors.New("PhaseToStart returned an error: " + err.Error())
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
		return errors.New("phaseEnum() returns a TreeEnum, and the phase to start is not a Leaf node")
	}

	return nil
}

// PhaseToStart uses the Phase provided via StartPhaseMoveConfig constructor
// (or 0 if NewStartPhaseConfig wasn't used). If you want a different behavior,
// override PhaseToStart in your embedding move.
func (s *StartPhase) PhaseToStart(currentPhase enum.EnumKey) (enum.EnumKey, error) {
	config := s.CustomConfiguration()
	val, ok := config[configPropStartPhase]
	if !ok {
		return 0, errors.New("no phase configured; did you call WithPhaseToStart?")
	}
	keyVal, ok := val.(enum.EnumKey)
	if !ok {
		return 0, errors.New("configured phase value is not an EnumKey")
	}
	return keyVal, nil
}

// Apply call BeforeLeavePhase() (if it exists), then BeforeEnterPhase() (if it
// exists),then SetCurrentPhase to the phase index returned by PhaseToStart
// from this move type.
func (s *StartPhase) Apply(state boardgame.State) error {

	phaseEnterer, ok := s.Info().ConcreteMove().(phaseToStarter)

	if !ok {
		return errors.New("The embedding move does not have PhaseToStart()")
	}

	delegate := state.Manager().Delegate()

	currentPhaseVal, currentPhaseKey := currentPhaseInfo(state)

	phaseToEnter, err := phaseEnterer.PhaseToStart(currentPhaseKey)
	if err != nil {
		return errors.New("PhaseToStart returned an error: " + err.Error())
	}

	phaseSetter, ok := state.GameState().(interfaces.CurrentPhaseSetter)

	if !ok {
		return errors.New("The gameState does not implement CurrentPhaseSetter")
	}

	beforeLeaver, ok := state.GameState().(interfaces.BeforeLeavePhaser)

	if ok && currentPhaseVal != nil {
		if err := beforeLeaver.BeforeLeavePhase(currentPhaseVal, state); err != nil {
			return errors.New("Before Leave Phase errored: " + err.Error())
		}
	}

	beforeEnterer, ok := state.GameState().(interfaces.BeforeEnterPhaser)

	if ok {
		phaseEnum := delegate.PhaseEnum()
		if phaseEnum == nil {
			return errors.New("GameState implements BeforeEnterPhaser but PhaseEnum() returns nil")
		}
		phaseToEnterVal, err := phaseEnum.NewImmutableVal(phaseToEnter)
		if err != nil {
			return errors.New("Before Enter Phase: could not create val for phase: " + err.Error())
		}
		if err := beforeEnterer.BeforeEnterPhase(phaseToEnterVal, state); err != nil {
			return errors.New("Before Enter Phase errored: " + err.Error())
		}
	}

	phaseSetter.SetCurrentPhase(phaseToEnter)

	return nil
}

// FallbackName returns "Start Phase PHASENAME" where PHASENAME is the
// string value of the phase to start that was passed via WithPhaseToStart, or
// the int value if no enum was passed.
func (s *StartPhase) FallbackName(m *boardgame.GameManager) string {

	return "Start Phase " + s.phaseStringValue()
}

// FallbackHelpText returns "Enters phase PHASENAME" where PHASENAME is the
// string value of the phase to start that was passed via WithPhaseToStart, or
// the int value if no enum was passed.
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

	keyVal, ok := val.(enum.EnumKey)

	if !ok {
		return "InvalidPhase"
	}

	if phaseEnum != nil {
		return phaseEnum.String(keyVal)
	}

	return strconv.Itoa(keyVal.Int())

}
