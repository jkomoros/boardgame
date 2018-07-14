/*

	count defines a collection of interfaces.ValidCounter for use in
	moves/groups/ParallelCount and friends.

*/
package count

import (
	"errors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//anyFunc will be returned for Any. Since there are no values to close over,
//we can return the same item each time.
func anyFunc(currentCount, length int) error {
	switch currentCount {
	case 0:
		return errors.New("Not enough count have occurred")
	case 1:
		return nil
	default:
		return errors.New("Too many count have occurred.")
	}
}

//allFunc will be returned from All. Since there are no values to close over,
//we can return the same item each time and avoid memory allocation.
func allFunc(currentCount, length int) error {
	if currentCount < length {
		return errors.New("Not enough count have occurred")
	} else if currentCount == length {
		return nil
	}

	return errors.New("Too many count have occurred.")
}

//Any will return nil if currentCount is 1, denoting that any item has matched.
//Equivalent to MinMax(0,1).
func Any() interfaces.ValidCounter {
	return anyFunc
}

//All will return nil if currentCount is precisely the same length as length.
//Equivalent to MinMax(0,-1).
func All() interfaces.ValidCounter {
	return allFunc
}
