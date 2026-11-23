func (g *gameState) CurrentPlayerHasCardsToReveal() bool {
	_, players := concreteStates(g.State())

	p := players[g.CurrentPlayer]

	return p.CardsLeftToReveal > 0
}