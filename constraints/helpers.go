package constraints

import (
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

// intEffectiveValue parses a string as an integer, or looks it up as a named
// constant in the chest.
func intEffectiveValue(str string, chest *boardgame.ComponentChest) (int, error) {
	if chest != nil {
		val := chest.Constant(str)
		if val != nil {
			i, ok := val.(int)
			if !ok {
				return 0, errors.New(str + " is a constant, but not of type int")
			}
			return i, nil
		}
	}

	intVal, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.New(str + " is not convertible to int: " + err.Error())
	}
	return intVal, nil
}
