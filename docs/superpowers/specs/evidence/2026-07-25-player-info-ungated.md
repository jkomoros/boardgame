# Evidence pack: player-info (roster) animations are invisible to the completion gate

**Task:** Task 10 of the animatable-item unification plan
(`docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md`, Phase 2,
item 2). This is #714's headline fix.

**Claim:** `boardgame-player-roster` is a DOM **sibling** of `boardgame-render-game`
(both are children of `boardgame-game-view`; see
`server/static/src/components/boardgame-game-view.ts`'s `render()`, where
`<boardgame-player-roster id="player">` and `<boardgame-render-game id="render">` are
adjacent top-level elements in the same template). `render-game` installs its
`will-animate`/`animation-done` gate listeners on **itself**
(`boardgame-render-game.ts`'s `firstUpdated()`: `this.addEventListener('will-animate',
...)`), so events bubbling out of a roster-hosted animatable (a
`boardgame-status-text`'s nested `boardgame-fading-text`, gate participants since
Task 4/Phase 1) bubble past `boardgame-game-view` and up the document tree, but never
pass through `render-game`'s own listeners -- a sibling's bubble path structurally
cannot cross it. The gate closes (`all-animations-done` fires, `isAnimating` flips
false) the moment the **board's own** animations settle, with zero awareness that a
roster animation may still be in flight.

The design doc's current-state inventory already names this exact gap: "Structural
gaps: ... `boardgame-player-roster` is a DOM sibling of `boardgame-render-game`, so
player-info animations bubble past the gate listeners installed at
`boardgame-render-game.ts:347-348` and are silently un-gated." The plan brief
(`.superpowers/sdd/task-10-brief.md`) quotes #714's own checklist item verbatim:
**"verify that status-text and friends in render-player-info will also be waited
for"** -- i.e. the literal ask this task closes.

## Reproduction

A genuine, deterministic vehicle was needed to demonstrate the desync with a wide,
unambiguous timing margin (see the "Vehicle choice" discussion in
`tests/animations/parity/player-info-gate.spec.ts` for the full reasoning, including
why the two more literal vehicles -- a real memory scoring flow, and a real per-game
player-info renderer such as pig's own "Round Score" `boardgame-status-text` -- were
each ruled out, the latter by discovering a separate pre-existing bug documented
below).

Setup: create a real offline `pig` game, drain creation setup, then mount a **real**
`boardgame-fading-text` element (the actual production class, not a stand-in) as a
light-DOM child of the real, live, already-mounted `<boardgame-player-roster>`
element:

```js
const roster = /* deep-query into boardgame-game-view's shadow root */;
const el = document.createElement('boardgame-fading-text');
el.id = 'task10-roster-probe';
el.autoMessage = 'fixed';
roster.appendChild(el);
el.trigger = 1; // establishes baseline; no fade fires yet
```

Then, while a **real** board cycle is genuinely open (confirmed via the `animHooks`
log observing `{ ev: 'play', detail: 'boardgame-die' }` -- i.e. a real Roll-die move's
own board animation has started, exactly analogous to memory's WonCards changing
alongside a card-flip cycle), the probe is given a real `postAnimationDelay` (a public
`BoardgameAnimatableItem` property, not a fake timer) and re-triggered:

```js
el.postAnimationDelay = 2500;
el.trigger = 2; // real change -> fires a real gated play()
```

The postAnimationDelay makes the roster fade's own physical settlement land ~2.5s
after the click -- far past the die's own ~250ms spin -- turning "does the gate wait
for the roster?" from a same-order-of-magnitude race into an unmissable margin.

## Result: the gate closes without waiting

Captured in-page with `performance.now()` timestamps (both signals measured entirely
inside the browser via a `requestAnimationFrame` polling loop, to avoid Node-side
round-trip noise on a sub-second-to-few-second window):

```
gate reported closed at 4746.4ms (page-relative)
roster probe (boardgame-fading-text#task10-roster-probe) settled at 7013.3ms
```

`render-game.isAnimating` flips back to `false` (and `all-animations-done` fires) a
full **~2.3 seconds before** the roster-hosted fade it should have been waiting for
even settles. This is not a close race decided by scheduling luck -- the gap is
almost exactly the probe's declared 2500ms `postAnimationDelay`, i.e. the gate closed
on the board's own animation alone and had no knowledge the roster participant
existed. `tests/animations/parity/player-info-gate.spec.ts`'s first test
(`roster fading-text holds the gate: close waits for its settle during a real board
cycle`) encodes this reproduction and fails against the current (pre-Task-10) code
with exactly this shape of assertion failure:

```
Error: gate reported closed at 2916ms (page-relative) but the roster probe did not
settle until 5183.7ms; expected the gate to wait for the roster participant
```

(Run-to-run absolute timestamps vary with machine/network jitter in the setup phase;
the ~2.5s gap between gate-close and roster-settle is the stable, reproducible
signal -- confirmed across multiple runs.)

## Non-wedging direction is already (trivially) clean

The suite's second test asserts the complementary direction: a roster animation
firing with **no board cycle open** must never open, extend, or otherwise disturb the
gate. Against the current (unmodified) code this test already passes -- trivially,
since nothing is piped into the gate at all yet, so an out-of-cycle roster animation
is definitionally invisible to it. This test is retained unchanged through the
implementation below; per the plan, it "should pass before AND after -- it pins the
guard" that Step 3's implementation must preserve once piping is added.

## Aside: a separate, pre-existing, out-of-scope bug

While investigating the most literal vehicle (a real per-game player-info renderer,
e.g. pig's `boardgame-status-text` bound to `RoundScore`), a **separate, unrelated**
bug surfaced: `boardgame-player-roster.ts`'s template forwards its loaded flag to
`<boardgame-player-roster-item>` via `?renderer-loaded="${this.rendererLoaded}"` (a
boolean-**attribute** binding), but `boardgame-player-roster-item.ts`'s
`rendererLoaded` property has no explicit `attribute:` option, so Lit derives the
observed attribute name `rendererloaded` (no hyphen) -- confirmed empirically via
`customElements.get('boardgame-player-roster-item').observedAttributes`. The mismatch
means `rendererLoaded` never reaches `true` on the roster item, so
`boardgame-render-player-info.ts`'s `instantiateRenderer()` guard never passes: **no
per-game player-info renderer mounts in the live app currently, for any game**
(confirmed directly: on a fresh pig game, `<boardgame-render-player-info>`'s shadow
root contains only its placeholder comment even after 15+ seconds). This bug predates
this plan (introduced 2026-02-07, commit `a87137230`) and is unrelated to gate
topology; it has been flagged separately (not fixed here) since fixing it is not one
of this spec's two declared Phase 2 behavior changes and is out of scope for Task 10.
It is the reason this evidence pack's vehicle mounts a fading-text directly onto the
roster rather than through a real per-game renderer -- once that separate bug is
fixed, the exact same gate-piping mechanism this task adds will equally cover a real
`boardgame-status-text`, since it fires through the identical `will-animate`/
`animation-done` events from the same DOM subtree.

## Conclusion

The reproduction confirms the design doc's stated gap and #714's literal checklist
item: player-info (roster) animations are not currently waited for by the completion
gate. This justifies Task 10's declared change: `boardgame-game-view` piping
`will-animate`/`animation-done` bubbling out of `boardgame-player-roster` into
`render-game`'s gate instance (via new `gateWillAnimate`/`gateAnimationDone` thin
delegates), gated by `renderGame.isAnimating` on the `will-animate` direction only (so
an out-of-cycle roster animation can never open/wedge the gate) and always forwarded
on the `animation-done` direction (so a participant admitted at open can always
settle).
