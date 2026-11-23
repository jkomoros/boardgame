func (m *moveMoveBetweenHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.VisibleStack.NumComponents() < 5 {
		return game.HiddenStack.First().MoveTo(game.VisibleStack, 2)
	}

	return game.VisibleStack.ComponentAt(2).MoveToFirstSlot(game.HiddenStack)

}