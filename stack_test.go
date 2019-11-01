package boardgame

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/workfit/tester/assert"
)

func TestShuffleCount(t *testing.T) {
	game := testDefaultGame(t, false)

	gameState, _ := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck

	assert.For(t).ThatActual(drawStack.ShuffleCount()).Equals(0)

	err := drawStack.Shuffle()
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(drawStack.ShuffleCount()).Equals(1)

	err = drawStack.PublicShuffle()
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(drawStack.ShuffleCount()).Equals(2)

	//Verify that shuffleCount makes it through round-tripping.
	blob, err := json.Marshal(drawStack)

	assert.For(t).ThatActual(err).IsNil()

	newGame := testDefaultGame(t, false)

	newGameState, _ := concreteStates(newGame.CurrentState())

	newDrawStack := newGameState.DrawDeck

	assert.For(t).ThatActual(newDrawStack.ShuffleCount()).Equals(0)

	err = json.Unmarshal(blob, newDrawStack)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(newDrawStack.ShuffleCount()).Equals(2)

	//TODO: also test this for sized stacks.

}

func TestEmptyStacks(t *testing.T) {
	game := testDefaultGame(t, false)

	gState := game.CurrentState().ImmutableGameState().(*testGameState)

	space := gState.MyBoard.SpaceAt(2)

	c := space.ComponentAt(0)

	assert.For(t).ThatActual(c).IsNil()

	c = space.ComponentAt(space.Len() - 1)

	assert.For(t).ThatActual(c).IsNil()
}

func TestMultipleStackMoves(t *testing.T) {
	manager, err := NewGameManager(defaultTestGameDelegate(10), newTestStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	deck := manager.Chest().Deck("test")

	st := game.CurrentState().(*state)

	verifyContainingComponent(t, st, deck)

	gameState := st.GameState().(*testGameState)

	drawDeck := gameState.DrawDeck

	destination := gameState.MyBoard.SpaceAt(0)

	for i := 0; i < 5; i++ {
		err := drawDeck.ComponentAt(0).MoveToNextSlot(destination)
		assert.For(t, i).ThatActual(err).IsNil()
		verifyContainingComponent(t, st, deck)
	}

}

func TestContainingComponent(t *testing.T) {
	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	assert.For(t).ThatActual(deck).IsNotNil()

	verifyContainingComponent(t, game.CurrentState(), deck)

	//TODO: test a rehyrdated state, one after a move, a sanitized state.

	move := game.MoveByName("Draw Card")

	err := <-game.ProposeMove(move, AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	verifyContainingComponent(t, game.CurrentState(), deck)

	refriedGame := game.Manager().Game(game.ID())

	assert.For(t).ThatActual(refriedGame).IsNotNil()

	verifyContainingComponent(t, refriedGame.CurrentState(), deck)

	sanitizedState := game.CurrentState().SanitizedForPlayer(0)

	//DrawDeck is the on that is sanitized by default
	componentsInDrawDeck := make(map[*component]bool)

	for _, c := range game.CurrentState().ImmutableGameState().(*testGameState).DrawDeck.Components() {
		componentsInDrawDeck[c.ptr()] = true
	}

	for i, c := range deck.Components() {

		stack, slotIndex, err := c.ImmutableInstance(sanitizedState).ContainingImmutableStack()

		//DrawDeck is sanitized, so we should get errors for that.
		if componentsInDrawDeck[c.ptr()] {
			assert.For(t, i).ThatActual(err).IsNotNil()
			continue
		}

		assert.For(t, i).ThatActual(stack).IsNotNil()

		otherC := stack.ImmutableComponentAt(slotIndex)

		assert.For(t, i).ThatActual(otherC.ptr()).Equals(c)

	}

	//TODO: ideally we'd also test here that SwapComponents and
	//ShuffleComponents pass the test.

}

func verifyContainingComponent(t *testing.T, st ImmutableState, deck *Deck) {

	_, _, err := deck.GenericComponent().ImmutableInstance(st).ContainingImmutableStack()

	assert.For(t).ThatActual(err).IsNotNil()

	for i, c := range deck.Components() {
		stack, slotIndex, err := c.ImmutableInstance(st).ContainingImmutableStack()
		assert.For(t, i).ThatActual(err).IsNil()
		assert.For(t, i).ThatActual(stack).IsNotNil()

		otherC := stack.ImmutableComponentAt(slotIndex)

		assert.For(t, i).ThatActual(otherC.ptr()).Equals(c)

	}
}

func TestMergedValidStack(t *testing.T) {
	game := testDefaultGame(t, true)

	testDeck := game.Manager().Chest().Deck("test")

	sized := testDeck.NewSizedStack(3).(*sizedStack)

	st := game.CurrentState().(*state)

	sized.setState(st)

	sized.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(st))
	sized.insertComponentAt(2, testDeck.ComponentAt(1).ImmutableInstance(st))

	growable := testDeck.NewStack(2).(*growableStack)

	growable.setState(game.CurrentState().(*state))

	growable.insertNext(testDeck.ComponentAt(2).ImmutableInstance(st))
	growable.insertNext(testDeck.ComponentAt(3).ImmutableInstance(st))

	otherSized := testDeck.NewSizedStack(4).(*sizedStack)

	otherSized.setState(st)

	merged := NewConcatenatedStack(sized, nil)

	assert.For(t).ThatActual(merged).IsNotNil()

	assert.For(t).ThatActual(merged.Valid()).IsNotNil()

	merged = NewOverlappedStack(sized, growable)

	assert.For(t).ThatActual(merged.Valid()).IsNotNil()

	merged = NewOverlappedStack(sized, otherSized)

	assert.For(t).ThatActual(merged.Valid()).IsNotNil()

	otherSized.SetSize(sized.Len())

	assert.For(t).ThatActual(merged.Valid()).IsNil()
}

func TestConcatenatedStack(t *testing.T) {
	game := testDefaultGame(t, true)

	testDeck := game.Manager().Chest().Deck("test")

	sized := testDeck.NewSizedStack(3).(*sizedStack)

	st := game.CurrentState().(*state)
	sized.setState(st)

	sized.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(st))
	sized.insertComponentAt(2, testDeck.ComponentAt(1).ImmutableInstance(st))

	growable := testDeck.NewStack(2).(*growableStack)

	growable.setState(st)

	growable.insertNext(testDeck.ComponentAt(2).ImmutableInstance(st))
	growable.insertNext(testDeck.ComponentAt(3).ImmutableInstance(st))

	merged := NewConcatenatedStack(sized, growable)

	assert.For(t).ThatActual(merged).IsNotNil()

	assert.For(t).ThatActual(merged.Valid()).IsNil()

	assert.For(t).ThatActual(merged.Len()).Equals(5)
	assert.For(t).ThatActual(merged.NumComponents()).Equals(4)
	assert.For(t).ThatActual(merged.SlotsRemaining()).Equals(1)
	assert.For(t).ThatActual(merged.ImmutableSizedStack()).IsNil()

	expectedIds := []string{
		"2a7effe5e4000914791f95f6c1a711e54d346020",
		"",
		"26367debb9c2ce3d0a24de425c5797f43ac43909",
		"eb8210726e015b0e472c225e512166e7f1ac34de",
		"667bc729077137a8932a2cffc5ca4b0cec905956",
	}

	assert.For(t).ThatActual(merged.Ids()).Equals(expectedIds)

	//TODO: test other aspects are what are expected

}

