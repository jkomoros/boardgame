func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {
	game, players := concreteStates(state)

	// During gathering no roles are assigned yet, so aliveWerewolves
	// would be 0 and the "villagers win" condition would fire at game
	// creation. The game can't end before it begins.
	if game.Phase.Value() == phaseGathering {
		return false, nil
	}

	var aliveWerewolves, aliveVillagers int

	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Eliminated {
			continue
		}
		if p.Role.Value() == roleWerewolf {
			aliveWerewolves++
		} else {
			aliveVillagers++
		}
	}

	villagersWin := aliveWerewolves == 0
	werewolvesWin := !villagersWin && aliveWerewolves >= aliveVillagers

	if !villagersWin && !werewolvesWin {
		return false, nil
	}

	for i, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		isWerewolf := p.Role.Value() == roleWerewolf
		if (villagersWin && !isWerewolf) || (werewolvesWin && isWerewolf) {
			winners = append(winners, boardgame.PlayerIndex(i))
		}
	}

	return true, winners
}