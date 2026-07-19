package behaviors

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/enum/graph"
)

/*
LocationBehavior is a struct designed to be anonymously embedded in a SubState
(typically a playerState or gameState) to represent the position of a token
within a SizedStack. It is a [Connectable] behavior that is automatically
connected by the framework.

The simplest way to configure it is with a struct tag on the embedding site.
The "location" tag specifies the name of a SizedStack field on the same struct:

	type playerState struct {
	    base.SubState
	    behaviors.LocationBehavior `location:"Spaces"`
	    Spaces boardgame.SizedStack `sizedstack:"tokens,24"`
	}

This eliminates the need for a FinishStateSetUp override. If you also need to
connect a graph for adjacency and pathfinding queries, use FinishStateSetUp for
the graph while still using the tag for the stack:

	type playerState struct {
	    base.SubState
	    behaviors.LocationBehavior `location:"Spaces"`
	    Spaces boardgame.SizedStack `sizedstack:"tokens,24"`
	}

	func (p *playerState) FinishStateSetUp() {
	    p.LocationBehavior.ConnectGraph(connectedGraph)
	}

You can also skip the tag entirely and wire everything in FinishStateSetUp:

	func (p *playerState) FinishStateSetUp() {
	    p.LocationBehavior.ConnectLocationStack(p.Location)
	    p.LocationBehavior.ConnectGraph(connectedGraph)
	}

RemainingPath stores the remaining hops during multi-hop movement. It is
serialized via TypeIntSlice and is used by the HopAlongPath FixUp move.
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

const locationStructTag = "location"

// ConfigureFromTags reads struct tags from the embedding site and configures
// the behavior. The "location" tag should contain the name of a SizedStack
// field on the containing SubState that tracks the token's position.
//
// Example: `location:"Spaces"` will look up the field named "Spaces" on the
// containing struct and call ConnectLocationStack with it.
//
// If the "location" tag is absent, this method does nothing, allowing manual
// configuration in FinishStateSetUp instead.
func (l *LocationBehavior) ConfigureFromTags(tags reflect.StructTag, containingSubState boardgame.SubState) error {
	fieldName := tags.Get(locationStructTag)
	if fieldName == "" {
		return nil
	}

	sizedStack, err := lookupSizedStackField(containingSubState, fieldName)
	if err != nil {
		return fmt.Errorf("LocationBehavior: location tag: %w", err)
	}

	l.ConnectLocationStack(sizedStack)
	return nil
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

func (l *LocationBehavior) moveIndexes(targetIndex int) (int, int, error) {
	if l.locationStack == nil {
		return 0, 0, errors.New("LocationBehavior: locationStack not connected")
	}
	currentKey, ok := l.LocationIndexKey()
	if !ok {
		return 0, 0, errors.New("LocationBehavior: no component found in location stack")
	}
	return currentKey.Int(), targetIndex, nil
}

// MayMoveTo reports whether MoveTo could move the token to targetIndex without
// modifying the location stack.
func (l *LocationBehavior) MayMoveTo(targetIndex int) error {
	currentIndex, targetIndex, err := l.moveIndexes(targetIndex)
	if err != nil {
		return err
	}
	return l.locationStack.MaySwapComponents(currentIndex, targetIndex)
}

// MoveTo moves the token from its current position to the target slot index by
// swapping components in the SizedStack.
func (l *LocationBehavior) MoveTo(targetIndex int) error {
	currentIndex, targetIndex, err := l.moveIndexes(targetIndex)
	if err != nil {
		return err
	}
	return l.locationStack.SwapComponents(currentIndex, targetIndex)
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
