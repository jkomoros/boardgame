func (m *moveMoveTokenSanitized) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.SanitizedTokensFrom.NumComponents() < 10 {
		return game.SanitizedTokensTo.First().MoveTo(game.SanitizedTokensFrom, 2)
	}

	return game.SanitizedTokensFrom.ComponentAt(2).MoveToFirstSlot(game.SanitizedTokensTo)

}