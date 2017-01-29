package boardgame

import (
	"testing"
)

//TODO: this should probably be somewhere more central.
func componentsEqual(one Component, two Component) bool {
	if one == nil && two == nil {
		return true
	}
	if one == nil || two == nil {
		return false
	}
	if one.Deck() != two.Deck() {
		return false
	}
	if one.DeckIndex() != two.DeckIndex() {
		return false
	}
	return true
}

//Begin tests

func TestStackInsert(t *testing.T) {

	game := testGame()

	chest := game.Chest

	stack := &Stack{
		game,
		"test",
		[]int{},
	}

	if stack.Len() != 0 {
		t.Error("Empty stack didn't report empty")
	}

	if stack.DeckName != "test" {
		t.Error("Stack didn't report right deck name")
	}

	if stack.ComponentAt(0) != nil {
		t.Error("Stack returned something even though it was empty:", stack.ComponentAt(0))
	}

	one := chest["test"].Components[0]
	two := chest["test"].Components[1]

	if componentsEqual(one, two) {
		t.Error("Two components that are not equal were thought to be equal", one, two)
	}

	if !componentsEqual(one, one) {
		t.Error("Two components that are equal were not thought to be equal")
	}

	stack.InsertFront(two)

	if stack.Len() != 1 {
		t.Error("Insertinga  component didn't increase len to 1")
	}

	if !componentsEqual(stack.ComponentAt(0), two) {
		t.Error("The component that we inserted did not come back", stack.ComponentAt(0), two, stack)
	}

	stack.InsertFront(one)

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
