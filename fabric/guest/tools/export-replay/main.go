// Export a complete Memory game as a guest replay: every state version, for
// every viewer, with the sanitized move record that produced it.
//
// boardgame-util/lib/scenario captures only settled decision boundaries and
// no move records, which is the right artifact for renderer review but skips
// the fix-up versions the animator needs. This tool reuses the scenario
// package's deterministic setup (seeded id and salt, manual clock) and the
// ordinary API sanitizers, and changes no game semantics.
//
// Run from the repository root:
//
//	./scripts/go-local run ./fabric/guest/tools/export-replay -out fabric/guest/replays/memory-full.json
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/filesystem"
)

type viewerEntry struct {
	Game json.RawMessage `json:"game"`
	// Move is the sanitized record of the move that produced this version;
	// null for version 0.
	Move json.RawMessage `json:"move"`
	// The three fields below exist only at settled decision boundaries.
	MoveLegality         map[string]api.ViewerMoveLegality `json:"moveLegality,omitempty"`
	ZeroInputMoves       []string                          `json:"zeroInputMoves,omitempty"`
	ProjectedMoveChoices json.RawMessage                   `json:"projectedMoveChoices,omitempty"`
}

type versionEntry struct {
	Version int `json:"version"`
	// Settled is true when the engine was waiting for a player at this
	// version, false for a version a fix-up move immediately replaced.
	Settled bool `json:"settled"`
	// Label describes the scripted step that ended at this version.
	Label   string                 `json:"label,omitempty"`
	Viewers map[string]viewerEntry `json:"viewers"`
}

type replay struct {
	SchemaVersion              int             `json:"schemaVersion"`
	Name                       string          `json:"name"`
	GameName                   string          `json:"gameName"`
	NumPlayers                 int             `json:"numPlayers"`
	Chest                      json.RawMessage `json:"chest"`
	MoveInputSchemaFingerprint string          `json:"moveInputSchemaFingerprint"`
	Versions                   []versionEntry  `json:"versions"`
}

type settled struct {
	label    string
	legality map[string]map[string]api.ViewerMoveLegality
	choices  map[string]json.RawMessage
	timers   json.RawMessage
}

func main() {
	out := flag.String("out", "memory-full.json", "replay output path")
	seed := flag.String("seed", "fabric-guest-memory-v1", "deterministic seed")
	flag.Parse()
	if err := run(*out, *seed); err != nil {
		log.Fatal(err)
	}
}

