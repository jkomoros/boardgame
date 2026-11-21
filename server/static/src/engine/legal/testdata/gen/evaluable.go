package main

import (
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame"
	memorygame "github.com/jkomoros/boardgame/examples/memory"
)

/*
Evaluability vectors: boardgame.LegalReadEvaluable, the public face of the
unexported facetSurvives table, asked about real states under real
sanitization.

The policy a property is sanitized with comes from the delegate, so a wrapped
memory delegate (synthDelegate) is given each policy in turn. That yields the
full policy x facet table from Go itself rather than from a transcription of
it, plus the Visibility block the server would send each viewer - which is
what a client actually has to work from.
*/

type exportedFacetRow struct {
	Policy string `json:"policy"`
	// Survives maps each facet to LegalReadEvaluable's answer for a property
	// under Policy.
	Survives map[string]bool `json:"survives"`
	// Visibility is the facet list the viewer's state JSON carries for that
	// property.
	Visibility []string `json:"visibility"`
}

type exportedReadCase struct {
	Config             string          `json:"config"`
	State              string          `json:"state"`
	Viewer             int             `json:"viewer"`
	CurrentPlayerIndex int             `json:"currentPlayerIndex"`
	Visibility         json.RawMessage `json:"visibility"`
	// Reads is [path, facet, evaluable].
	Reads [][]interface{} `json:"reads"`
}

type evaluableFile struct {
	fixtures   map[string]exportedFixture
	games      map[string]exportedGame
	facetTable []exportedFacetRow
	readCases  []exportedReadCase
}

func (e evaluableFile) sections() []section {
	return []section{
		{key: "comment", scalar: generatedComment},
		{key: "facetTable", list: anyList(e.facetTable)},
		{key: "readCases", list: anyList(e.readCases)},
	}
}

var allPolicies = []struct {
	name   string
	policy boardgame.Policy
}{
	{"visible", boardgame.PolicyVisible},
	{"order", boardgame.PolicyOrder},
	{"len", boardgame.PolicyLen},
	{"nonempty", boardgame.PolicyNonEmpty},
	{"hidden", boardgame.PolicyHidden},
}

var allFacets = []boardgame.LegalFacet{
	boardgame.LegalFacetValues, boardgame.LegalFacetCount, boardgame.LegalFacetOccupancy,
	boardgame.LegalFacetOrder, boardgame.LegalFacetNonEmpty,
}

func constantPolicy(p boardgame.Policy) policyOverride {
	return func(map[string]bool) boardgame.Policy { return p }
}

func buildEvaluable() evaluableFile {
	out := evaluableFile{}

	for _, p := range allPolicies {
		game, state := newGame("evaluable", &synthDelegate{
			GameDelegate: memorygame.NewDelegate(),
			policies:     map[string]policyOverride{"game.HiddenCards": constantPolicy(p.policy)},
		})
		row := exportedFacetRow{Policy: p.name, Survives: map[string]bool{}}
		for _, facet := range allFacets {
			row.Survives[facetName(facet)] = boardgame.LegalReadEvaluable(state, 1, boardgame.LegalRead{Path: "game.HiddenCards", Facet: facet})
		}
		var visibility struct {
			Game map[string][]string
		}
		must(json.Unmarshal(viewVisibility(game, state, 1), &visibility), "decoding visibility")
		row.Visibility = visibility.Game["HiddenCards"]
		out.facetTable = append(out.facetTable, row)
	}

	selfOnly := func(groups map[string]bool) boardgame.Policy {
		if groups["self"] {
			return boardgame.PolicyVisible
		}
		return boardgame.PolicyHidden
	}
	configs := []struct {
		name     string
		policies map[string]policyOverride
	}{
		{"asShipped", nil},
		{"mixed", map[string]policyOverride{
			"game.HiddenCards":         constantPolicy(boardgame.PolicyOrder),
			"game.VisibleCards":        constantPolicy(boardgame.PolicyLen),
			"game.UnusedCards":         constantPolicy(boardgame.PolicyNonEmpty),
			"game.NumCards":            constantPolicy(boardgame.PolicyHidden),
			"player.CardsLeftToReveal": selfOnly,
			"player.WonCards":          constantPolicy(boardgame.PolicyLen),
		}},
		{"playersHidden", map[string]policyOverride{
			"player.CardsLeftToReveal": constantPolicy(boardgame.PolicyHidden),
			"player.SeatFilled":        selfOnly,
		}},
	}
	states := []struct {
		name   string
		mutate func(state boardgame.State)
	}{
		{"playerZeroTurn", func(boardgame.State) {}},
		{"playerOneTurn", func(state boardgame.State) {
			must(state.GameState().ReadSetter().SetPlayerIndexProp("CurrentPlayer", 1), "setting CurrentPlayer")
		}},
	}
	paths := []string{
		"game.HiddenCards", "game.VisibleCards", "game.UnusedCards", "game.NumCards", "game.CurrentPlayer", "game.NoSuchProp",
		"player.CardsLeftToReveal", "player.WonCards", "player.SeatFilled", "player.NoSuchProp",
		"proposer.CardsLeftToReveal", "proposer.WonCards", "proposer.SeatFilled",
		"players[*].CardsLeftToReveal", "players[*].WonCards", "players[*].SeatFilled",
		"players[move.TargetPlayerIndex].CardsLeftToReveal", "players[move.TargetPlayerIndex].SeatFilled",
		"move.CardIndex", "move.NoSuchField",
		"nodothere", "nokind.X", "game.", "players[0].X",
	}
	for _, config := range configs {
		for _, st := range states {
			game, state := newGame("evaluable", &synthDelegate{GameDelegate: memorygame.NewDelegate(), policies: config.policies})
			st.mutate(state)
			for _, viewer := range []int{0, 1, -1, -2} {
				rc := exportedReadCase{
					Config: config.name, State: st.name, Viewer: viewer,
					CurrentPlayerIndex: int(state.CurrentPlayerIndex()),
					Visibility:         viewVisibility(game, state, viewer),
				}
				for _, path := range paths {
					for _, facet := range allFacets {
						rc.Reads = append(rc.Reads, []interface{}{path, facetName(facet),
							boardgame.LegalReadEvaluable(state, boardgame.PlayerIndex(viewer), boardgame.LegalRead{Path: boardgame.LegalPropPath(path), Facet: facet})})
					}
				}
				out.readCases = append(out.readCases, rc)
			}
		}
	}
	if len(out.readCases) == 0 {
		panic(fmt.Sprint("gen: no evaluability cases"))
	}
	return out
}
