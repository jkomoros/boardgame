func (m *moveResetPlayerForNewRound) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if player.Eliminated || player.Stood {
			return nil
		}
	}
	return errors.New("no active players need resetting")
}