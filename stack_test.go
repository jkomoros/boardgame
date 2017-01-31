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

	deck := chest.Deck("test")

	one := deck.Components()[0]
	two := deck.Components()[1]

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

	values := testingComponentValues(stack.ComponentValues())

	if len(values) != 2 {
		t.Error("stack.ComponentValues returned wrong len", len(values), "wanted", 2)
	}

	for i := 0; i < 2; i++ {
		if values[i] != deck.Components()[i].Values.(*testingComponent) {
			t.Error("Got wrong value out of stack.Components at", i, "got", values[i], "wanted", deck.Components()[i].Values.(*testingComponent))
		}
	}

	component := stack.RemoveFirst()

	if component != one {
		t.Error("REmoving component from front was wrong component. Got", component, "wanted", one)
	}

	if stack.Len() != 1 {
		t.Error("Removing a component from front did not decrement len.")
	}

	component = stack.ComponentAt(0)

	if component != two {
		t.Error("Removing a component didn't move the other component down. Got", component, "wanted", two)
	}

	component = stack.RemoveFirst()

	if component != two {
		t.Error("removing last component didn't return the right one. Got", component, "wanted", two)
	}

	if stack.Len() != 0 {
		t.Error("After removing two components the stack wasn't 0", stack.Len())
	}

	component = stack.RemoveFirst()

	if component != nil {
		t.Error("Was still able to remove a component even though it was empty.", component)
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

	values := testingComponentValues(stack.ComponentValues())

	if len(values) != stackSize {
		t.Error("stack.ComponentValues returned wrong len", len(values), "wanted", stackSize)
	}

	for i := 0; i < stackSize; i++ {
		if i == 2 {
			if values[i] != nil {
				t.Error("Expected nil at 2")
			}
			continue
		}
		if values[i] != deck.Components()[i].Values.(*testingComponent) {
			t.Error("Got wrong value out of stack.Components at", i, "got", values[i], "wanted", deck.Components()[i].Values.(*testingComponent))
		}
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

	values = testingComponentValues(stack.ComponentValues())

	if len(values) != stackSize {
		t.Error("stack.ComponentValues returned wrong len", len(values), "wanted", stackSize)
	}

	for i := 0; i < stackSize; i++ {
		if values[i] != deck.Components()[i].Values.(*testingComponent) {
			t.Error("Got wrong value out of stack.Components at", i, "got", values[i], "wanted", deck.Components()[i].Values.(*testingComponent))
		}
	}

	component := stack.RemoveFirst()

	if component != deck.Components()[0] {
		t.Error("Removing first componetn didn't give first component got", component, "wanted", deck.Components()[0])
	}

	if stack.SlotsRemaining() != stackSize-2 {
		t.Error("SlotsRemaining didn't change when removing first item")
	}

	component = stack.ComponentAt(0)

	if component != nil {
		t.Error("After removing a component from a slot there were other components in that slot", component)
	}

	component = stack.ComponentAt(1)

	if component != deck.Components()[1] {
		t.Error("AFter removing a component from slot one, the second component was not what we expected", component, "wanted", deck.Components()[1])
	}

	component = stack.RemoveFirst()

	if component != deck.Components()[1] {
		t.Error("Removing a second component didn't give right item", component, "wanted", deck.Components()[1])
	}

	if err := stack.InsertFront(deck.Components()[0]); err != nil {
		t.Error("Couldn't insert an item even though we'd removed two", err)
	}

	component = stack.ComponentAt(0)

	if component != deck.Components()[0] {
		t.Error("After inserting again, the first component was not in the slot we wanted", component, deck.Components()[0])
	}

	component = stack.RemoveFirst()

	if component != deck.Components()[0] {
		t.Error("After removing again, we got wrong component.", component, "wanted", deck.Components()[0])
	}

	component = stack.RemoveFirst()

	if component != deck.Components()[2] {
		t.Error("Removing final item didn't give what we expected", component, "wanted", deck.Components()[2])
	}

	component = stack.RemoveFirst()

	if component != nil {
		t.Error("removefirst from empty stack didn't give us nil", component)
	}

}
