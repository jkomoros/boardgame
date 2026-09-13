// Package scenario runs bounded, explicit game scripts through the real engine.
// A run yields canonical regression history and sanitized viewer snapshots of
// settled player decisions. It does not sandbox creator Go or enumerate moves.
package scenario

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/bots"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/filesystem"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
)

// Spec defines a reproducible setup and an explicit sequence of decisions.
// Seed controls engine RNG identity; creator code must use State.Rand() too.
// Legacy agents are not enabled; explicit Bot steps use detached observations.
type Spec struct {
	Name    string
	Seed    string
	Players int
	Variant boardgame.Variant
	Viewers []boardgame.PlayerIndex
	Steps   []Step
	// MaxMoves includes setup/fix-ups, not just explicit script steps. Zero=256.
	MaxMoves int
}

// Step selects exactly one operation: a creator move, bot decision, timer, or seat.
// Input follows MoveWithInput's strict creator contract. Timer and seating
// actions reuse the same debug hooks as golden replay.
type Step struct {
	Label  string
	Move   string
	Player boardgame.PlayerIndex
	Input  map[string]interface{}
	Timer  bool
	Seat   *boardgame.PlayerIndex
	Bot    bots.Policy
}

type MoveLegality = api.ViewerMoveLegality

// ViewerSnapshot is a complete API game snapshot for exactly one viewer.
// MoveLegality describes default-bound moves, as the ordinary /info tray does;
// it is not a claim that every possible argument to that move is legal.
type ViewerSnapshot struct {
	Viewer               boardgame.PlayerIndex   `json:"viewer"`
	Game                 json.RawMessage         `json:"game"`
	MoveLegality         map[string]MoveLegality `json:"moveLegality"`
	ProjectedMoveChoices json.RawMessage         `json:"projectedMoveChoices,omitempty"`
}

type Frame struct {
	Label   string           `json:"label"`
	Version int              `json:"version"`
	Viewers []ViewerSnapshot `json:"viewers"`
}

// Replay is the portable visual artifact. It excludes authoritative history,
// move inputs, and RNG salts. Each ViewerSnapshot must still be shown only to
// its intended viewer; a combined multi-viewer artifact is a local review tool.
type Replay struct {
	SchemaVersion              int             `json:"schemaVersion"`
	Name                       string          `json:"name"`
	GameName                   string          `json:"gameName"`
	Chest                      json.RawMessage `json:"chest"`
	MoveInputSchemaFingerprint string          `json:"moveInputSchemaFingerprint"`
	Frames                     []Frame         `json:"frames"`
}

type Result struct {
	Replay Replay
	// History is the existing filesystem record format, suitable for goldens.
	// It contains authoritative hidden state and is never part of Replay.
	History *record.Record
}

// Run executes synchronously, returning the completed prefix with any error.
// Its limits bound committed moves and script steps, not time spent inside an
// arbitrary creator callback. Use process isolation for a hard execution limit.
func Run(delegate boardgame.GameDelegate, spec Spec) (*Result, error) {
	return RunContext(context.Background(), delegate, spec)
}

