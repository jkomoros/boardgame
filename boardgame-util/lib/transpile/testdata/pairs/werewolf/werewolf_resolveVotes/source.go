func resolveVotes(state boardgame.State) error {
	game, players := concreteStates(state)

	phase := game.Phase.Value()

	// Tally votes
	voteCounts := make(map[boardgame.PlayerIndex]int)
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Eliminated {
			continue
		}
		if phase == phaseNight && p.Role.Value() != roleWerewolf {
			continue
		}
		vote := voteForPhase(p, phase)
		if vote >= 0 {
			voteCounts[vote]++
		}
	}

	// Find the player with most votes. Iterate by player index (not map
	// order) so tie-breaks are deterministic — golden-style replays would
	// otherwise produce different victims per run.
	var maxVotes int
	var eliminated boardgame.PlayerIndex = -1
	tied := false

	for i := 0; i < len(players); i++ {
		target := boardgame.PlayerIndex(i)
		count := voteCounts[target]
		if count == 0 {
			continue
		}
		if count > maxVotes {
			maxVotes = count
			eliminated = target
			tied = false
		} else if count == maxVotes {
			tied = true
		}
	}

	// During day phase, ties mean no elimination
	if phase == phaseDay && tied {
		eliminated = -1
	}

	// Eliminate the target (if any)
	if eliminated >= 0 && int(eliminated) < len(players) {
		players[eliminated].SetEliminated()
	}

	// Reset all votes
	for _, p := range players {
		p.DayVote = -1
		p.NightVote = -1
	}

	// Transition phase
	if phase == phaseDay {
		game.SetCurrentPhase(phaseNight)
	} else {
		game.RoundNumber++
		game.SetCurrentPhase(phaseDay)
	}

	return nil
}