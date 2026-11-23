# Declarative legality: component paths over static and dynamic values

Status: implemented on `claude/legal-dynamic-values-b9f79b`.
Scope: the "DynamicComponentValues have no path grammar equivalent" limit in
`legal/doc.go`. Siblings in flight at the same time handle seam scope,
`player.X` vs `proposer.X`, and honor-system `Reads`. This change leaves those
alone.

## 1. Problem

A legality check that reads a component's per-game *dynamic* values (a
species' `Population`, a checkers token's `Crowned`, a stack nested inside a
component's dynamic state) cannot be declared. It must go in `LegalCustom`.
That makes the move opaque to the ledger, keeps its bindings off the wire,
and means the check is never boot-validated.

The static half already has a partial answer. `ComponentPropEquals` and its
three siblings take a sealed `ComponentSelector` (`FirstOccupied`,
`OccupiedOrdinal`, `MoveIndex`), but they are a parallel family. They read
only static `Values()`, only string or enum fields, and only for equality.

## 2. Survey

Sources: the seven example games in this repository, and every game in the
GAMES repository (`darwin`, `metaltrader`, `murdermrmonroe`, `pass`,
`secretgroups`, `sequenceforge`, `terrainweave`, `valentine`). The survey
covered `LegalCustom` bodies and imperative `Legal()` overrides, looking for
any check that reads a component's values.

| Game / move | Check | Value source | Shape |
|---|---|---|---|
| darwin `moveIncreasePopulation` | `species.Population < MaxSpeciesPopulation` | dynamic int | compare to a literal |
| darwin `moveIncreaseBodySize` | `species.BodySize < MaxSpeciesBodySize` | dynamic int | compare to a literal |
| darwin `moveRemoveTrait` | `species.Traits.ComponentAt(move.TargetTraitCard) != nil` | dynamic **stack** | occupancy of a stack nested in a component |
| darwin `moveAddTrait` | `species.Traits.NumComponents() < maxTraits(numPlayers)` | dynamic stack | count, but the bound depends on player count (needs a computed property) |
| darwin `moveAddTrait` | no existing trait on the species has the proposed `Trait` | dynamic stack × static | loop over components (rule 3: stays custom) |
| darwin feeding/attack/nocturnal `Legal()` | `HasTrait`, `CanTakeFood`, `NocturnalUsed`, attack graph | dynamic | mostly computation; `NocturnalUsed == false` is a plain bool |
| valentine `MovePlayCard` | the other hand card is a Countess and this one is a King or Prince | static enum × 2 | two-component disjunction-of-conjunctions (stays custom) |
| valentine activate moves | `ActiveCard[first].Type == X` | static enum | already declarative via `ComponentPropEquals` |
| metaltrader `moveActivateMetalCard` | `ActiveCard[0].Type == Metal`, then pay out by the card's static counts | static enum, static ints used as counts | compare to a literal, plus a count *argument* read from a component |
| sequenceforge `MovePlayCard` | card value == pile count + 1 or wild | static int vs count | arithmetic (stays custom) |
| memory `moveRevealCard` match check | `VisibleCards[first].Type == VisibleCards[occupied 1].Type` | static enum × 2 | compare two components (already `ComponentPropsEqual`) |
| checkers `moveMoveToken` | destination is in `FreeNextSpaces`/`LegalCaptureSpaces` | dynamic `Crowned` + graph | graph walk (stays custom) |
| checkers `moveCrownToken` | (`DefaultComponent` base, opaque seam) | dynamic bool | n/a; that base is not on the seam |

The shapes, by how often they appear:

1. **A component value compared to a literal.** This is the most common
   shape, and it shows up for both static and dynamic values.
2. **Occupancy or count of a stack that lives inside a component's dynamic
   values.** darwin's `Traits`, `Food`, and `Fat` are all of this kind.
3. **Two components compared to each other.** These are all static today and
   are already served by `ComponentPropsEqual`.
4. **Loops, arithmetic, and graph walks.** Rule 3 keeps these out of the
   catalog, as it should.
5. **A component value compared to a player or game property.**
   `ComponentPropEqualsCurrentPlayer` is the existing named predicate for
   this.

Shapes 1 and 2 are the blocking ones. Both are *reads*. Neither needs a new
relation, only a way to name the value.

## 2a. Related history

