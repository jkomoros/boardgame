func (m *moveCurrentPlayerHit) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	_, players := concreteStates(state)

	currentPlayer := players[state.CurrentPlayerIndex()]

	if currentPlayer.HandValue() >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	return nil
}