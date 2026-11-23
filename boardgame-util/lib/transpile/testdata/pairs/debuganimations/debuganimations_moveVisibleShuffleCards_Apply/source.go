func (m *moveVisibleShuffleCards) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	return game.FanStack.PublicShuffle()

}