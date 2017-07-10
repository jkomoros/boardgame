package boardgame

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
	"log"
)

//Place to define testing structs and helpers that are useful throughout

type testAgent struct{}

func (t *testAgent) Name() string {
	return "Test"
}

func (t *testAgent) DisplayName() string {
	return "Robby the Robot"
}

func (t *testAgent) SetUpForGame(game *Game, player PlayerIndex) (state []byte) {
	return nil
}

func (t *testAgent) ProposeMove(game *Game, player PlayerIndex, agentState []byte) (move Move, newAgentState []byte) {

	state := game.CurrentState()

	gameState, _ := concreteStates(state)

	if gameState.CurrentPlayer != player {
		return nil, nil
	}

	move = game.PlayerMoveByName("Test")

	if move == nil {
		log.Println("Couldn't find move Test")
		return nil, nil
	}

	move.(*testMove).TargetPlayerIndex = player

	return move, nil
}

//testingComponent is a very basic thing that fufills the Component interface.
type testingComponent struct {
	String  string
	Integer int
}

type testingComponentDynamic struct {
	IntVar int
	Stack  *SizedStack
	Enum   enum.Var
}

const testGameName = "Test Game"

func (t *testingComponent) Reader() PropertyReader {
	return getDefaultReader(t)
}

func (t *testingComponentDynamic) Reader() PropertyReader {
	return getDefaultReader(t)
}

func (t *testingComponentDynamic) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testingComponentDynamic) Copy() MutableSubState {
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

	players := make([]*testPlayerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
		players[i] = player.(*testPlayerState)
	}

	game, ok := state.GameState().(*testGameState)

	if !ok {
		return nil, nil
	}

	return game, players
}

type testGameState struct {
	CurrentPlayer      PlayerIndex
	DrawDeck           *GrowableStack
	Timer              *Timer
	MyIntSlice         []int
	MyStringSlice      []string
	MyBoolSlice        []bool
	MyPlayerIndexSlice []PlayerIndex
	MyEnumValue        enum.Var
	//TODO: have a Stack here.
}

func (t *testGameState) Reader() PropertyReader {
	return getDefaultReader(t)
}

func (t *testGameState) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
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
	EnumVal           enum.Var
}

func (t *testPlayerState) PlayerIndex() PlayerIndex {
	return t.playerIndex
}

func (t *testPlayerState) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testPlayerState) Reader() PropertyReader {
	return getDefaultReader(t)
}

type testMoveInvalidPlayerIndex struct {
	baseMove
	//This move is a dangerous one and also a fix-up. So make it so by default
	//it doesn't apply.
	CurrentlyLegal bool
}

var testMoveInvalidPlayerIndexConfig = MoveTypeConfig{
	Name:     "Invalid PlayerIndex",
	HelpText: "Set one of the PlayerIndex properties to an invalid number, so we can verify that ApplyMove catches it.",
	MoveConstructor: func() Move {
		return new(testMoveInvalidPlayerIndex)
	},
	IsFixUp: true,
}

func (t *testMoveInvalidPlayerIndex) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testMoveInvalidPlayerIndex) Legal(state State, propopser PlayerIndex) error {

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
	baseMove
	TargetPlayerIndex PlayerIndex
}

var testMoveIncrementCardInHandConfig = MoveTypeConfig{
	Name:     "Increment IntValue of Card in Hand",
	HelpText: "Increments the IntValue of the card in the hand",
	MoveConstructor: func() Move {
		return new(testMoveIncrementCardInHand)
	},
}

func (t *testMoveIncrementCardInHand) DefaultsForState(state State) {
	t.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
}

func (t *testMoveIncrementCardInHand) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testMoveIncrementCardInHand) Legal(state State, proposer PlayerIndex) error {
	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	if !t.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The proposer is not the target Player index")
	}

	if !t.TargetPlayerIndex.Equivalent(game.CurrentPlayer) {
		return errors.New("The target player index is not the current player")
	}

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
	baseMove
	TargetPlayerIndex PlayerIndex
}

