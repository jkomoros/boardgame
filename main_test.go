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
		nil,
	}

	return game
}
