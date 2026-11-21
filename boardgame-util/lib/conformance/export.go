package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/filesystem"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/sirupsen/logrus"
)

// NewDelegate returns a fresh delegate not yet attached to a manager. An
// export may need more than one.
type NewDelegate func() boardgame.GameDelegate

// FromScenario runs spec through scenario.Run and exports the resulting
// history. spec.Viewers is ignored: every viewer is always exported.
func FromScenario(newDelegate NewDelegate, spec scenario.Spec, covers string) (*Document, error) {
	result, err := scenario.Run(newDelegate(), spec)
	if err != nil {
		return nil, fmt.Errorf("scenario %q: %w", spec.Name, err)
	}
	if result.History == nil {
		return nil, fmt.Errorf("scenario %q produced no history", spec.Name)
	}
	return FromHistory(newDelegate, result.History, spec.Name, Source{Kind: "scenario", Description: spec.Name, Covers: covers})
}

// FromGolden exports the golden history at path. displayPath is recorded in
// the Document in place of path, which may be machine-specific.
func FromGolden(newDelegate NewDelegate, path, displayPath, name, covers string) (*Document, error) {
	history, err := record.New(path)
	if err != nil {
		return nil, fmt.Errorf("golden %s: %w", displayPath, err)
	}
	return FromHistory(newDelegate, history, name, Source{Kind: "golden", Path: displayPath, Description: history.Description(), Covers: covers})
}

// FromHistory replays history on the real engine and exports every version.
// It fails if the engine does not reproduce the history.
func FromHistory(newDelegate NewDelegate, history *record.Record, name string, source Source) (*Document, error) {
	if newDelegate == nil || history == nil || history.Game() == nil {
		return nil, fmt.Errorf("conformance export requires a delegate and a history")
	}
	want := history.Game()
	for _, agent := range want.Agents {
		if agent != "" {
			return nil, fmt.Errorf("history uses agent %q; agent decisions are not replayable", agent)
		}
	}

	storage := &replayStorage{StorageManager: filesystem.NewStorageManager("")}
	storage.DebugNoDisk = true
	delegate := newDelegate()
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		return nil, fmt.Errorf("create manager: %w", err)
	}
	defer manager.Internals().Close()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	if err := manager.Internals().UseManualClock(Clock); err != nil {
		return nil, err
	}

	for version := 1; version <= want.Version; version++ {
		rec, err := history.Move(version)
		if err != nil {
			return nil, fmt.Errorf("history move %d: %w", version, err)
		}
		if isSeatMove(manager.ExampleMoveByName(rec.Name)) {
			storage.seating = true
		}
	}

	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: want.Name, ID: want.ID, SecretSalt: want.SecretSalt,
		NumPlayers: want.NumPlayers, Variant: want.Variant, Agents: want.Agents,
	})
	if err != nil {
		return nil, fmt.Errorf("recreate game: %w", err)
	}

	e := &exporter{manager: manager, game: game, storage: storage, history: history}
	doc, err := e.header(name, source)
	if err != nil {
		return nil, err
	}
	e.doc = doc

	// Version 0 and whatever chain the engine ran on its own during setup.
	if err := e.capture("setup"); err != nil {
		return nil, err
	}
	for game.Version() < want.Version {
		if game.Finished() {
			return nil, fmt.Errorf("game finished at version %d but the history continues to %d", game.Version(), want.Version)
		}
		before := game.Version()
		next, err := history.Move(before + 1)
		if err != nil {
			return nil, fmt.Errorf("history move %d: %w", before+1, err)
		}
		trigger, err := e.drive(next)
		if err != nil {
			return nil, fmt.Errorf("version %d (%s): %w", before+1, next.Name, err)
		}
		if game.Version() == before {
			return nil, fmt.Errorf("version %d (%s): the engine applied nothing", before+1, next.Name)
		}
		if err := e.capture(trigger); err != nil {
			return nil, err
		}
	}
	if game.Version() != want.Version {
		return nil, fmt.Errorf("replay ended at version %d; the history ends at %d", game.Version(), want.Version)
	}
	doc.Game.Finished = game.Finished()
	doc.Game.Winners = winners(game.Winners())
	doc.Game.FinalVersion = game.Version()
	return doc, nil
}

