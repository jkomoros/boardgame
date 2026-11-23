func (m *moveIncrementRoundsCompleted) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	game, _ := concreteStates(state)
	// Legal only if no player has cards and flags are reset (cleanup already done)
	for _, p := range state.ImmutablePlayerStates() {
		player := p.(*playerState)
		if player.HiddenHand.NumComponents() > 0 || player.VisibleHand.NumComponents() > 0 {
			return errors.New("cards haven't been collected yet")
		}
	}
	// Check that we haven't already incremented (by looking at whether we're
	// still in the cleanup phase — the StartPhase move after us will change it)
	if game.Phase.Value() != phaseRoundCleanup {
		return errors.New("not in cleanup phase")
	}
	return nil
}