func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, _ := concreteStates(state)
	return game.RoundsCompleted >= game.MaxRounds
}