func TestOverlappedStack(t *testing.T) {
	game := testDefaultGame(t, true)

	testDeck := game.Manager().Chest().Deck("test")

	first := testDeck.NewSizedStack(4).(*sizedStack)

	st := game.CurrentState().(*state)
	first.setState(st)

	first.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(st))
	first.insertComponentAt(2, testDeck.ComponentAt(1).ImmutableInstance(st))

	second := testDeck.NewSizedStack(4).(*sizedStack)

	second.setState(st)

	second.insertComponentAt(0, testDeck.ComponentAt(2).ImmutableInstance(st))
	second.insertComponentAt(1, testDeck.ComponentAt(3).ImmutableInstance(st))

	merged := NewOverlappedStack(first, second)

	assert.For(t).ThatActual(merged).IsNotNil()

	assert.For(t).ThatActual(merged.Valid()).IsNil()

	assert.For(t).ThatActual(merged.Len()).Equals(4)
	assert.For(t).ThatActual(merged.NumComponents()).Equals(3)
	assert.For(t).ThatActual(merged.SlotsRemaining()).Equals(1)
	assert.For(t).ThatActual(merged.ImmutableSizedStack()).IsNotNil()

	expectedIds := []string{
		"2a7effe5e4000914791f95f6c1a711e54d346020",
		"667bc729077137a8932a2cffc5ca4b0cec905956",
		"26367debb9c2ce3d0a24de425c5797f43ac43909",
		"",
	}

	assert.For(t).ThatActual(merged.Ids()).Equals(expectedIds)

	//TODO: test other aspects are what are expected

}

