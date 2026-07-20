package api

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jkomoros/boardgame"
)

const (
	projectedMoveChoicesStatusReady  = "ready"
	projectedMoveChoicesStatusFailed = "failed"

	maxProjectedMoveChoiceSets        = boardgame.MoveChoiceProjectionMaxSets
	maxProjectedMoveCandidatesPerSet  = boardgame.MoveChoiceProjectionMaxCandidatesPerSet
	maxProjectedMoveLegalEvaluations  = boardgame.MoveChoiceProjectionMaxLegalEvaluations
	maxProjectedMoveChoicesBytes      = boardgame.MoveChoiceProjectionMaxWireBytes
	maxProjectedMoveReconcileAttempts = 3
)

// projectedMoveChoicesSnapshot is a version-pinned, actor-only read model of
// finite move inputs. It intentionally contains no presentation text or legal
// error detail. A failed snapshot tells the client not to mistake a projection
// failure for an authoritative empty result.
type projectedMoveChoicesSnapshot struct {
	StateVersion                          int                      `json:"StateVersion"`
	MoveChoiceProjectionSchemaFingerprint string                   `json:"MoveChoiceProjectionSchemaFingerprint"`
	ProjectionSchemaVersion               int                      `json:"ProjectionSchemaVersion"`
	Status                                string                   `json:"Status"`
	Sets                                  []projectedMoveChoiceSet `json:"Sets,omitempty"`
}

type projectedMoveChoiceSet struct {
	MoveName   string                         `json:"MoveName"`
	FieldName  string                         `json:"FieldName"`
	Source     boardgame.MoveChoiceSource     `json:"Source"`
	Candidates []projectedMoveChoiceCandidate `json:"Candidates"`
}

type projectedMoveChoiceCandidate struct {
	Value     interface{} `json:"Value"`
	Available bool        `json:"Available"`
}

type projectedMoveChoiceSourceValue struct {
	value interface{}
	wire  string
}

type projectedMoveChoiceBudget struct {
	legalEvaluations int
}

type preparedProjectedMoveChoiceSet struct {
	schema boardgame.MoveChoiceProjectionSchema
	values []projectedMoveChoiceSourceValue
}

// reconcileProjectedMoveChoiceFrontier repairs missing durable boundary
// evidence before /info selects a state. Legacy records, a crash between the
// state write and marker write, and a partially committed fix-up chain all
// present identically as an unknown frontier. ForceFixUp serializes recovery
// through the game's ordinary main loop and resolves only after its recursive
// fix-up closure and terminal marker write.
func reconcileProjectedMoveChoiceFrontier(game *boardgame.Game) (*boardgame.Game, error) {
	if game == nil {
		return nil, fmt.Errorf("cannot reconcile projected choices for a nil game")
	}
	schema, err := boardgame.BuildMoveChoiceProjectionSchema(game.Manager())
	if err != nil {
		return nil, fmt.Errorf("build projected-choice schema: %w", err)
	}
	if len(schema) == 0 {
		return game, nil
	}
	if !boardgame.SupportsProposalFrontierStorage(game.Manager().Storage()) {
		return game, fmt.Errorf("projected choices require durable proposal-frontier storage after reload")
	}
	// The normal path is already certified. Consult the manager so an active
	// in-memory marker also counts, but require it to match this exact snapshot
	// before avoiding the storage reload.
	if projectedMoveChoiceFrontierKnown(game) {
		return game, nil
	}

	current := game.Manager().Game(game.ID())
	if current == nil {
		return game, fmt.Errorf("reload game before projected-choice reconciliation")
	}
	var lastErr error
	for attempt := 0; attempt < maxProjectedMoveReconcileAttempts; attempt++ {
		if projectedMoveChoiceFrontierKnown(current) {
			return current, nil
		}
		lastErr = <-current.Manager().Internals().ForceFixUp(current)

		// Always reload durable state, even after an error. A concurrent server
		// may have won the CAS and completed reconciliation while this request
		// observed a stale marker write.
		refreshed := current.Manager().Game(current.ID())
		if refreshed == nil {
			return current, fmt.Errorf("reload game after projected-choice reconciliation")
		}
		current = refreshed
		if projectedMoveChoiceFrontierKnown(current) {
			return current, nil
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("terminal fix-up check did not persist a proposal frontier")
		}
	}
	return current, fmt.Errorf("projected-choice frontier reconciliation failed after %d attempts: %w", maxProjectedMoveReconcileAttempts, lastErr)
}

func projectedMoveChoiceFrontierKnown(game *boardgame.Game) bool {
	if game == nil {
		return false
	}
	frontier, known := game.Manager().ProposalFrontierVersion(game.ID())
	return known && frontier == game.Version()
}

