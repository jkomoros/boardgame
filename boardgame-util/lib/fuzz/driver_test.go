package fuzz

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
)

// TestEveryGameRuns runs one short seed of every known game and exports it.
func TestEveryGameRuns(t *testing.T) {
	for _, game := range Games() {
		game := game
		t.Run(game, func(t *testing.T) {
			doc, result, err := Document(Config{Game: game, Seed: 1, MaxVersions: 40})
			if err != nil {
				t.Fatal(err)
			}
			if len(doc.Versions) == 0 {
				t.Fatal("no versions")
			}
			t.Logf("%s: %d versions, end %s, %d proposals, %d refusals, seated %t, players %d",
				game, len(doc.Versions), result.End, result.Proposals, len(result.Refusals), result.Derived.Seated, result.Derived.NumPlayers)
		})
	}
}

// TestRunIsDeterministic runs the same Config twice and requires identical
// document bytes.
func TestRunIsDeterministic(t *testing.T) {
	for _, cfg := range []Config{
		{Game: "memory", Seed: 3, MaxVersions: 60},
		{Game: "tictactoe", Seed: 5},
		{Game: "werewolf", Seed: 2, MaxVersions: 40},
	} {
		var runs [2][]byte
		for i := range runs {
			doc, _, err := Document(cfg)
			if err != nil {
				t.Fatalf("%+v: %v", cfg, err)
			}
			if runs[i], err = doc.Marshal(); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(runs[0], runs[1]) {
			t.Errorf("%+v: two runs differ", cfg)
		}
	}
}

// TestTruncatedRunIsPrefix requires a run cut short by a smaller
// MaxVersions to be a prefix of the longer run: the same versions, byte for
// byte, and the rejections recorded before its last version.
func TestTruncatedRunIsPrefix(t *testing.T) {
	for _, cfg := range []Config{
		{Game: "memory", Seed: 4, MaxVersions: 80},
		{Game: "pig", Seed: 6, MaxVersions: 80},
		{Game: "blackjack", Seed: 2, MaxVersions: 60},
	} {
		full, _, err := Document(cfg)
		if err != nil {
			t.Fatal(err)
		}
		short := cfg
		short.MaxVersions = cfg.MaxVersions / 3
		prefix, _, err := Document(short)
		if err != nil {
			t.Fatal(err)
		}
		if len(prefix.Versions) >= len(full.Versions) {
			t.Fatalf("%s: truncated run has %d versions, full run %d", cfg.Game, len(prefix.Versions), len(full.Versions))
		}
		for i, version := range prefix.Versions {
			a, _ := json.Marshal(version)
			b, _ := json.Marshal(full.Versions[i])
			if !bytes.Equal(a, b) {
				t.Fatalf("%s: version %d differs between the truncated and full runs", cfg.Game, i)
			}
		}
		last := prefix.Versions[len(prefix.Versions)-1].Version
		var want []conformance.Rejection
		for _, rejection := range full.Rejections {
			if rejection.AtVersion < last {
				want = append(want, rejection)
			}
		}
		a, _ := json.Marshal(prefix.Rejections)
		b, _ := json.Marshal(want)
		if len(want) == 0 {
			b = []byte("null")
		}
		if !bytes.Equal(a, b) {
			t.Fatalf("%s: rejections are not a prefix:\n%s\n%s", cfg.Game, a, b)
		}
		if len(prefix.Rejections) == 0 && len(full.Rejections) > 0 && prefix.SchemaVersion != conformance.SchemaVersion {
			t.Fatalf("%s: a document without rejections must be version 1", cfg.Game)
		}
	}
}
