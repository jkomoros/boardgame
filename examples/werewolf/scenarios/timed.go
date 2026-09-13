// Package scenarios contains executable Werewolf review scripts.
package scenarios

import (
	"encoding/json"
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/werewolf"
)

// TimedVote records a partial daytime vote and its durable deadline resolution.
func TimedVote() (scenario.Spec, error) {
	spec := scenario.Spec{Name: "Werewolf: a timed partial vote", Seed: "werewolf-timed-v1", Players: 4,
		Variant: boardgame.Variant{"timedvoting": "on"}, Viewers: []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, 0, 1, 2, 3},
	}
	for i := 0; i < spec.Players; i++ {
		player := boardgame.PlayerIndex(i)
		spec.Steps = append(spec.Steps, scenario.Step{Label: fmt.Sprintf("Player %d joins", i+1), Seat: &player})
	}
	probe, err := scenario.Run(werewolf.NewDelegate(), spec)
	if err != nil {
		return spec, err
	}
	target := -1
	for _, view := range probe.Replay.Frames[len(probe.Replay.Frames)-1].Viewers {
		if view.Viewer <= 0 {
			continue
		}
		var wire struct {
			CurrentState struct{ Players []struct{ Role string } }
		}
		if err := json.Unmarshal(view.Game, &wire); err != nil {
			return spec, err
		}
		if wire.CurrentState.Players[view.Viewer].Role == "Villager" {
			target = int(view.Viewer)
			break
		}
	}
	if target < 0 {
		return spec, fmt.Errorf("no villager target available")
	}
	spec.Steps = append(spec.Steps, []scenario.Step{
		{Label: "One player votes; others have not", Move: "Cast Vote", Player: 0, Input: map[string]interface{}{"VoteTarget": target}},
		{Label: "Deadline resolves the votes cast so far", Timer: true},
	}...)
	return spec, nil
}