// projectMoveChoiceSet binds each member of one sealed candidate universe to
// a fresh move and asks the move's full Legal method exactly once. A nil set
// means either that the canonical move name is hidden from this actor or that
// no candidate is legal; both cases intentionally publish no choice set.
func projectMoveChoiceSet(
	game *boardgame.Game,
	state boardgame.ImmutableState,
	actor boardgame.PlayerIndex,
	schema boardgame.MoveChoiceProjectionSchema,
	budget *projectedMoveChoiceBudget,
) (*projectedMoveChoiceSet, error) {
	prepared, err := prepareMoveChoiceSet(game, state, actor, schema)
	if err != nil || prepared == nil {
		return nil, err
	}
	if err := preflightProjectedMoveChoiceSets(
		&projectedMoveChoicesSnapshot{Status: projectedMoveChoicesStatusReady},
		[]preparedProjectedMoveChoiceSet{*prepared}, budget,
	); err != nil {
		return nil, err
	}
	return evaluatePreparedMoveChoiceSet(game, state, actor, *prepared, budget)
}

func prepareMoveChoiceSet(
	game *boardgame.Game,
	state boardgame.ImmutableState,
	actor boardgame.PlayerIndex,
	schema boardgame.MoveChoiceProjectionSchema,
) (*preparedProjectedMoveChoiceSet, error) {
	if schema.Disclosure != boardgame.MoveChoiceDisclosureActorExact {
		return nil, fmt.Errorf("move %q uses unsupported choice disclosure %q", schema.MoveName, schema.Disclosure)
	}

	probe := game.MoveByNameForState(schema.MoveName, state)
	if probe == nil {
		return nil, fmt.Errorf("projected-choice move %q is not installed", schema.MoveName)
	}
	visible, err := game.MoveNameVisibleToPlayer(probe, actor, actor, state)
	if err != nil {
		return nil, fmt.Errorf("resolve move-name visibility for %q: %w", schema.MoveName, err)
	}
	if !visible {
		return nil, nil
	}

	values, err := projectedMoveChoiceSourceValues(state, actor, schema)
	if err != nil {
		return nil, fmt.Errorf("move %q field %q: %w", schema.MoveName, schema.FieldName, err)
	}
	if len(values) > maxProjectedMoveCandidatesPerSet {
		return nil, fmt.Errorf("move %q has %d projected candidates; limit is %d", schema.MoveName, len(values), maxProjectedMoveCandidatesPerSet)
	}
	return &preparedProjectedMoveChoiceSet{schema: schema, values: values}, nil
}

func evaluatePreparedMoveChoiceSet(
	game *boardgame.Game,
	state boardgame.ImmutableState,
	actor boardgame.PlayerIndex,
	prepared preparedProjectedMoveChoiceSet,
	budget *projectedMoveChoiceBudget,
) (*projectedMoveChoiceSet, error) {
	schema := prepared.schema

	set := &projectedMoveChoiceSet{
		MoveName:   schema.MoveName,
		FieldName:  schema.FieldName,
		Source:     schema.Source,
		Candidates: make([]projectedMoveChoiceCandidate, 0, len(prepared.values)),
	}
	legalCandidates := 0
	for _, sourceValue := range prepared.values {
		if budget.legalEvaluations >= maxProjectedMoveLegalEvaluations {
			return nil, fmt.Errorf("projected choices exceed %d Legal evaluations", maxProjectedMoveLegalEvaluations)
		}
		move := game.MoveByNameForState(schema.MoveName, state)
		if move == nil {
			return nil, fmt.Errorf("projected-choice move %q disappeared during projection", schema.MoveName)
		}
		if err := bindMoveFields(move, func(name string) (string, bool) {
			if name == schema.FieldName {
				return sourceValue.wire, true
			}
			return "", false
		}); err != nil {
			return nil, fmt.Errorf("bind candidate %q for move %q: %w", sourceValue.wire, schema.MoveName, err)
		}

		budget.legalEvaluations++
		// The count bound limits how many game-authored legality calls one
		// projection may make. Like every existing legality endpoint, this
		// assumes an individual Legal implementation terminates normally; the
		// framework deliberately does not launch unbounded goroutines to time it.
		available := move.Legal(state, actor) == nil
		if available {
			legalCandidates++
		}
		set.Candidates = append(set.Candidates, projectedMoveChoiceCandidate{
			Value:     sourceValue.value,
			Available: available,
		})
	}
	if legalCandidates == 0 {
		return nil, nil
	}
	return set, nil
}

