// Export per-version viewer replays for the declarative-elements experiment.
//
// examples/memory/client/scenario-replay.json records only settled decision
// boundaries and never captures a pair, so it cannot show a card flying to a
// won-cards stack or the "both cards face up" state that Memory's
// post-animation-delay exists for. This command drives the real engine the
// same way boardgame-util/lib/scenario does (manual clock, seeded id and salt,
// in-memory filesystem storage) but writes one frame for EVERY state version,
// fix-up moves included, in the same JSON shape as scenario.Replay plus a
// `move` name per frame.
//
// It is additive tooling for fabric/declarative only. It changes no game.
//
//	./scripts/go-local run ./fabric/declarative/replays/cmd/export --out fabric/declarative/replays
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/filesystem"
)

type viewerSnapshot struct {
	Viewer boardgame.PlayerIndex `json:"viewer"`
	Game   json.RawMessage       `json:"game"`
	// Legality exists only at a settled decision boundary, where the engine
	// can project it. Fix-up versions in between have none.
	MoveLegality         map[string]api.ViewerMoveLegality `json:"moveLegality,omitempty"`
	ProjectedMoveChoices json.RawMessage                   `json:"projectedMoveChoices,omitempty"`
}

type frame struct {
	Label   string           `json:"label"`
	Version int              `json:"version"`
	Move    string           `json:"move"`
	Viewers []viewerSnapshot `json:"viewers"`
}

type replay struct {
	SchemaVersion int             `json:"schemaVersion"`
	Name          string          `json:"name"`
	GameName      string          `json:"gameName"`
	EveryVersion  bool            `json:"everyVersion"`
	Chest         json.RawMessage `json:"chest"`
	Frames        []frame         `json:"frames"`
}

type run struct {
	manager *boardgame.GameManager
	game    *boardgame.Game
	replay  *replay
	viewers []boardgame.PlayerIndex
	next    int
}

func start(delegate boardgame.GameDelegate, name, seed string, players int) (*run, error) {
	storage := filesystem.NewStorageManager("")
	storage.DebugNoDisk = true
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		return nil, err
	}
	if err := manager.Internals().UseManualClock(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		return nil, err
	}
	id := sha256.Sum256([]byte("boardgame scenario id:" + seed))
	salt := sha256.Sum256([]byte("boardgame scenario salt:" + seed))
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: delegate.Name(), ID: strings.ToUpper(hex.EncodeToString(id[:8])),
		SecretSalt: hex.EncodeToString(salt[:]), NumPlayers: players,
	})
	if err != nil {
		return nil, err
	}
	chest, err := json.Marshal(manager.Chest())
	if err != nil {
		return nil, err
	}
	r := &run{manager: manager, game: game,
		replay: &replay{SchemaVersion: 1, Name: name, GameName: delegate.Name(), EveryVersion: true, Chest: chest}}
	r.viewers = []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex}
	for i := 0; i < players; i++ {
		r.viewers = append(r.viewers, boardgame.PlayerIndex(i))
	}
	return r, r.capture("Setup")
}

// capture appends one frame per version committed since the last capture.
func (r *run) capture(label string) error {
	for ; r.next <= r.game.Version(); r.next++ {
		state := r.game.State(r.next)
		if state == nil {
			return fmt.Errorf("state %d unavailable", r.next)
		}
		f := frame{Label: label, Version: r.next}
		if r.next > 0 {
			move, err := r.game.Move(r.next)
			if err != nil {
				return err
			}
			f.Move = move.Info().Name()
		}
		for _, viewer := range r.viewers {
			view, err := r.game.JSONForPlayer(viewer, state)
			if err != nil {
				return err
			}
			blob, err := json.Marshal(view)
			if err != nil {
				return err
			}
			// The Go-rendered text diagram is large and no element reads it.
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(blob, &fields); err != nil {
				return err
			}
			delete(fields, "Diagram")
			// JSONForPlayer stamps the head version; a frame is one version.
			fields["Version"] = json.RawMessage(fmt.Sprint(r.next))
			if blob, err = json.Marshal(fields); err != nil {
				return err
			}
			snapshot := viewerSnapshot{Viewer: viewer, Game: blob}
			if r.next == r.game.Version() && r.game.AtProposalFrontier() {
				if snapshot.MoveLegality, err = api.MoveLegalityForViewer(r.game, viewer); err != nil {
					return err
				}
				if snapshot.ProjectedMoveChoices, err = api.ProjectMoveChoicesForViewer(r.game, viewer); err != nil {
					return err
				}
			}
			f.Viewers = append(f.Viewers, snapshot)
		}
		r.replay.Frames = append(r.replay.Frames, f)
	}
	return nil
}

func (r *run) move(label, name string, player boardgame.PlayerIndex, input map[string]interface{}) error {
	move, err := r.game.MoveWithInput(name, input)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err := <-r.game.ProposeMoveAtVersion(move, player, r.game.Version()); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return r.capture(label)
}

