func (p *playerState) TurnDone() error {
	if !p.Eliminated && !p.Stood {
		return errors.New("they have neither busted nor decided to stand")
	}
	return nil
}