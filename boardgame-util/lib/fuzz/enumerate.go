package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/server/api"
)

// candidate is one bound input for one move, with the seated players for
// whom Legal accepted and refused it.
type candidate struct {
	move    string
	input   map[string]interface{}
	legal   []boardgame.PlayerIndex
	illegal []boardgame.PlayerIndex
}

// step is everything available at one settled version.
type step struct {
	candidates []*candidate
	// actions are the legal actions, in a fixed order: proposals by move
	// table order, candidate order and actor, then the next timer, then a
	// seat.
	actions []action
	seat    *action
}

// actors are the players who may propose: every player in an unseated run;
// the players whose seat is filled in a seated one. The admin is not an
// actor (see the package's status note: the exporter replays a non-seat
// admin move as a timer firing, so an admin proposal could not be replayed).
func (d *driver) actors(state boardgame.ImmutableState) []boardgame.PlayerIndex {
	var result []boardgame.PlayerIndex
	for i, player := range state.ImmutablePlayerStates() {
		if d.derived.Seated {
			if seater, ok := player.(interfaces.Seater); ok && !seater.SeatIsFilled() {
				continue
			}
		}
		result = append(result, boardgame.PlayerIndex(i))
	}
	return result
}

func (d *driver) enumerate(uncovered map[string]bool) (*step, error) {
	state := d.game.CurrentState()
	actors := d.actors(state)
	projected, err := d.projections(actors)
	if err != nil {
		return nil, err
	}
	longest := longestStack(state)
	result := &step{}
	for _, move := range d.schema {
		defaults := d.game.MoveByName(move.Name)
		if defaults == nil {
			return nil, fmt.Errorf("move %q is not configured", move.Name)
		}
		var names []string
		var domains [][]interface{}
		for _, field := range move.Fields {
			if field.Disposition != string(boardgame.MoveInputRequired) && field.Disposition != string(boardgame.MoveInputServerDefaulted) {
				continue
			}
			domain, isDefaultOnly, err := d.domain(move.Name, field, defaults, projected, longest)
			if err != nil {
				return nil, err
			}
			if isDefaultOnly {
				key := move.Name + "." + field.Name
				if !uncovered[key] {
					uncovered[key] = true
					d.result.Uncovered = append(d.result.Uncovered, key)
				}
			}
			names = append(names, field.Name)
			domains = append(domains, domain)
		}
		for _, index := range d.candidateIndexes(domains) {
			input := map[string]interface{}{}
			rest := index
			for i := len(domains) - 1; i >= 0; i-- {
				size := int64(len(domains[i]))
				input[names[i]] = domains[i][rest%size]
				rest /= size
			}
			c := &candidate{move: move.Name, input: input}
			for _, actor := range actors {
				// A fresh move per actor: Legal may write to its move
				// (moves.AnyPlayer fills TargetPlayerIndex from the proposer
				// on its first call).
				bound, err := d.game.MoveWithInput(move.Name, input)
				if err != nil {
					// The binder refused it (an enum string or range value
					// out of bounds); it never reaches Legal.
					break
				}
				legalErr, panicked := safeLegal(bound, state, actor)
				switch {
				case panicked:
					d.result.LegalPanics++
				case legalErr == nil && actor.Valid(state):
					// The engine checks the proposer is valid (an active
					// player) before Legal; an inactive seated player's
					// proposal is refused however legal the move.
					c.legal = append(c.legal, actor)
				default:
					c.illegal = append(c.illegal, actor)
				}
			}
			if len(c.legal) == 0 && len(c.illegal) == 0 {
				continue
			}
			result.candidates = append(result.candidates, c)
			for _, actor := range c.legal {
				result.actions = append(result.actions, action{kind: "propose", move: move.Name, input: input, actor: actor})
			}
		}
	}
	active, err := d.timerActive()
	if err != nil {
		return nil, err
	}
	if active {
		result.actions = append(result.actions, action{kind: "timer", actor: boardgame.AdminPlayerIndex})
	}
	if seatAction := d.seatAction(state); seatAction != nil {
		result.seat = seatAction
		result.actions = append(result.actions, *seatAction)
	}
	return result, nil
}

