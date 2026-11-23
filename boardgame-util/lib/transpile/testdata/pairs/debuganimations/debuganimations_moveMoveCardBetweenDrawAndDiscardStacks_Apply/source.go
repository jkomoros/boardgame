func (m *moveMoveCardBetweenDrawAndDiscardStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.DiscardStack
	to := game.DrawStack

	if m.FromDraw {
		from = game.DrawStack
		to = game.DiscardStack
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}