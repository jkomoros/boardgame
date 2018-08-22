package config

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"
	"strings"
)

/*
GameNode is a special struct that can represent either the terminal leaf
game packages, or the mid-folders. It has its own UnmarshalJSON which
decides based on the raw json whether to populate Leafs or Mids. This means
that the actual JSON ingested can elide the "Leafs" or "Mids" keys for
brevity.

A Leafs that has the value "" means "one key with no additional suffix". This
allows terminating leafs within a node that is otherwise a mid.

Example valid JSON to ingest for GameNode:
	{
		"github.com/jkomoros": {
			"boardgame/examples": [
				"checkers",
				"blackjack"
			],
			"other-games-repo": [
				"pass"
			],
			"other-mixed-leaf-mid-repo": {
				"leaf": [
					""
				],
				"subdir": [
					"foo"
				]
			}
		}
	}

*/
type GameNode struct {
	Leafs []string
	Mids  map[string]*GameNode
}

//NewGameNode takes the given values and creates a reduced GameNode tree where
//all of the common prefixes are factored out.
func NewGameNode(paths ...string) *GameNode {

	panic("Not yet tested")

	result := newGameNodeItem()

	for _, path := range paths {
		splitPath := strings.Split(path, string(filepath.Separator))
		result.addPath(splitPath)
	}

	result.reduceTerminals()

	return result.elideSimpleMids()

}

//newGameNodeItem returns a very basic initalized gameNode. In particular,
//mids will not be nil.
func newGameNodeItem() *GameNode {
	return &GameNode{
		Mids: make(map[string]*GameNode),
	}
}

//addPath adds the given split string into this GameNode's mids, creating new
//sub-game nodes if necessary. Designed to only be called in NewGameNode,
//because further normalization must happen later.
func (g *GameNode) addPath(path []string) {
	//This shouldn't happen
	if len(path) == 0 {
		return
	}

	newItem := newGameNodeItem()

	//Base case
	g.Mids[path[0]] = newItem

	if len(path) == 1 {
		return
	}

	newItem.addPath(path[1:])

}

//reduceTerminals goes through each node and if all of its mids are GameNodes
//with no Mids or Leafs, then makes this child a Leafs terminal. Designed to
//be called only at the end of NewGameNode. This expects there to be no Leafs
//yet.
func (g *GameNode) reduceTerminals() {

	if len(g.Mids) == 0 {
		g.Leafs = []string{""}
		g.Mids = nil
		return
	}

	nonLeafFound := false

	for _, node := range g.Mids {
		if len(node.Mids) != 0 {
			nonLeafFound = true
		}
	}

	if !nonLeafFound {
		//All of the children are terminal, hoist into self.
		var leafs []string
		for key := range g.Mids {
			leafs = append(leafs, key)
		}
		g.Leafs = leafs
		g.Mids = nil
		return
	}

	//At least one of mids is non-terminal. The base case of the rest will
	//give them the terminal leaves.
	for _, node := range g.Mids {
		node.reduceTerminals()
	}

}

//addPrefix joins the given prefix to the front of all Mids and Leafs in this
//node.
func (g *GameNode) addPrefix(prefix string) {

	if len(g.Mids) > 0 {
		newMids := make(map[string]*GameNode, len(g.Mids))

		for key, node := range g.Mids {
			newMids[filepath.Join(prefix, key)] = node
		}

		g.Mids = newMids
	}

	if len(g.Leafs) > 0 {
		newLeafs := make([]string, len(g.Leafs))

		for i, leaf := range g.Leafs {
			newLeafs[i] = filepath.Join(prefix, leaf)
		}

		g.Leafs = newLeafs

	}
}

//elideSimpleMids removes any nodes that have a single child, merging the
//children with the Mid name and reducing this node. It returns the new root
//GameNode. Designed to be called as the last step of NewGameNode.
func (g *GameNode) elideSimpleMids() *GameNode {

	var childKey string
	var child *GameNode

	for key, node := range g.Mids {
		childKey = key
		child = node.elideSimpleMids()
		g.Mids[key] = child
	}

	if len(g.Mids) != 1 {
		return g
	}

	//If there's precisely one child, then hoist the child and add our key to it.

	//child is set to the one child, childKey is its key
	child.addPrefix(childKey)

	return child

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

/*
List returns a flattened, alphabetized, unique list of paths implied by the
contents of this node. Mids are joined by filepath.Separator.

Input:
	{
		"github.com/jkomoros": {
			"boardgame/examples": [
				"checkers",
				"blackjack"
			],
			"other-games-repo": [
				"pass"
			],
			"other-mixed-leaf-mid-repo": {
				"leaf": [
					""
				],
				"subdir": [
					"foo"
				]
			}
		}
	}

Output:
	[
		"github.com/jkomoros/boardgame/examples/blackjack",
		"github.com/jkomoros/boardgame/examples/checkers",
		"github.com/jkomoros/other-games-repo/pass",
		"github.com/jkomoros/other-mixed-leaf-mid-repo/leaf",
		"github.com/jkomoros/other-mixed-leaf-mid-repo/subdir/foo",
	]

*/
func (g *GameNode) List() []string {

	return alphabetizeUnique(g.listRecursive(""))

}

func alphabetizeUnique(in []string) []string {

	if in == nil {
		return nil
	}

	set := make(map[string]bool, len(in))
	for _, str := range in {
		set[str] = true
	}

	result := make([]string, len(set))

	i := 0
	for str := range set {
		result[i] = str
		i++
	}

	sort.Strings(result)

	return result
}

//listRecursive is the main implementaiton of List. prior is the prior part of
//the path implied so far.
func (g *GameNode) listRecursive(prior string) []string {

	if g == nil {
		return nil
	}

	var result []string

	if len(g.Leafs) > 0 {
		//Base case
		for _, leaf := range g.Leafs {
			result = append(result, filepath.Join(prior, leaf))
		}
	}

	for name, node := range g.Mids {
		result = append(result, node.listRecursive(filepath.Join(prior, name))...)
	}

	return result

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

//normalize ensures that the given node has either all mids or all leafs. If
//it has any mids, converts all of the leafs to be terminal mids.
func (g *GameNode) normalize() {

	if g == nil {
		return
	}

	//If no mids, then 0 or more leafs are fine.s
	if len(g.Mids) == 0 {
		return
	}

	for _, node := range g.Mids {
		node.normalize()
	}

	//If no leafs, than 0 or more mids are fine.
	if len(g.Leafs) == 0 {
		return
	}

	//We have both leafs and mids. Go through and convert each leaf to a mid.
	for _, leaf := range g.Leafs {

		newMid := &GameNode{
			Leafs: []string{
				"",
			},
		}

		g.Mids[leaf] = newMid

	}

	g.Leafs = nil
}

//extend takes an other GameNode and returns a *new* GameNode representing the
//merging of the two, where the keys in other overwrite the keys in this.
//Should call normalize() after calling on the top-level node.
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
