package boardgame

import (
	"encoding/json"
	stderrors "errors"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	jd "github.com/josephburnett/jd/lib"
	"github.com/workfit/tester/assert"
)

// TagTestBehavior is a minimal TagConfigurable implementation for testing the
// autoConnectBehaviors tag-forwarding machinery. Exported so that
// CanInterface() returns true (matching how real behaviors work).
type TagTestBehavior struct {
	Configured bool
	TagValue   string
}

func (b *TagTestBehavior) ConfigureFromTags(tags reflect.StructTag, containingSubState SubState) error {
	b.TagValue = tags.Get("testcfg")
	if b.TagValue == "error" {
		return stderrors.New("forced test error")
	}
	if b.TagValue != "" {
		b.Configured = true
	}
	return nil
}

// tagConfigTestState is a minimal SubState embedding TagTestBehavior with a tag.
type tagConfigTestState struct {
	state           State
	ref             StatePropertyRef
	TagTestBehavior `testcfg:"hello"`
	Score           int
}

func (t *tagConfigTestState) Reader() PropertyReader         { return getDefaultReader(t) }
func (t *tagConfigTestState) ReadSetter() PropertyReadSetter { return getDefaultReadSetter(t) }
func (t *tagConfigTestState) ConnectContainingState(state State, ref StatePropertyRef) {
	t.state = state
	t.ref = ref
}
func (t *tagConfigTestState) FinishStateSetUp()                  {}
func (t *tagConfigTestState) State() State                       { return t.state }
func (t *tagConfigTestState) ImmutableState() ImmutableState     { return t.state }
func (t *tagConfigTestState) StatePropertyRef() StatePropertyRef { return t.ref }

// tagConfigTestStateNoTag embeds TagTestBehavior without a struct tag.
type tagConfigTestStateNoTag struct {
	state           State
	ref             StatePropertyRef
	TagTestBehavior // no struct tag
	Score           int
}

func (t *tagConfigTestStateNoTag) Reader() PropertyReader         { return getDefaultReader(t) }
func (t *tagConfigTestStateNoTag) ReadSetter() PropertyReadSetter { return getDefaultReadSetter(t) }
func (t *tagConfigTestStateNoTag) ConnectContainingState(state State, ref StatePropertyRef) {
	t.state = state
	t.ref = ref
}
func (t *tagConfigTestStateNoTag) FinishStateSetUp()                  {}
func (t *tagConfigTestStateNoTag) State() State                       { return t.state }
func (t *tagConfigTestStateNoTag) ImmutableState() ImmutableState     { return t.state }
func (t *tagConfigTestStateNoTag) StatePropertyRef() StatePropertyRef { return t.ref }

// tagConfigTestStateError embeds TagTestBehavior with a tag that triggers an error.
type tagConfigTestStateError struct {
	state           State
	ref             StatePropertyRef
	TagTestBehavior `testcfg:"error"`
	Score           int
}

func (t *tagConfigTestStateError) Reader() PropertyReader         { return getDefaultReader(t) }
func (t *tagConfigTestStateError) ReadSetter() PropertyReadSetter { return getDefaultReadSetter(t) }
func (t *tagConfigTestStateError) ConnectContainingState(state State, ref StatePropertyRef) {
	t.state = state
	t.ref = ref
}
func (t *tagConfigTestStateError) FinishStateSetUp()                  {}
func (t *tagConfigTestStateError) State() State                       { return t.state }
func (t *tagConfigTestStateError) ImmutableState() ImmutableState     { return t.state }
func (t *tagConfigTestStateError) StatePropertyRef() StatePropertyRef { return t.ref }

