func (p *playerState) ResetForTurnStart() error {
	p.CardsLeftToReveal = 2
	return nil
}