func TestMoveExtreme(t *testing.T) {
	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	testDeck := chest.Deck("test")

	sized := testDeck.NewSizedStack(5).(*sizedStack)

	st := game.CurrentState().(*state)

	sized.setState(st)

	sized.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(st))
	sized.insertComponentAt(1, testDeck.ComponentAt(1).ImmutableInstance(st))
	sized.insertComponentAt(3, testDeck.ComponentAt(2).ImmutableInstance(st))

	assert.For(t).ThatActual(sized.indexes).Equals([]int{0, 1, -1, 2, -1})

	err := sized.First().SlideToLastSlot()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(sized.indexes).Equals([]int{-1, 1, -1, 2, 0})

	err = sized.ComponentAt(1).SlideToFirstSlot()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(sized.indexes).Equals([]int{1, -1, -1, 2, 0})

	growable := testDeck.NewStack(0).(*growableStack)

	growable.setState(st)

	growable.insertNext(testDeck.ComponentAt(0).ImmutableInstance(st))
	growable.insertNext(testDeck.ComponentAt(1).ImmutableInstance(st))
	growable.insertNext(testDeck.ComponentAt(2).ImmutableInstance(st))

	assert.For(t).ThatActual(growable.indexes).Equals([]int{0, 1, 2})

	err = growable.First().SlideToLastSlot()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(growable.indexes).Equals([]int{1, 2, 0})

	err = growable.ComponentAt(1).SlideToFirstSlot()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(growable.indexes).Equals([]int{2, 1, 0})
}

func TestExpandContractSizedStackSize(t *testing.T) {
	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	testDeck := chest.Deck("test")

	st := game.CurrentState().(*state)

	sized := testDeck.NewSizedStack(5).(*sizedStack)

	sized.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(st))
	sized.insertComponentAt(1, testDeck.ComponentAt(1).ImmutableInstance(st))
	sized.insertComponentAt(3, testDeck.ComponentAt(2).ImmutableInstance(st))

	err := sized.ExpandSize(-2)

	assert.For(t).ThatActual(err).IsNotNil()

	err = sized.ExpandSize(1)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(sized.size).Equals(6)
	assert.For(t).ThatActual(len(sized.indexes)).Equals(6)

	var nilComponent Component

	assert.For(t).ThatActual(sized.ComponentAt(5)).Equals(nilComponent)

	err = sized.ContractSize(-2)

	assert.For(t).ThatActual(err).IsNotNil()

	err = sized.ContractSize(2)

	assert.For(t).ThatActual(err).IsNotNil()

	err = sized.ContractSize(4)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(sized.size).Equals(4)
	assert.For(t).ThatActual(len(sized.indexes)).Equals(4)

	//Make sure the slot was taken from the right, not the middle.
	assert.For(t).ThatActual(sized.ComponentAt(2)).Equals(nilComponent)

	err = sized.ContractSize(3)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(sized.size).Equals(3)
	assert.For(t).ThatActual(len(sized.indexes)).Equals(3)

}

func TestChangedSizeStackRoundTrip(t *testing.T) {
	game := testDefaultGame(t, false)

	testDeck := game.Manager().Chest().Deck("test")

	cState := game.CurrentState()

	g, _ := concreteStates(cState)

	g.DownSizeStack.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(cState))
	g.DownSizeStack.insertComponentAt(2, testDeck.ComponentAt(1).ImmutableInstance(cState))

	assert.For(t).ThatActual(g.DownSizeStack.NumComponents()).Equals(2)
	assert.For(t).ThatActual(g.DownSizeStack.Len()).Equals(4)

	err := g.DownSizeStack.ContractSize(3)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(g.DownSizeStack.Len()).Equals(3)

	rec := cState.StorageRecord()

	refriedState, err := game.Manager().stateFromRecord(rec)

	assert.For(t).ThatActual(err).IsNil()

	rG, _ := concreteStates(refriedState)

	originalStack := g.DownSizeStack.(*sizedStack)
	refriedStack := rG.DownSizeStack.(*sizedStack)

	assert.For(t).ThatActual(refriedStack.indexes).Equals(originalStack.indexes)
	assert.For(t).ThatActual(refriedStack.size).Equals(originalStack.size)

}

func TestExpandContractDefaultStackSize(t *testing.T) {
	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	testDeck := chest.Deck("test")

	stack := testDeck.NewStack(0)

	//Fails because maxSize is 0
	err := stack.ExpandSize(5)

	assert.For(t).ThatActual(err).IsNotNil()

	//Fails because maxSize is 0
	err = stack.ContractSize(3)

	assert.For(t).ThatActual(err).IsNotNil()

	err = stack.SetSize(3)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(stack.MaxSize()).Equals(3)

	err = stack.ExpandSize(2)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(stack.MaxSize()).Equals(5)

	st := game.CurrentState().(*state)

	stack.insertComponentAt(0, testDeck.ComponentAt(0).ImmutableInstance(st))
	stack.insertComponentAt(1, testDeck.ComponentAt(1).ImmutableInstance(st))

	//Fails: too many components in stack
	err = stack.ContractSize(1)

	assert.For(t).ThatActual(err).IsNotNil()

	err = stack.ContractSize(2)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(stack.MaxSize()).Equals(2)
}

func TestFixedSize(t *testing.T) {

	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	testDeck := chest.Deck("test")

	sized := testDeck.NewSizedStack(5)

	assert.For(t).ThatActual(sized.SizedStack()).IsNotNil()

	growable := testDeck.NewStack(0)

	assert.For(t).ThatActual(growable.SizedStack()).IsNil()

}

