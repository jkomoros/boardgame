package werewolf

import (
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame"
)

func TestViewerRoleKnowledgeAndComputedLabels(t *testing.T) {
	_, game := newSeatedGame(t, 4)
	for _, viewer := range []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, 0, boardgame.AdminPlayerIndex} {
		view, err := game.JSONForPlayer(viewer, nil)
		if err != nil {
			t.Fatal(err)
		}
		blob, err := json.Marshal(view)
		if err != nil {
			t.Fatal(err)
		}
		var wire struct {
			CurrentState struct {
				Visibility boardgame.StateVisibility
				Computed   struct{ Players []map[string]interface{} }
			}
		}
		if err := json.Unmarshal(blob, &wire); err != nil {
			t.Fatal(err)
		}
		for i, player := range wire.CurrentState.Computed.Players {
			wantKnown := viewer == boardgame.AdminPlayerIndex || int(viewer) == i
			known := false
			for _, facet := range wire.CurrentState.Visibility.Players[i]["Role"] {
				known = known || facet == "values"
			}
			_, hasLabel := player["RoleValue"]
			if known != wantKnown || hasLabel != wantKnown {
				t.Fatalf("viewer=%v owner=%d: known=%v computedLabel=%v want=%v", viewer, i, known, hasLabel, wantKnown)
			}
		}
	}
}
