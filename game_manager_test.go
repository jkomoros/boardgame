package boardgame

import (
	"strings"
	"testing"
)

func newTestGameChest() *ComponentChest {
	chest := NewComponentChest()

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

	manager.AddPlayerMove(&testMove{})
	manager.AddFixUpMove(&testMoveAdvanceCurentPlayer{})
	manager.AddPlayerMove(&testMoveIncrementCardInHand{})
	manager.AddPlayerMove(&testMoveDrawCard{})

	return manager
}

type nilStackGameDelegate struct {
	testGameDelegate
	//We wil always return a player state that has nil. But if this is false, we will also return one for game.
	nilForPlayer bool
}

func (n *nilStackGameDelegate) EmptyPlayerState(playerIndex int) MutablePlayerState {
	return &testPlayerState{}
}

func (n *nilStackGameDelegate) EmptyGameState() MutableGameState {
	if n.nilForPlayer {
		//return a non-nil one.
		return n.testGameDelegate.EmptyGameState()
	}

	return &testGameState{}
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

	if err := game.SetUp(0); err == nil {
		t.Error("Didn't get error when setting up with an empty game state with nil stacks")
	}

	//Switch so gameState is valid, but playerState is still not, so we can
	//make sure we do the same test for playerStates.
	manager.delegate.(*nilStackGameDelegate).nilForPlayer = true

	if err := game.SetUp(0); err == nil {
		t.Error("Didn't get an error when given an empty player state with nil stacks")
	}

}

func TestMisshappenComputedProperties(t *testing.T) {
	delegate := &stateComputeDelegate{
		config: &ComputedPropertiesConfig{
			Global: map[string]ComputedGlobalPropertyDefinition{
				"ThisPropertyIsNotSupported": ComputedGlobalPropertyDefinition{
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

	game.SetUp(0)

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

	if manager.PlayerMoves() != nil {
		t.Error("Got moves back before SetUp was called")
	}

	if manager.PlayerMoveByName("Test") != nil {
		t.Error("Move by name returned a move before SetUp was called")
	}

	manager.SetUp()

	moves := manager.PlayerMoves()

	if moves == nil {
		t.Error("Got nil player moves even after setting up")
	}

	for i := 0; i < len(moves); i++ {
		if moves[i] == manager.playerMoves[i] {
			t.Error("PlayerMoves didn't return a copy; got same item at", i)
		}
	}

	if manager.PlayerMoveByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if manager.PlayerMoveByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

}
