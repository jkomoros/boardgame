func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {

	game, players := concreteStates(state)

	tokens := make([]string, totalDim)

	for i := 0; i < totalDim; i++ {
		tokens[i] = game.tokenValueAtIndex(i)
	}

	finished, winner := checkGameFinished(tokens)

	if finished {

		if winner == Empty {
			//Draw
			return true, nil
		}

		var winningPlayer int

		for i, player := range players {
			if player.TokenValue == winner {
				winningPlayer = i
			}
		}

		return true, []boardgame.PlayerIndex{boardgame.PlayerIndex(winningPlayer)}
	}

	return false, nil

}