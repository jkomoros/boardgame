func (m *moveAccumulateScores) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	// Only legal if at least one player has cards in hand (scores not yet collected)
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if player.Hand.NumComponents() > 0 {
			return nil
		}
	}
	return errors.New("no active players have cards to score")
}