// preflightProjectedMoveChoiceSets validates the complete visible candidate
// payload before any game-authored Legal method runs. It includes sets that may
// later prove all-illegal, so suppressed output cannot evade resource limits.
func preflightProjectedMoveChoiceSets(
	snapshot *projectedMoveChoicesSnapshot,
	prepared []preparedProjectedMoveChoiceSet,
	budget *projectedMoveChoiceBudget,
) error {
	preflight := *snapshot
	preflight.Sets = make([]projectedMoveChoiceSet, 0, len(prepared))
	totalEvaluations := budget.legalEvaluations
	for _, item := range prepared {
		totalEvaluations += len(item.values)
		if totalEvaluations > maxProjectedMoveLegalEvaluations {
			return fmt.Errorf("projected choices exceed %d Legal evaluations", maxProjectedMoveLegalEvaluations)
		}
		set := projectedMoveChoiceSet{
			MoveName:   item.schema.MoveName,
			FieldName:  item.schema.FieldName,
			Source:     item.schema.Source,
			Candidates: make([]projectedMoveChoiceCandidate, 0, len(item.values)),
		}
		for _, value := range item.values {
			// false is one byte longer than true in JSON, making this a safe
			// upper bound for the eventual status payload.
			set.Candidates = append(set.Candidates, projectedMoveChoiceCandidate{Value: value.value})
		}
		preflight.Sets = append(preflight.Sets, set)
	}
	return validateProjectedMoveChoicesSize(&preflight)
}

// projectMoveChoicesSnapshot projects only a recoverable proposal frontier.
// nil,nil means the supplied state or audience is intentionally ineligible;
// callers must omit the wire field in that case.
func projectMoveChoicesSnapshot(
	game *boardgame.Game,
	state boardgame.ImmutableState,
	actor boardgame.PlayerIndex,
	viewer boardgame.PlayerIndex,
) (*projectedMoveChoicesSnapshot, error) {
	if game == nil || state == nil {
		return nil, fmt.Errorf("projected choices require a game and immutable state")
	}
	if state.Game() != game {
		return nil, fmt.Errorf("projected-choice state does not belong to the game")
	}
	if state.Sanitized() {
		return nil, fmt.Errorf("projected choices require authoritative unsanitized state")
	}
	// Version one authorizes exact membership/status only to the proposing
	// actor. Observers, admins, and other players receive no snapshot.
	if actor < 0 || viewer != actor || !actor.Valid(state) {
		return nil, nil
	}

	schema, err := boardgame.BuildMoveChoiceProjectionSchema(game.Manager())
	if err != nil {
		return nil, err
	}
	// Do not add a feature envelope to every legacy game. Once at least one
	// projection is configured, an empty ready Sets array is authoritative.
	if len(schema) == 0 {
		return nil, nil
	}
	snapshot, err := newProjectedMoveChoicesSnapshot(game, state, projectedMoveChoicesStatusReady)
	if err != nil {
		return nil, err
	}
	frontier, known := game.Manager().ProposalFrontierVersion(game.ID())
	if !known {
		// Unknown is a failure only for the current durable head. A historical
		// animation bundle remains intentionally absent.
		fresh := game.Manager().Game(game.ID())
		if fresh == nil || fresh.Version() != state.Version() {
			return nil, nil
		}
		snapshot.Status = projectedMoveChoicesStatusFailed
		snapshot.Sets = nil
		return snapshot, nil
	}
	if state.Version() != frontier {
		return nil, nil
	}
	if len(schema) > maxProjectedMoveChoiceSets {
		return snapshot, fmt.Errorf("game configures %d projected choice sets; limit is %d", len(schema), maxProjectedMoveChoiceSets)
	}

	budget := new(projectedMoveChoiceBudget)
	prepared := make([]preparedProjectedMoveChoiceSet, 0, len(schema))
	for _, declaration := range schema {
		set, err := prepareMoveChoiceSet(game, state, actor, declaration)
		if err != nil {
			return snapshot, err
		}
		if set == nil {
			continue
		}
		prepared = append(prepared, *set)
	}
	if err := preflightProjectedMoveChoiceSets(snapshot, prepared, budget); err != nil {
		return snapshot, err
	}
	for _, preparedSet := range prepared {
		set, err := evaluatePreparedMoveChoiceSet(game, state, actor, preparedSet, budget)
		if err != nil {
			return snapshot, err
		}
		if set != nil {
			snapshot.Sets = append(snapshot.Sets, *set)
		}
	}
	return snapshot, nil
}

func newProjectedMoveChoicesSnapshot(game *boardgame.Game, state boardgame.ImmutableState, status string) (*projectedMoveChoicesSnapshot, error) {
	fingerprint, err := boardgame.MoveChoiceProjectionSchemaFingerprint(game.Manager())
	if err != nil {
		return nil, err
	}
	return &projectedMoveChoicesSnapshot{
		StateVersion:                          state.Version(),
		MoveChoiceProjectionSchemaFingerprint: fingerprint,
		ProjectionSchemaVersion:               boardgame.MoveChoiceProjectionSchemaVersion,
		Status:                                status,
		Sets:                                  make([]projectedMoveChoiceSet, 0),
	}, nil
}

