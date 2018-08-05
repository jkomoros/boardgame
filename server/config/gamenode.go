package config

import (
	"encoding/json"
	"errors"
)

/*
GameNode is a special struct that can represent either the terminal leaf
game packages, or the mid-folders. It has its own UnmarshalJSON which
decides based on the raw json whether to populate Leafs or Mids. This means
that the actual JSON ingested can elide the "Leafs" or "Mids" keys for
brevity.

Example valid JSON to ingest for GameNode:
	{
		"github.com/jkomoros": {
			"boardgame/examples": [
				"checkers",
				"blackjack"
			],
			"other-games-repo": [
				"pass"
			]
		}
	}

*/
type GameNode struct {
	Leafs []string
	Mids  map[string]*GameNode
}

func (g *GameNode) UnmarshalJSON(raw []byte) error {
	var strs []string
	var mids map[string]*GameNode

	if err := json.Unmarshal(raw, &strs); err == nil {
		//Looks like it was a Leafs.
		g.Leafs = strs
		return nil
	}

	if err := json.Unmarshal(raw, &mids); err == nil {
		//Looks like it aas a mids.
		g.Mids = mids
		return nil
	}

	return errors.New("Node didn't appear to be either Mids or Leafs")

}

func (g *GameNode) MarshalJSON() ([]byte, error) {
	if g.Leafs != nil {
		return json.Marshal(g.Leafs)
	}
	if g.Mids != nil {
		return json.Marshal(g.Mids)
	}
	return nil, nil
}