var testMoveDrawCardConfig = MoveTypeConfig{
	Name:     "Draw Card",
	HelpText: "Draws one card from draw deck into player's hand",
	MoveConstructor: func() Move {
		return new(testMoveDrawCard)
	},
}

func (t *testMoveDrawCard) DefaultsForState(state State) {
	t.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
}

func (t *testMoveDrawCard) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testMoveDrawCard) Legal(state State, proposer PlayerIndex) error {
	game, players := concreteStates(state)

	player := players[game.CurrentPlayer]

	if !t.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The proposer is not equivalent to targetplayerindex")
	}

	if !t.TargetPlayerIndex.Equivalent(game.CurrentPlayer) {
		return errors.New("The target player is not the current player")
	}

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

type testMoveAdvanceCurentPlayer struct {
	baseMove
}

var testMoveAdvanceCurrentPlayerConfig = MoveTypeConfig{
	Name:     "Advance Current Player",
	HelpText: "Advances to the next player when the current player has no more legal moves they can make this turn.",
	MoveConstructor: func() Move {
		return new(testMoveAdvanceCurentPlayer)
	},
	IsFixUp: true,
}

func (t *testMoveAdvanceCurentPlayer) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testMoveAdvanceCurentPlayer) Legal(state State, proposer PlayerIndex) error {

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
	baseMove
	AString           string
	ScoreIncrement    int
	TargetPlayerIndex PlayerIndex
	ABool             bool
}

var testMoveConfig = MoveTypeConfig{
	Name:     "Test",
	HelpText: "Advances the score of the current player by the specified amount.",
	MoveConstructor: func() Move {
		return new(testMove)
	},
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

func (t *testMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testMove) Legal(state State, proposer PlayerIndex) error {

	game, _ := concreteStates(state)

	if !t.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The target player index is not equivalent to the proposer")
	}

	if !t.TargetPlayerIndex.Equivalent(game.CurrentPlayer) {
		return errors.New("The target player is not hte current player")
	}

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

type testAlwaysLegalMove struct {
	baseMove
}

var testAlwaysLegalMoveConfig = MoveTypeConfig{
	Name:     "Test Always Legal Move",
	HelpText: "A move that is always legal",
	MoveConstructor: func() Move {
		return new(testAlwaysLegalMove)
	},
	IsFixUp: true,
}

func (t *testAlwaysLegalMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}

func (t *testAlwaysLegalMove) Legal(state State, proposer PlayerIndex) error {

	return nil

}

func (t *testAlwaysLegalMove) Apply(state MutableState) error {

	//This move doesn't do anything

	return nil
}

type illegalMove struct {
	baseMove
	Enum enum.Var
}

func (i *illegalMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(i)
}

func (i *illegalMove) Legal(state State, proposer PlayerIndex) error {
	return nil
}

func (i *illegalMove) Apply(state MutableState) error {
	return nil
}

var testIllegalMoveConfig = MoveTypeConfig{
	Name:     "Illegal Move",
	HelpText: "Move that is illegal because it has an illegal property type on it",
	MoveConstructor: func() Move {
		return new(illegalMove)
	},
}

//testingComponentValues is designed to be run on a stack.ComponentValues() of
//a stack of testingComponents, in order to convert them all to the specified
//underlying struct.
func testingComponentValues(in []SubState) []*testingComponent {
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

func makeTestGameIdsStable(game *Game) {
	//having the same fixed salt helps make the test predictable regarding
	//component ids.
	game.secretSalt = "FAKESALTFORTESTING"
	game.id = "FAKEIDFORTESTING"
}

//testGame returns a Game that has not yet had SetUp() called.
func testGame() *Game {

	manager := newTestGameManger()

	manager.SetUp()

	game := NewGame(manager)

	return game
}
