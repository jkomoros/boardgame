package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

//ShuffleStack is a move, typically used in SetUp phases, that simply shuffles
//a given stack. The struct you embed this in should implement SourceStacker.
type ShuffleStack struct {
	Base
}

//SourceStack by default just returns the property on GameState with the name
//passed to DefaultConfig by WithSourceStack. If that is not sufficient,
//override this in your embedding struct.
func (s *ShuffleStack) SourceStack(state boardgame.MutableState) boardgame.MutableStack {
	config := s.Info().Type().CustomConfiguration()

	stackName, ok := config[configNameSourceStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := state.MutableGameState().ReadSetter().MutableStackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//We don't need a Legal method because the pass-through to moves.Base is sufficient.

//Apply shuffles the stack that the embedding move selects by the return value
//from SourceStack().
func (s *ShuffleStack) Apply(state boardgame.MutableState) error {
	embeddingMove := s.TopLevelStruct()

	stacker, ok := embeddingMove.(moveinterfaces.SourceStacker)

	if !ok {
		return errors.New("Embedding move unexpectedly did not implement SourceStacker")
	}

	stack := stacker.SourceStack(state)

	if stack == nil {
		return errors.New("PrimaryStack returned a nil stack")
	}

	return stack.Shuffle()
}

func (s *ShuffleStack) ValidConfiguration(exampleState boardgame.MutableState) error {
	testMove := s.TopLevelStruct()

	sourceStacker, ok := testMove.(moveinterfaces.SourceStacker)

	if !ok {
		return errors.New("The embedding Move doesn't implement SourceStacker")
	}

	if sourceStacker.SourceStack(exampleState) == nil {
		return errors.New("SourceStack returned nil")
	}

	return nil
}

func (s *ShuffleStack) stackName(manager *boardgame.GameManager) string {

	name := "UnknownStack"

	stacker, ok := s.TopLevelStruct().(moveinterfaces.SourceStacker)

	if !ok {
		return name
	}

	state := manager.ExampleState()

	stack := stacker.SourceStack(state)

	if stack == nil {
		return name
	}

	if actualName := stackPropName(stack, state); actualName != "" {
		return actualName
	}

	return name
}

func (s *ShuffleStack) MoveTypeFallbackName(manager *boardgame.GameManager) string {
	return "Shuffle " + s.stackName(manager)
}

func (s *ShuffleStack) MoveTypeFallbackHelpText(manager *boardgame.GameManager) string {
	return "Shuffles the " + s.stackName(manager) + " stack"
}

func (s *ShuffleStack) MoveTypeIsFixUp(manager *boardgame.GameManager) bool {
	return true
}
