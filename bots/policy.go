// Package bots provides a narrow observation-and-decision boundary for creator
// bots. Policies receive detached viewer JSON, never an engine Game or State.
// This is an in-process API boundary, not isolation from arbitrary creator Go.
package bots

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame"
)

// Random supplies caller-owned randomness independent of the game RNG.
type Random interface{ Intn(int) int }

// Observation contains only the API state visible to Player and public static
// component data. JSON values are detached: editing them cannot mutate the game.
type Observation struct {
	Player  boardgame.PlayerIndex
	Version int
	Game    json.RawMessage
	Chest   json.RawMessage
}

// MoveSpec is one ordinary creator move. The caller binds context-owned fields.
type MoveSpec struct {
	Name  string
	Input map[string]interface{}
}

// Policy cooperatively honors ctx and returns a decision or an explanatory error.
// A policy can retain its own private memory, but receives no game capability.
type Policy interface {
	Decide(ctx context.Context, observation Observation, random Random) (MoveSpec, error)
}

type PolicyFunc func(context.Context, Observation, Random) (MoveSpec, error)

func (f PolicyFunc) Decide(ctx context.Context, observation Observation, random Random) (MoveSpec, error) {
	return f(ctx, observation, random)
}

// Observe captures one concrete player's settled view. It rejects observers and
// admin so accidentally supplying a privileged index cannot expose hidden state.
func Observe(game *boardgame.Game, player boardgame.PlayerIndex) (Observation, error) {
	if game == nil {
		return Observation{}, fmt.Errorf("bot requires a game")
	}
	if player < 0 || int(player) >= game.NumPlayers() {
		return Observation{}, fmt.Errorf("bot requires a configured player")
	}
	if !game.AtProposalFrontier() {
		return Observation{}, fmt.Errorf("bot requires a settled decision boundary")
	}
	view, err := game.JSONForPlayer(player, nil)
	if err != nil {
		return Observation{}, err
	}
	blob, err := json.Marshal(view)
	if err != nil {
		return Observation{}, err
	}
	chest, err := json.Marshal(game.Manager().Chest())
	if err != nil {
		return Observation{}, err
	}
	// Read version from the captured wire state, not a later game-head read.
	var version struct{ Version int }
	if err := json.Unmarshal(blob, &version); err != nil {
		return Observation{}, err
	}
	return Observation{Player: player, Version: version.Version, Game: blob, Chest: chest}, nil
}

// Play calls one policy and submits its ordinary move against the observed
// version. Cancellation is checked before and after Decide; it cannot interrupt
// arbitrary Go code or roll back a proposal already sent to the engine.
func Play(ctx context.Context, game *boardgame.Game, player boardgame.PlayerIndex, policy Policy, random Random) error {
	if ctx == nil || policy == nil || random == nil {
		return fmt.Errorf("bot requires context, policy, and randomness")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	observation, err := Observe(game, player)
	if err != nil {
		return err
	}
	spec, err := policy.Decide(ctx, observation, random)
	if err != nil {
		return fmt.Errorf("bot player %d version %d: %w", player, observation.Version, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// BuildMoveInputSchema excludes fix-ups and other engine-owned moves.
	schema, err := boardgame.BuildMoveInputSchema(game.Manager())
	if err != nil {
		return err
	}
	allowed := false
	for _, entry := range schema {
		allowed = allowed || entry.Name == spec.Name
	}
	if !allowed {
		return fmt.Errorf("bot selected unknown or engine-owned move %q", spec.Name)
	}
	move, err := game.MoveWithInput(spec.Name, spec.Input)
	if err != nil {
		return err
	}
	return <-game.ProposeMoveAtVersion(move, player, observation.Version)
}
