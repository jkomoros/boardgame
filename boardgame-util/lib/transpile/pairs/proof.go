package pairs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Replay is the TypeScript half's answer for one game or one substitution
// (server/static/scripts/pairs-proof.mjs).
type Replay struct {
	Game       string   `json:"game"`
	Func       string   `json:"func,omitempty"`
	Passed     bool     `json:"passed"`
	Documents  int      `json:"documents"`
	Versions   int      `json:"versions"`
	Rejections int      `json:"rejections"`
	Failure    *Failure `json:"failure"`
	Served     []string `json:"served"`
	Fallbacks  []string `json:"fallbacks"`
}

// Failure is the first document that failed and what differed in it,
// including the fuzz runner's divergence report.
type Failure struct {
	File     *string  `json:"file"`
	Problems []string `json:"problems"`
}

func (f *Failure) String() string {
	if f == nil {
		return ""
	}
	file := "(no document)"
	if f.File != nil {
		file = *f.File
	}
	return file + ": " + strings.Join(f.Problems, "\n    ")
}

// Options configure a proof run.
type Options struct {
	// Node is the Node binary (Node()).
	Node string
	// Rewrite edits the translator's output before it is substituted.
	Rewrite Rewrite
	// WorkDir holds the translated modules; it must exist.
	WorkDir string
}

type substitution struct {
	Game string `json:"game"`
	Func string `json:"func"`
	File string `json:"file"`
	Name string `json:"name"`
}

type replayJob struct {
	Baseline      []string       `json:"baseline"`
	Substitutions []substitution `json:"substitutions"`
}

type replayResult struct {
	Baseline      []Replay `json:"baseline"`
	Substitutions []Replay `json:"substitutions"`
}

func runReplay(root string, opts Options, job replayJob) (*replayResult, error) {
	blob, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
	out, err := RunScript(opts.Node, root, "pairs-proof.mjs", blob)
	if err != nil {
		return nil, err
	}
	var result replayResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("pairs-proof.mjs output: %w", err)
	}
	return &result, nil
}

// PairResult is one stored pair in a substitution proof.
type PairResult struct {
	Pair        Pair
	Translation Translation
	// Helper pairs are never resolved by the loader; their translation is
	// exercised inside every substituted caller's module.
	Helper bool
	// Replay is the corpus run with the translation substituted (nil for a
	// helper or a refused function).
	Replay *Replay
	// TextuallyEqual: the translator's text is the pair's, whitespace aside.
	TextuallyEqual bool
}

// GameProof is the general substitution proof for one game.
type GameProof struct {
	Game     string
	Baseline Replay
	Results  []PairResult
}

// Counts summarizes a game's proof.
func (g GameProof) Counts() (pairs, translated, substituted, passing, equal, failing, helpers int) {
	for _, r := range g.Results {
		pairs++
		if r.Translation.Accepted {
			translated++
		}
		if r.Helper && r.Translation.Accepted {
			helpers++
		}
		if r.TextuallyEqual {
			equal++
		}
		if r.Replay != nil {
			substituted++
			if r.Replay.Passed {
				passing++
			} else {
				failing++
			}
		}
	}
	return
}

