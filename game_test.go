package boardgame

import (
	"encoding/json"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

type testInfiniteLoopGameDelegate struct {
	testGameDelegate
}

func (t *testInfiniteLoopGameDelegate) ProposeFixUpMove(state State) Move {
	return t.Manager().FixUpMoveTypeByName("Test Always Legal Move").NewMove(state)
}

func TestMoveModifyDynamicValues(t *testing.T) {
	game := testGame(t)

	makeTestGameIdsStable(game)

	game.SetUp(0, nil, nil)

	drawCardMove := game.PlayerMoveByName("Draw Card")

	if drawCardMove == nil {
		t.Fatal("Couldn't find move draw card")
	}

	if err := <-game.ProposeMove(drawCardMove, ObserverPlayerIndex); err == nil {
		t.Error("Expected error proposing move from ObserverPlayerIndex")
	}

	if err := <-game.ProposeMove(drawCardMove, PlayerIndex(5)); err == nil {
		t.Error("Expected error proposing move with invalid Proposer")
	}

	if err := <-game.ProposeMove(drawCardMove, AdminPlayerIndex); err != nil {
		t.Error("Unexpected error trying to draw card: " + err.Error())
	}

	move := game.PlayerMoveByName("Increment IntValue of Card in Hand")

	if move == nil {
		t.Fatal("Couldn't find move Increment IntValue of Card in Hand")
	}

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Unexpected error trying to increment dynamic component state: " + err.Error())
	}

	//Apply the move again. This implicitly tests that deserializing a non-zero dynamic component value works.

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("unexpected error trying to increment dynamic component state a second time: ", err.Error())
	}

	gameState, playerStates := concreteStates(game.CurrentState())

	player := playerStates[gameState.CurrentPlayer]

	component := player.Hand.ComponentAt(0)

	dynamic := component.DynamicValues(game.CurrentState())

	if dynamic == nil {
		t.Error("Component unexpectedly had nil dynamic values")
	}

	easyDynamic := dynamic.(*testingComponentDynamic)

	if easyDynamic.IntVar != 7 {
		t.Error("Dynamic state of component unexpected value: ", easyDynamic.IntVar)
	}

	var stateNil *state

	assert.For(t).ThatActual(easyDynamic.Stack.state()).DoesNotEqual(stateNil)

	currentJSON, _ := json.MarshalIndent(game.CurrentState(), "", "\t")

	golden := goldenJSON("basic_state_after_dynamic_component_move.json", t)

	compareJSONObjects(currentJSON, golden, "Comparing json after two dynamic moves", t)

}

func TestProposeMoveNonModifiableGame(t *testing.T) {
	game := testGame(t)

	game.SetUp(0, nil, nil)

	manager := game.Manager()

	id := game.Id()

	//At this point, the game has stored state in storage.

	refriedGame := manager.Game(id)

	if refriedGame == nil {
		t.Fatal("Couldn't get a game out refried")
	}

	rawMove := game.PlayerMoveByName("test")

	move := rawMove.(*testMove)

	move.AString = "foo"
	move.ScoreIncrement = 3
	move.TargetPlayerIndex = 0
	move.ABool = true

	if err := <-refriedGame.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Propose move on refried game failed:", err)
	}

	//Update it from server
	refriedGame.Refresh()

	if refriedGame.Version() != 2 {
		t.Error("The proposed move didn't actually modify the underlying game in storage: ", refriedGame.Version())
	}

}

func TestGameSetUp(t *testing.T) {
	game := testGame(t)

	id := game.Id()

	if len(id) != gameIDLength {
		t.Error("Game didn't have an ID of correct length. Wanted", gameIDLength, "got", id)
	}

	if game.PlayerMoves() != nil {
		t.Error("Got moves back before SetUp was called")
	}

	if game.PlayerMoveByName("Test") != nil {
		t.Error("Move by name returned a move before SetUp was called")
	}

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	originalTestMove := move

	delayedError := game.ProposeMove(move, AdminPlayerIndex)

	select {
	case <-delayedError:
		//Good!
	case <-time.After(time.Millisecond * 5):
		t.Error("We never got an error from proposing a move on a game that hadn't even started")
	}

	if err := game.SetUp(15, nil, nil); err == nil {
		t.Error("Calling set up with an illegal number of players didn't fail")
	}

	if err := game.SetUp(2, GameConfig{"color": "illegal"}, nil); err == nil {
		t.Error("Calling game set up with an illegal config did not fail")
	}

	if err := game.SetUp(-5, nil, nil); err == nil {
		t.Error("Calling set up with negative number of players didn't fail")
	}

	if err := game.SetUp(3, nil, []string{"", "bam"}); err == nil {
		t.Error("Calling set up with wrong-sized agent config didn't fail")
	}

	//TODO: we no longer test that SetUp calls the Component distribution logic.

	//Blue is a legal color according to our delegate's Configs()
	if err := game.SetUp(0, GameConfig{"color": "blue"}, nil); err != nil {
		t.Error("Calling SetUp on a previously errored game did not succeed", err)
	}

	if wrapper, err := game.Manager().Storage().State(game.Id(), 0); wrapper == nil {
		t.Error("State 0 was not saved in storage when game set up")
	} else if err != nil {
		t.Error("Storing state 0 failed: " + err.Error())
	}

	if game.CurrentState() == nil {
		t.Error("Game had no current state after saving")
	}

	if game.PlayerMoveByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if game.PlayerMoveByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

	if originalTestMove == game.PlayerMoveByName("Test") {
		t.Error("MoveByName returned a non-copy")
	}

	//Test to verify that game has stack's state property set
	currentState := game.CurrentState()
	gameState, playerStates := concreteStates(currentState)

	if gameState.DrawDeck.state() != currentState {
		t.Error("GameState's drawdeck didn't have state set correctly. Got", gameState.DrawDeck.state(), "wanted", currentState)
	}

	if playerStates[0].Hand.state() != currentState {
		t.Error("PlayerStates Hand didn't have state set correctly. Got", playerStates[0].Hand.state(), "wanted", currentState)
	}

	deck := game.Chest().Deck("test").Components()

	if gameState.DrawDeck.Len() != len(deck) {
		t.Error("All of the components were not distributed in SetUp")
	}

	stateCopy := currentState.Copy(false)

	gameState, playerStates = concreteStates(stateCopy)

	if gameState.DrawDeck.state() != stateCopy {
		t.Error("The copy of state's stacks had the old state in gamestate")
	}

	if playerStates[0].Hand.state() != stateCopy {
		t.Error("The copy of state's stacks had the old state in playerstate")
	}

}

