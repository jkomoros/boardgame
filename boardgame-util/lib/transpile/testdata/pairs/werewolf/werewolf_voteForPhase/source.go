func voteForPhase(player *playerState, phase enum.EnumKey) boardgame.PlayerIndex {
	if phase == phaseNight {
		return player.NightVote
	}
	return player.DayVote
}