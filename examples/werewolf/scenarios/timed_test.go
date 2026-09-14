package scenarios_test

import (
	"encoding/json"
	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/werewolf"
	"github.com/jkomoros/boardgame/examples/werewolf/scenarios"
	"path/filepath"
	"testing"
)

func TestTimedVoteReplay(t *testing.T) {
	spec, err := scenarios.TimedVote()
	if err != nil {
		t.Fatal(err)
	}
	result, err := scenario.Run(werewolf.NewDelegate(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Replay.Frames) != 7 {
		t.Fatal("missing recorded decision")
	}
	for _, frame := range result.Replay.Frames {
		for _, view := range frame.Viewers {
			var wire struct {
				CurrentState struct{ Players []struct{ Role string } }
			}
			if err := json.Unmarshal(view.Game, &wire); err != nil {
				t.Fatal(err)
			}
			if view.Viewer < 0 || wire.CurrentState.Players[view.Viewer].Role != "Werewolf" {
				if _, disclosed := view.MoveLegality["Cast Night Vote"]; disclosed {
					t.Fatal("private move metadata reached an ineligible viewer")
				}
			}
		}
	}
	path := filepath.Join(t.TempDir(), "vote.json")
	if err := result.History.Save(path, false); err != nil {
		t.Fatal(err)
	}
	if err := golden.Compare(werewolf.NewDelegate(), path, false); err != nil {
		t.Fatal(err)
	}
}
