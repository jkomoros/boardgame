package graph

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
)

//EdgeFilter is a type of function that can be passed to filter in edges. Only
//edges that return true will be kept.
type EdgeFilter func(enum enum.Enum, from, to int) bool

//DirectionUp will return true if to is in a strictly lower-indexed row then
//from.
func DirectionUp(enum enum.Enum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[0] > toIndexes[0]
}

//DirectionDown will return true if to is in a strictly higher-indexed row
//then from.
func DirectionDown(enum enum.Enum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[0] > toIndexes[0]
}

//DirectionLeft will return true if to is in a strictly lower-indexed col then
//from.
func DirectionLeft(enum enum.Enum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[1] > toIndexes[1]
}

//DirectionRight will return true if to is in a strictly higher-indexed col
//then from.
func DirectionRight(enum enum.Enum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[1] > toIndexes[1]
}

//DirectionPerpendicular will return true if to is perpendicular to from (in the
//same row or col).
func DirectionPerpendicular(enum enum.Enum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	if fromIndexes[0] == toIndexes[0] {
		return true
	}
	return fromIndexes[1] == fromIndexes[1]
}

//DirectionDiagonal will return true if to is non-perpendicular to from.
func DirectionDiagonal(enum enum.Enum, from, to int) bool {
	return !DirectionPerpendicular(enum, from, to)
}

//NewGridConnectedness is a helper function to create a finished graph
//representing the connections between a grid. By default it adds edges
//between each of the 8 adjacent cells. However, all neighbors must pass the
//provided filters to be added. This package also defines a number of
//Direction* EdgeFilters. The enum passed must be a ranged, 2 dimensional enum.
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
