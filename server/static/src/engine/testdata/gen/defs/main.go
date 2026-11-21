// Command defs emits ../../defs/<game>.def.json: the engine GameDefinition
// for the vector game and for the example games whose golden histories the
// engine replays, each derived from the game's live Go manager. Run from the
// repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/defs
package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	delegates := []boardgame.GameDelegate{&vecgame.Delegate{}, memory.NewDelegate(), blackjack.NewDelegate(), pig.NewDelegate(), tictactoe.NewDelegate()}
	for _, delegate := range delegates {
		manager, err := boardgame.NewGameManager(delegate, memstore.NewStorageManager())
		check(err)
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		manager.SetLogger(logger)
		def, err := vecgame.DefinitionFor(manager)
		check(err)
		blob, err := json.MarshalIndent(def, "", "\t")
		check(err)
		check(os.WriteFile("server/static/src/engine/testdata/defs/"+def.Name+".def.json", append(blob, '\n'), 0644))
	}
}
