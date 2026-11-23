// Package pairs is the store of known answers the translator is measured
// against (fabric/TRANSPILE-PROPOSAL.md, "The ratchet and its metrics",
// "Stored pairs"), and the proof that a translator rule is right
// ("Turning a refusal reason into a rule").
//
// The store is boardgame-util/lib/transpile/testdata/pairs/:
//
//	<game>/<ts function name>/
//	  source.go     the Go function, verbatim, at the hash in meta.json
//	  accepted.ts   the accepted TypeScript: one exported function, named
//	                <ts function name> (transpile.Mangle), with the role's
//	                calling convention (server/static/src/engine/rules/data.ts)
//	  meta.json     Meta
//	index.ts        imports every accepted.ts; generated with the store
//	tsconfig.json   type-checks the store under the engine's sandbox options
//
// The six hand ports seed it (producer "hand"), split by
// server/static/scripts/pairs-extract.mjs and written by
// testdata/gen (`./scripts/go-local run ./boardgame-util/lib/transpile/pairs/testdata/gen`).
//
// A hand pair whose Go source has changed since it was stored is stale: it is
// reported, never a failure, because the hand port is not regenerated from
// Go and its pair says what the hand port does, not what the Go now does.
// Any other producer's stale pair is a failure.
package pairs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// StoreDir is the store, relative to the module root.
const StoreDir = "boardgame-util/lib/transpile/testdata/pairs"

// The files of one pair.
const (
	SourceFile   = "source.go"
	AcceptedFile = "accepted.ts"
	MetaFile     = "meta.json"
)

// ProducerHand is a pair split from a hand port.
const ProducerHand = "hand"

// The extraction kinds of a hand pair (pairs-extract.mjs explains each).
const (
	ExtractionMechanical = "mechanical"
	ExtractionWrapped    = "wrapped"
	ExtractionNone       = "none"
)

// Games are the six hand-ported games, in the order the store lists them.
var Games = []string{"memory", "blackjack", "tictactoe", "pig", "debuganimations", "werewolf"}

// Meta is a pair's meta.json.
type Meta struct {
	Func       string `json:"func"`
	Role       string `json:"role"`
	SourceHash string `json:"sourceHash"`
	Producer   string `json:"producer"`
	// Refusals are the checker's refusals of the Go source when the pair was
	// stored ([] when the translator accepted it). A rule for one of these
	// is proven against this pair (ProveRule).
	Refusals []transpile.Refusal `json:"refusals"`
	// Extraction and Note say how a hand pair's accepted.ts was cut from the
	// hand port.
	Extraction string `json:"extraction,omitempty"`
	Note       string `json:"note,omitempty"`
	Fuzz       Fuzz   `json:"fuzz"`
}

// Fuzz is what the checked-in corpus says about a pair. A null count was not
// measured; Why says why.
type Fuzz struct {
	// Seeds are the game's checked-in fuzz documents (conformance/fuzz/<game>/).
	Seeds []int `json:"seeds"`
	// BlocksCovered and BlocksTotal are Go block coverage of the function
	// over those documents (`boardgame-util fuzz coverage --blocks`).
	BlocksCovered *int `json:"blocksCovered"`
	BlocksTotal   *int `json:"blocksTotal"`
	// Unreached lists each unreached block as "line.col-line.col" in the
	// game's Go file.
	Unreached []string `json:"unreached,omitempty"`
	// MutantsKilled and MutantsTotal are body mutants of accepted.ts
	// (server/static/scripts/pairs-mutants.mjs) and how many the corpus kills.
	MutantsKilled *int `json:"mutantsKilled"`
	MutantsTotal  *int `json:"mutantsTotal"`
	// Survivors describe the mutants the corpus did not kill.
	Survivors []string `json:"survivors,omitempty"`
	// Why explains a null count.
	Why string `json:"why,omitempty"`
}

