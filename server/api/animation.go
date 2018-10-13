package api

import (
	"github.com/jkomoros/boardgame"
	"time"
)

//If moves in your game implement PreAnimationDelay, the server will use that
//information to inject non-semantic animation delays on the client. All moves
//derived from moves.Base implement this interface. See TUTORIAL.md for more
//on animation delays.
type PreAnimationDelayer interface {
	PreAnimationDelay() time.Duration
}

//If moves in your game implement PreAnimationDelay, the server will use that
//information to inject non-semantic animation delays on the client. All moves
//derived from moves.Base implement this interface. See TUTORIAL.md for more
//on animation delays.
type PostAnimationDelayer interface {
	PostAnimationDelay() time.Duration
}

func configureDelays(manager *boardgame.GameManager) (preDelays map[string]time.Duration, postDelays map[string]time.Duration) {

	preDelays = make(map[string]time.Duration)
	postDelays = make(map[string]time.Duration)

	for _, move := range manager.ExampleMoves() {

		var preDelay time.Duration
		var postDelay time.Duration

		if preDelayer, ok := move.(PreAnimationDelayer); ok {
			preDelay = preDelayer.PreAnimationDelay()
		}

		if postDelayer, ok := move.(PostAnimationDelayer); ok {
			postDelay = postDelayer.PostAnimationDelay()
		}

		name := move.Info().Name()

		preDelays[name] = preDelay
		postDelays[name] = postDelay
	}

	return preDelays, postDelays

}
