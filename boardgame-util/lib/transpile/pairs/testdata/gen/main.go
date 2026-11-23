// Command gen rebuilds the pair store
// (boardgame-util/lib/transpile/testdata/pairs/) from the six hand ports and
// fills each pair's measured fuzz fields. Run from the repository root:
//
//	BOARDGAME_NODE=<node 22.18+> ./scripts/go-local run ./boardgame-util/lib/transpile/pairs/testdata/gen [flags]
//
// With no flags it runs server/static/scripts/pairs-extract.mjs, writes
// every pair's source.go (the Go function at the manifest's hash),
// accepted.ts and meta.json, and index.ts, and removes hand pairs the
// extraction no longer has. Measured fields carry over while what they
// measured is unchanged; otherwise they are reset to null.
//
//	-check          write nothing; exit 1 if the store differs from what would be written
//	-coverage       measure Go block coverage of every pair's function over the
//	                checked-in fuzz corpus (a go build -cover of boardgame-util,
//	                `fuzz corpus --check`), and record it
//	-profile FILE   use this text coverage profile instead of building one
//	-mutants FILE   record the body-mutant counts in FILE (the JSON written by
//	                server/static/scripts/pairs-mutants.mjs --json FILE)
//	-not-mutated W  for every pair FILE has no entry for, record null mutant
//	                counts with W as the reason
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/pairs"
)

func main() {
	check := flag.Bool("check", false, "write nothing; exit 1 if the store is not current")
	coverage := flag.Bool("coverage", false, "measure and record Go block coverage")
	profile := flag.String("profile", "", "a text coverage profile to read instead of building one")
	mutants := flag.String("mutants", "", "the pairs-mutants.mjs JSON to record")
	notMutated := flag.String("not-mutated", "", "the reason recorded for pairs the mutant JSON has no entry for")
	flag.Parse()
	if err := run(*check, *coverage, *profile, *mutants, *notMutated); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(check, coverage bool, profile, mutants, notMutated string) error {
	root, err := pairs.ModuleRoot()
	if err != nil {
		return err
	}
	if coverage || profile != "" || mutants != "" {
		return measure(root, coverage, profile, mutants, notMutated)
	}
	node, err := pairs.Node()
	if err != nil {
		return err
	}
	ex, err := pairs.Extract(node, root)
	if err != nil {
		return err
	}
	manifests, err := pairs.Manifests(root, pairs.Games)
	if err != nil {
		return err
	}
	previous, err := pairs.Load(root)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	files, err := pairs.Build(root, ex, manifests, previous)
	if err != nil {
		return err
	}
	counts := map[string]int{}
	for _, p := range ex.Pairs {
		counts[p.Game]++
	}
	for _, game := range pairs.Games {
		fmt.Printf("%s: %d pairs\n", game, counts[game])
	}
	for _, n := range ex.None {
		fmt.Printf("no pair: %s (%s): %s\n", n.Func, n.Role, n.Reason)
	}
	if check {
		store := filepath.Join(root, filepath.FromSlash(pairs.StoreDir))
		stale := 0
		for rel, want := range files {
			got, err := os.ReadFile(filepath.Join(store, filepath.FromSlash(rel)))
			if err != nil || string(got) != string(want) {
				fmt.Println("stale: " + rel)
				stale++
			}
		}
		if stale > 0 {
			return fmt.Errorf("%d files of the store are not current", stale)
		}
		fmt.Println("The pair store is current")
		return nil
	}
	written, removed, err := pairs.WriteStore(root, files, previous)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d files, removed %d pairs\n", len(written), len(removed))
	return nil
}

// MutantResult is one pair's entry in pairs-mutants.mjs's JSON.
type MutantResult struct {
	Game      string   `json:"game"`
	Func      string   `json:"func"`
	Total     int      `json:"total"`
	Killed    int      `json:"killed"`
	Survivors []string `json:"survivors"`
	Accepted  string   `json:"acceptedSha256"`
}

func measure(root string, coverage bool, profile, mutants, notMutated string) error {
	stored, err := pairs.Load(root)
	if err != nil {
		return err
	}
	byGame := map[string][]pairs.Pair{}
	for _, p := range stored {
		byGame[p.Game] = append(byGame[p.Game], p)
	}
	if coverage || profile != "" {
		if profile == "" {
			work, err := os.MkdirTemp("", "pairs-coverage-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(work)
			if profile, err = fuzz.BuildBlockProfile(root, work, "", os.Stderr); err != nil {
				return err
			}
		}
		f, err := os.Open(profile)
		if err != nil {
			return err
		}
		blocks, err := fuzz.ParseProfile(f)
		f.Close()
		if err != nil {
			return err
		}
		manifests, err := pairs.Manifests(root, pairs.Games)
		if err != nil {
			return err
		}
		for _, game := range pairs.Games {
			spans, err := pairs.Spans(manifests[game])
			if err != nil {
				return err
			}
			measured := pairs.BlockCoverage(blocks, game, spans)
			for _, p := range byGame[game] {
				c, ok := measured[p.Meta.Func]
				if err := pairs.UpdateMeta(p, func(m *pairs.Meta) {
					if !ok || m.SourceHash != pairs.SourceHashOf(manifests[game], m.Func) {
						m.Fuzz.BlocksCovered, m.Fuzz.BlocksTotal, m.Fuzz.Unreached = nil, nil, nil
						return
					}
					covered, total := c.Covered, c.Total
					m.Fuzz.BlocksCovered, m.Fuzz.BlocksTotal, m.Fuzz.Unreached = &covered, &total, c.Unreached
				}); err != nil {
					return err
				}
				fmt.Printf("%s %s: %d of %d blocks\n", game, p.Meta.Func, c.Covered, c.Total)
			}
		}
		if mutants == "" {
			return nil
		}
		if stored, err = pairs.Load(root); err != nil {
			return err
		}
	}
	blob, err := os.ReadFile(mutants)
	if err != nil {
		return err
	}
	var results struct {
		Command string         `json:"command"`
		Pairs   []MutantResult `json:"pairs"`
	}
	if err := json.Unmarshal(blob, &results); err != nil {
		return err
	}
	byFunc := map[string]MutantResult{}
	for _, r := range results.Pairs {
		byFunc[r.Game+" "+r.Func] = r
	}
	var missing []string
	for _, p := range stored {
		r, ok := byFunc[p.Game+" "+p.Meta.Func]
		if ok && r.Accepted != pairs.SHA256(p.Accepted) {
			return fmt.Errorf("%s: the mutant run was of another accepted.ts; run it again", p.Meta.Func)
		}
		if !ok {
			missing = append(missing, p.Meta.Func)
			if notMutated == "" {
				continue
			}
		}
		if err := pairs.UpdateMeta(p, func(m *pairs.Meta) {
			if !ok {
				m.Fuzz.MutantsKilled, m.Fuzz.MutantsTotal, m.Fuzz.Survivors, m.Fuzz.Why = nil, nil, nil, notMutated
				return
			}
			killed, total := r.Killed, r.Total
			m.Fuzz.MutantsKilled, m.Fuzz.MutantsTotal, m.Fuzz.Survivors, m.Fuzz.Why = &killed, &total, r.Survivors, ""
		}); err != nil {
			return err
		}
	}
	sort.Strings(missing)
	fmt.Printf("recorded mutants for %d pairs; %d without an entry\n", len(stored)-len(missing), len(missing))
	return nil
}
