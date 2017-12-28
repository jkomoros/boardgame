package graph

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
)

//EdgeFilter is a type of function that can be passed to filter in edges.
type EdgeFilter func(enum enum.Enum, from, to int) bool

//TODO: create a bunch of EdgeFilters.

//NewGridConnectedness is a helper function to create a finished graph
//representing the connections between a grid. By default it adds edges
//between each of the 8 adjacent cells. However, all neighbors must pass the
//provided filters to be added. This package also defines a number of
//Neighbor* EdgeFilters. The enum passed must be a ranged, 2 dimensional enum.
func NewGridConnectedness(ranged2DEnum enum.Enum, filter ...EdgeFilter) (Graph, error) {
	if !ranged2DEnum.IsRange() {
		return nil, errors.New("The enum was not created with AddRange")
	}
	if len(ranged2DEnum.RangeDimensions()) != 2 {
		return nil, errors.New("The enum did not have two dimensions")
	}

	graph := New(false, ranged2DEnum)

	for _, val := range ranged2DEnum.Values() {

		theNeighbors := neighbors(ranged2DEnum, val)

		for _, theFilter := range filter {
			var tempNeighbors []int
			for _, n := range theNeighbors {
				if theFilter(ranged2DEnum, val, n) {
					tempNeighbors = append(tempNeighbors, n)
				}
			}
			theNeighbors = tempNeighbors
		}

		if err := graph.AddEdges(val, theNeighbors...); err != nil {
			return nil, err
		}

	}

	graph.Finish()

	return graph, nil

}

//assumes that theEnum is a 2d ranged enum, and that start is a valid value in
//it.
func neighbors(theEnum enum.Enum, start int) []int {
	var result []int
	indexes := theEnum.ValueToRange(start)
	for rOffset := -1; rOffset < 2; rOffset++ {
		for cOffset := -1; cOffset < 2; cOffset++ {

			if rOffset == 0 && cOffset == 0 {
				//This is the start cell
				continue
			}

			r := indexes[0] + rOffset
			c := indexes[1] + cOffset

			val := theEnum.RangeToValue(r, c)

			if val > 0 {
				result = append(result, val)
			}

		}
	}
	return result
}
