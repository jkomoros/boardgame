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
	MoveTypeFallbackName() string
	MoveTypeFallbackHelpText() string
	MoveTypeFallbackIsFixUp() bool
}

//MustDefaultConfig is a wrapper around DefaultConfig that if it errors will
//panic. Only suitable for being used during setup.
func MustDefaultConfig(exampleStruct moveinterfaces.DefaultConfigMove, options ...CustomConfigurationOption) *boardgame.MoveTypeConfig {
	result, err := DefaultConfig(exampleStruct, options...)

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
func DefaultConfig(exampleStruct moveinterfaces.DefaultConfigMove, options ...CustomConfigurationOption) (*boardgame.MoveTypeConfig, error) {

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

	throwAwayMoveType, err := throwAwayConfig.NewMoveType(nil)

	if err != nil {
		//Look for exatly the single kind of error we're OK with. Yes, this is a hack.
		if err.Error() != "No manager passed, so we can'd do validation" {
			return nil, errors.New("Couldn't create intermediate move type: " + err.Error())
		}
	}

	//the move returned from NewMove is guaranteed to implement
	//DefaultConfigMove, because it's fundamentally an exampleStruct.
	actualExample := throwAwayMoveType.NewMove(nil).(moveinterfaces.DefaultConfigMove)

	name := actualExample.MoveTypeName()
	helpText := actualExample.MoveTypeHelpText()
	isFixUp := actualExample.MoveTypeIsFixUp()

	moveTypeConfig, err := newMoveTypeConfig(name, helpText, isFixUp, exampleStruct, config), nil

	return moveTypeConfig, err

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

//A func that will fail to compile if all of the moves don't have a valid fallback.
func ensureAllMovesSatisfyFallBack() {
	var m defaultConfigFallbackMoveType
	m = new(ApplyUntil)
	m = new(ApplyUntilCount)
	m = new(ApplyCountTimes)
	m = new(Base)
	m = new(CollectCountComponents)
	m = new(CollectComponentsUntilGameCountReached)
	m = new(CollectComponentsUntilPlayerCountLeft)
	m = new(CurrentPlayer)
	m = new(DealCountComponents)
	m = new(DealComponentsUntilGameCountLeft)
	m = new(DealComponentsUntilPlayerCountReached)
	m = new(FinishTurn)
	m = new(MoveCountComponents)
	m = new(MoveComponentsUntilCountLeft)
	m = new(MoveComponentsUntilCountReached)
	m = new(RoundRobin)
	m = new(RoundRobinNumRounds)
	m = new(ShuffleStack)
	m = new(StartPhase)
	if m != nil {
		return
	}
}