func TestApplyMove(t *testing.T) {
	game := testGame(t)

	makeTestGameIdsStable(game)

	game.SetUp(0, nil, nil)

	rawMove := game.PlayerMoveByName("test")

	move := rawMove.(*testMove)

	move.AString = "foo"
	move.ScoreIncrement = 3
	move.TargetPlayerIndex = 0
	move.ABool = true

	manager := game.Manager()

	oldMoves := manager.playerMoves
	oldMovesByName := manager.playerMovesByName

	manager.playerMoves = nil
	manager.playerMovesByName = make(map[string]*MoveType)

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err == nil {
		t.Error("Game allowed a move that wasn't configured as part of game to be applied")
	}

	manager.playerMoves = oldMoves
	manager.playerMovesByName = oldMovesByName

	//testMove checks to make sure game.state.currentPlayerIndex is targetplayerindex

	move.TargetPlayerIndex = 1

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err == nil {
		t.Error("Game allowed a move to be applied where the wrong playe was current")
	}

	move.TargetPlayerIndex = 0

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Game didn't allow a legal move to be made")
	}

	//Verify that the move was made. Note that because our Delegate has a
	//FixUp move, this is also testing that not just the main move, but also
	//the fixup move was made.

	record, err := game.Manager().Storage().State(game.Id(), game.Version())

	if err != nil {
		t.Error("Unexpected error", err)
	}

	wrapper, err := game.Manager().stateFromRecord(record)

	if err != nil {
		t.Error("Error state from from blob", err)
	}

	wrapper.game = game

	currentJson, _ := json.Marshal(wrapper)
	golden := goldenJSON("basic_state_after_move.json", t)

	compareJSONObjects(currentJson, golden, "Basic state after test move", t)

	//Apply a move that should finish the game (any player has score > 5)
	newRawMove := game.PlayerMoveByName("test")

	newMove := newRawMove.(*testMove)

	newMove.AString = "foo"
	newMove.ScoreIncrement = 6
	newMove.TargetPlayerIndex = 1
	newMove.ABool = true

	if err := <-game.ProposeMove(newMove, AdminPlayerIndex); err != nil {
		t.Error("Game didn't allow a move to be made even though it was legal: ", err)
	}

	if wrapper, _ := game.Manager().Storage().State(game.Id(), 1); wrapper == nil {
		t.Error("We didn't get back state for state 1; game must not be persisting states to DB.")
	}

	//By the time err has resolved above, any fixup moves have been applied.

	if !game.Finished() {
		t.Error("Game didn't notice that a user had won")
	}

	if !reflect.DeepEqual(game.Winners(), []PlayerIndex{PlayerIndex(1)}) {
		t.Error("Game thought the wrong players had won")
	}

	moveAfterFinished := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 2,
		ABool:             true,
	}

	if err := <-game.ProposeMove(moveAfterFinished, AdminPlayerIndex); err == nil {
		t.Error("Game allowed a move to be applied after the game was finished")
	}
}

