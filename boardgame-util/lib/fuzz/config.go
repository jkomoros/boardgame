// Package fuzz drives one game from one seed on the real engine, choosing
// among the actions the engine says are legal, and records the history plus
// every refused input it provoked on purpose. The history is exported with
// conformance.FromHistory, unchanged, so a fuzz document is replayed,
// captured and normalized exactly as the checked-in conformance corpus is;
// the refusals become the document's rejections.
//
// A run is a pure function of its Config: the same Config gives a
// byte-identical document, and a run cut short by a smaller MaxVersions is a
// prefix of the longer run. Runs are bounded by counts only. Nothing in this
// package reads a wall clock, sleeps, retries or polls; the engine's own
// channels (ProposeMoveAtVersion, ForceFixUp) and the synchronous
// ForceNextTimerWithError are the only synchronization.
//
// See fabric/TRANSPILE-PROPOSAL.md, "The differential fuzzer".
package fuzz

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
)

// Defaults for the zero values of Config.
const (
	DefaultMaxProposals = 300
	DefaultMaxVersions  = 600
	DefaultRefusalRate  = 1.0 / 8
	// MaxCandidatesPerMove caps the product of a move's field domains. Above
	// it, that many candidates are sampled with the run's generator.
	MaxCandidatesPerMove = 4096
)

// Policy names.
const (
	PolicyRandomLegal = "random-legal"
	PolicyAgent       = "agent"
	// PolicyBotPrefix is followed by a name from Bots().
	PolicyBotPrefix = "bot:"
)

// Config describes one run.
type Config struct {
	// Game is the game type's name, as its delegate's Name() returns it.
	Game string
	// Seed is the only source of variation. Everything the run chooses is
	// derived from it by SHA-256 with a prefix per use (see Derive).
	Seed int64
	// Policy is PolicyRandomLegal (the default), PolicyBotPrefix+name, or
	// PolicyAgent. The agent policy is not reproducible (memory's agent draws
	// from the global math/rand) and is only for transient runs.
	Policy string
	// MaxProposals bounds the outside inputs the driver sends, refused ones
	// included. Zero means DefaultMaxProposals.
	MaxProposals int
	// MaxVersions stops the run at the first settled version at or past it.
	// Zero means DefaultMaxVersions.
	MaxVersions int
	// RefusalRate is the probability, per step, that the driver sends an
	// input it expects the engine to refuse. Zero means DefaultRefusalRate;
	// a negative value means never.
	RefusalRate float64
}

func (c Config) withDefaults() Config {
	if c.Policy == "" {
		c.Policy = PolicyRandomLegal
	}
	if c.MaxProposals == 0 {
		c.MaxProposals = DefaultMaxProposals
	}
	if c.MaxVersions == 0 {
		c.MaxVersions = DefaultMaxVersions
	}
	if c.RefusalRate == 0 {
		c.RefusalRate = DefaultRefusalRate
	}
	if c.RefusalRate < 0 {
		c.RefusalRate = 0
	}
	return c
}

// Derived is everything a run takes from its seed before the game exists.
type Derived struct {
	ID         string
	SecretSalt string
	NumPlayers int
	Variant    boardgame.Variant
	// Seated is true when players join through the seating rendezvous. It
	// is only ever true for a game that configures a seat-player move.
	Seated bool
	// generatorSeed seeds the driver's own choices.
	generatorSeed int64
}

// seedHash is SHA-256 of a prefix naming the use, the game and the seed.
// This is scenario.Run's scheme ("boardgame scenario id:" + seed), with the
// fuzzer's own prefixes so a fuzz run never collides with a scenario.
func seedHash(use, game string, seed int64) [32]byte {
	return sha256.Sum256([]byte("boardgame fuzz " + use + ":" + game + ":" + strconv.FormatInt(seed, 10)))
}

func seedUint(use, game string, seed int64) uint64 {
	sum := seedHash(use, game, seed)
	return binary.LittleEndian.Uint64(sum[:8])
}

// Derive computes a run's identity from its seed. hasSeatMove says whether
// the game configures a seat-player move; without one a run is never
// seated.
func Derive(manager *boardgame.GameManager, seed int64, hasSeatMove bool) (*Derived, error) {
	delegate := manager.Delegate()
	game := delegate.Name()
	id := seedHash("id", game, seed)
	salt := seedHash("salt", game, seed)
	result := &Derived{
		ID:            strings.ToUpper(hex.EncodeToString(id[:8])),
		SecretSalt:    hex.EncodeToString(salt[:]),
		generatorSeed: int64(seedUint("generator", game, seed)),
	}
	low, high := delegate.MinNumPlayers(), delegate.MaxNumPlayers()
	if low < 1 || high < low {
		return nil, fmt.Errorf("game %s has player bounds %d..%d", game, low, high)
	}
	result.NumPlayers = low + int(seedUint("players", game, seed)%uint64(high-low+1))
	variant := boardgame.Variant{}
	config := manager.Variants()
	keys := make([]string, 0, len(config))
	for key := range config {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		values := make([]string, 0, len(config[key].Values))
		for value := range config[key].Values {
			values = append(values, value)
		}
		if len(values) == 0 {
			continue
		}
		sort.Strings(values)
		variant[key] = values[seedUint("variant "+key, game, seed)%uint64(len(values))]
	}
	if err := config.LegalVariant(variant); err != nil {
		return nil, fmt.Errorf("derived variant %v: %w", variant, err)
	}
	result.Variant = variant
	result.Seated = hasSeatMove && seedUint("seated", game, seed)%2 == 1
	return result, nil
}

func (d *Derived) generator() *rand.Rand {
	return rand.New(rand.NewSource(d.generatorSeed))
}

// ParseSeeds reads "N" or "A..B" (inclusive).
func ParseSeeds(text string) ([]int64, error) {
	if text == "" {
		return nil, fmt.Errorf("no seeds given")
	}
	parts := strings.Split(text, "..")
	if len(parts) > 2 {
		return nil, fmt.Errorf("seed range %q is not N or A..B", text)
	}
	first, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("seed %q: %w", parts[0], err)
	}
	last := first
	if len(parts) == 2 {
		if last, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
			return nil, fmt.Errorf("seed %q: %w", parts[1], err)
		}
	}
	if last < first {
		return nil, fmt.Errorf("seed range %q is empty", text)
	}
	if last-first >= 1000000 {
		return nil, fmt.Errorf("seed range %q is larger than a million", text)
	}
	seeds := make([]int64, 0, last-first+1)
	for seed := first; seed <= last; seed++ {
		seeds = append(seeds, seed)
	}
	return seeds, nil
}
