package vecgame

import (
	"encoding/json"
	"io"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

// Definition is the engine's GameDefinition (see ../../../definition.ts),
// derived from a live GameManager rather than written by hand: property
// types, stack shapes and struct-tag sanitization policies are read back from
// the manager the Go core actually built. The derivation lives in
// boardgame-util/lib/gamedef; these are its types.
type Definition = gamedef.Definition

type DeckDef = gamedef.DeckDef

type EnumDef = gamedef.EnumDef

type SubStateDef = gamedef.SubStateDef

type PropDef = gamedef.PropDef

// PolicyName is the lowercase name the engine uses for a policy.
func PolicyName(p boardgame.Policy) string { return gamedef.PolicyName(p) }

// DefinitionFor derives the engine definition for any game's manager, in the
// shape the checked-in ../../defs/*.def.json have: gamedef.DefinitionFor
// without the fields it adds (Definition.WithoutAdditions).
func DefinitionFor(manager *boardgame.GameManager) (*Definition, error) {
	def, err := gamedef.DefinitionFor(manager)
	if err != nil {
		return nil, err
	}
	return def.WithoutAdditions(), nil
}

// DefFile is one ../../defs/<Name>.def.json and its bytes.
type DefFile struct {
	Name string
	Blob []byte
}

// DefFiles derives every checked-in ../../defs/*.def.json: the vector game
// and the example games whose golden histories the engine replays, each from
// a fresh manager. gen/defs writes them; boardgame-util/lib/gamedef's tests
// compare them with the checked-in files.
func DefFiles() ([]DefFile, error) {
	delegates := []boardgame.GameDelegate{&Delegate{}, memory.NewDelegate(), blackjack.NewDelegate(), pig.NewDelegate(), tictactoe.NewDelegate()}
	var result []DefFile
	for _, delegate := range delegates {
		manager, err := boardgame.NewGameManager(delegate, memstore.NewStorageManager())
		if err != nil {
			return nil, err
		}
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		manager.SetLogger(logger)
		def, err := DefinitionFor(manager)
		if err != nil {
			return nil, err
		}
		blob, err := json.MarshalIndent(def, "", "\t")
		if err != nil {
			return nil, err
		}
		result = append(result, DefFile{Name: def.Name, Blob: append(blob, '\n')})
	}
	return result, nil
}