func TestAutoConnectBehaviorsTagConfigurable(t *testing.T) {
	// Happy path: tag present and ConfigureFromTags receives it
	s := &tagConfigTestState{}
	err := autoConnectBehaviors(s)
	assert.For(t, "happy path err").ThatActual(err).IsNil()
	assert.For(t, "configured").ThatActual(s.TagTestBehavior.Configured).IsTrue()
	assert.For(t, "tag value").ThatActual(s.TagTestBehavior.TagValue).Equals("hello")

	// No tag: ConfigureFromTags is called but tag is empty, so no-op
	sNoTag := &tagConfigTestStateNoTag{}
	err = autoConnectBehaviors(sNoTag)
	assert.For(t, "no tag err").ThatActual(err).IsNil()
	assert.For(t, "not configured").ThatActual(sNoTag.TagTestBehavior.Configured).IsFalse()

	// Error path: ConfigureFromTags returns an error, which is propagated
	sErr := &tagConfigTestStateError{}
	err = autoConnectBehaviors(sErr)
	assert.For(t, "error propagated").ThatActual(err).IsNotNil()
	assert.For(t, "error contains field name").ThatActual(strings.Contains(err.Error(), "TagTestBehavior")).IsTrue()
	assert.For(t, "error contains cause").ThatActual(strings.Contains(err.Error(), "forced test error")).IsTrue()
}

func TestPlayerIndex(t *testing.T) {
	game := testDefaultGame(t, false)

	state := game.CurrentState()

	for i := 0; i < game.NumPlayers(); i++ {
		pState := state.ImmutablePlayerStates()[i]
		assert.For(t, i).ThatActual(pState.StatePropertyRef().PlayerIndex).Equals(PlayerIndex(i))
	}
}

func TestStatePropertyRefValidate(t *testing.T) {
	game := testDefaultGame(t, false)

	state := game.CurrentState()

	tests := []struct {
		description string
		ref         StatePropertyRef
		errExpected bool
	}{
		{
			"No prop name",
			StatePropertyRef{
				Group: StateGroupGame,
			},
			false,
		},
		{
			"Basic existing game property",
			StatePropertyRef{
				Group:    StateGroupGame,
				PropName: "MyIntSlice",
			},
			false,
		},
		{
			"Basic existing player property",
			StatePropertyRef{
				Group:    StateGroupPlayer,
				PropName: "IsFoo",
			},
			false,
		},
		{
			"Basic existing dynamic component values property",
			StatePropertyRef{
				Group:    StateGroupDynamicComponentValues,
				PropName: "IntVar",
				DeckName: "test",
			},
			false,
		},
		{
			"Basic existing component values property",
			StatePropertyRef{
				Group:    StateGroupComponentValues,
				PropName: "Integer",
				DeckName: "test",
			},
			false,
		},
		{
			"Basic existing game property that doesn't exist",
			StatePropertyRef{
				Group:    StateGroupGame,
				PropName: "NonExistent",
			},
			true,
		},
		{
			"Basic existing player property that doesn't exist",
			StatePropertyRef{
				Group:    StateGroupPlayer,
				PropName: "Nonexistent",
			},
			true,
		},
		{
			"Basic existing dynamic component values property that doesn't exist",
			StatePropertyRef{
				Group:    StateGroupDynamicComponentValues,
				PropName: "Nonexistent",
				DeckName: "test",
			},
			true,
		},
		{
			"Basic existing component values property that doesn't exist",
			StatePropertyRef{
				Group:    StateGroupComponentValues,
				PropName: "Nonexistent",
				DeckName: "test",
			},
			true,
		},
		{
			"Basic existing dynamic component values property missing deck",
			StatePropertyRef{
				Group:    StateGroupDynamicComponentValues,
				PropName: "IntVar",
			},
			true,
		},
		{
			"Basic existing component values property missing deck",
			StatePropertyRef{
				Group:    StateGroupComponentValues,
				PropName: "Integer",
			},
			true,
		},
		{
			"Basic existing dynamic component values property invalid deck",
			StatePropertyRef{
				Group:    StateGroupDynamicComponentValues,
				PropName: "IntVar",
				DeckName: "invaliddeckname",
			},
			true,
		},
		{
			"Basic existing component values property invalid deck",
			StatePropertyRef{
				Group:    StateGroupComponentValues,
				PropName: "Integer",
				DeckName: "invaliddeckname",
			},
			true,
		},
		{
			"Basic existing dynamic component values property invalid index",
			StatePropertyRef{
				Group:     StateGroupDynamicComponentValues,
				PropName:  "IntVar",
				DeckName:  "test",
				DeckIndex: 1000,
			},
			true,
		},
		{
			"Basic existing component values property invalid index",
			StatePropertyRef{
				Group:     StateGroupComponentValues,
				PropName:  "Integer",
				DeckName:  "test",
				DeckIndex: 1000,
			},
			true,
		},
		{
			"Basic existing player property negative player index",
			StatePropertyRef{
				Group:       StateGroupPlayer,
				PropName:    "IsFoo",
				PlayerIndex: -2,
			},
			true,
		},
		{
			"Basic existing player property too-high player index",
			StatePropertyRef{
				Group:       StateGroupPlayer,
				PropName:    "IsFoo",
				PlayerIndex: 10,
			},
			true,
		},
		{
			"Game property with player index set",
			StatePropertyRef{
				Group:       StateGroupGame,
				PropName:    "DrawDeck",
				PlayerIndex: 2,
			},
			true,
		},
		{
			"Game property with dynamic component deck set",
			StatePropertyRef{
				Group:    StateGroupGame,
				PropName: "DrawDeck",
				DeckName: "nonempty",
			},
			true,
		},
	}

	for i, test := range tests {
		err := test.ref.Validate(state)
		if test.errExpected {
			assert.For(t, i, test.description).ThatActual(err).IsNotNil()
		} else {
			assert.For(t, i, test.description).ThatActual(err).IsNil()
		}
	}
}

