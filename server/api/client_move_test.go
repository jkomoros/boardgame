package api

import (
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/tictactoe"
)

func TestMoveJSONForPlayerSanitizesNameAndProperties(t *testing.T) {
	storage := newLegalLedgerStorage()
	manager, err := boardgame.NewGameManager(tictactoe.NewDelegate(), storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	server := &Server{}
	hasPlaceToken := func(forms []*moveForm) bool {
		for _, form := range forms {
			if form.Name == "Place Token" {
				return true
			}
		}
		return false
	}
	if !hasPlaceToken(server.generateFormsWithLegality(game, game.CurrentState(), 0)) {
		t.Fatal("current actor did not receive private Place Token form")
	}
	if hasPlaceToken(server.generateFormsWithLegality(game, game.CurrentState(), 1)) {
		t.Fatal("other player received private Place Token form")
	}
	if hasPlaceToken(server.generateFormsWithLegality(game, game.CurrentState(), boardgame.ObserverPlayerIndex)) {
		t.Fatal("observer received private Place Token form")
	}

	move := game.MoveByName("Place Token")
	if err := move.ReadSetter().SetIntProp("Slot", 3); err != nil {
		t.Fatal(err)
	}
	version := move.Info().Version()
	if err := <-game.ProposeMove(move, 0); err != nil {
		t.Fatal(err)
	}
	record, err := storage.Move(game.ID(), version)
	if err != nil {
		t.Fatal(err)
	}

	assertWire := func(viewer boardgame.PlayerIndex, want string) {
		t.Helper()
		bundles, err := server.moveBundles(game, []*boardgame.MoveStorageRecord{record}, viewer, false, false)
		if err != nil {
			t.Fatal(err)
		}
		projected := bundles[0]["Move"]
		encoded, err := json.Marshal(projected)
		if err != nil {
			t.Fatal(err)
		}
		if string(encoded) != want {
			t.Fatalf("viewer %d move = %s, want %s", viewer, encoded, want)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"Blob", "Proposer", "Initiator", "Phase", "Timestamp"} {
			if _, exists := decoded[forbidden]; exists {
				t.Fatalf("viewer %d leaked %s: %s", viewer, forbidden, encoded)
			}
		}
	}

	assertWire(0, `{"AnimationKey":"Place Token","Properties":{"Slot":3},"Version":1}`)
	assertWire(1, `{"AnimationKey":"Hidden Action","Version":1}`)
	assertWire(boardgame.ObserverPlayerIndex, `{"AnimationKey":"Hidden Action","Version":1}`)
	assertWire(boardgame.AdminPlayerIndex, `{"AnimationKey":"Place Token","Properties":{"Slot":3,"TargetPlayerIndex":0},"Version":1}`)

	if got := string(record.Blob); got == "" || got == "null" {
		t.Fatal("projection destructively changed the stored move blob")
	}
}
