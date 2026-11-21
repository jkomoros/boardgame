# Fabric guest

One self-contained HTML file that plays a boardgame from JSON messages. It
mounts the real `boardgame-render-game` with a game's unmodified renderer,
inside `<iframe sandbox="allow-scripts">`, with no server, storage or
network.

It is built on the two client packages, `@boardgame/elements` and
`@boardgame/renderer-kit` (`server/static/packages/`, installed here as
`file:` dependencies), and implements the kit's `RendererHost` contract. It
imports nothing from the webapp in `server/static/src`.

Background: [`../PROPOSAL.md`](../PROPOSAL.md), stream B, step 2.

| File | What |
|---|---|
| [`PROTOCOL.md`](PROTOCOL.md), [`protocol.ts`](protocol.ts) | The message contract and its types |
| [`FINDINGS.md`](FINDINGS.md) | Every place the client assumed the webapp, the shim used, and the change that would remove it; bundle size; CSP |
| `src/guest-host.ts` | The renderer host: version queue, timers, move and preview transports |
| `src/version-queue.ts` | Overlapping windows in, each version once and in order out |
| `src/games.ts` | The renderers compiled into the bundle |
| `src/client.ts`, `src/types/boardgame-types.ts` | The two paths a game client imports the framework by (`../../src/client.js`), re-exporting the packages |
| `game-src/<game>` | Links to each bundled game's `client/` directory, the webapp's convention |
| `src/guest-main.ts` | Entry point of the standalone guest: `postMessage` in and out |
| `src/fabric-main.ts`, `src/fabric-bridge.ts` | Entry point of the guest inside a Common Fabric `cf-iframe`: cells and streams in place of `postMessage`, and a lobby before the deal |
| `src/guest-document.ts` | Theme and root element, shared by both entry points |
| `tools/vendor-fabric-client.sh` | Bundles Common Fabric's guest client out of a labs checkout into `vendor/` |
| `harness.html` | Embeds the built guest and plays a recording by posting messages |
| `tools/export-replay` | Go: records a complete Memory game, every version, every viewer |
| `tools/serve.mjs` | Static server for the harness, port 8719 |
| `tests/` | Playwright suite |

## Use

Node 24, as in CI (Node 20 also works). Run everything from this directory.

```bash
npm ci
```

```bash
npm run build
```

```bash
npm run serve
```

Then open <http://127.0.0.1:8719/fabric/guest/harness.html>. Set `GUEST_PORT`
to use another port.

```bash
npm test
```

`npm test` builds first. `npm run type-check` runs `tsc` over the guest and
the client source it imports. `npm run replay` regenerates
`replays/memory-full.json` (needs Go; it runs through `scripts/go-local`).

## The build output

`dist/guest.html` is the deliverable: an HTML shell with the whole bundle in
one inline classic script. A `cf-iframe` guest is a document, so one document
is the unit to hand over, and it works from `srcdoc`, from a `file:` URL, or
served. `dist/guest.js` is the same code as a file, for a host that would
rather reference a script than allow an inline one. `dist/bundle-report.json`
lists every module and its size.

## The Common Fabric build

```bash
tools/vendor-fabric-client.sh /path/to/labs
```

```bash
npm run build:fabric
```

This writes `../patterns/guest/guest.js` and `guest.html`, which a pattern
reads with `dataFile()` and hands to `cf-iframe` as `script` and `src`. It is
the same renderer host and the same bundled renderers as the standalone
guest. The transport differs: `src/fabric-bridge.ts` reads the pattern's
`view`, `info` and `me` cells, turns each new `view` into the `init` and
`state` messages of [`PROTOCOL.md`](PROTOCOL.md), and sends the host's
`propose-move` and `tick` to the pattern's `move` stream. Neither build
output nor `vendor/` is checked in. The pattern that uses it is
[`../patterns/memory`](../patterns/memory/README.md).

## Harness controls

- **Step** sends the next version. **Deliver window** sends several in one
  message. **Play all** steps on each `animations-done`.
- **Overlap** is how many already-sent versions each message repeats, the way
  a latest-value cell would. The guest ignores them.
- **Seat** re-sends the current window computed for another seat. The guest
  cuts to it.
- **Proposed moves** chooses how the harness answers a `propose-move`. A
  recording cannot apply an arbitrary move, so "accept" only steps the
  recording forward.
- Query parameters: `?replay=scenario` plays
  `examples/memory/client/scenario-replay.json`; `?srcdoc=1` embeds by
  `srcdoc`; `?csp=network` or `?csp=strict` puts the guest under a
  Content-Security-Policy; `?autostart=0` waits for the console.
- `window.harness` exposes the same operations to the console and the tests.

## Adding a game

Link its client directory under `game-src/`
(`ln -s ../../../examples/<game>/client game-src/<game>`), import its
renderer modules in `src/games.ts` and add its row. Nothing else is per-game:
the renderer is chosen by the `gameName` of `init`.
