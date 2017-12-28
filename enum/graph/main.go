/*

	graph is a simple package that provides facilities for creating simple
	graphs where each Node is a particular value in an enum.

	graph is useful for modeling adjacency of spaces in a gameboard.

*/
package graph

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
	"strconv"
)

type Graph interface {
	//AddEdge adds the edge to the graph if it doesn't exist, and if the graph
	//isn't finished yet. Will error if from or to aren't in the given enum.
	AddEdge(from, to int) error
	Connected(from, to int) bool
	Neighbors(start int) []int

	//Defaults to 0 for edges that haven't had SetEdgeWeight called.
	EdgeWeight(from, to int) int
	//SetEdgeWeight sets the weight between the two nodes. Errors if the graph
	//is already finished, or if those two nodes aren't connected.
	SetEdgeWeight(from, to int, weight int) error

	//After finish is called, no modifications may be made to the graph.
	Finish()
}

type graph struct {
	undirected  bool
	finished    bool
	theEnum     enum.Enum
	edges       map[int]map[int]bool
	edgeWeights map[string]int
}

//New returns a new, unfinished graph based on the given enum, where each node
//in the graph is one of the values in the Enum. If undirected is true, then
//adding an edge from -> to also adds the edge to -> from automatically.
func New(undirected bool, enum enum.Enum) Graph {
	return &graph{
		undirected,
		false,
		enum,
		make(map[int]map[int]bool, len(enum.Values())),
		make(map[string]int),
	}
}

func (g *graph) Finish() {
	g.finished = true
}

func (g *graph) AddEdge(from, to int) error {
	if err := g.addEdgeImpl(from, to); err != nil {
		return err
	}
	if g.undirected {
		return g.addEdgeImpl(to, from)
	}
	return nil
}

func (g *graph) addEdgeImpl(from, to int) error {
	if !g.theEnum.Valid(from) {
		return errors.New("from value is not legal in that enum")
	}
	if !g.theEnum.Valid(to) {
		return errors.New("to value is not legal in that enum")
	}
	if g.finished {
		return errors.New("graph is finished so no modifications may be made")
	}
	edgeMap := g.edges[from]
	if edgeMap == nil {
		edgeMap = make(map[int]bool)
		g.edges[from] = edgeMap
	}
	edgeMap[to] = true
	return nil
}

func (g *graph) Connected(from, to int) bool {
	edgeMap := g.edges[from]
	if edgeMap == nil {
		return false
	}
	return edgeMap[to]
}

func (g *graph) Neighbors(start int) []int {
	edgeMap := g.edges[start]
	if edgeMap == nil {
		return nil
	}
	result := make([]int, len(edgeMap))
	counter := 0
	for key, _ := range edgeMap {
		result[counter] = key
		counter++
	}
	return result
}

func keyForEdge(from, to int) string {
	return strconv.Itoa(from) + "-" + strconv.Itoa(to)
}

func (g *graph) EdgeWeight(from, to int) int {
	//If the edge doesn't exist, the default of 0 is fine
	return g.edgeWeights[keyForEdge(from, to)]
}

func (g *graph) SetEdgeWeight(from, to int, weight int) error {
	if !g.Connected(from, to) {
		return errors.New("from and to do not share an edge")
	}
	if g.finished {
		return errors.New("graph is finished so no modifications may be made")
	}
	g.edgeWeights[keyForEdge(from, to)] = weight
	return nil
}
