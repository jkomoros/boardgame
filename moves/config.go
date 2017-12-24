package moves

import (
	"github.com/jkomoros/boardgame"
)

//CustomConfigurationOption is a function that takes a PropertyCollection and
//modifies a key on it. This package defines a number of functions that
//satisfy this interface and can be used in DefaultConfig to pass in
//configuration to the base moves without requiring verbose embedding and
//method overriding. All of those functions in this package start with "With".
type CustomConfigurationOption func(boardgame.PropertyCollection)
