package moves

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
)

func namedBehaviorField(state boardgame.ImmutableState, fieldName string, behaviorPointer any) (any, error) {
	if state == nil {
		return nil, fmt.Errorf("state is nil")
	}
	return namedBehaviorValue(state.ImmutableGameState(), fieldName, behaviorPointer)
}

func namedBehaviorValue(immutableGameState any, fieldName string, behaviorPointer any) (any, error) {
	if strings.TrimSpace(fieldName) == "" {
		return nil, fmt.Errorf("behavior field name is empty")
	}
	expectedPointer := reflect.TypeOf(behaviorPointer)
	if expectedPointer == nil || expectedPointer.Kind() != reflect.Ptr || expectedPointer.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("internal behavior lookup expected a pointer to struct, got %T", behaviorPointer)
	}
	gameState := reflect.ValueOf(immutableGameState)
	if !gameState.IsValid() || gameState.Kind() != reflect.Ptr || gameState.IsNil() || gameState.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("gameState must be a non-nil pointer to struct, got %T", immutableGameState)
	}
	value := gameState.Elem()
	structField, ok := value.Type().FieldByName(fieldName)
	if !ok {
		return nil, fmt.Errorf("gameState has no field %q", fieldName)
	}
	if len(structField.Index) != 1 {
		return nil, fmt.Errorf("gameState field %q is promoted; want a direct value field", fieldName)
	}
	if structField.Type != expectedPointer.Elem() {
		return nil, fmt.Errorf("gameState field %q has type %s, want direct value field %s", fieldName, structField.Type, expectedPointer.Elem())
	}
	field := value.FieldByIndex(structField.Index)
	if !field.CanAddr() || !field.Addr().CanInterface() {
		return nil, fmt.Errorf("gameState field %q is not accessible", fieldName)
	}
	result := field.Addr().Interface()
	if reflect.TypeOf(result) != expectedPointer {
		return nil, fmt.Errorf("gameState field %q address has type %T, want %s", fieldName, result, expectedPointer)
	}
	return result, nil
}

func configuredString(config boardgame.PropertyCollection, key, optionName string) (string, bool, error) {
	raw, exists := config[key]
	if !exists {
		return "", false, nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", true, fmt.Errorf("%s config is not a string", optionName)
	}
	if strings.TrimSpace(value) == "" {
		return "", true, fmt.Errorf("%s config cannot be empty", optionName)
	}
	return value, true, nil
}