type exporter struct {
	manager  *boardgame.GameManager
	game     *boardgame.Game
	storage  *replayStorage
	history  *record.Record
	doc      *Document
	captured int // number of versions captured so far
}

// drive repeats the outside input that produced history move next.
func (e *exporter) drive(next *boardgame.MoveStorageRecord) (string, error) {
	if next.Proposer >= 0 {
		move, err := e.manager.Internals().InflateMoveStorageRecord(next, e.game)
		if err != nil {
			return "", fmt.Errorf("inflate: %w", err)
		}
		return "propose", <-e.game.ProposeMoveAtVersion(move, next.Proposer, e.game.Version())
	}
	example := e.game.MoveByName(next.Name)
	if example == nil {
		return "", fmt.Errorf("move is not configured")
	}
	if isSeatMove(example) {
		move, err := e.manager.Internals().InflateMoveStorageRecord(next, e.game)
		if err != nil {
			return "", fmt.Errorf("inflate: %w", err)
		}
		index, err := move.Reader().PlayerIndexProp("TargetPlayerIndex")
		if err != nil {
			return "", fmt.Errorf("seat target: %w", err)
		}
		e.storage.pendingSeat = &seat{index: index, storage: e.storage}
		if err := <-e.manager.Internals().ForceFixUp(e.game); err != nil {
			return "", err
		}
		if e.storage.pendingSeat != nil {
			return "", fmt.Errorf("the game did not accept seat %d", index)
		}
		return "seat", nil
	}
	fired, err := e.manager.Internals().ForceNextTimerWithError()
	if err != nil {
		return "", err
	}
	if !fired {
		return "", fmt.Errorf("the history applies it as admin, but it is not a seat move and no timer is pending")
	}
	return "timer", nil
}

func isSeatMove(move boardgame.Move) bool {
	seater, ok := move.(interfaces.SeatPlayerMover)
	return ok && seater.IsSeatPlayerMove()
}

// capture records every version not yet captured, up to the settled head.
func (e *exporter) capture(trigger string) error {
	if !e.game.AtProposalFrontier() {
		return fmt.Errorf("version %d is not a settled boundary", e.game.Version())
	}
	head := e.game.Version()
	first := e.captured
	for version := first; version <= head; version++ {
		entry, err := e.version(version, head, first, trigger)
		if err != nil {
			return fmt.Errorf("capture version %d: %w", version, err)
		}
		e.doc.Versions = append(e.doc.Versions, *entry)
		e.captured++
	}
	return nil
}

