package graph

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"sort"
	"testing"
)

/*
 | 0 | 1 | 2 | 3 |
-|----------------
0| 0 | 1 | 2 | 3 |
-|----------------
1| 4 | 5 | 6 | 7 |
-|----------------
2| 8 | 9 | 10| 11|
-|----------------
3| 12| 13| 14| 15|
-|----------------
*/

func TestBasic(t *testing.T) {

	set := enum.NewSet()

	e := set.MustAddRange("whatever", 4, 4)

	graph := New(false, e)

	assert.For(t).ThatActual(graph).IsNotNil()

	assert.For(t).ThatActual(graph.Connected(e.RangeToValue(0, 1), e.RangeToValue(1, 0))).IsFalse()

	err := graph.AddEdge(e.RangeToValue(0, 1), e.RangeToValue(1, 1))

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(graph.Connected(e.RangeToValue(0, 1), e.RangeToValue(1, 1))).IsTrue()

	err = graph.AddEdge(e.RangeToValue(0, 1), e.RangeToValue(0, 0))

	n := graph.Neighbors(e.RangeToValue(0, 1))

	sort.Ints(n)

	assert.For(t).ThatActual(n).Equals([]int{0, 5})

	//0,4 is not a valid index
	err = graph.AddEdge(e.RangeToValue(0, 1), e.RangeToValue(0, 4))

	assert.For(t).ThatActual(err).IsNotNil()

	graph.Finish()

	//the graph has been finished so no modifications may be made
	err = graph.AddEdge(e.RangeToValue(0, 1), e.RangeToValue(2, 0))

	assert.For(t).ThatActual(err).IsNotNil()
}

func TestNewWithPlainEnum(t *testing.T) {
	// Test that New() works with a plain Enum (not RangeEnum)
	set := enum.NewSet()

	e := set.MustAdd("spaces", map[int]string{
		0: "A",
		1: "B",
		2: "C",
	})

	g := New(true, e)

	assert.For(t).ThatActual(g).IsNotNil()

	err := g.AddEdge(0, 1)
	assert.For(t).ThatActual(err).IsNil()

	err = g.AddEdge(1, 2)
	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(g.Connected(0, 1)).IsTrue()
	assert.For(t).ThatActual(g.Connected(1, 0)).IsTrue() // undirected
	assert.For(t).ThatActual(g.Connected(0, 2)).IsFalse()
	assert.For(t).ThatActual(g.Connected(1, 2)).IsTrue()

	g.Finish()
}

// Build a simple undirected graph for path tests:
//
//	0 --- 1 --- 2
//	|           |
//	3 --- 4 --- 5
//
// All edges weight 1 (default).
func makeSimpleGraph(t *testing.T) Graph {
	t.Helper()
	set := enum.NewSet()
	e := set.MustAdd("nodes", map[int]string{
		0: "N0", 1: "N1", 2: "N2",
		3: "N3", 4: "N4", 5: "N5",
	})

	g := New(true, e)

	assert.For(t).ThatActual(g.AddEdge(0, 1)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(1, 2)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(0, 3)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(2, 5)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(3, 4)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(4, 5)).IsNil()

	g.Finish()
	return g
}

func TestShortestPathDirect(t *testing.T) {
	g := makeSimpleGraph(t)

	// Direct neighbor
	path, err := g.ShortestPath(0, 1)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(path).Equals([]int{0, 1})
}

func TestShortestPathMultiHop(t *testing.T) {
	g := makeSimpleGraph(t)

	// Multi-hop: 0 -> 1 -> 2 (length 2) or 0 -> 3 -> 4 -> 5 -> 2 (length 4)
	path, err := g.ShortestPath(0, 2)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(path)).Equals(3)
	assert.For(t).ThatActual(path[0]).Equals(0)
	assert.For(t).ThatActual(path[len(path)-1]).Equals(2)
}

func TestShortestPathSameNode(t *testing.T) {
	g := makeSimpleGraph(t)

	path, err := g.ShortestPath(3, 3)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(path).Equals([]int{3})
}

func TestShortestPathNoPath(t *testing.T) {
	// Create a disconnected graph
	set := enum.NewSet()
	e := set.MustAdd("islands", map[int]string{
		0: "A", 1: "B", 2: "C",
	})

	g := New(true, e)
	assert.For(t).ThatActual(g.AddEdge(0, 1)).IsNil()
	// Node 2 is disconnected
	g.Finish()

	path, err := g.ShortestPath(0, 2)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(len(path)).Equals(0)
}

func TestShortestPathInvalidNodes(t *testing.T) {
	g := makeSimpleGraph(t)

	_, err := g.ShortestPath(99, 0)
	assert.For(t).ThatActual(err).IsNotNil()

	_, err = g.ShortestPath(0, 99)
	assert.For(t).ThatActual(err).IsNotNil()
}

func TestShortestPathWeighted(t *testing.T) {
	// Weighted graph:
	//   0 ---(10)--- 1
	//   |            |
	//  (1)          (1)
	//   |            |
	//   2 ---(1)---- 3
	//
	// Shortest 0->1: 0->2->3->1 (cost 3) NOT 0->1 (cost 10)
	set := enum.NewSet()
	e := set.MustAdd("weighted", map[int]string{
		0: "W0", 1: "W1", 2: "W2", 3: "W3",
	})

	g := New(true, e)
	assert.For(t).ThatActual(g.AddEdge(0, 1)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(0, 2)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(2, 3)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(3, 1)).IsNil()

	assert.For(t).ThatActual(g.SetEdgeWeight(0, 1, 10)).IsNil()
	assert.For(t).ThatActual(g.SetEdgeWeight(1, 0, 10)).IsNil()
	// Other edges left at default (0 → effective weight 1)

	g.Finish()

	path, err := g.ShortestPath(0, 1)
	assert.For(t).ThatActual(err).IsNil()
	// Should go through 0->2->3->1, not direct 0->1
	assert.For(t).ThatActual(len(path)).Equals(4)
	assert.For(t).ThatActual(path[0]).Equals(0)
	assert.For(t).ThatActual(path[len(path)-1]).Equals(1)
}

func TestDistance(t *testing.T) {
	g := makeSimpleGraph(t)

	// Same node
	d, err := g.Distance(0, 0)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(0)

	// Direct neighbor (weight 1)
	d, err = g.Distance(0, 1)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(1)

	// Two hops
	d, err = g.Distance(0, 2)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(2)

	// No path
	set := enum.NewSet()
	e := set.MustAdd("islands", map[int]string{0: "A", 1: "B", 2: "C"})
	disconnected := New(true, e)
	assert.For(t).ThatActual(disconnected.AddEdge(0, 1)).IsNil()
	disconnected.Finish()

	d, err = disconnected.Distance(0, 2)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(d).Equals(-1)
}

func TestDistanceWeighted(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("weighted", map[int]string{
		0: "W0", 1: "W1", 2: "W2", 3: "W3",
	})

	g := New(true, e)
	assert.For(t).ThatActual(g.AddEdge(0, 1)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(0, 2)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(2, 3)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(3, 1)).IsNil()

	assert.For(t).ThatActual(g.SetEdgeWeight(0, 1, 10)).IsNil()
	assert.For(t).ThatActual(g.SetEdgeWeight(1, 0, 10)).IsNil()

	g.Finish()

	// 0->2->3->1 = 1+1+1 = 3, not 0->1 = 10
	d, err := g.Distance(0, 1)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(3)
}
