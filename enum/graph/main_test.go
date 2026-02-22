package graph

import (
	"slices"
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
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

	slices.Sort(n)

	assert.For(t).ThatActual(n).Equals([]enum.EnumKey{0, 5})

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

	e := set.MustAdd("spaces", map[enum.EnumKey]string{
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
	e := set.MustAdd("nodes", map[enum.EnumKey]string{
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
	assert.For(t).ThatActual(path).Equals([]enum.EnumKey{0, 1})
}

func TestShortestPathMultiHop(t *testing.T) {
	g := makeSimpleGraph(t)

	// Multi-hop: 0 -> 1 -> 2 (length 2) or 0 -> 3 -> 4 -> 5 -> 2 (length 4)
	path, err := g.ShortestPath(0, 2)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(path)).Equals(3)
	assert.For(t).ThatActual(path[0]).Equals(enum.EnumKey(0))
	assert.For(t).ThatActual(path[len(path)-1]).Equals(enum.EnumKey(2))
}

func TestShortestPathSameNode(t *testing.T) {
	g := makeSimpleGraph(t)

	path, err := g.ShortestPath(3, 3)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(path).Equals([]enum.EnumKey{3})
}

func TestShortestPathNoPath(t *testing.T) {
	// Create a disconnected graph
	set := enum.NewSet()
	e := set.MustAdd("islands", map[enum.EnumKey]string{
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
	e := set.MustAdd("weighted", map[enum.EnumKey]string{
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
	assert.For(t).ThatActual(path[0]).Equals(enum.EnumKey(0))
	assert.For(t).ThatActual(path[len(path)-1]).Equals(enum.EnumKey(1))
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
	e := set.MustAdd("islands", map[enum.EnumKey]string{0: "A", 1: "B", 2: "C"})
	disconnected := New(true, e)
	assert.For(t).ThatActual(disconnected.AddEdge(0, 1)).IsNil()
	disconnected.Finish()

	d, err = disconnected.Distance(0, 2)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(d).Equals(-1)
}

func TestDistanceWeighted(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("weighted", map[enum.EnumKey]string{
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

// --- EnumGraph tests ---

func makeSimpleEnumGraph(t *testing.T) *EnumGraph {
	t.Helper()
	set := enum.NewSet()
	e := set.MustAdd("nodes", map[enum.EnumKey]string{
		0: "N0", 1: "N1", 2: "N2",
		3: "N3", 4: "N4", 5: "N5",
	})

	g := NewEnumGraph(e)

	v0 := e.MustNewImmutableVal(0)
	v1 := e.MustNewImmutableVal(1)
	v2 := e.MustNewImmutableVal(2)
	v3 := e.MustNewImmutableVal(3)
	v4 := e.MustNewImmutableVal(4)
	v5 := e.MustNewImmutableVal(5)

	assert.For(t).ThatActual(g.AddEdge(v0, v1)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(v1, v2)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(v0, v3)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(v2, v5)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(v3, v4)).IsNil()
	assert.For(t).ThatActual(g.AddEdge(v4, v5)).IsNil()

	g.Finish()
	return g
}

func TestEnumGraphBasic(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("spaces", map[enum.EnumKey]string{
		0: "A", 1: "B", 2: "C",
	})

	g := NewEnumGraph(e)

	assert.For(t).ThatActual(g.Enum()).Equals(e)
	assert.For(t).ThatActual(g.Inner()).IsNotNil()

	v0 := e.MustNewImmutableVal(0)
	v1 := e.MustNewImmutableVal(1)
	v2 := e.MustNewImmutableVal(2)

	// Add edge and check connectivity
	assert.For(t).ThatActual(g.AddEdge(v0, v1)).IsNil()
	assert.For(t).ThatActual(g.Connected(v0, v1)).IsTrue()
	assert.For(t).ThatActual(g.Connected(v1, v0)).IsTrue() // undirected
	assert.For(t).ThatActual(g.Connected(v0, v2)).IsFalse()

	// AddEdges
	assert.For(t).ThatActual(g.AddEdges(v1, v2)).IsNil()
	assert.For(t).ThatActual(g.Connected(v1, v2)).IsTrue()

	g.Finish()
}

func TestEnumGraphByKey(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("spaces", map[enum.EnumKey]string{
		0: "A", 1: "B", 2: "C",
	})

	g := NewEnumGraph(e)

	// ByKey convenience methods
	assert.For(t).ThatActual(g.AddEdgeByKey(0, 1)).IsNil()
	assert.For(t).ThatActual(g.AddEdgesByKey(1, 2)).IsNil()

	v0 := e.MustNewImmutableVal(0)
	v2 := e.MustNewImmutableVal(2)

	// Should be connected via the ByKey edges
	assert.For(t).ThatActual(g.Connected(v0, e.MustNewImmutableVal(1))).IsTrue()
	assert.For(t).ThatActual(g.Connected(e.MustNewImmutableVal(1), v2)).IsTrue()

	g.Finish()
}

func TestEnumGraphShortestPath(t *testing.T) {
	g := makeSimpleEnumGraph(t)
	e := g.Enum()

	v0 := e.MustNewImmutableVal(0)
	v2 := e.MustNewImmutableVal(2)

	path, err := g.ShortestPath(v0, v2)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(path)).Equals(3)
	assert.For(t).ThatActual(path[0].Value()).Equals(enum.EnumKey(0))
	assert.For(t).ThatActual(path[len(path)-1].Value()).Equals(enum.EnumKey(2))
}

func TestEnumGraphDistance(t *testing.T) {
	g := makeSimpleEnumGraph(t)
	e := g.Enum()

	v0 := e.MustNewImmutableVal(0)
	v1 := e.MustNewImmutableVal(1)
	v5 := e.MustNewImmutableVal(5)

	d, err := g.Distance(v0, v0)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(0)

	d, err = g.Distance(v0, v1)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(1)

	d, err = g.Distance(v0, v5)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(3) // 0->3->4->5
}

func TestEnumGraphNeighbors(t *testing.T) {
	g := makeSimpleEnumGraph(t)
	e := g.Enum()

	v0 := e.MustNewImmutableVal(0)

	neighbors := g.Neighbors(v0)
	assert.For(t).ThatActual(len(neighbors)).Equals(2)

	// Collect neighbor keys
	keys := make([]enum.EnumKey, len(neighbors))
	for i, n := range neighbors {
		keys[i] = n.Value()
	}
	slices.Sort(keys)
	assert.For(t).ThatActual(keys).Equals([]enum.EnumKey{1, 3})
}

func TestEnumGraphNilChecks(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("spaces", map[enum.EnumKey]string{
		0: "A", 1: "B",
	})

	g := NewEnumGraph(e)
	v0 := e.MustNewImmutableVal(0)

	// AddEdge nil checks
	assert.For(t).ThatActual(g.AddEdge(nil, v0)).IsNotNil()
	assert.For(t).ThatActual(g.AddEdge(v0, nil)).IsNotNil()

	// AddEdges nil checks
	assert.For(t).ThatActual(g.AddEdges(nil, v0)).IsNotNil()
	assert.For(t).ThatActual(g.AddEdges(v0, nil)).IsNotNil()

	// Connected nil checks
	assert.For(t).ThatActual(g.Connected(nil, v0)).IsFalse()
	assert.For(t).ThatActual(g.Connected(v0, nil)).IsFalse()

	// Neighbors nil check
	assert.For(t).ThatActual(g.Neighbors(nil) == nil).IsTrue()

	// ShortestPath nil checks
	_, err := g.ShortestPath(nil, v0)
	assert.For(t).ThatActual(err).IsNotNil()
	_, err = g.ShortestPath(v0, nil)
	assert.For(t).ThatActual(err).IsNotNil()

	// Distance nil checks
	d, err := g.Distance(nil, v0)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(d).Equals(-1)

	// SetEdgeWeight nil checks
	assert.For(t).ThatActual(g.SetEdgeWeight(nil, v0, 5)).IsNotNil()

	// EdgeWeight nil checks
	assert.For(t).ThatActual(g.EdgeWeight(nil, v0)).Equals(0)

	g.Finish()
}

func TestEnumGraphWrap(t *testing.T) {
	// WrapGraph should wrap an existing Graph
	g := makeSimpleGraph(t)
	eg := WrapGraph(g)

	assert.For(t).ThatActual(eg.Inner()).Equals(g)
	assert.For(t).ThatActual(eg.Enum()).Equals(g.Enum())

	v0 := eg.Enum().MustNewImmutableVal(0)
	v1 := eg.Enum().MustNewImmutableVal(1)

	assert.For(t).ThatActual(eg.Connected(v0, v1)).IsTrue()
}

func TestEnumGraphByKeyQueries(t *testing.T) {
	g := makeSimpleEnumGraph(t)

	// ConnectedByKey
	assert.For(t).ThatActual(g.ConnectedByKey(0, 1)).IsTrue()
	assert.For(t).ThatActual(g.ConnectedByKey(1, 0)).IsTrue()
	assert.For(t).ThatActual(g.ConnectedByKey(0, 2)).IsFalse()

	// NeighborsByKey
	neighbors := g.NeighborsByKey(0)
	slices.Sort(neighbors)
	assert.For(t).ThatActual(neighbors).Equals([]enum.EnumKey{1, 3})

	// ShortestPathByKey
	path, err := g.ShortestPathByKey(0, 2)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(path)).Equals(3)
	assert.For(t).ThatActual(path[0]).Equals(enum.EnumKey(0))
	assert.For(t).ThatActual(path[len(path)-1]).Equals(enum.EnumKey(2))

	// DistanceByKey
	d, err := g.DistanceByKey(0, 1)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(1)

	d, err = g.DistanceByKey(0, 5)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(d).Equals(3)

	// EdgeWeightByKey (default weight 0)
	w := g.EdgeWeightByKey(0, 1)
	assert.For(t).ThatActual(w).Equals(0)
}

func TestEnumGraphSetEdgeWeightByKey(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("nodes", map[enum.EnumKey]string{
		0: "N0", 1: "N1", 2: "N2",
	})

	g := NewEnumGraph(e)
	assert.For(t).ThatActual(g.AddEdgeByKey(0, 1)).IsNil()
	assert.For(t).ThatActual(g.AddEdgeByKey(1, 2)).IsNil()
	assert.For(t).ThatActual(g.SetEdgeWeightByKey(0, 1, 5)).IsNil()

	assert.For(t).ThatActual(g.EdgeWeightByKey(0, 1)).Equals(5)
	assert.For(t).ThatActual(g.EdgeWeightByKey(1, 2)).Equals(0)

	g.Finish()
}

func TestNewDirectedEnumGraph(t *testing.T) {
	set := enum.NewSet()
	e := set.MustAdd("spaces", map[enum.EnumKey]string{
		0: "A", 1: "B",
	})

	g := NewDirectedEnumGraph(e)

	v0 := e.MustNewImmutableVal(0)
	v1 := e.MustNewImmutableVal(1)

	assert.For(t).ThatActual(g.AddEdge(v0, v1)).IsNil()
	assert.For(t).ThatActual(g.Connected(v0, v1)).IsTrue()
	assert.For(t).ThatActual(g.Connected(v1, v0)).IsFalse() // directed

	g.Finish()
}

func TestDirectionFilters(t *testing.T) {
	// Use a 4x4 grid to test direction filters.
	//
	// | 0 | 1 | 2 | 3 |
	// | 4 | 5 | 6 | 7 |
	// | 8 | 9 |10 |11 |
	// |12 |13 |14 |15 |
	set := enum.NewSet()
	e := set.MustAddRange("grid", 4, 4)

	// DirectionDown: neighbors should have strictly higher row index.
	// Cell 5 is at row 1, col 1. Its Down neighbors should be in row 2.
	downNeighbors := []enum.EnumKey{}
	for _, n := range neighbors(e, 5) {
		if DirectionDown(e, 5, n) {
			downNeighbors = append(downNeighbors, n)
		}
	}
	slices.Sort(downNeighbors)
	// Row 2, cols 0-2: cells 8, 9, 10
	assert.For(t).ThatActual(downNeighbors).Equals([]enum.EnumKey{8, 9, 10})

	// DirectionUp: neighbors should have strictly lower row index.
	// Cell 5 is at row 1. Its Up neighbors should be in row 0.
	upNeighbors := []enum.EnumKey{}
	for _, n := range neighbors(e, 5) {
		if DirectionUp(e, 5, n) {
			upNeighbors = append(upNeighbors, n)
		}
	}
	slices.Sort(upNeighbors)
	// Row 0, cols 0-2: cells 0, 1, 2
	assert.For(t).ThatActual(upNeighbors).Equals([]enum.EnumKey{0, 1, 2})

	// DirectionRight: neighbors should have strictly higher col index.
	// Cell 5 is at row 1, col 1. Its Right neighbors should be in col 2.
	rightNeighbors := []enum.EnumKey{}
	for _, n := range neighbors(e, 5) {
		if DirectionRight(e, 5, n) {
			rightNeighbors = append(rightNeighbors, n)
		}
	}
	slices.Sort(rightNeighbors)
	// Col 2, rows 0-2: cells 2, 6, 10
	assert.For(t).ThatActual(rightNeighbors).Equals([]enum.EnumKey{2, 6, 10})

	// DirectionLeft: neighbors should have strictly lower col index.
	// Cell 5 is at row 1, col 1. Its Left neighbors should be in col 0.
	leftNeighbors := []enum.EnumKey{}
	for _, n := range neighbors(e, 5) {
		if DirectionLeft(e, 5, n) {
			leftNeighbors = append(leftNeighbors, n)
		}
	}
	slices.Sort(leftNeighbors)
	// Col 0, rows 0-2: cells 0, 4, 8
	assert.For(t).ThatActual(leftNeighbors).Equals([]enum.EnumKey{0, 4, 8})

	// Test grid connectedness with DirectionDown + DirectionDiagonal filter
	// (used by checkers for downward movement).
	graph, err := NewGridConnectedness(e, DirectionDown, DirectionDiagonal)
	assert.For(t).ThatActual(err).IsNil()

	// Cell 0 (row 0, col 0) should connect to cell 5 (row 1, col 1) — down-right diagonal
	assert.For(t).ThatActual(graph.Connected(0, 5)).IsTrue()
	// Cell 0 should NOT connect to cell 1 (same row, not down)
	assert.For(t).ThatActual(graph.Connected(0, 1)).IsFalse()
	// Cell 0 should NOT connect to cell 4 (down but not diagonal — perpendicular)
	assert.For(t).ThatActual(graph.Connected(0, 4)).IsFalse()
}
