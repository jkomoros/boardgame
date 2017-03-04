package boardgame

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestInflate(t *testing.T) {
	game := testGame()

	game.SetUp(0)

	chest := game.Chest()

	testDeck := chest.Deck("test")

	gStack := NewGrowableStack(testDeck, 0)

	gStack.insertNext(testDeck.Components()[0])

	sStack := NewSizedStack(testDeck, 2)

	sStack.insertNext(testDeck.Components()[1])

	if gStack.ComponentAt(0) == nil {
		t.Error("Couldnt' get component from inflated gstack")
	}

	if sStack.ComponentAt(0) == nil {
		t.Error("Couldn't get component from inflated sstack")
	}

	if err := gStack.Inflate(chest); err == nil {
		t.Error("An inflated g stack was able to inflate again")
	}

	if err := sStack.Inflate(chest); err == nil {
		t.Error("An inflated s stack was able to inflate again")
	}

	gStackBlob, err := json.Marshal(gStack)

	if err != nil {
		t.Error("Gstack didn't serialize", err)
	}

	sStackBlob, err := json.Marshal(sStack)

	if err != nil {
		t.Error("SStack didn't serialize", err)
	}

	reGStack := &GrowableStack{}

	if err := json.Unmarshal(gStackBlob, reGStack); err != nil {
		t.Error("Couldn't reconstitute gStack", err)
	}

	reSStack := &SizedStack{}

	if err := json.Unmarshal(sStackBlob, reSStack); err != nil {
		t.Error("Couldn't reconstitute sStack", err)
	}

	if reGStack.Inflated() {
		t.Error("Reconstituted g stack thought it was inflated")
	}

	if reSStack.Inflated() {
		t.Error("Reconstituted s stack thought it was inflated")
	}

	if reGStack.ComponentAt(0) != nil {
		t.Error("Uninflated g stack still returned a component")
	}

	if reSStack.ComponentAt(0) != nil {
		t.Error("Uninflated s stack still returned a component")
	}

	if err := reGStack.Inflate(chest); err != nil {
		t.Error("Uninflated g stack wasn't able to inflate", err)
	}

	if err := reSStack.Inflate(chest); err != nil {
		t.Error("Uninflated s stack wasn't able to inflate", err)
	}

	if !reGStack.Inflated() {
		t.Error("After inflating g stack it didn't think it was inflated")
	}

	if !reSStack.Inflated() {
		t.Error("After inflating s stack it didn't think it was inflated")
	}

	c := reGStack.ComponentAt(0)

	if c != testDeck.Components()[0] {
		t.Error("After inflating g stack, got wrong component. Wanted", testDeck.Components()[0], "got", c)
	}

	c = reSStack.ComponentAt(0)

	if c != testDeck.Components()[1] {
		t.Error("After inflating s stack, got wrong component. Wanted", testDeck.Components()[1], "got", c)
	}
}

func TestMoveComponent(t *testing.T) {

	game := testGame()

	deck := game.Chest().Deck("test")

	gStack := NewGrowableStack(deck, 0)

	sStack := NewSizedStack(deck, 5)

	gStackMaxLen := NewGrowableStack(deck, 4)

	sStackMaxLen := NewSizedStack(deck, 4)

	for _, c := range deck.Components() {
		gStack.insertNext(c)
		gStackMaxLen.insertNext(c)
		sStack.insertNext(c)
		sStackMaxLen.insertNext(c)
	}

	fakeState := &State{}

	gStack.statePtr = fakeState
	sStack.statePtr = fakeState
	gStackMaxLen.statePtr = fakeState
	sStackMaxLen.statePtr = fakeState

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

	sStackOtherState := sStack.Copy()
	sStackOtherState.statePtr = &State{}

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
			FirstComponentIndex,
			0,
			FirstSlotIndex,
			4,
			false,
			"Move from growable first component to sized stack first slot",
		},
		{
			sStack,
			gStack,
			FirstSlotIndex,
			4,
			FirstSlotIndex,
			0,
			true,
			"Move an empty slot in sized stack to growable stack",
		},
		{
			sStack,
			gStack,
			FirstComponentIndex,
			0,
			LastSlotIndex,
			4,
			false,
			"Move first component in sized stack to growable stack",
		},
		{
			sStackOtherState,
			gStack,
			FirstComponentIndex,
			0,
			LastSlotIndex,
			4,
			true,
			"Move from a stack in one state to another",
		},
		{
			sStack,
			sStack,
			FirstComponentIndex,
			0,
			LastSlotIndex,
			4,
			true,
			"Moving from same stack",
		},
		{
			sStack,
			gStackMaxLen,
			FirstComponentIndex,
			0,
			LastSlotIndex,
			4,
			true,
			"Moving to a gstack with no more space",
		},
		{
			gStack,
			sStackMaxLen,
			FirstComponentIndex,
			0,
			LastSlotIndex,
			-1,
			true,
			"Moving from a growable stack to a slot that has no more space.",
		},
		{
			gStack,
			sStack,
			10,
			10,
			LastSlotIndex,
			4,
			true,
			"Invalid component index",
		},
		{
			gStack,
			sStack,
			2,
			2,
			LastSlotIndex,
			4,
			false,
			"Moving from middle of growable stack to sized stack",
		},
		{
			gStack,
			sStack,
			FirstComponentIndex,
			0,
			NextSlotIndex,
			4,
			false,
			"NextSlotIndex from growable to sized",
		},
		{
			sStack,
			gStack,
			FirstComponentIndex,
			0,
			NextSlotIndex,
			4,
			false,
			"NextSlotIndex from sized to growable",
		},
	}

	for i, test := range tests {
		var source Stack
		var destination Stack

		switch s := test.source.(type) {
		case *GrowableStack:
			source = s.Copy()
		case *SizedStack:
			source = s.Copy()
		}

		//Some tests deliberately want to make sure that copies within same source and dest aren't allowed
		if test.source == test.destination {
			destination = source
		} else {

			switch s := test.destination.(type) {
			case *GrowableStack:
				destination = s.Copy()
			case *SizedStack:
				destination = s.Copy()
			}
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

		if finalComponent := destination.ComponentAt(test.resolvedSlotIndex); finalComponent != component {
			t.Error("After the move, the component that was supposed to be moved was not moved to the target slot.", i, test.description)
		}
	}

}