func TestSort(t *testing.T) {

	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	testDeck := chest.Deck("test")

	gStack := testDeck.NewStack(0)

	st := game.CurrentState().(*state)

	gStack.setState(st)

	gStack.insertNext(testDeck.Components()[0].ImmutableInstance(st))
	gStack.insertNext(testDeck.Components()[1].ImmutableInstance(st))
	gStack.insertNext(testDeck.Components()[2].ImmutableInstance(st))
	gStack.insertNext(testDeck.Components()[3].ImmutableInstance(st))

	for stackSorted(gStack) {
		if err := gStack.Shuffle(); err != nil {
			t.Fatal("Couldn't shuffle: " + err.Error())
		}
	}

	lessFunc := func(i, j ImmutableComponentInstance) bool {
		if i == nil {
			return true
		}
		if j == nil {
			return false
		}
		return i.Values().(*testingComponent).Integer < j.Values().(*testingComponent).Integer
	}

	err := gStack.SortComponents(lessFunc)

	assert.For(t).ThatActual(err).IsNil()

	sorted := stackSorted(gStack)

	assert.For(t).ThatActual(sorted).IsTrue()

	sStack := testDeck.NewSizedStack(5)

	sStack.setState(st)

	sStack.insertComponentAt(0, testDeck.Components()[0].ImmutableInstance(st))
	sStack.insertComponentAt(1, testDeck.Components()[1].ImmutableInstance(st))
	sStack.insertComponentAt(2, testDeck.Components()[2].ImmutableInstance(st))
	//Deliberately leave a nil
	sStack.insertComponentAt(4, testDeck.Components()[3].ImmutableInstance(st))

	//Shuffle at least once. But if we happen to accidentally shuffle ito
	//sorted order, shuffle again.
	sStack.Shuffle()

	for stackSorted(sStack) {
		if err := sStack.Shuffle(); err != nil {
			t.Fatal("Couldn't shuffle: " + err.Error())
		}
	}

	err = sStack.SortComponents(lessFunc)

	assert.For(t).ThatActual(err).IsNil()

	sorted = stackSorted(sStack)

	assert.For(t).ThatActual(sorted).IsTrue()

}

func stackSorted(stack Stack) bool {
	last := -1

	for _, c := range stack.Components() {
		if c == nil {
			if last == -1 {
				//That's OK
				continue
			}
			return false
		}
		current := c.Values().(*testingComponent).Integer
		if last < current {
			last = current
		} else {
			return false
		}
	}

	return true
}

func TestInflate(t *testing.T) {
	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	testDeck := chest.Deck("test")

	gStack := testDeck.NewStack(0)

	st := game.CurrentState().(*state)

	gStack.setState(st)

	gStack.insertNext(testDeck.Components()[0].ImmutableInstance(st))

	sStack := testDeck.NewSizedStack(2)

	sStack.setState(st)

	sStack.insertNext(testDeck.Components()[1].ImmutableInstance(st))

	if gStack.ComponentAt(0) == nil {
		t.Error("Couldnt' get component from inflated gstack")
	}

	if sStack.ComponentAt(0) == nil {
		t.Error("Couldn't get component from inflated sstack")
	}

	gStackBlob, err := json.Marshal(gStack)

	if err != nil {
		t.Error("Gstack didn't serialize", err)
	}

	sStackBlob, err := json.Marshal(sStack)

	if err != nil {
		t.Error("SStack didn't serialize", err)
	}

	reGStack := testDeck.NewStack(0)

	reGStack.setState(game.CurrentState().(*state))

	if err := json.Unmarshal(gStackBlob, reGStack); err != nil {
		t.Error("Couldn't reconstitute gStack", err)
	}

	reSStack := testDeck.NewSizedStack(0)

	reSStack.setState(game.CurrentState().(*state))

	if err := json.Unmarshal(sStackBlob, reSStack); err != nil {
		t.Error("Couldn't reconstitute sStack", err)
	}

	c := reGStack.ComponentAt(0)

	if !c.Equivalent(testDeck.Components()[0]) {
		t.Error("After inflating g stack, got wrong component. Wanted", testDeck.Components()[0], "got", c)
	}

	c = reSStack.ComponentAt(0)

	if !c.Equivalent(testDeck.Components()[1]) {
		t.Error("After inflating s stack, got wrong component. Wanted", testDeck.Components()[1], "got", c)
	}
}

func TestSecretMoveComponentGrowable(t *testing.T) {
	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	gStack := deck.NewStack(0)
	sStack := deck.NewSizedStack(5)

	fakeState := &state{
		game:            game,
		secretMoveCount: make(map[string][]int),
	}

	gStack.setState(fakeState)
	sStack.setState(fakeState)

	for i, c := range deck.Components() {
		if i >= 4 {
			//Only include up to first four components
			break
		}
		if i%2 == 0 {
			gStack.insertNext(c.ImmutableInstance(fakeState))
		} else {
			sStack.insertNext(c.ImmutableInstance(fakeState))
		}
	}

	assert.For(t).ThatActual(gStack.NumComponents()).Equals(len(deck.Components()) / 2)
	assert.For(t).ThatActual(sStack.NumComponents()).Equals(len(deck.Components()) / 2)

	secretMoveTestHelper(t, gStack, sStack, "growable to sized")

}

