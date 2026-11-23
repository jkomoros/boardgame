func (m *moveCastVote) Apply(state boardgame.State) error {
	game, players := concreteStates(state)
	voter := players[m.TargetPlayerIndex]
	if game.Phase.Value() == phaseNight {
		voter.NightVote = m.VoteTarget
	} else {
		voter.DayVote = m.VoteTarget
	}
	return nil
}