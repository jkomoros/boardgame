package fuzz

import (
	"fmt"
	"sort"

	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
)

// Name is a fuzz document's name: "<game>/seed-NNNN".
func Name(game string, seed int64) string {
	return fmt.Sprintf("%s/seed-%04d", game, seed)
}

// Document runs cfg and exports it: conformance.FromHistory, unchanged,
// over the run's history, then the run's refusals as rejections. A document
// with any rejection has conformance.SchemaVersionWithRejections.
func Document(cfg Config) (*conformance.Document, *Result, error) {
	newDelegate := Delegate(cfg.Game)
	if newDelegate == nil {
		return nil, nil, fmt.Errorf("unknown game %q; known: %v", cfg.Game, Games())
	}
	result, err := RunWithDelegate(newDelegate, cfg)
	if err != nil {
		return nil, nil, err
	}
	doc, err := DocumentFor(newDelegate, result)
	return doc, result, err
}

// DocumentFor exports a finished run.
func DocumentFor(newDelegate conformance.NewDelegate, result *Result) (*conformance.Document, error) {
	cfg := result.Config
	source := conformance.Source{
		// conformance/schema.ts allows only "scenario" or "golden"; a fuzz
		// run is a generated scenario, and Description says which.
		Kind: "scenario",
		Description: fmt.Sprintf("fuzz: game %s, seed %d, policy %s, max proposals %d, max versions %d, refusal rate %g",
			cfg.Game, cfg.Seed, cfg.Policy, cfg.MaxProposals, cfg.MaxVersions, cfg.RefusalRate),
		Covers: fmt.Sprintf("A seeded run of the fuzz driver (boardgame-util/lib/fuzz): %d players, variant %v, seated %t; ended: %s after %d proposals; %d refused inputs.",
			result.Derived.NumPlayers, sortedVariant(result.Derived.Variant), result.Derived.Seated, result.End, result.Proposals, len(result.Refusals)),
	}
	doc, err := conformance.FromHistory(newDelegate, result.History, Name(cfg.Game, cfg.Seed), source)
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", Name(cfg.Game, cfg.Seed), err)
	}
	if doc.Game.Seating != result.Derived.Seated {
		return nil, fmt.Errorf("export %s: the run was seated=%t but the exporter replayed seated=%t", Name(cfg.Game, cfg.Seed), result.Derived.Seated, doc.Game.Seating)
	}
	if len(result.Refusals) > 0 {
		doc.SchemaVersion = conformance.SchemaVersionWithRejections
		doc.Rejections = make([]conformance.Rejection, len(result.Refusals))
		for i, refusal := range result.Refusals {
			doc.Rejections[i] = refusal.Rejection
		}
	}
	return doc, nil
}

func sortedVariant(variant map[string]string) []string {
	result := make([]string, 0, len(variant))
	for key, value := range variant {
		result = append(result, key+"="+value)
	}
	sort.Strings(result)
	return result
}