func TestSecretMoveComponentSized(t *testing.T) {
	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	gStack := deck.NewStack(0)
	sStack := deck.NewSizedStack(5)

	fakeState := &state{
		game:            game,
		secretMoveCount: make(map[string][]int),
	}

	gStack.setState(fakeState)
	sStack.setState(fakeState)

	for i, c := range deck.Components() {
		if i >= 4 {
			//Only include up to first four components
			break
		}
		if i%2 == 0 {
			gStack.insertNext(c.ImmutableInstance(fakeState))
		} else {
			sStack.insertNext(c.ImmutableInstance(fakeState))
		}
	}

	assert.For(t).ThatActual(gStack.NumComponents()).Equals(len(deck.Components()) / 2)
	assert.For(t).ThatActual(sStack.NumComponents()).Equals(len(deck.Components()) / 2)

	secretMoveTestHelper(t, sStack, gStack, "sized to growable")

}

func secretMoveTestHelper(t *testing.T, from Stack, to Stack, description string) {
	lastIds := from.Ids()
	lastIdsSeen := from.IdsLastSeen()

	toLastIds := to.Ids()
	toLastIdsSeen := to.IdsLastSeen()

	err := from.ComponentAt(from.firstComponentIndex()).SecretMoveTo(to, to.firstSlot())

	assert.For(t, description).ThatActual(err).IsNil()

	assert.For(t, description).ThatActual(from.Ids()).DoesNotEqual(lastIds)

	actualNumIdsBefore := 0

	for _, id := range lastIds {
		if id == "" {
			continue
		}
		actualNumIdsBefore++
	}

	actualNumIds := 0

	for _, id := range from.Ids() {
		if id == "" {
			continue
		}
		actualNumIds++
	}

	assert.For(t, description).ThatActual(actualNumIds).Equals(actualNumIdsBefore - 1)

	assert.For(t, description).ThatActual(to.Ids()).DoesNotEqual(toLastIds)

	//Make sure all of hte Ids have changed
	for _, id := range to.Ids() {
		if id == "" {
			continue
		}
		for _, oldID := range toLastIds {
			if oldID == "" {
				continue
			}
			assert.For(t, description).ThatActual(id).DoesNotEqual(oldID)
		}
	}

	assert.For(t, description).ThatActual(len(from.IdsLastSeen())).Equals(len(lastIdsSeen))

	assert.For(t, description).ThatActual(len(to.IdsLastSeen())).Equals(len(toLastIdsSeen)*2 + 2)
}

