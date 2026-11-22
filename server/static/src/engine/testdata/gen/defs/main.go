// Command defs emits ../../defs/<game>.def.json: the engine GameDefinition
// for the vector game and for the example games whose golden histories the
// engine replays, each derived from the game's live Go manager by
// gamedef.DefinitionFor, without the fields gamedef adds
// (Definition.WithoutAdditions). Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/defs
//
// boardgame-util/lib/gamedef's tests check that the files are current.
package main

import (
	"os"

	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	files, err := vecgame.DefFiles()
	check(err)
	for _, file := range files {
		check(os.WriteFile("server/static/src/engine/testdata/defs/"+file.Name+".def.json", file.Blob, 0644))
	}
}
