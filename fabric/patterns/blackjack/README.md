# Blackjack on Common Fabric

Blackjack for two to seven people at one table, as a Common Fabric pattern,
on the TypeScript engine's port of `examples/blackjack`. It is the first
table whose seats see different things: each player sees their own hole card
and nobody else's.

It runs to the end of a game in `cf test`. It does not yet play in a
browser: the guest has no blackjack renderer and reads one shared view (see
"What the guest needs").

| File | What |
|---|---|
| [`table.tsx`](table.tsx) | The sealed table: the policy, the stored state, the one writer `move`, the one released lift `seatView`. **The two interim seams are here** |
| [`game.ts`](game.ts) | The engine's blackjack, compiled; the public facts |
| [`main.tsx`](main.tsx) | The piece: the `cf-iframe`, and `identify` |
| [`parity.test.ts`](parity.test.ts) | Every recorded Go game of blackjack, replayed through the table; no view holds another seat's hole card |
| [`parity.test.tsx`](parity.test.tsx) | The opening of one recording, replayed inside a lift in the pattern sandbox |
| [`multi-user.test.tsx`](multi-user.test.tsx) | Three players and a watcher, each in their own runtime |

## How it differs from Memory

- **One view a seat.** `seatView({ table, seat })` is the engine's
  sanitization for that seat (`seatViewOf` in `../toolkit/table.ts`). The
  pattern asks for the watcher's view and for each of seven seats, as
  separate values, so that each can be released to a different audience.
- **Everything else is the same**: one space, one sealed record, one `move`
  handler, the same `Proposal`s (`sit`, `leave`, `start`, `play`, `restart`;
  blackjack has no timers), and the same toolkit.

## What is not held today

`table.tsx`, under "The policy, and the interim seams", says both in full.

| Seam | Designed | Shipped | Not held |
|---|---|---|---|
| The input witness | A view is released only over a table `move` wrote | No witness | The release is of whatever is in the cell. Same as Memory (its FINDINGS F2) |
| The seat's audience | Seat i's view is released to seat i's occupant list (`fabric/RESEARCH.md`, "Dealing to a player"; specs#59) | Every seat's view is released to the table's space, like the watcher's | **Every reader of the space can read every seat's view, hole cards included.** A guest shows a seat only its own because it asks only for its own. That holds against an unmodified client and nothing else |

The designed release for a seat is written out in `table.tsx`. A labs spike
ran it on a port of the list-release draft (labs#8608): a sealed key a seat
carrying `ifc.members`, occupant lists written by a second handler, players
listing themselves, one rule adding the key's `Members`. Who a seat's view
is kept from, in each version:

| Kept from | Shipped | Designed, on the port |
|---|---|---|
| An unmodified client | yes | yes |
| A modified client at the table | **no** | yes |
| A watcher with access to the space | **no** | yes |
| The table's owner | **no** | **no** |

The owner's row is the spike's finding: the owner can bind the seat list's
writer from another module to a live seat's list and read that hand. It
needs a ruling that binds a declared writer to the run that owns the
document. This pattern does not try to block it.

And, as for Memory, anyone who can read the space can read the sealed table
itself, because labs delivers a space's documents to every reader
(specs#95).

## Tests

Node 24 and a labs checkout on branch `fix/served-handler-write-identity`
(Deno 2.9.8). From the repository root:

```bash
node --test fabric/patterns/blackjack/parity.test.ts
```

```bash
node fabric/patterns/tools/recording-fixture.mjs blackjack/golden-714104226cb9a317 fabric/patterns/blackjack/testdata/recording.json 16
```

```bash
CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/blackjack/parity.test.tsx --root "$PWD"
```

```bash
CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/blackjack/multi-user.test.tsx --root "$PWD"
```

The last takes about four minutes and ends with `92 passed, 0 failed`.

- **Parity.** All seven recordings (419 versions, 3,118 viewer outputs):
  state, move record, and every viewer's game object, sanitized move,
  legality and choices, at every version. They include a table that seats its
  players one at a time, seven players, and a deck that runs out.
- **Secrecy.** In every view of every version, for every viewer: each other
  seat's hole card is the hidden card, the draw stack shows no face, and a
  hole card a player's own view shows appears nowhere in anyone else's view
  of the same version. Checked under Node over whole games and in `cf test`
  at every step by every participant.
- **Refusals.** A move out of turn, a forged proposer, a move by someone
  with no seat, and an engine-owned move proposed by a player.

## What the guest needs

To play this in a browser, `fabric/guest` needs:

1. **A blackjack renderer in the bundle**: `examples/blackjack/client`,
   added to `src/games.ts`. Today the guest says `this guest has no renderer
   for "blackjack"`.
2. **A bridge that reads one view a seat.** The pattern hands the guest
   `view` (the watcher's) and `seat0` to `seat6`. The bridge should read
   `view` for the seats, find the viewer's seat from `me`, and then take its
   entries from `seat<N>` once seated. Each seat view is a `TableView` whose
   entries' `viewers` hold only that seat's key. Memory hands a guest `view`
   alone, with every seat's legality under `viewers`; a bridge that falls
   back to `view` when there is no `seat<N>` serves both.
3. **A lobby that is not Memory's.** The bridge's lobby is titled "Memory"
   and offers Memory's three board sizes. Blackjack's one option is
   `maxrounds`, 1 to 10. The options could come from `info`.
4. **Nothing for timers or art.** Blackjack has no timer, and the engine's
   chest carries each card's suit and rank.

What was seen in a browser (labs `4815052651`, local servers): the piece
deploys, two CLI identities sat and dealt, and a browser identity's guest
was handed the watcher's view of the dealt game (181 KB, ten versions),
showed "Watching" and whose turn it was, and reported the missing renderer.
That a seat's own view is handed over was not looked at; it is released by
the same rule.
