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

//copy returns a fresh, deep copy of gameNode.
func (g *GameNode) copy() *GameNode {

	if g == nil {
		return nil
	}

	result := &GameNode{}

	if g.Mids != nil {
		mids := make(map[string]*GameNode, len(g.Mids))
		for key, val := range g.Mids {
			mids[key] = val.copy()
		}
		result.Mids = mids
	}

	if g.Leafs != nil {
		leafs := make([]string, len(g.Leafs))
		for i, leaf := range g.Leafs {
			leafs[i] = leaf
		}
		result.Leafs = leafs
	}

	return result

}

//extend takes an other GameNode and returns a *new* GameNode representing the
//merging of the two, where the keys in other overwrite the keys in this.
func (g *GameNode) extend(other *GameNode) *GameNode {

	if g == nil {
		return other.copy()
	}

	//Note: it's possible that its' better to do a shallowCopy, where we copy
	//map/lists of the gameNode, but not the pointers, because we're going to
	//rewrite them anyway. Think about whether that will break some other behavior.
	result := g.copy()

	if other == nil {
		return result
	}

	result.Leafs = mergedStrList(result.Leafs, other.Leafs)

	processedNodeNames := make(map[string]bool)

	for nodeName, node := range result.Mids {
		//This will handle nil nodes in other correctly
		result.Mids[nodeName] = node.extend(other.Mids[nodeName])
		processedNodeNames[nodeName] = true
	}

	//Now copy in the node names in other that weren't in g
	for nodeName, node := range other.Mids {
		if processedNodeNames[nodeName] {
			continue
		}
		result.Mids[nodeName] = node.copy()
	}

	return result

}
