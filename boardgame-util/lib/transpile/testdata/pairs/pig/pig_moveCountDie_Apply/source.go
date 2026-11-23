func (m *moveCountDie) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	value := game.Die.ComponentAt(0).DynamicValues().(*dice.DynamicValue).Value

	if value == 1 {
		//Bust!
		p.Eliminated = true
	} else {
		p.RoundScore += value
	}

	p.DieCounted = true

	return nil
}