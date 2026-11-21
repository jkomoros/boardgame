# Common Fabric patterns

Games from this repository as Common Fabric patterns, on the TypeScript
engine (`server/static/src/engine`). There are two:

- [`memory/`](memory/README.md) plays in a browser, for two to six people.
  Its README has the commands to test, deploy and play it.
- [`blackjack/`](blackjack/README.md) has hidden hands, so each seat has its
  own view. It plays in a browser too, slowly: its README has the numbers.

```
patterns/
  toolkit/    shared by every game: no Common Fabric imports
  guest/      the built boardgame guest; written by fabric/guest, not checked in
  tools/      a script that writes a recorded Go game as a test fixture
  memory/
  blackjack/
```

## What is shared

- **The engine.** A game's rules are a `GameSpec` under
  `server/static/src/engine/games/`, the same one the engine's conformance
  tests hold to Go. A pattern compiles it (`compileGame`) and writes no rules
  of its own.
- **`toolkit/table.ts`.** Everything about a table that is not the game: the
  seats, the deal's one draw of randomness, and the three values the engine
  asks a host to keep (game record, move history, stored state), for the
  versions of the last input. `stepTable(game, table, proposal, context)` is
  the whole write side. `viewOf(game, table)` is one view for everyone, for a
  game whose seats all see the same; `seatViewOf(game, table, seat)` is one
  seat's own. All are pure, and take the clock, the randomness and the
  proposer as arguments.
- **`toolkit/view.ts`, `toolkit/json.ts`.** The wire types (`TableView`,
  `ViewEntry`, `Proposal`, `TableInfo`): a view entry is the boardgame
  client's state bundle. And JSON as a type a cell can hold, with the copy
  every released value has to be.
- **`toolkit/parity.ts`.** Replays a recorded Go game from `conformance/`
  through a table and compares the state, the move record and every
  viewer's output at every version. It runs under Node and inside a pattern.
- **The guest**, `fabric/guest`. One bundle holds every game's renderer.

## Adding a game

1. **The rules.** A `GameSpec` in `server/static/src/engine/games/`, passing
   the engine's own conformance test against recorded Go games. The engine's
   README says how, and what a pattern's sandbox requires of the file.
2. **`<game>/game.ts`.** Compile it, and name the seat colours.
3. **`<game>/table.tsx`.** Copy Memory's (every seat sees the same) or
   blackjack's (a view a seat). It is a short module of wiring that cannot
   be shared: a policy's rule names its lift and its handler by the module
   they are defined in. Keep the interim seams as they are in the one you
   copy, and change them in every game together.
4. **`<game>/main.tsx`.** Copy; only the names change.
5. **Tests.** `parity.test.ts` (copy; it finds the game's recordings by
   name), `parity.test.tsx` for the sandbox, and a `multi-user.test.tsx`.
6. **The guest.** Add the game's renderer to `fabric/guest/src/games.ts` and
   rebuild.

## What the next game will find missing

- **Release to a seat.** A seat's view cannot yet be released to that seat's
  occupants alone (blackjack's README, "What is not held today").
- **Speed.** A move takes seconds, in proportion to the size of the state
  (Memory's `FINDINGS.md`, "Latency"). Blackjack's deal with two players is
  ten versions in one input, a 181 KB view.
- **Seating through the engine.** A table seats people itself and creates
  the game with every seat filled. A game whose rules react to a player
  arriving mid-game is replayed correctly in the parity tests, but a
  `Proposal` cannot yet cause it.
