func (m *moveRevealCard) Apply(state boardgame.State) error {

	if err := m.MoveComponentToSlot.Apply(state); err != nil {
		return err
	}

	game, players := concreteStates(state)

	players[game.CurrentPlayer.EnsureValid(state)].CardsLeftToReveal--

	//If the cards are the same, the FixUpMove CaptureCards will fire after this.

	return nil
}