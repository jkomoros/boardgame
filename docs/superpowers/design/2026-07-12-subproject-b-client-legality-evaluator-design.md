# Sub-project B — Client Legality UI (design, critique-hardened v2)

> **Status:** design, hardened by a 5-lens adversarial critique (correctness,
> scope, TS-fit, conformance, completeness), each grounded in real code.
> Spec-first per the follow-ups roadmap. Not implemented. Branch
> `declarative-legality-design`. Supersedes the v1 draft; §"Corrections"
> records what the critique overturned.

## 0. TL;DR

The genuinely valuable, safe, *on-the-client* work is **authority-driven
move-graying + an explanation renderer** — and most of its plumbing already
exists. A **full client-side re-evaluator** (the original "Layer 3") turns
out to be the *wrong tool* for this codebase: in the real games every
field-dependent move puts its per-choice legality in `LegalCustom` (not the
catalog), the client can't resolve `player.X`/current-player or `any`
compositor children, a blessed imperative `Legal()` residue makes optimistic
"looks good" hints unsafe, and the conformance harness that would keep it
honest is large and easy to get false-green. If live *per-choice* preview is
wanted, a **debounced server-preview endpoint** previews the *authoritative*
legality (including `LegalCustom`) at a fraction of the cost — but it is a
round-trip, not "on the client." That tension is the one real decision (§8).

## 1. The load-bearing principle: authority vs. advisory (unchanged, verified)

- **Authoritative:** `LegalForPlayer`, `LegalForPlayerError`, `LegalForAnyone`
  — from the real `move.Legal()` calls `ProposeMove` gates on. Confirmed by
  the critique against `legalFormFromLedger` (main.go:1774-1785) and the F1
  regression tests (`legal_ledger_override_test.go`).
- **Advisory:** the `Preconditions` ledger — the declarative *plan's* view,
  which can disagree with the authority in **both** directions (F1).
- **Rule 0:** never fold ledger entries into the enable/disable decision.

But Rule 0 needs a correction the v1 draft got wrong (§Corrections C1): the
authority `LegalForPlayer` is computed against **`DefaultsForState`-bound**
move fields, not the player's live edits. So it is only a *stable* verdict
for **fieldless** moves or field-independent failures.

**Rule 0 (corrected).**
- **Visibility:** show a move iff `LegalForAnyone` (or absent-ledger opaque
  default). Unchanged.
- **Submit hard-disable** iff `LegalForPlayer===false` **AND** the ledger has
  **no `provisional` (field-dependent) entry** — i.e. no field choice could
  flip it. (Opaque moves: no ledger ⇒ treat as no provisional ⇒ authority is
  final.)
- **Otherwise keep submit reachable** (`LegalForPlayer===false` with a
  provisional entry ⇒ "your current choice isn't legal, pick another"), since
  the server verdict was only about the *default* field values.

## 2. What already exists in the client (verified — don't rebuild)

- `types/api.d.ts`: `PreconditionEntry`/`PreconditionMessage`/`MoveForm`
  legality fields, omitempty modeled correctly.
- `selectors.ts` `selectMoveLegality`: memoized map incl. `preconditions` —
  **consumed by nothing**.
- `boardgame-render-game.ts` `_deriveLegality`: the map game renderers
  *actually* receive — and it **drops `preconditions`** (a lossy fork of the
  selector). `isMoveCurrentlyLegal()`/`isMovePossible()` helpers exist and are
  already pushed to renderers.
- **The merged template table already ships**: `Chest.LegalTemplates` on
  /info (defaults + delegate overrides), reachable via `selectGameChest`.
  (§Corrections C3.)

So "Layer 1" is largely built; the real v1 gap is *adoption* + the generic
admin move-form + threading `preconditions` to the renderer.

## 3. Scope, reframed into three tiers

### v1 — Authority graying + inline reason (BUILD NOW; on the client; low risk)
1. Apply corrected Rule 0 to the **generic/admin move-form button** and the
   example game renderers (adopt `isMoveCurrentlyLegal`/`isMovePossible`).
