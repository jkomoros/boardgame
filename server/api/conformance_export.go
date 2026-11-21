package api

import (
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

// MoveFormsForViewerAtState exports, for conformance tooling, the move forms
// and legality /info would report to viewer if state were the game's head. It
// is additive: no server path calls it.
//
// It differs from generateFormsWithLegality in exactly one way: each move is
// bound with Game.MoveByNameForState(name, state) instead of Game.Moves(),
// which binds defaults against CurrentState. For the head state the two are
// the same call, so the result is the server's own (pinned by
// conformance_export_test.go). For an older state this is the snapshot-pure
// reading the server would have produced when that state was the head.
//
// Fields come back in formFields' map-iteration order, as on the wire; that
// order is not stable, so callers that need determinism must sort them.
func MoveFormsForViewerAtState(game *boardgame.Game, state boardgame.ImmutableState, viewer boardgame.PlayerIndex) (json.RawMessage, error) {
	if game == nil || state == nil || state.Game() != game {
		return nil, fmt.Errorf("move forms require a game and one of its states")
	}
	if state.Sanitized() {
		return nil, fmt.Errorf("move forms require authoritative unsanitized state")
	}
	if viewer < boardgame.AdminPlayerIndex || int(viewer) >= game.NumPlayers() {
		return nil, fmt.Errorf("move forms require admin, observer or a configured player")
	}
	server := &Server{}
	forms := make([]*moveForm, 0)
	for _, example := range game.Moves() {
		move := game.MoveByNameForState(example.Info().Name(), state)
		if move == nil {
			return nil, fmt.Errorf("move %q is not configured", example.Info().Name())
		}
		if base.IsFixUp(move) {
			continue
		}
		if !move.Info().MoveNamePublic() && !server.privateMoveAvailable(game, state, move, viewer) {
			continue
		}
		forms = append(forms, server.legalMoveForm(game, state, move, viewer))
	}
	return json.Marshal(forms)
}
