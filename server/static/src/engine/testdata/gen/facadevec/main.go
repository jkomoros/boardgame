// Command facadevec emits the known answers for the Go-shaped facade
// (engine/go/): ../../facade.vectors.json, from the Go standard library and
// the framework's enum, graph and behavior packages, and
// ../../facade-stacks.vectors.json, from scripted sequences of component
// and stack calls on real vecgame states, and ../../facade-moves.vectors.json,
// an enum-typed move field of real Go checkers (moves.go). Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/facadevec
package main

import (
	"encoding/json"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func write(path string, v interface{}) {
	blob, err := json.Marshal(v)
	check(err)
	check(os.WriteFile(path, append(blob, '\n'), 0644))
}

func main() {
	write("server/static/src/engine/testdata/facade.vectors.json", langVecs())
	write("server/static/src/engine/testdata/facade-stacks.vectors.json", stackVecs())
	write("server/static/src/engine/testdata/facade-moves.vectors.json", moveFieldVecs())
}
