package moves

import (
	"github.com/jkomoros/boardgame"
)

/*
RoundRobinGameStateProperties is designed to be embedded in your GameState anonymously to
automatically satisfy the RoundRobinProperties interface, making it easy to
use RoundRobin-basd moves. You typically embed this IN ADDITION TO base.SubState.

	//Example
	type gameState struct {
		base.SubState
		//By including this, we can use any round-robin based move without any
		//other changes
		moves.RoundRobinGameStateProperties
		MyInt int
	}

*/
type RoundRobinGameStateProperties struct {
	RRLastPlayer    boardgame.PlayerIndex
	RRStarterPlayer boardgame.PlayerIndex
	RRRoundCount    int
	RRHasStarted    bool
}

func (r *RoundRobinGameStateProperties) RoundRobinLastPlayer() boardgame.PlayerIndex {
	return r.RRLastPlayer
}

func (r *RoundRobinGameStateProperties) RoundRobinStarterPlayer() boardgame.PlayerIndex {
	return r.RRStarterPlayer
}

func (r *RoundRobinGameStateProperties) RoundRobinRoundCount() int {
	return r.RRRoundCount
}

func (r *RoundRobinGameStateProperties) RoundRobinHasStarted() bool {
	return r.RRHasStarted
}

func (r *RoundRobinGameStateProperties) SetRoundRobinLastPlayer(nextPlayer boardgame.PlayerIndex) {
	r.RRLastPlayer = nextPlayer
}

func (r *RoundRobinGameStateProperties) SetRoundRobinStarterPlayer(index boardgame.PlayerIndex) {
	r.RRStarterPlayer = index
}

func (r *RoundRobinGameStateProperties) SetRoundRobinRoundCount(count int) {
	r.RRRoundCount = count
}

func (r *RoundRobinGameStateProperties) SetRoundRobinHasStarted(val bool) {
	r.RRHasStarted = val
}