func TestSwapComponents(t *testing.T) {
	game := testGame()

	deck := game.Chest().Deck("test")

	stack := NewGrowableStack(deck, 0)

	for _, c := range deck.Components() {
		stack.insertNext(c)
	}

	swapComponentsTests(stack, t)

	sStack := NewSizedStack(deck, 10)

	for _, c := range deck.Components() {
		stack.insertNext(c)
	}

	swapComponentsTests(sStack, t)

}

func swapComponentsTests(stack Stack, t *testing.T) {
	if err := stack.SwapComponents(0, 1); err == nil {
		t.Error("Stack with no state allowed a swap")
	}

	fakeState := &State{}

	switch s := stack.(type) {
	case *GrowableStack:
		s.statePtr = fakeState
	case *SizedStack:
		s.statePtr = fakeState
	default:
		t.Fatal("Unknown type of stack")
	}

	zero := stack.ComponentAt(0)
	one := stack.ComponentAt(1)

	if err := stack.SwapComponents(0, 1); err != nil {
		t.Error("Legal swap not allowed")
	}

	if stack.ComponentAt(0) != one {
		t.Error("Swap did not actually position of #1")
	}

	if stack.ComponentAt(1) != zero {
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

	game := testGame()

	deck := game.Chest().Deck("test")

	fakeState := &State{}

	stack := NewGrowableStack(deck, 0)

	stack.statePtr = fakeState

	for _, c := range deck.Components() {
		stack.insertNext(c)
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
		stackCopy := stack.Copy()

		component := deck.ComponentAt(test.componentDeckIndex)

		if !reflect.DeepEqual(stackCopy.indexes, startingIndexes) {
			t.Error("Sanity check failed", i, "Starting indexes were", stackCopy.indexes, "wanted", startingIndexes)
		}

		stackCopy.insertComponentAt(test.slotIndex, component)

		if !reflect.DeepEqual(stackCopy.indexes, test.expectedIndexes) {
			t.Error("Test", i, test.description, "failed for insertComponentAt. Got", stackCopy.indexes, "wanted", test.expectedIndexes)
		}
	}
}

func TestGrowableStackRemoveComponentAt(t *testing.T) {
	//Splicing out parts of an array is so finicky that we need to make sure
	//to test it extra good...

	game := testGame()

	deck := game.Chest().Deck("test")

	fakeState := &State{}

	stack := NewGrowableStack(deck, 0)

	stack.statePtr = fakeState

	for _, c := range deck.Components() {
		stack.insertNext(c)
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
		stackCopy := stack.Copy()

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
	game := testGame()

	deck := game.Chest().Deck("test")

	stack := NewGrowableStack(deck, 0)

	fakeState := &State{}

	stack.statePtr = fakeState

	for _, c := range deck.Components() {
		stack.insertNext(c)
	}

	//The number of shuffles to do
	numShuffles := 10

	//Number of shuffles that were the same (which is bad)
	numShufflesTheSame := 0

	lastStackState := fmt.Sprint(stack.indexes)

	for i := 0; i < numShuffles; i++ {
		if err := stack.Shuffle(); err != nil {
			t.Error("Shuffle failed", err)
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

	sStack := NewSizedStack(deck, 5)

	sStack.statePtr = fakeState

	for _, c := range deck.Components() {
		sStack.insertNext(c)
	}

	//Number of shuffles that were the same (which is bad)
	numShufflesTheSame = 0

	lastStackState = fmt.Sprint(sStack.indexes)

	for i := 0; i < numShuffles; i++ {
		if err := sStack.Shuffle(); err != nil {
			t.Error("couldn't shuffle stack: ", err)
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
	game := testGame()

	deck := game.Chest().Deck("test")

	fakeState := &State{}

	to := NewGrowableStack(deck, 1)
	to.statePtr = fakeState

	from := NewSizedStack(deck, 2)
	from.statePtr = fakeState

	zero := deck.Components()[0]
	one := deck.Components()[1]

	from.insertNext(zero)

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

	to = NewGrowableStack(deck, 1)

	from = NewSizedStack(deck, 2)

	from.insertNext(zero)
	from.insertNext(one)

	if err := from.MoveAllTo(to); err == nil {
		t.Error("Got no error moving from a stack that was too big.")
	}

}