func TestPlayerIndexNextPrevious(t *testing.T) {

	game := testGame(t, false, 3, nil, nil)

	state := game.CurrentState()

	tests := []struct {
		p            PlayerIndex
		expectedNext PlayerIndex
		expectedPrev PlayerIndex
	}{
		{
			0,
			1,
			2,
		},
		{
			2,
			0,
			1,
		},
		{
			AdminPlayerIndex,
			AdminPlayerIndex,
			AdminPlayerIndex,
		},
		{
			ObserverPlayerIndex,
			ObserverPlayerIndex,
			ObserverPlayerIndex,
		},
		{
			AnyPlayerIndex,
			AnyPlayerIndex,
			AnyPlayerIndex,
		},
	}

	for i, test := range tests {
		result := test.p.Next(state)
		assert.For(t, "next", i).ThatActual(result).Equals(test.expectedNext)

		result = test.p.Previous(state)

		assert.For(t, "prev", i).ThatActual(result).Equals(test.expectedPrev)
	}
}

func TestPlayerIndexNextPreviousCustomOrder(t *testing.T) {
	game := testGame(t, false, 3, nil, nil)
	state := game.CurrentState()

	delegate := game.Manager().Delegate().(*testGameDelegate)

	// Custom order: 2, 0, 1 (reversed first player, shifted)
	delegate.customPlayerOrder = []PlayerIndex{2, 0, 1}

	tests := []struct {
		description  string
		p            PlayerIndex
		expectedNext PlayerIndex
		expectedPrev PlayerIndex
	}{
		{
			"Player 2 is first in custom order, next is 0",
			2,
			0,
			1,
		},
		{
			"Player 0 is second in custom order, next is 1",
			0,
			1,
			2,
		},
		{
			"Player 1 is last in custom order, wraps to 2",
			1,
			2,
			0,
		},
		{
			"Special indices are unaffected by custom order",
			AdminPlayerIndex,
			AdminPlayerIndex,
			AdminPlayerIndex,
		},
		{
			"Observer is unaffected by custom order",
			ObserverPlayerIndex,
			ObserverPlayerIndex,
			ObserverPlayerIndex,
		},
	}

	for i, test := range tests {
		result := test.p.Next(state)
		assert.For(t, "custom next", i, test.description).ThatActual(result).Equals(test.expectedNext)

		result = test.p.Previous(state)
		assert.For(t, "custom prev", i, test.description).ThatActual(result).Equals(test.expectedPrev)
	}

	// Test: player not found in order falls back to same index
	delegate.customPlayerOrder = []PlayerIndex{0, 1} // only 2 entries for 3 players
	result := PlayerIndex(2).Next(state)
	// Player 2 not in order, should return self
	assert.For(t, "not in order next").ThatActual(result).Equals(PlayerIndex(2))

	result = PlayerIndex(2).Previous(state)
	assert.For(t, "not in order prev").ThatActual(result).Equals(PlayerIndex(2))

	// Test: skipping inactive players in custom order
	delegate.customPlayerOrder = []PlayerIndex{2, 0, 1}
	delegate.inactivePlayers = map[PlayerIndex]bool{0: true} // player 0 is inactive

	// Next from 2 should skip inactive 0 and land on 1
	result = PlayerIndex(2).Next(state)
	assert.For(t, "skip inactive next").ThatActual(result).Equals(PlayerIndex(1))

	// Previous from 1 should skip inactive 0 and land on 2
	result = PlayerIndex(1).Previous(state)
	assert.For(t, "skip inactive prev").ThatActual(result).Equals(PlayerIndex(2))

	// Test: multiple inactive players, only one active
	delegate.inactivePlayers = map[PlayerIndex]bool{0: true, 1: true} // only player 2 active

	result = PlayerIndex(2).Next(state)
	assert.For(t, "only one active next").ThatActual(result).Equals(PlayerIndex(2))

	result = PlayerIndex(2).Previous(state)
	assert.For(t, "only one active prev").ThatActual(result).Equals(PlayerIndex(2))

	// Test: all players inactive returns self
	delegate.inactivePlayers = map[PlayerIndex]bool{0: true, 1: true, 2: true}

	result = PlayerIndex(0).Next(state)
	assert.For(t, "all inactive next").ThatActual(result).Equals(PlayerIndex(0))

	result = PlayerIndex(0).Previous(state)
	assert.For(t, "all inactive prev").ThatActual(result).Equals(PlayerIndex(0))

	// Reset for clean state
	delegate.customPlayerOrder = nil
	delegate.inactivePlayers = nil
}

