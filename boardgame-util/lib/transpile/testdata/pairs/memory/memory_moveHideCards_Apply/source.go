func (m *moveHideCards) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	//Cancel a timer in case it was still going.
	game.HideCardsTimer.Cancel()

	for i, c := range game.VisibleCards.Components() {
		if c != nil {
			if err := c.MoveTo(game.HiddenCards, i); err != nil {
				return errors.New("Couldn't move component: " + err.Error())
			}
		}
	}

	return nil
}