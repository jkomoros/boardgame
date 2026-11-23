func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	component := c.Values().(*playerToken)

	_, players := concreteStates(state)

	switch component.Value {
	case X:
		return players[0].UnusedTokens, nil
	case O:
		return players[1].UnusedTokens, nil
	}
	return nil, errors.New("Component with unexpected value")
}