package behaviors

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/enum/graph"
)

/*
LocationBehavior is a struct designed to be anonymously embedded in a SubState
(typically a playerState or gameState) to represent the position of a token
within a SizedStack. It is a Connectable behavior, so you must call
ConnectBehavior from within your SubState's FinishStateSetUp. You must also
call ConnectLocationStack to point it at the SizedStack that tracks the token's
position. Optionally, call ConnectGraph to associate a graph for adjacency and
pathfinding queries.

RemainingPath stores the remaining hops during multi-hop movement. It is
serialized via TypeIntSlice and is used by the HopAlongPath FixUp move.

Example wiring:

	func (p *playerState) FinishStateSetUp() {
	    p.LocationBehavior.ConnectBehavior(p)
	    p.LocationBehavior.ConnectLocationStack(p.Location)
	    p.LocationBehavior.ConnectGraph(connectedGraph)
	}
*/
type LocationBehavior struct {
	// Unexported runtime fields (not serialized)
	container     boardgame.SubState
	locationStack boardgame.SizedStack
	locGraph      *graph.EnumGraph

	// Exported serializable field
	LocRemainingPath []int
}

// ConnectBehavior stores a reference to the containing SubState.
func (l *LocationBehavior) ConnectBehavior(containingSubState boardgame.SubState) {
	l.container = containingSubState
}

// ConnectLocationStack sets the SizedStack that this behavior tracks for
// position.
func (l *LocationBehavior) ConnectLocationStack(stack boardgame.SizedStack) {
	l.locationStack = stack
}

// ConnectGraph associates an EnumGraph for adjacency and pathfinding
// operations. This is optional; if not connected, methods like Neighbors,
// IsConnectedTo, ShortestPathTo, and DistanceTo will return errors.
func (l *LocationBehavior) ConnectGraph(g *graph.EnumGraph) {
	l.locGraph = g
}

// ValidConfiguration returns an error if the behavior hasn't been properly
// connected.
func (l *LocationBehavior) ValidConfiguration(example boardgame.State) error {
	if l.container == nil {
		return errors.New("LocationBehavior: ConnectBehavior hasn't been called. See the behaviors package doc for more on initializing Connectable behaviors")
	}
	if l.locationStack == nil {
		return errors.New("LocationBehavior: ConnectLocationStack hasn't been called. The behavior needs a SizedStack to track position")
	}
	return nil
}

// LocationEnum returns the enum associated with the connected graph, or nil if
// no graph has been connected. This is useful for constructing ImmutableVals
// from location indices.
func (l *LocationBehavior) LocationEnum() enum.Enum {
	if l.locGraph == nil {
		return nil
	}
	return l.locGraph.Enum()
}

// LocationIndex scans the connected SizedStack for the first non-nil component
// and returns its slot index as an enum.ImmutableVal. Returns nil if no graph
// is connected or no component is found.
func (l *LocationBehavior) LocationIndex() enum.ImmutableVal {
	if l.locationStack == nil || l.locGraph == nil {
		return nil
	}
	for i, c := range l.locationStack.Components() {
		if c == nil {
			continue
		}
		return l.locGraph.Enum().MustNewImmutableVal(enum.EnumKey(i))
	}
	return nil
}

// LocationIndexKey is a convenience method that returns the raw EnumKey of the
// current location. Returns (0, false) if no component is found or no stack is
// connected.
func (l *LocationBehavior) LocationIndexKey() (enum.EnumKey, bool) {
	if l.locationStack == nil {
		return 0, false
	}
	for i, c := range l.locationStack.Components() {
		if c == nil {
			continue
		}
		return enum.EnumKey(i), true
	}
	return 0, false
}

// MoveTo moves the token from its current position to the target slot index by
// swapping components in the SizedStack.
func (l *LocationBehavior) MoveTo(targetIndex int) error {
	if l.locationStack == nil {
		return errors.New("LocationBehavior: locationStack not connected")
	}
	currentKey, ok := l.LocationIndexKey()
	if !ok {
		return errors.New("LocationBehavior: no component found in location stack")
	}
	return l.locationStack.SwapComponents(currentKey.Int(), targetIndex)
}

// Neighbors returns the indices of all spaces adjacent to the current location
// in the connected graph as ImmutableVals. Returns nil if no graph is connected.
func (l *LocationBehavior) Neighbors() []enum.ImmutableVal {
	if l.locGraph == nil {
		return nil
	}
	idx := l.LocationIndex()
	if idx == nil {
		return nil
	}
	return l.locGraph.Neighbors(idx)
}

// IsConnectedTo returns whether the current location is directly connected to
// the target in the graph. Returns false if no graph is connected.
func (l *LocationBehavior) IsConnectedTo(target enum.ImmutableVal) bool {
	if l.locGraph == nil || target == nil {
		return false
	}
	idx := l.LocationIndex()
	if idx == nil {
		return false
	}
	return l.locGraph.Connected(idx, target)
}

// ShortestPathTo returns the shortest path from the current location to the
// target, inclusive of both endpoints, as ImmutableVals. Returns an error if
// no graph is connected or no path exists.
func (l *LocationBehavior) ShortestPathTo(target enum.ImmutableVal) ([]enum.ImmutableVal, error) {
	if l.locGraph == nil {
		return nil, errors.New("LocationBehavior: no graph connected")
	}
	if target == nil {
		return nil, errors.New("LocationBehavior: target is nil")
	}
	idx := l.LocationIndex()
	if idx == nil {
		return nil, errors.New("LocationBehavior: no component found in location stack")
	}
	return l.locGraph.ShortestPath(idx, target)
}

// DistanceTo returns the total weight of the shortest path from the current
// location to the target. Returns -1 and an error if no graph is connected or
// no path exists.
func (l *LocationBehavior) DistanceTo(target enum.ImmutableVal) (int, error) {
	if l.locGraph == nil {
		return -1, errors.New("LocationBehavior: no graph connected")
	}
	if target == nil {
		return -1, errors.New("LocationBehavior: target is nil")
	}
	idx := l.LocationIndex()
	if idx == nil {
		return -1, errors.New("LocationBehavior: no component found in location stack")
	}
	return l.locGraph.Distance(idx, target)
}

// Graph returns the connected EnumGraph, or nil if none.
func (l *LocationBehavior) Graph() *graph.EnumGraph {
	return l.locGraph
}

// Token returns the ComponentInstance at the current location in the SizedStack.
// Returns nil if no stack is connected or no component is found.
func (l *LocationBehavior) Token() boardgame.ComponentInstance {
	if l.locationStack == nil {
		return nil
	}
	currentKey, ok := l.LocationIndexKey()
	if !ok {
		return nil
	}
	return l.locationStack.ComponentAt(currentKey.Int())
}

// GetLocationBehavior returns the LocationBehavior itself. This method exists
// so that any struct that embeds LocationBehavior automatically satisfies the
// HasLocationBehavior interface, allowing the HopAlongPath FixUp to find it.
func (l *LocationBehavior) GetLocationBehavior() *LocationBehavior {
	return l
}

// HasLocationBehavior is implemented by any SubState that embeds a
// LocationBehavior. It allows framework moves like HopAlongPath to find
// LocationBehaviors by scanning player and game states via type assertion.
type HasLocationBehavior interface {
	GetLocationBehavior() *LocationBehavior
}