func TestMoveRoundTrip(t *testing.T) {
	game := testGame(t)

	err := game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	move := game.PlayerMoveByName("test")

	testMove := move.(*testMove)

	testMove.AString = "foo"
	testMove.ScoreIncrement = 3
	testMove.TargetPlayerIndex = 0
	testMove.ABool = true

	err = <-game.ProposeMove(move, AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(testMove.Info().Timestamp().IsZero()).IsFalse()

	refriedMove, err := game.Move(1)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(move).Equals(refriedMove)

	assert.For(t).ThatActual(refriedMove.Info().Timestamp()).Equals(move.Info().Timestamp())

	fixUpMove, err := game.Move(2)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(fixUpMove.Info().Initiator()).Equals(1)

}

func TestIllegalMove(t *testing.T) {

	manager := newTestGameManger(t)

	_, err := (&testIllegalMoveConfig).NewMoveType(manager)

	assert.For(t).ThatActual(err).IsNotNil()

}

func TestInfiniteProposeFixUp(t *testing.T) {
	//This test makes sure that if our GameDelegate is going to always return
	//moves that are legal, we'll bail at a certain point.

	moveInstaller := func(manager *GameManager) *MoveTypeConfigBundle {
		bundle := NewMoveTypeConfigBundle()
		bundle.AddMoves(
			&testMoveConfig,
			&testAlwaysLegalMoveConfig,
		)
		return bundle
	}

	delegate := &testInfiniteLoopGameDelegate{
		testGameDelegate{
			moveInstaller: moveInstaller,
		},
	}

	manager, err := NewGameManager(delegate, newTestGameChest(), newTestStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	game := manager.NewGame()

	err = game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).Equals(ErrTooManyFixUps)

}

func TestIllegalPlayerIndex(t *testing.T) {
	game := testGame(t)

	game.SetUp(2, nil, nil)

	previousVersion := game.Version()

	move := game.FixUpMoveByName("Invalid PlayerIndex")

	assert.For(t).ThatActual(move).IsNotNil()

	move.(*testMoveInvalidPlayerIndex).CurrentlyLegal = true

	err := <-game.ProposeMove(move, AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(game.Version()).Equals(previousVersion)

}

func TestAgent(t *testing.T) {
	game := testGame(t)

	game.instantAgentMoves = true

	assert.For(t).ThatActual(game.NumAgentPlayers()).Equals(0)

	err := game.SetUp(3, nil, []string{"", "Test", "Test"})

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(game.Version()).Equals(0)

	assert.For(t).ThatActual(game.NumAgentPlayers()).Equals(2)

	move := game.PlayerMoveByName("Test")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, 0)

	assert.For(t).ThatActual(err).IsNil()

	//After we make that move, the next two players will make moves and it
	//will advance back to main player.

	<-time.After(time.Millisecond * 50)

	assert.For(t).ThatActual(game.Version()).Equals(6)

	gameState, _ := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.CurrentPlayer).Equals(PlayerIndex(0))

}

func TestGameSalt(t *testing.T) {
	game := testGame(t)

	game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(game.secretSalt).DoesNotEqual("")

	refriedGame := game.Manager().Game(game.Id())

	if !assert.For(t).ThatActual(refriedGame).IsNotNil().Passed() {
		t.FailNow()
	}

	assert.For(t).ThatActual(game.secretSalt).Equals(refriedGame.secretSalt)

	mainC := game.Chest().Deck("test").ComponentAt(0)
	refriedC := refriedGame.Chest().Deck("test").ComponentAt(0)

	mainCId := mainC.ID(game.CurrentState())

	assert.For(t).ThatActual(mainCId).DoesNotEqual("")
	assert.For(t).ThatActual(mainCId).Equals(refriedC.ID(refriedGame.CurrentState()))

	otherGame := testGame(t)

	otherGame.SetUp(0, nil, nil)

	otherC := otherGame.Chest().Deck("test").ComponentAt(0)

	assert.For(t).ThatActual(mainCId).DoesNotEqual(otherC.ID(otherGame.CurrentState()))
}

func TestGameState(t *testing.T) {
	game := testGame(t)

	makeTestGameIdsStable(game)

	game.SetUp(0, nil, nil)

	if game.Name() != testGameName {
		t.Error("Game name was not correct")
	}

	blob, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		t.Error("Json marshal of game failed:", err)
	}

	goldenBlob, err := ioutil.ReadFile("test/game_blob.json")

	if err != nil {
		t.Error("Couldn't load golden file", err)
	}

	compareJSONObjects(blob, goldenBlob, "Sanity checking game json", t)

	//Getting this now helps verify that we invalidate currentState cache when
	//we apply a move.
	state := game.CurrentState()

	state0 := game.State(0)

	//This is lame, but when you create json for a State, it touches Computed,
	//which will make it non-nil, so if you're doing direct comparison they
	//won't compare equal even though they basically are. At this point
	//CurrentState has already been touched above by creating a json blob. So
	//just touch state0, too. ¯\_(ツ)_/¯
	_, _ = json.Marshal(state0)

	assert.For(t).ThatActual(state).Equals(state0).ThenDiffOnFail()

	move := game.PlayerMoveByName("Test")

	if move == nil {
		t.Fatal("Couldn't find a move to make")
	}

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Couldn't make move")
	}

	state = game.State(-1)

	if state != nil {
		t.Error("Returned a state for a non-sensiscal version -1", state)
	}

	state = game.State(game.Version() + 1)

	if state != nil {
		t.Error("Returned a state for a too-high version", state)
	}

	currentState := game.CurrentState()
	state = game.State(game.Version())

	if !reflect.DeepEqual(currentState, state) {
		t.Error("State(game.Version()) and CurrentState() weren't equivalent", currentState, state)

	}

}
