# Blackjack on Common Fabric

Blackjack for two to seven people at one table, as a Common Fabric pattern,
on the TypeScript engine's port of `examples/blackjack`. It is the first
table whose seats see different things: each player sees their own hole card
and nobody else's.

It plays in a browser: three identities and a watcher who owns the table's
space have played whole games through the deployed piece, each handed their
own hole card and nobody else's (see "In a browser").

| File | What |
|---|---|
| [`table.tsx`](table.tsx) | The sealed table: the policy, the stored state, the one writer `move`, the one released lift `viewFor` |
| [`game.ts`](game.ts) | The engine's blackjack (`server/static/src/engine/games/blackjack/`, generated from the Go), compiled; the public facts |
| [`main.tsx`](main.tsx) | The piece: the `cf-iframe` |
| [`parity.test.ts`](parity.test.ts) | Every recorded Go game of blackjack, replayed through the table; no view holds another seat's hole card |
| [`parity.test.tsx`](parity.test.tsx) | The opening of one recording, replayed inside a lift in the pattern sandbox |
| [`multi-user.test.tsx`](multi-user.test.tsx) | Three players and a watcher, each in their own runtime |
| [`attacks/display-fit.test.ts`](attacks/display-fit.test.ts) | What each principal is shown at the display boundary, and what the table's owner gets by going around the table. Runs inside a labs checkout |
| [`tools/play-three-users.mjs`](tools/play-three-users.mjs) | Plays a game in four browsers as four identities |

## How it differs from Memory

It does not, in shape. Both tables are one sealed record, one `move`
handler and one released lift, `viewFor`, which asks the runtime who it is
computed for and returns the engine's view for that principal's seat
(`viewForPrincipal` in `../toolkit/table.ts`). The two `table.tsx` files
differ in the game they name. What differs is the engine's answer: at
Memory every seat is shown the same board, and at blackjack a seat is shown
its own hole card. Blackjack has no timers, so it never sends `tick`.

## Who is shown what

The release needs labs' viewer-bound release, which is on a local labs
branch behind `EXPERIMENTAL_VIEWER_BOUND_RELEASE` and is not ruled
(`fabric/status/viewer-bound-port.md` says which commit, and what ran).

The rule in `table.tsx` releases what `viewFor` computed to the principal it
was computed for, when everything confidential it read was written by
`move`. Nothing is per seat: no key, no list, no copy of a view.

| A seat's hole card is kept from | Before | Now |
|---|---|---|
| An unmodified client | yes | yes |
| Another player, or a watcher, at the runtime's display boundary | **no** | yes |
| The table's owner, at the display boundary | **no** | yes |
| The owner binding `move` from another module, or running the module again over the live table | **no** | yes |
| Other code writing the table | yes | yes |
| A client that reads the space's documents without the runtime's display boundary | **no** | **not shown either way; see below** |

"Before" is the shape this replaced: a view a seat, each released to the
whole space.

Two things this does not hold, neither of them the pattern's to fix:

