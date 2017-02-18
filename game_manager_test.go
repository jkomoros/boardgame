package boardgame

import (
	"strings"
	"testing"
)

func newTestGameChest() *ComponentChest {
	chest := NewComponentChest()

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
