package boardgame

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
)

func TestPolicyFromStructTag(t *testing.T) {

	errorMap := map[string]Policy{
		"all": PolicyInvalid,
	}

	tests := []struct {
		in       string
		expected map[string]Policy
	}{
		{
			"",
			map[string]Policy{
				"all": PolicyVisible,
			},
		},
		{
			"hidden",
			map[string]Policy{
				"all": PolicyHidden,
			},
		},
		{
			"other:hidden",
			map[string]Policy{
				"other": PolicyHidden,
			},
		},
		{
			"all:order,other:hidden",
			map[string]Policy{
				"other": PolicyHidden,
				"all":   PolicyOrder,
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

type testSizedStackWithGrowableTag struct {
	Cards SizedStack `stack:"test"`
}

func (t *testSizedStackWithGrowableTag) Reader() PropertyReader { return getDefaultReader(t) }
func (t *testSizedStackWithGrowableTag) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testSizedStackWithGrowableTag) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

type testStackWithConflictingConcreteTags struct {
	Cards Stack `stack:"test" sizedstack:"test,2"`
}

func (t *testStackWithConflictingConcreteTags) Reader() PropertyReader { return getDefaultReader(t) }
func (t *testStackWithConflictingConcreteTags) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testStackWithConflictingConcreteTags) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

type testMergedStackWithConflictingTags struct {
	First  Stack
	Second Stack
	Cards  MergedStack `concatenate:"First,Second" overlap:"First,Second"`
}

func (t *testMergedStackWithConflictingTags) Reader() PropertyReader { return getDefaultReader(t) }
func (t *testMergedStackWithConflictingTags) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testMergedStackWithConflictingTags) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

type testStackWithConcreteAndMergedTags struct {
	First Stack
	Cards Stack `stack:"test" concatenate:"First"`
}

func (t *testStackWithConcreteAndMergedTags) Reader() PropertyReader { return getDefaultReader(t) }
func (t *testStackWithConcreteAndMergedTags) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testStackWithConcreteAndMergedTags) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

type testNonStackWithStackTag struct {
	Count int `stack:"test"`
}

func (t *testNonStackWithStackTag) Reader() PropertyReader { return getDefaultReader(t) }
func (t *testNonStackWithStackTag) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testNonStackWithStackTag) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

type testBoardWithoutBoardTag struct {
	Spaces Board `stack:"test"`
}

func (t *testBoardWithoutBoardTag) Reader() PropertyReader { return getDefaultReader(t) }
func (t *testBoardWithoutBoardTag) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testBoardWithoutBoardTag) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}

func TestStructInflaterRejectsContradictoryStackTags(t *testing.T) {
	game := testDefaultGame(t, false)
	tests := []struct {
		name    string
		example Reader
		want    string
	}{
		{"sized field with growable tag", new(testSizedStackWithGrowableTag), "declared as SizedStack"},
		{"both concrete tags", new(testStackWithConflictingConcreteTags), "both stack and sizedstack"},
		{"both merged tags", new(testMergedStackWithConflictingTags), "both concatenate and overlap"},
		{"concrete and merged tags", new(testStackWithConcreteAndMergedTags), "mixed concrete stack tag"},
		{"stack tag on scalar", new(testNonStackWithStackTag), "not a Stack or Board property"},
		{"board missing board tag", new(testBoardWithoutBoardTag), "no board struct tag"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewStructInflater(test.example, nil, game.manager.Chest())
			if err == nil {
				t.Fatal("NewStructInflater accepted contradictory stack tags")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewStructInflater error = %q, want substring %q", err, test.want)
			}
		})
	}
}

func TestStructInflater(t *testing.T) {

	example := &testGeneralReadSetter{}

	game := testDefaultGame(t, false)

	validator, err := NewStructInflater(example, nil, game.manager.Chest())

	assert.For(t).ThatActual(err).IsNil()

	autoFilledObj := &testGeneralReadSetter{}

	err = validator.Valid(autoFilledObj)

	assert.For(t).ThatActual(err).IsNotNil()

	err = validator.Inflate(autoFilledObj, game.CurrentState())

	assert.For(t).ThatActual(err).IsNil()

	err = validator.Valid(autoFilledObj)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(autoFilledObj.EnumConst.Enum()).Equals(testColorEnum)

	assert.For(t).ThatActual(autoFilledObj.EnumVar.Enum()).Equals(testColorEnum)

	assert.For(t).ThatActual(autoFilledObj.TheTimer).IsNotNil()

	assert.For(t).ThatActual(validator.sanitizationPolicy["TheInt"]).Equals(map[string]Policy{
		"all": PolicyHidden,
	})

	assert.For(t).ThatActual(validator.sanitizationPolicy["TheGrowableStack"]).Equals(map[string]Policy{
		"all": PolicyOrder,
	})

	assert.For(t).ThatActual(validator.sanitizationPolicy["TheTimer"]).Equals(map[string]Policy{
		"all": PolicyVisible,
	})

}
