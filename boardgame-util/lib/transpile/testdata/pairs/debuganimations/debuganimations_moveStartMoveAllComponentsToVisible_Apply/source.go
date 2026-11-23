func (m *moveStartMoveAllComponentsToVisible) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	for game.AllHiddenStack.NumComponents() > 0 {
		if err := game.AllHiddenStack.First().MoveToNextSlot(game.AllVisibleStack); err != nil {
			return err
		}
	}

	return nil
}