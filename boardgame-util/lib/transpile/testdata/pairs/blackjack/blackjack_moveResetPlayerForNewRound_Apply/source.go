func (m *moveResetPlayerForNewRound) Apply(state boardgame.State) error {
	_, players := concreteStates(state)
	for _, p := range players {
		p.Eliminated = false
		p.Stood = false
	}
	return nil
}