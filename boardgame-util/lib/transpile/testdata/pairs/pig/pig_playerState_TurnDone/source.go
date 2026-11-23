func (p *playerState) TurnDone() error {
	if !p.DieCounted {
		return errors.New("the most recent die roll has not been counted")
	}

	if !p.Eliminated && !p.Done {
		return errors.New("they have not either busted or signaled that they are done")
	}

	return nil
}