func failedProjectedMoveChoicesForInfo(game *boardgame.Game, state boardgame.ImmutableState, actor boardgame.PlayerIndex) *projectedMoveChoicesSnapshot {
	if game == nil || state == nil || actor < 0 || !actor.Valid(state) {
		return nil
	}
	schema, err := boardgame.BuildMoveChoiceProjectionSchema(game.Manager())
	if err != nil || len(schema) == 0 {
		return nil
	}
	snapshot, err := newProjectedMoveChoicesSnapshot(game, state, projectedMoveChoicesStatusFailed)
	if err != nil {
		return nil
	}
	snapshot.Sets = nil
	return snapshot
}

func (s *Server) projectedMoveChoicesForInfo(
	game *boardgame.Game,
	state boardgame.ImmutableState,
	actor boardgame.PlayerIndex,
	audienceEligible bool,
	reconciliationErr error,
) *projectedMoveChoicesSnapshot {
	if !audienceEligible {
		return nil
	}
	if reconciliationErr != nil {
		return failedProjectedMoveChoicesForInfo(game, state, actor)
	}
	return s.projectedMoveChoicesForBundle(game, state, actor)
}

func validateProjectedMoveChoicesSize(snapshot *projectedMoveChoicesSnapshot) error {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode projected choices: %w", err)
	}
	if len(encoded) > maxProjectedMoveChoicesBytes {
		return fmt.Errorf("projected choices exceed %d bytes", maxProjectedMoveChoicesBytes)
	}
	return nil
}

// projectedMoveChoicesForBundle converts internal failures to an explicit,
// generic failed snapshot while logging the diagnostic only on the server.
func (s *Server) projectedMoveChoicesForBundle(
	game *boardgame.Game,
	state boardgame.ImmutableState,
	actor boardgame.PlayerIndex,
) *projectedMoveChoicesSnapshot {
	snapshot, err := projectMoveChoicesSnapshot(game, state, actor, actor)
	if err == nil {
		return snapshot
	}
	if s != nil && s.logger != nil {
		entry := s.logger.WithError(err)
		if game != nil {
			entry = entry.WithField("gameID", game.ID())
		}
		entry.Error("Projected move choices failed")
	}
	if snapshot == nil {
		// Invalid/non-versioned direct calls have no safe wire identity. Normal
		// bundle delivery always has an initialized snapshot before projection.
		return nil
	}
	snapshot.Status = projectedMoveChoicesStatusFailed
	snapshot.Sets = nil
	return snapshot
}

func projectedMoveChoiceSourceValues(state boardgame.ImmutableState, actor boardgame.PlayerIndex, schema boardgame.MoveChoiceProjectionSchema) ([]projectedMoveChoiceSourceValue, error) {
	switch schema.Source {
	case boardgame.MoveChoiceSourcePlayers:
		result := make([]projectedMoveChoiceSourceValue, 0, len(state.ImmutablePlayerStates()))
		for i := range state.ImmutablePlayerStates() {
			value := boardgame.PlayerIndex(i)
			if value.Valid(state) {
				result = append(result, projectedMoveChoiceSourceValue{value: value, wire: strconv.Itoa(i)})
			}
		}
		return result, nil
	case boardgame.MoveChoiceSourceEnumValues:
		result := make([]projectedMoveChoiceSourceValue, 0, len(schema.CandidateValues))
		for _, value := range schema.CandidateValues {
			result = append(result, projectedMoveChoiceSourceValue{value: value, wire: value})
		}
		return result, nil
	case boardgame.MoveChoiceSourceStackSlots:
		stack, err := boardgame.ResolveMoveChoiceStack(state, actor, schema.StackSource)
		if err != nil {
			return nil, err
		}
		if stack.Len() > boardgame.MoveChoiceProjectionMaxStackSlotsInspected {
			return nil, fmt.Errorf("stack has %d slots; inspected-slot limit is %d", stack.Len(), boardgame.MoveChoiceProjectionMaxStackSlotsInspected)
		}
		result := make([]projectedMoveChoiceSourceValue, 0, stack.NumComponents())
		for index := 0; index < stack.Len(); index++ {
			if stack.ImmutableComponentAt(index) == nil {
				continue
			}
			result = append(result, projectedMoveChoiceSourceValue{value: index, wire: strconv.Itoa(index)})
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported choice source %q", schema.Source)
	}
}
