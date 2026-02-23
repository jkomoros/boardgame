package graph

import (
	"errors"

	"github.com/jkomoros/boardgame/enum"
)

// EnumGraph wraps a Graph with Val-centric methods so that game developers
// configure spatial games without seeing EnumKey. All methods mirror the
// existing Graph interface but accept and return enum.ImmutableVal instead of
// enum.EnumKey.
type EnumGraph struct {
	inner   Graph
	theEnum enum.Enum
}

// NewEnumGraph returns a new EnumGraph backed by a fresh undirected graph for
// the given enum.
func NewEnumGraph(e enum.Enum) *EnumGraph {
	return &EnumGraph{
		inner:   New(true, e),
		theEnum: e,
	}
}

// NewDirectedEnumGraph returns a new EnumGraph backed by a fresh directed graph
// for the given enum.
func NewDirectedEnumGraph(e enum.Enum) *EnumGraph {
	return &EnumGraph{
		inner:   New(false, e),
		theEnum: e,
	}
}

// WrapGraph wraps an existing Graph into an EnumGraph. The enum is taken from
// the graph itself.
func WrapGraph(g Graph) *EnumGraph {
	return &EnumGraph{
		inner:   g,
		theEnum: g.Enum(),
	}
}

// Inner returns the underlying Graph, for use by framework internals that
// still operate on EnumKey.
func (eg *EnumGraph) Inner() Graph {
	return eg.inner
}

// Enum returns the enum associated with this graph.
func (eg *EnumGraph) Enum() enum.Enum {
	return eg.theEnum
}

// AddEdge adds an edge between two nodes specified as ImmutableVals.
func (eg *EnumGraph) AddEdge(from, to enum.ImmutableVal) error {
	if from == nil || to == nil {
		return errors.New("EnumGraph.AddEdge: from and to must not be nil")
	}
	return eg.inner.AddEdge(from.Value(), to.Value())
}

// AddEdges adds edges from one node to multiple target nodes, all specified as
// ImmutableVals.
func (eg *EnumGraph) AddEdges(from enum.ImmutableVal, to ...enum.ImmutableVal) error {
	if from == nil {
		return errors.New("EnumGraph.AddEdges: from must not be nil")
	}
	keys := make([]enum.EnumKey, len(to))
	for i, v := range to {
		if v == nil {
			return errors.New("EnumGraph.AddEdges: to values must not be nil")
		}
		keys[i] = v.Value()
	}
	return eg.inner.AddEdges(from.Value(), keys...)
}

// Connected returns whether from is directly connected to to. Returns false if
// either argument is nil.
func (eg *EnumGraph) Connected(from, to enum.ImmutableVal) bool {
	if from == nil || to == nil {
		return false
	}
	return eg.inner.Connected(from.Value(), to.Value())
}

// Neighbors returns all nodes adjacent to start as ImmutableVals. Returns nil
// if start is nil.
func (eg *EnumGraph) Neighbors(start enum.ImmutableVal) []enum.ImmutableVal {
	if start == nil {
		return nil
	}
	keys := eg.inner.Neighbors(start.Value())
	result := make([]enum.ImmutableVal, len(keys))
	for i, k := range keys {
		result[i] = eg.theEnum.MustNewImmutableVal(k)
	}
	return result
}

// ShortestPath returns the shortest path between start and end as ImmutableVals.
func (eg *EnumGraph) ShortestPath(start, end enum.ImmutableVal) ([]enum.ImmutableVal, error) {
	if start == nil || end == nil {
		return nil, errors.New("EnumGraph.ShortestPath: start and end must not be nil")
	}
	keys, err := eg.inner.ShortestPath(start.Value(), end.Value())
	if err != nil {
		return nil, err
	}
	result := make([]enum.ImmutableVal, len(keys))
	for i, k := range keys {
		result[i] = eg.theEnum.MustNewImmutableVal(k)
	}
	return result, nil
}

// Distance returns the total weight of the shortest path between start and end.
func (eg *EnumGraph) Distance(start, end enum.ImmutableVal) (int, error) {
	if start == nil || end == nil {
		return -1, errors.New("EnumGraph.Distance: start and end must not be nil")
	}
	return eg.inner.Distance(start.Value(), end.Value())
}

// SetEdgeWeight sets the weight of an edge between two nodes.
func (eg *EnumGraph) SetEdgeWeight(from, to enum.ImmutableVal, weight int) error {
	if from == nil || to == nil {
		return errors.New("EnumGraph.SetEdgeWeight: from and to must not be nil")
	}
	return eg.inner.SetEdgeWeight(from.Value(), to.Value(), weight)
}

// EdgeWeight returns the weight of the edge between two nodes. Returns 0 if
// either argument is nil.
func (eg *EnumGraph) EdgeWeight(from, to enum.ImmutableVal) int {
	if from == nil || to == nil {
		return 0
	}
	return eg.inner.EdgeWeight(from.Value(), to.Value())
}

// Finish marks the graph as finished, preventing further modifications.
func (eg *EnumGraph) Finish() {
	eg.inner.Finish()
}

// AddEdgeByKey adds an edge using EnumKey values directly. This is a
// convenience for game developers who configure graphs using their const values.
func (eg *EnumGraph) AddEdgeByKey(from, to enum.EnumKey) error {
	return eg.inner.AddEdge(from, to)
}

// AddEdgesByKey adds edges from one node to multiple target nodes using EnumKey
// values directly.
func (eg *EnumGraph) AddEdgesByKey(from enum.EnumKey, to ...enum.EnumKey) error {
	return eg.inner.AddEdges(from, to...)
}

// SetEdgeWeightByKey sets the weight of an edge using EnumKey values directly.
func (eg *EnumGraph) SetEdgeWeightByKey(from, to enum.EnumKey, weight int) error {
	return eg.inner.SetEdgeWeight(from, to, weight)
}

// ConnectedByKey returns whether from is directly connected to to, using
// EnumKey values directly.
func (eg *EnumGraph) ConnectedByKey(from, to enum.EnumKey) bool {
	return eg.inner.Connected(from, to)
}

// NeighborsByKey returns all nodes adjacent to start as EnumKey values.
func (eg *EnumGraph) NeighborsByKey(start enum.EnumKey) []enum.EnumKey {
	return eg.inner.Neighbors(start)
}

// ShortestPathByKey returns the shortest path between start and end as
// EnumKey values.
func (eg *EnumGraph) ShortestPathByKey(start, end enum.EnumKey) ([]enum.EnumKey, error) {
	return eg.inner.ShortestPath(start, end)
}

// DistanceByKey returns the total weight of the shortest path between start
// and end, using EnumKey values directly.
func (eg *EnumGraph) DistanceByKey(start, end enum.EnumKey) (int, error) {
	return eg.inner.Distance(start, end)
}

// EdgeWeightByKey returns the weight of the edge between two nodes, using
// EnumKey values directly.
func (eg *EnumGraph) EdgeWeightByKey(from, to enum.EnumKey) int {
	return eg.inner.EdgeWeight(from, to)
}
