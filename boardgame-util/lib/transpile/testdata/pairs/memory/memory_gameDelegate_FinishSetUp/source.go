func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, players := concreteStates(state)

	//First, shuffle unused cards to ensure a different set of cards that
	//adhere to config each time.

	game.UnusedCards.Shuffle()

	//Now, go assemble the deck by going through each component from the
	//front, seeing if it matches. If it does, put it in the HiddenCards array
	//and find its match and also put it in the HiddenCards. If it doesn't,
	//put it in the UnusedCardsScratch (along with its pair) to get it out of
	//the way.

	for game.HiddenCards.NumComponents() < game.NumCards {

		//The card to match.
		firstCard := game.UnusedCards.ComponentAt(0).Values().(*cardValue)

		//Now find its pair. If we keep it, we'll also keep its pair. If we
		//move it to scratch, we'll also move its pair to scratch.
		var pairCardIndex int

		for i := 1; i < game.UnusedCards.Len(); i++ {
			candidateCard := game.UnusedCards.ComponentAt(i).Values().(*cardValue)

			if candidateCard.Type == firstCard.Type {
				pairCardIndex = i
				break
			}
		}

		if pairCardIndex == 0 {
			//Uh oh, couldn't find the pair...

			return errors.New("unexpectedly unable to find the pair when sorting cards to include")
		}

		useCard := false

		if game.CardSet == cardSetAll {
			useCard = true
		} else if game.CardSet == firstCard.CardSet {
			useCard = true
		}

		//Doing the pair card first means that its index doesn't have to be
		//modified down by 1
		if useCard {
			if err := game.UnusedCards.ComponentAt(pairCardIndex).MoveToNextSlot(game.HiddenCards); err != nil {
				return errors.New("Couldn't move pair card to other slot: " + err.Error())
			}
			if err := game.UnusedCards.First().MoveToNextSlot(game.HiddenCards); err != nil {
				return errors.New("Couldn't move first card to other slot: " + err.Error())
			}
		} else {
			if err := game.UnusedCards.ComponentAt(pairCardIndex).SlideToLastSlot(); err != nil {
				return errors.New("Couldn't move pair card to end: " + err.Error())
			}
			if err := game.UnusedCards.First().SlideToLastSlot(); err != nil {
				return errors.New("Couldn't move first card to end: " + err.Error())
			}
		}

	}

	game.HiddenCards.Shuffle()

	players[0].CardsLeftToReveal = 2

	return nil
}