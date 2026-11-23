func (m *moveAccumulateScores) Apply(state boardgame.State) error {
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if !player.Eliminated {
			player.Score += player.HandValue()
		}
	}
	return nil
}