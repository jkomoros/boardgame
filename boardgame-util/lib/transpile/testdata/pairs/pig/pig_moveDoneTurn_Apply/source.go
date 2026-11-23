func (m *moveDoneTurn) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	p.Done = true

	return nil
}