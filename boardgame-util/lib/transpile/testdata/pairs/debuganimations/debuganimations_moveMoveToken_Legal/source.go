func (m *moveMoveToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.TokensFrom.NumComponents() == 10 && game.TokensTo.NumComponents() == 9 {
		return nil
	}

	if game.TokensFrom.NumComponents() == 9 && game.TokensTo.NumComponents() == 10 {
		return nil
	}

	return errors.New("tokens aren't in known position")
}