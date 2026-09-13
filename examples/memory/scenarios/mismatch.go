// Package scenarios contains executable Memory examples for tests and review.
package scenarios

import (
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/memory"
)

// Mismatch reveals two different cards, then explicitly fires the hide timer.
// The setup probe chooses indices from an isolated authoritative record; those
// indices become ordinary explicit moves in the resulting reproducible script.
func Mismatch() (scenario.Spec, error) {
	spec := scenario.Spec{Name: "Memory: reveal, compare, hide", Seed: "memory-mismatch-v1", Players: 2,
		Viewers: []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, 0, 1}}
	probe, err := scenario.Run(memory.NewDelegate(), spec)
	if err != nil {
		return spec, err
	}
	var setup struct {
		Game struct {
			HiddenCards struct {
				Deck    string
				Indexes []int
			}
		}
	}
	blob, err := probe.History.State(probe.History.Game().Version)
	if err != nil {
		return spec, err
	}
	if err := json.Unmarshal(blob, &setup); err != nil {
		return spec, err
	}
	var chest struct {
		Decks map[string][]struct{ Values struct{ Type string } }
	}
	if err := json.Unmarshal(probe.Replay.Chest, &chest); err != nil {
		return spec, err
	}
	indexes := setup.Game.HiddenCards.Indexes
	deck := chest.Decks[setup.Game.HiddenCards.Deck]
	if len(indexes) < 2 {
		return spec, fmt.Errorf("Memory needs at least two cards")
	}
	first, second := -1, -1
	for slot, index := range indexes {
		if index < 0 || index >= len(deck) {
			continue
		}
		if first == -1 {
			first = slot
			continue
		}
		if deck[index].Values.Type != deck[indexes[first]].Values.Type {
			second = slot
			break
		}
	}
	if second < 0 {
		return spec, fmt.Errorf("Memory setup has no mismatching pair")
	}
	var game struct{ CurrentPlayerIndex boardgame.PlayerIndex }
	if err := json.Unmarshal(probe.Replay.Frames[0].Viewers[0].Game, &game); err != nil {
		return spec, err
	}
	spec.Steps = []scenario.Step{
		{Label: "Reveal the first card", Move: "Reveal Card", Player: game.CurrentPlayerIndex, Input: map[string]interface{}{"CardIndex": first}},
		{Label: "Reveal a different card", Move: "Reveal Card", Player: game.CurrentPlayerIndex, Input: map[string]interface{}{"CardIndex": second}},
		{Label: "Hide the pair and pass the turn", Timer: true},
	}
	return spec, nil
}
