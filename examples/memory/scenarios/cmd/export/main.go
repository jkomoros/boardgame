// Export the real Memory scenario for local renderer review and golden replay.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/memory/scenarios"
)

func main() {
	replayPath := flag.String("replay", "memory-replay.json", "viewer replay output")
	historyPath := flag.String("history", "", "optional authoritative golden output")
	flag.Parse()
	spec, err := scenarios.Mismatch()
	if err != nil {
		log.Fatal(err)
	}
	result, err := scenario.Run(memory.NewDelegate(), spec)
	if err != nil {
		log.Fatal(err)
	}
	blob, err := json.MarshalIndent(result.Replay, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*replayPath, append(blob, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
	if *historyPath != "" {
		if err := result.History.Save(*historyPath, false); err != nil {
			log.Fatal(err)
		}
	}
}
