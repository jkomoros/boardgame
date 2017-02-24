package boardgame

import (
	"errors"
)

//Place to define testing structs and helpers that are useful throughout

//testingComponent is a very basic thing that fufills the Component interface.
type testingComponent struct {
	String  string
	Integer int
}

const testGameName = "Test Game"

func (t *testingComponent) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testingComponent) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

func componentsEqual(one *Component, two *Component) bool {
	if one == nil && two == nil {
		return true
	}
	if one == nil || two == nil {
		return false
	}
	if one.Deck.Name() != two.Deck.Name() {
		return false
	}
	if one.DeckIndex != two.DeckIndex {
		return false
	}
	return true
}

type testState struct {
	Game    *testGameState
	Players []*testPlayerState
}

func (t *testState) GameState() GameState {
	return t.Game
}

func (t *testState) PlayerStates() []PlayerState {
	result := make([]PlayerState, len(t.Players))
	for i, player := range t.Players {
		result[i] = player
	}
	return result
}

func (t *testState) Diagram() string {
	return "IMPLEMENT ME"
}

func (t *testState) Copy() State {

	result := &testState{}

	if t.Game != nil {
		result.Game = t.Game.Copy().(*testGameState)
	}

	if t.Players == nil {
		return result
	}

	array := make([]*testPlayerState, len(t.Players))
	for i, player := range t.Players {
		array[i] = player.Copy().(*testPlayerState)
	}
	result.Players = array

	return result
}

type testGameState struct {
	CurrentPlayer int
	DrawDeck      *GrowableStack
	//TODO: have a Stack here.
}

func (t *testGameState) Copy() GameState {
	var result testGameState
	result = *t
	result.DrawDeck = t.DrawDeck.Copy()
	return &result
}

func (t *testGameState) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testGameState) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

type testPlayerState struct {
	//Note: PlayerIndex is stored ehre, but not a normal property or
	//serialized, because it's really just a convenience method because it's
	//implied by its position in the State.Users array.
	playerIndex       int
	Score             int
	MovesLeftThisTurn int
	IsFoo             bool
}

func (t *testPlayerState) PlayerIndex() int {
	return t.playerIndex
}

func (t *testPlayerState) Copy() PlayerState {
	var result testPlayerState
	result = *t
	return &result
}

func (t *testPlayerState) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testPlayerState) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

type testMoveAdvanceCurentPlayer struct{}

func (t *testMoveAdvanceCurentPlayer) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testMoveAdvanceCurentPlayer) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

func (t *testMoveAdvanceCurentPlayer) SetProp(name string, val interface{}) error {
	return PropertySetImpl(t, name, val)
}

func (t *testMoveAdvanceCurentPlayer) Copy() Move {
	var result testMoveAdvanceCurentPlayer
	result = *t
	return &result
}

func (t *testMoveAdvanceCurentPlayer) DefaultsForState(state State) {
	//No defaults to set
}

func (t *testMoveAdvanceCurentPlayer) Name() string {
	return "Advance Current Player"
}

func (t *testMoveAdvanceCurentPlayer) Description() string {
	return "Advances to the next player when the current player has no more legal moves they can make this turn."
}

func (t *testMoveAdvanceCurentPlayer) Legal(state State) error {
	payload := state.(*testState)

	player := payload.Players[payload.Game.CurrentPlayer]

	if player.MovesLeftThisTurn > 0 {
		return errors.New("The current player still has moves left this turn.")
	}

	return nil
}

func (t *testMoveAdvanceCurentPlayer) Apply(state State) error {

	payload := state.(*testState)

	//Make sure we're leaving it at 0
	payload.Players[payload.Game.CurrentPlayer].MovesLeftThisTurn = 0

	payload.Game.CurrentPlayer++

	if payload.Game.CurrentPlayer >= len(payload.Players) {
		payload.Game.CurrentPlayer = 0
	}

	payload.Players[payload.Game.CurrentPlayer].MovesLeftThisTurn = 1

	return nil
}

type testMove struct {
	AString           string
	ScoreIncrement    int
	TargetPlayerIndex int
	ABool             bool
}

func (t *testMove) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testMove) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

func (t *testMove) SetProp(name string, val interface{}) error {
	return PropertySetImpl(t, name, val)
}

func (t *testMove) Copy() Move {
	var result testMove
	result = *t
	return &result
}

func (t *testMove) Name() string {
	return "Test"
}

func (t *testMove) Description() string {
	return "Advances the score of the current player by the specified amount."
}

func (t *testMove) DefaultsForState(state State) {
	s := state.(*testState)

	t.TargetPlayerIndex = s.Game.CurrentPlayer
	t.ScoreIncrement = 3
}

func (t *testMove) Legal(state State) error {

	payload := state.(*testState)

	if payload.Game.CurrentPlayer != t.TargetPlayerIndex {
		return errors.New("The current player is not the same as the target player")
	}

	return nil

}

func (t *testMove) Apply(state State) error {

	payload := state.(*testState)

	payload.Players[payload.Game.CurrentPlayer].Score += t.ScoreIncrement

	payload.Players[payload.Game.CurrentPlayer].MovesLeftThisTurn -= 1

	return nil
}

type testAlwaysLegalMove struct{}

func (t *testAlwaysLegalMove) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testAlwaysLegalMove) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

func (t *testAlwaysLegalMove) SetProp(name string, val interface{}) error {
	return PropertySetImpl(t, name, val)
}

func (t *testAlwaysLegalMove) Copy() Move {
	var result testAlwaysLegalMove
	result = *t
	return &result
}

func (t *testAlwaysLegalMove) Name() string {
	return "Test Always Legal Move"
}

func (t *testAlwaysLegalMove) Description() string {
	return "A move that is always legal"
}

func (t *testAlwaysLegalMove) DefaultsForState(state State) {
	//Pass
}

func (t *testAlwaysLegalMove) Legal(state State) error {

	//This move is always legal

	return nil

}

func (t *testAlwaysLegalMove) Apply(state State) error {

	//This move doesn't do anything

	return nil
}

//testingComponentValues is designed to be run on a stack.ComponentValues() of
//a stack of testingComponents, in order to convert them all to the specified
//underlying struct.
func testingComponentValues(in []PropertyReader) []*testingComponent {
	result := make([]*testingComponent, len(in))
	for i := 0; i < len(in); i++ {
		if in[i] == nil {
			result[i] = nil
			continue
		}
		result[i] = in[i].(*testingComponent)
	}
	return result
}

//testGame returns a Game that has not yet had SetUp() called.
func testGame() *Game {

	manager := newTestGameManger()

	manager.SetUp()

	game := NewGame(manager)

	return game
}
