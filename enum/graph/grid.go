package graph

import (
	"errors"
	"fmt"
	"github.com/jkomoros/boardgame/enum"
	"strconv"
)

//EdgeFilter is a type of function that can be passed to filter in edges. Only
//edges that return true will be kept. This package defines a large number of
//them, all of which start with "Direction".
type EdgeFilter func(enum enum.RangeEnum, from, to int) bool

//DirectionUp will return true if to is in a strictly lower-indexed row then
//from.
func DirectionUp(enum enum.RangeEnum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[0] > toIndexes[0]
}

//DirectionDown will return true if to is in a strictly higher-indexed row
//then from.
func DirectionDown(enum enum.RangeEnum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[0] > toIndexes[0]
}

//DirectionLeft will return true if to is in a strictly lower-indexed col then
//from.
func DirectionLeft(enum enum.RangeEnum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[1] > toIndexes[1]
}

//DirectionRight will return true if to is in a strictly higher-indexed col
//then from.
func DirectionRight(enum enum.RangeEnum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	return fromIndexes[1] > toIndexes[1]
}

//DirectionPerpendicular will return true if to is perpendicular to from (in the
//same row or col).
func DirectionPerpendicular(enum enum.RangeEnum, from, to int) bool {
	fromIndexes := enum.ValueToRange(from)
	toIndexes := enum.ValueToRange(to)
	if fromIndexes[0] == toIndexes[0] {
		return true
	}
	return fromIndexes[1] == toIndexes[1]
}

//DirectionDiagonal will return true if to is non-perpendicular to from.
func DirectionDiagonal(enum enum.RangeEnum, from, to int) bool {
	return !DirectionPerpendicular(enum, from, to)
}

//MustNewGridConnectedness is like NewGridConnectedness, but if it would have
//returned an error, it panics instead. Only appropriate to be called during
//setup.
func MustNewGridConnectedness(ranged2DEnum enum.RangeEnum, filter ...EdgeFilter) Graph {
	graph, err := NewGridConnectedness(ranged2DEnum, filter...)

	if err != nil {
		panic(err.Error())
	}

	return graph
}

//NewGridConnectedness is a helper function to create a finished graph
//representing the connections between a grid. By default it adds edges
//between each of the 8 adjacent cells. However, all neighbors must pass the
//provided filters to be added. This package also defines a number of
//Direction* EdgeFilters. The enum passed must be a ranged, 2 dimensional enum.
//	//Returns a graph that has all cells connected to each of their neighbors.
//	NewGridConnectedness(e)
//
//  //Returns a graph that creates connections upward and diagonally from each
//  //cell.
//	NewGridConnectedness(e, DirectionUp, DirectionDiagonal)
//
func NewGridConnectedness(ranged2DEnum enum.RangeEnum, filter ...EdgeFilter) (Graph, error) {
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
			return nil, errors.New("Couldn't add " + strconv.Itoa(val) + ": " + fmt.Sprintf("%v", theNeighbors) + ": " + err.Error())
		}

	}

	graph.Finish()

	return graph, nil

}

//assumes that theEnum is a 2d ranged enum, and that start is a valid value in
//it.
func neighbors(theEnum enum.RangeEnum, start int) []int {
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

			if val != enum.IllegalValue {
				result = append(result, val)
			}

		}
	}
	return result
}
