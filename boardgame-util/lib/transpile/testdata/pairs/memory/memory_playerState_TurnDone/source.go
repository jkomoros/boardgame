func (p *playerState) TurnDone() error {
	if p.CardsLeftToReveal > 0 {
		return errors.New("they still have cards left to reveal")
	}

	game, _ := concreteStates(p.State())

	if game.VisibleCards.NumComponents() > 0 {
		return errors.New("there are still some cards revealed, which they must hide")
	}

	return nil
}