// CoverageComplete reports whether every Go block of the function was
// reached: the third condition of a rule proof.
func (f Fuzz) CoverageComplete() bool {
	return f.BlocksCovered != nil && f.BlocksTotal != nil && *f.BlocksTotal > 0 && *f.BlocksCovered == *f.BlocksTotal
}

// Pair is one stored pair.
type Pair struct {
	Game string
	// Name is the directory name: the mangled TypeScript name.
	Name     string
	Dir      string
	Meta     Meta
	Source   []byte
	Accepted []byte
}

// Load reads every pair under the store at root (the module root).
func Load(root string) ([]Pair, error) {
	store := filepath.Join(root, filepath.FromSlash(StoreDir))
	games, err := os.ReadDir(store)
	if err != nil {
		return nil, err
	}
	var out []Pair
	for _, game := range games {
		if !game.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(store, game.Name()))
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			p, err := loadPair(filepath.Join(store, game.Name(), entry.Name()))
			if err != nil {
				return nil, err
			}
			p.Game = game.Name()
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Game != out[j].Game {
			return out[i].Game < out[j].Game
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func loadPair(dir string) (Pair, error) {
	p := Pair{Name: filepath.Base(dir), Dir: dir}
	var err error
	if p.Source, err = os.ReadFile(filepath.Join(dir, SourceFile)); err != nil {
		return p, err
	}
	if p.Accepted, err = os.ReadFile(filepath.Join(dir, AcceptedFile)); err != nil {
		return p, err
	}
	blob, err := os.ReadFile(filepath.Join(dir, MetaFile))
	if err != nil {
		return p, err
	}
	decoder := json.NewDecoder(bytes.NewReader(blob))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p.Meta); err != nil {
		return p, fmt.Errorf("%s: %w", filepath.Join(dir, MetaFile), err)
	}
	return p, nil
}

// Validate checks what a pair says about itself: its directory is the
// mangled name of its function, source.go hashes to meta.sourceHash,
// accepted.ts exports that one name, and its fields are well formed.
func (p Pair) Validate() error {
	name, err := transpile.Mangle(p.Meta.Func)
	if err != nil {
		return err
	}
	if name != p.Name {
		return fmt.Errorf("%s/%s: directory is not Mangle(%s) = %s", p.Game, p.Name, p.Meta.Func, name)
	}
	if !strings.HasPrefix(p.Meta.Func, p.Game+".") {
		return fmt.Errorf("%s/%s: func %s is not in package %s", p.Game, p.Name, p.Meta.Func, p.Game)
	}
	if got := gamedef.SourceHash(p.Source); got != p.Meta.SourceHash {
		return fmt.Errorf("%s/%s: source.go hashes to %s, meta.json says %s", p.Game, p.Name, got, p.Meta.SourceHash)
	}
	if p.Meta.Producer == "" || p.Meta.Role == "" {
		return fmt.Errorf("%s/%s: producer and role are required", p.Game, p.Name)
	}
	if p.Meta.Producer == ProducerHand {
		switch p.Meta.Extraction {
		case ExtractionMechanical, ExtractionWrapped, ExtractionNone:
		default:
			return fmt.Errorf("%s/%s: a hand pair's extraction is mechanical, wrapped or none, not %q", p.Game, p.Name, p.Meta.Extraction)
		}
		if p.Meta.Extraction != ExtractionMechanical && p.Meta.Note == "" {
			return fmt.Errorf("%s/%s: a %s extraction needs a note", p.Game, p.Name, p.Meta.Extraction)
		}
	}
	if p.Meta.Refusals == nil {
		return fmt.Errorf("%s/%s: refusals must be a list ([] for none)", p.Game, p.Name)
	}
	exports := exportedFunctions(p.Accepted)
	if len(exports) != 1 || exports[0] != p.Name {
		return fmt.Errorf("%s/%s: accepted.ts must export exactly the function %s; it exports %v", p.Game, p.Name, p.Name, exports)
	}
	return nil
}

// exportedFunctions lists the names of `export function NAME(` at the start
// of a line.
func exportedFunctions(text []byte) []string {
	var out []string
	for _, line := range strings.Split(string(text), "\n") {
		if !strings.HasPrefix(line, "export ") {
			continue
		}
		rest := strings.TrimPrefix(line, "export function ")
		if rest == line {
			out = append(out, "<"+strings.Fields(line)[1]+">")
			continue
		}
		if i := strings.IndexAny(rest, "(<"); i > 0 {
			out = append(out, rest[:i])
		}
	}
	return out
}

// Signature is the parameter types and result type of the exported
// function of a TypeScript text, with parameter names dropped:
// "(RulesState, number): void".
func Signature(text, name string) (string, error) {
	for _, line := range strings.Split(text, "\n") {
		prefix := "export function " + name + "("
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := line[len(prefix):]
		depth := 1
		end := -1
	scan:
		for i := 0; i < len(rest); i++ {
			switch rest[i] {
			case '(', '[', '{', '<':
				depth++
			case ')', ']', '}':
				depth--
			case '>':
				if i == 0 || rest[i-1] != '=' {
					depth--
				}
			}
			if depth == 0 {
				end = i
				break scan
			}
		}
		if end < 0 {
			return "", fmt.Errorf("unterminated parameter list of %s", name)
		}
		params := splitTop(rest[:end])
		types := make([]string, 0, len(params))
		for _, param := range params {
			colon := strings.Index(param, ":")
			if colon < 0 {
				return "", fmt.Errorf("parameter %q of %s has no type", param, name)
			}
			types = append(types, strings.TrimSpace(param[colon+1:]))
		}
		result := strings.TrimSpace(rest[end+1:])
		result = strings.TrimPrefix(result, ":")
		result = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(result), "{"))
		return "(" + strings.Join(types, ", ") + "): " + result, nil
	}
	return "", fmt.Errorf("no `export function %s(` line", name)
}

