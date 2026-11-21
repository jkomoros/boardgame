# Common Fabric patterns

Games from this repository as Common Fabric patterns. Today there is one,
[`memory/`](memory/README.md), and that README has the commands to test,
deploy and play it.

```
patterns/
  toolkit/   shared by every game: no Common Fabric imports
  guest/     the built boardgame guest; written by fabric/guest, not checked in
  memory/    one game
```

## What is shared

- **`toolkit/view.ts`.** The wire types: `TableView`, `ViewEntry`, `Proposal`,
  `TableInfo`. A view entry is the boardgame client's state bundle, so the
  guest needs nothing game-specific to show one.
- **`toolkit/table.ts`.** Everything about a table that is not the game:
  seats, the deal's one draw of randomness, the window of recent versions,
  the move history, timers and their deadlines. `stepTable(rules, table,
  proposal, context)` is the whole write side and `viewOf(rules, table)` the
  whole read side. Both are pure, and take the clock, the randomness and the
  proposer as arguments.
- **`GameRules<G, S>`**, in the same file. What a game supplies: `setUp`,
  `step(game, version, identity, moveName, args, seat)` returning the proposed
  move's version and one for each fix-up after it, `sanitize`, `viewerPart`,
  `currentPlayer` and `outcome`. The TypeScript engine's rules layer is meant
  to implement this interface from a game definition, at which point a
  game's `rules.ts` goes away.
- **The guest**, `fabric/guest`. One bundle holds every game's renderer. Its
  bridge (`src/fabric-bridge.ts`) expects five resources by name: `view`,
  `info`, `me`, `move` and `identify`.

## Adding a second game

1. **`<game>/rules.ts`.** Implement `GameRules`. Return it from a function,
   not a constant: the pattern sandbox admits only plain data at module
   scope. Everything `sanitize` and `viewerPart` return must be newly built;
   an object passed through from the game's state is withheld from the guest
   along with the rest of the view (Memory's `FINDINGS.md`, F4).
2. **`<game>/table.tsx`.** Copy Memory's and change the types and the rules
   it calls. It is a short module of wiring that cannot be shared: a
   policy's rule names its lift and its handler by the module they are
   defined in, so each game's policy, `move` and `seatView` live in that
   game's own module. Keep the interim seam as it is in Memory's until the
   findings it cites are closed, and change it in every game together.
3. **`<game>/main.tsx`.** Copy Memory's; only the names and `info` change.
4. **The guest.** Add the game's two renderer imports and one row to
   `fabric/guest/src/games.ts`, then rebuild (`npm run build:fabric`).
5. **Tests.** A `rules.test.ts` for the rules with a moved clock, and a
   `multi-user.test.tsx` through the pattern.

## What a second game will find missing

- **Per-seat secrets.** `viewOf` releases one view to every reader and puts
  only legality under each seat. Memory has nothing a seat alone may see. A
  game with hands needs one released view a seat, and a way to release it to
  that seat's occupant (RESEARCH.md, "Dealing to a player"), which is not
  built.
- **The lobby** in the guest's bridge says "Memory" and offers Memory's three
  board sizes. A second game needs its setup options to come from `info`.
- **Size.** A table and its view are each rewritten whole on every move.
  Memory's are about 30 KB and a move takes over two seconds to reach a
  guest. A game with more state needs the view split by version first.
