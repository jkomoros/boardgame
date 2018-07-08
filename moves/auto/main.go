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
	//DeriveName() will be called to generate the name. This might be an
	//expensive method, so it will only be called during installation.
	DeriveName() string
}

//MustConfig is a wrapper around Config that if it errors will panic. Only
//suitable for being used during setup.
func MustConfig(exampleStruct AutoConfigurableMove, options ...interfaces.CustomConfigurationOption) boardgame.MoveConfig {
	result, err := Config(exampleStruct, options...)

	if err != nil {
		panic("Couldn't Config: " + err.Error())
	}

	return result
}

//Config is a powerful default MoveConfig generator. In many cases
//you'll implement moves that are very thin embeddings of moves in this
//package. Generating a MoveConfig for each is a pain. This method auto-
//generates the MoveConfig based on an example zero type of your move to
//install. Moves need a few extra methods that are consulted to generate the
//move name, helptext, and isFixUp; anything based on moves.Base automatically
//satisfies the necessary interface. See the package doc for an example of
//use.
func Config(exampleStruct AutoConfigurableMove, options ...interfaces.CustomConfigurationOption) (boardgame.MoveConfig, error) {

	config := make(boardgame.PropertyCollection, len(options))

	for _, option := range options {
		option(config)
	}

	if exampleStruct == nil {
		return boardgame.MoveConfig{}, errors.New("nil struct provided")
	}

	//We'll create a throw-away move type config first to get a fully-
	//initialized and expanded move (e.g. with all tag-based autoinflation)
	//that we can then pass to the MoveType* methods, so they'll have more to work with.

	throwAwayConfig := newMoveConfig("Temporary Move", exampleStruct, config)

	generatedExample, err := throwAwayConfig.OrphanExampleMove()

	if err != nil {
		return boardgame.MoveConfig{}, err
	}

	actualExample := generatedExample.(AutoConfigurableMove)

	name := actualExample.DeriveName()

	moveTypeConfig, err := newMoveConfig(name, exampleStruct, config), nil

	return moveTypeConfig, err

}

func newMoveConfig(name string, exampleStruct boardgame.Move, config boardgame.PropertyCollection) boardgame.MoveConfig {
	val := reflect.ValueOf(exampleStruct)

	//We can accept either pointer or struct types.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	return boardgame.MoveConfig{
		Name: name,
		MoveConstructor: func() boardgame.Move {
			return reflect.New(typ).Interface().(boardgame.Move)
		},
		CustomConfiguration: config,
	}
}
