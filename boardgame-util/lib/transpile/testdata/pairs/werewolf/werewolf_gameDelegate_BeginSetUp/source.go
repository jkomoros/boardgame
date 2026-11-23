func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	_, players := concreteStates(state)

	// Initialize all player votes to -1 (no vote).
	for _, p := range players {
		p.DayVote = -1
		p.NightVote = -1
	}

	return nil
}