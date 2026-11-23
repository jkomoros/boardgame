func (m *moveMoveCardBetweenShortStacks) DefaultsForState(state boardgame.ImmutableState) {
	gameState, _ := concreteStates(state)

	if gameState.FirstShortStack.NumComponents() < 1 {
		m.FromFirst = false
	} else {
		m.FromFirst = true
	}
}