func TestMoveComponent(t *testing.T) {

	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	gStack := deck.NewStack(0).(*growableStack)

	sStack := deck.NewSizedStack(5).(*sizedStack)

	gStackMaxLen := deck.NewStack(4).(*growableStack)

	sStackMaxLen := deck.NewSizedStack(4).(*sizedStack)

	fakeState := &state{
		game: game,
	}

	gStack.setState(fakeState)
	sStack.setState(fakeState)
	gStackMaxLen.setState(fakeState)
	sStackMaxLen.setState(fakeState)

	for i, c := range deck.Components() {
		if i >= 4 {
			//Only include up to first four components
			break
		}
		instance := c.ImmutableInstance(fakeState)
		gStack.insertNext(instance)
		gStackMaxLen.insertNext(instance)
		sStack.insertNext(instance)
		sStackMaxLen.insertNext(instance)
	}

	if !reflect.DeepEqual(gStack.indexes, []int{0, 1, 2, 3}) {
		t.Error("gStack was not initialized like expected. Got", gStack.indexes)
	}

	if !reflect.DeepEqual(sStack.indexes, []int{0, 1, 2, 3, -1}) {
		t.Error("sStack was not initalized like expected. Got", sStack.indexes)
	}

	if !reflect.DeepEqual(gStackMaxLen.indexes, []int{0, 1, 2, 3}) {
		t.Error("gStackMaxLen was not initalized like expected. got", gStackMaxLen.indexes)
	}

	if !reflect.DeepEqual(sStackMaxLen.indexes, []int{0, 1, 2, 3}) {
		t.Error("sStackMaxLen was not initalized like expected. Got", sStackMaxLen.indexes)
	}

	sStackOtherState := deck.NewSizedStack(5)
	sStackOtherState.importFrom(sStack)
	sStackOtherState.setState(&state{})

	tests := []struct {
		source                 Stack
		destination            Stack
		componentIndex         int
		resolvedComponentIndex int
		slotIndex              int
		resolvedSlotIndex      int
		expectError            bool
		description            string
	}{
		{
			gStack,
			sStack,
			0,
			0,
			4,
			4,
			false,
			"Move from growable to sized 0 to last slot",
		},
		{
			gStack,
			sStack,
			gStack.firstComponentIndex(),
			0,
			sStack.firstSlot(),
			4,
			false,
			"Move from growable first component to sized stack first slot",
		},
		{
			sStack,
			gStack,
			sStack.firstSlot(),
			4,
			gStack.firstSlot(),
			0,
			true,
			"Move an empty slot in sized stack to growable stack",
		},
		{
			sStack,
			gStack,
			sStack.firstComponentIndex(),
			0,
			gStack.lastSlot(),
			4,
			false,
			"Move first component in sized stack to growable stack",
		},
		{
			sStackOtherState,
			gStack,
			sStackOtherState.firstComponentIndex(),
			0,
			gStack.lastSlot(),
			4,
			true,
			"Move from a stack in one state to another",
		},
		{
			sStack,
			sStack,
			sStack.firstComponentIndex(),
			0,
			sStack.lastSlot(),
			4,
			true,
			"Moving from same stack",
		},
		{
			sStack,
			gStackMaxLen,
			sStack.firstComponentIndex(),
			0,
			gStackMaxLen.lastSlot(),
			4,
			true,
			"Moving to a gstack with no more space",
		},
		{
			gStack,
			sStackMaxLen,
			gStack.firstComponentIndex(),
			0,
			sStackMaxLen.lastSlot(),
			-1,
			true,
			"Moving from a growable stack to a slot that has no more space.",
		},
		{
			gStack,
			sStack,
			10,
			10,
			sStack.lastSlot(),
			4,
			true,
			"Invalid component index",
		},
		{
			gStack,
			sStack,
			2,
			2,
			sStack.lastSlot(),
			4,
			false,
			"Moving from middle of growable stack to sized stack",
		},
		{
			gStack,
			sStack,
			gStack.firstComponentIndex(),
			0,
			sStack.nextSlot(),
			4,
			false,
			"NextSlotIndex from growable to sized",
		},
		{
			sStack,
			gStack,
			sStack.firstComponentIndex(),
			0,
			gStack.nextSlot(),
			4,
			false,
			"NextSlotIndex from sized to growable",
		},
	}

	for i, test := range tests {
		var source Stack
		var destination Stack

		switch s := test.source.(type) {
		case *growableStack:
			source = deck.NewStack(0)
			source.importFrom(s)
		case *sizedStack:
			source = deck.NewSizedStack(5)
			source.importFrom(s)
		}

		source.setState(test.source.state())

		//Some tests deliberately want to make sure that copies within same source and dest aren't allowed
		if test.source == test.destination {
			destination = source
		} else {

			switch s := test.destination.(type) {
			case *growableStack:
				destination = deck.NewStack(0)
				destination.importFrom(s)
			case *sizedStack:
				destination = deck.NewSizedStack(5)
				destination.importFrom(s)
			}
			destination.setState(test.destination.state())
		}

		preMoveSourceNumComponents := source.NumComponents()
		preMoveDestinationNumComponents := destination.NumComponents()

		component := source.ComponentAt(test.resolvedComponentIndex)

		err := moveComonentImpl(source, test.componentIndex, destination, test.slotIndex)

		if err == nil && test.expectError {
			t.Error("Got no error but expected one for", i, test.description)
		} else if err != nil && !test.expectError {
			t.Error("Got an error but didn't expect one for", i, test.description, err)
		}

		if err != nil && test.expectError {
			continue
		}

		if preMoveSourceNumComponents != source.NumComponents()+1 {
			t.Error("After the successful move, sourcew as not one component smaller.", i, test.description)
		}
		if preMoveDestinationNumComponents != destination.NumComponents()-1 {
			t.Error("After the successful move, destination was not one component bigger", i, test.description)
		}

		if finalComponent := destination.ComponentAt(test.resolvedSlotIndex); !finalComponent.Equivalent(component) {
			t.Error("After the move, the component that was supposed to be moved was not moved to the target slot.", i, test.description)
		}
	}

}

func TestSwapComponents(t *testing.T) {
	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	stack := deck.NewStack(0)

	fakeState := &state{
		game: game,
	}

	stack.setState(fakeState)

	for _, c := range deck.Components() {
		stack.insertNext(c.ImmutableInstance(fakeState))
	}

	swapComponentsTests(stack, t)

	sStack := deck.NewSizedStack(10)

	sStack.setState(fakeState)

	for _, c := range deck.Components() {
		sStack.insertNext(c.ImmutableInstance(fakeState))
	}

	swapComponentsTests(sStack, t)

}