// splitTop splits a parameter list at the commas outside brackets.
func splitTop(s string) []string {
	var out []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '(', '[', '{', '<':
			depth++
		case ')', ']', '}':
			depth--
		case '>':
			if i > 0 && s[i-1] != '=' {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if tail := strings.TrimSpace(s[start:]); tail != "" {
		out = append(out, tail)
	}
	return out
}

// Staleness is one pair whose Go function no longer is what was stored.
type Staleness struct {
	Pair Pair
	// Why is "source changed (sha256:... now)" or "no longer in the manifest".
	Why string
}

// Stale compares each pair with its game's current manifest.
func Stale(pairs []Pair, manifests map[string]*gamedef.StaticReport) []Staleness {
	var out []Staleness
	for _, p := range pairs {
		report := manifests[p.Game]
		if report == nil {
			continue
		}
		var entry *gamedef.FunctionEntry
		for i := range report.Functions {
			if report.Functions[i].Func == p.Meta.Func {
				entry = &report.Functions[i]
			}
		}
		switch {
		case entry == nil:
			out = append(out, Staleness{p, "no longer in the manifest"})
		case entry.SourceHash != p.Meta.SourceHash:
			out = append(out, Staleness{p, "source changed (" + entry.SourceHash + " now)"})
		case entry.Role != p.Meta.Role:
			out = append(out, Staleness{p, "role changed (" + entry.Role + " now)"})
		}
	}
	return out
}

// SourceHashOf is the manifest's hash of a function, or "" when the
// manifest has no such function.
func SourceHashOf(report *gamedef.StaticReport, fn string) string {
	for _, entry := range report.Functions {
		if entry.Func == fn {
			return entry.SourceHash
		}
	}
	return ""
}

// SHA256 is the lower-case hex SHA-256 of b.
func SHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ModuleRoot is the module root above the working directory.
func ModuleRoot() (string, error) {
	return transpile.ModuleRoot(".")
}
