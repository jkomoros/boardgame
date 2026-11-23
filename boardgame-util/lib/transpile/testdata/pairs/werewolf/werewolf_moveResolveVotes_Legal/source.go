func (m *moveResolveVotes) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	phase := game.Phase.Value()
	if phase != phaseDay && phase != phaseNight {
		return errors.New("not in a voting phase")
	}
	// Check that all eligible voters have voted
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Eliminated {
			continue
		}
		if phase == phaseNight && p.Role.Value() != roleWerewolf {
			// Villagers don't vote at night
			continue
		}
		if voteForPhase(p, phase) < 0 {
			return errors.New("not all eligible players have voted")
		}
	}

	return nil
}