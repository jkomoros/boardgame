package auto

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"reflect"
)

//Note that tests for this package are actually in the underlying moves
//package, effectively, because it's too much of a pain to recreate them here,
//or to do the tests for eh basic moves wihtout having auto config.

//AutoConfigurableMove is the interface that moves passed to auto.Config must
//implement. These methods are interrogated to set the move name,
//helptext,isFixUp, and legalPhases to good values. moves.Base defines
//powerful stubs for these, so any moves that embed moves.Base (or embed a
//move that embeds moves.Base, etc) satisfy this interface.
type AutoConfigurableMove interface {
	//DefaultConfigMoves all must implement all Move methods.
	boardgame.Move
	//The name for the move type
	MoveTypeName() string
	//The HelpText to use.
	MoveTypeHelpText() string
	//Whether the move should be a fix up.
	MoveTypeIsFixUp() bool
	//Result will be used for LegalPhases in the config.
	MoveTypeLegalPhases() []int
}

//MustConfig is a wrapper around Config that if it errors will panic. Only
//suitable for being used during setup.
func MustConfig(exampleStruct AutoConfigurableMove, options ...interfaces.CustomConfigurationOption) *boardgame.MoveTypeConfig {
	result, err := Config(exampleStruct, options...)

	if err != nil {
		panic("Couldn't Config: " + err.Error())
	}

	return result
}

//Config is a powerful default MoveTypeConfig generator. In many cases
//you'll implement moves that are very thin embeddings of moves in this
//package. Generating a MoveTypeConfig for each is a pain. This method auto-
//generates the MoveTypeConfig based on an example zero type of your move to
//install. Moves need a few extra methods that are consulted to generate the
//move name, helptext, and isFixUp; anything based on moves.Base automatically
//satisfies the necessary interface. See the package doc for an example of
//use.
func Config(exampleStruct AutoConfigurableMove, options ...interfaces.CustomConfigurationOption) (*boardgame.MoveTypeConfig, error) {

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

	throwAwayConfig := newMoveTypeConfig("Temporary Move", "Temporary Move Help Text", false, nil, exampleStruct, config)

	throwAwayMoveType, err := throwAwayConfig.NewMoveType(nil)

	if err != nil {
		//Look for exatly the single kind of error we're OK with. Yes, this is a hack.
		if err.Error() != "No manager passed, so we can'd do validation" {
			return nil, errors.New("Couldn't create intermediate move type: " + err.Error())
		}
	}

	//the move returned from NewMove is guaranteed to implement
	//DefaultConfigMove, because it's fundamentally an exampleStruct.
	actualExample := throwAwayMoveType.NewMove(nil).(AutoConfigurableMove)

	name := actualExample.MoveTypeName()
	helpText := actualExample.MoveTypeHelpText()
	isFixUp := actualExample.MoveTypeIsFixUp()
	legalPhases := actualExample.MoveTypeLegalPhases()

	moveTypeConfig, err := newMoveTypeConfig(name, helpText, isFixUp, legalPhases, exampleStruct, config), nil

	return moveTypeConfig, err

}

func newMoveTypeConfig(name, helpText string, isFixUp bool, legalPhases []int, exampleStruct boardgame.Move, config boardgame.PropertyCollection) *boardgame.MoveTypeConfig {
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
		LegalPhases:         legalPhases,
	}
}