- **Delivery.** The rule decides what a runtime shows. Whether the memory
  server hands the sealed table itself to a reader of the space is a
  separate matter (specs#95). `fabric/status/viewer-bound-port.md` says
  what was observed on this labs branch.
- **A client is trusted for its own user.** With server execution off, each
  client computes its own instance from a replica of the sealed table.

One `TODO(cfc)` remains in `table.tsx`: the table starts as `null`, because
a record default is nobody's and would keep every view sealed.

**The game is generated from the Go.** `game.ts` imports `blackjack()` from
`server/static/src/engine/games/blackjack/index.ts`, which
`boardgame-util transpile ./examples/blackjack` writes from
`examples/blackjack`. It replaced the hand port `games/blackjack.ts`, which
is gone (`fabric/transpile/status/replacement.md`). The engine modules a
table imports are part of its policy's identity, so the switch gave this
table a new one: a table created before it is stranded on the old identity,
and any later change to the generated files' text does the same again
(`fabric/TRANSPILE-PROPOSAL.md`, risk 6).

## Tests

Node 24 and a labs checkout that has the viewer-bound release (Deno 2.9.8).
From the repository root:

```bash
node --test fabric/patterns/blackjack/parity.test.ts
```

```bash
node fabric/patterns/tools/recording-fixture.mjs blackjack/golden-714104226cb9a317 fabric/patterns/blackjack/testdata/recording.json 16
```

```bash
EXPERIMENTAL_VIEWER_BOUND_RELEASE=true CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/blackjack/parity.test.tsx --root "$PWD"
```

```bash
EXPERIMENTAL_VIEWER_BOUND_RELEASE=true CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/blackjack/multi-user.test.tsx --root "$PWD"
```

The last ends with `92 passed, 0 failed`. `EXPERIMENTAL_SERVER_EXECUTION`
set to `true` or `false` chooses whether a serving runtime runs the table.

The display-boundary test uses labs' own harness, so it is copied into a
labs checkout to run:

```bash
cp fabric/patterns/blackjack/attacks/display-fit.test.ts /path/to/labs/runtime/html/test/boardgame-blackjack-display-fit.test.ts
```

```bash
(cd /path/to/labs/runtime/html && BOARDGAME_ROOT="$OLDPWD" deno test -A test/boardgame-blackjack-display-fit.test.ts)
```

`SERVED=1` runs it with a serving runtime.

- **Parity.** All seven recordings (419 versions, 3,118 viewer outputs):
  state, move record, and every viewer's game object, sanitized move,
  legality and choices, at every version. They include a table that seats its
  players one at a time, seven players, and a deck that runs out.
- **Secrecy.** In every view of every version, for every viewer: each other
  seat's hole card is the hidden card, the draw stack shows no face, and a
  hole card a player's own view shows appears nowhere in anyone else's view
  of the same version. Checked under Node over whole games. In `cf test`
  each participant checks their own instance at every step; a participant
  cannot read another's.
- **The display boundary and the owner.** Each player is shown their own
  hole card; nobody is shown another principal's view, the table or its
  salt; an owner who binds `move` from another module makes only their own
  proposals and is shown no hole card; a second run of the module over the
  live table is shown to nobody; another module's handler cannot write the
  table.
- **Refusals.** A move out of turn, a forged proposer, a move by someone
  with no seat, and an engine-owned move proposed by a player.

## What the guest does

Agreed between this pattern and `fabric/guest`, and written up in
`fabric/guest/PROTOCOL.md` ("Inside a `cf-iframe`"):

1. **The bundle holds blackjack's renderer**, chosen by the view's
   `gameName`.
2. **The bridge reads one view**, which the runtime computed for whoever is
   looking. The view's `viewer` says which seat is theirs. The guest asks
   for nothing per seat and never learns anyone's principal.
3. **The lobby is built from `info`**: its title is `info.displayName` and
   it offers `info.setupOptions`, which are the engine's variants. Blackjack's
   one option is `maxrounds`, 1 to 10.
4. **`info.moveInputSchemaFingerprint` is passed in `init`**, so the renderer
   checks itself against the engine it is playing on.
5. **Nothing for timers or art.** Blackjack has no timer, and the engine's
   chest carries each card's suit and rank.

## In a browser

With labs' local servers running with `EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`
and the guest built (Memory's README, "Running it", steps 1 to 3, deploying
`fabric/patterns/blackjack/main.tsx`):

```bash
ROUNDS=2 node fabric/patterns/blackjack/tools/play-three-users.mjs "$CF_API_URL/$CF_SPACE/blackjack"
```

It registers four fresh identities in four browser contexts. Alice, Bob and
Carol are granted `WRITE` and sit; Dave is granted `OWNER` of the table's
space and only watches. It deals and plays to the end, hitting below 15 and
standing otherwise. Before every move, and after the deal, it reads from
each unmodified client both the state its renderer host was handed and the
cards on screen, and checks that every hole card still hidden is given to,
and shown face up to, its own seat, is given to nobody else, is shown face
down to everybody else, and appears nowhere else in what anybody else was
given. It exits non-zero if any check fails.

On 2026-11-21 against labs `70f5c390ec` (the viewer-bound branch): three
whole games on the final code, two rounds each, 10, 9 and 8 moves. The same
result on all four screens, 424 of 424 checks, 745 to 1,002 animations in
each guest and no watchdog firing. Screenshots are in `screenshots/`, from
the first: `02-dealt-dave` is the owner's screen with two hole cards face
down, and `02-dealt-bob` is Bob's, with his own face up.

**How long it takes.** The deal reached every screen in 1.2 s. A move
reached its mover's screen in a median 2.6 s (1.1 to 12.1 s over 27 moves),
and the time follows how many versions the move produces: about two seconds
for one, 2.6 s for three, and 9.6 s for the sixteen of a round's end, which
finishes one round and deals the next a card at a time. A table that held
only the newest version of each input would make every move 1.2 to 2.1 s
and move those cards together; that is measured and not shipped.
`fabric/status/viewer-bound-port.md` has the numbers for every step and
where the time goes.
