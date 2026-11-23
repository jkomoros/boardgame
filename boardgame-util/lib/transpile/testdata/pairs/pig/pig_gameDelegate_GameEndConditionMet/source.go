func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, players := concreteStates(state)

	for _, player := range players {
		if player.Score >= game.TargetScore {
			return true
		}
	}

	return false
}