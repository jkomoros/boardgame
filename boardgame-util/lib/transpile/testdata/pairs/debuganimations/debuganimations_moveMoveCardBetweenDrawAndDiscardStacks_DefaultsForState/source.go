func (m *moveMoveCardBetweenDrawAndDiscardStacks) DefaultsForState(state boardgame.ImmutableState) {
	gameState, _ := concreteStates(state)

	if gameState.DiscardStack.NumComponents() < 3 {
		m.FromDraw = true
	} else {
		m.FromDraw = false
	}
}