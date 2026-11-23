func(player boardgame.ImmutableSubState) int {
			return player.(*playerState).HandValue()
		}