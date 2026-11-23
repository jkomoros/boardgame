func populateFellowWolves(players []*playerState) {
	var werewolves []boardgame.PlayerIndex
	for i, p := range players {
		p.FellowWolves = nil
		if behaviors.PlayerIsInactive(p) || p.Role.Value() != roleWerewolf {
			continue
		}
		werewolves = append(werewolves, boardgame.PlayerIndex(i))
	}
	for _, wolf := range werewolves {
		for _, fellow := range werewolves {
			if fellow != wolf {
				players[wolf].FellowWolves = append(players[wolf].FellowWolves, fellow)
			}
		}
	}
}