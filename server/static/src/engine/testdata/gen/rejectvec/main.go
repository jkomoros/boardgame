// Command rejectvec emits ../../rejections.vectors.json: what the Go engine
// does with proposals it refuses, which the conformance corpus cannot show
// (every proposal recorded there succeeded).
//
// Each case creates a game with a fixed id and salt, applies some accepted
// moves, then makes one attempt that Go rejects: a wrong proposer, a stale
// version, illegal arguments, input the binder refuses, a move after game
// over, an unknown move, a move from another game. The attempt's error text
// is recorded, and the generator itself asserts that the stored state and
// version did not change.
//
// Two further cases reach the fix-up cap (ErrTooManyFixUps) with a wrapped
// tictactoe that has an always-legal fix-up move. There Go DOES commit
// versions before failing; the vector records how many, and the final state.
//
// Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/rejectvec
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/interfaces"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// spinDelegate is tictactoe plus "Spin", a fix-up that is always legal:
// everywhere when phase is negative, otherwise only in that phase. It must
// forward the optional interfaces core finds by type assertion.
type spinDelegate struct {
	boardgame.GameDelegate
	phase int
}

// Name must match this package's name (NewGameManager checks it).
func (s *spinDelegate) Name() string { return "main" }

func (s *spinDelegate) ConfigureLegalTemplates() map[string]string {
	if tc, ok := s.GameDelegate.(legal.TemplateConfigurer); ok {
		return tc.ConfigureLegalTemplates()
	}
	return nil
}

func (s *spinDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(s)
	spin := auto.MustConfig(new(moves.NoOp), moves.WithMoveName("Spin"), moves.WithHelpText("A fix-up that is always legal"))
	if s.phase < 0 {
		return moves.Combine(s.GameDelegate.ConfigureMoves(), moves.Add(spin))
	}
	return moves.Combine(s.GameDelegate.ConfigureMoves(), moves.AddForPhase(enum.EnumKey(s.phase), spin))
}

