func (m *moveRollDice) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	die := game.Die.ComponentAt(0)

	die.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())

	p.DieCounted = false

	return nil
}