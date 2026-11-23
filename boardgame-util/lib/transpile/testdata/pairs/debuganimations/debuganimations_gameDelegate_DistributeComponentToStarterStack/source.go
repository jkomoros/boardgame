func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)

	if c.Deck().Name() == tokensDeckName {

		if game.TokensTo.NumComponents() < 9 {
			return game.TokensTo, nil
		}

		if game.SanitizedTokensFrom.NumComponents() < 10 {
			return game.SanitizedTokensFrom, nil
		}

		if game.SanitizedTokensTo.NumComponents() < 9 {
			return game.SanitizedTokensTo, nil
		}

		return game.TokensFrom, nil
	}

	if game.FirstShortStack.NumComponents() < 1 {
		return game.FirstShortStack, nil
	}

	if game.SecondShortStack.NumComponents() < 1 {
		return game.SecondShortStack, nil
	}

	if game.DiscardStack.NumComponents() < 2 {
		return game.DiscardStack, nil
	}

	if game.HiddenCard.NumComponents() < 1 {
		return game.HiddenCard, nil
	}

	if game.FanStack.NumComponents() < 6 {
		return game.FanStack, nil
	}

	if game.FanDiscard.NumComponents() < 3 {
		return game.FanDiscard, nil
	}

	if game.VisibleStack.NumComponents() < 5 {
		return game.VisibleStack, nil
	}

	if game.HiddenStack.NumComponents() < 4 {
		return game.HiddenStack, nil
	}

	if game.AllVisibleStack.NumComponents() < 4 {
		return game.AllVisibleStack, nil
	}

	return game.DrawStack, nil

}