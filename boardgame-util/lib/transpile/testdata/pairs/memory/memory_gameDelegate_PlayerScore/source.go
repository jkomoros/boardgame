func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	return pState.(*playerState).GameScore()
}