package bots_test

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/bots"
	gameMemory "github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/storage/memory"
)

func newGame(t *testing.T) *boardgame.Game {
	t.Helper()
	manager, err := boardgame.NewGameManager(gameMemory.NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatal(err)
	}
	manager.Internals().UseManualTimers()
	t.Cleanup(manager.Internals().Close)
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	return game
}

func TestObservationIsDetachedAndHidesCardPositions(t *testing.T) {
	game := newGame(t)
	for _, player := range []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, boardgame.AdminPlayerIndex, boardgame.PlayerIndex(game.NumPlayers())} {
		if _, err := bots.Observe(game, player); err == nil {
			t.Fatalf("accepted privileged/invalid player %d", player)
		}
	}
	observation, err := bots.Observe(game, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"SecretSalt", "SecretMoveCount", "\"Timers\""} {
		if strings.Contains(string(observation.Game), forbidden) {
			t.Fatalf("exposed %s", forbidden)
		}
	}
	var wire struct {
		CurrentState struct {
			Game struct{ HiddenCards struct{ Indexes []int } }
		}
	}
	if err := json.Unmarshal(observation.Game, &wire); err != nil {
		t.Fatal(err)
	}
	occupied := 0
	for _, index := range wire.CurrentState.Game.HiddenCards.Indexes {
		if index >= 0 {
			t.Fatal("hidden card location disclosed")
		}
		if index == -2 {
			occupied++
		}
	}
	if occupied == 0 {
		t.Fatal("test must include hidden occupied cards")
	}
	observation.Game[0] = 'x'
	fresh, err := bots.Observe(game, 0)
	if err != nil || !json.Valid(fresh.Game) {
		t.Fatal("policy buffer mutation reached engine")
	}
}

func TestPolicyCancellationAndErrorsDoNotSubmit(t *testing.T) {
	for _, mode := range []string{"cancel", "error", "context field", "fixup"} {
		t.Run(mode, func(t *testing.T) {
			game := newGame(t)
			version := game.Version()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sentinel := errors.New("no decision")
			policy := bots.PolicyFunc(func(context.Context, bots.Observation, bots.Random) (bots.MoveSpec, error) {
				spec := bots.MoveSpec{Name: "Reveal Card", Input: map[string]interface{}{"CardIndex": 0}}
				switch mode {
				case "cancel":
					cancel()
				case "error":
					return spec, sentinel
				case "context field":
					spec.Input["TargetPlayerIndex"] = 1
				case "fixup":
					spec.Name = "Capture Cards"
				}
				return spec, nil
			})
			err := bots.Play(ctx, game, game.CurrentState().CurrentPlayerIndex(), policy, rand.New(rand.NewSource(1)))
			if err == nil {
				t.Fatal("accepted invalid/canceled decision")
			}
			if mode == "error" && !errors.Is(err, sentinel) {
				t.Fatalf("lost original error: %v", err)
			}
			if game.Version() != version {
				t.Fatal("failed decision committed")
			}
		})
	}
}

func TestDecisionCannotSubmitAgainstAnAdvancedVersion(t *testing.T) {
	game := newGame(t)
	player := game.CurrentState().CurrentPlayerIndex()
	policy := bots.PolicyFunc(func(context.Context, bots.Observation, bots.Random) (bots.MoveSpec, error) {
		// Simulate another actor committing while the bot is thinking.
		move, err := game.MoveWithInput("Reveal Card", map[string]interface{}{"CardIndex": 0})
		if err != nil {
			return bots.MoveSpec{}, err
		}
		if err := <-game.ProposeMove(move, player); err != nil {
			return bots.MoveSpec{}, err
		}
		return bots.MoveSpec{Name: "Reveal Card", Input: map[string]interface{}{"CardIndex": 1}}, nil
	})
	version := game.Version()
	if err := bots.Play(context.Background(), game, player, policy, rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("stale decision accepted")
	}
	if game.Version() != version+1 {
		t.Fatalf("stale proposal changed history: %d", game.Version())
	}
}