// Step is one proposal. Fields holds every stored field of the move (as a
// replay supplies them); Input holds only what a proposer supplies, bound by
// Game.MoveWithInput. Exactly one of the two is used.
type Step struct {
	Name     string                 `json:"name"`
	Proposer int                    `json:"proposer"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	Input    map[string]interface{} `json:"input,omitempty"`
	// UseInput distinguishes an empty Input from an absent one.
	UseInput bool `json:"useInput,omitempty"`
	// ExpectedVersion, when >= 0, is passed to ProposeMoveAtVersion.
	ExpectedVersion int `json:"expectedVersion"`
	// From names another game whose move type is proposed here.
	From string `json:"from,omitempty"`
}

type Case struct {
	Name string `json:"name"`
	// Game is "memory", "blackjack", "tictactoe", "pig", "spin-always" or
	// "spin-after-first-move".
	Game       string            `json:"game"`
	ID         string            `json:"id"`
	SecretSalt string            `json:"secretSalt"`
	NumPlayers int               `json:"numPlayers"`
	Variant    map[string]string `json:"variant,omitempty"`
	// Seating is whether the game was told players will be seated.
	Seating bool `json:"seating,omitempty"`
	// Setup is applied first; every step of it is accepted.
	Setup []Step `json:"setup"`
	// Attempt is the proposal Go refuses. Absent when creation itself fails.
	Attempt *Step `json:"attempt,omitempty"`
	// Error is err.Error() of the refusal.
	Error string `json:"error"`
	// VersionBefore and VersionAfter are the game's version around the
	// attempt. They are equal except at the fix-up cap.
	VersionBefore int `json:"versionBefore"`
	VersionAfter  int `json:"versionAfter"`
	// FinalState is the stored state after the attempt, when it changed.
	FinalState json.RawMessage `json:"finalState,omitempty"`
	// FinalMoves are the names of the moves applied by the attempt.
	FinalMoves []string `json:"finalMoves,omitempty"`
}

// seatingStorage tells the game that players will be seated (Go's
// WillSeatPlayer rendezvous), as the server's storage does.
type seatingStorage struct {
	*memstore.StorageManager
}

func (s *seatingStorage) FetchInjectedDataForGame(gameID, dataType string) interface{} {
	if dataType == interfaces.WillSeatPlayerRendezvousDataType {
		return true
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}

func newManager(name string, seating bool) *boardgame.GameManager {
	var delegate boardgame.GameDelegate
	switch name {
	case "memory":
		delegate = memory.NewDelegate()
	case "blackjack":
		delegate = blackjack.NewDelegate()
	case "tictactoe":
		delegate = tictactoe.NewDelegate()
	case "pig":
		delegate = pig.NewDelegate()
	case "spin-always":
		delegate = &spinDelegate{GameDelegate: tictactoe.NewDelegate(), phase: -1}
	case "spin-after-first-move":
		delegate = &spinDelegate{GameDelegate: tictactoe.NewDelegate(), phase: 1}
	default:
		panic("unknown game " + name)
	}
	var storage boardgame.StorageManager = memstore.NewStorageManager()
	if seating {
		storage = &seatingStorage{memstore.NewStorageManager()}
	}
	manager, err := boardgame.NewGameManager(delegate, storage)
	check(err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	manager.Internals().UseManualTimers()
	return manager
}

func canonical(blob []byte) json.RawMessage {
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(blob))
	decoder.UseNumber()
	check(decoder.Decode(&value))
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	check(encoder.Encode(value))
	return bytes.TrimRight(out.Bytes(), "\n")
}

func storedState(game *boardgame.Game) []byte {
	blob, err := game.Manager().Storage().State(game.ID(), game.Version())
	check(err)
	return canonical(blob)
}

// propose makes one proposal and returns Go's error, or nil.
func propose(game *boardgame.Game, step Step) error {
	var move boardgame.Move
	if step.UseInput {
		bound, err := game.MoveWithInput(step.Name, step.Input)
		if err != nil {
			return err
		}
		move = bound
	} else if step.From != "" {
		// A move type of another game, proposed to this one.
		other := newManager(step.From, false)
		defer other.Internals().Close()
		otherGame, err := other.Internals().RecreateGame(&boardgame.GameStorageRecord{
			Name: other.Delegate().Name(), ID: "FFFFFFFFFFFFFFFF", SecretSalt: "other", NumPlayers: other.Delegate().DefaultNumPlayers(),
		})
		check(err)
		move = otherGame.MoveByName(step.Name)
		if move == nil {
			panic("no move " + step.Name + " in " + step.From)
		}
	} else {
		blob, err := json.Marshal(step.Fields)
		check(err)
		inflated, err := game.Manager().Internals().InflateMoveStorageRecord(&boardgame.MoveStorageRecord{Name: step.Name, Blob: blob}, game)
		if err != nil {
			return err
		}
		move = inflated
	}
	if step.ExpectedVersion >= 0 {
		return <-game.ProposeMoveAtVersion(move, boardgame.PlayerIndex(step.Proposer), step.ExpectedVersion)
	}
	return <-game.ProposeMove(move, boardgame.PlayerIndex(step.Proposer))
}

func run(c Case) Case {
	manager := newManager(c.Game, c.Seating)
	defer manager.Internals().Close()
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: manager.Delegate().Name(), ID: c.ID, SecretSalt: c.SecretSalt, NumPlayers: c.NumPlayers, Variant: c.Variant,
	})
	if c.Attempt == nil {
		// The case is about creation itself.
		if err == nil {
			panic(c.Name + ": creation was expected to fail")
		}
		c.Error = err.Error()
		// The game exists in storage even though creation reported failure.
		stored := manager.Game(c.ID)
		if stored == nil {
			panic(c.Name + ": no game was stored")
		}
		c.VersionAfter = stored.Version()
		c.FinalState = storedState(stored)
		records := stored.MoveRecords(stored.Version())
		for _, rec := range records {
			c.FinalMoves = append(c.FinalMoves, rec.Name)
		}
		return c
	}
	check(err)
	for i, step := range c.Setup {
		if err := propose(game, step); err != nil {
			panic(fmt.Sprintf("%s: setup step %d (%s) was refused: %v", c.Name, i, step.Name, err))
		}
	}
	c.VersionBefore = game.Version()
	before := storedState(game)
	finishedBefore := game.Finished()
	err = propose(game, *c.Attempt)
	if err == nil {
		panic(c.Name + ": the attempt was accepted")
	}
	c.Error = err.Error()
	stored := manager.Game(c.ID)
	c.VersionAfter = stored.Version()
	after := storedState(stored)
	if c.VersionAfter == c.VersionBefore {
		if !bytes.Equal(before, after) {
			panic(c.Name + ": the state changed although the version did not")
		}
		if stored.Finished() != finishedBefore {
			panic(c.Name + ": Finished changed on a refused proposal")
		}
		return c
	}
	c.FinalState = after
	for _, rec := range stored.MoveRecords(stored.Version())[c.VersionBefore:] {
		c.FinalMoves = append(c.FinalMoves, rec.Name)
	}
	return c
}

func fields(name string, proposer int, f map[string]interface{}) Step {
	return Step{Name: name, Proposer: proposer, Fields: f, ExpectedVersion: -1}
}

func input(name string, proposer int, in map[string]interface{}) *Step {
	return &Step{Name: name, Proposer: proposer, Input: in, UseInput: true, ExpectedVersion: -1}
}

type M = map[string]interface{}

func main() {
	const admin, observer = -2, -1
	place := func(player, slot int) Step {
		return fields("Place Token", player, M{"Slot": slot, "TargetPlayerIndex": player})
	}
	won := []Step{place(0, 0), place(1, 3), place(0, 1), place(1, 4), place(0, 2)}
	base := func(name, game string, players int) Case {
		return Case{Name: name, Game: game, ID: "ABCDEF0123456789", SecretSalt: "rejection-vectors-salt", NumPlayers: players}
	}
	with := func(c Case, setup []Step, attempt *Step) Case {
		c.Setup = setup
		if c.Setup == nil {
			c.Setup = []Step{}
		}
		c.Attempt = attempt
		return c
	}
	at := func(step *Step, version int) *Step {
		step.ExpectedVersion = version
		return step
	}
	ptr := func(step Step) *Step { return &step }
	seated := func(c Case) Case {
		c.Seating = true
		return c
	}

	cases := []Case{
		// ----- wrong proposer -------------------------------------------
		with(base("tictactoe: the other player proposes", "tictactoe", 2), nil, input("Place Token", 1, M{"Slot": 4})),
		with(base("tictactoe: a player names another player as the target", "tictactoe", 2), nil, ptr(fields("Place Token", 1, M{"Slot": 4, "TargetPlayerIndex": 0}))),
		with(base("tictactoe: a player moves for the wrong target", "tictactoe", 2), nil, ptr(fields("Place Token", 0, M{"Slot": 4, "TargetPlayerIndex": 1}))),
		with(base("tictactoe: the observer proposes", "tictactoe", 2), nil, input("Place Token", observer, M{"Slot": 4})),
		with(base("tictactoe: a proposer past the last seat", "tictactoe", 2), nil, input("Place Token", 5, M{"Slot": 4})),
		with(base("tictactoe: a proposer below every special index", "tictactoe", 2), nil, input("Place Token", -7, M{"Slot": 4})),
		with(base("memory: the other player reveals", "memory", 2), nil, input("Reveal Card", 1, M{"CardIndex": 0})),
		with(base("memory: the other player hides", "memory", 2), []Step{fields("Reveal Card", 0, M{"CardIndex": 0, "TargetPlayerIndex": 0}), fields("Reveal Card", 0, M{"CardIndex": 1, "TargetPlayerIndex": 0})}, input("Hide Cards", 1, M{})),
		with(base("blackjack: the next player hits out of turn", "blackjack", 3), nil, input("Current Player Hit", 1, M{})),
		with(base("blackjack: a player forces the turn to end", "blackjack", 3), nil, input("Force Finish Turn", 0, M{})),
		with(base("pig: the player whose turn it is not rolls", "pig", 2), nil, nil),

		// ----- stale version --------------------------------------------
		with(base("tictactoe: a proposal made against an old version", "tictactoe", 2), []Step{place(0, 0)}, at(input("Place Token", 1, M{"Slot": 4}), 0)),
		with(base("tictactoe: a proposal made against a future version", "tictactoe", 2), nil, at(input("Place Token", 0, M{"Slot": 4}), 3)),
		with(base("memory: a stale reveal", "memory", 2), []Step{fields("Reveal Card", 0, M{"CardIndex": 0, "TargetPlayerIndex": 0})}, at(input("Reveal Card", 0, M{"CardIndex": 1}), 0)),

		// ----- illegal arguments ----------------------------------------
		// Move names are matched without regard to case: the lower-case set-up step is accepted.
		with(base("tictactoe: an occupied slot", "tictactoe", 2), []Step{*input("place token", 0, M{"Slot": 4})}, input("PLACE TOKEN", 1, M{"Slot": 4})),
		with(base("tictactoe: a slot past the board", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": 9})),
		with(base("tictactoe: a negative slot", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": -1})),
		with(base("memory: a card index past the layout", "memory", 2), nil, input("Reveal Card", 0, M{"CardIndex": 99})),
		with(base("memory: a negative card index", "memory", 2), nil, input("Reveal Card", 0, M{"CardIndex": -1})),
		with(base("memory: a card that is already revealed", "memory", 2), []Step{fields("Reveal Card", 0, M{"CardIndex": 0, "TargetPlayerIndex": 0})}, input("Reveal Card", 0, M{"CardIndex": 0})),
		with(base("memory: a third reveal", "memory", 2), []Step{fields("Reveal Card", 0, M{"CardIndex": 0, "TargetPlayerIndex": 0}), fields("Reveal Card", 0, M{"CardIndex": 1, "TargetPlayerIndex": 0})}, input("Reveal Card", 0, M{"CardIndex": 2})),
		with(base("memory: hiding with cards still to reveal", "memory", 2), nil, input("Hide Cards", 0, M{})),
		with(base("pig: done, then done again after the turn has passed", "pig", 2), nil, nil),
		with(base("blackjack: standing again after the turn has passed", "blackjack", 3), []Step{fields("Current Player Stand", 0, M{"TargetPlayerIndex": 0})}, input("Current Player Stand", 0, M{})),

		// ----- input the binder refuses ---------------------------------
		with(base("tictactoe: a required field is missing", "tictactoe", 2), nil, input("Place Token", 0, M{})),
		with(base("tictactoe: an unknown field", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": 4, "Column": 1})),
		with(base("tictactoe: a context-owned field is supplied", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": 4, "TargetPlayerIndex": 0})),
		with(base("tictactoe: a field of the wrong type", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": "four"})),
		with(base("tictactoe: a fractional slot", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": 1.5})),
		with(base("tictactoe: a null field", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": nil})),
		with(base("tictactoe: a boolean for an integer", "tictactoe", 2), nil, input("Place Token", 0, M{"Slot": true})),
		with(base("memory: a field on a move that takes none", "memory", 2), nil, input("Hide Cards", 0, M{"CardIndex": 0})),

		// ----- unknown and engine-owned moves ---------------------------
		with(base("tictactoe: a move that does not exist", "tictactoe", 2), nil, input("No Such Move", 0, M{})),
		with(base("tictactoe: a fix-up move through the proposer's entry point", "tictactoe", 2), nil, input("Finish Turn", 0, M{})),
		with(base("tictactoe: a fix-up move proposed directly by a player", "tictactoe", 2), nil, ptr(fields("Finish Turn", 0, M{}))),
		with(base("tictactoe: a fix-up move proposed directly by the admin when it is not legal", "tictactoe", 2), nil, ptr(fields("Finish Turn", admin, M{}))),
		with(base("tictactoe: the seating move with no seat pending", "tictactoe", 2), nil, ptr(fields("Seat Player", admin, M{"TargetPlayerIndex": 0}))),
		with(base("tictactoe: another game's move", "tictactoe", 2), nil, &Step{Name: "Roll Dice", Proposer: 0, From: "pig", ExpectedVersion: -1}),
		with(base("blackjack: a move that is legal only in another phase", "blackjack", 3), nil, ptr(fields("Deal Initial Hidden Card", admin, M{}))),
		with(base("blackjack: a move of yet another phase", "blackjack", 3), nil, ptr(fields("Accumulate Scores", admin, M{}))),
		with(seated(base("blackjack, seating: a move out of its place in the progression", "blackjack", 3)), nil, ptr(fields("Begin Game", admin, M{}))),
		with(seated(base("blackjack, seating: waiting for players who are not there", "blackjack", 3)), nil, ptr(fields("Wait For Enough Players - Gathering", admin, M{}))),
		with(seated(base("blackjack, seating: a play move while gathering", "blackjack", 3)), nil, input("Current Player Hit", 0, M{})),
		with(seated(base("blackjack, seating: inactivating a seat before enough players came", "blackjack", 3)), nil, ptr(fields("Inactivate Empty Seat - Gathering", admin, M{"TargetPlayerIndex": 0}))),

		// ----- after game over ------------------------------------------
		with(base("tictactoe: a proposal after the game is won", "tictactoe", 2), won, input("Place Token", 1, M{"Slot": 5})),
		with(base("tictactoe: the admin proposes after the game is won", "tictactoe", 2), won, ptr(fields("Finish Turn", admin, M{}))),
		with(base("tictactoe: an unknown move after the game is won", "tictactoe", 2), won, input("No Such Move", 1, M{})),

		// ----- the fix-up cap -------------------------------------------
		with(base("the fix-up cap, reached by a player's move", "spin-after-first-move", 2), nil, input("Place Token", 0, M{"Slot": 4})),
		base("the fix-up cap, reached while the game is created", "spin-always", 2),
	}

	var out []Case
	for _, c := range cases {
		if c.Game == "pig" && c.Attempt == nil {
			// Built here because it needs the starting player, which the salt decides.
			manager := newManager("pig", false)
			game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{Name: "pig", ID: c.ID, SecretSalt: c.SecretSalt, NumPlayers: c.NumPlayers})
			check(err)
			current := int(manager.Delegate().CurrentPlayerIndex(game.CurrentState()))
			manager.Internals().Close()
			if c.Name == "pig: the player whose turn it is not rolls" {
				c.Setup = []Step{}
				c.Attempt = input("Roll Dice", 1-current, M{})
			} else {
				c.Setup = []Step{fields("Done Turn", current, M{"TargetPlayerIndex": current})}
				// Done ends the turn, so the same player is no longer current.
				c.Attempt = input("Done Turn", current, M{})
			}
		}
		out = append(out, run(c))
	}

	blob, err := json.MarshalIndent(map[string]interface{}{
		"comment": "GENERATED by server/static/src/engine/testdata/gen/rejectvec from the real Go engine. Do not edit; regenerate with: ./scripts/go-local run ./server/static/src/engine/testdata/gen/rejectvec",
		"cases":   out,
	}, "", "\t")
	check(err)
	check(os.WriteFile("server/static/src/engine/testdata/rejections.vectors.json", append(blob, '\n'), 0644))
}
