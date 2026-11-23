func (m *moveShuffleHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if err := game.FanDiscard.Shuffle(); err != nil {
		return err
	}

	game.FanShuffleCount++

	return nil
}