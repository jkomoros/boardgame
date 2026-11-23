func (m *moveBeginGame) Apply(state boardgame.State) error {
	if err := m.StartPhase.Apply(state); err != nil {
		return err
	}

	_, players := concreteStates(state)

	var active []*playerState
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		active = append(active, p)
	}

	numWerewolves := 1
	if len(active) >= 6 {
		numWerewolves = 2
	}

	indices := state.Rand().Perm(len(active))

	for i, p := range active {
		isWerewolf := false
		for j := 0; j < numWerewolves; j++ {
			if indices[j] == i {
				isWerewolf = true
				break
			}
		}
		if isWerewolf {
			p.Role.SetValue(roleWerewolf)
		} else {
			p.Role.SetValue(roleVillager)
		}
	}

	populateFellowWolves(players)

	return startVoteTimer(state)
}