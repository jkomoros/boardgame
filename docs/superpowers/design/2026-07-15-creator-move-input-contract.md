# Creator Move-Input Contract

## Motivation

Persisted fields on a Go move are not the same thing as inputs a renderer must
supply. `moves.CurrentPlayer.TargetPlayerIndex`, for example, is populated from
proposal context; requiring every renderer to send it is boilerplate and makes
impersonation mistakes easy. Conversely, a custom `DefaultsForState` method may
choose a convenient form default without making omission a stable API promise.

The contract therefore classifies responsibility explicitly. It never executes
`DefaultsForState` to guess ownership from one sample state.

## Game-Creator API

The common case needs no configuration:

- supported custom fields are `required`;
- `moves.CurrentPlayer`, `moves.AnyPlayer`, and `moves.AdminPlayer` contribute a
  `context-owned` `TargetPlayerIndex` automatically through `auto.MustConfig`;
- standard Go properties infer their normal creator codec;
- range-enum board indexes are creator-facing numbers even though the legacy
  form wire encodes them as strings.

A genuine server default is one explicit option:

```go
auto.MustConfig(
    new(moveChooseCard),
    moves.WithMoveInputDefault("CardIndex"),
)
```

The field becomes optional and overrideable in TypeScript. Merely implementing
`DefaultsForState` does not make it optional.

For unusual ownership or representation, use
`moves.WithMoveInputField(name, disposition, codec)`. Reconfiguring a field
owned by an embedded behavior requires the intentionally loud
`moves.WithMoveInputFieldOverride`. Ordinary configuration cannot silently
replace a behavior rule. Unknown fields, repeated declarations, multiple
codecs, unmatched overrides, and codec/property mismatches fail during
`auto.Config` with the field name.

External reusable behaviors may implement `boardgame.MoveInputFieldsProvider`.
One embedded provider is promoted automatically by Go's normal method-set
rules, including when the behavior type is unexported in another package. A
wrapper behavior that combines several providers implements one method that
returns its complete composed contract. This avoids visibility-sensitive
value reflection and keeps precedence explicit. Forgetting that composition
fails configuration and names the ambiguous provider types instead of silently
making their fields required. A nil embedded pointer whose provider panics fails
configuration with an actionable error. Framework current-player behaviors use
private marker interfaces for their automatic context-owned field.

## Dispositions

- `required`: the renderer must provide the field.
- `server-defaulted`: the renderer may omit it or explicitly override it.
- `context-owned`: the safe renderer API forbids it; proposal context owns it.
- `unsupported`: the safe renderer API forbids it and records the limitation.

Unclassified supported fields are required. Unclassified collection, stack,
board, timer, and other unsupported property types fail generation with an
actionable diagnostic instead of disappearing.

## Codecs and Layers

The codec describes the creator representation separately from the persisted Go
property and legacy form wire:

- `integer` and `player-index`: finite integer `number` in TypeScript;
- `boolean`: `boolean`, serialized internally as `"1"` or `"0"`;
- `enum`: an exact string-literal union when enum values are known;
- `string`: `string`.

Every generated module contains three models:

1. `MoveInputs`: native fields a renderer may supply; defaults are optional and
   context/unsupported fields are absent.
2. `ResolvedMoveInputs`: the complete supported model after server defaults and
   proposal context are applied.
3. `MoveWireInputs`: the internal form-encoded string representation; this is
   not a game-creator API.

The module also contains readonly runtime schema metadata and a canonical
SHA-256 fingerprint. Runtime validation rejects missing, extra, forbidden,
fractional, non-finite, wrongly typed, and invalid enum values before form
serialization.

Generated/bound renderers route both typed `proposeMove` calls and legacy
`propose-move`/`data-arg-*` controls through the same fingerprint and exact
validation boundary. The legacy adapter first converts dataset strings using
the generated codec metadata; it is a compatibility path, not a second
transport API.

## Freshness and Failure Behavior

The generator and server both use `boardgame.BuildMoveInputSchema`; the server
publishes its fingerprint as `MoveInputSchemaFingerprint` in `/info`. A safe
client fails closed with `BOARDGAME_STALE_MOVE_INPUT_SCHEMA` when the value is
missing or differs from its generated fingerprint.

`boardgame-util emit-move-args --check` performs extraction and comparison but
does not write. Normal generation builds every output in memory, prepares and
syncs every replacement beside its destination, and only then performs
same-filesystem atomic renames. Preparation failure leaves the previous
generation untouched; a crash during the rename phase is detected on the next
`--check` and by per-game fingerprint mismatch.

## Deferred Transport

Slice fields and non-integer numeric fields are deliberately unsupported. Their
empty/repeated/form and floating-point semantics have not been proven end to
end. Advanced custom codecs also require an explicit registration and
round-trip-test API before they are accepted; a string label alone is not
treated as executable serialization logic.
