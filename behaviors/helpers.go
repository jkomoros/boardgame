package behaviors

import (
	"fmt"
	"reflect"

	"github.com/jkomoros/boardgame"
)

// lookupStackField uses reflection to find a boardgame.Stack field by name on
// the containing SubState. Used by tag-configurable behaviors (DrawDiscardPair,
// FaceUpMarket) to resolve struct tag references to stack fields.
func lookupStackField(containingSubState boardgame.SubState, fieldName string) (boardgame.Stack, error) {
	v := reflect.ValueOf(containingSubState).Elem()
	t := v.Type()

	structField, ok := t.FieldByName(fieldName)
	if !ok {
		return nil, fmt.Errorf("field %q does not exist on the containing struct", fieldName)
	}

	fieldVal := v.FieldByIndex(structField.Index)

	if !fieldVal.CanInterface() {
		return nil, fmt.Errorf("field %q is not accessible", fieldName)
	}

	stack, ok := fieldVal.Interface().(boardgame.Stack)
	if !ok {
		return nil, fmt.Errorf("field %q is not a boardgame.Stack", fieldName)
	}

	if stack == nil {
		return nil, fmt.Errorf("field %q is a Stack but is nil (was it inflated?)", fieldName)
	}

	return stack, nil
}

// lookupSizedStackField is like lookupStackField but for boardgame.SizedStack.
// Used by LocationBehavior.
func lookupSizedStackField(containingSubState boardgame.SubState, fieldName string) (boardgame.SizedStack, error) {
	v := reflect.ValueOf(containingSubState).Elem()
	t := v.Type()

	structField, ok := t.FieldByName(fieldName)
	if !ok {
		return nil, fmt.Errorf("field %q does not exist on the containing struct", fieldName)
	}

	fieldVal := v.FieldByIndex(structField.Index)

	if !fieldVal.CanInterface() {
		return nil, fmt.Errorf("field %q is not accessible", fieldName)
	}

	sizedStack, ok := fieldVal.Interface().(boardgame.SizedStack)
	if !ok {
		return nil, fmt.Errorf("field %q is not a boardgame.SizedStack", fieldName)
	}

	if sizedStack == nil {
		return nil, fmt.Errorf("field %q is a SizedStack but is nil (was it inflated?)", fieldName)
	}

	return sizedStack, nil
}

func validateAttachedStackPair(example boardgame.State, container boardgame.SubState, behaviorName, sourceRole string, source boardgame.Stack, destinationRole string, destination boardgame.Stack) error {
	if example == nil {
		return fmt.Errorf("%s: example state is nil", behaviorName)
	}
	if container == nil || container.State() != example {
		return fmt.Errorf("%s must be connected to the example state", behaviorName)
	}
	if source == destination {
		return fmt.Errorf("%s: %s and %s must be different stacks", behaviorName, sourceRole, destinationRole)
	}
	if err := boardgame.ValidateStackAttachment(example, source); err != nil {
		return fmt.Errorf("%s: %s stack is not attached to the game state: %w", behaviorName, sourceRole, err)
	}
	if err := boardgame.ValidateStackAttachment(example, destination); err != nil {
		return fmt.Errorf("%s: %s stack is not attached to the game state: %w", behaviorName, destinationRole, err)
	}
	if source.Deck() == nil || destination.Deck() == nil {
		return fmt.Errorf("%s: %s and %s must both have decks", behaviorName, sourceRole, destinationRole)
	}
	if source.Deck() != destination.Deck() {
		return fmt.Errorf("%s: %s deck %q does not match %s deck %q", behaviorName, sourceRole, source.Deck().Name(), destinationRole, destination.Deck().Name())
	}
	return nil
}
