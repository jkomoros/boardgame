/*

	roundrobinhelpers is a package that contains structs that are useful to
	use with RoundRobin moves.

*/
package roundrobinhelpers

import (
	"github.com/jkomoros/boardgame"
)

//BaseGameState is designed to be embedded in your GameState anonymously to
//automatically satisfy the RoundRobinProperties interface, making it easy to
//use RoundRobin-basd moves. Because this embeds boardgame.BaseSubState
//itself, you should embed this INSTEAD of boardgame.BaseSubState.
type BaseGameState struct {
	boardgame.BaseSubState
	RRLastPlayer    boardgame.PlayerIndex
	RRStarterPlayer boardgame.PlayerIndex
	RRRoundCount    int
	RRHasStarted    bool
}

func (r *BaseGameState) RoundRobinLastPlayer() boardgame.PlayerIndex {
	return r.RRLastPlayer
}

func (r *BaseGameState) RoundRobinStarterPlayer() boardgame.PlayerIndex {
	return r.RRStarterPlayer
}

func (r *BaseGameState) RoundRobinRoundCount() int {
	return r.RRRoundCount
}

func (r *BaseGameState) RoundRobinHasStarted() bool {
	return r.RRHasStarted
}

func (r *BaseGameState) SetRoundRobinLastPlayer(nextPlayer boardgame.PlayerIndex) {
	r.RRLastPlayer = nextPlayer
}

func (r *BaseGameState) SetRoundRobinStarterPlayer(index boardgame.PlayerIndex) {
	r.RRStarterPlayer = index
}

func (r *BaseGameState) SetRoundRobinRoundCount(count int) {
	r.RRRoundCount = count
}

func (r *BaseGameState) SetRoundRobinHasStarted(val bool) {
	r.RRHasStarted = val
}
