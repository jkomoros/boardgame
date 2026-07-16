package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jkomoros/boardgame"
)

func TestAnimationMoveRecordExcludesPrivateStorageFields(t *testing.T) {
	storage := &boardgame.MoveStorageRecord{
		Name: "Choose Secret Card", Version: 17, Initiator: 9, Phase: 3,
		Proposer: 2, Timestamp: time.Unix(123, 0), Blob: json.RawMessage(`{"Target":4}`),
	}
	encoded, err := json.Marshal(animationMoveRecord(storage))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["Name"] != storage.Name || got["Version"] != float64(storage.Version) {
		t.Fatalf("client move = %s; want only Name and Version", encoded)
	}
	for _, forbidden := range []string{"Blob", "Proposer", "Initiator", "Phase", "Timestamp"} {
		if _, exists := got[forbidden]; exists {
			t.Fatalf("client move leaked %s: %s", forbidden, encoded)
		}
	}
	if animationMoveRecord(nil) != nil {
		t.Fatal("nil storage move should remain nil for the initial bundle")
	}
}
