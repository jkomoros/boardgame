# Conformance corpus

Each `<game>/<name>.json.gz` here is a complete record of one game run on the
Go engine. A second implementation of the engine (the TypeScript one in
`fabric/PROPOSAL.md`) is correct for a run when, given the same game id, salt,
variant and outside input, it reproduces every version of the document.

For every state version, fix-up moves included, a document holds the
unsanitized state as stored, the move that produced it, and for every viewer
(admin, observer, each player) the game object a client receives, the
sanitized move record, and the legality of each offered move. The header
carries the chest, variant configuration and move table, so a document needs
nothing else.

- `schema.ts` describes the format (types only).
- `index.json` lists every document: game, versions, viewers, what it covers.
- The `normalization` section inside each document lists every field that is
  omitted or rewritten relative to the live wire, and why. The same list is in
  the package doc of `boardgame-util/lib/conformance`.

## Why here, and why gzipped

The corpus is a cross-language contract read by Go tests and by TypeScript,
and it spans games, so it sits at the repository root instead of under one Go
package's `testdata` or one game's `client/`.

A full per-viewer record is large but repetitive: the biggest document is
12.5 MB of JSON and 200 KB gzipped. Files are gzipped deterministically (no
timestamp or name in the header). To read one:

```sh
gunzip -c conformance/memory/scenario-mismatch.json.gz | less
```

or write uncompressed copies somewhere else:

```sh
./scripts/go-local run ./boardgame-util conformance export --plain --out /some/dir
```

In Node, `JSON.parse(zlib.gunzipSync(fs.readFileSync(path)).toString())`.

## Regenerating

The corpus is generated; do not edit it by hand.

```sh
./scripts/go-local run ./boardgame-util conformance export          # rewrite
./scripts/go-local run ./boardgame-util conformance export --check  # verify only
```

`./scripts/go-local test ./boardgame-util/lib/conformance/...` regenerates the
corpus in memory and fails if any checked-in document differs, so a change to
the engine or to an example game that alters behavior must regenerate it. The
test compares decompressed bytes, so a change in Go's compressor cannot fail
it.

Every golden history under `examples/{blackjack,memory,pig,tictactoe}/testdata/golden`
is exported automatically. Scripted scenarios are listed in
`boardgame-util/lib/conformance/corpus/corpus.go`; add one there to cover
something new.