func (e *exporter) version(version, head, first int, trigger string) (*Version, error) {
	stored, err := e.storage.State(e.game.ID(), version)
	if err != nil {
		return nil, err
	}
	if expected, err := e.history.State(version); err != nil {
		return nil, fmt.Errorf("history state: %w", err)
	} else if err := sameState(stored, expected); err != nil {
		return nil, fmt.Errorf("the engine no longer reproduces the history: %w", err)
	}
	entry := &Version{Version: version, Settled: version == head}
	if entry.State, err = canonical(stored); err != nil {
		return nil, err
	}
	if version == head {
		entry.Finished = e.game.Finished()
		entry.Winners = winners(e.game.Winners())
	}

	var moveRecord *boardgame.MoveStorageRecord
	if version > 0 {
		moveRecord, err = e.storage.Move(e.game.ID(), version)
		if err != nil {
			return nil, err
		}
		expected, err := e.history.Move(version)
		if err != nil {
			return nil, fmt.Errorf("history move: %w", err)
		}
		if moveRecord.Name != expected.Name || moveRecord.Proposer != expected.Proposer {
			return nil, fmt.Errorf("the engine applied %q by %d; the history has %q by %d", moveRecord.Name, moveRecord.Proposer, expected.Name, expected.Proposer)
		}
		moveTrigger := "fixup"
		if moveRecord.Initiator == moveRecord.Version {
			moveTrigger = trigger
			if version != first && trigger != "setup" {
				return nil, fmt.Errorf("move %q starts a causal chain inside another step", moveRecord.Name)
			}
		}
		arguments, err := canonical(moveRecord.Blob)
		if err != nil {
			return nil, err
		}
		entry.Move = &MoveRecord{
			Name: moveRecord.Name, Version: moveRecord.Version, Initiator: moveRecord.Initiator,
			Phase: int(moveRecord.Phase), PhaseName: e.phaseName(int(moveRecord.Phase)),
			Proposer: moveRecord.Proposer, Arguments: arguments,
			IsFixUp: base.IsFixUp(e.manager.ExampleMoveByName(moveRecord.Name)), Trigger: moveTrigger,
		}
	}

	state := e.game.State(version)
	if state == nil || state.Version() != version {
		return nil, fmt.Errorf("state is unavailable")
	}
	activeTimers, err := activeTimersFromState(stored)
	if err != nil {
		return nil, err
	}
	for _, viewer := range e.doc.Viewers {
		view, err := e.view(viewer, state, moveRecord, entry, activeTimers)
		if err != nil {
			return nil, fmt.Errorf("viewer %d: %w", viewer, err)
		}
		entry.Views = append(entry.Views, *view)
	}
	return entry, nil
}

func (e *exporter) view(viewer boardgame.PlayerIndex, state boardgame.ImmutableState, moveRecord *boardgame.MoveStorageRecord, entry *Version, activeTimers map[string]interface{}) (*View, error) {
	wire, err := e.game.JSONForPlayer(viewer, state)
	if err != nil {
		return nil, err
	}
	blob, err := json.Marshal(wire)
	if err != nil {
		return nil, err
	}
	var game map[string]interface{}
	if err := decode(blob, &game); err != nil {
		return nil, err
	}
	if entry.Settled {
		// Here the version is the head, so the scheduler's own payload must
		// agree with what we rebuild from the state for every other version.
		if !reflect.DeepEqual(game["ActiveTimers"], interface{}(activeTimers)) {
			return nil, fmt.Errorf("ActiveTimers on the wire %v differ from the state's timers %v", game["ActiveTimers"], activeTimers)
		}
	}
	delete(game, "Diagram")
	game["ActiveTimers"] = activeTimers
	game["Version"] = entry.Version
	game["Finished"] = entry.Finished
	game["Winners"] = entry.Winners
	view := &View{Viewer: viewer}
	if view.Game, err = json.Marshal(game); err != nil {
		return nil, err
	}

	view.Move = json.RawMessage("null")
	if moveRecord != nil {
		sanitized, err := e.game.MoveJSONForPlayer(viewer, moveRecord)
		if err != nil {
			return nil, fmt.Errorf("sanitize move: %w", err)
		}
		if sanitized != nil {
			raw, err := json.Marshal(sanitized)
			if err != nil {
				return nil, err
			}
			if view.Move, err = canonical(raw); err != nil {
				return nil, err
			}
		}
	}

	forms, err := api.MoveFormsForViewerAtState(e.game, state, viewer)
	if err != nil {
		return nil, err
	}
	if view.Legality, err = legalityFromForms(forms); err != nil {
		return nil, err
	}

	if entry.Settled && viewer >= 0 {
		choices, err := api.ProjectMoveChoicesForViewer(e.game, viewer)
		if err != nil {
			return nil, fmt.Errorf("projected choices: %w", err)
		}
		if choices != nil {
			if view.ProjectedMoveChoices, err = canonical(choices); err != nil {
				return nil, err
			}
		}
	}
	return view, nil
}

func (e *exporter) phaseName(phase int) string {
	phases := e.manager.Delegate().PhaseEnum()
	if phases == nil {
		return ""
	}
	return phases.String(enum.EnumKey(phase))
}

