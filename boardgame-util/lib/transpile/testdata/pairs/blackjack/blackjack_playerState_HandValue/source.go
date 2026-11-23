func (p *playerState) HandValue() int {
	return handValue(p.Hand.ImmutableComponents())
}