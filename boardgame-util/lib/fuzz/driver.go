package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/filesystem"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/sirupsen/logrus"
)

// Refusal reasons: why the driver expected, or found, the engine to refuse.
const (
	// ReasonIllegalInput is a candidate that bound (MoveWithInput succeeded)
	// but that Legal refused for every seated player.
	ReasonIllegalInput = "illegal-input"
	// ReasonWrongActor is a candidate Legal accepts for some player,
	// proposed by a player for whom Legal refuses it.
	ReasonWrongActor = "wrong-actor"
	// ReasonStaleVersion is a legal candidate proposed at the previous
	// version.
	ReasonStaleVersion = "stale-version"
	// ReasonApplyError is a legal candidate (Legal passed) whose proposal
	// the engine still refused, typically because the move's Apply returned
	// an error.
	ReasonApplyError = "apply-error"
	// ReasonTimerError is a forced timer whose completion move the engine
	// refused.
	ReasonTimerError = "timer-error"
)

// Refusal is one refused input: the document's Rejection plus why.
type Refusal struct {
	conformance.Rejection
	Reason string `json:"reason"`
}

// MoveStats counts what happened to one move type in a run.
type MoveStats struct {
	// Applied counts versions this move produced, fix-ups included.
	Applied int `json:"applied"`
	// ProposedLegally counts proposals the driver sent believing them
	// legal (Legal passed), whether or not they then applied.
	ProposedLegally int `json:"proposedLegally"`
	// Refused counts refusals that named this move.
	Refused int `json:"refused"`
}

// End reasons.
const (
	EndFinished     = "finished"
	EndNoAction     = "no-action"
	EndMaxProposals = "max-proposals"
	EndMaxVersions  = "max-versions"
)

// Result is a finished run.
type Result struct {
	// Config is the run's configuration with defaults filled in.
	Config  Config
	Derived Derived
	// History is the committed history, in the filesystem record format.
	History  *record.Record `json:"-"`
	Refusals []Refusal
	// End is why the run stopped (End* constants).
	End string
	// Proposals counts outside inputs sent, refused ones included.
	Proposals int
	// Moves has an entry for every configured move type.
	Moves map[string]*MoveStats
	// MoveOrder is the move table's order.
	MoveOrder []string
	// Uncovered names move fields whose domain is only their default value
	// (strings), as "Move.Field".
	Uncovered []string
	// LegalPanics counts candidate legality checks that panicked; such a
	// candidate is never proposed.
	LegalPanics int
}

// Run runs cfg.Game, which must be one of Games().
func Run(cfg Config) (*Result, error) {
	newDelegate := Delegate(cfg.Game)
	if newDelegate == nil {
		return nil, fmt.Errorf("unknown game %q; known: %v", cfg.Game, Games())
	}
	return RunWithDelegate(newDelegate, cfg)
}

