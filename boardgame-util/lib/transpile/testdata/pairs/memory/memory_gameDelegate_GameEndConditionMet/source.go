func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, _ := concreteStates(state)

	if game.Cards.NumComponents() > 0 {
		return false
	}

	return true
}