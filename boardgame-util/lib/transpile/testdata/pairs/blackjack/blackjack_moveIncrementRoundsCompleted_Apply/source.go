func (m *moveIncrementRoundsCompleted) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)
	game.RoundsCompleted++
	return nil
}