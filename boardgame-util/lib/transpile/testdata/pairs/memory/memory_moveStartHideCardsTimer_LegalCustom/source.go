func (m *moveStartHideCardsTimer) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	game, _ := concreteStates(state)

	if game.HideCardsTimer.Active() {
		return errors.New("the timer is already active")
	}

	return nil
}