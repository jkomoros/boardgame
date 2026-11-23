func (m *moveCollectCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	for _, p := range state.ImmutablePlayerStates() {
		player := p.(*playerState)
		if player.HiddenHand.NumComponents() > 0 || player.VisibleHand.NumComponents() > 0 {
			return nil
		}
	}
	return errors.New("no player has cards to collect")
}