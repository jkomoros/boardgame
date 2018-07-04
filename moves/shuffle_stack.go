package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//ShuffleStack is a move, typically used in SetUp phases, that simply shuffles
//a given stack. The struct you embed this in should implement SourceStacker.
//
//In practice it is common to just use this move directly in your game, and
//pass the stack via WithSourceStack to auto.Config.
//
//+autoreader
type ShuffleStack struct {
	FixUp
}

type moveInfoer interface {
	Info() *boardgame.MoveInfo
}

func sourceStackFromConfig(m moveInfoer, state boardgame.State) boardgame.Stack {
	config := m.Info().Type().CustomConfiguration()

	stackName, ok := config[configNameSourceStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := state.GameState().ReadSetter().StackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//SourceStack by default just returns the property on GameState with the name
//passed to DefaultConfig by WithSourceStack. If that is not sufficient,
//override this in your embedding struct.
func (s *ShuffleStack) SourceStack(state boardgame.State) boardgame.Stack {
	return sourceStackFromConfig(s, state)
}

//We don't need a Legal method because the pass-through to moves.Base is sufficient.

//Apply shuffles the stack that the embedding move selects by the return value
//from SourceStack().
func (s *ShuffleStack) Apply(state boardgame.State) error {
	embeddingMove := s.TopLevelStruct()

	stacker, ok := embeddingMove.(interfaces.SourceStacker)

	if !ok {
		return errors.New("Embedding move unexpectedly did not implement SourceStacker")
	}

	stack := stacker.SourceStack(state)

	if stack == nil {
		return errors.New("PrimaryStack returned a nil stack")
	}

	return stack.Shuffle()
}

func (s *ShuffleStack) ValidConfiguration(exampleState boardgame.State) error {
	testMove := s.TopLevelStruct()

	sourceStacker, ok := testMove.(interfaces.SourceStacker)

	if !ok {
		return errors.New("The embedding Move doesn't implement SourceStacker")
	}

	if sourceStacker.SourceStack(exampleState) == nil {
		return errors.New("SourceStack returned nil")
	}

	return nil
}

//FallbackName returns "Shuffle STACK" where STACK is the name of the
//stack set by WithSourceStack.
func (s *ShuffleStack) FallbackName() string {
	return "Shuffle " + stackName(s, configNameSourceStack)
}

//FallbackHelpText returns "Shuffles the STACK stack" where STACK is the name
//of the stack set by WithSourceStack.
func (s *ShuffleStack) FallbackHelpText() string {
	return "Shuffles " + stackName(s, configNameSourceStack)
}
