package legal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// TestExportConformanceFixtures serializes every named fixture in
// legalFixtureBuilders to testdata/conformance/fixtures/<name>.json so the
// TypeScript conformance suite (server/static/src/legal/*.conformance.test.ts)
// can reconstruct the exact same Context the Go evaluator sees and prove
// Go<->TS predicate equivalence against testdata/conformance/*.json.
//
// It is gated behind EXPORT_CONFORMANCE_FIXTURES=1 so a normal `go test
// ./legal/...` run does not rewrite committed testdata; regenerate with
//
//	EXPORT_CONFORMANCE_FIXTURES=1 go test ./legal/ -run TestExportConformanceFixtures
//
// The state is serialized via state.StorageRecord() — the canonical
// {Version, Game, Players, Components} JSON the server already sends the
// client (matching server/static/src/types/game-state.d.ts RawGameState), so
// the TS resolver runs the SAME representation in the test and in production.
// currentPlayerIndex is the delegate-resolved index (not a state field, so
// not recoverable from RawGameState alone — see design C5). The move's
// property values are exported flat for move.* path resolution.
func TestExportConformanceFixtures(t *testing.T) {
	if os.Getenv("EXPORT_CONFORMANCE_FIXTURES") != "1" {
		t.Skip("set EXPORT_CONFORMANCE_FIXTURES=1 to regenerate the TS conformance fixtures")
	}

	outDir := filepath.Join("testdata", "conformance", "fixtures")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", outDir, err)
	}

	names := make([]string, 0, len(legalFixtureBuilders))
	for name := range legalFixtureBuilders {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fx := legalFixtureBuilders[name](t)

		exported := exportedFixture{
			Fixture:            name,
			CurrentPlayerIndex: int(fx.state.CurrentPlayerIndex()),
			State:              json.RawMessage(fx.state.StorageRecord()),
			Move:               exportMoveProps(t, fx.move),
			Chest:              exportChest(t, fx.chest),
		}

		b, err := json.MarshalIndent(exported, "", "  ")
		if err != nil {
			t.Fatalf("marshal fixture %q: %v", name, err)
		}
		b = append(b, '\n')
		if err := os.WriteFile(filepath.Join(outDir, name+".json"), b, 0644); err != nil {
			t.Fatalf("write fixture %q: %v", name, err)
		}
	}
	t.Logf("exported %d conformance fixtures to %s", len(names), outDir)
}

type exportedFixture struct {
	Fixture            string                 `json:"fixture"`
	CurrentPlayerIndex int                    `json:"currentPlayerIndex"`
	State              json.RawMessage        `json:"state"`
	Move               map[string]interface{} `json:"move"`
	Chest              json.RawMessage        `json:"chest"`
}

// exportChest serializes the fixture's ComponentChest exactly as it ships on
// /info (server/api/main.go: gin.H "Chest" = game.Manager().Chest()), so the TS
// evaluator runs the SAME chest representation in the conformance test and in
// production. Enum/component predicates (propEquals's enum arm,
// componentPresentAtKey, componentPropEqualsCurrentPlayer) need it: the state
// serializes an enum prop as its value NAME with no enum identity, and a deck
// component's immutable Values (e.g. checkers token Color) live only here, not
// in the state's dynamic Components.
func exportChest(t *testing.T, chest *boardgame.ComponentChest) json.RawMessage {
	if chest == nil {
		return json.RawMessage("null")
	}
	b, err := json.Marshal(chest)
	if err != nil {
		t.Fatalf("marshal chest: %v", err)
	}
	return json.RawMessage(b)
}

// exportMoveProps flattens a move's readable properties into a JSON object for
// move.* path resolution (nil move -> null, exercising the "no move" Unknown
// path). Values are exported in the same shape the TS resolver expects: ints
// as numbers, bools as bools, strings as strings, enums as their string
// value, PlayerIndex as its integer.
func exportMoveProps(t *testing.T, move boardgame.Move) map[string]interface{} {
	if move == nil {
		return nil
	}
	reader := move.Reader()
	out := map[string]interface{}{}
	for propName, propType := range reader.Props() {
		switch propType {
		case boardgame.TypeInt:
			v, err := reader.IntProp(propName)
			if err == nil {
				out[propName] = v
			}
		case boardgame.TypeBool:
			v, err := reader.BoolProp(propName)
			if err == nil {
				out[propName] = v
			}
		case boardgame.TypeString:
			v, err := reader.StringProp(propName)
			if err == nil {
				out[propName] = v
			}
		case boardgame.TypePlayerIndex:
			v, err := reader.PlayerIndexProp(propName)
			if err == nil {
				out[propName] = int(v)
			}
		case boardgame.TypeEnum:
			v, err := reader.ImmutableEnumProp(propName)
			if err == nil && v != nil {
				out[propName] = enumValueString(v)
			}
		default:
			// Other property types (stacks, timers, etc.) are not addressable
			// by a move.* legal path; skip them.
		}
	}
	return out
}

func enumValueString(v enum.ImmutableVal) string {
	return v.String()
}
