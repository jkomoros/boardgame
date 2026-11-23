func (p *playerState) TurnDone() error {
	if p.TokensToPlaceThisTurn > 0 {
		return errors.New("they still have tokens left to place this turn")
	}
	return nil
}