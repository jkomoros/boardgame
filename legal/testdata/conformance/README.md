# Conformance corpus

Each file in this directory is a table of `(spec, fixture, proposer, expect)`
cases for exactly one catalog predicate. This format is dumb JSON on purpose:
it IS the future Go<->TypeScript conformance contract (design spec §6, §9) —
`legal/conformance_test.go` checks it against the Go catalog now, and a
future TS evaluator's test suite is expected to consume it verbatim.
Divergence between the two evaluators on any case here is a bug on
whichever side disagrees.

## File shape

```jsonc
{
  "predicate": "propAtLeast",
  "cases": [
    {
      "spec": {"name": "propAtLeast", "args": ["player.CardsLeftToReveal", "1"]},
      "fixture": "memoryDefault",
      "proposer": 0,
      "expect": "pass"
    }
  ]
}
```

- `predicate` must equal every case's `spec.name` in the file (checked by
  the loader).
- `spec` is a `legal.Spec` (`{name, args, sub, message}`) — the same JSON
  shape a predicate's `Spec` builder produces.
- `fixture` names a fixture built in `legal/conformance_test.go`'s
  `legalFixtureBuilders` map. See that file's doc comment for why the
  fixtures are two existing example games (`examples/memory`,
  `examples/checkers`) rather than a hand-rolled one.
- `proposer` is a player index passed through to `Context.Proposer`.
- `expect` is one of `"pass"`, `"fail"`, or `"unknown"`.

## Adding a predicate's corpus

1. Add or reuse a named fixture in `legalFixtureBuilders`
   (`legal/conformance_test.go`) that puts the state (and, if the predicate
   reads `move.*`, a `Move`) into the arrangement your case needs.
2. Add `<predicateName>.json` here with at least 3 cases: at minimum one
   `pass`, one `fail`, and one `unknown` (typically a nil-`Move` fixture for
   a predicate that declares a `move.*` read).
3. `go test ./legal/...` picks the file up automatically —
   `TestConformanceCorpus` globs every `*.json` in this directory.
