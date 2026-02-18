/*

Package graph is a simple package that provides facilities for creating simple
graphs where each Node is a particular value in an enum.

graph is useful for modeling adjacency of spaces in a gameboard. In addition to
basic adjacency queries (Connected, Neighbors), it supports weighted shortest
path computation via ShortestPath and Distance, which use Dijkstra's algorithm.
Edge weights default to 1 when not explicitly set via SetEdgeWeight, so
unweighted graphs behave as BFS.

NewGridConnectedness is a graph creator that connects all spaces in a grid that
are neighbors, with the ability to filter to only include some types of
neighbors.

See also behaviors.LocationBehavior and moves.MoveOnGraph, which build on
graphs to provide reusable spatial game mechanics.

*/
package graph

import (
	"container/heap"
	"errors"
	"strconv"

	"github.com/jkomoros/boardgame/enum"
)

//Graph is the primary type of this package. It represents a directed graph
//where the nodes are all values in an enum.
type Graph interface {
	//AddEdge adds the edge to the graph if it doesn't exist, and if the graph
	//isn't finished yet. Will error if from or to aren't in the given enum.
	AddEdge(from, to int) error
	//AddEdges is a convenience wrapper around AddEdge, with multiple to
	//nodes. Will error if adding any errors.
	AddEdges(from int, to ...int) error
	Connected(from, to int) bool
	Neighbors(start int) []int

	//Defaults to 0 for edges that haven't had SetEdgeWeight called.
	EdgeWeight(from, to int) int
	//SetEdgeWeight sets the weight between the two nodes. Errors if the graph
	//is already finished, or if those two nodes aren't connected.
	SetEdgeWeight(from, to int, weight int) error

	//ShortestPath returns the shortest path from start to end (inclusive of
	//both endpoints) using Dijkstra's algorithm. Edge weights default to 1 if
	//not explicitly set via SetEdgeWeight. Returns nil and an error if no path
	//exists or if start/end are not valid nodes.
	ShortestPath(start, end int) ([]int, error)

	//Distance returns the total weight of the shortest path between start and
	//end. Returns -1 and an error if no path exists or if start/end are not
	//valid nodes.
	Distance(start, end int) (int, error)

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

func (g *graph) AddEdges(from int, to ...int) error {
	for i, item := range to {
		if err := g.AddEdge(from, item); err != nil {
			return errors.New("Couldn't add " + strconv.Itoa(i) + " edge: " + err.Error())
		}
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
	for key := range edgeMap {
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

func (g *graph) effectiveWeight(from, to int) int {
	w := g.EdgeWeight(from, to)
	if w == 0 {
		return 1
	}
	return w
}

func (g *graph) ShortestPath(start, end int) ([]int, error) {
	if !g.theEnum.Valid(start) {
		return nil, errors.New("start is not a valid node in the graph")
	}
	if !g.theEnum.Valid(end) {
		return nil, errors.New("end is not a valid node in the graph")
	}
	if start == end {
		return []int{start}, nil
	}

	const inf = int(^uint(0) >> 1) // max int

	dist := make(map[int]int)
	prev := make(map[int]int)
	visited := make(map[int]bool)

	for _, v := range g.theEnum.Values() {
		dist[v] = inf
	}
	dist[start] = 0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{node: start, dist: 0})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		u := item.node

		if visited[u] {
			continue
		}
		visited[u] = true

		if u == end {
			break
		}

		for _, v := range g.Neighbors(u) {
			if visited[v] {
				continue
			}
			alt := dist[u] + g.effectiveWeight(u, v)
			if alt < dist[v] {
				dist[v] = alt
				prev[v] = u
				heap.Push(pq, &pqItem{node: v, dist: alt})
			}
		}
	}

	if dist[end] == inf {
		return nil, errors.New("no path exists between " + strconv.Itoa(start) + " and " + strconv.Itoa(end))
	}

	// Reconstruct path from end to start
	var path []int
	for cur := end; cur != start; cur = prev[cur] {
		path = append(path, cur)
	}
	path = append(path, start)

	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, nil
}

func (g *graph) Distance(start, end int) (int, error) {
	path, err := g.ShortestPath(start, end)
	if err != nil {
		return -1, err
	}

	total := 0
	for i := 0; i < len(path)-1; i++ {
		total += g.effectiveWeight(path[i], path[i+1])
	}
	return total, nil
}

// pqItem is an item in the priority queue for Dijkstra.
type pqItem struct {
	node  int
	dist  int
	index int // index in the heap
}

// priorityQueue implements heap.Interface for Dijkstra's algorithm.
type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].dist < pq[j].dist }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*pqItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}
