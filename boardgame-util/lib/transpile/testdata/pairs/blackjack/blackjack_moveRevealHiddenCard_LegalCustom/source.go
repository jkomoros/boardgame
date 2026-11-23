func (m *moveRevealHiddenCard) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	first := p.HiddenHand.ImmutableFirst()
	if err := first.MayMoveTo(p.VisibleHand); err != nil {
		return err
	}

	return nil
}