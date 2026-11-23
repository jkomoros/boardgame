func (m *moveMoveTokenSanitized) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.SanitizedTokensFrom.NumComponents() == 10 && game.SanitizedTokensTo.NumComponents() == 9 {
		return nil
	}

	if game.SanitizedTokensFrom.NumComponents() == 9 && game.SanitizedTokensTo.NumComponents() == 10 {
		return nil
	}

	return errors.New("tokens aren't in known position")
}