package graph

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
	"sort"
	"testing"
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

	sort.Ints(n)

	assert.For(t).ThatActual(n).Equals([]int{0, 5})

	//0,4 is not a valid index
	err = graph.AddEdge(e.RangeToValue(0, 1), e.RangeToValue(0, 4))

	assert.For(t).ThatActual(err).IsNotNil()

	graph.Finish()

	//the graph has been finished so no modifications may be made
	err = graph.AddEdge(e.RangeToValue(0, 1), e.RangeToValue(2, 0))

	assert.For(t).ThatActual(err).IsNotNil()
}
