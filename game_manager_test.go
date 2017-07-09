package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"strings"
	"testing"
)

const (
	colorRed enum.Constant = iota
	colorBlue
	colorGreen
)

var testEnums = enum.NewSet()

var testColorEnum = testEnums.MustAdd("color", map[enum.Constant]string{
	colorRed:   "Red",
	colorBlue:  "Blue",
	colorGreen: "Green",
})

func newTestGameChest() *ComponentChest {

	chest := NewComponentChest(testEnums)

	deck := NewDeck()

	deck.AddComponent(&testingComponent{
		"foo",
		1,
	})

	deck.AddComponent(&testingComponent{
		"bar",
		2,
	})

	deck.AddComponent(&testingComponent{
		"baz",
		5,
	})

	deck.AddComponent(&testingComponent{
		"slam",
		10,
	})

	deck.SetShadowValues(&testShadowValues{
		Message: "Foo",
	})

	chest.AddDeck("test", deck)

	chest.Finish()

	return chest
}

func newTestGameManger() *GameManager {
	manager := NewGameManager(&testGameDelegate{}, newTestGameChest(), newTestStorageManager())

	manager.AddAgent(&testAgent{})

	manager.BulkAddMoveTypes([]*MoveTypeConfig{
		&testMoveConfig,
		&testMoveIncrementCardInHandConfig,
		&testMoveDrawCardConfig,
		&testMoveAdvanceCurrentPlayerConfig,
		&testMoveInvalidPlayerIndexConfig,
	})

	return manager
}

type nilStackGameDelegate struct {
	testGameDelegate
	//We wil always return a player state that has nil. But if this is false, we will also return one for game.
	nilForPlayer bool
}

func (n *nilStackGameDelegate) EmptyPlayerState(playe PlayerIndex) MutablePlayerState {
	return &testPlayerState{}
}

func (n *nilStackGameDelegate) EmptyGameState() MutableSubState {
	if n.nilForPlayer {
		//return a non-nil one.
		return n.testGameDelegate.EmptyGameState()
	}

	return &testGameState{}
}

func TestDefaultMove(t *testing.T) {
	//Tests that Moves based on DefaultMove copy correctly

	game := testGame()

	game.SetUp(0, nil)

	//FixUpMoveByName calls Copy under the covers.
	move := game.FixUpMoveByName("Advance Current Player")

	assert.For(t).ThatActual(move).IsNotNil()

	assert.For(t).ThatActualString(move.Type().Name()).Equals("Advance Current Player")

	convertedMove, ok := move.(*testMoveAdvanceCurentPlayer)

	assert.For(t).ThatActual(ok).Equals(true)
	assert.For(t).ThatActual(convertedMove).IsNotNil()
}

func TestNilStackErrors(t *testing.T) {
	manager := NewGameManager(&nilStackGameDelegate{}, newTestGameChest(), newTestStorageManager())

	if err := manager.SetUp(); err != nil {
		t.Fatal("Couldn't set up nilStackGameDelegate-based manager")
	}

	game := NewGame(manager)

	if game == nil {
		t.Error("No game provided from new game")
	}

	if err := game.SetUp(0, nil); err == nil {
		t.Error("Didn't get error when setting up with an empty game state with nil stacks")
	}

	//Switch so gameState is valid, but playerState is still not, so we can
	//make sure we do the same test for playerStates.
	manager.delegate.(*nilStackGameDelegate).nilForPlayer = true

	if err := game.SetUp(0, nil); err == nil {
		t.Error("Didn't get an error when given an empty player state with nil stacks")
	}

}

func TestMisshappenComputedProperties(t *testing.T) {
	delegate := &stateComputeDelegate{
		config: &ComputedPropertiesConfig{
			Global: map[string]ComputedGlobalPropertyDefinition{
				"ThisPropertyIsNotSupported": {
					Dependencies: []StatePropertyRef{},
					PropType:     TypeGrowableStack,
					Compute: func(state State) (interface{}, error) {
						return nil, nil
					},
				},
			},
		},
		returnDefaultCollection: true,
	}

	manager := NewGameManager(delegate, newTestGameChest(), newTestStorageManager())

	if manager.SetUp() == nil {
		t.Error("We had a misshapen config object but didn't get an error at setup")
	}
}

func TestGameManagerModifiableGame(t *testing.T) {
	game := testGame()

	game.SetUp(0, nil)

	manager := game.Manager()

	//use ToLower to verify that ID comparison is not case sensitive.
	otherGame := manager.ModifiableGame(strings.ToLower(game.Id()))

	if game != otherGame {
		t.Error("ModifiableGame didn't give back the same game that already existed")
	}

	//OK, forget about the real game to test us making a new one.
	manager.modifiableGames = make(map[string]*Game)

	otherGame = manager.ModifiableGame(game.Id())

	if otherGame == nil {
		t.Error("Other game didn't return anything even though it was in storage!")
	}

	if game == otherGame {
		t.Error("ModifiableGame didn't grab a game from storage fresh")
	}

	otherGame = manager.ModifiableGame("NOGAMEATTHISID")

	if otherGame != nil {
		t.Error("ModifiableGame returned a game even for an invalid ID")
	}

}

func TestGameManagerSetUp(t *testing.T) {

	manager := newTestGameManger()

	if manager.PlayerMoveTypes() != nil {
		t.Error("Got moves back before SetUp was called")
	}

	if manager.PlayerMoveTypeByName("Test") != nil {
		t.Error("Move by name returned a move before SetUp was called")
	}

	if manager.Agents() != nil {
		t.Error("Agent before setup was not nil")
	}

	if manager.AgentByName("Test") != nil {
		t.Error("Agent test before setup was not nil")
	}

	manager.SetUp()

	moves := manager.PlayerMoveTypes()

	if moves == nil {
		t.Error("Got nil player moves even after setting up")
	}

	if manager.Agents() == nil {
		t.Error("Agents after setup was nil")
	}

	if manager.AgentByName("test") == nil {
		t.Error("Agent test after setup was nil")
	}

	if manager.PlayerMoveTypeByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if manager.PlayerMoveTypeByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

}
