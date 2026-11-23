func (m *moveStartHideCardsTimer) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	move := state.Game().MoveByName(hideCardMoveName)

	game.HideCardsTimer.Start(hideCardsDuration, move)

	return nil
}