// Substitute is the general form of the rule proof: for every stored pair of
// the games whose Go source the translator accepts, the game built from the
// stored pairs with that one function's translation in place of its
// accepted.ts must pass the game's whole corpus. Games run in parallel, one
// Node process each.
func Substitute(root string, stored []Pair, games []string, opts Options) ([]GameProof, error) {
	byGame := map[string][]Pair{}
	for _, p := range stored {
		byGame[p.Game] = append(byGame[p.Game], p)
	}
	proofs := make([]GameProof, len(games))
	errs := make([]error, len(games))
	var wg sync.WaitGroup
	for i, game := range games {
		wg.Add(1)
		go func(i int, game string) {
			defer wg.Done()
			proofs[i], errs[i] = substituteGame(root, game, byGame[game], opts, true)
		}(i, game)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return proofs, nil
}

// SubstituteGame is Substitute for some pairs of one game, without the
// baseline replay.
func SubstituteGame(root, game string, pairs []Pair, opts Options) (GameProof, error) {
	return substituteGame(root, game, pairs, opts, false)
}

func substituteGame(root, game string, pairs []Pair, opts Options, baseline bool) (GameProof, error) {
	proof := GameProof{Game: game}
	refs := make([]string, len(pairs))
	for i, p := range pairs {
		refs[i] = p.Meta.Func
	}
	translations, err := Translate(root, game, refs, opts.Rewrite)
	if err != nil {
		return proof, err
	}
	dir := filepath.Join(opts.WorkDir, game)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return proof, err
	}
	job := replayJob{}
	if baseline {
		job.Baseline = []string{game}
	}
	index := map[string]int{}
	for _, p := range pairs {
		t := translations[p.Meta.Func]
		r := PairResult{Pair: p, Translation: t, Helper: p.Meta.Role == "helper"}
		r.TextuallyEqual = t.Accepted && TextuallyEqual(t.Code, p.Accepted, p.Name)
		if t.Accepted && !r.Helper {
			file := filepath.Join(dir, p.Name+".ts")
			if err := os.WriteFile(file, []byte(t.Module), 0644); err != nil {
				return proof, err
			}
			index[p.Meta.Func] = len(proof.Results)
			job.Substitutions = append(job.Substitutions, substitution{Game: game, Func: p.Meta.Func, File: file, Name: p.Name})
		}
		proof.Results = append(proof.Results, r)
	}
	result, err := runReplay(root, opts, job)
	if err != nil {
		return proof, err
	}
	if baseline {
		if len(result.Baseline) != 1 {
			return proof, fmt.Errorf("%s: no baseline replay", game)
		}
		proof.Baseline = result.Baseline[0]
	}
	for i := range result.Substitutions {
		replay := result.Substitutions[i]
		at, ok := index[replay.Func]
		if !ok {
			return proof, fmt.Errorf("%s: a replay for %s, which was not substituted", game, replay.Func)
		}
		proof.Results[at].Replay = &replay
	}
	for ref, at := range index {
		if proof.Results[at].Replay == nil {
			return proof, fmt.Errorf("%s: no replay came back for %s", game, ref)
		}
	}
	return proof, nil
}

// RulePair is one stored pair that carried the refusal a rule is about.
type RulePair struct {
	Pair Pair
	// Accepted: the translator now accepts the Go source (condition 1).
	Accepted bool
	Refusals []string
	// Replays are the corpus runs with the translation substituted
	// (condition 2): the pair's own, or for a helper every caller's.
	Replays []Replay
	// CoverageComplete is condition 3; a pair without it does not count.
	CoverageComplete bool
	TextuallyEqual   bool
}

// Passed reports conditions 1 and 2.
func (r RulePair) Passed() bool {
	if !r.Accepted || len(r.Replays) == 0 {
		return false
	}
	for _, replay := range r.Replays {
		if !replay.Passed {
			return false
		}
	}
	return true
}

// RuleProof is ProveRule's answer.
type RuleProof struct {
	Reason, Detail string
	Pairs          []RulePair
}

// Proven reports whether every pair that carried the refusal passes
// conditions 1 and 2 and at least one of them has complete coverage.
func (p RuleProof) Proven() bool {
	counting := 0
	for _, r := range p.Pairs {
		if !r.Passed() {
			return false
		}
		if r.CoverageComplete {
			counting++
		}
	}
	return counting > 0
}

// StillRefused lists the pairs the translator still refuses.
func (p RuleProof) StillRefused() []RulePair {
	var out []RulePair
	for _, r := range p.Pairs {
		if !r.Accepted {
			out = append(out, r)
		}
	}
	return out
}

