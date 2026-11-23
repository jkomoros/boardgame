func (m *moveMoveCardBetweenShortStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.SecondShortStack
	to := game.FirstShortStack

	if m.FromFirst {
		from = game.FirstShortStack
		to = game.SecondShortStack
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}