// RunWithDelegate runs a game built by newDelegate. cfg.Game must equal its
// Name().
func RunWithDelegate(newDelegate conformance.NewDelegate, cfg Config) (*Result, error) {
	cfg = cfg.withDefaults()
	if cfg.MaxProposals < 0 || cfg.MaxVersions < 0 || cfg.RefusalRate > 1 {
		return nil, fmt.Errorf("invalid bounds in %+v", cfg)
	}
	storage := &driverStorage{StorageManager: newRecordStorage()}
	manager, err := newManager(newDelegate, storage)
	if err != nil {
		return nil, err
	}
	defer manager.Internals().Close()
	if manager.Delegate().Name() != cfg.Game {
		return nil, fmt.Errorf("delegate is %q, config names %q", manager.Delegate().Name(), cfg.Game)
	}
	seatMove := seatMoveName(manager)
	derived, err := Derive(manager, cfg.Seed, seatMove != "")
	if err != nil {
		return nil, err
	}
	storage.seating = derived.Seated
	schema, err := boardgame.BuildMoveInputSchema(manager)
	if err != nil {
		return nil, err
	}
	d := &driver{
		cfg: cfg, manager: manager, storage: storage, derived: derived,
		random: derived.generator(), schema: schema, seatMove: seatMove,
		result: &Result{Config: cfg, Derived: *derived, Moves: map[string]*MoveStats{}},
	}
	for _, move := range manager.ExampleMoves() {
		name := move.Info().Name()
		d.result.MoveOrder = append(d.result.MoveOrder, name)
		d.result.Moves[name] = &MoveStats{}
	}
	if err := d.policySetUp(); err != nil {
		return nil, err
	}
	d.game, err = manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: cfg.Game, ID: derived.ID, SecretSalt: derived.SecretSalt,
		NumPlayers: derived.NumPlayers, Variant: derived.Variant,
	})
	if err != nil {
		return nil, fmt.Errorf("create game: %w", err)
	}
	if err := d.policyStart(); err != nil {
		return nil, err
	}
	if err := d.loop(); err != nil {
		return nil, fmt.Errorf("%s seed %d at version %d: %w", cfg.Game, cfg.Seed, d.game.Version(), err)
	}
	history, err := storage.RecordForID(derived.ID)
	if err != nil {
		return nil, err
	}
	d.result.History = history
	for version := 1; version <= history.Game().Version; version++ {
		move, err := history.Move(version)
		if err != nil {
			return nil, err
		}
		if stats := d.result.Moves[move.Name]; stats != nil {
			stats.Applied++
		}
	}
	return d.result, nil
}

func newRecordStorage() *filesystem.StorageManager {
	storage := filesystem.NewStorageManager("")
	storage.DebugNoDisk = true
	return storage
}

