func (p *playerState) ResetForTurn() {
	p.Done = false
	p.Eliminated = false
	p.RoundScore = 0
	p.DieCounted = true
}