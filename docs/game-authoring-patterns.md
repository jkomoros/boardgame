# Composing game machinery

Start with ordinary components, state, and moves. Add a shared declaration where
it removes a repeated decision; custom Go and custom renderers remain available
at each layer. These examples show how the same rules reach authoring, controls,
and executable review.

| Need | Smallest useful layer | Working example |
| --- | --- | --- |
| Pick a card or player | `moves.WithChoices`, then optional native projected-choice adapters | Darwin food cards; Werewolf votes |
| Distinguish hidden from a legitimate zero | `viewGameProp` / `viewPlayerProp` over sanitizer availability | Werewolf role and vote status |
| Compare an active card's static type | `legal.ComponentPropEquals` | Valentine Guard |
| Compare two visible cards | `legal.ComponentPropsEqual` | Memory captures |
| Repeated piles | `Board` with optional enum keys and one indexed stack-path step | Sequence Forge |
| Words and covers at the same position | Rectangular board `stacks` / `componentViews` | Secret Groups |
| Place tiles in an expanding layout | Ordinary components with dynamic coordinates, game-local geometry, spatial renderer | Terrain Weave in the Games repository |
| Time a rule | Existing timer `Start` / `Cancel` inside `Apply` | Memory; Werewolf's timed variant |
| Review rules and visuals together | `scenario.Run`, golden history, viewer replay | Memory, Darwin, Werewolf, Secret Groups |
| Choose moves using only a player's information | `bots.Policy` and version-checked `bots.Play` | Tic-Tac-Toe random-empty policy |

## Declare choices once

Configure bounded choices on the authoritative move. The standard tray works
without renderer code. A custom hand can pass `projectedStackChoices` to a
component zone; a player picker can pass `projectedPlayerChoices` to a target
list. These adapters reuse the projected actions, including their inputs,
availability, and version. They do not recreate legality. See the
[native-choice tutorial](../TUTORIAL.md) and [scenario guide](scenario-replays.md).

For hidden information, use the [viewer knowledge helpers](viewer-knowledge.md).
A sanitized default is not evidence that the underlying value is known.

## Compare static component values

Selectors name a bounded component lookup, then a predicate expresses the rule:

```go
legal.ComponentPropEquals(
    legal.FirstOccupied("player.ActiveCard"), "Type", "Guard",
)
legal.ComponentPropsEqual(
    legal.FirstOccupied("game.VisibleCards"),
    legal.OccupiedOrdinal("game.VisibleCards", 1), "Type",
)
```

`FirstOccupied` and `OccupiedOrdinal` skip empty slots; `MoveIndex` uses the
move's actual slot index. Manager setup validates paths, decks, and comparable
static string/enum fields. These predicates currently evaluate on the server;
client controls receive authoritative previews. Dynamic values and branching
rules still use ordinary `LegalCustom` code.

## Collect piles without losing their meaning

A repeated collection can be a single state property:

```go
BuildPiles boardgame.Board `stack:"cards,12" board:"4" enum:"pile"`
```

The registered enum must have exactly four values. `SpaceAt` takes a positional
index; `SpaceAtKey` takes an enum key, including noncontiguous keys. Saved boards
record the enum and ordered keys so a changed key layout cannot silently
reinterpret old piles. Existing saves without key metadata remain readable.

Reusable moving moves accept paths such as `game.BuildPiles[move.TargetPile]`
or `player.DiscardPiles[move.TargetPile]`. An integer field selects a position;
an enum field selects its key. Custom moves can call `moves.ResolveStackPath`
with the same grammar. Sequence Forge's legacy decoder shows the separate
migration needed when replacing old state fields with a Board.

Presentation does not need a heterogeneous physical stack. Use rectangular
board `stacks` and matching `componentViews` to align independently sanitized
layers. Supply `labelFor` to describe all visible occupants at each position.
Secret Groups keeps Words, Key, and Revealed as separate stacks.

## Exercise the same engine

Write a [scenario](scenario-replays.md) when a rule needs a reproducible sequence,
then use its history for golden replay and its viewer snapshots for visual
review. Explicit timer steps reuse the engine's durable completion path. A replay
control records an intended move; advancing a recorded step shows the next real
engine output. It is not a second implementation of the game.

[Bots](observation-bots.md) return the same ordinary creator inputs. The first
pilot demonstrates detached observations and stale-decision rejection;
cooperative callbacks are not a secure execution sandbox.
