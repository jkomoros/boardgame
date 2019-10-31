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

//RoundRobinLastPlayer returns the value set via SetRoundRobinLastPlayer.
func (r *RoundRobinGameStateProperties) RoundRobinLastPlayer() boardgame.PlayerIndex {
	return r.RRLastPlayer
}

//RoundRobinStarterPlayer returns the value set via SetRoundRobinStarterPlayer
func (r *RoundRobinGameStateProperties) RoundRobinStarterPlayer() boardgame.PlayerIndex {
	return r.RRStarterPlayer
}

//RoundRobinRoundCount returns the value set via SetRoundRobinRoundCount
func (r *RoundRobinGameStateProperties) RoundRobinRoundCount() int {
	return r.RRRoundCount
}

//RoundRobinHasStarted returns the value set via SetRoundRobinHasStarted
func (r *RoundRobinGameStateProperties) RoundRobinHasStarted() bool {
	return r.RRHasStarted
}

//SetRoundRobinLastPlayer sets the value to return for RoundRobinLastPlayer
func (r *RoundRobinGameStateProperties) SetRoundRobinLastPlayer(nextPlayer boardgame.PlayerIndex) {
	r.RRLastPlayer = nextPlayer
}

//SetRoundRobinStarterPlayer sets the value to return for RoundRobinStarterPlayer
func (r *RoundRobinGameStateProperties) SetRoundRobinStarterPlayer(index boardgame.PlayerIndex) {
	r.RRStarterPlayer = index
}

//SetRoundRobinRoundCount sets the value to return for RoundrobinRoundCount
func (r *RoundRobinGameStateProperties) SetRoundRobinRoundCount(count int) {
	r.RRRoundCount = count
}

//SetRoundRobinHasStarted sets the value to return for RoundRobinHasStarted.
func (r *RoundRobinGameStateProperties) SetRoundRobinHasStarted(val bool) {
	r.RRHasStarted = val
}
