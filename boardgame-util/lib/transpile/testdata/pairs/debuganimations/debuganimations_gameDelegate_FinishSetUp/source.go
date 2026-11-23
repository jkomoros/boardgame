func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil

}