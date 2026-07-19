package boardgame

// StackConstraint is a function checked before component(s) are inserted into
// a destination stack. If the function returns a non-nil error, the move is
// rejected and the destination stack is left unmodified. Constraints are added
// to a Stack via AddConstraint and are checked automatically whenever
// components move into that stack through MoveTo, MoveToNextSlot, MoveAllTo,
// or any other method that uses moveComponentImpl. They are also checked
// automatically during Legal() for moves that implement SourceStacker and
// DestinationStacker interfaces.
//
// destination is the stack that the component(s) would be added to, in its
// pre-insertion state (the proposed components are NOT yet in the stack).
// proposed contains the component(s) that are proposed to be added.
// state is the ImmutableState that the stack is part of.
//
// Constraint implementations must account for the fact that proposed
// components are not yet in the destination stack. For example, to check a
// maximum count, use dest.NumComponents() + len(proposed).
//
// Constraints must be deterministic, copy-stable predicates. They may base
// their result only on logical values reachable through destination, proposed,
// and state. They must not modify supplied or captured state, consume
// randomness, schedule callbacks, perform I/O, modify captured values, inspect
// pointer identity, consult clocks, depend on invocation count, or retain the
// supplied objects after returning. The engine may evaluate a constraint on a
// copied state, and may evaluate it independently during Legal and Apply.
//
// Constraints must not panic.
//
// Constraints are NOT checked during initial game setup
// (DistributeComponentToStarterStack), because that uses insertComponentAt
// directly rather than moveComponentImpl.
//
// Constraints are set at setup time (via struct tags or in FinishSetUp) and
// don't need individual removal. Use ClearConstraints to reset all
// constraints on a stack.
type StackConstraint func(destination ImmutableStack, proposed []ImmutableComponentInstance, state ImmutableState) error

// StackConstraintConstructor is used for struct-tag-based constraint
// configuration. Name is the identifier used in struct tags (e.g. "max" for
// `sizedstack:"tokens,9,max(1)"`). Constructor parses the tag arguments and
// returns a StackConstraint. It receives the ComponentChest so it can
// resolve constant names to values.
type StackConstraintConstructor struct {
	Name        string
	Constructor func(args []string, chest *ComponentChest) (StackConstraint, error)
}
