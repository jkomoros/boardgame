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
//Equivalent to Between(0,1).
func Any() interfaces.ValidCounter {
	return anyFunc
}

//All will return nil if currentCount is precisely the same length as length.
//Equivalent to Between(0,-1).
func All() interfaces.ValidCounter {
	return allFunc
}

//AtLeast will return nil if currentCount is min or greater.
func AtLeast(min int) interfaces.ValidCounter {
	return func(currentCount, length int) error {
		if currentCount >= min {
			return nil
		}
		return errors.New("currentCount not yet greater than min configuration")
	}
}

//AtMost will return nil as long as currentCount is less than or equal to max. A
//max argument of less than 0 will be interpreted to mean precisely the length
//parameter passed into ValidCounter.
func AtMost(max int) interfaces.ValidCounter {
	return func(currentCount, length int) error {
		if max < 0 {
			max = length
		}
		if currentCount <= max {
			return nil
		}
		return errors.New("currentCount is greater than max configuration")
	}
}

//Between returns nil as long as the value is greater than or equal to min and
//less than or equal to max. A max argument of less than 0 will be interpreted
//to mean precise the length parameter passed into ValidCounter.
func Between(min, max int) interfaces.ValidCounter {
	return func(currentCount, length int) error {
		if max < 0 {
			max = length
		}
		if currentCount < min {
			return errors.New("Count below min")
		}
		if currentCount > max {
			return errors.New("Count above max")
		}
		return nil
	}
}

//Exactly returns nil if currentCount is precisely equaly to targetCount.
//Equivalent to Between(targetCount,targetCount).
func Exactly(targetCount int) interfaces.ValidCounter {
	return func(currentCount, length int) error {
		if targetCount == currentCount {
			return nil
		}
		if targetCount > currentCount {
			return errors.New("currentCount is not yet targetCount")
		}
		return errors.New("currentCount has already passed targetCount")
	}
}
