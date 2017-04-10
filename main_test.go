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

type testingComponentDynamic struct {
	IntVar int
	Stack  *SizedStack
}

const testGameName = "Test Game"

func (t *testingComponent) Reader() PropertyReader {
	return DefaultReader(t)
}

func (t *testingComponentDynamic) Reader() PropertyReader {
	return DefaultReader(t)
}

func (t *testingComponentDynamic) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testingComponentDynamic) Copy() MutableDynamicComponentValues {
	var result testingComponentDynamic
	result = *t
	result.Stack = t.Stack.Copy()
	return &result
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
	CurrentPlayer PlayerIndex
	DrawDeck      *GrowableStack
	Timer         *Timer
	//TODO: have a Stack here.
}

func (t *testGameState) MutableCopy() MutableGameState {
	var result testGameState
	result = *t
	result.DrawDeck = t.DrawDeck.Copy()
	result.Timer = t.Timer.Copy()
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
	playerIndex       PlayerIndex
	Score             int
	MovesLeftThisTurn int
	Hand              *SizedStack
	IsFoo             bool
}

func (t *testPlayerState) PlayerIndex() PlayerIndex {
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

type testMoveInvalidPlayerIndex struct {
	//This move is a dangerous one and also a fix-up. So make it so by default
	//it doesn't apply.
	CurrentlyLegal bool
}

func (t *testMoveInvalidPlayerIndex) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testMoveInvalidPlayerIndex) Copy() Move {
	var result testMoveInvalidPlayerIndex
	result = *t
	return &result
}

func (t *testMoveInvalidPlayerIndex) DefaultsForState(state State) {
	return
}

func (t *testMoveInvalidPlayerIndex) Name() string {
	return "Invalid PlayerIndex"
}

func (t *testMoveInvalidPlayerIndex) Description() string {
	return "Set one of the PlayerIndex properties to an invalid number, so we can verify that ApplyMove catches it."
}

func (t *testMoveInvalidPlayerIndex) Legal(state State) error {

	if !t.CurrentlyLegal {
		return errors.New("Move not currently legal")
	}

	return nil
}

func (t *testMoveInvalidPlayerIndex) Apply(state MutableState) error {
	game, players := concreteStates(state)

	game.CurrentPlayer = PlayerIndex(len(players))

	return nil
}

type testMoveIncrementCardInHand struct {
	TargetPlayerIndex PlayerIndex
}

func (t *testMoveIncrementCardInHand) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testMoveIncrementCardInHand) Copy() Move {
	var result testMoveIncrementCardInHand
	result = *t
	return &result
}

func (t *testMoveIncrementCardInHand) DefaultsForState(state State) {

	game, _ := concreteStates(state)

	t.TargetPlayerIndex = game.CurrentPlayer
}

func (t *testMoveIncrementCardInHand) Name() string {
	return "Increment IntValue of Card in Hand"
}

func (t *testMoveIncrementCardInHand) Description() string {
	return "Increments the IntValue of the card in the hand"
}

func (t *testMoveIncrementCardInHand) Legal(state State) error {
	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	if player.Hand.NumComponents() == 0 {
		return errors.New("The current player does not have any components in their hand")
	}

	if game.DrawDeck.NumComponents() < 1 {
		return errors.New("There aren't enough cards left over in the draw deck")
	}

	return nil
}

func (t *testMoveIncrementCardInHand) Apply(state MutableState) error {
	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	for _, component := range player.Hand.Components() {
		if component == nil {
			continue
		}

		values := component.DynamicValues(state)

		if values == nil {
			return errors.New("DynamicValues was nil")
		}

		easyValues := values.(*testingComponentDynamic)

		easyValues.IntVar += 3
		game.DrawDeck.MoveComponent(LastComponentIndex, easyValues.Stack, FirstSlotIndex)

		return nil

	}

	return errors.New("Didn't find a component in hand")
}

type testMoveDrawCard struct {
	TargetPlayerIndex PlayerIndex
}

func (t *testMoveDrawCard) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testMoveDrawCard) Copy() Move {
	var result testMoveDrawCard
	result = *t
	return &result
}

func (t *testMoveDrawCard) DefaultsForState(state State) {

	game, _ := concreteStates(state)

	t.TargetPlayerIndex = game.CurrentPlayer
}

func (t *testMoveDrawCard) Name() string {
	return "Draw Card"
}

func (t *testMoveDrawCard) Description() string {
	return "Draws one card from draw deck into player's hand"
}

func (t *testMoveDrawCard) Legal(state State) error {
	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	if player.Hand.SlotsRemaining() == 0 {
		return errors.New("The current player does not have enough slots in their hand")
	}

	if game.DrawDeck.NumComponents() == 0 {
		return errors.New("there are no cards to draw")
	}

	return nil
}

func (t *testMoveDrawCard) Apply(state MutableState) error {
	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	if err := game.DrawDeck.MoveComponent(FirstComponentIndex, player.Hand, FirstSlotIndex); err != nil {
		return errors.New("couldn't move component from draw deck to hand: " + err.Error())
	}

	return nil
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

	if !game.CurrentPlayer.Valid(state) {
		//Must have looped back around to 0
		game.CurrentPlayer = 0
	}

	players[game.CurrentPlayer].MovesLeftThisTurn = 1

	return nil
}

type testMove struct {
	AString           string
	ScoreIncrement    int
	TargetPlayerIndex PlayerIndex
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
