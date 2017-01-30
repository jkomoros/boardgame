package boardgame

//Place to define testing structs and helpers that are useful throughout

//testingComponent is a very basic thing that fufills the Component interface.
type testingComponent struct {
	String  string
	Integer int
}

const testGameName = "testgame"

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

type testStatePayload struct {
	game  *testGameState
	users []*testUserState
}

func (t *testStatePayload) Game() GameState {
	return t.game
}

func (t *testStatePayload) Users() []UserState {
	result := make([]UserState, len(t.users))
	for i, user := range t.users {
		result[i] = user
	}
	return result
}

func (t *testStatePayload) Copy() StatePayload {

	result := &testStatePayload{}

	if t.game != nil {
		result.game = t.game.Copy().(*testGameState)
	}

	if t.users == nil {
		return result
	}

	array := make([]*testUserState, len(t.users))
	for i, user := range t.users {
		array[i] = user.Copy().(*testUserState)
	}
	result.users = array

	return result
}

func (t *testStatePayload) JSON() JSONObject {

	usersArray := make([]JSONObject, len(t.users))

	for i, user := range t.users {
		usersArray[i] = user.JSON()
	}

	return JSONMap{
		"Game":  t.game.JSON(),
		"Users": usersArray,
	}
}

type testGameState struct {
	CurrentPlayer int
	//TODO: have a Stack here.
}

func (t *testGameState) Copy() GameState {
	return &testGameState{
		CurrentPlayer: t.CurrentPlayer,
	}
}

func (t *testGameState) JSON() JSONObject {

	return t
}

func (t *testGameState) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testGameState) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

type testUserState struct {
	//Note: PlayerIndex is stored ehre, but not a normal property or
	//serialized, because it's really just a convenience method because it's
	//implied by its position in the State.Users array.
	playerIndex       int
	Score             int
	MovesLeftThisTurn int
	IsFoo             bool
}

func (t *testUserState) PlayerIndex() int {
	return t.playerIndex
}

func (t *testUserState) Copy() UserState {
	return &testUserState{
		playerIndex:       t.playerIndex,
		Score:             t.Score,
		IsFoo:             t.IsFoo,
		MovesLeftThisTurn: t.MovesLeftThisTurn,
	}
}

func (t *testUserState) JSON() JSONObject {
	return t
}

func (t *testUserState) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testUserState) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

type testMoveAdvanceCurentPlayer struct{}

func (t *testMoveAdvanceCurentPlayer) GameName() string {
	return testGameName
}

func (t *testMoveAdvanceCurentPlayer) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testMoveAdvanceCurentPlayer) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

func (t *testMoveAdvanceCurentPlayer) JSON() JSONObject {
	return t
}

func (t *testMoveAdvanceCurentPlayer) Legal(state StatePayload) bool {
	payload := state.(*testStatePayload)

	user := payload.users[payload.game.CurrentPlayer]

	if user.MovesLeftThisTurn > 0 {
		return false
	}

	return true
}

func (t *testMoveAdvanceCurentPlayer) Apply(state StatePayload) StatePayload {
	result := state.Copy()

	payload := result.(*testStatePayload)

	//Make sure we're leaving it at 0
	payload.users[payload.game.CurrentPlayer].MovesLeftThisTurn = 0

	payload.game.CurrentPlayer++

	if payload.game.CurrentPlayer >= len(payload.users) {
		payload.game.CurrentPlayer = 0
	}

	payload.users[payload.game.CurrentPlayer].MovesLeftThisTurn = 1

	return result
}

type testMove struct {
	AString           string
	ScoreIncrement    int
	TargetPlayerIndex int
	ABool             bool
}

func (t *testMove) GameName() string {
	return testGameName
}

func (t *testMove) Props() []string {
	return PropertyReaderPropsImpl(t)
}

func (t *testMove) Prop(name string) interface{} {
	return PropertyReaderPropImpl(t, name)
}

func (t *testMove) JSON() JSONObject {
	return t
}

func (t *testMove) Legal(state StatePayload) bool {

	payload := state.(*testStatePayload)

	if payload.game.CurrentPlayer != t.TargetPlayerIndex {
		return false
	}

	return true

}

func (t *testMove) Apply(state StatePayload) StatePayload {
	result := state.Copy()

	payload := result.(*testStatePayload)

	payload.users[payload.game.CurrentPlayer].Score += t.ScoreIncrement

	payload.users[payload.game.CurrentPlayer].MovesLeftThisTurn -= 1

	return result
}

func testGame() *Game {

	chest := NewComponentChest(testGameName)

	deck := &Deck{}

	deck.AddComponent(&Component{
		Values: &testingComponent{
			"foo",
			1,
		},
	})

	deck.AddComponent(&Component{
		Values: &testingComponent{
			"bar",
			2,
		},
	})

	deck.AddComponent(&Component{
		Values: &testingComponent{
			"baz",
			5,
		},
	})

	deck.AddComponent(&Component{
		Values: &testingComponent{
			"slam",
			10,
		},
	})

	chest.AddDeck("test", deck)

	chest.Finish()

	game := &Game{
		Name:     testGameName,
		Delegate: &testGameDelegate{},
		State: &State{
			Version: 0,
			Schema:  0,
			Payload: &testStatePayload{
				game: &testGameState{
					CurrentPlayer: 0,
				},
				users: []*testUserState{
					&testUserState{
						playerIndex:       0,
						Score:             0,
						MovesLeftThisTurn: 1,
						IsFoo:             false,
					},
					&testUserState{
						playerIndex:       1,
						Score:             0,
						MovesLeftThisTurn: 0,
						IsFoo:             false,
					},
					&testUserState{
						playerIndex:       2,
						Score:             0,
						MovesLeftThisTurn: 0,
						IsFoo:             true,
					},
				},
			},
		},
	}

	game.SetChest(chest)

	return game
}