func (e *exporter) header(name string, source Source) (*Document, error) {
	manager, game, delegate := e.manager, e.game, e.manager.Delegate()
	doc := &Document{
		SchemaVersion: SchemaVersion, Name: name, Source: source, Normalization: Normalization(),
		Game: Game{
			Name: delegate.Name(), DisplayName: delegate.DisplayName(), ID: game.ID(),
			SecretSalt: game.StorageRecord().SecretSalt, NumPlayers: game.NumPlayers(),
			MinNumPlayers: delegate.MinNumPlayers(), MaxNumPlayers: delegate.MaxNumPlayers(),
			DefaultNumPlayers: delegate.DefaultNumPlayers(), Variant: game.Variant(), Agents: game.Agents(),
			Seating: e.storage.seating,
		},
		Engine: Engine{
			Clock: Clock.Format(time.RFC3339), LegalCatalogVersion: boardgame.LegalCatalogVersion,
			MoveChoiceProjectionSchemaVersion: boardgame.MoveChoiceProjectionSchemaVersion,
		},
		Viewers: []boardgame.PlayerIndex{boardgame.AdminPlayerIndex, boardgame.ObserverPlayerIndex},
	}
	if doc.Game.Variant == nil {
		doc.Game.Variant = boardgame.Variant{}
	}
	if doc.Game.Agents == nil {
		doc.Game.Agents = []string{}
	}
	for i := 0; i < game.NumPlayers(); i++ {
		doc.Viewers = append(doc.Viewers, boardgame.PlayerIndex(i))
	}
	if phases := delegate.PhaseEnum(); phases != nil {
		doc.Engine.PhaseEnum = phases.Name()
	}
	if groups := delegate.GroupEnum(); groups != nil {
		doc.Engine.GroupEnum = groups.Name()
	}

	var err error
	if doc.Engine.MoveInputSchemaFingerprint, err = boardgame.MoveInputSchemaFingerprint(manager); err != nil {
		return nil, err
	}
	if doc.Engine.MoveChoiceProjectionSchemaFingerprint, err = boardgame.MoveChoiceProjectionSchemaFingerprint(manager); err != nil {
		return nil, err
	}
	inputSchema, err := boardgame.BuildMoveInputSchema(manager)
	if err != nil {
		return nil, err
	}
	choiceSchema, err := boardgame.BuildMoveChoiceProjectionSchema(manager)
	if err != nil {
		return nil, err
	}
	for target, value := range map[*json.RawMessage]interface{}{
		&doc.MoveInputSchema: inputSchema, &doc.MoveChoiceProjectionSchema: choiceSchema,
		&doc.ComputedProperties: manager.ComputedPropertyDescriptors(),
		&doc.VariantConfig:      manager.Variants(), &doc.Chest: manager.Chest(),
	} {
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if *target, err = canonical(raw); err != nil {
			return nil, err
		}
	}

	for _, move := range manager.ExampleMoves() {
		info := MoveType{
			Name: move.Info().Name(), HelpText: move.HelpText(), IsFixUp: base.IsFixUp(move),
			NamePublic: move.Info().MoveNamePublic(), AnimationKeys: move.Info().AnimationKeys(),
			Fields: []MoveField{},
		}
		if starter, ok := move.(interfaces.GatheringStartMover); ok && starter.IsGatheringStartMove() {
			info.IsGatheringStart = true
		}
		for fieldName, fieldType := range move.ReadSetter().Props() {
			field := MoveField{Name: fieldName, Type: fieldType.String()}
			if fieldType == boardgame.TypeEnum {
				if value, _ := move.ReadSetter().EnumProp(fieldName); value != nil {
					field.EnumName = value.Enum().Name()
				}
			}
			info.Fields = append(info.Fields, field)
		}
		sort.Slice(info.Fields, func(i, j int) bool { return info.Fields[i].Name < info.Fields[j].Name })
		doc.Moves = append(doc.Moves, info)
	}
	return doc, nil
}

