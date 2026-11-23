func (p *playerState) ResetForTurnStart() error {
	p.TokensToPlaceThisTurn = 1
	return nil
}