func safeLegal(move boardgame.Move, state boardgame.ImmutableState, actor boardgame.PlayerIndex) (err error, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	return move.Legal(state, actor), false
}

// domain is one field's candidate values, per the proposal: both booleans;
// every player index; every enum key; for an integer, its projected choices
// if a projection declares them, else the default and 0 through the length
// of the longest stack; for a string, the default only. An integer-coded
// enum (a range enum) takes every key of its enum.
func (d *driver) domain(move string, field boardgame.MoveInputSchemaField, defaults boardgame.Move, projected map[string][]interface{}, longest int) ([]interface{}, bool, error) {
	reader := defaults.ReadSetter()
	switch boardgame.MoveInputCodec(field.Codec) {
	case boardgame.MoveInputCodecBoolean:
		return []interface{}{false, true}, false, nil
	case boardgame.MoveInputCodecPlayerIndex:
		var values []interface{}
		for i := 0; i < d.game.NumPlayers(); i++ {
			values = append(values, i)
		}
		return values, false, nil
	case boardgame.MoveInputCodecEnum:
		var values []interface{}
		for _, value := range field.EnumValues {
			values = append(values, value)
		}
		return values, false, nil
	case boardgame.MoveInputCodecString:
		value, err := reader.StringProp(field.Name)
		if err != nil {
			return nil, false, fmt.Errorf("%s.%s default: %w", move, field.Name, err)
		}
		return []interface{}{value}, true, nil
	case boardgame.MoveInputCodecInteger:
		if values, ok := projected[move+"\x00"+field.Name]; ok {
			return values, false, nil
		}
		if reader.Props()[field.Name] == boardgame.TypeEnum {
			value, err := reader.ImmutableEnumProp(field.Name)
			if err != nil || value == nil {
				return nil, false, fmt.Errorf("%s.%s enum: %v", move, field.Name, err)
			}
			keys := append([]enum.EnumKey(nil), value.Enum().Values()...)
			sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
			var values []interface{}
			for _, key := range keys {
				values = append(values, int(key))
			}
			return values, false, nil
		}
		def, err := reader.IntProp(field.Name)
		if err != nil {
			return nil, false, fmt.Errorf("%s.%s default: %w", move, field.Name, err)
		}
		values := []interface{}{def}
		for i := 0; i <= longest; i++ {
			if i != def {
				values = append(values, i)
			}
		}
		return values, false, nil
	default:
		return nil, false, fmt.Errorf("%s.%s has codec %q", move, field.Name, field.Codec)
	}
}

