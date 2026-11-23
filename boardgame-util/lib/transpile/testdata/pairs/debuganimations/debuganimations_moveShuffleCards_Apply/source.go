func (m *moveShuffleCards) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	return game.FanStack.Shuffle()

}