# Memory on Common Fabric

Memory for two to six people at one table, as a Common Fabric pattern. The
board is drawn by the boardgame client's own renderer, unmodified, in a
`cf-iframe`.

Background: [`../../PROPOSAL.md`](../../PROPOSAL.md) (stream B) and
[`../../RESEARCH.md`](../../RESEARCH.md) ("The shape: one space, labels do
the work"). What the runtime would and would not do is in
[`FINDINGS.md`](FINDINGS.md).

## How it is put together

| File | What |
|---|---|
| [`table.tsx`](table.tsx) | The sealed table: the policy, the stored state, the one writer `move`, the one released lift `viewFor` |
| [`game.ts`](game.ts) | The engine's Memory (`server/static/src/engine/games/memory/`, generated from the Go), compiled; the public facts. No rules are written in this directory |
| [`main.tsx`](main.tsx) | The piece: the `cf-iframe` |
| [`../toolkit/table.ts`](../toolkit/table.ts) | Seats and the engine's three stored values; `stepTable`, `viewForPrincipal` |
| [`../toolkit/view.ts`](../toolkit/view.ts), [`json.ts`](../toolkit/json.ts) | The wire types shared with the guest; JSON as a type a cell can hold |
| [`../toolkit/parity.ts`](../toolkit/parity.ts) | Replays a recorded Go game through a table and compares every version |
| [`../guest/`](../guest) | The built guest (`guest.js`, `guest.html`); not checked in |
| [`parity.test.ts`](parity.test.ts), [`parity.test.tsx`](parity.test.tsx) | Memory here is Memory in Go: under Node, and inside the pattern sandbox |
| [`table.test.ts`](table.test.ts) | The table with a clock the test moves |
| [`multi-user.test.tsx`](multi-user.test.tsx) | Two players and a watcher, each in their own runtime |
| [`tools/play-two-users.mjs`](tools/play-two-users.mjs) | Plays a game in two browsers as two identities |
| [`screenshots/`](screenshots) | From that script |

- **One space.** Nothing is hidden by where it is stored.
- **One sealed record** holds seats and what the engine asks a host to keep:
  the game record with its secret salt, the stored state of the versions the
  last input produced, and the whole move history. Its only confidentiality
  clause is this module's policy.
- **The rules are the engine's.** The shuffle, card ids, fix-up moves, the
  hide-cards timer and what each viewer may see are Go's, held to recorded
  Go games.
- **One writer.** `move` takes a `Proposal`: `sit`, `leave`, `start`, `play`,
  `tick` or `restart`. The proposer is `currentPrincipal()`. A refused
  proposal writes nothing.
- **One release.** `viewFor` asks the runtime who it is computed for and
  returns that principal's `TableView`: which seat is theirs, each seat's
  name, and each version the last input produced as the boardgame client's
  state bundle (stacks with `Indexes`, `IDs` and `IDsLastSeen`, the move that
  produced the version, timers, outcome), with what the viewer may do at the
  head. The runtime keeps one instance for each principal who looks, and the
  rule releases each to the principal it was computed for, when everything
  confidential it read was written by `move`.
- **Timers.** A mismatch starts a four-second timer: a deadline in the
  engine's stored state. A guest whose own clock runs out sends `tick`, and
  the engine hides the cards only if the table's clock has passed it.

**What is not held.** The rule decides what a runtime shows. Whether the
memory server hands the sealed record itself to a reader of the space is a
separate matter (specs#95), and a space's owner may release a clause by
intent; `fabric/status/viewer-bound-port.md` says what was observed. One
`TODO(cfc)` remains in `table.tsx`: the table starts as `null`, because a
record default is nobody's and would keep every view sealed.

**Which labs.** The viewer-bound release, on a local labs branch behind
`EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`
(`fabric/status/viewer-bound-port.md` names the commit). It also carries
the fixes for served handlers (`FINDINGS.md` F1) and for a record its writer
reads and sets (F2). Set the variable for every `cf` command below and for
the servers.

**The game is generated from the Go.** `game.ts` imports `memory()` from
`server/static/src/engine/games/memory/index.ts`, which
`boardgame-util transpile ./examples/memory` writes from `examples/memory`
(its exported definition and its translated functions). It replaced the
hand port `games/memory.ts`, which is gone
(`fabric/transpile/status/replacement.md`). The engine modules a table
imports are part of its policy's identity, so the switch gave this table a
new one: a table created before it is stranded on the old identity, and any
later change to the generated files' text (a regeneration that changes
them) does the same again (`fabric/TRANSPILE-PROPOSAL.md`, risk 6).

## Tests, with no browser

Node 24 and a labs checkout with Deno 2.9.8. From the repository root:

```bash
node --test fabric/patterns/memory/parity.test.ts fabric/patterns/memory/table.test.ts
```

```bash
node fabric/patterns/tools/recording-fixture.mjs memory/scenario-mismatch fabric/patterns/memory/testdata/recording.json
```

```bash
EXPERIMENTAL_VIEWER_BOUND_RELEASE=true CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/memory/parity.test.tsx --root "$PWD"
```

```bash
EXPERIMENTAL_VIEWER_BOUND_RELEASE=true CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/memory/multi-user.test.tsx --root "$PWD"
```

- **Parity** (24 Node tests): all six recorded Go games of Memory, 172
  versions and 688 viewer outputs, match at every version through the
  functions `move` and the view lift call; again through a copy that
  reorders every object's keys, as a cell does; and one of them inside a
  lift under the pattern sandbox.
- **The table** : seating, every refusal in Go's words, card ids across a
  reveal, the timer against a moved clock, a whole game.
- **Multi-user** takes about a minute and ends with `202 passed, 0 failed`:
  a full two-player game through `move`, each participant reading their own
  instance of the view, no view showing a hidden card's face, wrong-player,
  forged-proposer, unseated and illegal moves refused, and the two players'
  views agreeing.

## Running it

All commands from the repository root unless a `cd` says otherwise. `$LABS`
is a labs checkout (see "Which labs" above).

**1. Build the guest.** Once, and again after any change under `fabric/guest`
or to the client it bundles.

```bash
export LABS=/path/to/labs
```

```bash
cd fabric/guest && npm ci && tools/vendor-fabric-client.sh "$LABS" && npm run build:fabric && cd ../..
```

**2. Start labs' local servers.** Read
`$LABS/docs/development/LOCAL_DEV_SERVERS.md` first. With another checkout's
servers on the default ports, use an offset:

```bash
(cd "$LABS" && EXPERIMENTAL_VIEWER_BOUND_RELEASE=true ./scripts/start-local-dev.sh --port-offset=340)
```

**3. Mint a key, create a space, deploy.**

```bash
export CF_LABS_ROOT="$LABS" CF_API_URL=http://localhost:8340 EXPERIMENTAL_VIEWER_BOUND_RELEASE=true
```

```bash
mkdir -p fabric/patterns/.cf && test -e fabric/patterns/.cf/owner.key || "$LABS/bin/cf" id new > fabric/patterns/.cf/owner.key
```

```bash
export CF_IDENTITY="$PWD/fabric/patterns/.cf/owner.key"
```

```bash
export CF_SPACE="$("$LABS/bin/cf" space create --label "Memory table")"
```

```bash
"$LABS/bin/cf" piece new fabric/patterns/memory/main.tsx --root "$PWD" --slug memory
```

`fabric/patterns/.cf/` is ignored by Git. `--root` is the repository root
because the pattern imports from `server/static/src/engine`.

**4a. Play it as two people, by script.**

```bash
node fabric/patterns/memory/tools/play-two-users.mjs "$CF_API_URL/$CF_SPACE/memory"
```

It opens two isolated browser contexts, registers a new passphrase identity
in each on the shell's login screen, grants each `WRITE` with `cf acl set`,
seats Alice and Bob, deals, and plays to the end, waiting out the hide-cards
timer on each mismatch. It writes `screenshots/` and prints a line per move
and a summary. `HEADED=1` shows the browsers; `CFC_DEBUG=1` keeps the
runtime's reasons for any refusal. A table can be played once: deploy under
a new `--slug` for another run.

**4b. Or by hand.** Open the address `cf piece new` printed in two browser
profiles. In each, choose **Register**, **Use Passphrase**, **Generate
Passphrase**, **I've Saved It**. The page then says the slug was not found;
the console names the browser's DID. Grant it and reload:

```bash
"$LABS/bin/cf" acl set <browser DID> WRITE
```

**5. Stop the servers.**

```bash
(cd "$LABS" && ./scripts/stop-local-dev.sh --port-offset=340)
```

## What was run

On 2026-11-21 against labs `70f5c390ec` (the viewer-bound branch), with the
script: two complete games between two fresh identities on the ten-card
board, 16 and 18 reveals, seven mismatches hidden by the timer (4.0 to
5.2 s after the second card showed), the same winner shown to both, 282 and
326 animations in each guest and no watchdog firing. A reveal reached its
mover's guest in a median 1.4 s and 1.5 s (0.9 to 2.4 s). On the same
servers the pattern as it was before took 2.2 s and 2.4 s.
`fabric/status/viewer-bound-port.md` has every step.

The screenshots are from the first of those games.

| Screenshot | Shows |
|---|---|
| `01-lobby-*` | Both seated, before the deal |
| `02-dealt-*` | Ten cards face down, for each player |
| `03-flipping-alice` | Just after the first click |
| `04-mismatch-timer-*` | Two different cards up and the timer bar running, for both |
| `05-capturing-*` | A match held, then leaving the board for the player's pile |
| `06-finished-*` | The end, for both |
