func (m *moveMoveCardBetweenDrawAndDiscardStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	var from boardgame.Stack
	var to boardgame.Stack
	if m.FromDraw {
		from = game.DrawStack
		to = game.DiscardStack
	} else {
		from = game.DiscardStack
		to = game.DrawStack
	}

	first := from.ImmutableFirst()
	if first == nil {
		return errors.New("source stack has no cards to move")
	}

	return first.MayMoveTo(to)
}