package memory

import (
	"github.com/jkomoros/boardgame"
)

type Agent struct{}

type agentCardInfo struct {
	Value string
	Index int
}

type agentState struct {
	LastCards    []agentCardInfo
	MemoryLength int
}

func (a *Agent) Name() string {
	return "ai"
}

func (a *Agent) DisplayName() string {
	return "Robby the Robot"
}

func (a *Agent) SetUpForGame(game *boardgame.Game, player boardgame.PlayerIndex) []byte {
	//TODO: do something
	return nil
}

func (a *Agent) ProposeMove(game *boardgame.Game, player boardgame.PlayerIndex, agentState []byte) (move boardgame.Move, newState []byte) {
	//TODO: do something
	return nil, nil
}

//CardSeen is called when a card is visible. If will return true if that was
//new information, or false if not.
func (a *agentState) CardSeen(value string, index int) bool {

	//Is this card currently in the known set of cards?
	for _, card := range a.LastCards {
		if card.Value == value && card.Index == index {
			return false
		}
	}

	//Add it to the list.

	info := agentCardInfo{
		Value: value,
		Index: index,
	}

	a.LastCards = append([]agentCardInfo{info}, a.LastCards...)

	if len(a.LastCards) > a.MemoryLength {
		//Trim it if there are more cards than we can remember at once.
		a.LastCards = a.LastCards[:a.MemoryLength]
	}

	return true

}
