package scenario_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/memory/scenarios"
)

func TestMemoryScenarioProducesRealHistoryAndViewerFrames(t *testing.T) {
	spec, err := scenarios.Mismatch()
	if err != nil {
		t.Fatal(err)
	}
	result, err := scenario.Run(memory.NewDelegate(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Replay.Frames) != 4 {
		t.Fatalf("frames=%d", len(result.Replay.Frames))
	}
	for i, frame := range result.Replay.Frames {
		if i > 0 && frame.Version <= result.Replay.Frames[i-1].Version {
			t.Fatal("scenario did not advance the engine")
		}
		for _, snapshot := range frame.Viewers {
			if strings.Contains(string(snapshot.Game), "SecretSalt") || strings.Contains(string(snapshot.Game), "SecretMoveCount") {
				t.Fatal("viewer artifact contains authoritative secrets")
			}
			var wire struct {
				CurrentState struct {
					Game struct {
						VisibleCards struct{ Indexes []int }
						HiddenCards  struct{ Indexes []int }
					}
				}
			}
			if err := json.Unmarshal(snapshot.Game, &wire); err != nil {
				t.Fatal(err)
			}
			visible := 0
			for _, index := range wire.CurrentState.Game.VisibleCards.Indexes {
				if index >= 0 {
					visible++
				}
			}
			want := []int{0, 1, 2, 0}[i]
			if visible != want {
				t.Fatalf("frame=%d viewer=%d visible=%d want=%d", i, snapshot.Viewer, visible, want)
			}
			for _, index := range wire.CurrentState.Game.HiddenCards.Indexes {
				if index >= 0 {
					t.Fatal("hidden card identity exposed")
				}
			}
		}
	}
	path := filepath.Join(t.TempDir(), "memory.json")
	if err := result.History.Save(path, false); err != nil {
		t.Fatal(err)
	}
	if err := golden.Compare(memory.NewDelegate(), path, false); err != nil {
		t.Fatalf("scenario history did not replay: %v", err)
	}
}

func TestScenarioFailureReturnsCompletedPrefix(t *testing.T) {
	spec, err := scenarios.Mismatch()
	if err != nil {
		t.Fatal(err)
	}
	spec.Steps[1].Input = spec.Steps[0].Input // revealing the same slot is illegal
	result, err := scenario.Run(memory.NewDelegate(), spec)
	if err == nil || !strings.Contains(err.Error(), "step 2") {
		t.Fatalf("expected annotated failure, got %v", err)
	}
	if result == nil || len(result.Replay.Frames) != 2 {
		t.Fatal("lost completed trace")
	}
	if result.History.Game().Version != result.Replay.Frames[1].Version {
		t.Fatal("rejected move changed durable history")
	}
}

func TestScenarioBudgetCountsAutomaticMoves(t *testing.T) {
	spec, err := scenarios.Mismatch()
	if err != nil {
		t.Fatal(err)
	}
	spec.MaxMoves = 2
	_, err = scenario.Run(memory.NewDelegate(), spec)
	if err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("expected setup fix-ups to exhaust budget, got %v", err)
	}
}