// RefusalsCarried lists every (reason, detail) some stored pair carried,
// sorted, with the pairs carrying each.
func RefusalsCarried(stored []Pair) map[[2]string][]Pair {
	out := map[[2]string][]Pair{}
	for _, p := range stored {
		seen := map[[2]string]bool{}
		for _, r := range p.Meta.Refusals {
			key := [2]string{r.Reason, r.Detail}
			if !seen[key] {
				seen[key] = true
				out[key] = append(out[key], p)
			}
		}
	}
	return out
}

// ProveRule checks a proposed rule for refusal reason (and detail, when not
// empty) against every stored pair that carried it: (1) the translator now
// accepts the Go source; (2) with the translator's output substituted for
// accepted.ts the game's conformance and fuzz corpora pass; (3) the pair's
// recorded block coverage is complete (else it does not count). A helper's
// translation is substituted through every stored caller that reaches it.
func ProveRule(root string, stored []Pair, reason, detail string, opts Options) (RuleProof, error) {
	proof := RuleProof{Reason: reason, Detail: detail}
	byGame := map[string][]Pair{}
	for _, p := range stored {
		byGame[p.Game] = append(byGame[p.Game], p)
	}
	carried := map[string][]Pair{}
	for _, p := range stored {
		for _, r := range p.Meta.Refusals {
			if r.Reason == reason && (detail == "" || r.Detail == detail) {
				carried[p.Game] = append(carried[p.Game], p)
				break
			}
		}
	}
	games := make([]string, 0, len(carried))
	for game := range carried {
		games = append(games, game)
	}
	sort.Strings(games)
	for _, game := range games {
		all := byGame[game]
		refs := make([]string, len(all))
		for i, p := range all {
			refs[i] = p.Meta.Func
		}
		translations, err := Translate(root, game, refs, opts.Rewrite)
		if err != nil {
			return proof, err
		}
		var subjects []Pair
		for _, p := range carried[game] {
			t := translations[p.Meta.Func]
			r := RulePair{Pair: p, Accepted: t.Accepted, CoverageComplete: p.Meta.Fuzz.CoverageComplete()}
			for _, refusal := range t.Refusals {
				r.Refusals = append(r.Refusals, refusal.Reason+" "+refusal.Detail)
			}
			if t.Accepted {
				r.TextuallyEqual = TextuallyEqual(t.Code, p.Accepted, p.Name)
				subjects = append(subjects, p)
			}
			proof.Pairs = append(proof.Pairs, r)
		}
		if len(subjects) == 0 {
			continue
		}
		// Which pairs to substitute: each accepted subject itself, or for a
		// helper, every non-helper pair whose translation bundles it.
		need := map[string][]string{}
		for _, p := range subjects {
			if p.Meta.Role != "helper" {
				need[p.Meta.Func] = append(need[p.Meta.Func], p.Meta.Func)
				continue
			}
			for _, caller := range all {
				if caller.Meta.Role == "helper" {
					continue
				}
				for _, callee := range translations[caller.Meta.Func].Callees {
					if callee == p.Meta.Func {
						need[caller.Meta.Func] = append(need[caller.Meta.Func], p.Meta.Func)
					}
				}
			}
		}
		var substituted []Pair
		for _, p := range all {
			if len(need[p.Meta.Func]) > 0 {
				substituted = append(substituted, p)
			}
		}
		sub, err := substituteGame(root, game, substituted, opts, false)
		if err != nil {
			return proof, err
		}
		for i := range proof.Pairs {
			r := &proof.Pairs[i]
			if r.Pair.Game != game || !r.Accepted {
				continue
			}
			for _, result := range sub.Results {
				if result.Replay == nil {
					continue
				}
				for _, subject := range need[result.Pair.Meta.Func] {
					if subject == r.Pair.Meta.Func {
						r.Replays = append(r.Replays, *result.Replay)
					}
				}
			}
		}
	}
	return proof, nil
}
