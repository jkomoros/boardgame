package gamedef

import (
	"errors"

	"github.com/jkomoros/boardgame"
)

// Export is the run-time half of the exporter: it reads the live manager
// (moves in registration order, the chest, the constructors, plans,
// progressions, variants) and joins static, the static half from Static,
// into one GameData.
//
// manager is a GameManager built over the game's delegate, as
// testdata/gen/defs and lib/conformance build one:
//
//	manager, err := boardgame.NewGameManager(memory.NewDelegate(), memstore.NewStorageManager())
//
// static must come from Static on the same game's package; Export fails
// when a configured move's Go type has no entry in static.Moves.
//
// Not implemented: workstream 1 owns it.
func Export(manager *boardgame.GameManager, static *StaticInfo) (*GameData, error) {
	return nil, errors.New("gamedef: export not implemented")
}
