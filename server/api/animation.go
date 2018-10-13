package api

import (
	"github.com/gin-gonic/gin"
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

func (s *Server) moveBundles(game *boardgame.Game, moves []*boardgame.MoveStorageRecord, playerIndex boardgame.PlayerIndex, autoCurrentPlayer bool) []gin.H {
	var bundles []gin.H

	animationInfo := s.animationDelays[game.Manager().Delegate().Name()]

	currentInitiator := moves[0].Initiator

	//Delay from the last move. Starts at 0 (no delay)
	var postDelay time.Duration

	for i, move := range moves {

		preDelay := animationInfo.preDelays[move.Name]
		//The current' move's postDelay. Before moving on to the next thing, will set it to most
		nextPostDelay := animationInfo.postDelays[move.Name]

		//We make a bundle edge unless all of the following are true:
		// 1) We aren't the last move in the total run
		// 2) the last move didn't have a postDelay
		// 3) Current move doesn't have a preDelay
		// 4) The move.initiatior is the same as the last move's initiator (no new causal chain)
		if i != len(moves)-1 {
			if postDelay == 0 {
				if preDelay == 0 {
					if move.Initiator == currentInitiator {
						//OK, if all of these things are true, we don't have a bundle break here.
						postDelay = nextPostDelay
						continue
					}
				}
			}
		}

		//This is the state for the end of the bundle.
		state := game.State(move.Version)

		if autoCurrentPlayer {
			newPlayerIndex := game.Manager().Delegate().CurrentPlayerIndex(state)
			if newPlayerIndex.Valid(state) {
				playerIndex = newPlayerIndex
			}
		}

		delay := preDelay
		if postDelay > preDelay {
			delay = postDelay
		}

		if len(bundles) == 0 {
			delay = 0
		}

		//If state is nil, JSONForPlayer will basically treat it as just "give the
		//current version" which is a reasonable fallback.
		bundle := gin.H{
			"Game":            game.JSONForPlayer(playerIndex, state),
			"Move":            move,
			"Delay":           delay,
			"ViewingAsPlayer": playerIndex,
			"Forms":           s.generateForms(game),
		}

		bundles = append(bundles, bundle)

		currentInitiator = move.Initiator
		postDelay = nextPostDelay
	}

	return bundles
}
