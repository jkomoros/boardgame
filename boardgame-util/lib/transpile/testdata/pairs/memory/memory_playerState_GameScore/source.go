func (p *playerState) GameScore() int {
	return p.WonCards.NumComponents()
}