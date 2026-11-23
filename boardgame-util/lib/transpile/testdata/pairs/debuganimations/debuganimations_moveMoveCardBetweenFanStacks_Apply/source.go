func (m *moveMoveCardBetweenFanStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() < 6 {
		return game.FanDiscard.First().MoveTo(game.FanStack, 2)
	}

	return game.FanStack.ComponentAt(2).MoveToFirstSlot(game.FanDiscard)
}