package conformance_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
	"github.com/jkomoros/boardgame/examples/memory"
	memoryscenarios "github.com/jkomoros/boardgame/examples/memory/scenarios"
	"github.com/jkomoros/boardgame/examples/werewolf"
	werewolfscenarios "github.com/jkomoros/boardgame/examples/werewolf/scenarios"
)

func decodeObject(t *testing.T, blob []byte) map[string]interface{} {
	t.Helper()
	var value map[string]interface{}
	if err := json.Unmarshal(blob, &value); err != nil {
		t.Fatalf("decode %.80s: %v", blob, err)
	}
	return value
}

// TestExportAgreesWithScenarioReplay checks the export against an
// independent witness: scenario.Run captures each viewer's API snapshot
// live, at the moment each version is the head. The export of the same
// script must carry the same game object, legality and projected choices
// for those versions and viewers.
func TestExportAgreesWithScenarioReplay(t *testing.T) {
	cases := map[string]struct {
		newDelegate conformance.NewDelegate
		spec        func() (scenario.Spec, error)
	}{
		"memory":   {func() boardgame.GameDelegate { return memory.NewDelegate() }, memoryscenarios.Mismatch},
		"werewolf": {func() boardgame.GameDelegate { return werewolf.NewDelegate() }, werewolfscenarios.TimedVote},
	}
	for name, test := range cases {
		spec, err := test.spec()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		witness, err := scenario.Run(test.newDelegate(), spec)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		doc, err := conformance.FromScenario(test.newDelegate, spec, "")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(doc.Versions) != witness.History.Game().Version+1 {
			t.Fatalf("%s: %d versions exported, history has %d", name, len(doc.Versions), witness.History.Game().Version+1)
		}
		if len(doc.Viewers) != spec.Players+2 {
			t.Fatalf("%s: viewers = %v", name, doc.Viewers)
		}
		compared := 0
		for _, frame := range witness.Replay.Frames {
			version := doc.Versions[frame.Version]
			if !version.Settled {
				t.Errorf("%s: version %d is a scenario frame but not settled", name, frame.Version)
			}
			for _, snapshot := range frame.Viewers {
				var view *conformance.View
				for i := range version.Views {
					if version.Views[i].Viewer == snapshot.Viewer {
						view = &version.Views[i]
					}
				}
				if view == nil {
					t.Fatalf("%s: version %d has no view for %d", name, frame.Version, snapshot.Viewer)
				}
				want := decodeObject(t, snapshot.Game)
				delete(want, "Diagram")
				if winners, _ := want["Winners"].([]interface{}); winners == nil {
					want["Winners"] = []interface{}{}
				}
				if got := decodeObject(t, view.Game); !reflect.DeepEqual(got, want) {
					gotJSON, _ := json.Marshal(got)
					wantJSON, _ := json.Marshal(want)
					t.Errorf("%s: version %d viewer %d: game differs from the live snapshot\n got: %s\nwant: %s", name, frame.Version, snapshot.Viewer, gotJSON, wantJSON)
				}
				if len(view.Legality) != len(snapshot.MoveLegality) {
					t.Errorf("%s: version %d viewer %d: %d moves offered, live snapshot has %d", name, frame.Version, snapshot.Viewer, len(view.Legality), len(snapshot.MoveLegality))
				}
				for _, legality := range view.Legality {
					live, ok := snapshot.MoveLegality[legality.Name]
					if !ok || live.LegalForPlayer != legality.LegalForPlayer || live.LegalForAnyone != legality.LegalForAnyone {
						t.Errorf("%s: version %d viewer %d: %q legality %+v, live snapshot %+v (present %v)", name, frame.Version, snapshot.Viewer, legality.Name, legality, live, ok)
					}
				}
				if (len(snapshot.ProjectedMoveChoices) == 0) != (len(view.ProjectedMoveChoices) == 0) {
					t.Errorf("%s: version %d viewer %d: projected choices presence differs", name, frame.Version, snapshot.Viewer)
				} else if len(snapshot.ProjectedMoveChoices) > 0 && !reflect.DeepEqual(decodeObject(t, view.ProjectedMoveChoices), decodeObject(t, snapshot.ProjectedMoveChoices)) {
					t.Errorf("%s: version %d viewer %d: projected choices differ", name, frame.Version, snapshot.Viewer)
				}
				compared++
			}
		}
		if compared == 0 {
			t.Fatalf("%s: nothing was compared", name)
		}
	}
}

// TestExportShape checks the per-version invariants a consumer relies on.
func TestExportShape(t *testing.T) {
	spec, err := memoryscenarios.Mismatch()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := conformance.FromScenario(func() boardgame.GameDelegate { return memory.NewDelegate() }, spec, "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Game.SecretSalt == "" || doc.Game.ID == "" {
		t.Fatal("the id and secret salt must be exported")
	}
	if len(doc.Normalization) == 0 {
		t.Fatal("no normalization list")
	}
	sawTimer, sawFixUp := false, false
	for i, version := range doc.Versions {
		if version.Version != i {
			t.Fatalf("versions[%d].version = %d", i, version.Version)
		}
		if (version.Move == nil) != (i == 0) {
			t.Fatalf("version %d: move presence is wrong", i)
		}
		if len(version.Views) != len(doc.Viewers) {
			t.Fatalf("version %d: %d views for %d viewers", i, len(version.Views), len(doc.Viewers))
		}
		state := decodeObject(t, version.State)
		if _, ok := state["Timers"]; !ok {
			t.Errorf("version %d: stored state has no Timers", i)
		}
		if _, ok := state["Computed"]; ok {
			t.Errorf("version %d: stored state must not carry Computed", i)
		}
		for j, view := range version.Views {
			if view.Viewer != doc.Viewers[j] {
				t.Fatalf("version %d: view %d is for %d, want %d", i, j, view.Viewer, doc.Viewers[j])
			}
			game := decodeObject(t, view.Game)
			if _, ok := game["Diagram"]; ok {
				t.Errorf("version %d viewer %d: Diagram was not omitted", i, view.Viewer)
			}
			if game["Version"] != float64(i) {
				t.Errorf("version %d viewer %d: game.Version = %v", i, view.Viewer, game["Version"])
			}
			current := game["CurrentState"].(map[string]interface{})
			for _, key := range []string{"Game", "Players", "Computed", "Visibility", "Components"} {
				if _, ok := current[key]; !ok {
					t.Errorf("version %d viewer %d: CurrentState has no %s", i, view.Viewer, key)
				}
			}
			if _, ok := current["SecretMoveCount"]; ok && view.Viewer != boardgame.AdminPlayerIndex {
				t.Errorf("version %d viewer %d: sanitized state leaks SecretMoveCount", i, view.Viewer)
			}
			if view.ProjectedMoveChoices != nil && (!version.Settled || view.Viewer < 0) {
				t.Errorf("version %d viewer %d: unexpected projected choices", i, view.Viewer)
			}
		}
		if version.Move != nil {
			sawTimer = sawTimer || version.Move.Trigger == "timer"
			sawFixUp = sawFixUp || (version.Move.Trigger == "fixup" && version.Move.IsFixUp)
			if version.Move.Trigger == "fixup" && version.Move.Initiator == version.Move.Version {
				t.Errorf("version %d: a fixup trigger cannot start its own chain", i)
			}
		}
	}
	if !sawTimer || !sawFixUp {
		t.Fatalf("expected a timer-fired move and a fix-up move (timer %v, fixup %v)", sawTimer, sawFixUp)
	}
	if !doc.Versions[len(doc.Versions)-1].Settled {
		t.Fatal("the final version must be settled")
	}
}
