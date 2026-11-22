// Command stockvec emits ../../stock/*.json.gz: gzipped conformance documents
// (boardgame-util/lib/conformance) of short games of ../stockgame and
// ../seatgame played on the real Go engine. Each records, at every version,
// the stored state, the move record and every viewer's view, legality (with
// Go's error texts) and choices, so rules/stock.test.ts can replay them on
// the TypeScript engine with the stock moves of rules/stock.ts.
//
// fixups.json beside them holds, for each document and version, each
// fix-up move's Legal error text for the admin, which a conformance document
// does not record.
//
// Every input is fixed (ids, salts, steps), so a rerun with no Go change
// writes the same bytes.
//
// Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/stockvec
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/seatgame"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/stockgame"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

const outDir = "server/static/src/engine/testdata/stock"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func move(label, name string, player boardgame.PlayerIndex, input map[string]interface{}) scenario.Step {
	if input == nil {
		input = map[string]interface{}{}
	}
	return scenario.Step{Label: label, Move: name, Player: player, Input: input}
}

func seat(player boardgame.PlayerIndex) scenario.Step {
	p := player
	return scenario.Step{Label: fmt.Sprintf("Seat player %d", player), Seat: &p}
}

func place(player boardgame.PlayerIndex, slot int) scenario.Step {
	return move(fmt.Sprintf("Player %d places on slot %d", player, slot), "Place Card", player, map[string]interface{}{"Slot": slot})
}

type doc struct {
	file        string
	newDelegate conformance.NewDelegate
	spec        scenario.Spec
	covers      string
}

func main() {
	stock := func() boardgame.GameDelegate { return stockgame.NewDelegate() }
	seats := func() boardgame.GameDelegate { return seatgame.NewDelegate() }
	docs := []doc{
		{
			file: "stockgame-full.json.gz", newDelegate: stock,
			covers: "every deal, collect and move-components move; ShuffleStack; StartPhase; DefaultComponent with and without a legal type; FinishTurn; Done; PlayerColor and ComponentColor; the game's end",
			spec: scenario.Spec{Name: "stockgame: a full two-player game", Seed: "stockgame-full-v1", Players: 2, Steps: []scenario.Step{
				place(0, 0), place(1, 1),
				move("Player 1 passes", "Pass", 1, nil),
				place(0, 2), place(1, 3), place(0, 4), place(1, 5),
				move("Player 0 calls cleanup", "Call Cleanup", 0, nil),
			}},
		},
		{
			file: "stockgame-three.json.gz", newDelegate: stock,
			covers: "the deal and collect moves with three players and an uneven deck; DefaultComponent refusing an odd slot",
			spec: scenario.Spec{Name: "stockgame: three players, a few placements", Seed: "stockgame-three-v1", Players: 3, Steps: []scenario.Step{
				place(0, 3), place(1, 0),
				move("Player 0 passes", "Pass", 0, nil),
				place(2, 2),
			}},
		},
		{
			file: "seatgame-seated.json.gz", newDelegate: seats,
			covers: "SeatPlayer, InactivateEmptySeat, ActivateFilledSeat, CloseAllSeats, ActivateEmptySeat, CloseEmptySeat",
			spec: scenario.Spec{Name: "seatgame: two of four seats filled", Seed: "seatgame-seated-v1", Players: 4, Steps: []scenario.Step{
				seat(0), seat(2),
				move("Player 2 closes the seats", "Close All Seats", 2, nil),
			}},
		},
		{
			file: "seatgame-slow.json.gz", newDelegate: seats,
			covers: "CloseAllSeats refused below its target count, then accepted",
			spec: scenario.Spec{Name: "seatgame: one seat, then a second", Seed: "seatgame-slow-v1", Players: 3, Steps: []scenario.Step{
				seat(1), seat(0),
				move("Player 0 closes the seats", "Close All Seats", 0, nil),
			}},
		},
		{
			file: "seatgame-unseated.json.gz", newDelegate: seats,
			covers: "the seat moves in a game that never seats players",
			spec:   scenario.Spec{Name: "seatgame: no seating", Seed: "seatgame-unseated-v1", Players: 2},
		},
	}
	check(os.MkdirAll(outDir, 0o755))
	var index []map[string]interface{}
	fixups := map[string][]map[string]string{}
	for _, d := range docs {
		document, err := conformance.FromScenario(d.newDelegate, d.spec, d.covers)
		check(err)
		blob, err := json.Marshal(document)
		check(err)
		var zipped bytes.Buffer
		w, err := gzip.NewWriterLevel(&zipped, gzip.BestCompression)
		check(err)
		_, err = w.Write(append(blob, '\n'))
		check(err)
		check(w.Close())
		check(os.WriteFile(filepath.Join(outDir, d.file), zipped.Bytes(), 0o644))
		index = append(index, map[string]interface{}{"file": d.file, "game": document.Game.Name, "name": d.spec.Name, "versions": len(document.Versions), "covers": d.covers})
		fixups[d.file] = fixUpLegality(d, document.Game.Seating)
	}
	blob, err := json.MarshalIndent(fixups, "", " ")
	check(err)
	check(os.WriteFile(filepath.Join(outDir, "fixups.json"), append(blob, '\n'), 0o644))
	blob, err = json.MarshalIndent(index, "", " ")
	check(err)
	check(os.WriteFile(filepath.Join(outDir, "index.json"), append(blob, '\n'), 0o644))
}

// seatingStorage answers the WillSeatPlayer rendezvous as the run did.
type seatingStorage struct {
	*memstore.StorageManager
	seating bool
}

func (s *seatingStorage) FetchInjectedDataForGame(gameID, dataType string) interface{} {
	if dataType == interfaces.WillSeatPlayerRendezvousDataType && s.seating {
		return true
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}

// fixUpLegality is what the conformance document cannot show: a fix-up's
// Legal error text (a fix-up is never in a viewer's legality). It reruns
// the scenario, stores its history in a fresh manager, and at every version
// binds each fix-up move to that state (MoveByNameForState, so its
// DefaultsForState runs) and records Legal's error for the admin; "" is
// legal.
func fixUpLegality(d doc, seating bool) []map[string]string {
	result, err := scenario.Run(d.newDelegate(), d.spec)
	check(err)
	history := result.History
	storage := &seatingStorage{StorageManager: memstore.NewStorageManager(), seating: seating}
	manager, err := boardgame.NewGameManager(d.newDelegate(), storage)
	check(err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	want := history.Game()
	for v := 0; v <= want.Version; v++ {
		blob, err := history.State(v)
		check(err)
		rec := *want
		rec.Version = v
		rec.Finished = false
		rec.Winners = nil
		var move *boardgame.MoveStorageRecord
		if v > 0 {
			move, err = history.Move(v)
			check(err)
		}
		check(storage.SaveGameAndCurrentState(&rec, blob, move))
	}
	game := manager.Game(want.ID)
	if game == nil {
		panic("game did not load")
	}
	var out []map[string]string
	for v := 0; v <= want.Version; v++ {
		state := game.State(v)
		row := map[string]string{}
		for _, m := range manager.ExampleMoves() {
			if !base.IsFixUp(m) {
				continue
			}
			bound := game.MoveByNameForState(m.Info().Name(), state)
			if bound == nil {
				panic("no move " + m.Info().Name())
			}
			msg := ""
			if err := bound.Legal(state, boardgame.AdminPlayerIndex); err != nil {
				msg = err.Error()
			}
			row[m.Info().Name()] = msg
		}
		out = append(out, row)
	}
	return out
}
