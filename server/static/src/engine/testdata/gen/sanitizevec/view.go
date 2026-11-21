package main

import (
	"encoding/json"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
)

type enumKeyT = enum.EnumKey

func enumKey(i int) enum.EnumKey { return enum.EnumKey(i) }

// viewOf returns what viewer receives as CurrentState: the object
// Game.JSONForPlayer builds, with Computed removed (computed properties are
// game rules, outside the engine core).
func viewOf(h *vecgame.Harness, vs vecState, st boardgame.State, viewer int) ([]byte, string, bool) {
	game := st.Game()
	obj, err := game.JSONForPlayer(boardgame.PlayerIndex(viewer), st)
	if err != nil {
		return nil, err.Error(), false
	}
	blob, err := json.Marshal(obj)
	check(err)
	var parsed struct{ CurrentState map[string]json.RawMessage }
	check(json.Unmarshal(blob, &parsed))
	delete(parsed.CurrentState, "Computed")
	out, err := json.Marshal(parsed.CurrentState)
	check(err)
	return out, "", false
}
