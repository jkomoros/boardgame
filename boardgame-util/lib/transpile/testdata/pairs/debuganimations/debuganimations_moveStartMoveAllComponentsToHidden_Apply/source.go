func (m *moveStartMoveAllComponentsToHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	for game.AllVisibleStack.NumComponents() > 0 {
		if err := game.AllVisibleStack.First().MoveToNextSlot(game.AllHiddenStack); err != nil {
			return err
		}
	}

	return nil
}