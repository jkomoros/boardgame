func (m *moveFlipHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.HiddenCard.NumComponents() < 1 && game.VisibleCard.NumComponents() < 1 {
		return errors.New("Neither the HiddenCard nor RevealedCard is set")
	}

	if game.HiddenCard.NumComponents() > 0 && game.VisibleCard.NumComponents() > 0 {
		return errors.New("both hidden and revealed are full")
	}

	return nil
}