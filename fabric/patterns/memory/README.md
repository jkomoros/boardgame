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
| [`table.tsx`](table.tsx) | The sealed table: the policy, the stored state, the one writer `move`, the one released lift `seatView`. **The interim seam is here** |
| [`game.ts`](game.ts) | The engine's Memory (`server/static/src/engine/games/memory.ts`), compiled; the public facts. No rules are written in this directory |
| [`main.tsx`](main.tsx) | The piece: the `cf-iframe`, and `identify` |
| [`../toolkit/table.ts`](../toolkit/table.ts) | Seats and the engine's three stored values; `stepTable`, `viewOf`, `seatViewOf` |
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
- **One release.** `seatView` returns a `TableView`: who is seated, and each
  version the table still holds as the boardgame client's state bundle
  (stacks with `Indexes`, `IDs` and `IDsLastSeen`, the move that produced the
  version, timers, outcome, and legality for each seat).
- **Timers.** A mismatch starts a four-second timer: a deadline in the
  engine's stored state. A guest whose own clock runs out sends `tick`, and
  the engine hides the cards only if the table's clock has passed it.

**What is weaker than designed.** `table.tsx`, under "The policy, and the
interim seam", ships a release rule without the input witness: the view is
released over whatever is in the table's cell, not only over what `move`
wrote. The runtime keeps no witness for a record its writer reads and then
sets (`FINDINGS.md` F2), and the alternative, one JSON string, is not taken.
Separately, anyone who can read the space can read the sealed record today,
because labs delivers a space's documents to every reader (specs#95).

**Which labs.** Branch `fix/served-handler-write-identity` (draft PR
commonfabric/labs#10111): labs main with the fix for served handlers
(`FINDINGS.md` F1). The release as designed needs it. The rule as shipped has
no witness and so does not itself need it; it was not rerun on labs main.

## Tests, with no browser

Node 24 and a labs checkout with Deno 2.9.8. From the repository root:

```bash
node --test fabric/patterns/memory/parity.test.ts fabric/patterns/memory/table.test.ts
```

```bash
node fabric/patterns/tools/recording-fixture.mjs memory/scenario-mismatch fabric/patterns/memory/testdata/recording.json
```

```bash
CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/memory/parity.test.tsx --root "$PWD"
```

```bash
CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/memory/multi-user.test.tsx --root "$PWD"
```

- **Parity** (24 Node tests): all six recorded Go games of Memory, 172
  versions and 688 viewer outputs, match at every version through the
  functions `move` and the view lift call; again through a copy that
  reorders every object's keys, as a cell does; and one of them inside a
  lift under the pattern sandbox.
- **The table** : seating, every refusal in Go's words, card ids across a
  reveal, the timer against a moved clock, a whole game.
- **Multi-user** takes about three minutes and ends with `202 passed, 0
  failed`: a full two-player game through `move`, no view showing a hidden
  card's face, wrong-player, forged-proposer, unseated and illegal moves
  refused, and the two players' views agreeing.

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
(cd "$LABS" && ./scripts/start-local-dev.sh --port-offset=340)
```

**3. Mint a key, create a space, deploy.**

```bash
export CF_LABS_ROOT="$LABS" CF_API_URL=http://localhost:8340
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

On 2026-11-21 against labs `4815052651` (`fix/served-handler-write-identity`),
with the script, on the engine: a complete game between two identities, 16
reveals, three mismatches hidden by the timer, five pairs captured, the same
winner shown to both, up to ten animations running on each move. A move took
a median of 3.7 s to reach a guest; `FINDINGS.md` says why that is slower
than the hand-written rules were.

Rerun on the same day and labs commit after the guest was rebuilt on the
client packages and the lobby was made data-driven: two more whole games
between two fresh identities, on the ten-card board. In a space of its own,
16 reveals, a median of **3.0 s** from proposal to the mover's guest (2.2 to
5.3 s); in a space that also held two blackjack tables, 14 reveals and a
median of 4.3 s (3.0 to 8.3 s). The screenshots below are from the first of
those.

| Screenshot | Shows |
|---|---|
| `01-lobby-*` | Both seated, before the deal |
| `02-dealt-*` | Ten cards face down, for each player |
| `03-flipping-alice` | Just after the first click |
| `04-mismatch-timer-*` | Two different cards up and the timer bar running, for both |
| `05-capturing-*` | A match held, then leaving the board for the player's pile |
| `06-finished-*` | The end, for both |