func TestPlayerIndexValid(t *testing.T) {

	gameThreePlayers := testGame(t, false, 3, nil, nil)

	stateThreePlayers := gameThreePlayers.CurrentState()

	tests := []struct {
		p        PlayerIndex
		state    ImmutableState
		expected bool
	}{
		{
			0,
			stateThreePlayers,
			true,
		},
		{
			ObserverPlayerIndex,
			stateThreePlayers,
			true,
		},
		{
			AdminPlayerIndex,
			stateThreePlayers,
			true,
		},
		{
			AnyPlayerIndex,
			stateThreePlayers,
			true,
		},
		{
			AnyPlayerIndex - 1,
			stateThreePlayers,
			false,
		},
		{
			4,
			stateThreePlayers,
			false,
		},
		{
			3,
			stateThreePlayers,
			false,
		},
	}

	for i, test := range tests {
		result := test.p.Valid(test.state)
		assert.For(t, "valid", i).ThatActual(result).Equals(test.expected)
	}
}

func TestPlayerIndexEquivalent(t *testing.T) {

	equivalentTests := []struct {
		p        PlayerIndex
		other    PlayerIndex
		expected bool
	}{
		{
			0,
			0,
			true,
		},
		{
			0,
			1,
			false,
		},
		{
			AdminPlayerIndex,
			0,
			true,
		},
		{
			AdminPlayerIndex,
			ObserverPlayerIndex,
			false,
		},
		{
			ObserverPlayerIndex,
			1,
			false,
		},
		{
			0,
			AdminPlayerIndex,
			true,
		},
		{
			AdminPlayerIndex,
			AdminPlayerIndex,
			true,
		},
		{
			ObserverPlayerIndex,
			ObserverPlayerIndex,
			false,
		},
		{
			AnyPlayerIndex,
			0,
			true,
		},
		{
			0,
			AnyPlayerIndex,
			true,
		},
		{
			AnyPlayerIndex,
			AnyPlayerIndex,
			true,
		},
		{
			AnyPlayerIndex,
			ObserverPlayerIndex,
			false,
		},
		{
			AnyPlayerIndex,
			AdminPlayerIndex,
			true,
		},
		{
			AdminPlayerIndex,
			AnyPlayerIndex,
			true,
		},
	}

	for i, test := range equivalentTests {
		result := test.p.Equivalent(test.other)

		assert.For(t, "equivalent", i).ThatActual(result).Equals(test.expected)
	}
}

