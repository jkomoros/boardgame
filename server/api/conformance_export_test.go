package api

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/jkomoros/boardgame"
	memorygame "github.com/jkomoros/boardgame/examples/memory"
	werewolfgame "github.com/jkomoros/boardgame/examples/werewolf"
)

// formsSortedByField decodes forms JSON and sorts each form's Fields by name,
// the one part of the wire whose order Go map iteration leaves unstable.
func formsSortedByField(t *testing.T, blob []byte) []map[string]interface{} {
	t.Helper()
	var forms []map[string]interface{}
	if err := json.Unmarshal(blob, &forms); err != nil {
		t.Fatalf("decode forms: %v", err)
	}
	for _, form := range forms {
		fields, _ := form["Fields"].([]interface{})
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].(map[string]interface{})["Name"].(string) < fields[j].(map[string]interface{})["Name"].(string)
		})
	}
	return forms
}

// TestMoveFormsForViewerAtStateMatchesServerAtHead pins the conformance
// export to the server: for the head state it must report exactly what
// generateFormsWithLegality (the /info and bundle path) reports.
func TestMoveFormsForViewerAtStateMatchesServerAtHead(t *testing.T) {
	delegates := map[string]boardgame.GameDelegate{
		"memory":   memorygame.NewDelegate(),   // declarative ledger moves
		"werewolf": werewolfgame.NewDelegate(), // opaque moves
	}
	for name, delegate := range delegates {
		manager, err := boardgame.NewGameManager(delegate, newLegalLedgerStorage())
		if err != nil {
			t.Fatalf("%s manager: %v", name, err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("%s game: %v", name, err)
		}
		if name == "memory" {
			// Move off the setup state so defaults and verdicts are not the
			// trivial initial ones.
			current := manager.Delegate().CurrentPlayerIndex(game.CurrentState())
			if err := <-game.ProposeMove(game.MoveByName("Reveal Card"), current); err != nil {
				t.Fatalf("reveal: %v", err)
			}
		}
		server := &Server{}
		state := game.CurrentState()
		for viewer := boardgame.AdminPlayerIndex; int(viewer) < game.NumPlayers(); viewer++ {
			want, err := json.Marshal(server.generateFormsWithLegality(game, state, viewer))
			if err != nil {
				t.Fatal(err)
			}
			got, err := MoveFormsForViewerAtState(game, state, viewer)
			if err != nil {
				t.Fatalf("%s viewer %d: %v", name, viewer, err)
			}
			if !reflect.DeepEqual(formsSortedByField(t, got), formsSortedByField(t, want)) {
				t.Errorf("%s viewer %d: export differs from the server\n got: %s\nwant: %s", name, viewer, got, want)
			}
		}
		if game.Version() > 0 {
			// An older state must be accepted and bound to that state, not
			// to the head.
			if _, err := MoveFormsForViewerAtState(game, game.State(0), 0); err != nil {
				t.Errorf("%s: historical state: %v", name, err)
			}
		}
		if sanitized, err := state.SanitizedForPlayer(0); err != nil {
			t.Fatal(err)
		} else if _, err := MoveFormsForViewerAtState(game, sanitized, 0); err == nil {
			t.Errorf("%s: a sanitized state was accepted", name)
		}
	}
}