func swapComponentsTests(stack Stack, t *testing.T) {

	zero := stack.ComponentAt(0)
	one := stack.ComponentAt(1)

	if err := stack.SwapComponents(0, 1); err != nil {
		t.Error("Legal swap not allowed")
	}

	if !stack.ComponentAt(0).Equivalent(one) {
		t.Error("Swap did not actually position of #1")
	}

	if !stack.ComponentAt(1).Equivalent(zero) {
		t.Error("Swap did not actualy change position of #0")
	}

	if err := stack.SwapComponents(-1, 0); err == nil {
		t.Error("Stack swap with illgal lower bound succeeded")
	}

	if err := stack.SwapComponents(0, stack.Len()); err == nil {
		t.Error("Stack swap with illegal upper bound succeeded")
	}

	if err := stack.SwapComponents(0, 0); err == nil {
		t.Error("Stack swap that was no op succeeded")
	}
}

func TestGrowableStackInsertComponentAt(t *testing.T) {
	//Splicing out parts of an array is so finicky that we need to make sure
	//to test it extra good...

	game := testDefaultGame(t, true)

	deck := game.Manager().Chest().Deck("test")

	fakeState := &state{
		game: game,
	}

	stack := deck.NewStack(0)

	stack.setState(fakeState)

	for i, c := range deck.Components() {
		if i >= 4 {
			//Only include up to first four components
			break
		}
		stack.insertNext(c.ImmutableInstance(fakeState))
	}

	//stack.indexes = [0, 1, 2, 3]

	startingIndexes := []int{0, 1, 2, 3}

	tests := []struct {
		slotIndex          int
		componentDeckIndex int
		expectedIndexes    []int
		description        string
	}{
		{
			0,
			2,
			[]int{2, 0, 1, 2, 3},
			"Add 2 at index 0",
		},
		{
			4,
			2,
			[]int{0, 1, 2, 3, 2},
			"Insert 2 at end",
		},
		{
			1,
			3,
			[]int{0, 3, 1, 2, 3},
			"Insert 3 at #1",
		},
		{
			3,
			1,
			[]int{0, 1, 2, 1, 3},
			"inserting 1 at #3",
		},
	}

	for i, test := range tests {
		stackCopy := deck.NewStack(0).(*growableStack)

		stackCopy.importFrom(stack)

		component := deck.ComponentAt(test.componentDeckIndex)

		if !reflect.DeepEqual(stackCopy.indexes, startingIndexes) {
			t.Error("Sanity check failed", i, "Starting indexes were", stackCopy.indexes, "wanted", startingIndexes)
		}

		stackCopy.insertComponentAt(test.slotIndex, component.ImmutableInstance(fakeState))

		if !reflect.DeepEqual(stackCopy.indexes, test.expectedIndexes) {
			t.Error("Test", i, test.description, "failed for insertComponentAt. Got", stackCopy.indexes, "wanted", test.expectedIndexes)
		}
	}
}

func TestGrowableStackRemoveComponentAt(t *testing.T) {
	//Splicing out parts of an array is so finicky that we need to make sure
	//to test it extra good...

	game := testDefaultGame(t, true)

	deck := game.Manager().Chest().Deck("test")

	fakeState := &state{
		game: game,
	}

	stack := deck.NewStack(0)

	stack.setState(fakeState)

	for i, c := range deck.Components() {
		if i >= 4 {
			//Only include up to first four components
			break
		}
		stack.insertNext(c.ImmutableInstance(fakeState))
	}

	//stack.indexes = [0, 1, 2, 3]
	startingIndexes := []int{0, 1, 2, 3}

	tests := []struct {
		componentIndex  int
		expectedIndexes []int
		description     string
	}{
		{
			0,
			[]int{1, 2, 3},
			"Remove 0",
		},
		{
			3,
			[]int{0, 1, 2},
			"remove last",
		},
		{
			1,
			[]int{0, 2, 3},
			"remove #1",
		},
		{
			2,
			[]int{0, 1, 3},
			"remove #2",
		},
	}

	for i, test := range tests {
		stackCopy := deck.NewStack(0).(*growableStack)

		stackCopy.importFrom(stack)

		if !reflect.DeepEqual(stackCopy.indexes, startingIndexes) {
			t.Error("Sanity check failed for", i, "Starting indexes were", stackCopy.indexes, "wanted", startingIndexes)
		}

		stackCopy.removeComponentAt(test.componentIndex)

		if !reflect.DeepEqual(stackCopy.indexes, test.expectedIndexes) {
			t.Error("Test", i, test.description, "failed. Got", stackCopy.indexes, "wanted", test.expectedIndexes)
		}
	}
}

