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
*/
package constraints
