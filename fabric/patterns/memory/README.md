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
| [`rules.ts`](rules.ts) | Memory's rules, by hand, behind the `GameRules` interface. The engine's rules layer replaces this file |
| [`main.tsx`](main.tsx) | The piece: the `cf-iframe`, and `identify` |
| [`../toolkit/table.ts`](../toolkit/table.ts) | Seats, the window of versions, history, timers; `stepTable` and `viewOf` |
| [`../toolkit/view.ts`](../toolkit/view.ts) | The wire types shared with the guest |
| [`../guest/`](../guest) | The built guest (`guest.js`, `guest.html`); not checked in |
| [`multi-user.test.tsx`](multi-user.test.tsx) | Two players and a watcher, each in their own runtime |
| [`rules.test.ts`](rules.test.ts) | The rules with a clock the test moves |
| [`tools/play-two-users.mjs`](tools/play-two-users.mjs) | Plays a game in two browsers as two identities |
| [`screenshots/`](screenshots) | From that script |

- **One space.** Nothing is hidden by where it is stored.
- **One sealed record** holds seats, the layout, face-up slots, captures, the
  turn, the version, running timers, the last six versions and the whole
  move history. Its only confidentiality clause is this module's policy.
- **One writer.** `move` takes a `Proposal`: `sit`, `leave`, `start`, `play`,
  `tick` or `restart`. The proposer is `currentPrincipal()`. A refused
  proposal writes nothing.
- **One release.** `seatView` returns a `TableView`: who is seated, and each
  version the table still holds as the boardgame client's state bundle
  (stacks with `Indexes`, `IDs` and `IDsLastSeen`, the move that produced the
  version, timers, outcome, and legality for each seat).
- **Timers.** A mismatch starts a four-second timer: a deadline in the sealed
  record. A guest whose own clock runs out sends `tick`, and `move` hides the
  cards only if its clock agrees.

**What is weaker than designed.** `table.tsx`, under "The policy, and the
interim seam", ships a release rule without the input witness: the view is
released over whatever is in the table's cell, not only over what `move`
wrote. The runtime keeps no witness for a record its writer reads and then
sets (`FINDINGS.md` F2), and the alternative, one JSON string, is not taken.
Separately, anyone who can read the space can read the sealed record today,
because labs delivers a space's documents to every reader (specs#95).

**Which labs.** The release as designed needs labs' fix for served handlers
(`5b1c700b4a`, on branch `worktree-agent-ad697e53f88bab905`; `FINDINGS.md`
F1), and everything recorded here was run on that branch. The rule as shipped
has no witness and so does not itself need the fix; it was not rerun on labs
main.

## Tests, with no browser

Node 24 and a labs checkout with Deno 2.9.8. From the repository root:

```bash
node --test fabric/patterns/memory/rules.test.ts
```

```bash
CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/memory/multi-user.test.tsx --root "$PWD"
```

The second takes a little over a minute and prints many `get() took >50ms`
warnings; it ends with `202 passed, 0 failed`. It covers a full two-player
game through `move`, that no view shows a hidden card's face, the refusals
(wrong player, forged proposer, no seat, illegal moves), and that the two
players' views agree.

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

On 2026-11-21 against labs `6e12b01bf7` (the fix branch), with the script: a
complete game between two identities, 14 reveals, two mismatches hidden by
the timer, five pairs captured, the same winner shown to both. The client's
animation counters read 29 cycles and 238 animations in each guest, with no
watchdog firing. Numbers and what they mean are in `FINDINGS.md`.

| Screenshot | Shows |
|---|---|
| `01-lobby-*` | Both seated, before the deal |
| `02-dealt-*` | Ten cards face down, for each player |
| `03-flipping-alice` | Just after the first click |
| `04-mismatch-timer-*` | Two different cards up and the timer bar running, for both |
| `05-capturing-*` | A match held, then leaving the board for the player's pile |
| `06-finished-*` | The end, for both |
