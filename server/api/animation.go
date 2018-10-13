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

	currentInitiator := moves[0].Initiator

	for i, move := range moves {

		//Slide along until we find the last move in an initator chain, or the last move
		if move.Initiator == currentInitiator && i != len(moves)-1 {
			continue
		}

		//This is the state for the end of the bundle.
		state := game.State(move.Version)

		if autoCurrentPlayer {
			newPlayerIndex := game.Manager().Delegate().CurrentPlayerIndex(state)
			if newPlayerIndex.Valid(state) {
				playerIndex = newPlayerIndex
			}
		}

		delay := 500

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
	}

	return bundles
}
