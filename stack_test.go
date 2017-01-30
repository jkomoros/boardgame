package boardgame

import (
	"math"
	"testing"
)

func TestStackInsert(t *testing.T) {

	game := testGame()

	chest := game.Chest()

	stack := NewGrowableStack(chest.Deck("test"), 0)

	if stack.SlotsRemaining() != math.MaxInt64 {
		t.Error("A stack with no cap reported a non-huge SlotsRemaining")
	}

	if stack.Len() != 0 {
		t.Error("Empty stack didn't report empty")
	}

	if stack.deck.Name() != "test" {
		t.Error("Stack didn't report right deck name")
	}

	if stack.ComponentAt(0) != nil {
		t.Error("Stack returned something even though it was empty:", stack.ComponentAt(0))
	}

	one := chest.Deck("test").Components()[0]
	two := chest.Deck("test").Components()[1]

	if componentsEqual(one, two) {
		t.Error("Two components that are not equal were thought to be equal", one, two)
	}

	if !componentsEqual(one, one) {
		t.Error("Two components that are equal were not thought to be equal")
	}

	if err := stack.InsertFront(two); err != nil {
		t.Error("Got unexpected error when inserting:", err)
	}

	if stack.Len() != 1 {
		t.Error("Insertinga  component didn't increase len to 1")
	}

	if !componentsEqual(stack.ComponentAt(0), two) {
		t.Error("The component that we inserted did not come back", stack.ComponentAt(0), two, stack)
	}

	if err := stack.InsertFront(one); err != nil {
		t.Error("Got unexpected error when inserting:", err)
	}

	if stack.Len() != 2 {
		t.Error("Inserting a second component didn't increase len to 2")
	}

	if !componentsEqual(stack.ComponentAt(0), one) {
		t.Error("Inseting front didn't put right component at front")
	}

	if !componentsEqual(stack.ComponentAt(1), two) {
		t.Error("Inserting front didn't move the previous front back by one")
	}

}

func TestStackCap(t *testing.T) {
	game := testGame()

	stack := NewGrowableStack(game.Chest().Deck("test"), 2)

	deck := game.Chest().Deck("test")

	if stack.SlotsRemaining() != 2 {
		t.Error("An empty stack with cap 2 reported wrong slots remaining. Got", stack.SlotsRemaining(), "wanted 2")
	}

	if err := stack.InsertFront(deck.Components()[0]); err != nil {
		t.Error("got unexpected error on insertion", err)
	}

	if stack.SlotsRemaining() != 1 {
		t.Error("A stack with one item cap 2 reported wrong slots remaining. Got", stack.SlotsRemaining(), "wanted 1")
	}

	if err := stack.InsertFront(deck.Components()[1]); err != nil {
		t.Error("Got unexpected error on insertion:", err)
	}

	if stack.SlotsRemaining() != 0 {
		t.Error("A stack with two items cap two said it still had slots left. Got", stack.SlotsRemaining(), "wanted 0")
	}

	if err := stack.InsertFront(deck.Components()[2]); err == nil {
		t.Error("Inserting after a stack hit its cap succeeded")
	}

	if stack.Len() > 2 {
		t.Error("InsertFront after a stack had hit its cap succeeded")
	}
}

func TestSizedStack(t *testing.T) {
	game := testGame()

	deck := game.Chest().Deck("test")

	stackSize := 3

	stack := NewSizedStack(deck, stackSize)

	if stack.Len() != stackSize {
		t.Error("Sized stack had wrong len. Got", stack.Len(), "wanted", stackSize)
	}

	for i := 0; i < stackSize; i++ {
		component := stack.ComponentAt(i)

		if component != nil {
			t.Error("Expected nil component but got one at", i, "got", component, "wanted nil")
		}
	}

	if stack.SlotsRemaining() != stackSize {
		t.Error("Got wrong SlotsRemaining for empty stack. Got", stack.SlotsRemaining(), "wanted", stackSize)
	}

	if err := stack.InsertAtSlot(deck.Components()[1], 1); err != nil {
		t.Error("Insertion unexpectedly failed", err)
	}

	if stack.SlotsRemaining() != stackSize-1 {
		t.Error("After inserting a component, slots remaining was wrong. Got", stack.SlotsRemaining(), "wanted", stackSize-1)
	}

	if stack.ComponentAt(1) != deck.Components()[1] {
		t.Error("Got wrong component out. Got", stack.ComponentAt(1), "wanted", deck.Components()[1])
	}

	if err := stack.InsertFront(deck.Components()[0]); err != nil {
		t.Error("Insertion unexpectedly failed", err)
	}

	if stack.SlotsRemaining() != stackSize-2 {
		t.Error("After inserting a component, slots remaining was wrong. Got", stack.SlotsRemaining(), "wanted", stackSize-2)
	}

	if stack.ComponentAt(0) != deck.Components()[0] {
		t.Error("Stack InsertFirstEmptySlot put a component in the wrong slot.")
	}

	if err := stack.InsertFront(deck.Components()[2]); err != nil {
		t.Error("Insertion unexpectedly failed", err)
	}

	if stack.SlotsRemaining() != stackSize-3 {
		t.Error("After inserting a component, slots remaining was wrong. Got", stack.SlotsRemaining(), "wanted", stackSize-3)
	}

	if stack.ComponentAt(2) != deck.Components()[2] {
		t.Error("Stack insertnextemptyslot didn't insert the item at the right slot")
	}

	if err := stack.InsertFront(deck.Components()[3]); err == nil {
		t.Error("Trying to insert a compnent after there were no more slots succeeded")
	}

	if err := stack.InsertAtSlot(deck.Components()[3], 0); err == nil {
		t.Error("Trying to insert a component at a taken slot succeeded")
	}

}
