func(state boardgame.ImmutableState) bool {
			game, _ := concreteStates(state)
			return game.CurrentPlayerHasCardsToReveal()
		}