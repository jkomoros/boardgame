/*
Package constraints provides pre-built StackConstraint constructors for
common patterns. These constraints can be used programmatically via
Stack.AddConstraint, or declaratively via struct tags on stack fields.

# Available Constraints

MaxNumComponents(max) rejects a move if the stack would contain more than
max components after the addition.

Unique(propPath) rejects a move if any two components in the stack share
the same value for the named property.

Same(propPath) rejects a move if not all components in the stack have the
same value for the named property.

MaxDistinctValues(propPath, max) rejects a move if the stack would contain
more than max distinct values for the named property.

# Property Path Syntax

The propPath parameter follows a simple convention:

  - "color" — checks Component.Values() first, then falls back to
    ImmutableDynamicValues()
  - "component.color" — only checks Component.Values()
  - "dynamic.color" — only checks ImmutableDynamicValues()

The property is read via the generic Prop() method on PropertyReader, and
values are compared using fmt.Sprintf("%v", val) for equality.

# Struct Tag Syntax

Constraint expressions can be added after the deck name (and size, for
sizedstack) in struct tags. Arguments within a constraint use commas,
just like the top-level tag fields — the tag parser is parenthesis-aware,
so commas inside parentheses are treated as constraint arguments:

	`stack:"cards,max(3)"`
	`sizedstack:"cards,5,max(3)"`
	`sizedstack:"cards,5,maxdistinct(color,2)"`
	`sizedstack:"cards,5,max(3),unique(color)"`

Struct-tag constraints work out of the box: the default base.GameDelegate
returns DefaultConstructors() from ConfigureStackConstraintConstructors.
Override that method only to add custom types via ExtendDefaults(), or
return nil to disable struct-tag constraints entirely.

# How Constraints Are Checked

Constraints use pre-insertion semantics: the destination stack is in its
pre-insertion state (the proposed components are NOT yet in the stack), and
the proposed parameter contains the components that would be added.
Constraint implementations must account for this — for example,
MaxNumComponents uses dest.NumComponents() + len(proposed) to predict the
post-move count.

Constraints are deterministic validation predicates, not event callbacks. The
engine may invoke them with a copied destination, proposed components, and
state. Custom constraints may capture immutable configuration such as a limit
or property name, but must resolve all runtime state through the arguments they
receive. Do not capture the game state, player state, destination stack, a
counter, or any other mutable runtime object.

A safe custom factory looks like this:

	func maxCombinedTokens(max int) boardgame.StackConstraint {
		return func(dest boardgame.ImmutableStack, proposed []boardgame.ImmutableComponentInstance, state boardgame.ImmutableState) error {
			game := state.ImmutableGameState().(*gameState)
			if dest.NumComponents()+game.Reserve.NumComponents()+len(proposed) > max {
				return errors.New("too many combined tokens")
			}
			return nil
		}
	}

Here max is immutable configuration and game is resolved from the supplied
state on every invocation. Capturing a gameState or Reserve value outside the
returned function is unsafe because a copied validation would still consult
the original object graph.

Constraints are checked automatically in two places:
  - During Legal(), for moves that declare source/destination via
    WithSourceProperty/WithDestinationProperty (no move code needed).
  - During Apply(), inside moveComponentImpl, as a safety net before
    the component is actually inserted.

For custom moves, use MayMoveTo or MayMoveToSlot in Legal() to check
constraints along with all other slot-independent or slot-specific
validation in a single call. See the ImmutableComponentInstance
documentation for details.

Automatic source/destination checking proposes the source's first component.
That is sufficient for a one-component Apply, but not for an Apply that calls
MoveAllTo: a later component can violate an order-dependent constraint. Such a
move should call source.MayMoveAllTo(destination) in Legal(). MoveAllTo repeats
the full validation transactionally in Apply, so an ignored or handled error
still cannot expose a partial transfer.

# Future Work

Source-side constraints (checked on the stack a component is being removed
from) are not yet supported. This would enable patterns like "don't allow
removal below N components."
*/
package constraints
