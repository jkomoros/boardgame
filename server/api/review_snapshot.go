package api

import (
	"fmt"
	"github.com/jkomoros/boardgame"
)

// ViewerMoveLegality is the default-bound legality shown by the ordinary move
// tray. It omits free-form errors and private move names unavailable to a viewer.
type ViewerMoveLegality struct {
	LegalForPlayer bool `json:"legalForPlayer"`
	LegalForAnyone bool `json:"legalForAnyone"`
}

// MoveLegalityForViewer reuses /info's visibility and legality rules for local
// review tooling. It does not claim legality for arbitrary non-default inputs.
func MoveLegalityForViewer(game *boardgame.Game, viewer boardgame.PlayerIndex) (map[string]ViewerMoveLegality, error) {
	if game == nil || !game.AtProposalFrontier() {
		return nil, fmt.Errorf("move review requires a settled game")
	}
	if viewer < boardgame.ObserverPlayerIndex || int(viewer) >= game.NumPlayers() {
		return nil, fmt.Errorf("move review requires observer or a configured player")
	}
	server := &Server{}
	forms := server.generateFormsWithLegality(game, game.CurrentState(), viewer)
	result := make(map[string]ViewerMoveLegality, len(forms))
	for _, form := range forms {
		result[form.Name] = ViewerMoveLegality{LegalForPlayer: form.LegalForPlayer, LegalForAnyone: form.LegalForAnyone}
	}
	return result, nil
}
