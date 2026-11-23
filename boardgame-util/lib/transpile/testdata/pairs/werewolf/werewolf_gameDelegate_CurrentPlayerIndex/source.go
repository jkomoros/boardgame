func (g *gameDelegate) CurrentPlayerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	phase := game.Phase.Value()
	if phase == phaseDay || phase == phaseNight {
		// Simultaneous voting: any player may act.
		return boardgame.AnyPlayerIndex
	}
	// During gathering, no current player.
	return boardgame.ObserverPlayerIndex
}