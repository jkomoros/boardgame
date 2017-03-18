package boardgame

import (
	"errors"
	"log"
)

//Place to define testing structs and helpers that are useful throughout

//testingComponent is a very basic thing that fufills the Component interface.
type testingComponent struct {
	String  string
	Integer int
}

const testGameName = "Test Game"

func (t *testingComponent) Reader() PropertyReader {
	return DefaultReader(t)
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

//Every game should do such a convenience method. state might be nil.
func concreteStates(state State) (*testGameState, []*testPlayerState) {

	if state == nil {
		return nil, nil
	}

	players := make([]*testPlayerState, len(state.Players()))

	for i, player := range state.Players() {
		players[i] = player.(*testPlayerState)
	}

	game, ok := state.Game().(*testGameState)

	if !ok {
		return nil, nil
	}

	return game, players
}

type testGameState struct {
	CurrentPlayer int
	DrawDeck      *GrowableStack
	//TODO: have a Stack here.
}

func (t *testGameState) MutableCopy() MutableGameState {
	var result testGameState
	result = *t
	result.DrawDeck = t.DrawDeck.Copy()
	return &result
}

func (t *testGameState) Copy() GameState {
	return t.MutableCopy()
}

func (t *testGameState) Reader() PropertyReader {
	return DefaultReader(t)
}

func (t *testGameState) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

type testPlayerState struct {
	//Note: PlayerIndex is stored ehre, but not a normal property or
	//serialized, because it's really just a convenience method because it's
	//implied by its position in the State.Users array.
	playerIndex       int
	Score             int
	MovesLeftThisTurn int
	Hand              *SizedStack
	IsFoo             bool
}

func (t *testPlayerState) PlayerIndex() int {
	return t.playerIndex
}

func (t *testPlayerState) MutableCopy() MutablePlayerState {
	var result testPlayerState
	result = *t

	result.Hand = t.Hand.Copy()

	return &result
}

func (t *testPlayerState) Copy() PlayerState {
	return t.MutableCopy()
}

func (t *testPlayerState) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testPlayerState) Reader() PropertyReader {
	return DefaultReader(t)
}

type testMoveAdvanceCurentPlayer struct{}

func (t *testMoveAdvanceCurentPlayer) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
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

	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	if player.MovesLeftThisTurn > 0 {
		return errors.New("The current player still has moves left this turn.")
	}

	return nil
}

func (t *testMoveAdvanceCurentPlayer) Apply(state MutableState) error {

	game, players := concreteStates(state)

	//Make sure we're leaving it at 0
	players[game.CurrentPlayer].MovesLeftThisTurn = 0

	game.CurrentPlayer++

	if game.CurrentPlayer >= len(players) {
		game.CurrentPlayer = 0
	}

	players[game.CurrentPlayer].MovesLeftThisTurn = 1

	return nil
}

type testMove struct {
	AString           string
	ScoreIncrement    int
	TargetPlayerIndex int
	ABool             bool
}

func (t *testMove) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
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
	game, _ := concreteStates(state)

	if game == nil {
		blob, _ := DefaultMarshalJSON(state)
		log.Println(string(blob))
		return
	}

	t.TargetPlayerIndex = game.CurrentPlayer
	t.ScoreIncrement = 3
}

func (t *testMove) Legal(state State) error {

	game, _ := concreteStates(state)

	if game.CurrentPlayer != t.TargetPlayerIndex {
		return errors.New("The current player is not the same as the target player")
	}

	return nil

}

func (t *testMove) Apply(state MutableState) error {

	game, players := concreteStates(state)

	players[game.CurrentPlayer].Score += t.ScoreIncrement

	players[game.CurrentPlayer].MovesLeftThisTurn -= 1

	return nil
}

type testAlwaysLegalMove struct{}

func (t *testAlwaysLegalMove) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
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

func (t *testAlwaysLegalMove) Apply(state MutableState) error {

	//This move doesn't do anything

	return nil
}

//testingComponentValues is designed to be run on a stack.ComponentValues() of
//a stack of testingComponents, in order to convert them all to the specified
//underlying struct.
func testingComponentValues(in []ComponentValues) []*testingComponent {
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
