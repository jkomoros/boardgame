func startVoteTimer(state boardgame.State) error {
	if !timedVotingEnabled(state) {
		return nil
	}
	game, _ := concreteStates(state)
	moveName := "Resolve Day Votes on Timeout"
	if game.Phase.Value() == phaseNight {
		moveName = "Resolve Night Votes on Timeout"
	}
	completion := state.Game().MoveByNameForState(moveName, state)
	if completion == nil {
		return errors.New("timed vote resolver was not configured")
	}
	game.VoteTimer.Start(voteDuration, completion)
	return nil
}