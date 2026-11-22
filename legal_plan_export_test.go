package boardgame_test

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/werewolf"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

var planExampleGames = []struct {
	name        string
	newDelegate func() boardgame.GameDelegate
	// inCorpus is true for the games with documents under conformance/.
	inCorpus bool
}{
	{"memory", memory.NewDelegate, true},
	{"blackjack", blackjack.NewDelegate, true},
	{"pig", pig.NewDelegate, true},
	{"tictactoe", tictactoe.NewDelegate, true},
	{"werewolf", werewolf.NewDelegate, true},
	{"debuganimations", debuganimations.NewDelegate, true},
	{"checkers", checkers.NewDelegate, false},
}

func newPlanManager(t *testing.T, newDelegate func() boardgame.GameDelegate) *boardgame.GameManager {
	t.Helper()
	manager, err := boardgame.NewGameManager(newDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	t.Cleanup(func() { manager.Internals().Close() })
	return manager
}

// TestLegalPlanSpecsCallable: every move of every example game answers, an
// unknown name does not, and the result is a copy.
func TestLegalPlanSpecsCallable(t *testing.T) {
	declaredTotal := 0
	for _, game := range planExampleGames {
		manager := newPlanManager(t, game.newDelegate)
		var declared []string
		for _, move := range manager.ExampleMoves() {
			name := move.Info().Name()
			specs, isDeclared, hasCustom, ok := manager.LegalPlanSpecs(name)
			if !ok {
				t.Errorf("%s %q: LegalPlanSpecs not ok", game.name, name)
				continue
			}
			if !isDeclared {
				if specs != nil || hasCustom {
					t.Errorf("%s %q: undeclared plan with specs %v, custom %v", game.name, name, specs, hasCustom)
				}
				continue
			}
			declared = append(declared, name)
			declaredTotal++

			// Case is ignored, as for ExampleMoveByName.
			upper, upperDeclared, upperCustom, upperOK := manager.LegalPlanSpecs(strings.ToUpper(name))
			if !upperOK || !upperDeclared || upperCustom != hasCustom || fmt.Sprint(upper) != fmt.Sprint(specs) {
				t.Errorf("%s %q: upper-case lookup differs", game.name, name)
			}

			// Mutating the result does not reach the manager.
			before := fmt.Sprintf("%#v", specs)
			for i := range specs {
				specs[i].Name = "mutated"
				for j := range specs[i].Args {
					specs[i].Args[j] = "mutated"
				}
				for j := range specs[i].Sub {
					specs[i].Sub[j].Name = "mutated"
				}
			}
			again, _, _, _ := manager.LegalPlanSpecs(name)
			if after := fmt.Sprintf("%#v", again); after != before {
				t.Errorf("%s %q: mutating the returned specs changed the manager's plan:\nbefore %s\nafter  %s", game.name, name, before, after)
			}
		}
		if _, _, _, ok := manager.LegalPlanSpecs("No Such Move In This Game"); ok {
			t.Errorf("%s: unknown move name is ok", game.name)
		}
		t.Logf("%s: %d moves, declared plans: %q", game.name, len(manager.ExampleMoves()), declared)
	}
	var nilManager *boardgame.GameManager
	if _, _, _, ok := nilManager.LegalPlanSpecs("x"); ok {
		t.Error("nil manager is ok")
	}
	if declaredTotal == 0 {
		t.Error("no example game declares a plan")
	}
}

// ledgerEntry is what a conformance document records of one predicate:
// name, args and, when its verdict is not pass, the message template.
type ledgerEntry struct {
	Name    string   `json:"name"`
	Args    []string `json:"args"`
	Message *struct {
		Template string `json:"template"`
	} `json:"message"`
}

type ledgerDocument struct {
	Moves []struct {
		Name string `json:"name"`
	} `json:"moves"`
	Versions []struct {
		Views []struct {
			Legality []struct {
				Name          string          `json:"name"`
				Preconditions json.RawMessage `json:"preconditions"`
			} `json:"legality"`
		} `json:"views"`
	} `json:"versions"`
}

func readLedgerDocument(t *testing.T, path string) ledgerDocument {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	var doc ledgerDocument
	if err := json.NewDecoder(reader).Decode(&doc); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return doc
}

func entryKey(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + "(" + strings.Join(args, ", ") + ")"
}

// TestLegalPlanSpecsMatchLedgers: for the games in the conformance corpus,
// every precondition ledger recorded for a move lists exactly the plan
// LegalPlanSpecs returns, in order: each spec's name and args, then "custom"
// when the plan has the LegalCustom tail. A spec's Message, when set, equals
// the ledger's message template wherever the ledger records one (only for a
// non-pass verdict). The ledger records neither a compositor's Sub specs nor
// AdminPolicy, so those are not compared. A move offered in some view with
// no ledger must have no plan (or an empty one).
func TestLegalPlanSpecsMatchLedgers(t *testing.T) {
	for _, game := range planExampleGames {
		if !game.inCorpus {
			continue
		}
		manager := newPlanManager(t, game.newDelegate)
		paths, err := filepath.Glob(filepath.Join("conformance", game.name, "*.json.gz"))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) == 0 {
			t.Errorf("%s: no conformance documents", game.name)
			continue
		}

		type observed struct {
			ledgers map[string]bool // entry keys joined by " | "
			absent  bool            // offered at least once with no ledger
		}
		byMove := map[string]*observed{}
		ledgerCount := 0
		messageChecks := 0
		for _, path := range paths {
			doc := readLedgerDocument(t, path)
			for _, move := range doc.Moves {
				if _, _, _, ok := manager.LegalPlanSpecs(move.Name); !ok {
					t.Errorf("%s: %s lists move %q, which LegalPlanSpecs does not know", game.name, path, move.Name)
				}
			}
			for _, version := range doc.Versions {
				for _, view := range version.Views {
					for _, legality := range view.Legality {
						seen := byMove[legality.Name]
						if seen == nil {
							seen = &observed{ledgers: map[string]bool{}}
							byMove[legality.Name] = seen
						}
						raw := strings.TrimSpace(string(legality.Preconditions))
						if raw == "" || raw == "null" {
							seen.absent = true
							continue
						}
						var entries []ledgerEntry
						if err := json.Unmarshal(legality.Preconditions, &entries); err != nil {
							t.Fatalf("%s: %s: %v", game.name, path, err)
						}
						ledgerCount++
						keys := make([]string, len(entries))
						for i, entry := range entries {
							keys[i] = entryKey(entry.Name, entry.Args)
						}
						seen.ledgers[strings.Join(keys, " | ")] = true

						specs, declared, _, _ := manager.LegalPlanSpecs(legality.Name)
						if !declared {
							continue // reported below
						}
						for i, spec := range specs {
							if i >= len(entries) || spec.Message == "" || entries[i].Message == nil {
								continue
							}
							messageChecks++
							if entries[i].Message.Template != spec.Message {
								t.Errorf("%s %q: spec %d message %q, ledger template %q", game.name, legality.Name, i, spec.Message, entries[i].Message.Template)
							}
						}
					}
				}
			}
		}

		names := make([]string, 0, len(byMove))
		for name := range byMove {
			names = append(names, name)
		}
		sort.Strings(names)
		compared := 0
		for _, name := range names {
			seen := byMove[name]
			specs, declared, hasCustom, ok := manager.LegalPlanSpecs(name)
			if !ok {
				t.Errorf("%s: ledger move %q is unknown to LegalPlanSpecs", game.name, name)
				continue
			}
			keys := make([]string, 0, len(specs)+1)
			for _, spec := range specs {
				keys = append(keys, entryKey(spec.Name, spec.Args))
			}
			if hasCustom {
				keys = append(keys, "custom")
			}
			want := strings.Join(keys, " | ")
			emptyPlan := len(keys) == 0
			if seen.absent && declared && !emptyPlan {
				t.Errorf("%s %q: offered with no ledger, but LegalPlanSpecs has a plan: %s", game.name, name, want)
			}
			for ledger := range seen.ledgers {
				compared++
				if !declared {
					t.Errorf("%s %q: ledger %s recorded, but LegalPlanSpecs reports no plan", game.name, name, ledger)
					continue
				}
				if ledger != want {
					t.Errorf("%s %q:\n ledger %s\n plan   %s", game.name, name, ledger, want)
				}
			}
		}
		// Moves with a plan that no view ever offered (fix-ups, private
		// moves) cannot be compared.
		var unobserved []string
		for _, move := range manager.ExampleMoves() {
			name := move.Info().Name()
			if _, declared, _, _ := manager.LegalPlanSpecs(name); declared {
				if seen := byMove[name]; seen == nil || len(seen.ledgers) == 0 {
					unobserved = append(unobserved, name)
				}
			}
		}
		t.Logf("%s: %d documents, %d ledgers, %d distinct (move, ledger) pairs compared, %d message checks; declared plans with no recorded ledger: %q", game.name, len(paths), ledgerCount, compared, messageChecks, unobserved)
	}
}
