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

func TestNewGridConnectedness(t *testing.T) {
	set := enum.NewSet()

	e := set.MustAddRange("whatever", 4, 4)

	g, err := NewGridConnectedness(e)

	assert.For(t).ThatActual(err).IsNil()

	n := g.Neighbors(6)

	sort.Ints(n)

	assert.For(t).ThatActual(n).Equals([]int{1, 2, 3, 5, 7, 9, 10, 11})

}
