func finishVoteResolution(state boardgame.State) error {
	game, _ := concreteStates(state)
	game.VoteTimer.Cancel()
	if err := resolveVotes(state); err != nil {
		return err
	}
	if finished, _ := state.Manager().Delegate().CheckGameFinished(state); finished {
		return nil
	}
	return startVoteTimer(state)
}