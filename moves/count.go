package moves

import (
	"errors"
)

//ValidCounter is the signature of objects in the moves/count package. It is
//expected within groups in the move/groups package for items like
//ParallelCount. currentCount is the value of the counter in question, and
//length is the context-specific length of the important item, often the
//number of children in the parrent group. If ValidCounter returns nil, the
//count is considered valid and complete; if it is not valid it should return
//a descriptive error. Typically these functions are closures that close over
//configuration options.
type ValidCounter func(currentCount, length int) error

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

//CountAny will return nil if currentCount is 1, denoting that any item has matched.
//Equivalent to CountBetween(0,1).
func CountAny() ValidCounter {
	return anyFunc
}

//CountAll will return nil if currentCount is precisely the same length as length.
//Equivalent to CountBetween(0,-1).
func CountAll() ValidCounter {
	return allFunc
}

//CountAtLeast will return nil if currentCount is min or greater.
func CountAtLeast(min int) ValidCounter {
	return func(currentCount, length int) error {
		if currentCount >= min {
			return nil
		}
		return errors.New("currentCount not yet greater than min configuration")
	}
}

//CountAtMost will return nil as long as currentCount is less than or equal to max. A
//max argument of less than 0 will be interpreted to mean precisely the length
//parameter passed into ValidCounter.
func CountAtMost(max int) ValidCounter {
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

//CountBetween returns nil as long as the value is greater than or equal to min and
//less than or equal to max. A max argument of less than 0 will be interpreted
//to mean precise the length parameter passed into ValidCounter.
func CountBetween(min, max int) ValidCounter {
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

//CountExactly returns nil if currentCount is precisely equaly to targetCount.
//Equivalent to CountBetween(targetCount,targetCount).
func CountExactly(targetCount int) ValidCounter {
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