func run(out, seed string) error {
	delegate := memory.NewDelegate()
	storage := filesystem.NewStorageManager("")
	storage.DebugNoDisk = true
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		return err
	}
	defer manager.Internals().Close()
	if err := manager.Internals().UseManualClock(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		return err
	}
	id := sha256.Sum256([]byte("boardgame scenario id:" + seed))
	salt := sha256.Sum256([]byte("boardgame scenario salt:" + seed))
	const players = 2
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: delegate.Name(), ID: strings.ToUpper(hex.EncodeToString(id[:8])),
		SecretSalt: hex.EncodeToString(salt[:]), NumPlayers: players,
		Variant: boardgame.Variant{"numcards": "small", "cardset": "general"},
	})
	if err != nil {
		return err
	}
	viewers := []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, 0, 1}
	schema, err := boardgame.BuildMoveInputSchema(manager)
	if err != nil {
		return err
	}
	zeroInput := map[string]bool{}
	for _, move := range schema {
		hasInput := false
		for _, field := range move.Fields {
			if field.Disposition == string(boardgame.MoveInputRequired) || field.Disposition == string(boardgame.MoveInputUnsupported) {
				hasInput = true
			}
		}
		zeroInput[move.Name] = !hasInput
	}

	frontier := map[int]*settled{}
	capture := func(label string) error {
		if !game.AtProposalFrontier() {
			return fmt.Errorf("version %d is not settled", game.Version())
		}
		entry := &settled{label: label, legality: map[string]map[string]api.ViewerMoveLegality{}, choices: map[string]json.RawMessage{}}
		for _, viewer := range viewers {
			key := strconv.Itoa(int(viewer))
			legality, err := api.MoveLegalityForViewer(game, viewer)
			if err != nil {
				return err
			}
			entry.legality[key] = legality
			choices, err := api.ProjectMoveChoicesForViewer(game, viewer)
			if err != nil {
				return err
			}
			entry.choices[key] = choices
		}
		// Game.JSONForPlayer reports timers as of the call, not as of the
		// state it is given, so they have to be read while this version is
		// the head.
		live, err := game.JSONForPlayer(boardgame.ObserverPlayerIndex, nil)
		if err != nil {
			return err
		}
		liveBlob, err := json.Marshal(live)
		if err != nil {
			return err
		}
		var liveObject struct{ ActiveTimers json.RawMessage }
		if err := json.Unmarshal(liveBlob, &liveObject); err != nil {
			return err
		}
		entry.timers = liveObject.ActiveTimers
		frontier[game.Version()] = entry
		return nil
	}
	if err := capture("Setup"); err != nil {
		return err
	}

	// The script: the first two turns end in a mismatch (hidden once by the
	// explicit move, once by the timer); every later turn is a match, so the
	// game runs to its end and exercises capture fix-ups and the outcome.
	mismatches := 0
	for step := 0; !game.Finished(); step++ {
		if step > 200 {
			return fmt.Errorf("script did not finish the game")
		}
		var state struct {
			Game struct {
				CurrentPlayer boardgame.PlayerIndex
				HiddenCards   struct{ Indexes []int }
				VisibleCards  struct{ Indexes []int }
			}
		}
		blob, err := json.Marshal(game.CurrentState())
		if err != nil {
			return err
		}
		if err := json.Unmarshal(blob, &state); err != nil {
			return err
		}
		var visible []int
		for _, index := range state.Game.VisibleCards.Indexes {
			if index >= 0 {
				visible = append(visible, index)
			}
		}
		player := state.Game.CurrentPlayer
		reveal := func(slot int, label string) error {
			move, err := game.MoveWithInput("Reveal Card", map[string]interface{}{"CardIndex": slot})
			if err != nil {
				return err
			}
			if err := <-game.ProposeMoveAtVersion(move, player, game.Version()); err != nil {
				return fmt.Errorf("%s: %w", label, err)
			}
			return capture(label)
		}
		cardType := func(index int) (string, error) {
			return cardTypeFromChest(manager, index)
		}
		switch len(visible) {
		case 0:
			for slot, index := range state.Game.HiddenCards.Indexes {
				if index >= 0 {
					if err := reveal(slot, fmt.Sprintf("Player %d reveals a card", player)); err != nil {
						return err
					}
					break
				}
			}
		case 1:
			want, err := cardType(visible[0])
			if err != nil {
				return err
			}
			wantMatch := mismatches >= 2
			chosen := -1
			for slot, index := range state.Game.HiddenCards.Indexes {
				if index < 0 {
					continue
				}
				got, err := cardType(index)
				if err != nil {
					return err
				}
				if (got == want) == wantMatch {
					chosen = slot
					break
				}
			}
			if chosen < 0 {
				return fmt.Errorf("no suitable second card")
			}
			label := fmt.Sprintf("Player %d reveals the match", player)
			if !wantMatch {
				label = fmt.Sprintf("Player %d reveals a different card", player)
			}
			if err := reveal(chosen, label); err != nil {
				return err
			}
		default:
			// Two mismatched cards are showing and the engine is waiting.
			mismatches++
			if mismatches == 1 {
				move, err := game.MoveWithInput("Hide Cards", nil)
				if err != nil {
					return err
				}
				if err := <-game.ProposeMoveAtVersion(move, player, game.Version()); err != nil {
					return fmt.Errorf("hide cards: %w", err)
				}
				if err := capture(fmt.Sprintf("Player %d hides the cards", player)); err != nil {
					return err
				}
			} else {
				fired, err := manager.Internals().ForceNextTimerWithError()
				if err != nil {
					return err
				}
				if !fired {
					return fmt.Errorf("no timer is pending")
				}
				if err := capture("The hide timer fires"); err != nil {
					return err
				}
			}
		}
	}

	chest, err := json.Marshal(manager.Chest())
	if err != nil {
		return err
	}
	fingerprint, err := boardgame.MoveInputSchemaFingerprint(manager)
	if err != nil {
		return err
	}
	result := replay{SchemaVersion: 1, Name: "Memory: a complete small game", GameName: delegate.Name(),
		NumPlayers: players, Chest: chest, MoveInputSchemaFingerprint: fingerprint}
	// Read storage directly: Game.MoveRecords caches a prefix that can lag
	// the head by the final fix-up.
	records, err := manager.Storage().Moves(game.ID(), 0, game.Version())
	if err != nil {
		return err
	}
	byVersion := map[int]*boardgame.MoveStorageRecord{}
	for _, record := range records {
		if record == nil {
			continue
		}
		byVersion[record.Version] = record
	}
	for version := 0; version <= game.Version(); version++ {
		state := game.State(version)
		if state == nil {
			return fmt.Errorf("state %d is unavailable", version)
		}
		entry := versionEntry{Version: version, Viewers: map[string]viewerEntry{}}
		at := frontier[version]
		if at != nil {
			entry.Settled = true
			entry.Label = at.label
		}
		for _, viewer := range viewers {
			key := strconv.Itoa(int(viewer))
			view, err := game.JSONForPlayer(viewer, state)
			if err != nil {
				return err
			}
			gameBlob, err := json.Marshal(view)
			if err != nil {
				return err
			}
			// The Go-rendered diagram is a debugging aid the guest never shows.
			var object map[string]json.RawMessage
			if err := json.Unmarshal(gameBlob, &object); err != nil {
				return err
			}
			delete(object, "Diagram")
			// JSONForPlayer fills these four from the head of the game, not
			// from the state it was given. Restate them for this version.
			final := version == game.Version()
			object["Version"] = json.RawMessage(strconv.Itoa(version))
			if !final {
				object["Finished"] = json.RawMessage("false")
				object["Winners"] = json.RawMessage("null")
			}
			object["ActiveTimers"] = json.RawMessage("{}")
			if at != nil && len(at.timers) > 0 {
				object["ActiveTimers"] = at.timers
			}
			if gameBlob, err = json.Marshal(object); err != nil {
				return err
			}
			moveBlob := json.RawMessage("null")
			if record := byVersion[version]; record != nil {
				sanitized, err := game.MoveJSONForPlayer(viewer, record)
				if err != nil {
					return err
				}
				if moveBlob, err = json.Marshal(sanitized); err != nil {
					return err
				}
			}
			item := viewerEntry{Game: gameBlob, Move: moveBlob}
			if at != nil {
				item.MoveLegality = at.legality[key]
				item.ProjectedMoveChoices = at.choices[key]
				for name := range at.legality[key] {
					if zeroInput[name] {
						item.ZeroInputMoves = append(item.ZeroInputMoves, name)
					}
				}
			}
			entry.Viewers[key] = item
		}
		result.Versions = append(result.Versions, entry)
	}
	blob, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return os.WriteFile(out, append(blob, '\n'), 0644)
}

func cardTypeFromChest(manager *boardgame.GameManager, index int) (string, error) {
	blob, err := json.Marshal(manager.Chest())
	if err != nil {
		return "", err
	}
	var chest struct {
		Decks map[string][]struct{ Values struct{ Type string } }
	}
	if err := json.Unmarshal(blob, &chest); err != nil {
		return "", err
	}
	deck := chest.Decks["cards"]
	if index < 0 || index >= len(deck) {
		return "", fmt.Errorf("card index %d is out of range", index)
	}
	return deck[index].Values.Type, nil
}
