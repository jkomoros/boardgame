package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	jd "github.com/josephburnett/jd/lib"
	"github.com/workfit/tester/assert"
)

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
				StateGroupGame,
				"",
				statePropertyRefDefaultIndex,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing game property",
			StatePropertyRef{
				StateGroupGame,
				"MyIntSlice",
				statePropertyRefDefaultIndex,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			false,
		},
		{
			"Basic existing player property",
			StatePropertyRef{
				StateGroupPlayer,
				"IsFoo",
				statePropertyRefDefaultIndex,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			false,
		},
		{
			"Basic existing dynamic component values property",
			StatePropertyRef{
				StateGroupDynamicComponentValues,
				"IntVar",
				statePropertyRefDefaultIndex,
				"test",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			false,
		},
		{
			"Basic existing game property that doesn't exist",
			StatePropertyRef{
				StateGroupGame,
				"NonExistent",
				statePropertyRefDefaultIndex,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing player property that doesn't exist",
			StatePropertyRef{
				StateGroupPlayer,
				"Nonexistent",
				statePropertyRefDefaultIndex,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing dynamic component values property that doesn't exist",
			StatePropertyRef{
				StateGroupDynamicComponentValues,
				"Nonexistent	",
				statePropertyRefDefaultIndex,
				"test",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing dynamic component values property missing deck",
			StatePropertyRef{
				StateGroupDynamicComponentValues,
				"IntVar",
				statePropertyRefDefaultIndex,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing dynamic component values property invalid deck",
			StatePropertyRef{
				StateGroupDynamicComponentValues,
				"IntVar",
				statePropertyRefDefaultIndex,
				"invaliddeckname",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing player property negative player index",
			StatePropertyRef{
				StateGroupPlayer,
				"IsFoo",
				-2,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Basic existing player property too-high player index",
			StatePropertyRef{
				StateGroupPlayer,
				"IsFoo",
				10,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Game property with player index set",
			StatePropertyRef{
				StateGroupGame,
				"DrawDeck",
				2,
				"",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Game property with dynamic component deck set",
			StatePropertyRef{
				StateGroupGame,
				"DrawDeck",
				statePropertyRefDefaultIndex,
				"nonempty",
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Game property with stack and too-low stack index",
			StatePropertyRef{
				StateGroupGame,
				"DrawDeck",
				statePropertyRefDefaultIndex,
				"",
				-10,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
		{
			"Game property with stack and too-high stack index",
			StatePropertyRef{
				StateGroupGame,
				"DrawDeck",
				statePropertyRefDefaultIndex,
				"",
				1000,
				statePropertyRefDefaultIndex,
				statePropertyRefDefaultIndex,
			},
			true,
		},
	}

	for i, test := range tests {
		err := test.ref.Validate(state)
		if test.errExpected {
			assert.For(t, i).ThatActual(err).IsNotNil()
		} else {
			assert.For(t, i).ThatActual(err).IsNil()
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
	}

	for i, test := range tests {
		result := test.p.Next(state)
		assert.For(t, "next", i).ThatActual(result).Equals(test.expectedNext)

		result = test.p.Previous(state)

		assert.For(t, "prev", i).ThatActual(result).Equals(test.expectedPrev)
	}
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
			AdminPlayerIndex - 1,
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

	state, err := game.Manager().stateFromRecord(record)
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

	reconstitutedState, err := game.Manager().stateFromRecord(blob)

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

	//Can't do ThenDiffOnFail because components and sub-states now have references back to state (loops)
	assert.For(t).ThatActual(reconstitutedState).Equals(game.CurrentState())

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
