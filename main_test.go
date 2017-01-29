package boardgame

//Place to define testing structs and helpers that are useful throughout

//testingComponent is a very basic thing that fufills the Component interface.
type testingComponent struct {
	deckName  string
	deckIndex int
	String    string
	Integer   int
}

const testGameName = "testgame"

func (t *testingComponent) Props() []string {
	return []string{"String", "Integer"}
}

func (t *testingComponent) Prop(name string) interface{} {
	switch name {
	case "String":
		return t.String
	case "Integer":
		return t.Integer
	default:
		return nil
	}
}

func (t *testingComponent) Deck() string {
	return t.deckName
}

func (t *testingComponent) DeckIndex() int {
	return t.deckIndex
}

func (t *testingComponent) GameName() string {
	return testGameName
}

func componentsEqual(one Component, two Component) bool {
	if one == nil && two == nil {
		return true
	}
	if one == nil || two == nil {
		return false
	}
	if one.Deck() != two.Deck() {
		return false
	}
	if one.DeckIndex() != two.DeckIndex() {
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

	return JSONObject{
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
	//TODO: once JSONObject is more generic, we can just return ourselves

	return JSONObject{
		"CurrentPlayer": t.CurrentPlayer,
	}
}

func (t *testGameState) Props() []string {
	return []string{"CurrentPlayer"}
}

func (t *testGameState) Prop(name string) interface{} {
	switch name {
	case "CurrentPlayer":
		return t.CurrentPlayer
	default:
		return nil
	}
}

type testUserState struct {
	//Note: PlayerIndex is stored ehre, but not a normal property or
	//serialized, because it's really just a convenience method because it's
	//implied by its position in the State.Users array.
	playerIndex int
	Score       int
	IsFoo       bool
}

func (t *testUserState) PlayerIndex() int {
	return t.playerIndex
}

func (t *testUserState) Copy() UserState {
	return &testUserState{
		playerIndex: t.playerIndex,
		Score:       t.Score,
		IsFoo:       t.IsFoo,
	}
}

func (t *testUserState) JSON() JSONObject {
	return JSONObject{
		"Score": t.Score,
		"IsFoo": t.IsFoo,
	}
}

func (t *testUserState) Props() []string {
	return []string{"Score", "IsFoo"}
}

func (t *testUserState) Prop(name string) interface{} {
	switch name {
	case "Score":
		return t.Score
	case "IsFoo":
		return t.IsFoo
	default:
		return nil
	}
}

type testMove struct {
	AString        string
	ScoreIncrement int
	ABool          bool
}

func (t *testMove) GameName() string {
	return testGameName
}

func (t *testMove) Props() []string {
	return []string{"AString", "ScoreIncrement", "ABool"}
}

func (t *testMove) Prop(name string) interface{} {
	switch name {
	case "AString":
		return t.AString
	case "ScoreIncrement":
		return t.ScoreIncrement
	case "ABool":
		return t.ABool
	default:
		return nil
	}
}

func (t *testMove) JSON() JSONObject {
	return JSONObject{
		"AString":        t.AString,
		"ScoreIncrement": t.ScoreIncrement,
		"ABool":          t.ABool,
	}
}

func (t *testMove) Legal(state *State) bool {

	payload := state.Payload.(*testStatePayload)

	if payload.game.CurrentPlayer != 0 {
		return false
	}

	return true

}

func (t *testMove) Apply(state *State) *State {
	result := state.Copy()

	payload := result.Payload.(*testStatePayload)

	payload.game.CurrentPlayer++
	payload.users[1].Score += t.ScoreIncrement

	return result
}

func testGame() *Game {
	//TODO: some kind of way to set the deckName/Index automatically at insertion?
	chest := ComponentChest{
		"test": &Deck{
			Name: "test",
			Components: []Component{
				&testingComponent{
					"test",
					0,
					"foo",
					1,
				},
				&testingComponent{
					"test",
					1,
					"bar",
					2,
				},
			},
		},
	}

	game := &Game{
		testGameName,
		chest,
		&State{
			Version: 0,
			Schema:  0,
			Payload: &testStatePayload{
				game: &testGameState{
					CurrentPlayer: 0,
				},
				users: []*testUserState{
					&testUserState{
						playerIndex: 0,
						Score:       0,
						IsFoo:       false,
					},
					&testUserState{
						playerIndex: 1,
						Score:       0,
						IsFoo:       false,
					},
					&testUserState{
						playerIndex: 2,
						Score:       0,
						IsFoo:       true,
					},
				},
			},
		},
	}

	return game
}
