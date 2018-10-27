package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestPolicyFromStructTag(t *testing.T) {

	errorMap := map[int]Policy{
		GroupAll: PolicyInvalid,
	}

	tests := []struct {
		in       string
		expected map[int]Policy
	}{
		{
			"",
			map[int]Policy{
				GroupAll: PolicyVisible,
			},
		},
		{
			"hidden",
			map[int]Policy{
				GroupAll: PolicyHidden,
			},
		},
		{
			"other:hidden",
			map[int]Policy{
				GroupOther: PolicyHidden,
			},
		},
		{
			"all:order,other:hidden",
			map[int]Policy{
				GroupOther: PolicyHidden,
				GroupAll:   PolicyOrder,
			},
		},
		{
			"all:random:foo",
			errorMap,
		},
	}

	for i, test := range tests {
		result := policyFromStructTag(test.in, "all")
		assert.For(t, i).ThatActual(result).Equals(test.expected)
	}

}

func TestStructTag(t *testing.T) {

	type anonTestStruct struct {
		C int `enum:"C"`
	}

	type anonPointerTestStruct struct {
		D int `enum:"D"`
	}

	type testStruct struct {
		anonTestStruct
		*anonPointerTestStruct
		A int
		B enum.Val `enum:"B"`
	}

	theStruct := &testStruct{
		anonPointerTestStruct: &anonPointerTestStruct{},
	}

	assert.For(t).ThatActual(structTagForField(theStruct, "A", enumStructTag)).Equals("")
	assert.For(t).ThatActual(structTagForField(theStruct, "B", enumStructTag)).Equals("B")
	assert.For(t).ThatActual(structTagForField(theStruct, "C", enumStructTag)).Equals("C")
	assert.For(t).ThatActual(structTagForField(theStruct, "D", enumStructTag)).Equals("D")
	assert.For(t).ThatActual(structTagForField(theStruct, "Illegal", enumStructTag)).Equals("")

}

type testAutoEnumMove struct {
	info           *MoveInfo
	topLevelStruct Move
	A              enum.Val `enum:"color"`
}

func (t *testAutoEnumMove) Reader() PropertyReader {
	return getDefaultReader(t)
}

func (t *testAutoEnumMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testAutoEnumMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

func (t *testAutoEnumMove) Legal(state ImmutableState, proposer PlayerIndex) error {
	return nil
}

func (t *testAutoEnumMove) IsFixUp() bool {
	return false
}

func (t *testAutoEnumMove) Apply(state State) error {
	return nil
}

func (t *testAutoEnumMove) DefaultsForState(state ImmutableState) {
	//Pass
}

func (t *testAutoEnumMove) SetInfo(m *MoveInfo) {
	t.info = m
}

func (t *testAutoEnumMove) Info() *MoveInfo {
	return t.info
}

func (t *testAutoEnumMove) SetTopLevelStruct(m Move) {
	t.topLevelStruct = m
}

func (t *testAutoEnumMove) TopLevelStruct() Move {
	return t.topLevelStruct
}

func (t *testAutoEnumMove) ValidConfiguration(exampleState State) error {
	return nil
}

func (t *testAutoEnumMove) Description() string {
	return t.TopLevelStruct().HelpText()
}

func (t *testAutoEnumMove) HelpText() string {
	return "Test move that has a enum.Var that has to be created"
}

var testAutoEnumMoveConfig = NewMoveConfig(
	"AutoEnumMove",
	func() Move {
		return new(testAutoEnumMove)
	},
	nil)

func TestAutoEnum(t *testing.T) {

	moveInstaller := func(manager *GameManager) []MoveConfig {
		return []MoveConfig{
			testAutoEnumMoveConfig,
		}
	}

	manager, err := NewGameManager(&testGameDelegate{moveInstaller: moveInstaller}, newTestStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	move := game.MoveByName("AutoEnumMove")

	assert.For(t).ThatActual(move).IsNotNil()

	enumVar := move.(*testAutoEnumMove).A

	assert.For(t).ThatActual(enumVar.Enum()).Equals(testColorEnum)

}

type testGeneralReadSetter struct {
	TheInt            int               `sanitize:"hidden"`
	EnumConst         enum.ImmutableVal `enum:"color"`
	EnumVar           enum.Val          `enum:"color"`
	TheImmutableTimer ImmutableTimer
	TheTimer          Timer
	TheSizedStack     Stack `sizedstack:"test,0"`
	TheGrowableStack  Stack `stack:"test" sanitize:"order"`
}

func (t *testGeneralReadSetter) Reader() PropertyReader {
	return getDefaultReader(t)
}

func (t *testGeneralReadSetter) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testGeneralReadSetter) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

func TestReaderValidator(t *testing.T) {

	example := &testGeneralReadSetter{}

	game := testDefaultGame(t, false)

	validator, err := newReaderValidator(example, nil, game.manager.Chest())

	assert.For(t).ThatActual(err).IsNil()

	autoFilledObj := &testGeneralReadSetter{}

	err = validator.Valid(autoFilledObj.ReadSetter())

	assert.For(t).ThatActual(err).IsNotNil()

	err = validator.AutoInflate(autoFilledObj.ReadSetConfigurer(), game.CurrentState())

	assert.For(t).ThatActual(err).IsNil()

	err = validator.Valid(autoFilledObj.ReadSetter())

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(autoFilledObj.EnumConst.Enum()).Equals(testColorEnum)

	assert.For(t).ThatActual(autoFilledObj.EnumVar.Enum()).Equals(testColorEnum)

	assert.For(t).ThatActual(autoFilledObj.TheTimer).IsNotNil()

	assert.For(t).ThatActual(validator.sanitizationPolicy["TheInt"]).Equals(map[int]Policy{
		GroupAll: PolicyHidden,
	})

	assert.For(t).ThatActual(validator.sanitizationPolicy["TheGrowableStack"]).Equals(map[int]Policy{
		GroupAll: PolicyOrder,
	})

	assert.For(t).ThatActual(validator.sanitizationPolicy["TheTimer"]).Equals(map[int]Policy{
		GroupAll: PolicyVisible,
	})

}
