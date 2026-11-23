package main

import (
	"github.com/jkomoros/boardgame/base"
)

// delegateVecs are base.GameDelegate's answers that the loader
// (rules/data.ts) reproduces for a game that does not override the method.
func delegateVecs() map[string]interface{} {
	_, err := (&base.GameDelegate{}).DistributeComponentToStarterStack(nil, nil)
	return map[string]interface{}{
		"distributeComponentToStarterStackError": err.Error(),
	}
}
