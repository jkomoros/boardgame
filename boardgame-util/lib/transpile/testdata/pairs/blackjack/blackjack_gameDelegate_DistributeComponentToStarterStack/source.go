func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	game, _ := concreteStates(state)

	card := c.Values().(*playingcards.Card)

	if card.Rank.Value() == playingcards.RankJoker {
		return game.UnusedCards, nil
	}

	return game.DrawStack, nil

}