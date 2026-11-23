func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {
	//This debug game is never finished
	return false, nil
}