func TestShuffle(t *testing.T) {
	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	stack := deck.NewStack(0).(*growableStack)

	fakeState := &state{
		game:            game,
		version:         0,
		secretMoveCount: make(map[string][]int),
	}

	stack.setState(fakeState)

	for _, c := range deck.Components() {
		stack.insertNext(c.ImmutableInstance(fakeState))
	}

	//The number of shuffles to do
	numShuffles := 10

	//Number of shuffles that were the same (which is bad)
	numShufflesTheSame := 0

	lastStackState := fmt.Sprint(stack.indexes)

	lastIds := stack.Ids()
	assert.For(t).ThatActual(len(stack.IdsLastSeen())).Equals(len(lastIds))

	for i := 0; i < numShuffles; i++ {
		if err := stack.Shuffle(); err != nil {
			t.Error("Shuffle failed", err)
		}

		if i == 0 {
			//First time through, check that ids are scrambled correctly
			assert.For(t).ThatActual(len(stack.IdsLastSeen())).Equals(len(lastIds) * 2)
			for j, id := range lastIds {
				version, ok := stack.IdsLastSeen()[id]
				assert.For(t, j, id).ThatActual(ok).IsTrue()
				assert.For(t, j, id).ThatActual(version).Equals(0)
			}
			for j, id := range stack.Ids() {
				version, ok := stack.IdsLastSeen()[id]
				assert.For(t, j, id).ThatActual(ok).IsTrue()
				assert.For(t, j, id).ThatActual(version).Equals(0)
			}
		}

		stackState := fmt.Sprint(stack.indexes)
		if stackState == lastStackState {
			//Stack was teh same before and after. That's suspicious...
			numShufflesTheSame++
		}

		lastStackState = stackState
	}

	//We set this high because there aren't THAT many items, so the same shuffle will happen somewhat often.
	if numShufflesTheSame > 3 {
		t.Error("When we shuffled", numShuffles, "times, got the same state", numShufflesTheSame, "which is suspicious")
	}

	sStack := deck.NewSizedStack(5).(*sizedStack)

	sStack.setState(fakeState)

	for _, c := range deck.Components() {
		sStack.insertNext(c.ImmutableInstance(fakeState))
	}

	//Number of shuffles that were the same (which is bad)
	numShufflesTheSame = 0

	//Reset lastIds to be for sStack but skip empty ones.
	lastIds = nil
	for _, id := range sStack.Ids() {
		if id == "" {
			continue
		}
		lastIds = append(lastIds, id)
	}

	assert.For(t).ThatActual(len(sStack.IdsLastSeen())).Equals(len(lastIds))

	lastStackState = fmt.Sprint(sStack.indexes)

	for i := 0; i < numShuffles; i++ {
		if err := sStack.Shuffle(); err != nil {
			t.Error("couldn't shuffle stack: ", err)
		}

		if i == 0 {
			//First time through, check that ids are scrambled correctly
			assert.For(t).ThatActual(len(sStack.IdsLastSeen())).Equals(len(lastIds) * 2)
			for j, id := range lastIds {
				if id == "" {
					continue
				}
				version, ok := sStack.IdsLastSeen()[id]
				assert.For(t, j, id).ThatActual(ok).IsTrue()
				assert.For(t, j, id).ThatActual(version).Equals(0)
			}
			for j, id := range sStack.Ids() {
				if id == "" {
					continue
				}
				version, ok := sStack.IdsLastSeen()[id]
				assert.For(t, j, id).ThatActual(ok).IsTrue()
				assert.For(t, j, id).ThatActual(version).Equals(0)
			}
		}

		stackState := fmt.Sprint(sStack.indexes)
		if stackState == lastStackState {
			//Stack was teh same before and after. That's suspicious...
			numShufflesTheSame++
		}

		lastStackState = stackState
	}

	//We set this high because there aren't THAT many items, so the same shuffle will happen somewhat often.
	if numShufflesTheSame > 3 {
		t.Error("When we shuffled", numShuffles, "times, got the same state", numShufflesTheSame, "which is suspicious")
	}

}

func TestMoveAllTo(t *testing.T) {
	game := testDefaultGame(t, false)

	deck := game.Manager().Chest().Deck("test")

	fakeState := &state{
		game: game,
	}

	to := deck.NewStack(1)
	to.setState(fakeState)

	from := deck.NewSizedStack(2)
	from.setState(fakeState)

	zero := deck.Components()[0]
	one := deck.Components()[1]

	from.insertNext(zero.ImmutableInstance(fakeState))

	//This should succeed because although to only has one slot, there's only
	//actually one item in from.
	if err := from.MoveAllTo(to); err != nil {
		t.Error("Unexpected error moving from sized stack to other stack", err)
	}

	if from.NumComponents() != 0 {
		t.Error("MoveAllTo did not vacate from")
	}

	if to.NumComponents() != 1 {
		t.Error("MoveAllTo did not move the components to other")
	}

	to = deck.NewStack(1)
	to.setState(fakeState)

	from = deck.NewSizedStack(2)
	from.setState(fakeState)

	from.insertNext(zero.ImmutableInstance(fakeState))
	from.insertNext(one.ImmutableInstance(fakeState))

	if err := from.MoveAllTo(to); err == nil {
		t.Error("Got no error moving from a stack that was too big.")
	}

}
