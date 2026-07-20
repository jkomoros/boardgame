/*
Package boardgame is a framework that makes it possible to build boardgames
with minimial fuss.

This package contains the core boardgame logic engine. Other packages extend
and build of of this base. package boardgame/base provides base
implementations of the various types of objects your game logic must provide.
boardgame/moves provides a collection of powerful move objects to base your
own logic off of. boardgame/server is a package that, given a game definition,
creates a powerful Progressive Web App, complete with automatically-generated
animations.

The boardgame/boardgame-util command is a powerful swiss army knife of
functionality to help create game packages and run servers based on that
automatically.

The documentation in this package is primarily detail about how the various
concepts wire together. For a high-level overview of how everything works and
tour of the main concepts, see TUTORIAL.md.

StackConstraints provide a declarative way to express invariants on what a
stack will accept. Constraint functions are checked automatically before
components move into a stack, and the move is rejected if violated. See
StackConstraint for details. The constraints sub-package provides pre-built
constraints like MaxNumComponents, Unique, Same, and MaxDistinctValues.
Constraints can be added programmatically via Stack.AddConstraint, or
declaratively via struct tags. The default base.GameDelegate includes
constructors for all pre-built constraints; override
ConfigureStackConstraintConstructors only to add custom types.

Constraints are validation predicates, not event callbacks. They may capture
immutable configuration, but must resolve runtime game objects through their
supplied arguments. They must be deterministic, free of side effects, and
produce the same result for a logical state and its copy. This lets
MayMoveCountTo, MayMoveAllTo, and their transactional mutating counterparts
validate an ordered multi-component transfer on a disposable state before
committing it.

ImmutableComponentInstance provides "May" methods for pre-validating
component moves in Legal() before actually performing them in Apply():

  - MayMoveTo(dest) checks whether a component could move to the
    destination stack (slot-independent). It covers all of MoveTo,
    SecretMoveTo, MoveToFirstSlot, MoveToLastSlot, and MoveToNextSlot.
  - MayMoveToSlot(dest, slotIndex) additionally checks that a specific
    slot is valid and available.

ImmutableStack also provides MayMoveCountTo(dest, count), MayMoveAllTo(dest),
and MaySwapComponents(i, j) for pre-validating those operations. Pair
MayMoveCountTo with Stack.MoveCountTo when exactly N components should move as
one notional move; use moves.MoveCountComponents when they should instead have
separate persistence and animation boundaries.

These methods are designed so that if Legal() passes, Apply() will succeed
for the corresponding operation. The moves package uses them internally,
and game authors should use them in custom Legal() methods to replace ad-hoc
manual checks.

The primary entry point for use of this package is defining your own
GameDelegate. The methods and documentation from there will point to other
parts of this package.
*/
package boardgame