// RunContext checks cancellation between operations and passes it to bot policies.
// Cancellation does not interrupt creator Apply/Legal callbacks.
func RunContext(ctx context.Context, delegate boardgame.GameDelegate, spec Spec) (*Result, error) {
	if ctx == nil {
		return nil, fmt.Errorf("scenario requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if delegate == nil || spec.Name == "" {
		return nil, fmt.Errorf("scenario requires a delegate and name")
	}
	if len(spec.Steps) > 256 {
		return nil, fmt.Errorf("scenario exceeds 256 explicit steps")
	}
	if spec.MaxMoves == 0 {
		spec.MaxMoves = 256
	}
	if spec.MaxMoves < 1 || spec.MaxMoves > 4096 {
		return nil, fmt.Errorf("MaxMoves must be between 1 and 4096")
	}
	for i, step := range spec.Steps {
		operations := 0
		if step.Move != "" {
			operations++
		}
		if step.Bot != nil {
			operations++
		}
		if step.Timer {
			operations++
		}
		if step.Seat != nil {
			operations++
		}
		if operations != 1 || (step.Move == "" && len(step.Input) != 0) {
			return nil, fmt.Errorf("step %d must select exactly one operation", i+1)
		}
	}
	recordStorage := filesystem.NewStorageManager("")
	recordStorage.DebugNoDisk = true
	storage := &scenarioStorage{StorageManager: recordStorage, maxMoves: spec.MaxMoves}
	for _, step := range spec.Steps {
		storage.seating = storage.seating || step.Seat != nil
	}
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		return nil, fmt.Errorf("scenario setup: %w", err)
	}
	defer manager.Internals().Close()
	if err := manager.Internals().UseManualClock(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		return nil, err
	}
	if spec.Players == 0 {
		spec.Players = delegate.DefaultNumPlayers()
	}
	seed := spec.Seed
	if seed == "" {
		seed = spec.Name
	}
	id := sha256.Sum256([]byte("boardgame scenario id:" + seed))
	randomSeed := sha256.Sum256([]byte("boardgame scenario bot rng:" + seed))
	random := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(randomSeed[:8]))))
	salt := sha256.Sum256([]byte("boardgame scenario salt:" + seed))
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: delegate.Name(), ID: strings.ToUpper(hex.EncodeToString(id[:8])),
		SecretSalt: hex.EncodeToString(salt[:]), NumPlayers: spec.Players, Variant: spec.Variant,
	})
	history, _ := storage.RecordForID(strings.ToUpper(hex.EncodeToString(id[:8])))
	result := &Result{History: history, Replay: Replay{SchemaVersion: 1, Name: spec.Name, GameName: delegate.Name()}}
	if err != nil {
		return result, fmt.Errorf("scenario setup: %w", err)
	}
	result.Replay.Chest, err = json.Marshal(manager.Chest())
	if err != nil {
		return result, err
	}
	result.Replay.MoveInputSchemaFingerprint, err = boardgame.MoveInputSchemaFingerprint(manager)
	if err != nil {
		return result, err
	}
	if len(spec.Viewers) == 0 {
		spec.Viewers = []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex}
	}
	for _, viewer := range spec.Viewers {
		if viewer < boardgame.ObserverPlayerIndex || int(viewer) >= spec.Players {
			return result, fmt.Errorf("viewer %d must be observer or a configured player", viewer)
		}
	}
	schema, err := boardgame.BuildMoveInputSchema(manager)
	if err != nil {
		return result, err
	}
	capture := func(label string) error {
		if !game.AtProposalFrontier() {
			return fmt.Errorf("version %d is not a settled decision boundary", game.Version())
		}
		frame := Frame{Label: label, Version: game.Version()}
		for _, viewer := range spec.Viewers {
			view, err := game.JSONForPlayer(viewer, nil)
			if err != nil {
				return err
			}
			blob, err := json.Marshal(view)
			if err != nil {
				return err
			}
			legality, err := api.MoveLegalityForViewer(game, viewer)
			if err != nil {
				return err
			}
			choices, err := api.ProjectMoveChoicesForViewer(game, viewer)
			if err != nil {
				return err
			}
			frame.Viewers = append(frame.Viewers, ViewerSnapshot{Viewer: viewer, Game: blob, MoveLegality: legality, ProjectedMoveChoices: choices})
		}
		result.Replay.Frames = append(result.Replay.Frames, frame)
		return nil
	}
	if err := capture("Setup"); err != nil {
		return result, err
	}
	for i, step := range spec.Steps {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		label := step.Label
		if label == "" {
			label = step.Move
			if label == "" {
				label = fmt.Sprintf("Step %d", i+1)
			}
		}
		switch {
		case step.Bot != nil:
			err = bots.Play(ctx, game, step.Player, step.Bot, random)
		case step.Timer:
			var fired bool
			fired, err = manager.Internals().ForceNextTimerWithError()
			if err == nil && !fired {
				err = fmt.Errorf("no timer is pending")
			}
		case step.Seat != nil:
			if *step.Seat < 0 || int(*step.Seat) >= spec.Players {
				err = fmt.Errorf("invalid seat %d", *step.Seat)
				break
			}
			storage.pendingSeat = &seat{index: *step.Seat, storage: storage}
			err = <-manager.Internals().ForceFixUp(game)
			if err == nil && storage.pendingSeat != nil {
				err = fmt.Errorf("game did not accept seat %d", *step.Seat)
			}
		default:
			if step.Player < 0 || int(step.Player) >= spec.Players {
				err = fmt.Errorf("move requires a configured player")
				break
			}
			var move boardgame.Move
			move, err = game.MoveWithInput(step.Move, step.Input)
			if err == nil {
				playerMove := false
				for _, entry := range schema {
					playerMove = playerMove || entry.Name == move.Info().Name()
				}
				if !playerMove {
					err = fmt.Errorf("%q is not a creator move", step.Move)
				} else {
					err = <-game.ProposeMoveAtVersion(move, step.Player, game.Version())
				}
			}
		}
		if err != nil {
			return result, fmt.Errorf("scenario %q step %d (%s), committed version %d: %w", spec.Name, i+1, label, game.Version(), err)
		}
		if err := capture(label); err != nil {
			return result, fmt.Errorf("step %d capture: %w", i+1, err)
		}
	}
	result.History.SetDescription(spec.Name)
	return result, nil
}

type scenarioStorage struct {
	*filesystem.StorageManager
	maxMoves    int
	seating     bool
	pendingSeat *seat
}

func (s *scenarioStorage) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	if game.Version > s.maxMoves {
		return fmt.Errorf("scenario committed-move budget %d exhausted", s.maxMoves)
	}
	return s.StorageManager.SaveGameAndCurrentState(game, state, move)
}

func (s *scenarioStorage) FetchInjectedDataForGame(gameID, dataType string) interface{} {
	if dataType == interfaces.WillSeatPlayerRendezvousDataType && s.seating {
		return true
	}
	if dataType == interfaces.PlayerToSeatRendezvousDataType && s.pendingSeat != nil {
		return s.pendingSeat
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}

type seat struct {
	index   boardgame.PlayerIndex
	storage *scenarioStorage
}

func (s *seat) SeatIndex() boardgame.PlayerIndex { return s.index }
func (s *seat) Committed()                       { s.storage.pendingSeat = nil }