// newManager builds a manager on the fixed manual clock with logging off.
func newManager(newDelegate conformance.NewDelegate, storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	manager, err := boardgame.NewGameManager(newDelegate(), storage)
	if err != nil {
		return nil, fmt.Errorf("create manager: %w", err)
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	if err := manager.Internals().UseManualClock(conformance.Clock); err != nil {
		manager.Internals().Close()
		return nil, err
	}
	return manager, nil
}

// seatMoveName is the game's seat-player move, or "".
func seatMoveName(manager *boardgame.GameManager) string {
	for _, move := range manager.ExampleMoves() {
		if seater, ok := move.(interfaces.SeatPlayerMover); ok && seater.IsSeatPlayerMove() {
			return move.Info().Name()
		}
	}
	return ""
}

type driver struct {
	cfg      Config
	manager  *boardgame.GameManager
	game     *boardgame.Game
	storage  *driverStorage
	derived  *Derived
	random   *rand.Rand
	schema   []boardgame.MoveInputSchemaMove
	seatMove string
	result   *Result
	policy   policy
	seatedAt int // version of the first seat; -1 before it
}

func (d *driver) loop() error {
	d.seatedAt = -1
	uncovered := map[string]bool{}
	for {
		if d.game.Finished() {
			d.result.End = EndFinished
			return nil
		}
		// A seated run makes its first seat before any bound applies: the
		// exporter learns that the run is seated from a seat move in the
		// history, so a seated history must contain one.
		firstSeat := d.derived.Seated && d.seatedAt < 0
		if !firstSeat {
			if d.result.Proposals >= d.cfg.MaxProposals {
				d.result.End = EndMaxProposals
				return nil
			}
			if d.game.Version() >= d.cfg.MaxVersions {
				d.result.End = EndMaxVersions
				return nil
			}
		}
		if !d.game.AtProposalFrontier() {
			return fmt.Errorf("the game is not at a settled version")
		}
		step, err := d.enumerate(uncovered)
		if err != nil {
			return err
		}
		// One draw per step, always, so the generator's sequence does not
		// depend on which kinds of refusal happen to be available.
		refuse := d.random.Float64() < d.cfg.RefusalRate
		if firstSeat {
			if step.seat == nil {
				return fmt.Errorf("seated run: no seat can be filled at version %d", d.game.Version())
			}
			if err := d.execute(*step.seat); err != nil {
				return err
			}
			continue
		}
		if refuse {
			sent, err := d.refuse(step)
			if err != nil {
				return err
			}
			if sent {
				continue
			}
		}
		if len(step.actions) == 0 {
			d.result.End = EndNoAction
			return nil
		}
		chosen, err := d.choose(step)
		if err != nil {
			return err
		}
		if err := d.execute(chosen); err != nil {
			return err
		}
	}
}

type action struct {
	kind  string // "propose", "timer" or "seat"
	move  string
	input map[string]interface{}
	actor boardgame.PlayerIndex
	seat  boardgame.PlayerIndex
	// bound is a move to propose as is (the agent's), instead of binding
	// input.
	bound boardgame.Move
}

// snapshot is what must not change when the engine refuses an input.
type snapshot struct {
	version  int
	finished bool
	state    []byte
}

func (d *driver) snapshot() (snapshot, error) {
	version := d.game.Version()
	state, err := d.storage.State(d.game.ID(), version)
	if err != nil {
		return snapshot{}, err
	}
	return snapshot{version: version, finished: d.game.Finished(), state: append([]byte(nil), state...)}, nil
}

func (d *driver) unchanged(before snapshot) error {
	after, err := d.snapshot()
	if err != nil {
		return err
	}
	if after.version != before.version || after.finished != before.finished || !bytes.Equal(after.state, before.state) {
		return fmt.Errorf("the engine refused an input but the game changed (version %d -> %d)", before.version, after.version)
	}
	if !d.game.AtProposalFrontier() {
		return fmt.Errorf("the engine refused an input and left the game unsettled")
	}
	return nil
}

func (d *driver) execute(a action) error {
	d.result.Proposals++
	before, err := d.snapshot()
	if err != nil {
		return err
	}
	switch a.kind {
	case "propose":
		move := a.bound
		if move == nil {
			if move, err = d.game.MoveWithInput(a.move, a.input); err != nil {
				return fmt.Errorf("rebind %s: %w", a.move, err)
			}
		}
		d.result.Moves[a.move].ProposedLegally++
		if err := <-d.game.ProposeMoveAtVersion(move, a.actor, before.version); err != nil {
			if checkErr := d.unchanged(before); checkErr != nil {
				return checkErr
			}
			input := a.input
			if a.bound != nil {
				input = boundInput(a.bound)
			}
			d.record(ReasonApplyError, "propose", a.move, input, a.actor, before.version, err)
			return nil
		}
	case "timer":
		fired, err := d.manager.Internals().ForceNextTimerWithError()
		if err != nil {
			if checkErr := d.unchanged(before); checkErr != nil {
				return checkErr
			}
			d.record(ReasonTimerError, "timer", "", nil, boardgame.AdminPlayerIndex, before.version, err)
			return nil
		}
		if !fired {
			return fmt.Errorf("a timer was active but none fired")
		}
	case "seat":
		d.storage.pendingSeat = &seat{index: a.seat, storage: d.storage}
		err := <-d.manager.Internals().ForceFixUp(d.game)
		accepted := d.storage.pendingSeat == nil
		d.storage.pendingSeat = nil
		if err != nil {
			return fmt.Errorf("seat %d: %w", a.seat, err)
		}
		if !accepted {
			return fmt.Errorf("the game did not accept seat %d, which Legal allowed", a.seat)
		}
		if d.seatedAt < 0 {
			d.seatedAt = d.game.Version()
		}
	default:
		return fmt.Errorf("unknown action kind %q", a.kind)
	}
	if d.game.Version() == before.version {
		return fmt.Errorf("%s %q was accepted but applied nothing", a.kind, a.move)
	}
	if !d.game.AtProposalFrontier() {
		return fmt.Errorf("%s %q left the game unsettled", a.kind, a.move)
	}
	return nil
}

// refuse sends one input the driver expects the engine to refuse. It
// returns false, sending nothing, when no kind of refusal is available.
func (d *driver) refuse(step *step) (bool, error) {
	var illegal, wrongActor, anyLegal []*candidate
	for _, c := range step.candidates {
		switch {
		case len(c.legal) == 0 && len(c.illegal) > 0:
			illegal = append(illegal, c)
		case len(c.legal) > 0 && len(c.illegal) > 0:
			wrongActor = append(wrongActor, c)
		}
		if len(c.legal) > 0 {
			anyLegal = append(anyLegal, c)
		}
	}
	var kinds []string
	if len(illegal) > 0 {
		kinds = append(kinds, ReasonIllegalInput)
	}
	if len(wrongActor) > 0 {
		kinds = append(kinds, ReasonWrongActor)
	}
	if len(anyLegal) > 0 && d.game.Version() > 0 {
		kinds = append(kinds, ReasonStaleVersion)
	}
	if len(kinds) == 0 {
		return false, nil
	}
	kind := kinds[d.random.Intn(len(kinds))]
	var chosen *candidate
	var actor boardgame.PlayerIndex
	expected := d.game.Version()
	switch kind {
	case ReasonIllegalInput:
		chosen = illegal[d.random.Intn(len(illegal))]
		actor = chosen.illegal[d.random.Intn(len(chosen.illegal))]
	case ReasonWrongActor:
		chosen = wrongActor[d.random.Intn(len(wrongActor))]
		actor = chosen.illegal[d.random.Intn(len(chosen.illegal))]
	case ReasonStaleVersion:
		chosen = anyLegal[d.random.Intn(len(anyLegal))]
		actor = chosen.legal[d.random.Intn(len(chosen.legal))]
		expected--
	}
	d.result.Proposals++
	before, err := d.snapshot()
	if err != nil {
		return false, err
	}
	move, err := d.game.MoveWithInput(chosen.move, chosen.input)
	if err != nil {
		return false, fmt.Errorf("rebind %s: %w", chosen.move, err)
	}
	proposeErr := <-d.game.ProposeMoveAtVersion(move, actor, expected)
	if proposeErr == nil {
		return false, fmt.Errorf("the engine accepted %s %v by %d (%s), which the driver expected it to refuse", chosen.move, chosen.input, actor, kind)
	}
	if err := d.unchanged(before); err != nil {
		return false, err
	}
	d.record(kind, "propose", chosen.move, chosen.input, actor, expected, proposeErr)
	return true, nil
}

func (d *driver) record(reason, kind, name string, input map[string]interface{}, proposer boardgame.PlayerIndex, expected int, err error) {
	encoded := map[string]json.RawMessage{}
	for field, value := range input {
		blob, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			blob = []byte("null")
		}
		encoded[field] = blob
	}
	d.result.Refusals = append(d.result.Refusals, Refusal{
		Rejection: conformance.Rejection{
			AtVersion: d.game.Version(), Kind: kind, Name: name, Input: encoded,
			Proposer: proposer, ExpectedVersion: expected, Error: err.Error(),
		},
		Reason: reason,
	})
	if stats := d.result.Moves[name]; stats != nil {
		stats.Refused++
	}
}

