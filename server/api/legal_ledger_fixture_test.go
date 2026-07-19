package api

import (
	"errors"
	"sync"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

/*
This file is Task 10's server-ledger test fixture: a minimal, hand-built
opted-in-move game (legal_ledger_test.go exercises it). It cannot reuse an
existing example game because none has opted in to declarative legality yet
(Tasks 11/12 migrate examples/memory and examples/checkers) -- the only
un-migrated fixture Task 10 needs is examples/memory itself, used
separately by the frozen-wire test (legal_ledger_frozen_wire_test.go) for
the OPAQUE-move byte-identity guarantee.

Reader()/ReadSetter()/ReadSetConfigurer() below are hand-written rather than
`boardgame:codegen`-generated: this package is external to package
boardgame, so it cannot use the package-internal reflection reader
(property_reader.go's getDefaultReader, "only used for testing within the
package"), and running the real `boardgame-util codegen` tool against a
throwaway fixture risked a large, unrelated diff (confirmed by trial: it
disagreed with the moves package's own long-committed codegen output).
Hand-rolling follows the exact same generated-code shape codegen itself
produces (compare base/auto_reader.go's Move reader for the empty-fields
case), just reduced to only the fields this fixture actually has.
*/

// legalLedgerGameState is the fixture's game state: CurrentPlayer is read by
// base.GameDelegate's default CurrentPlayerIndex (by convention, the
// "CurrentPlayer" PlayerIndex prop) and by legal.ProposerIsCurrentPlayer's
// own declared "game.CurrentPlayer" Read; HiddenCounter is a plain int this
// fixture's delegate sanitizes to Hidden for every non-admin viewer, giving
// an authored precondition something to fail on and the ledger something to
// strip bindings from (design spec §6, #693 guard).
type legalLedgerGameState struct {
	base.SubState
	CurrentPlayer boardgame.PlayerIndex
	HiddenCounter int
}

var legalLedgerGameStateProps = map[string]boardgame.PropertyType{
	"CurrentPlayer": boardgame.TypePlayerIndex,
	"HiddenCounter": boardgame.TypeInt,
}

type legalLedgerGameStateReader struct {
	data *legalLedgerGameState
}

func (r *legalLedgerGameStateReader) Props() map[string]boardgame.PropertyType {
	return legalLedgerGameStateProps
}

func (r *legalLedgerGameStateReader) Prop(name string) (interface{}, error) {
	switch name {
	case "CurrentPlayer":
		return r.PlayerIndexProp(name)
	case "HiddenCounter":
		return r.IntProp(name)
	}
	return nil, errors.New("no such property: " + name)
}

func (r *legalLedgerGameStateReader) PropMutable(name string) bool {
	switch name {
	case "CurrentPlayer", "HiddenCounter":
		return true
	}
	return false
}

func (r *legalLedgerGameStateReader) SetProp(name string, value interface{}) error {
	return r.ConfigureProp(name, value)
}

func (r *legalLedgerGameStateReader) ConfigureProp(name string, value interface{}) error {
	switch name {
	case "CurrentPlayer":
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("provided value was not a PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case "HiddenCounter":
		val, ok := value.(int)
		if !ok {
			return errors.New("provided value was not an int")
		}
		return r.SetIntProp(name, val)
	}
	return errors.New("no such property: " + name)
}

func (r *legalLedgerGameStateReader) IntProp(name string) (int, error) {
	if name == "HiddenCounter" {
		return r.data.HiddenCounter, nil
	}
	return 0, errors.New("no such int prop: " + name)
}

func (r *legalLedgerGameStateReader) SetIntProp(name string, value int) error {
	if name == "HiddenCounter" {
		r.data.HiddenCounter = value
		return nil
	}
	return errors.New("no such int prop: " + name)
}

func (r *legalLedgerGameStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {
	if name == "CurrentPlayer" {
		return r.data.CurrentPlayer, nil
	}
	return 0, errors.New("no such PlayerIndex prop: " + name)
}

func (r *legalLedgerGameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {
	if name == "CurrentPlayer" {
		r.data.CurrentPlayer = value
		return nil
	}
	return errors.New("no such PlayerIndex prop: " + name)
}

func (r *legalLedgerGameStateReader) BoolProp(name string) (bool, error) {
	return false, errors.New("no such bool prop: " + name)
}
func (r *legalLedgerGameStateReader) SetBoolProp(name string, value bool) error {
	return errors.New("no such bool prop: " + name)
}
func (r *legalLedgerGameStateReader) StringProp(name string) (string, error) {
	return "", errors.New("no such string prop: " + name)
}
func (r *legalLedgerGameStateReader) SetStringProp(name string, value string) error {
	return errors.New("no such string prop: " + name)
}
func (r *legalLedgerGameStateReader) IntSliceProp(name string) ([]int, error) {
	return nil, errors.New("no such []int prop: " + name)
}
func (r *legalLedgerGameStateReader) SetIntSliceProp(name string, value []int) error {
	return errors.New("no such []int prop: " + name)
}
func (r *legalLedgerGameStateReader) BoolSliceProp(name string) ([]bool, error) {
	return nil, errors.New("no such []bool prop: " + name)
}
func (r *legalLedgerGameStateReader) SetBoolSliceProp(name string, value []bool) error {
	return errors.New("no such []bool prop: " + name)
}
func (r *legalLedgerGameStateReader) StringSliceProp(name string) ([]string, error) {
	return nil, errors.New("no such []string prop: " + name)
}
func (r *legalLedgerGameStateReader) SetStringSliceProp(name string, value []string) error {
	return errors.New("no such []string prop: " + name)
}
func (r *legalLedgerGameStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {
	return nil, errors.New("no such []PlayerIndex prop: " + name)
}
func (r *legalLedgerGameStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {
	return errors.New("no such []PlayerIndex prop: " + name)
}
func (r *legalLedgerGameStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {
	return nil, errors.New("no such enum prop: " + name)
}
func (r *legalLedgerGameStateReader) EnumProp(name string) (enum.Val, error) {
	return nil, errors.New("no such enum prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureEnumProp(name string, value enum.Val) error {
	return errors.New("no such enum prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {
	return errors.New("no such enum prop: " + name)
}
func (r *legalLedgerGameStateReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {
	return nil, errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerGameStateReader) EnumSliceProp(name string) (enum.EnumSlice, error) {
	return nil, errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {
	return errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {
	return errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerGameStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {
	return nil, errors.New("no such stack prop: " + name)
}
func (r *legalLedgerGameStateReader) StackProp(name string) (boardgame.Stack, error) {
	return nil, errors.New("no such stack prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {
	return errors.New("no such stack prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {
	return errors.New("no such stack prop: " + name)
}
func (r *legalLedgerGameStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {
	return nil, errors.New("no such board prop: " + name)
}
func (r *legalLedgerGameStateReader) BoardProp(name string) (boardgame.Board, error) {
	return nil, errors.New("no such board prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {
	return errors.New("no such board prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {
	return errors.New("no such board prop: " + name)
}
func (r *legalLedgerGameStateReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {
	return nil, errors.New("no such timer prop: " + name)
}
func (r *legalLedgerGameStateReader) TimerProp(name string) (boardgame.Timer, error) {
	return nil, errors.New("no such timer prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {
	return errors.New("no such timer prop: " + name)
}
func (r *legalLedgerGameStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {
	return errors.New("no such timer prop: " + name)
}

func (g *legalLedgerGameState) Reader() boardgame.PropertyReader {
	return &legalLedgerGameStateReader{g}
}
func (g *legalLedgerGameState) ReadSetter() boardgame.PropertyReadSetter {
	return &legalLedgerGameStateReader{g}
}
func (g *legalLedgerGameState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &legalLedgerGameStateReader{g}
}

// legalLedgerPlayerState is the fixture's player state: no persistable
// fields of its own, so its reader is the same empty-Props() shape
// base.Move's own codegen'd reader uses (base/auto_reader.go).
type legalLedgerPlayerState struct {
	base.SubState
}

var legalLedgerPlayerStateProps = map[string]boardgame.PropertyType{}

type legalLedgerPlayerStateReader struct {
	data *legalLedgerPlayerState
}

func (r *legalLedgerPlayerStateReader) Props() map[string]boardgame.PropertyType {
	return legalLedgerPlayerStateProps
}
func (r *legalLedgerPlayerStateReader) Prop(name string) (interface{}, error) {
	return nil, errors.New("no such property: " + name)
}
func (r *legalLedgerPlayerStateReader) PropMutable(name string) bool { return false }
func (r *legalLedgerPlayerStateReader) SetProp(name string, value interface{}) error {
	return errors.New("no such property: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureProp(name string, value interface{}) error {
	return errors.New("no such property: " + name)
}
func (r *legalLedgerPlayerStateReader) IntProp(name string) (int, error) {
	return 0, errors.New("no such int prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetIntProp(name string, value int) error {
	return errors.New("no such int prop: " + name)
}
func (r *legalLedgerPlayerStateReader) BoolProp(name string) (bool, error) {
	return false, errors.New("no such bool prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetBoolProp(name string, value bool) error {
	return errors.New("no such bool prop: " + name)
}
func (r *legalLedgerPlayerStateReader) StringProp(name string) (string, error) {
	return "", errors.New("no such string prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetStringProp(name string, value string) error {
	return errors.New("no such string prop: " + name)
}
func (r *legalLedgerPlayerStateReader) IntSliceProp(name string) ([]int, error) {
	return nil, errors.New("no such []int prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetIntSliceProp(name string, value []int) error {
	return errors.New("no such []int prop: " + name)
}
func (r *legalLedgerPlayerStateReader) BoolSliceProp(name string) ([]bool, error) {
	return nil, errors.New("no such []bool prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetBoolSliceProp(name string, value []bool) error {
	return errors.New("no such []bool prop: " + name)
}
func (r *legalLedgerPlayerStateReader) StringSliceProp(name string) ([]string, error) {
	return nil, errors.New("no such []string prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetStringSliceProp(name string, value []string) error {
	return errors.New("no such []string prop: " + name)
}
func (r *legalLedgerPlayerStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {
	return 0, errors.New("no such PlayerIndex prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {
	return errors.New("no such PlayerIndex prop: " + name)
}
func (r *legalLedgerPlayerStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {
	return nil, errors.New("no such []PlayerIndex prop: " + name)
}
func (r *legalLedgerPlayerStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {
	return errors.New("no such []PlayerIndex prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {
	return nil, errors.New("no such enum prop: " + name)
}
func (r *legalLedgerPlayerStateReader) EnumProp(name string) (enum.Val, error) {
	return nil, errors.New("no such enum prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureEnumProp(name string, value enum.Val) error {
	return errors.New("no such enum prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {
	return errors.New("no such enum prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {
	return nil, errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerPlayerStateReader) EnumSliceProp(name string) (enum.EnumSlice, error) {
	return nil, errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {
	return errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {
	return errors.New("no such enumslice prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {
	return nil, errors.New("no such stack prop: " + name)
}
func (r *legalLedgerPlayerStateReader) StackProp(name string) (boardgame.Stack, error) {
	return nil, errors.New("no such stack prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {
	return errors.New("no such stack prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {
	return errors.New("no such stack prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {
	return nil, errors.New("no such board prop: " + name)
}
func (r *legalLedgerPlayerStateReader) BoardProp(name string) (boardgame.Board, error) {
	return nil, errors.New("no such board prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {
	return errors.New("no such board prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {
	return errors.New("no such board prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {
	return nil, errors.New("no such timer prop: " + name)
}
func (r *legalLedgerPlayerStateReader) TimerProp(name string) (boardgame.Timer, error) {
	return nil, errors.New("no such timer prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {
	return errors.New("no such timer prop: " + name)
}
func (r *legalLedgerPlayerStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {
	return errors.New("no such timer prop: " + name)
}

func (p *legalLedgerPlayerState) Reader() boardgame.PropertyReader {
	return &legalLedgerPlayerStateReader{p}
}
func (p *legalLedgerPlayerState) ReadSetter() boardgame.PropertyReadSetter {
	return &legalLedgerPlayerStateReader{p}
}
func (p *legalLedgerPlayerState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &legalLedgerPlayerStateReader{p}
}

// legalLedgerDelegate is the fixture's GameDelegate: package name must be
// "api" (game_manager.go's gamePkgMatchesDelegateName requires the
// delegate's Name() to equal its owning package's last path component).
type legalLedgerDelegate struct {
	base.GameDelegate
}

func (d *legalLedgerDelegate) Name() string { return "api" }

func (d *legalLedgerDelegate) DefaultNumPlayers() int { return 2 }

func (d *legalLedgerDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(legalLedgerGameState)
}

func (d *legalLedgerDelegate) PlayerStateConstructor(boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(legalLedgerPlayerState)
}

// SanitizationPolicy hides HiddenCounter from every non-admin viewer
// (PolicyHidden -- LegalFacetValues never survives it), giving the fixture
// move's authored precondition something concrete to be inevaluable/
// bindings-stripped about.
func (d *legalLedgerDelegate) SanitizationPolicy(prop boardgame.StatePropertyRef, groupMembership map[string]bool) boardgame.Policy {
	if prop.Group == boardgame.StateGroupGame && prop.PropName == "HiddenCounter" {
		return boardgame.PolicyHidden
	}
	return boardgame.PolicyVisible
}

func (d *legalLedgerDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(d)
	return moves.Add(
		auto.MustConfig(
			new(legalLedgerMoveOptedIn),
			moves.WithMoveName("Opted In"),
			moves.WithLegalPreconditions(
				// game.HiddenCounter starts at 0, so this always fails --
				// exactly what the bindings-stripping / evaluable tests
				// need (a real Fail carrying real Bindings to strip).
				legal.PropAtLeast("game.HiddenCounter", 1000),
			),
		),
		auto.MustConfig(
			new(legalLedgerMoveOpaque),
			moves.WithMoveName("Opaque"),
		),
	)
}

// legalLedgerMoveOptedIn embeds moves.CurrentPlayer with NO extra
// persistable fields, so it inherits CurrentPlayer's own already-generated
// Reader() (moves/auto_reader.go, checked in) via method promotion --
// giving this fixture the "move.TargetPlayerIndex" field
// legal.ProposerIsCurrentPlayer needs for free, without any codegen of its
// own.
type legalLedgerMoveOptedIn struct {
	moves.CurrentPlayer
}

func (m *legalLedgerMoveOptedIn) Apply(state boardgame.State) error { return nil }

// legalLedgerMoveOpaque never calls WithLegalPreconditions -- an opaque
// (non-opted-in) move type sharing this fixture's manager, so a single test
// game can exercise both the ledger path and the frozen two-call path.
type legalLedgerMoveOpaque struct {
	moves.Default
}

func (m *legalLedgerMoveOpaque) Apply(state boardgame.State) error { return nil }

// legalLedgerStorage is a minimal hand-rolled boardgame.StorageManager.
// storage/memory (the repo's real in-memory implementation) can't be used
// here: it pulls in storage/internal/helpers, which imports package
// server/api itself -- an import cycle from THIS package's own test file.
// This fixture only needs enough persistence for NewDefaultGame/ProposeMove
// to round-trip within a single test process, so it skips
// storage/internal/helpers' shared plumbing (chat, extendedgame listing,
// AllGames) entirely and implements boardgame.StorageManager's nine methods
// directly against plain maps.
type legalLedgerStorage struct {
	mu     sync.Mutex
	states map[string]map[int]boardgame.StateStorageRecord
	moves  map[string]map[int]*boardgame.MoveStorageRecord
	games  map[string]*boardgame.GameStorageRecord
}

func newLegalLedgerStorage() *legalLedgerStorage {
	return &legalLedgerStorage{
		states: make(map[string]map[int]boardgame.StateStorageRecord),
		moves:  make(map[string]map[int]*boardgame.MoveStorageRecord),
		games:  make(map[string]*boardgame.GameStorageRecord),
	}
}

func (s *legalLedgerStorage) State(gameID string, version int) (boardgame.StateStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	versions, ok := s.states[gameID]
	if !ok {
		return nil, errors.New("no such game")
	}
	record, ok := versions[version]
	if !ok {
		return nil, errors.New("no such version for that game")
	}
	return record, nil
}

func (s *legalLedgerStorage) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	versions, ok := s.moves[gameID]
	if !ok {
		return nil, errors.New("no such game")
	}
	record, ok := versions[version]
	if !ok {
		return nil, errors.New("no such version for that game")
	}
	return record, nil
}

func (s *legalLedgerStorage) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	var result []*boardgame.MoveStorageRecord
	for v := fromVersion + 1; v <= toVersion; v++ {
		move, err := s.Move(gameID, v)
		if err != nil {
			return nil, err
		}
		result = append(result, move)
	}
	return result, nil
}

func (s *legalLedgerStorage) Game(id string) (*boardgame.GameStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.games[id]
	if !ok {
		return nil, errors.New("no such game")
	}
	return record, nil
}

func (s *legalLedgerStorage) AgentState(gameID string, player boardgame.PlayerIndex) ([]byte, error) {
	return nil, nil
}

func (s *legalLedgerStorage) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	if game == nil {
		return errors.New("no game provided")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.states[game.ID] == nil {
		s.states[game.ID] = make(map[int]boardgame.StateStorageRecord)
	}
	if s.moves[game.ID] == nil {
		s.moves[game.ID] = make(map[int]*boardgame.MoveStorageRecord)
	}

	s.states[game.ID][game.Version] = state
	if move != nil {
		s.moves[game.ID][game.Version] = move
	}
	s.games[game.ID] = game
	return nil
}

func (s *legalLedgerStorage) SaveProposalFrontier(gameID string, stateVersion, frontierVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	game := s.games[gameID]
	if game == nil || game.Version != stateVersion {
		return errors.New("proposal frontier used a stale or missing game version")
	}
	game.ProposalFrontierKnown = frontierVersion >= 0
	game.ProposalFrontierVersion = frontierVersion
	return nil
}

func (s *legalLedgerStorage) SaveAgentState(gameID string, player boardgame.PlayerIndex, state []byte) error {
	return nil
}

func (s *legalLedgerStorage) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	return nil
}

func (s *legalLedgerStorage) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	return nil
}

// newLegalLedgerGame builds a fresh two-player game on legalLedgerDelegate.
func newLegalLedgerGame(t interface {
	Fatalf(format string, args ...interface{})
}) (*boardgame.Game, *boardgame.GameManager) {
	manager, err := boardgame.NewGameManager(&legalLedgerDelegate{}, newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("legal ledger fixture: building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal ledger fixture: building game: %v", err)
	}
	return game, manager
}

// legalLedgerObserverDelegate is legalLedgerDelegate with FinishSetUp
// overridden to force the fixture into the "no current player" state the
// LegalForAnyone regression (docs/superpowers/... critical review finding)
// is about: game.CurrentPlayer = boardgame.ObserverPlayerIndex (a documented
// framework pattern -- base.GameDelegate.CurrentPlayerIndex's own doc
// comment: "If your game has different rounds where no one may move, return
// boardgame.ObserverPlayerIndex"), and game.HiddenCounter set high enough
// that the authored propAtLeast precondition PASSES -- isolating the
// contributed proposerIsCurrentPlayer atom as the ONLY thing that can fail,
// so a test can tell whether LegalForAnyone incorrectly exempts it.
type legalLedgerObserverDelegate struct {
	legalLedgerDelegate
}

func (d *legalLedgerObserverDelegate) FinishSetUp(state boardgame.State) error {
	if err := d.legalLedgerDelegate.FinishSetUp(state); err != nil {
		return err
	}
	gameState, ok := state.GameState().(*legalLedgerGameState)
	if !ok {
		return errors.New("legal ledger observer fixture: game state was not *legalLedgerGameState")
	}
	gameState.CurrentPlayer = boardgame.ObserverPlayerIndex
	gameState.HiddenCounter = 1000
	return nil
}

// newLegalLedgerObserverGame builds a fresh two-player game on
// legalLedgerObserverDelegate: same "Opted In"/"Opaque" moves as
// newLegalLedgerGame, but game.CurrentPlayer is forced to
// boardgame.ObserverPlayerIndex and game.HiddenCounter is pre-seeded past
// propAtLeast's threshold, so moves.CurrentPlayer's target-player checks are
// the only thing that can make "Opted In" illegal.
func newLegalLedgerObserverGame(t interface {
	Fatalf(format string, args ...interface{})
}) (*boardgame.Game, *boardgame.GameManager) {
	manager, err := boardgame.NewGameManager(&legalLedgerObserverDelegate{}, newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("legal ledger observer fixture: building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal ledger observer fixture: building game: %v", err)
	}
	return game, manager
}
