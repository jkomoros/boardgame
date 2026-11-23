func (p *playerState) ResetForTurnEnd() error {
	if p.Done {
		p.Score += p.RoundScore
	}
	p.ResetForTurn()
	return nil
}