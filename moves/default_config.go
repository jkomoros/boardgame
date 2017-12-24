package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
	"reflect"
)

//The interface that moves that can be handled by DefaultConfig implement.
type defaultConfigFallbackMoveType interface {
	//The last resort move-name generator that MoveName will fall back on if
	//none of the other options worked.
	MoveTypeFallbackName(manager *boardgame.GameManager) string
	MoveTypeFallbackHelpText(manager *boardgame.GameManager) string
	MoveTypeFallbackIsFixUp(manager *boardgame.GameManager) bool
}

//MustDefaultConfig is a wrapper around DefaultConfig that if it errors will
//panic. Only suitable for being used during setup.
func MustDefaultConfig(manager *boardgame.GameManager, exampleStruct moveinterfaces.DefaultConfigMove, options ...CustomConfigurationOption) *boardgame.MoveTypeConfig {
	result, err := DefaultConfig(manager, exampleStruct, options...)

	if err != nil {
		panic("Couldn't DefaultConfig: " + err.Error())
	}

	return result
}

//DefaultConfig is a powerful default MoveTypeConfig generator. In many cases
//you'll implement moves that are very thin embeddings of moves in this
//package. Generating a MoveTypeConfig for each is a pain. This method auto-
//generates the MoveTypeConfig based on an example zero type of your move to
//install. Moves need a few extra methods that are consulted to generate the
//move name, helptext, and isFixUp; anything based on moves.Base automatically
//satisfies the necessary interface. See the package doc for an example of
//use.
func DefaultConfig(manager *boardgame.GameManager, exampleStruct moveinterfaces.DefaultConfigMove, options ...CustomConfigurationOption) (*boardgame.MoveTypeConfig, error) {

	config := make(boardgame.PropertyCollection, len(options))

	for _, option := range options {
		option(config)
	}

	if exampleStruct == nil {
		return nil, errors.New("nil struct provided")
	}

	//We'll create a throw-away move type config first to get a fully-
	//initialized and expanded move (e.g. with all tag-based autoinflation)
	//that we can then pass to the MoveType* methods, so they'll have more to work with.

	throwAwayConfig := newMoveTypeConfig("Temporary Move", "Temporary Move Help Text", false, exampleStruct, config)

	throwAwayMoveType, err := throwAwayConfig.NewMoveType(manager)

	if err != nil {
		return nil, errors.New("Couldn't create temporary move type: " + err.Error())
	}

	//the move returned from NewMove is guaranteed to implement
	//DefaultConfigMove, because it's fundamentally an exampleStruct.
	actualExample := throwAwayMoveType.NewMove(manager.ExampleState()).(moveinterfaces.DefaultConfigMove)

	name := actualExample.MoveTypeName(manager)
	helpText := actualExample.MoveTypeHelpText(manager)
	isFixUp := actualExample.MoveTypeIsFixUp(manager)

	return newMoveTypeConfig(name, helpText, isFixUp, exampleStruct, config), nil

}

func newMoveTypeConfig(name, helpText string, isFixUp bool, exampleStruct boardgame.Move, config boardgame.PropertyCollection) *boardgame.MoveTypeConfig {
	val := reflect.ValueOf(exampleStruct)

	//We can accept either pointer or struct types.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	return &boardgame.MoveTypeConfig{
		Name:     name,
		HelpText: helpText,
		MoveConstructor: func() boardgame.Move {
			return reflect.New(typ).Interface().(boardgame.Move)
		},
		CustomConfiguration: config,
		IsFixUp:             isFixUp,
	}
}

//stackPropName takes a stack that was returned from a given state, and the
//state it was returned from. It searches through the sub- states in state
//until it finds the name of the property where it resides.
func stackPropName(stack boardgame.MutableStack, state boardgame.MutableState) string {

	if name := stackPropNameInReadSetter(stack, state.MutableGameState().ReadSetter()); name != "" {
		return name
	}

	if name := stackPropNameInReadSetter(stack, state.MutablePlayerStates()[0].ReadSetter()); name != "" {
		return name
	}

	for _, dynamicComponentValues := range state.MutableDynamicComponentValues() {
		if name := stackPropNameInReadSetter(stack, dynamicComponentValues[0].ReadSetter()); name != "" {
			return name
		}
	}

	return ""

}

func stackPropNameInReadSetter(stack boardgame.MutableStack, readSetter boardgame.PropertyReadSetter) string {
	for propName, propType := range readSetter.Props() {
		if propType != boardgame.TypeStack {
			continue
		}
		testStack, err := readSetter.MutableStackProp(propName)
		if err != nil {
			continue
		}
		if testStack == stack {
			return propName
		}
	}
	return ""
}
