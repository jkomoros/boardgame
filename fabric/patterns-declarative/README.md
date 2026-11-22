# Memory drawn by pattern JSX with boardgame elements

Stream D, step 3 of [`../PROPOSAL.md`](../PROPOSAL.md): the Memory board drawn
by a Common Fabric pattern's own JSX, through the real labs reconciler, with
the boardgame declarative elements of [`../declarative`](../declarative) in
the shell's document. No `cf-iframe`, no guest.

What was found is in [`FINDINGS.md`](FINDINGS.md). What ran is in
[`../status/declarative-pattern.md`](../status/declarative-pattern.md).

| Path | What |
|---|---|
| [`elements.ts`](elements.ts) | JSX types for the `bg-*` elements, merged into `JSX.IntrinsicElements` by the pattern itself |
| [`memory/main.tsx`](memory/main.tsx) | The pattern. Imports the table, `move` and `viewFor` of `../patterns/memory`; adds only a view |
| [`memory/per-stack.tsx`](memory/per-stack.tsx), [`memory/probe-release.tsx`](memory/probe-release.tsx) | Experiments on what of the released view a pattern may reshape |
| [`memory/state.ts`](memory/state.ts) | The two paths of Memory's released state those experiments read, as a type |
| [`reconciler/main.tsx`](reconciler/main.tsx) | Experiment on what the reconciler does to elements, with nothing confidential |
| [`memory/tools/`](memory/tools) | The browser scripts: a two-identity game, and one per experiment |
| [`labs-spike/`](labs-spike) | The labs change, as a patch |
| `memory/screenshots/`, `reconciler/screenshots/` | From those scripts, with each run's output as `run.log` |

## Running it

Node 24, Deno 2.9.8, and a labs checkout with the spike applied. All commands
from the repository root.

**1. Labs.** A worktree of labs with the patch. The table needs labs'
`viewerBoundRelease`, so the base is `70f5c390ec` (local branch
`worktree-agent-a53d3b2fae9b79d6b`) and every command below runs with
`EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`:

```bash
git -C /path/to/labs worktree add -b spike/boardgame-trusted-components-viewer-bound /path/to/labs-spike 70f5c390ec
```

```bash
git -C /path/to/labs-spike am "$PWD/fabric/patterns-declarative/labs-spike/0001-spike-shell-trusted-components.patch"
```

```bash
export LABS=/path/to/labs-spike EXPERIMENTAL_VIEWER_BOUND_RELEASE=true
```

**2. Build the elements** into one file, `fabric/declarative/dist/bg-elements.js`:

```bash
(cd fabric/declarative && npm ci && npm run build:bundle)
```

**3. Start labs' servers** with the shell told where the bundle is. Read
`$LABS/docs/development/LOCAL_DEV_SERVERS.md` first; 460 is the offset used
here.

```bash
(cd "$LABS" && CF_SPIKE_TRUSTED_COMPONENTS="$OLDPWD/fabric/declarative/dist/bg-elements.js" ./scripts/start-local-dev.sh --port-offset=460)
```

The shell bundles that file when it starts. After rebuilding the elements,
restart the servers.

**4. Key, space, deploy.**

```bash
export CF_LABS_ROOT="$LABS" CF_API_URL=http://localhost:8460
```

```bash
mkdir -p fabric/patterns-declarative/.cf && test -e fabric/patterns-declarative/.cf/owner.key || "$LABS/bin/cf" id new > fabric/patterns-declarative/.cf/owner.key
```

```bash
export CF_IDENTITY="$PWD/fabric/patterns-declarative/.cf/owner.key"
```

```bash
export CF_SPACE="$("$LABS/bin/cf" space create --label "Declarative Memory" --quiet | tail -1)"
```

```bash
"$LABS/bin/cf" piece new fabric/patterns-declarative/memory/main.tsx --root "$PWD" --slug memory --quiet
```

**5. Play it as two people.** A table is played once; deploy under a new
`--slug` for another run.

```bash
node fabric/patterns-declarative/memory/tools/play-two-users.mjs "$CF_API_URL/$CF_SPACE/memory"
```

The experiments are deployed and driven the same way: `memory/per-stack.tsx`
with `tools/probe-per-stack.mjs`, `memory/probe-release.tsx` with
`tools/probe-release.mjs`, `reconciler/main.tsx` with
`tools/probe-reconciler.mjs`.

**6. Stop the servers.**

```bash
(cd "$LABS" && ./scripts/stop-local-dev.sh --port-offset=460)
```
