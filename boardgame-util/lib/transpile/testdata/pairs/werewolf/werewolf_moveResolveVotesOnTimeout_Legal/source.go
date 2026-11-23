func (m *moveResolveVotesOnTimeout) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	if proposer != boardgame.AdminPlayerIndex {
		return errors.New("only the timer may resolve incomplete votes")
	}
	game, _ := concreteStates(state)
	if !timedVotingEnabled(state) || !game.VoteTimer.Active() {
		return errors.New("timed vote resolution is not active")
	}
	return nil
}