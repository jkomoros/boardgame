func (m *moveCollectCards) Apply(state boardgame.State) error {
	game, players := concreteStates(state)
	for _, p := range players {
		if err := p.HiddenHand.MoveAllTo(game.DiscardStack); err != nil {
			return err
		}
		if err := p.VisibleHand.MoveAllTo(game.DiscardStack); err != nil {
			return err
		}
	}
	return nil
}