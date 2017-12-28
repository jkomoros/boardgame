/*

	graph is a simple package that provides facilities for creating simple
	graphs where each Node is a particular value in an enum.

	graph is useful for modeling adjacency of spaces in a gameboard.

*/
package graph

import (
	"github.com/jkomoros/boardgame/enum"
)

type Graph interface {
	AddEdge(from, to int) error
	Connected(from, to int) bool
	Neighbors(start int) []int

	//Defaults to 0 for edges that haven't had SetEdgeWeight called.
	EdgeWeight(from, to int) int
	SetEdgeWeight(from, to int, weight int) error

	//After finish is called, no modifications may be made to the graph.
	Finish()
}

type graph struct {
	directed    bool
	edges       map[int]map[int]bool
	edgeWeights map[string]int
}

func New(directed bool, enum enum.Enum) Graph {
	return nil
}
