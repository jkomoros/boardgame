package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"reflect"
)

//GroupableMoveConfig is a type of MoveConfig that also has enough methods for
//it to be used as a MoveProgressionGroup in AddOrderedForPhase.
//AutoConfigurer.Configure() returns these so that they can be nested directly
//in any of the objects in the moves/groups package.
type GroupableMoveConfig interface {
	boardgame.MoveConfig
	interfaces.MoveProgressionGroup
}

//NewGroupableMoveConfig takes a generic boardgame.MoveConfig and makes it
//satisfy the GroupableMoveConfig interface, so it can be used as a child in
//the objects in moves/groups. The config returned will simply return a list
//with a single item of itself for MoveConfigs. For Satisfied, it will consume
//a move that shares its own name, and, if it implements
//AllowMultipleInProgression() and returns true from that, it will consume as
//many of those moves in a row as exist from the front of the tape.
//AutoConfigurer.Config() returns objects that have been run through this
//automatically, but it's a public function in case you want to decorate a
//move config you generated manually and not from AutoConfigurer.Config().
func NewGroupableMoveConfig(config boardgame.MoveConfig) GroupableMoveConfig {
	return &defaultMoveConfig{
		config,
	}
}

type defaultMoveConfig struct {
	boardgame.MoveConfig
}

func (d *defaultMoveConfig) MoveConfigs() []boardgame.MoveConfig {
	return []boardgame.MoveConfig{d.MoveConfig}
}

func (d *defaultMoveConfig) Satisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {

	if tape == nil {
		return nil, nil
	}

	allowMultiple := false

	if allowMultipler, ok := d.Constructor()().(interfaces.AllowMultipleInProgression); ok {
		if allowMultipler.AllowMultipleInProgression() {
			allowMultiple = true
		}
	}

	if allowMultiple {
		return d.multipleItemSatisfied(tape)
	}

	return d.singleItemSatisified(tape)
}

func (d *defaultMoveConfig) singleItemSatisified(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {
	//If there's one, and the next item either doesn't exist or has a different name, OK.

	if tape.MoveName != d.Name() {
		return errors.New("Move name does not match: " + tape.MoveName + " is not " + d.Name()), tape
	}

	return nil, tape.Rest

}

func (d *defaultMoveConfig) multipleItemSatisfied(tape *interfaces.MoveGroupHistoryItem) (error, *interfaces.MoveGroupHistoryItem) {
	//Must have at least one; consume as many as you can

	if tape.MoveName != d.Name() {
		return errors.New("Move name does not match: " + tape.MoveName + " is not " + d.Name()), tape
	}

	tape = tape.Rest

	//Consume as many as we can
	for tape != nil {

		if tape.MoveName != d.Name() {
			//Found the first one that wasn't us; return
			return nil, tape
		}

		//Keep consuming

		tape = tape.Rest
	}

	//If we get to here we fell off the end of the tape, which is legal.
	return nil, nil
}

//AutoConfigurableMove is the interface that moves passed to AutoConfigurer.Config must
//implement. These methods are interrogated to set the move name,
//helptext,isFixUp, and legalPhases to good values. moves.Base defines
//powerful stubs for these, so any moves that embed moves.Base (or embed a
//move that embeds moves.Base, etc) satisfy this interface.
type AutoConfigurableMove interface {
	//DefaultConfigMoves all must implement all Move methods.
	boardgame.Move
	//DeriveName() will be called to generate the name. This might be an
	//expensive method, so it will only be called during installation.
	DeriveName(manager *boardgame.GameManager) string
}

//AutoConfigurer is an object that makes it easy to configure moves. Get a new
//one with NewAutoConfigurer. See the package doc for much more on how to use
//it.
type AutoConfigurer struct {
	delegate boardgame.GameDelegate
}

//NewAutoConfigurer returns a new AutoConfigurer ready for use.
func NewAutoConfigurer(g boardgame.GameDelegate) *AutoConfigurer {
	return &AutoConfigurer{
		delegate: g,
	}
}

//MustConfig is a wrapper around Config that if it errors will panic. Only
//suitable for being used during setup.
func (a *AutoConfigurer) MustConfig(exampleStruct AutoConfigurableMove, options ...interfaces.CustomConfigurationOption) GroupableMoveConfig {
	result, err := a.Config(exampleStruct, options...)

	if err != nil {
		panic("Couldn't Config: " + err.Error())
	}

	return result
}

//Config is a powerful default MoveConfig generator. In many cases you'll
//implement moves that are very thin embeddings of moves in this package.
//Generating a MoveConfig for each is a pain. This method auto- generates the
//MoveConfig based on an example zero type of your move to install. Moves need
//a few extra methods that are consulted to generate the move name, helptext,
//and isFixUp; anything based on moves.Base automatically satisfies the
//necessary interface. See the package doc for an example of use. Instead of
//returning a boardgame.MoveConfig, it returns a GroupableMoveConfig
//equivalent to what you'd get from NewGroupableMoveConfig, which satisfies
//boardgame.MoveConfig but also adds enough methods to be useable as input to
//AddOrderedForPhase.
func (a *AutoConfigurer) Config(exampleStruct AutoConfigurableMove, options ...interfaces.CustomConfigurationOption) (GroupableMoveConfig, error) {

	if a.delegate == nil {
		return nil, errors.New("No delegate provided")
	}

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

	throwAwayConfig := newMoveConfig("Temporary Move", exampleStruct, config)

	generatedExample, err := boardgame.OrphanExampleMove(throwAwayConfig)

	if err != nil {
		return nil, err
	}

	actualExample := generatedExample.(AutoConfigurableMove)

	name := actualExample.DeriveName(a.delegate.Manager())

	moveTypeConfig, err := newMoveConfig(name, exampleStruct, config), nil

	return moveTypeConfig, err

}

func newMoveConfig(name string, exampleStruct boardgame.Move, config boardgame.PropertyCollection) GroupableMoveConfig {
	val := reflect.ValueOf(exampleStruct)

	//We can accept either pointer or struct types.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	constructor := func() boardgame.Move {
		return reflect.New(typ).Interface().(boardgame.Move)
	}

	return &defaultMoveConfig{
		boardgame.NewMoveConfig(name, constructor, config),
	}
}
