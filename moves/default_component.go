package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"strconv"
)

//legalTyper is the interface that moves that implement DefaultComponent must
//have.
type legalTyper interface {
	LegalType() int
}

//DefaultComponent is a fix up move type that iterates through SourceStac(),
//and for any non-nil component it encounters that implements
//interface.LegalComponent, calls Legal(). The first index that
//returns nil for that method will set ComponentIndex to that index and
//return. You must provide your own Apply(). This move is convenient for fix
//up moves that have to be done on given components when something becomes
//true. For example, it's useful in checkers to automatically crown tokens
//that make it to the other side of the board.
//
//+autoreader
type DefaultComponent struct {
	FixUpMulti
	ComponentIndex int
}

func (d *DefaultComponent) sourceStackImpl(state boardgame.MutableState) (boardgame.MutableStack, error) {
	sourceStacker, ok := d.TopLevelStruct().(interfaces.SourceStacker)
	if !ok {
		return nil, errors.New("The top level struct doesn't implement SourceStacker")
	}
	return sourceStacker.SourceStack(state), nil
}

func (d *DefaultComponent) legalTypeImpl() (int, error) {
	typer, ok := d.TopLevelStruct().(legalTyper)

	if !ok {
		return 0, errors.New("The top level struct doesn't implement LegalType")
	}
	return typer.LegalType(), nil
}

//DefaultsForState iterates through SourceStack() components one by one from
//start to finish. If the component is non-nil and implments
//interfaces.LegalComponent, calls Legal. If that returns nil,
//it sets the ComponentIndex property to that index and returns.
func (d *DefaultComponent) DefaultsForState(state boardgame.State) {

	//So sorry for this hack. :-(
	mState, ok := state.(boardgame.MutableState)
	if !ok {
		return
	}

	stack, _ := d.sourceStackImpl(mState)

	if stack == nil {
		return
	}

	legalType, _ := d.legalTypeImpl()

	for i, c := range stack.Components() {
		if c == nil {
			continue
		}
		if c.Values == nil {
			continue
		}
		legal, ok := c.Values.(interfaces.LegalComponent)
		if !ok {
			continue
		}
		if legal.Legal(state, legalType) != nil {
			continue
		}
		d.ComponentIndex = i
		return
	}
}

//Legal checks that the component specified by the ComponentIndex index within
//SourceStack implements interfaces.LegalComponent and then returns its return
//value. Otherwise, it errors.
func (d *DefaultComponent) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	if err := d.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}

	//So sorry for this hack. :-(
	mState, ok := state.(boardgame.MutableState)
	if !ok {
		return errors.New("State wasn't convertible to MutableState")
	}

	stack, err := d.sourceStackImpl(mState)
	if err != nil {
		return errors.New("Couldn't get source stack: " + err.Error())
	}
	if stack == nil {
		return errors.New("SourceStack returned nil stack")
	}
	legalType, err := d.legalTypeImpl()
	if err != nil {
		return errors.New("couldn't get legal type: " + err.Error())
	}

	c := stack.ComponentAt(d.ComponentIndex)

	if c == nil {
		return errors.New("ComponentIndex didn't specify a valid component")
	}

	if c.Values == nil {
		return errors.New("Specified component's values were nil")
	}

	legal, ok := c.Values.(interfaces.LegalComponent)

	if !ok {
		return errors.New("Specified Component didn't satisfy LegalComponent")
	}

	return legal.Legal(state, legalType)

}

//LegalType returns the value that will be passed to the Component's Legal()
//legalType argument. It will return the value passed to auto.Config with
//WithLegalType(), or 0 if none was provided. If that behavior isn't
//sufficient, you may override this method.
func (d *DefaultComponent) LegalType() int {
	config := d.Info().Type().CustomConfiguration()

	legalType, ok := config[configNameSourceStack]

	if !ok {
		return 0
	}

	legalTypeInt, ok := legalType.(int)

	if !ok {
		return 0
	}

	return legalTypeInt
}

//SourceStack returns the stack set in configuration by WithSourceStack on the
//GameState, or nil. If that is not sufficient for your needs you should
//override SourceStack yourself.
func (d *DefaultComponent) SourceStack(state boardgame.MutableState) boardgame.MutableStack {
	return sourceStackFromConfig(d, state)
}

func (d *DefaultComponent) ValidConfiguration(exampleState boardgame.MutableState) error {
	if err := d.FixUpMulti.ValidConfiguration(exampleState); err != nil {
		return err
	}

	stack, err := d.sourceStackImpl(exampleState)

	if err != nil {
		return errors.New("SourceStack errored: " + err.Error())
	}

	if stack == nil {
		return errors.New("SourceStack was nil")
	}

	_, err = d.legalTypeImpl()

	if err != nil {
		return errors.New("LegalType errored: " + err.Error())
	}

	//Make sure that at least one component in the deck implements
	//LegalComponent.
	for _, c := range stack.Deck().Components() {
		if c == nil {
			//This shouldn't happen.
			continue
		}
		if c.Values == nil {
			continue
		}

		if _, ok := c.Values.(interfaces.LegalComponent); ok {
			return nil
		}
	}

	return errors.New("No components in the SourceStack's deck implemented LegalComponent.")
}

//MoveTypeFallbackName returns a string based on the stackName passed to
//WithSourceStack, and the LegalType.
func (d *DefaultComponent) MoveTypeFallbackName() string {
	legalType, _ := d.legalTypeImpl()
	return "Default Component For " + stackName(d, configNameSourceStack) + " LegalType " + strconv.Itoa(legalType)
}

//MoveTypeFallbackName returns a string based on the stackName passed to
//WithSourceStack, and the LegalType.
func (d *DefaultComponent) MoveTypeFallbackHelpText() string {
	legalType, _ := d.legalTypeImpl()
	return "Searches " + stackName(d, configNameSourceStack) + " for a component that returns nil for Legal() with LegalType " + strconv.Itoa(legalType)
}