func TestSecretMoveCount(t *testing.T) {

	game := testDefaultGame(t, true)

	currentState := game.CurrentState()

	assert.For(t).ThatActual(currentState.Version()).Equals(game.Version())

	gameState, _ := concreteStates(currentState)

	s := currentState.(*state)

	for i, c := range gameState.DrawDeck.Components() {
		assert.For(t, i).ThatActual(c.secretMoveCount()).Equals(0)
	}

	idBefore := gameState.DrawDeck.ComponentAt(0).ID()

	gameState.DrawDeck.ComponentAt(0).movedSecretly()

	assert.For(t).ThatActual(gameState.DrawDeck.ComponentAt(0).ID()).DoesNotEqual(idBefore)

	for i, c := range gameState.DrawDeck.Components() {
		if i == 0 {
			assert.For(t, i).ThatActual(c.secretMoveCount()).Equals(1)
		} else {
			assert.For(t, i).ThatActual(c.secretMoveCount()).Equals(0)
		}
	}

	//We're going to do a faked save to verify that these things round trip
	game.version++

	blob, err := json.MarshalIndent(s, "", "\t")

	assert.For(t).ThatActual(err).IsNil()

	game.manager.Storage().SaveGameAndCurrentState(game.StorageRecord(), blob, nil)

	//Read back in the game and verify that the secretMoveCount round-tripped.

	refriedGame := game.manager.Game(game.ID())

	refriedGameState, _ := concreteStates(refriedGame.CurrentState())

	for i, c := range refriedGameState.DrawDeck.Components() {
		if i == 0 {
			assert.For(t, i).ThatActual(c.secretMoveCount()).Equals(1)
		} else {
			assert.For(t, i).ThatActual(c.secretMoveCount()).Equals(0)
		}
	}

}

func testSubStatesHaveStateSet(t *testing.T, state *state) {
	assert.For(t).ThatActual(state.GameState().(*testGameState).state).Equals(state)

	for i := 0; i < len(state.playerStates); i++ {
		assert.For(t, i).ThatActual(state.PlayerStates()[i].(*testPlayerState).state).Equals(state)
	}

	for _, dynamicComponents := range state.DynamicComponentValues() {
		for i, component := range dynamicComponents {
			assert.For(t, i).ThatActual(component.(*testingComponentDynamic).state).Equals(state)
		}
	}
}

func TestRand(t *testing.T) {
	game := testDefaultGame(t, false)

	//Make there be more than one state
	err := <-game.ProposeMove(game.Moves()[0], AdminPlayerIndex)
	assert.For(t).ThatActual(err).IsNil()

	zeroState := game.State(0).(State)

	r := zeroState.Rand()

	assert.For(t).ThatActual(r).IsNotNil()

	//Test Rand() returns same object
	assert.For(t).ThatActual(zeroState.Rand()).Equals(r)

	first := r.Int()
	second := r.Int()

	zeroStateCopy := game.State(0).(State)

	copyR := zeroStateCopy.Rand()

	//Make sure different state gives different object
	assert.For(t).ThatActual(copyR).DoesNotEqual(r)

	//Make sure new object gives save values
	assert.For(t).ThatActual(copyR.Int()).Equals(first)
	assert.For(t).ThatActual(copyR.Int()).Equals(second)

	oneState := game.State(1).(State)

	//Make sure a different version gives a different value.
	assert.For(t).ThatActual(oneState.Rand().Int()).DoesNotEqual(first)

	//Make sure different game same version has different value.
	otherGame := testDefaultGame(t, false)
	otherGameState := otherGame.State(0).(State)
	assert.For(t).ThatActual(otherGameState.Rand().Int()).DoesNotEqual(first)

}

