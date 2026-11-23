func (m *moveCurrentPlayerHit) Apply(state boardgame.State) error {

	if err := m.DrawToPlayer.Apply(state); err != nil {
		return err
	}

	_, players := concreteStates(state)

	currentPlayer := players[state.CurrentPlayerIndex()]

	handValue := currentPlayer.HandValue()

	if handValue > targetScore {
		currentPlayer.Eliminated = true
	}

	if handValue == targetScore {
		currentPlayer.Stood = true
	}

	return nil
}