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

To enable struct-tag constraints, your GameDelegate must return constructors
from ConfigureStackConstraintConstructors. Use DefaultConstructors() to get
all pre-built constraints, or ExtendDefaults() to add custom types.

# Future Work

Source-side constraints (checked on the stack a component is being removed
from) are not yet supported. This would enable patterns like "don't allow
removal below N components." The current constraints are all destination-side:
they validate the stack after a component is added.
*/
package constraints