func TestState(t *testing.T) {

	game := testDefaultGame(t, true)

	assert.For(t).ThatActual(game.CurrentState().Version()).Equals(game.Version())

	theState := game.CurrentState().(*state)

	testSubStatesHaveStateSet(t, theState)

	theStateCopy, err := theState.Copy(false)

	assert.For(t).ThatActual(err).IsNil()

	testSubStatesHaveStateSet(t, theStateCopy.(*state))

	record, err := game.Manager().Storage().State(game.ID(), game.Version())

	if err != nil {
		t.Error("Unexpected error", err)
	}

	state, err := game.Manager().stateFromRecord(record, game.Version())
	state.game = game

	if err != nil {
		t.Error("StateFromBlob err", err)
	}

	if state == nil {
		t.Error("State could not be created")
	}

	assert.For(t).ThatActual(state.Version()).Equals(game.Version())

	testSubStatesHaveStateSet(t, state)

	currentJSON, _ := json.Marshal(state)
	golden := goldenJSON("base.json", t)

	compareJSONObjects(currentJSON, golden, "Basic state", t)

	stateCopy, err := state.Copy(false)

	assert.For(t).ThatActual(err).IsNil()

	copyJSON, _ := DefaultMarshalJSON(stateCopy)

	compareJSONObjects(copyJSON, currentJSON, "Copy was not same", t)

	_, playerStatesCopy := concreteStates(stateCopy)

	playerStatesCopy[0].MovesLeftThisTurn = 10

	_, playerStates := concreteStates(state)

	if playerStates[0].MovesLeftThisTurn == 10 {
		t.Error("Modifying a copy change the original")
	}

	if state.Sanitized() {
		t.Error("State reported being sanitized even when it wasn't")
	}

	sanitizedStateCopy, err := stateCopy.Copy(true)

	assert.For(t).ThatActual(err).IsNil()

	if !sanitizedStateCopy.Sanitized() {
		t.Error("A copy that was told it was sanitized did not report being sanitized.")
	}

	//TODO: test that GAmeState and UserStates are also copies
}

func TestStateSerialization(t *testing.T) {

	game := testDefaultGame(t, false)

	gameState, _ := concreteStates(game.CurrentState())

	if gameState.Timer.state() == nil {
		t.Error("The set up timer did no thave a stateptr")
	}

	rawMove := game.MoveByName("test")

	move := rawMove.(*testMove)

	move.AString = "bam"
	move.ScoreIncrement = 3
	move.TargetPlayerIndex = 0
	move.ABool = true

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Fatal("Couldn't make move", err)
	}

	blob, err := json.Marshal(game.CurrentState())

	if err != nil {
		t.Fatal("Couldn't serialize state:", err)
	}

	reconstitutedState, err := game.Manager().stateFromRecord(blob, game.Version())

	if err != nil {
		t.Error("StateFromBlob returned unexpected err", err)
	}

	reconstitutedState.game = game

	gameState, _ = concreteStates(reconstitutedState)

	if gameState.DrawDeck.ComponentAt(0).DynamicValues().(*testingComponentDynamic).Stack.Deck() == nil {
		t.Error("The stack on a component's dynamic value was not inflated coming back from storage.")
	}

	if gameState.Timer.state() == nil {
		t.Error("The timer did not come back inflated from storage")
	}

	//This is lame, but when you create json for a State, it touches Computed,
	//which will make it non-nil, so if you're doing direct comparison they
	//won't compare equal even though they basically are. At this point
	//CurrentState has already been touched above by creating a json blob. So
	//just touch reconstitutedState, too. ¯\_(ツ)_/¯

	_, _ = json.Marshal(reconstitutedState)

	assertPersistedStatesEqual(t, reconstitutedState, game.CurrentState())

}

func compareJSONObjects(in []byte, golden []byte, message string, t *testing.T) {

	//recreated in server/internal/teststoragemanager

	inJSON, err := jd.ReadJsonString(string(in))

	if err != nil {
		t.Fatal(message + ": Couldn't read json in: " + err.Error())
	}

	goldenJSON, err := jd.ReadJsonString(string(golden))

	if err != nil {
		t.Fatal(message + ": Couldn't read json golden: " + err.Error())
	}

	diff := goldenJSON.Diff(inJSON)

	if len(diff) == 0 {
		return
	}

	t.Error(message + ": JSON comparison failed: " + diff.Render())

}

func goldenJSON(fileName string, t *testing.T) []byte {

	contents, err := ioutil.ReadFile("./testdata/" + fileName)

	if !assert.For(t, fileName).ThatActual(err).IsNil().Passed() {
		t.FailNow()
	}

	return contents

}