func (r *run) timer(label string) error {
	fired, err := r.manager.Internals().ForceNextTimerWithError()
	if err != nil {
		return err
	}
	if !fired {
		return fmt.Errorf("%s: no timer pending", label)
	}
	return r.capture(label)
}

// authoritative returns the unsanitized current state. It is used only to
// script the run and never reaches the replay.
func (r *run) authoritative(into interface{}) error {
	blob, err := json.Marshal(r.game.CurrentState())
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, into)
}

type memoryState struct {
	Game struct {
		CurrentPlayer boardgame.PlayerIndex
		HiddenCards   struct {
			Deck    string
			Indexes []int
		}
	}
}

// memoryPair finds two hidden slots that match (or differ) in Type.
func memoryPair(r *run, match bool) (boardgame.PlayerIndex, int, int, error) {
	var state memoryState
	if err := r.authoritative(&state); err != nil {
		return 0, 0, 0, err
	}
	var chest struct {
		Decks map[string][]struct{ Values struct{ Type string } }
	}
	if err := json.Unmarshal(r.replay.Chest, &chest); err != nil {
		return 0, 0, 0, err
	}
	deck := chest.Decks[state.Game.HiddenCards.Deck]
	indexes := state.Game.HiddenCards.Indexes
	for a, first := range indexes {
		if first < 0 {
			continue
		}
		for b := a + 1; b < len(indexes); b++ {
			second := indexes[b]
			if second < 0 {
				continue
			}
			if (deck[first].Values.Type == deck[second].Values.Type) == match {
				return state.Game.CurrentPlayer, a, b, nil
			}
		}
	}
	return 0, 0, 0, fmt.Errorf("no pair with match=%v", match)
}

func memoryTurn(r *run, match bool) error {
	player, a, b, err := memoryPair(r, match)
	if err != nil {
		return err
	}
	kind := "mismatch"
	if match {
		kind = "match"
	}
	if err := r.move(fmt.Sprintf("P%d reveals slot %d", player, a), "Reveal Card", player, map[string]interface{}{"CardIndex": a}); err != nil {
		return err
	}
	if err := r.move(fmt.Sprintf("P%d reveals slot %d (%s)", player, b, kind), "Reveal Card", player, map[string]interface{}{"CardIndex": b}); err != nil {
		return err
	}
	if !match {
		return r.timer("Hide timer fires")
	}
	return nil
}

func memoryReplay(name, seed string, turns []bool) (*replay, error) {
	r, err := start(memory.NewDelegate(), name, seed, 2)
	if err != nil {
		return nil, err
	}
	defer r.manager.Internals().Close()
	for _, match := range turns {
		if err := memoryTurn(r, match); err != nil {
			return r.replay, err
		}
	}
	return r.replay, nil
}

func blackjackReplay(name, seed string) (*replay, error) {
	r, err := start(blackjack.NewDelegate(), name, seed, 2)
	if err != nil {
		return nil, err
	}
	defer r.manager.Internals().Close()
	current := func() (boardgame.PlayerIndex, error) {
		var state struct {
			Game struct{ CurrentPlayer boardgame.PlayerIndex }
		}
		err := r.authoritative(&state)
		return state.Game.CurrentPlayer, err
	}
	// First seat hits once and stands if it survived; second seat stands.
	first, err := current()
	if err != nil {
		return r.replay, err
	}
	if err := r.move(fmt.Sprintf("P%d hits", first), "Current Player Hit", first, nil); err != nil {
		return r.replay, err
	}
	for i := 0; i < 2 && !r.game.Finished(); i++ {
		player, err := current()
		if err != nil {
			return r.replay, err
		}
		if err := r.move(fmt.Sprintf("P%d stands", player), "Current Player Stand", player, nil); err != nil {
			// The round may already have rolled over; keep the prefix.
			log.Printf("blackjack: stopping script: %v", err)
			break
		}
	}
	return r.replay, nil
}

func write(dir, file string, r *replay, err error) {
	if err != nil {
		log.Fatalf("%s: %v", file, err)
	}
	blob, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), append(blob, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s: %d frames, versions 0..%d", file, len(r.Frames), r.Frames[len(r.Frames)-1].Version)
}

func main() {
	out := flag.String("out", "fabric/declarative/replays", "output directory")
	flag.Parse()
	a, err := memoryReplay("Memory A: match, mismatch, match", "declarative-memory-a",
		[]bool{true, false, true})
	write(*out, "memory-a.json", a, err)
	b, err := memoryReplay("Memory B: mismatch, match, match", "declarative-memory-b",
		[]bool{false, true, true})
	write(*out, "memory-b.json", b, err)
	bj, err := blackjackReplay("Blackjack: deal, hit, stand", "declarative-blackjack")
	write(*out, "blackjack.json", bj, err)
}
