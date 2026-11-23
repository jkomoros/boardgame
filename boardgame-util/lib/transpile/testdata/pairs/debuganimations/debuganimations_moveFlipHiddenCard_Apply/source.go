func (m *moveFlipHiddenCard) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.VisibleCard
	to := game.HiddenCard

	if game.HiddenCard.NumComponents() > 0 {
		from = game.HiddenCard
		to = game.VisibleCard
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}