2. Render **`LegalForPlayerError` inline** on a disabled/kept-reachable
   control (it's already in `MoveLegalityInfo.error`). This is the
   always-safe reason for fieldless moves; for field-bearing moves show it as
   "current choice isn't legal."
3. **Bind only to settled live-head state** — gate the actionable legality UI
   on no-pending-animation-bundles AND viewing the current version, so
   animation/version bundles (which carry legality-free intermediate forms)
   don't flicker or mis-gate (§Corrections C4 / critique CRITICAL-completeness).
4. **Observer / simultaneous-phase treatment**: when `ViewingAsPlayer` is
   Observer (incl. simultaneous phases, where the server coerces the viewer to
   Observer), render moves view-only with one "you're observing / not seated"
   affordance — never per-move blank-reason disabled buttons (observers get no
   `LegalForPlayerError`).
5. **a11y:** prefer `aria-disabled` + inline visible/announced reason over
   native `disabled` (which drops the reason from tab order). Fold into the D2
   UI decision.
6. Cheap plumbing: (a) capture `LegalCatalogVersion` from /info into the store
   (version slice; it's absent from later bundles); (b) type
   `GameChest.LegalTemplates`; (c) make `_deriveLegality` reuse
   `selectMoveLegality` so `preconditions` reach the renderer (needed by v2).

### v2 — Per-predicate explanation renderer (MODERATE value/cost; shared with W6)
A TS `renderLegalMessage(template, bindings, table)` mirroring Go's
`RenderLegalMessage` **byte-for-byte** (table lookup with raw-key fallback;
`{name}` where `name`∈`[A-Za-z0-9_]+`; **missing binding → bare placeholder
name**, never blank/throw), fed by `Chest.LegalTemplates`. Used to render the
first non-`pass` ledger entry as *supplementary* detail — **never the primary
reason** (an `evaluable:false` fail has its bindings stripped by the #693
guard → bare-name garbage; and an all-`pass` ledger with `LegalForPlayer:false`
(F1) has no failing entry to show). Primary reason stays `LegalForPlayerError`.
- **This same renderer is Workstream 6's** post-submit `{template,bindings}`
  rejection renderer (roadmap: "W6 shares its client renderer"). Design them
  together so pre- and post-submit reasons are consistent. (i18n is out of
  scope — the table is single-locale English today.)
- Reading order: the ledger array is already in server **evaluation (bucket)
  order**; "first entry whose `verdict!=='pass'`" reproduces the server's
  latched reason with **no re-parsing** (§Corrections C2).

### v3 — Live per-choice preview (the "architectural" question; DECIDE — §8)
Three candidate mechanisms, in descending recommendation:
- **(a) Server-preview endpoint** (best ROI, *authoritative*): a debounced
  `POST …/movePreview` that binds the user's current form args
  (`getMoveFromForm` already exists server-side) and returns
  `move.Legal(state, proposer)` — the *same* authority as submit, **including
  the `LegalCustom` residue** the client can't see. ~50 LOC server + a
  debounced client call. **But it is a round-trip, not "on the client."**
- **(b) Narrow client target-legality** (genuinely on the client; safe where
  it applies): for *click-to-propose* renderers, evaluate per-**candidate
  target** legality *before* the click for the subset of moves whose
  target-legality is fully catalog-expressible and field-dependent-only (e.g.
  "which board cells are legal destinations"). Reuses a *minimal* evaluator
  (only the predicates those moves use), fail-closed to no-hint otherwise.
- **(c) Full client re-evaluator** (the original Layer 3 — *not recommended*):
  literal "on the client," but the critique shows its trustworthy scope is
  ~empty on the real corpus (LegalCustom everywhere), it can't resolve
  current-player or `any`-compositor children (§Corrections C5/C6), F1 residue
  makes optimistic hints unsafe (must emit only "undetermined"), and it needs
  the full conformance harness (§7). High cost, low real value.

## 4. Correctness decision table (client, per move) — corrected

Given the authority booleans + ledger (all per-viewer):

| Condition | Render |
|---|---|
| `LegalForAnyone` absent/false | hide the move |
| `LegalForPlayer` true | enabled |
| `LegalForPlayer` false, **some** `provisional` entry | keep reachable; hint "current choice isn't legal" (+ v2 detail) |
| `LegalForPlayer` false, **no** `provisional` entry | hard-disable (`aria-disabled`) + `LegalForPlayerError` reason |
| ViewingAsPlayer = Observer | view-only; single "observing" affordance |
| pending animation bundle / not live head | suppress actionable legality UI |

Never derive any of this by folding ledger `verdict`s (Rule 0).

## 5. Corrections the critique forced (record)

- **C1 (critical):** `LegalForPlayer` is default-field-bound; hard-disabling
  purely on it breaks field-bearing moves. → corrected Rule 0 (§1/§4).
- **C2:** the ledger array is in **bucket+plan (evaluation) order**, and
  `provisional == FieldDependent` tags membership. The v1 draft's O1
  (re-parse args / add a `fieldDependent` wire tag) is **withdrawn** — "first
  non-pass in array order" already matches the server.
- **C3:** the merged `LegalTemplates` table already ships in the Chest on
  /info. v1 draft's O2 ("add an additive /info field") is **withdrawn** — only
  client typing is needed.
- **C4:** animation/version bundles replace `moveForms` with legality-free
  forms; bind legality UI to settled live-head only.
- **C5:** `player.X`/current-player is **not** resolvable client-side
  (`CurrentPlayerIndex` is delegate code, not serialized). A full client
  evaluator must treat current-player-dependent predicates as inevaluable —
  or the server must additively ship the resolved index.
- **C6:** `any`/`allActivePlayers` ship as **one flat ledger entry** (no child
  sub-tree on the wire), so a client can't re-evaluate them even when
  `evaluable:true`. Scope compositors/quantifiers out of any client evaluator.
- **C7:** `LegalForPlayerError` is empty for observers and is the
  *default-bound* error for field-bearing moves — not "always exactly what
  ProposeMove returns."

## 6. If v3(c)/(b) is chosen — conformance harness (the correctness gate)
The corpus (`legal/testdata/conformance/*.json`) is the Go↔TS contract, but:
- it references **Go-built fixtures by name** → must serialize fixture state
  to JSON, and it must be the **per-viewer *sanitized* state** the client
  actually evaluates (not the full state), or the harness gives false green on
  the entire `evaluable`/sanitization dimension (critique CRITICAL-conformance);
- serialization must also carry **resolved `CurrentPlayerIndex` + the Chest**
  (not in `state.StorageRecord()`), or `proposerIsCurrentPlayer` /
  `componentPropEqualsCurrentPlayer` can't be reproduced;
- add a **plan-level** corpus (multi-predicate, both buckets) pinning overall
  verdict + first-failure template, or F7 bucket-ordering has zero coverage;
- generate fixtures **fresh in the same CI job** (or assert committed==fresh)
  so the snapshot can't silently drift from the live Go fixtures.

## 7. Integration points (verified)
- `selectors.ts::selectMoveLegality` (add nothing; already has `preconditions`).
- `boardgame-render-game.ts::_deriveLegality` (make it non-lossy; it feeds
  renderers).
- generic/admin move-form component (the "all moves enabled, bounce on submit"
  surface — primary v1 target).
- gathering/seating components read `LegalForPlayer` off the form directly —
  v1 graying should cover them consistently or explicitly scope them out.

## 8. Decisions for the user
- **D1 (the big one):** for live per-choice preview, pick **(a) server-preview
  endpoint** (authoritative, cheap, best ROI — but a round-trip), **(b) narrow
  client target-legality** (on-the-client, safe, limited surface), or **(c)
  full client re-evaluator** (literally "on the client," but low real value +
  high cost). *Recommendation: ship v1 now; do v3 as (a) unless "must be on the
  client" is a hard constraint, in which case (b).*
- **D2:** disabled-UI treatment — `aria-disabled`+inline reason (recommended,
  a11y-safe) vs native `disabled`.
- **D3:** scope gathering/seating and the admin move-form in or out of the v1
  graying pass.

## 9. Recommended next step
Build **v1** (real, on-the-client, low-risk, mostly plumbing-adoption) behind
the corrected Rule 0, then decide D1 for the preview tier.
