package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jkomoros/boardgame"
)

func TestFailedDiskWriteDoesNotPublishStagedTimerState(t *testing.T) {
	basePath := t.TempDir()
	storage := NewStorageManager(basePath)
	if err := storage.Connect(""); err != nil {
		t.Fatal(err)
	}

	game := &boardgame.GameStorageRecord{Name: "timer-test", ID: fmt.Sprintf("TIMER%d", time.Now().UnixNano()), Version: 0}
	const inactiveJSON = `{"Game":{},"Players":[],"Timers":[]}`
	const moveJSON = `{"choice":"original"}`
	inactive := boardgame.StateStorageRecord(inactiveJSON)
	if err := storage.SaveGameAndCurrentState(game, inactive, nil); err != nil {
		t.Fatal(err)
	}
	committedHead := *game
	committedHead.Version = 1
	move := &boardgame.MoveStorageRecord{Version: 1, Blob: []byte(moveJSON)}
	if err := storage.SaveGameAndCurrentState(&committedHead, inactive, move); err != nil {
		t.Fatal(err)
	}
	committedHead.Finished = true
	inactive[1] = 'X'
	move.Blob[11] = 'X'
	storedGame, err := storage.Game(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	storedState, err := storage.State(game.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	storedMove, err := storage.Move(game.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if storedGame.Finished || !json.Valid(storedState) || string(storedMove.Blob) != moveJSON {
		t.Fatalf("mutating save inputs changed committed filesystem record: finished=%t state=%s move=%s", storedGame.Finished, storedState, storedMove.Blob)
	}
	committedHead.Finished = false

	failedHead := committedHead
	failedHead.Version = 2
	deadline := time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
	active := boardgame.StateStorageRecord(`{"Game":{},"Players":[],"Timers":[` +
		`{"Ref":{},"ID":"uncommitted-timer","Generation":1,"Deadline":"` + deadline + `",` +
		`"Status":"active","Move":{"Name":"Secret Completion","Blob":{}}}]}`)

	blocker := filepath.Join(basePath, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block nested writes"), 0600); err != nil {
		t.Fatal(err)
	}
	storage.basePath = blocker
	if err := storage.SaveGameAndCurrentState(&failedHead, active, nil); err == nil {
		t.Fatal("save through a non-directory base unexpectedly succeeded")
	}
	storage.basePath = basePath

	stored, err := storage.Game(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Version != committedHead.Version {
		t.Fatalf("failed save published cached head version %d; want %d", stored.Version, committedHead.Version)
	}
	wakeups, err := storage.TimerWakeups(game.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(wakeups) != 0 {
		t.Fatalf("failed save published %d durable timer wakeups", len(wakeups))
	}
}
