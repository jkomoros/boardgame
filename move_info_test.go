package boardgame

import (
	"strings"
	"testing"
)

func TestMoveInfoConcreteMoveNilSafe(t *testing.T) {
	var info *MoveInfo
	if got := info.ConcreteMove(); got != nil {
		t.Fatalf("nil MoveInfo ConcreteMove() = %T; want nil", got)
	}
}

func TestMoveInfoConcreteMoveIdentity(t *testing.T) {
	manager := newTestGameManger(t)
	one := manager.ExampleMoveByName("Test")
	two := manager.ExampleMoveByName("Test")

	if one == nil || two == nil {
		t.Fatal("expected two constructed example moves")
	}
	if one == two {
		t.Fatal("separate constructions returned the same Move instance")
	}
	if one.Info() == two.Info() {
		t.Fatal("separate Move instances shared one MoveInfo")
	}
	if got := one.Info().ConcreteMove(); got != one {
		t.Fatalf("first ConcreteMove() = %T %p; want original %T %p", got, got, one, one)
	}
	if got := two.Info().ConcreteMove(); got != two {
		t.Fatalf("second ConcreteMove() = %T %p; want original %T %p", got, got, two, two)
	}
	describer, ok := one.(interface{ Description() string })
	if !ok {
		t.Fatalf("constructed move %T does not expose its embedded Description method", one)
	}
	if got, want := describer.Description(), one.HelpText(); got != want {
		t.Fatalf("embedded Description() = %q; want final concrete HelpText() %q", got, want)
	}
}

type discardingMoveInfoMove struct {
	testMove
}

func (m *discardingMoveInfoMove) SetInfo(_ *MoveInfo) {}

func TestMoveConstructorMustRetainCanonicalInfo(t *testing.T) {
	config := NewMoveConfig(
		"Discarding MoveInfo",
		func() Move { return new(discardingMoveInfoMove) },
		nil,
	)

	_, err := newMoveType(config, nil)
	if err == nil || !strings.Contains(err.Error(), "preserve the MoveInfo affiliation") {
		t.Fatalf("newMoveType() error = %v; want canonical MoveInfo error", err)
	}
}

type copyingMoveInfoMove struct {
	testMove
	copiedInfo MoveInfo
}

func (m *copyingMoveInfoMove) SetInfo(info *MoveInfo) {
	m.copiedInfo = *info
}

func (m *copyingMoveInfoMove) Info() *MoveInfo {
	return &m.copiedInfo
}

func TestMoveConstructorMayCopyMoveInfoValue(t *testing.T) {
	config := NewMoveConfig(
		"Copying MoveInfo",
		func() Move { return new(copyingMoveInfoMove) },
		nil,
	)
	move, err := NewOrphanMove(config)
	if err != nil {
		t.Fatalf("NewOrphanMove() error = %v", err)
	}
	if got := move.Info().ConcreteMove(); got != move {
		t.Fatalf("copied MoveInfo ConcreteMove() = %T %p; want %T %p", got, got, move, move)
	}
}

func TestNewOrphanMoveProvidesStandaloneAffiliation(t *testing.T) {
	move, err := NewOrphanMove(testMoveConfig)
	if err != nil {
		t.Fatalf("NewOrphanMove() error = %v", err)
	}
	if move.Info() == nil || move.Info().ConcreteMove() != move {
		t.Fatal("orphan move did not receive its concrete runtime affiliation")
	}
	if got, want := move.Info().Name(), testMoveConfig.Name(); got != want {
		t.Fatalf("orphan move name = %q; want %q", got, want)
	}
}

func TestMoveConstructorRejectsTypedNil(t *testing.T) {
	config := NewMoveConfig(
		"Typed Nil Move",
		func() Move {
			var move *testMove
			return move
		},
		nil,
	)

	_, err := newMoveType(config, nil)
	if err == nil || !strings.Contains(err.Error(), "non-nil pointer") {
		t.Fatalf("newMoveType() error = %v; want non-nil pointer error", err)
	}
}

type defaultsObserveConcreteMove struct {
	testMove
	sawConcreteSelf bool
}

func (m *defaultsObserveConcreteMove) DefaultsForState(ImmutableState) {
	m.sawConcreteSelf = m.Info() != nil && m.Info().ConcreteMove() == m
}

func TestConcreteMoveAffiliatedBeforeDefaults(t *testing.T) {
	config := NewMoveConfig(
		"Observe Concrete Move",
		func() Move { return new(defaultsObserveConcreteMove) },
		nil,
	)
	moveType, err := newMoveType(config, nil)
	if moveType == nil || err == nil || err.Error() != newMoveTypeErrNoManagerPassed {
		t.Fatalf("newMoveType() = (%v, %v); want usable managerless move type", moveType, err)
	}

	move, ok := moveType.NewMove(newTestGameManger(t).ExampleState()).(*defaultsObserveConcreteMove)
	if !ok || move == nil {
		t.Fatalf("NewMove() = %T; want *defaultsObserveConcreteMove", move)
	}
	if !move.sawConcreteSelf {
		t.Fatal("ConcreteMove was not affiliated before DefaultsForState")
	}
}
