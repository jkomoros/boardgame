# Evidence pack: per-game player-info renderers never mounted (pre-existing bug)

**Claim:** Since a87137230 (2026-02-07), no per-game player-info renderer
(pig's Round Score, memory's pair counts, blackjack/checkers equivalents)
has mounted in the live app.

**Mechanism:** boardgame-player-roster.ts forwarded its loaded flag with
`?renderer-loaded="${this.rendererLoaded}"` — an ATTRIBUTE binding — but
boardgame-player-roster-item's `rendererLoaded` property declares no
`attribute:` option, so Lit observes attribute `rendererloaded` (lowercased
camelCase, hyphen-free). The attribute set by the binding was never
observed; the flag stayed false; boardgame-render-player-info's mount
condition (`rendererLoaded`) never became true.

**Fix:** property binding (`.rendererLoaded=`) — the same style
roster-item itself uses to forward the flag downstream (line ~265).

**Proof:** red-first regression test
tests/animations/parity/player-info-mounts.spec.ts — fails against the
attribute binding (renderer never found in 20s), passes with the property
binding (mounts in <3s).

**Relevance to this branch:** Task 10 gates roster-subtree animations; this
bug meant the primary real-world producers of those animations never
existed at runtime. With the fix, per-game player-info renderers mount
again and their status-text/fading-text instances are live, gated
participants.
