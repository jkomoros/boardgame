package main

// An enum-typed move field (enum.RangeVal on checkers' Move Token), for the
// facade's `{ move }` Owner: how Go stores the field in a move record (the
// move's JSON, which storage.go's MoveStorageRecord.inflate unmarshals back),
// and what Value(), SetValue(k) and RangeValue() answer.

import (
	"encoding/json"
	"io"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/examples/checkers"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

// moveFieldRead is the field as stored in the move's JSON and read back.
type moveFieldRead struct {
	// Stored is the field's value in json.Marshal(move).
	Stored json.RawMessage
	Value  int
	// Range is RangeValue(); null where Go returns nil.
	Range []int
}

// moveFieldSet is one SetValue(Key) on the same move, in order: its error
// ("" for nil), then the field as read after it.
type moveFieldSet struct {
	Key   int
	Error string
	moveFieldRead
}

// moveFieldRecord is a move record whose field holds Stored, unmarshalled
// into a fresh move as Go inflates a stored move: the error ("" for nil),
// then the field as read after it.
type moveFieldRecord struct {
	Blob  json.RawMessage
	Error string
	moveFieldRead
}

type moveFieldVectors struct {
	Game    string
	Move    string
	Field   string
	Enum    string
	Initial moveFieldRead
	Sets    []moveFieldSet
	Records []moveFieldRecord
}

func readMoveField(move boardgame.Move, field string) moveFieldRead {
	blob, err := json.Marshal(move)
	check(err)
	var fields map[string]json.RawMessage
	check(json.Unmarshal(blob, &fields))
	val, err := move.ReadSetter().EnumProp(field)
	check(err)
	rangeVal := val.(enum.RangeVal)
	return moveFieldRead{Stored: fields[field], Value: int(val.Value()), Range: rangeVal.RangeValue()}
}

func moveFieldVecs() moveFieldVectors {
	manager, err := boardgame.NewGameManager(checkers.NewDelegate(), memstore.NewStorageManager())
	check(err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{Name: "checkers", ID: "FACADEMOVEVEC001", SecretSalt: "facadevec-salt", NumPlayers: 2})
	check(err)

	const moveName, field = "Move Token", "SpaceIndex"
	out := moveFieldVectors{Game: "checkers", Move: moveName, Field: field, Enum: "spaces"}

	move := game.MoveByName(moveName)
	out.Initial = readMoveField(move, field)
	for _, key := range []int{0, 1, 7, 8, 9, 35, 63, 64, -1, 100, 27, 2} {
		val, err := move.ReadSetter().EnumProp(field)
		check(err)
		set := moveFieldSet{Key: key}
		if err := val.SetValue(enum.EnumKey(key)); err != nil {
			set.Error = err.Error()
		}
		set.moveFieldRead = readMoveField(move, field)
		out.Sets = append(out.Sets, set)
	}

	for _, stored := range []string{`"0_0"`, `"3_4"`, `"7_7"`, `"2_5"`} {
		fresh := game.MoveByName(moveName)
		blob := json.RawMessage(`{"` + field + `":` + stored + `}`)
		record := moveFieldRecord{Blob: blob}
		if err := json.Unmarshal(blob, fresh); err != nil {
			record.Error = err.Error()
		}
		record.moveFieldRead = readMoveField(fresh, field)
		out.Records = append(out.Records, record)
	}
	return out
}
