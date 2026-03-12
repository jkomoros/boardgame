package boardgame

// StackConstraint is a function checked after component(s) are tentatively
// inserted into a destination stack. If the function returns a non-nil error,
// the move is rejected and the component is rolled back to its source
// position. Constraints are added to a Stack via AddConstraint and are
// checked automatically whenever components move into that stack through
// MoveTo, MoveToNextSlot, MoveAllTo, or any other method that uses
// moveComonentImpl.
//
// destination is the stack that the component(s) were just added to.
// justAdded contains the component(s) that were just added.
// state is the ImmutableState that the stack is part of.
//
// Constraints must be pure functions: they must not modify any state or
// produce side effects. If a constraint rejects a move, only the component
// move itself is rolled back — any side effects from the constraint function
// will persist.
//
// Constraints must not panic. A panic inside a constraint will prevent the
// rollback of the tentative component move, leaving the stack in an
// inconsistent state.
//
// Constraints are NOT checked during initial game setup
// (DistributeComponentToStarterStack), because that uses insertComponentAt
// directly rather than moveComonentImpl.
//
// Constraints are set at setup time (via struct tags or in FinishSetUp) and
// don't need individual removal. Use ClearConstraints to reset all
// constraints on a stack.
type StackConstraint func(destination ImmutableStack, justAdded []ImmutableComponentInstance, state ImmutableState) error

// StackConstraintConstructor is used for struct-tag-based constraint
// configuration. Name is the identifier used in struct tags (e.g. "max" for
// `sizedstack:"tokens,9,max(1)"`). Constructor parses the tag arguments and
// returns a StackConstraint. It receives the ComponentChest so it can
// resolve constant names to values.
type StackConstraintConstructor struct {
	Name        string
	Constructor func(args []string, chest *ComponentChest) (StackConstraint, error)
}
