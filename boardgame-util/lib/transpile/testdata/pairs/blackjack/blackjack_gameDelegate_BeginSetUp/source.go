func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	game, _ := concreteStates(state)
	maxRounds, err := strconv.Atoi(variant[variantKeyMaxRounds])
	if err != nil {
		maxRounds = defaultMaxRounds
	}
	game.MaxRounds = maxRounds
	return nil
}