// candidateIndexes returns the product indexes to try: all of them, or a
// seeded sample of MaxCandidatesPerMove in increasing order.
func (d *driver) candidateIndexes(domains [][]interface{}) []int64 {
	total := int64(1)
	for _, domain := range domains {
		total *= int64(len(domain))
		if total > 1<<40 {
			total = 1 << 40
		}
	}
	if total <= MaxCandidatesPerMove {
		result := make([]int64, total)
		for i := range result {
			result[i] = int64(i)
		}
		return result
	}
	// Floyd's algorithm: MaxCandidatesPerMove distinct indexes.
	chosen := make(map[int64]bool, MaxCandidatesPerMove)
	for j := total - MaxCandidatesPerMove; j < total; j++ {
		t := d.random.Int63n(j + 1)
		if chosen[t] {
			chosen[j] = true
		} else {
			chosen[t] = true
		}
	}
	result := make([]int64, 0, len(chosen))
	for index := range chosen {
		result = append(result, index)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// projections reads each actor's declared choice projection and returns,
// for each "move\x00field", the union of the candidate values (available or
// not) in first-seen order.
func (d *driver) projections(actors []boardgame.PlayerIndex) (map[string][]interface{}, error) {
	result := map[string][]interface{}{}
	seen := map[string]map[string]bool{}
	for _, actor := range actors {
		blob, err := api.ProjectMoveChoicesForViewer(d.game, actor)
		if err != nil {
			return nil, fmt.Errorf("projected choices for %d: %w", actor, err)
		}
		if len(blob) == 0 {
			continue
		}
		var snapshot struct {
			Status string
			Sets   []struct {
				MoveName   string
				FieldName  string
				Candidates []struct{ Value json.RawMessage }
			}
		}
		decoder := json.NewDecoder(bytes.NewReader(blob))
		decoder.UseNumber()
		if err := decoder.Decode(&snapshot); err != nil {
			return nil, err
		}
		for _, set := range snapshot.Sets {
			key := set.MoveName + "\x00" + set.FieldName
			if seen[key] == nil {
				seen[key] = map[string]bool{}
				result[key] = []interface{}{}
			}
			for _, c := range set.Candidates {
				if seen[key][string(c.Value)] {
					continue
				}
				seen[key][string(c.Value)] = true
				var value interface{}
				inner := json.NewDecoder(bytes.NewReader(c.Value))
				inner.UseNumber()
				if err := inner.Decode(&value); err != nil {
					return nil, err
				}
				result[key] = append(result[key], value)
			}
		}
	}
	return result, nil
}

// longestStack is the length of the longest stack (or board space) in the
// game and player states.
func longestStack(state boardgame.ImmutableState) int {
	longest := 0
	consider := func(sub boardgame.ImmutableSubState) {
		if sub == nil {
			return
		}
		reader := sub.Reader()
		for name, kind := range reader.Props() {
			switch kind {
			case boardgame.TypeStack:
				if stack, err := reader.ImmutableStackProp(name); err == nil && stack != nil && stack.Len() > longest {
					longest = stack.Len()
				}
			case boardgame.TypeBoard:
				if board, err := reader.ImmutableBoardProp(name); err == nil && board != nil {
					for _, space := range board.ImmutableSpaces() {
						if space.Len() > longest {
							longest = space.Len()
						}
					}
				}
			}
		}
	}
	consider(state.ImmutableGameState())
	for _, player := range state.ImmutablePlayerStates() {
		consider(player)
	}
	return longest
}

// timerActive says whether the current stored state has an active timer.
func (d *driver) timerActive() (bool, error) {
	stored, err := d.storage.State(d.game.ID(), d.game.Version())
	if err != nil {
		return false, err
	}
	var state struct {
		Timers []struct {
			ID     string
			Status string
		}
	}
	if err := json.Unmarshal(stored, &state); err != nil {
		return false, err
	}
	for _, timer := range state.Timers {
		if timer.Status == "active" && timer.ID != "" {
			return true, nil
		}
	}
	return false, nil
}

// seatAction is seating the first open seat, in a seated run, when the
// game's seat move would accept it now.
func (d *driver) seatAction(state boardgame.ImmutableState) *action {
	if !d.derived.Seated || d.seatMove == "" {
		return nil
	}
	for i, player := range state.ImmutablePlayerStates() {
		seater, ok := player.(interfaces.Seater)
		if !ok || seater.SeatIsFilled() || seater.SeatIsClosed() {
			continue
		}
		index := boardgame.PlayerIndex(i)
		d.storage.pendingSeat = &seat{index: index, storage: d.storage}
		move := d.game.MoveByName(d.seatMove)
		var legalErr error = fmt.Errorf("no seat move")
		if move != nil {
			var panicked bool
			legalErr, panicked = safeLegal(move, state, boardgame.AdminPlayerIndex)
			if panicked {
				legalErr = fmt.Errorf("panicked")
			}
		}
		d.storage.pendingSeat = nil
		if legalErr != nil {
			return nil
		}
		return &action{kind: "seat", seat: index, actor: boardgame.AdminPlayerIndex}
	}
	return nil
}
