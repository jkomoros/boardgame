func (g *gameDelegate) FinishSetUp(state boardgame.State) error {

	game, _ := concreteStates(state)

	//Pick a player to start randomly.
	startingPlayer := boardgame.PlayerIndex(state.Rand().Intn(len(state.PlayerStates())))

	game.CurrentPlayer = startingPlayer
	game.TargetScore = defaultTargetScore

	return nil

}