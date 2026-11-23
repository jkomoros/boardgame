func (m *moveCastVote) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	voter := players[m.TargetPlayerIndex]

	if voter.Eliminated {
		return errors.New("eliminated players cannot vote")
	}

	phase := game.Phase.Value()

	if voteForPhase(voter, phase) >= 0 {
		return errors.New("you have already voted this phase")
	}

	if phase == phaseNight {
		if voter.Role.Value() != roleWerewolf {
			return errors.New("only werewolves may act during the night")
		}
	} else if phase != phaseDay {
		return errors.New("voting is not allowed during this phase")
	}

	// Validate the target
	target := m.VoteTarget
	if target < 0 || int(target) >= len(players) {
		return errors.New("invalid vote target")
	}

	targetPlayer := players[target]
	if targetPlayer.Eliminated {
		return errors.New("cannot vote for an eliminated player")
	}

	if behaviors.PlayerIsInactive(targetPlayer) {
		return errors.New("cannot vote for an inactive player")
	}

	if target == m.TargetPlayerIndex {
		return errors.New("you cannot vote for yourself")
	}

	return nil
}