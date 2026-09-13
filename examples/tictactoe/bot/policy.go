// Package bot is a perfect-information policy using only detached viewer JSON.
package bot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame/bots"
)

// RandomEmpty chooses an empty square with the caller's injected randomness.
// Rules and turn validation remain in the engine; this example demonstrates the
// observation boundary, not a strong tic-tac-toe strategy.
type RandomEmpty struct{}

func (RandomEmpty) Decide(ctx context.Context, observation bots.Observation, random bots.Random) (bots.MoveSpec, error) {
	if err := ctx.Err(); err != nil {
		return bots.MoveSpec{}, err
	}
	var wire struct {
		CurrentState struct {
			Game struct{ Slots struct{ Indexes []int } }
		}
		CurrentPlayerIndex int
	}
	if err := json.Unmarshal(observation.Game, &wire); err != nil {
		return bots.MoveSpec{}, err
	}
	if wire.CurrentPlayerIndex != int(observation.Player) {
		return bots.MoveSpec{}, fmt.Errorf("not my turn")
	}
	var empty []int
	for slot, index := range wire.CurrentState.Game.Slots.Indexes {
		if index == -1 {
			empty = append(empty, slot)
		}
	}
	if len(empty) == 0 {
		return bots.MoveSpec{}, fmt.Errorf("no empty squares")
	}
	return bots.MoveSpec{Name: "Place Token", Input: map[string]interface{}{"Slot": empty[random.Intn(len(empty))]}}, nil
}
