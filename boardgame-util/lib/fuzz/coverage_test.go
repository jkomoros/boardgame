package fuzz

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestCoverageMatchesIndex reads the checked-in corpus and requires the
// move counts read from the documents to equal the ones the driver counted
// while running (conformance/fuzz/index.json), for every game.
func TestCoverageMatchesIndex(t *testing.T) {
	dir := filepath.Join(repoRoot, CorpusDir)
	docs, err := LoadDocuments(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	blob, err := os.ReadFile(filepath.Join(dir, CorpusIndexFile))
	if err != nil {
		t.Fatal(err)
	}
	var index Index
	if err := json.Unmarshal(blob, &index); err != nil {
		t.Fatal(err)
	}
	if len(docs) != len(index.Games) {
		t.Fatalf("%d games on disk, %d in the index", len(docs), len(index.Games))
	}
	for _, summary := range index.Games {
		report, err := Coverage(docs[summary.Game])
		if err != nil {
			t.Fatalf("%s: %v", summary.Game, err)
		}
		if report.Documents != summary.Seeds {
			t.Errorf("%s: %d documents, index says %d seeds", summary.Game, report.Documents, summary.Seeds)
		}
		if len(report.Moves) != len(summary.Moves) {
			t.Fatalf("%s: %d moves, index has %d", summary.Game, len(report.Moves), len(summary.Moves))
		}
		rejections := 0
		for _, n := range summary.RejectionsByKind {
			rejections += n
		}
		if report.Rejections != rejections {
			t.Errorf("%s: %d rejections, index says %d", summary.Game, report.Rejections, rejections)
		}
		for i, want := range summary.Moves {
			got := report.Moves[i]
			if got.Name != want.Name || got.IsFixUp != want.IsFixUp || got.Applied != want.Applied || got.Refused != want.Refused {
				t.Errorf("%s: move %d is %+v, index says %+v", summary.Game, i, got, want)
			}
			// Every legal proposal the driver made was accepted (the corpus
			// has no apply-error refusal), and is a "propose" version.
			if !want.IsFixUp && got.Proposed != want.ProposedLegally {
				t.Errorf("%s %s: %d accepted proposals, index says %d proposed legally", summary.Game, want.Name, got.Proposed, want.ProposedLegally)
			}
		}
		var neverApplied []string
		for _, move := range report.Moves {
			if move.Applied == 0 {
				neverApplied = append(neverApplied, move.Name)
			}
		}
		if !reflect.DeepEqual(report.Gaps().NeverApplied, neverApplied) || len(neverApplied) != len(summary.NeverApplied) {
			t.Errorf("%s: never applied %v, index says %v", summary.Game, report.Gaps().NeverApplied, summary.NeverApplied)
		}
		if text := report.Text(); !strings.HasPrefix(text, "## "+summary.Game+"\n") {
			t.Errorf("%s: text starts %.40q", summary.Game, text)
		}
	}
}

// TestCoverageMeasures checks the predicate and visibility rows on games
// whose shape is known from their source: tictactoe declares a plan and
// hides nothing; memory hides HiddenCards from everyone; werewolf's roles
// are hidden from other players.
func TestCoverageMeasures(t *testing.T) {
	docs, err := LoadDocuments(filepath.Join(repoRoot, CorpusDir), "")
	if err != nil {
		t.Fatal(err)
	}
	report := func(game string) Report {
		r, err := Coverage(docs[game])
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	tictactoe := report("tictactoe")
	if len(tictactoe.Visibility) != 0 {
		t.Errorf("tictactoe hides %+v", tictactoe.Visibility)
	}
	if len(tictactoe.Predicates) == 0 || tictactoe.Predicates[0].Move != "Place Token" {
		t.Errorf("tictactoe predicates %+v", tictactoe.Predicates)
	}
	verdicts := 0
	for _, doc := range docs["tictactoe"] {
		for _, version := range doc.Versions {
			for _, view := range version.Views {
				for _, legality := range view.Legality {
					var list []precondition
					if len(legality.Preconditions) > 0 {
						if err := json.Unmarshal(legality.Preconditions, &list); err != nil {
							t.Fatal(err)
						}
					}
					verdicts += len(list)
				}
			}
		}
	}
	counted := 0
	for _, p := range tictactoe.Predicates {
		counted += p.Pass + p.Fail + p.Unknown
	}
	if counted != verdicts {
		t.Errorf("tictactoe: %d verdicts counted, %d in the documents", counted, verdicts)
	}

	memory := report("memory")
	if len(memory.Visibility) != 1 || memory.Visibility[0].Property != "Game.HiddenCards" || memory.Visibility[0].SanitizedNonZero == 0 || memory.Visibility[0].RevealedNonZero != 0 {
		t.Errorf("memory visibility %+v", memory.Visibility)
	}
	werewolf := report("werewolf")
	var properties []string
	for _, v := range werewolf.Visibility {
		properties = append(properties, v.Property)
	}
	if !reflect.DeepEqual(properties, []string{"Players.FellowWolves", "Players.NightVote", "Players.Role"}) {
		t.Errorf("werewolf hidden properties %v", properties)
	}
}

func TestIsZero(t *testing.T) {
	for blob, want := range map[string]bool{
		`null`: true, `false`: true, `0`: true, `""`: true, `[]`: true, `[0, 0]`: true, `{}`: true,
		`true`: false, `-1`: false, `"Day"`: false, `[0, 1]`: false,
		`{"Deck":"cards","IDs":["",""],"Indexes":[-1,-1],"Size":2}`: true,
		`{"Deck":"cards","IDs":["x"],"Indexes":[3]}`:                false,
		`{"Deck":"cards","IDs":[],"Indexes":[]}`:                    true,
		`{"ID":"","IsTimer":true}`:                                  true,
		`{"ID":"abc","IsTimer":true}`:                               false,
	} {
		if got := isZero(json.RawMessage(blob)); got != want {
			t.Errorf("isZero(%s) = %t, want %t", blob, got, want)
		}
	}
}

// TestBlockCoverage maps a hand-written profile onto a small package.
func TestBlockCoverage(t *testing.T) {
	dir := t.TempDir()
	source := `package demo

func reached(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}

func (d *thing) never() {
	d.n++
}

type thing struct{ n int }
`
	if err := os.WriteFile(filepath.Join(dir, "demo.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	profile := `mode: set
example.com/demo/demo.go:3.25,4.11 1 1
example.com/demo/demo.go:4.11,6.3 1 0
example.com/demo/demo.go:7.2,7.10 1 1
example.com/demo/demo.go:3.25,4.11 1 0
example.com/demo/demo.go:10.25,12.2 1 0
example.com/other/other.go:1.1,2.2 1 1
`
	blocks, err := ParseProfile(strings.NewReader(profile))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 5 {
		t.Fatalf("%d blocks, want 5 (one duplicate merged)", len(blocks))
	}
	got, err := BlockCoverage(blocks, "example.com/demo", dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []FuncCoverage{
		{File: "demo.go", Line: 3, Func: "reached", Blocks: 3, Reached: 2, Unreached: []string{"4.11-6.3"}},
		{File: "demo.go", Line: 10, Func: "(*thing).never", Blocks: 1, Reached: 0, Unreached: []string{"10.25-12.2"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	gaps := Report{Blocks: got}.Gaps().Functions
	if len(gaps) != 2 {
		t.Errorf("function gaps %v", gaps)
	}
}
