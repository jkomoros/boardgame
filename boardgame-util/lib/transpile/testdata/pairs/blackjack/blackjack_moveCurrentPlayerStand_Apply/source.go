func (m *moveCurrentPlayerStand) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	currentPlayer.Stood = true

	return nil
}