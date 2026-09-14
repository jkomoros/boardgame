package main

import (
	"encoding/json"
	"flag"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/werewolf"
	"github.com/jkomoros/boardgame/examples/werewolf/scenarios"
	"log"
	"os"
)

func main() {
	path := flag.String("replay", "examples/werewolf/client/scenario-replay.json", "viewer replay output")
	flag.Parse()
	spec, err := scenarios.TimedVote()
	if err != nil {
		log.Fatal(err)
	}
	result, err := scenario.Run(werewolf.NewDelegate(), spec)
	if err != nil {
		log.Fatal(err)
	}
	blob, err := json.MarshalIndent(result.Replay, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*path, append(blob, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
}
