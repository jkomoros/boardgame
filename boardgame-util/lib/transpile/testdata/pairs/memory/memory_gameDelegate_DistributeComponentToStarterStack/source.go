func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)

	//For now, shunt all cards to UnusedCards. In FinishSetup we'll construct
	//the deck based on config.
	return game.UnusedCards, nil

}