I searched issues and discussions on jkomoros/boardgame and jkomoros/games.
boardgame has one discussion (#776, the welcome post), and games has
discussions disabled, so all the history below comes from issues.

| Issue | Relationship |
|---|---|
| boardgame#189 (capture `Legal` logic so the client can reject obviously illegal moves) | **Advances.** This is the root issue for declarative legality. The largest remaining class of opaque checks, dynamic component values, becomes declarative and ledger-visible. |
| boardgame#213 (client `proposeMove` should bail early) | **Advances**, through #189. Component-path verdicts now carry honest per-viewer evaluability, which is what a client early-bail needs. |
| boardgame#41 (implement dynamic component state), #595, #201 (Values/DynamicValues/computed are "all the same") | **Context.** #201's observation that these are all sub-states is the reason the path does *not* make the author name the storage split (§4 alternative 3). |
| boardgame#415 (expand DynamicComponentValues client-side) | **Prerequisite for a future TypeScript evaluator** of component paths. It is not closed: the server-side ledger is the only consumer today. |
| boardgame#736 / #741 (stack-constraint predicates; computed index helpers for boards) | **Adjacent.** Those are about *move-time* stack constraints and computed properties. Component paths are the read-side equivalent for legality, and they compose with #741-style computed properties when a check needs real computation (rule 3). |
| boardgame#791 (`only(propPath, value)` stack constraint) | **Adjacent.** It uses the `constraints` package's `component.`/`dynamic.` prefix grammar (`constraints/prop_path.go`). That grammar is deliberately not reused (§4 alternative 3). Unifying the two prop-path grammars is a possible follow-up. |
| boardgame#553 (cards that do different moves at different times), games#15 (cards should only light up if they're legal) | **Advances.** Per-card legality that depends on a card's current state is now expressible declaratively, and so it can reach the ledger. |
| boardgame#65, #624, #473, #640, #644, #295 | Checked and **not affected**. They cover fix-up logging, deal-move emptiness, parallel evaluation, progression performance, state-dependent repeats, and AI helpers. No part of this change touches them. |

This change does not fully close any issue. #189 and #213 stay open until a
client evaluator exists.

## 3. Decision: a component segment in the path grammar

A selector becomes a **path segment**, and the terminal property resolves
against the selected component:

```
KIND.Stack[selector].Prop
KIND.Board[move.Space][selector].Prop
```

`selector` is one of:

| Bracket | Meaning | Builder |
|---|---|---|
| `[move.Field]` | the slot whose index (int field) or enum key (enum field) is the move's `Field` | `legal.MoveIndex(stack, "move.Field")` |
| `[first]` | the first occupied slot | `legal.FirstOccupied(stack)` |
| `[occupied:N]` | the N-th occupied slot (zero-based, skipping holes) | `legal.OccupiedOrdinal(stack, N)` |

`Prop` resolves against the component's static `Values()` **or** its
`DynamicComponentValues`. Exactly one of the two must define it, which is
checked at boot (§5). The author does not say which one. The value is just a
property of the component.

`ComponentSelector` gains one method, `Prop(name) string`, which returns the
canonical path. The sealed selector stays the typed way to build a component
reference, and the path string is its wire spelling:

```go
species := legal.MoveIndex("players[move.TargetPlayerIndex].Species", "move.TargetSpecies")

legal.PropCompare(species.Prop("Population"), legal.OpLessThan, MaxSpeciesPopulation)
legal.ComponentPresentAt(species.Prop("Traits"), "move.TargetTraitCard")
legal.PropEquals(legal.MoveIndex("game.Spaces", "move.TokenIndexToMove").Prop("Crowned"), "false")
```

Every predicate that resolves through `ctx.ResolvePath` works over these paths
with no change. That covers `PropEquals`/`PropNotEquals`, `PropAtLeast`,
`PropCompare`, `StackCount`/`StackEmpty`/`StackNotEmpty`,
`ComponentPresentAt`/`ComponentAbsentAt`/`ComponentPresentAtKey`, and the
`MayMove`/`MaySwap` family when the terminal property is a stack. No
`Dynamic*` family is added, no catalog predicate is added, and
`LegalCatalogVersion` does not move. The catalog's rules of growth hold:
rule 1 says a relation over a path belongs in the catalog, and this change
adds paths, not relations.

### Why this reads naturally

The bracket is the existing board-index bracket, generalized. Before this
change, `game.Board[move.Space]` already meant "index this collection with a
move field". Now `game.Stack[move.Slot]` means the same thing one level
down, from a stack to a component, and the parser tells the two apart
structurally:

- one bracket with no suffix → a board index (unchanged), which is a stack;
- one bracket followed by `.Prop` → a component selector on a stack;
- two brackets followed by `.Prop` → a board index, then a component selector.

A component is not a `PropertyType`, so a selector without `.Prop` is a parse
error. There is never a value the path could "be" except the property.

## 4. Alternatives considered

1. **A parallel `Dynamic*` predicate family** (`DynamicPropEquals`,
   `DynamicPropAtLeast`, ...). Rejected. It doubles the relation surface for
   a difference in *where a value lives*, and it is exactly the growth that
   rule 1 says to avoid. Authors would also have to learn which family to
   use.
2. **Widening `ComponentPropEquals` and its siblings to read dynamic fields
   and more types.** Rejected as the primary mechanism. It still could not
   express a comparison, a count, or occupancy over a dynamic value without
   adding `ComponentPropAtLeast`, `ComponentStackCount`, and so on, which is
   the same parallel family under a different prefix.
3. **Explicit `dynamic.`/`component.` prefixes** (the precedent is
   `constraints/prop_path.go`), for example `game.Spaces[move.X].dynamic.Crowned`.
   Rejected. Static and dynamic values come from two Go structs, chosen per
   deck by the game, and the author does not think of them as two
   namespaces. Requiring a prefix would make authors learn the storage
   split. Name collisions are vanishingly rare, and they are caught at boot
   (§5), with an error that suggests renaming one field.
4. **Arbitrary path traversal** (`a.b.c[...]...`). Rejected for this round.
   One component hop covers every surveyed case, including darwin's nested
   `Traits` stack used as a terminal stack. A second hop
   (`...Traits[move.T].Trait`) is grammatically a small step, but it adds
   transitive sanitization reasoning (§6) that no surveyed move needs, so it
   is a parse error for now and recorded as future work.
5. **Reusing the existing `ComponentSelector` wire strings inside brackets**
   (`[move-index:move.X]`, `[first-occupied]`). Rejected for readability, and
   because `[move.X]` already exists as bracket syntax. The selector *types*
   are shared. Only their in-path spelling is new, and `Prop` produces it so
   authors never hand-spell it.

## 5. Boot validation

`validateLegalPath` and `validateLegalPathTypes` (core, one grammar for boot
and evaluation) handle the new segment as follows:

- The stack property exists and is `TypeStack`. When there is a board
  bracket, the board checks run first as before, and then the space is the
  stack.
- A `[move.Field]` selector names an existing move field of type `TypeInt`
  or `TypeEnum`.
- Every deck the stack can hold is collected: the game reader for `game.*`,
  and **every** player's reader for player-scoped kinds. For each deck,
  `Prop` is classified like this:
  - **static** when *every* component's `Values()` reader defines it with
    one `PropertyType`. If only some components define it, that is a boot
    error naming the component.
  - **dynamic** when the deck's `DynamicComponentValues` reader defines it.
  - Both static and dynamic is a boot error (ambiguous). Neither is a boot
    error that names the deck.
- Across decks, the type must agree.
- The resolved type then feeds the normal `RequiredReadTypes` and
  `AllowedReadTypes` contracts. For example, `PropCompare` over a dynamic
  `Traits` stack fails at boot because the path is `TypeStack`, not
  `TypeInt`.

Evaluation applies the same rule: static `Values()` first, then dynamic
values. Boot has already proved that exactly one of them defines `Prop` on
every component, so this order cannot change the outcome.

An **empty slot or an out-of-range selector** resolves to an error, which
yields a fail-closed `Unknown` verdict. A plan that wants a specific message
for "no component there" puts `ComponentPresentAt` (or `StackNotEmpty`)
first, which is what every surveyed move already does.

## 6. Sanitization, facets, and the #693 guard

Dynamic values have their own sanitization rule. Every dynamic property is
`PolicyHidden` unless the component is in a `PolicyVisible` stack, directly
or transitively. When it is, the deck's configured dynamic-value policy
applies. `LegalReadEvaluable` mirrors that rule exactly, and it fails closed:

1. **The stack hop must be fully visible.** The stack (or board) property's
   policy for this viewer must survive `FacetValues`, which means
   `PolicyVisible`. Picking a component, and seeing anything about it,
   needs its identity and values. This is judged with the same owner rules
   as any other path: the current player for `player.*`, and every player
   for `proposer.*` and `players[move.F].*`.
2. **Static terminal property.** A visible component's static values are
   visible, so the read is evaluable when (1) holds.
3. **Dynamic terminal property.** The read's own `Facet` must also survive
   the deck's dynamic-value policy for this viewer. For example, a
   `StackCount` over a dynamic `Traits` stack declares `FacetCount`, which
   survives `PolicyLen`. A `PropEquals` over a dynamic int declares
   `FacetValues`, which needs `PolicyVisible`.
4. The transitive case (a component visible only because it sits in a
   stack that is itself inside a visible component's dynamic values) is not
   reachable, because component paths have one hop. Rule (1) is sufficient
   and conservative.

A read that fails these checks is not evaluable. The ledger then strips the
bindings, which are the dynamic value itself, so `{value}` for
`PropEquals`/`PropCompare` never reaches a viewer who could not see it.
`ClientEvaluable` keeps meaning "the generic client knows this predicate's
semantics". The component segment is part of the shared path grammar that a
future TypeScript evaluator must implement, the same as the board bracket.

## 7. Field-dependent vs field-independent

A path is **field-dependent** when *any* segment reads the move:

- the `move.*` kind;
- `players[move.F].*`;
- a board bracket `[move.F]`. This was a pre-existing gap:
  `legalReadsIncludeMovePath` ignored board brackets, so a predicate whose
  only move dependency was a board index would have been memoized as
  field-independent. It is fixed here because the component selector shares
  the same classification;
- a component bracket `[move.F]`.

`[first]` and `[occupied:N]` read only state, so those paths are
field-independent and memoizable per state version. Dynamic values are
state, so version-keyed memoization stays sound.

## 8. Relationship to the existing component predicates

`ComponentPropEquals`/`ComponentPropNotEquals` are now the same check as
`PropEquals`/`PropNotEquals` over `selector.Prop(field)` for static enum
fields. The path form also works for dynamic fields and for int, bool, and
`PlayerIndex` types, but it does not support string fields, because
`PropEquals` has no string arm. They are kept unchanged, because their wire
names, template keys, and bindings (`{prop}`) are shipped surface.
`ComponentPropsEqual`/`ComponentPropsNotEqual` remain the only way to compare
two components, because the catalog has no path-vs-path relation (see open
questions).

## 9. Migrations

- **examples (this repo):** no example game's `LegalCustom` reads a dynamic
  value with a non-computational check. checkers' `Crowned` is read only
  inside the graph walk, and pig's die value is read only in `Apply`. The
  new grammar is exercised through the checkers conformance fixture
  (`game.Spaces[move.TokenIndexToMove].Crowned` and `.Color`).
- **GAMES** (branch `legal-dynamic-values`, paired worktree, not pushed):
  - darwin `moveIncreasePopulation`/`moveIncreaseBodySize` are fully
    declarative, using
    `PropCompare(targetSpecies.Prop("Population"), "<", MaxSpeciesPopulation)`.
  - darwin `moveRemoveTrait` is fully declarative, using
    `ComponentPresentAt(targetSpecies.Prop("Traits"), "move.TargetTraitCard")`.
  - metaltrader `moveActivateMetalCard` is fully declarative, using
    `PropEquals(activeCard.Prop("Type"), "Metal")` and four
    `MayMoveCountTo(bowl, player.X, activeCard.Prop("MetalX"))`. The card's
    static values used to force `LegalCustom` here, and they now feed a
    count argument directly.
  - darwin `moveAddTrait` stays partially custom. Its cap depends on player
    count, and its duplicate-trait check is a loop (rule 3).
  - darwin's feeding and attack `Legal()` overrides stay imperative. They
    are mostly computation (`CanTakeFood`, attack graph).
  - The darwin goldens were extended (max and below-max caps, a missing
    species, a present trait), and a metaltrader activate golden was added.
    Every case matches the legacy messages exactly. `go test ./...` passes
    across the GAMES module through the task-local `go.work`.

## 10. Open questions

1. **Path-vs-path relations.** Should the catalog gain one equality and
   comparison predicate over two paths (`PropsEqual(a, b)`,
   `PropsCompare(a, op, b)`)? That would retire
   `ComponentPropsEqual`/`ComponentPropEqualsCurrentPlayer` as special cases,
   and it fits rule 1. It is a catalog version bump, so it is deferred.
2. **Converging `ComponentPropEquals` onto the path form.** Should the
   builder emit `propEquals` once `PropEquals` grows a string arm and
   `{prop}` stops being a load-bearing binding?
3. **A second component hop.** Darwin's duplicate-trait check would still
   need a loop, so no surveyed move needs `...Traits[move.T].Trait` today.
4. ~~Count arguments read from components.~~ Resolved: `MayMoveCountTo`'s
   `countPath` accepts a component path as-is, as metaltrader shows.
