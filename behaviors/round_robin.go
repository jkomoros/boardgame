package behaviors

import (
	"github.com/jkomoros/boardgame"
)

/*
RoundRobin is designed to be embedded in your GameState anonymously to
automatically satisfy the moves/interfaces.RoundRobinProperties interface,
making it easy to use RoundRobin-based moves. You typically embed this IN
ADDITION TO base.SubState.

    //Example
    type gameState struct {
        base.SubState
        //By including this, we can use any round-robin based move without any
        //other changes
        behaviors.RoundRobin
        MyInt int
    }

*/
type RoundRobin struct {
	RRLastPlayer    boardgame.PlayerIndex
	RRStarterPlayer boardgame.PlayerIndex
	RRRoundCount    int
	RRHasStarted    bool
}

//ConnectBehavior doesn't do anything
func (r *RoundRobin) ConnectBehavior(containingStruct boardgame.SubState) {
	//Pass
}

//RoundRobinLastPlayer returns the value set via SetRoundRobinLastPlayer.
func (r *RoundRobin) RoundRobinLastPlayer() boardgame.PlayerIndex {
	return r.RRLastPlayer
}

//RoundRobinStarterPlayer returns the value set via SetRoundRobinStarterPlayer
func (r *RoundRobin) RoundRobinStarterPlayer() boardgame.PlayerIndex {
	return r.RRStarterPlayer
}

//RoundRobinRoundCount returns the value set via SetRoundRobinRoundCount
func (r *RoundRobin) RoundRobinRoundCount() int {
	return r.RRRoundCount
}

//RoundRobinHasStarted returns the value set via SetRoundRobinHasStarted
func (r *RoundRobin) RoundRobinHasStarted() bool {
	return r.RRHasStarted
}

//SetRoundRobinLastPlayer sets the value to return for RoundRobinLastPlayer
func (r *RoundRobin) SetRoundRobinLastPlayer(nextPlayer boardgame.PlayerIndex) {
	r.RRLastPlayer = nextPlayer
}

//SetRoundRobinStarterPlayer sets the value to return for RoundRobinStarterPlayer
func (r *RoundRobin) SetRoundRobinStarterPlayer(index boardgame.PlayerIndex) {
	r.RRStarterPlayer = index
}

//SetRoundRobinRoundCount sets the value to return for RoundrobinRoundCount
func (r *RoundRobin) SetRoundRobinRoundCount(count int) {
	r.RRRoundCount = count
}

//SetRoundRobinHasStarted sets the value to return for RoundRobinHasStarted.
func (r *RoundRobin) SetRoundRobinHasStarted(val bool) {
	r.RRHasStarted = val
}
