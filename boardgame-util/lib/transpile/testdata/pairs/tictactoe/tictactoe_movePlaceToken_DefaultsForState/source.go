func (m *movePlaceToken) DefaultsForState(state boardgame.ImmutableState) {
	game, _ := concreteStates(state)

	m.MoveComponentToSlot.DefaultsForState(state)

	//Default to setting a slot that's empty.
	for i, c := range game.Slots.Components() {
		if c == nil {
			m.Slot = i
			break
		}
	}
}