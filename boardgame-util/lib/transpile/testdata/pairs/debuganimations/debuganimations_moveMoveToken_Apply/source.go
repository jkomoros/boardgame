func (m *moveMoveToken) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.TokensFrom.NumComponents() < 10 {
		return game.TokensTo.First().MoveTo(game.TokensFrom, 2)
	}

	return game.TokensFrom.ComponentAt(2).MoveToFirstSlot(game.TokensTo)

}