func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	game, _ := concreteStates(state)

	game.CardSet = variant[variantKeyCardSet]

	switch variant[variantKeyNumCards] {
	case numCardsSmall:
		game.NumCards = 10
	case numCardsMedium:
		game.NumCards = 20
	case numCardsLarge:
		game.NumCards = 40
	default:
		game.NumCards = 20
	}

	if err := game.HiddenCards.SetSize(game.NumCards); err != nil {
		return errors.New("Couldn't set up hidden cards: " + err.Error())
	}
	if err := game.VisibleCards.SetSize(game.NumCards); err != nil {
		return errors.New("Couldn't set up revealed cards: " + err.Error())
	}

	return nil
}