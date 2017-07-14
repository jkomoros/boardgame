package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"testing"
)

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
		B enum.Var `enum:"B"`
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
	moveType *MoveType
	A        enum.Var `enum:"color"`
}

func (t *testAutoEnumMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testAutoEnumMove) Legal(state State, proposer PlayerIndex) error {
	return nil
}

func (t *testAutoEnumMove) Apply(state MutableState) error {
	return nil
}

func (t *testAutoEnumMove) DefaultsForState(state State) {
	//Pass
}

func (t *testAutoEnumMove) SetType(mType *MoveType) {
	t.moveType = mType
}

func (t *testAutoEnumMove) Type() *MoveType {
	return t.moveType
}

func (t *testAutoEnumMove) Description() string {
	return t.Type().HelpText()
}

var testAutoEnumMoveConfig = MoveTypeConfig{
	Name:     "AutoEnumMove",
	HelpText: "Test move that has a enum.Var that has to be created",
	MoveConstructor: func() Move {
		return new(testAutoEnumMove)
	},
}

func TestAutoEnum(t *testing.T) {
	manager := NewGameManager(&testGameDelegate{}, newTestGameChest(), newTestStorageManager())

	err := manager.AddMoveType(&testAutoEnumMoveConfig)

	assert.For(t).ThatActual(err).IsNil()

	manager.SetUp()

	game := NewGame(manager)

	game.SetUp(0, nil)

	moveType := manager.PlayerMoveTypeByName("AutoEnumMove")

	assert.For(t).ThatActual(moveType).IsNotNil()

	move := moveType.NewMove(game.CurrentState())

	assert.For(t).ThatActual(move).IsNotNil()

	enumVar := move.(*testAutoEnumMove).A

	assert.For(t).ThatActual(enumVar.Enum()).Equals(testColorEnum)

}

type testGeneralReadSetter struct {
	TheInt    int
	EnumConst enum.Const `enum:"color"`
	EnumVar   enum.Var   `enum:"color"`
	TheTimer  *Timer
}

func (t *testGeneralReadSetter) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func TestReaderValidator(t *testing.T) {

	example := &testGeneralReadSetter{}

	game := testGame()

	game.SetUp(0, nil)

	validator, err := newReaderValidator(example.ReadSetter(), example, nil, game.manager.Chest())

	assert.For(t).ThatActual(err).IsNil()

	autoFilledObj := &testGeneralReadSetter{}

	err = validator.Valid(autoFilledObj.ReadSetter())

	assert.For(t).ThatActual(err).IsNotNil()

	err = validator.AutoInflate(autoFilledObj.ReadSetter(), game.CurrentState())

	assert.For(t).ThatActual(err).IsNil()

	err = validator.Valid(autoFilledObj.ReadSetter())

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(autoFilledObj.EnumConst.Enum()).Equals(testColorEnum)

	assert.For(t).ThatActual(autoFilledObj.EnumVar.Enum()).Equals(testColorEnum)

	assert.For(t).ThatActual(autoFilledObj.TheTimer).IsNotNil()
}