// legalityFromForms reduces the server's move forms to their legality.
func legalityFromForms(forms json.RawMessage) ([]MoveLegality, error) {
	var wire []struct {
		Name   string
		Fields []struct {
			Name         string
			DefaultValue json.RawMessage
		}
		LegalForPlayer      bool
		LegalForPlayerError string
		LegalForAnyone      bool
		Preconditions       json.RawMessage
	}
	if err := json.Unmarshal(forms, &wire); err != nil {
		return nil, err
	}
	result := make([]MoveLegality, 0, len(wire))
	for _, form := range wire {
		legality := MoveLegality{
			Name: form.Name, LegalForPlayer: form.LegalForPlayer, LegalForPlayerError: form.LegalForPlayerError,
			LegalForAnyone: form.LegalForAnyone, Defaults: map[string]json.RawMessage{},
		}
		for _, field := range form.Fields {
			value, err := canonical(field.DefaultValue)
			if err != nil {
				return nil, err
			}
			legality.Defaults[field.Name] = value
		}
		if len(form.Preconditions) > 0 && string(form.Preconditions) != "null" {
			preconditions, err := canonical(form.Preconditions)
			if err != nil {
				return nil, err
			}
			legality.Preconditions = preconditions
		}
		result = append(result, legality)
	}
	return result, nil
}

// activeTimersFromState rebuilds the ActiveTimers wire payload from a stored
// state: one entry per active timer, with the time left at Clock.
func activeTimersFromState(stored []byte) (map[string]interface{}, error) {
	var state struct {
		Timers []struct {
			ID       string
			Status   string
			Deadline time.Time
		}
	}
	if err := json.Unmarshal(stored, &state); err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	for _, timer := range state.Timers {
		if timer.Status != "active" || timer.ID == "" {
			continue
		}
		result[timer.ID] = map[string]interface{}{"TimeLeft": json.Number(fmt.Sprint(int64(timer.Deadline.Sub(Clock) / time.Millisecond)))}
	}
	return result, nil
}

// sameState compares a replayed state with the history's, ignoring timer
// deadlines (a golden recorded them from a wall clock) and, for histories
// older than durable timers, the Timers section entirely.
func sameState(actual, expected []byte) error {
	var got, want map[string]interface{}
	if err := decode(actual, &got); err != nil {
		return err
	}
	if err := decode(expected, &want); err != nil {
		return err
	}
	if _, ok := want["Timers"]; !ok {
		delete(got, "Timers")
	}
	for _, state := range []map[string]interface{}{got, want} {
		timers, _ := state["Timers"].([]interface{})
		for _, timer := range timers {
			if fields, ok := timer.(map[string]interface{}); ok {
				delete(fields, "Deadline")
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		return fmt.Errorf("state differs\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
	return nil
}

func winners(indexes []boardgame.PlayerIndex) []boardgame.PlayerIndex {
	if indexes == nil {
		return []boardgame.PlayerIndex{}
	}
	return indexes
}

func decode(blob []byte, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(blob))
	decoder.UseNumber()
	return decoder.Decode(target)
}

// canonical re-encodes JSON with sorted object keys and no extra whitespace.
func canonical(blob []byte) (json.RawMessage, error) {
	var value interface{}
	if err := decode(blob, &value); err != nil {
		return nil, fmt.Errorf("canonicalize %.60q: %w", blob, err)
	}
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(out.Bytes(), "\n"), nil
}

// replayStorage is in-memory record storage plus the seating rendezvous.
type replayStorage struct {
	*filesystem.StorageManager
	seating     bool
	pendingSeat *seat
}

func (s *replayStorage) FetchInjectedDataForGame(gameID, dataType string) interface{} {
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
	storage *replayStorage
}

func (s *seat) SeatIndex() boardgame.PlayerIndex { return s.index }
func (s *seat) Committed()                       { s.storage.pendingSeat = nil }