// boundInput reads the supplied-field values off a move built elsewhere.
func boundInput(move boardgame.Move) map[string]interface{} {
	fields, err := boardgame.ResolveMoveInputFields(move)
	if err != nil {
		return map[string]interface{}{}
	}
	input := map[string]interface{}{}
	for _, field := range fields {
		if field.Disposition != boardgame.MoveInputRequired && field.Disposition != boardgame.MoveInputServerDefaulted {
			continue
		}
		value, err := move.ReadSetter().Prop(field.Name)
		if err != nil {
			continue
		}
		input[field.Name] = value
	}
	return input
}

// driverStorage is in-memory record storage plus the seating rendezvous, as
// scenario.Run and the conformance replay use it.
type driverStorage struct {
	*filesystem.StorageManager
	seating     bool
	pendingSeat *seat
}

func (s *driverStorage) FetchInjectedDataForGame(gameID, dataType string) interface{} {
	if dataType == interfaces.WillSeatPlayerRendezvousDataType {
		return s.seating
	}
	if dataType == interfaces.PlayerToSeatRendezvousDataType {
		if s.pendingSeat == nil {
			return nil
		}
		return s.pendingSeat
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}

type seat struct {
	index   boardgame.PlayerIndex
	storage *driverStorage
}

func (s *seat) SeatIndex() boardgame.PlayerIndex { return s.index }
func (s *seat) Committed()                       { s.storage.pendingSeat = nil }
