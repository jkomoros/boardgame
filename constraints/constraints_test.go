package constraints_test

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/constraints"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
)

func newTestGame(t *testing.T) *boardgame.Game {
	t.Helper()
	manager, err := boardgame.NewGameManager(tictactoe.NewDelegate(), memory.NewStorageManager())
	assert.For(t).ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t).ThatActual(err).IsNil()
	manager.Internals().AllowMutableConstraints(game)
	return game
}

// getStack returns a mutable stack from the game or player state by name.
func getStack(st boardgame.ImmutableState, group string, name string, playerIdx int) boardgame.Stack {
	mutableSt := st.(boardgame.State)
	switch group {
	case "game":
		rs := mutableSt.GameState().ReadSetter()
		s, _ := rs.StackProp(name)
		return s
	case "player":
		rs := mutableSt.PlayerStates()[playerIdx].ReadSetter()
		s, _ := rs.StackProp(name)
		return s
	}
	return nil
}

func TestMaxNumComponents(t *testing.T) {
	game := newTestGame(t)
	st := game.CurrentState()

	slots := getStack(st, "game", "Slots", 0)
	p0Unused := getStack(st, "player", "UnusedTokens", 0)

	assert.For(t).ThatActual(slots).IsNotNil()
	assert.For(t).ThatActual(p0Unused).IsNotNil()
	assert.For(t, "slots empty at start").ThatActual(slots.NumComponents()).Equals(0)
	assert.For(t, "player has tokens").ThatActual(p0Unused.NumComponents() > 0).Equals(true)

	slots.AddConstraint(constraints.MaxNumComponents(1))

	// First move should succeed.
	err := p0Unused.First().MoveTo(slots, slots.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(slots.NumComponents()).Equals(1)

	// Second should fail.
	nextSlot := -1
	for i := 0; i < slots.Len(); i++ {
		if slots.ComponentAt(i) == nil {
			nextSlot = i
			break
		}
	}
	assert.For(t, "found empty slot").ThatActual(nextSlot >= 0).Equals(true)

	err = p0Unused.First().MoveTo(slots, nextSlot)
	assert.For(t, "second move rejected").ThatActual(err).IsNotNil()
	assert.For(t, "still 1 component").ThatActual(slots.NumComponents()).Equals(1)
}

func TestSameConstraint(t *testing.T) {
	game := newTestGame(t)
	st := game.CurrentState()

	slots := getStack(st, "game", "Slots", 0)
	p0Unused := getStack(st, "player", "UnusedTokens", 0)
	p1Unused := getStack(st, "player", "UnusedTokens", 1)

	// The "Value" property on playerToken distinguishes X from O.
	slots.AddConstraint(constraints.Same("Value"))

	// Move player 0's token (Value="X") in.
	err := p0Unused.First().MoveTo(slots, 0)
	assert.For(t).ThatActual(err).IsNil()

	// Move another of player 0's tokens in — same value → OK.
	err = p0Unused.First().MoveTo(slots, 1)
	assert.For(t).ThatActual(err).IsNil()

	// Move player 1's token (Value="O") — different → rejected.
	err = p1Unused.First().MoveTo(slots, 2)
	assert.For(t, "different value rejected").ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(slots.NumComponents()).Equals(2)
}

func TestMaxDistinctValuesConstraint(t *testing.T) {
	game := newTestGame(t)
	st := game.CurrentState()

	slots := getStack(st, "game", "Slots", 0)
	p0Unused := getStack(st, "player", "UnusedTokens", 0)
	p1Unused := getStack(st, "player", "UnusedTokens", 1)

	// Allow at most 1 distinct value for the "Value" property.
	slots.AddConstraint(constraints.MaxDistinctValues("Value", 1))

	err := p0Unused.First().MoveTo(slots, 0)
	assert.For(t).ThatActual(err).IsNil()

	err = p0Unused.First().MoveTo(slots, 1)
	assert.For(t).ThatActual(err).IsNil()

	// Different player's token — 2 distinct values > 1 → rejected.
	err = p1Unused.First().MoveTo(slots, 2)
	assert.For(t, "too many distinct values").ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(slots.NumComponents()).Equals(2)
}

func TestUniqueConstraint(t *testing.T) {
	game := newTestGame(t)
	st := game.CurrentState()

	slots := getStack(st, "game", "Slots", 0)
	p0Unused := getStack(st, "player", "UnusedTokens", 0)

	// All of player 0's tokens have the same Value ("X").
	// So constraints.Unique("Value") should reject the second one.
	slots.AddConstraint(constraints.Unique("Value"))

	err := p0Unused.First().MoveTo(slots, 0)
	assert.For(t).ThatActual(err).IsNil()

	err = p0Unused.First().MoveTo(slots, 1)
	assert.For(t, "duplicate value rejected").ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(slots.NumComponents()).Equals(1)
}

func TestMaxNumComponentsConstructorParsing(t *testing.T) {
	c := constraints.MaxNumComponentsConstructor()
	assert.For(t).ThatActual(c.Name).Equals("max")

	constraint, err := c.Constructor([]string{"3"}, nil)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(constraint).IsNotNil()

	_, err = c.Constructor([]string{"1", "2"}, nil)
	assert.For(t, "too many args").ThatActual(err).IsNotNil()

	_, err = c.Constructor([]string{"abc"}, nil)
	assert.For(t, "not an int").ThatActual(err).IsNotNil()
}

func TestDefaultConstructors(t *testing.T) {
	constructors := constraints.DefaultConstructors()
	assert.For(t).ThatActual(len(constructors)).Equals(4)

	names := make(map[string]bool)
	for _, c := range constructors {
		names[c.Name] = true
	}
	assert.For(t).ThatActual(names["max"]).Equals(true)
	assert.For(t).ThatActual(names["unique"]).Equals(true)
	assert.For(t).ThatActual(names["same"]).Equals(true)
	assert.For(t).ThatActual(names["maxdistinct"]).Equals(true)
}

func newTestChest(t *testing.T) *boardgame.ComponentChest {
	t.Helper()
	manager, err := boardgame.NewGameManager(tictactoe.NewDelegate(), memory.NewStorageManager())
	assert.For(t).ThatActual(err).IsNil()
	return manager.Chest()
}

func TestPropPathValidation(t *testing.T) {
	chest := newTestChest(t)

	// --- UniqueConstructor ---
	uc := constraints.UniqueConstructor()

	// Valid property name should succeed.
	_, err := uc.Constructor([]string{"Value"}, chest)
	assert.For(t, "unique with valid prop").ThatActual(err).IsNil()

	// Typo should fail at construction time.
	_, err = uc.Constructor([]string{"Valeu"}, chest)
	assert.For(t, "unique with typo").ThatActual(err).IsNotNil()

	// Prefixed valid name should succeed.
	_, err = uc.Constructor([]string{"component.Value"}, chest)
	assert.For(t, "unique with component prefix").ThatActual(err).IsNil()

	// Prefixed typo should fail.
	_, err = uc.Constructor([]string{"component.Valeu"}, chest)
	assert.For(t, "unique with component prefix typo").ThatActual(err).IsNotNil()

	// Dynamic prefix with nonexistent prop should fail (tictactoe has no dynamic values).
	_, err = uc.Constructor([]string{"dynamic.Value"}, chest)
	assert.For(t, "unique with dynamic prefix on game without dynamic values").ThatActual(err).IsNotNil()

	// --- SameConstructor ---
	sc := constraints.SameConstructor()

	_, err = sc.Constructor([]string{"Value"}, chest)
	assert.For(t, "same with valid prop").ThatActual(err).IsNil()

	_, err = sc.Constructor([]string{"colour"}, chest)
	assert.For(t, "same with typo").ThatActual(err).IsNotNil()

	// --- MaxDistinctValuesConstructor ---
	mdc := constraints.MaxDistinctValuesConstructor()

	_, err = mdc.Constructor([]string{"Value", "2"}, chest)
	assert.For(t, "maxdistinct with valid prop").ThatActual(err).IsNil()

	_, err = mdc.Constructor([]string{"colour", "2"}, chest)
	assert.For(t, "maxdistinct with typo").ThatActual(err).IsNotNil()

	// --- nil chest should skip validation (programmatic use) ---
	_, err = uc.Constructor([]string{"anything"}, nil)
	assert.For(t, "unique with nil chest skips validation").ThatActual(err).IsNil()

	_, err = sc.Constructor([]string{"anything"}, nil)
	assert.For(t, "same with nil chest skips validation").ThatActual(err).IsNil()

	_, err = mdc.Constructor([]string{"anything", "3"}, nil)
	assert.For(t, "maxdistinct with nil chest skips validation").ThatActual(err).IsNil()
}

func TestExtendDefaults(t *testing.T) {
	custom := &boardgame.StackConstraintConstructor{
		Name: "custom",
		Constructor: func(args []string, chest *boardgame.ComponentChest) (boardgame.StackConstraint, error) {
			return nil, nil
		},
	}

	extended := constraints.ExtendDefaults(custom)
	assert.For(t, "extended length").ThatActual(len(extended)).Equals(5)
	assert.For(t, "custom is last").ThatActual(extended[4].Name).Equals("custom")

	// Verify DefaultConstructors is not mutated.
	assert.For(t, "defaults unchanged").ThatActual(len(constraints.DefaultConstructors())).Equals(4)
}
