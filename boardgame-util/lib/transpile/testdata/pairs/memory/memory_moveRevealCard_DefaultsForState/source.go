func (m *moveRevealCard) DefaultsForState(state boardgame.ImmutableState) {

	m.MoveComponentToSlot.DefaultsForState(state)

	game, _ := concreteStates(state)

	for i, c := range game.HiddenCards.Components() {
		if c != nil {
			m.CardIndex = i
			break
		}
	}
}