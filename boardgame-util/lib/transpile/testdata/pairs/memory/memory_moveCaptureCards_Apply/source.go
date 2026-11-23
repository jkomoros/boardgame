func (m *moveCaptureCards) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	for i, c := range game.VisibleCards.Components() {
		if c != nil {
			game.VisibleCards.ComponentAt(i).MoveToNextSlot(p.WonCards)
		}
	}

	return nil
}