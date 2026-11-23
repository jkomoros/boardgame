func (m *moveMoveCardBetweenFanStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() == 6 && game.FanDiscard.NumComponents() == 3 {
		return nil
	}

	if game.FanStack.NumComponents() == 5 && game.FanDiscard.NumComponents() == 4 {
		return nil
	}

	return errors.New("Fan stacks aren't in known toggle state")
}