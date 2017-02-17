package boardgame

import (
	"testing"
)

func newTestGameChest() *ComponentChest {
	chest := NewComponentChest(testGameName)

	deck := &Deck{}

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

	chest.AddDeck("test", deck)

	chest.Finish()

	return chest
}

func newTestGameManger() *GameManager {
	manager := NewGameManager(&testGameDelegate{}, newTestGameChest(), NewInMemoryStorageManager())

	manager.AddPlayerMove(&testMove{})
	manager.AddFixUpMove(&testMoveAdvanceCurentPlayer{})

	return manager
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
