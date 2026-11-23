func (m *movePlaceToken) Apply(state boardgame.State) error {

	if err := m.MoveComponentToSlot.Apply(state); err != nil {
		return err
	}

	game, players := concreteStates(state)

	players[m.TargetPlayerIndex.EnsureValid(state)].TokensToPlaceThisTurn--

	game.Phase.SetValue(phaseAfterFirstMove)

	return nil
}