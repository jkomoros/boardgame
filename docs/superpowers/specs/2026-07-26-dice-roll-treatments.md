# Dice-roll presentation treatments — design research catalog

**Status:** research input for a future dice API design doc. Not a plan, not a spec.
**Date:** 2026-07-26
**Scope:** enumerate the distinct *presentation treatments* a game author might want for a
dice roll, so the eventual API can make the common case trivial and the exotic case possible.

---

## 0. Grounding: what exists today

| Thing | File | Relevant facts |
|---|---|---|
| The die | `server/static/src/components/boardgame-die.ts` | Extends `BoardgameAnimatableItem`. Renders a **vertical strip of 2D faces** inside a 50px rounded square with `overflow: hidden`; "rolling" is `translateY(-size * selectedFace)` on `#inner`, played as a WAAPI motion track (`target: 'visual'`). Size is fixed at `--effective-die-size: 50px` scaled by `--die-scale`. Pips are absolutely-positioned divs; faces 1–6 only have pip layouts, everything else falls back to a numeral `<span>` that is currently `display:none` for 1–6. |
| Only consumer | `examples/pig/client/boardgame-render-game-pig.ts` | One `<boardgame-die>` bound to `.item` and `.action`, plus an `effectsForTransition()` that pulses/bursts on 1 and on max. Its `.die { height:100px }` CSS rule is dead — no element carries that class — which is itself evidence that **host-box sizing does not currently work**. |
| Server model | `components/dice/main.go` | `Value{Faces []int}` and `DynamicValue{Value, SelectedFace}`. Faces are **`[]int` only** — there is no server-side notion of a symbol, a color, a shape, or a polyhedron. `BasicDie(min,max)` is the only constructor. |
| Containers dice could live in | `boardgame-component-stack.ts` (`board \| fan \| grid \| pile \| spatial \| spread \| stack`), `boardgame-component-zone.ts` (semantic wrapper, excludes board/spatial), `boardgame-spatial-board.ts`, `boardgame-board-viewport.ts` | Stacks own their children's resting `layoutTransform` (a single write per relayout, deliberately). Zones add a labeled bordered box. |
| Clipping hazards (measured) | `grep overflow:hidden` | `boardgame-game-surface` (#48), `boardgame-game-board` (#65), `boardgame-board-viewport` (#48, #74), `boardgame-spatial-board` (#114, #152), `boardgame-effect-layer` (#152, #160), `boardgame-card`, `boardgame-player-panel`, and `boardgame-die` itself (#39). **Effectively every plausible dice ancestor clips.** |
| Motion ownership rule | `docs/superpowers/specs/2026-07-17-motion-foundations-design.md` § "Physical animation ownership" | Conflicts are prevented by **DOM ownership**, not a runtime arbiter: component host owns structural position/size; component-owned inner element owns face/orientation; overlays own emphasis; document overlay owns screen treatments. Any dice design must slot into this partition rather than invent a channel arbiter. |
| Reduced motion | `boardgame-animatable-item.ts:284`, `ARCHITECTURE.md:198` | Reduced motion is a **complete scheduling policy**, not a CSS afterthought — the gate already knows how to collapse a motion to its end state and settle. |

Two structural facts drive almost everything below:

1. **A die's own border box is a safe harbor.** Anything that tumbles strictly inside the
   element needs no portal, no measurement handoff, and no coordination with ancestors.
2. **Leaving that box is a cliff, not a slope.** The first pixel outside triggers the
   clipping/`transform`-containment hazard against a component tree where every candidate
   ancestor already clips. This is the single most important boundary in the API.

---

## 1. The catalog

Treatments are grouped by how far they stray from the die's own box, which is also roughly
their implementation cost order.

### Group A — Inside the die's own box (no portal)

---

#### T1. Reel (status die) — *what ships today*

**What the player sees.** A small rounded square sitting in the layout like a badge. On roll,
the face strip scrolls vertically past the window, decelerating, and stops on the result. It
reads unmistakably as "a value changed" and never as "an object was thrown." Extremely legible
at small sizes and on low-end devices; zero risk of the result being ambiguous.

**Exemplars.** Slot machines; Board Game Arena's compact score/dice widgets; most mobile
euro-game adaptations where the die is a scoreboard element rather than a prop; the framework's
own Pig renderer.

| | |
|---|---|
| **Rest** | In place, in the layout flow. Always present, always showing the last value. |
| **Rolls** | In place, inside its own `overflow:hidden` window. |
| **Camera/scale** | None. |
| **Container** | None. |
| **Count/variety** | 1–3. Homogeneous. Arbitrary face count is free (it's just a longer strip). |

**Author writes.** `<boardgame-die .item=${d} .action=${this.move(...)}>`. This is already true.

**Feasibility.** Trivially safe. No portal. No 3D. Sub-500ms. Regression-harness friendly
(pure transform on a known element). This is the honest floor and the reduced-motion terminal
state for every other treatment.

---

#### T2. Tumble in place — *proposed default*

**What the player sees.** A real cube that pitches and yaws through several rotations, bounces
once or twice against an invisible floor at the bottom of its own box, wobbles, and settles
showing the outcome face. The die never grows and never leaves its slot in the layout — a
player scanning the page sees the same rectangle before and after, but during the roll it is
unmistakably a physical object.

**Exemplars.** The die widget in most digital Yahtzee/Farkle implementations; Roll20's inline
d6; the die in Camel Up's app when shown in the player bar; countless "roll a die" web widgets.

| | |
|---|---|
| **Rest** | In place, in layout flow. Persistent. |
| **Rolls** | In place, within its own border box (padded so corners don't clip during rotation). |
| **Camera/scale** | None. Optionally a small "hop" of a few px, still inside the box. |
| **Container** | None (an invisible floor plane at the box bottom). |
| **Count/variety** | 1–5, laid out by ordinary flow/stack. Homogeneous d6. |

**Author writes.**

```ts
<boardgame-die .item=${this.state.Game.Die.Components[0]}
               .action=${this.move(MoveNames.RollDice)}></boardgame-die>
```

Identical to today. The *behavior* upgrades; the authoring surface does not.

**Feasibility.**
- ✅ No portal. The 3D subtree is entirely inside the die's shadow root, so `transform-style: preserve-3d` and `perspective` are the die's own business.
- ⚠️ Requires removing `overflow:hidden` from `#main` and giving the host padding, or the cube's corners clip during rotation. That padding changes the element's layout footprint — an incompatibility with T1 that must be handled (probably: the box is sized for the *rotating* extent, ~1.4× the face).
- ⚠️ The current 50px hard-coded size must become author-settable (host box or a token), or "a bigger die" is impossible without `--die-scale` hacks.
- ✅ Blocking duration ~700–900ms for one die; fine.
- ✅ Reduced motion: skip to landed face, announce.

---

#### T3. Symbol faces (a *modifier*, not a motion)

**What the player sees.** The same geometry as T1/T2, but each face carries an icon, a color
field, or a word instead of pips — a lightning bolt, a heart, a claw, a blank. Reading the
result is pattern recognition, not counting.

**Exemplars.** King of Tokyo (1/2/3/energy/heart/claw); Star Wars X-Wing (hit/crit/focus/blank);
Zombie Dice (brain/shotgun/runner); Camel Up (colored camels); Dice Forge; Marvel Champions;
Elder Sign.

| | |
|---|---|
| **Rest / rolls / camera / container** | Inherited from whichever motion treatment it modifies. |
| **Count/variety** | Any. Often mixed *colors* of an otherwise identical die. |

**Author writes.** This is where the server model bites: `Faces []int` cannot express "claw."
The face index is the only stable handle, so the author needs a client-side face renderer:

```ts
<boardgame-die .item=${d} .faceContent=${(face, index) => html`<kot-icon .kind=${ICONS[index]}>`}>
```

or, more declarative and more in keeping with the framework's slot idioms:

```ts
<boardgame-die .item=${d}>
  <template slot="face" data-index="0">⚡</template>
  <template slot="face" data-index="1">❤️</template>
  ...
</boardgame-die>
```

**Feasibility.**
- ✅ No motion implications at all.
- ⚠️ Accessibility: the numeric `Value` stops being the human-readable result. Needs a parallel
  `faceLabel(index) → string` for `aria-label` and the live-region announcement, or screen-reader
  users hear "4" for "claw."
- ⚠️ Cataloged as a treatment because it is a *presentation* decision authors reach for early,
  but it is really an axis (face content) that crosses every row of this table.

---

### Group B — Inside a container the author placed (portal usually still avoidable)

---

#### T4. Tray roll

**What the player sees.** A visible rectangular tray — felt floor, low walls, a bit of depth —
occupying a defined region of the layout. Dice live in the tray between rolls, scattered where
they last landed. On roll they leap, tumble, ricochet off the walls, collide with the floor, and
come to rest in new positions inside the tray. The tray is the stage; nothing outside it moves.

**Exemplars.** Yahtzee's tray/lid; King of Tokyo's dice area; Dice Throne; Board Game Arena's
dice widget for Can't Stop and Yahtzee; Roll20's roll tray; Quarriors' bag-and-roll area;
most "dice roller" apps (Dice by PCalc).

| | |
|---|---|
| **Rest** | In the tray, at their last landing position. Persistent between rolls. |
| **Rolls** | On the tray floor, bounded by tray walls. |
| **Camera/scale** | None. The tray is already the right size. |
| **Container** | Rendered 3D box/tray with real collision walls. |
| **Count/variety** | 2–8. Usually homogeneous; sometimes two colors. |

**Author writes.**

```ts
<boardgame-dice-tray .stack=${this.state.Game.Dice}
                     .action=${this.move(MoveNames.RollDice)}>
</boardgame-dice-tray>
```

i.e. **one element instead of one element** — the tray substitutes for the loose dice and takes
the stack, exactly the way `boardgame-component-stack` takes a stack today. Attributes for
`surface="felt|wood|none"`, `walls`, `floor-tilt`.

**Feasibility.**
- ✅ No portal *if* the tray's own box is the stage. The 3D scene is inside the tray's shadow root.
- ⚠️ **Transform ownership conflict.** A stack owns its children's resting `layoutTransform`
  (single write per relayout, by design). A physics sim wants to own the die's transform during
  flight. This is exactly the conflict the motion-foundations doc resolves by DOM ownership —
  so the tray must own resting *slots* (where a die comes to rest is presentation state, not game
  state) and hand transform ownership to the sim for the flight, returning it at settle. Getting
  this seam wrong is the main risk in the whole design.
- ⚠️ Landing positions are **presentation state with no server backing**. They must survive a
  re-render but must not be treated as truth; on reload they can be re-derived from a seed.
- ⚠️ Die-vs-die collision is an N-body problem (see anti-goals). V1 should give each die a
  reserved landing slot and let dice collide only with walls.
- ✅ Reduced motion: dice appear in their landing slots showing final faces.

---

#### T5. Cup shake and cast

**What the player sees.** An opaque cup sits mouth-down or mouth-up. On roll the cup lifts,
shakes (with a rattling implication), inverts, and dumps the dice onto the surface; the dice
tumble out and settle. Optionally the cup lands over the dice and must be lifted to reveal them
(Liar's Dice). The cup is an *actor* with its own choreography, and the dice do not exist
visually until it opens.

**Exemplars.** Yahtzee's cup; Backgammon dice cups; Liar's Dice / Perudo (cup conceals your own
dice); craps' "stickman" is the analog; Board Game Arena's Perudo implementation; Dudo apps.

| | |
|---|---|
| **Rest** | Inside the cup — i.e. **not present** until cast. After the cast they rest on the surface, or back in the cup. |
| **Rolls** | On the surface in front of/below the cup. |
| **Camera/scale** | None, though the cup's arc reads as a small camera move. |
| **Container** | Physical-ish cup + a landing surface. |
| **Count/variety** | 2–6. Homogeneous. |

**Author writes.**

```ts
<boardgame-dice-tray vessel="cup" .stack=${this.state.Game.Dice} .action=${...}>
```

— i.e. the cup is a **`vessel` attribute on the tray**, not a new element, because the tray
already owns "the region dice live in and land on." A `conceal` boolean covers the Liar's Dice
case (dice land face-down / under the cup until a reveal move).

**Feasibility.**
- ⚠️ Dice **appear** mid-animation. That collides with component continuity, which tracks stable
  component IDs across versions. Either the dice are always in the DOM and merely hidden inside
  the cup, or the framework's "faux component" machinery has to cover them. Prefer the former.
- ⚠️ Concealment is a *sanitization* question, not a presentation one: for Liar's Dice the server
  must genuinely not send other players' faces. The cup then renders from an unknown-value die,
  which the current `boardgame-die` has no concept of. **This treatment has a server-side
  prerequisite the others don't.**
- ✅ Otherwise same portal profile as T4.

---

#### T6. Personal per-player dice

**What the player sees.** Each player's dice sit in their own area of their player panel, in
their color. When it's your turn, *your* dice animate; everyone else's sit inert showing their
last result. The spatial association between dice and owner is the point — you can read the
table state without reading any text.

**Exemplars.** King of Tokyo; Backgammon (each side's dice belong to a side and land in that
player's quadrant); Dice Throne; Sagrada's private dice; Yahtzee scorepads in multiplayer apps;
Warhammer (each player rolls their own handful).

| | |
|---|---|
| **Rest** | In the owning player's panel/zone. Persistent. |
| **Rolls** | In place within that panel. |
| **Camera/scale** | Usually none. Sometimes the active player's tray scales up slightly (see T8). |
| **Container** | Small per-player tray, or none. |
| **Count/variety** | 1–6 per player × N players — **the total on screen can be large even when each roll is small.** |

**Author writes.** Nothing new — it's T2 or T4 instantiated inside the player-info renderer
(`boardgame-render-player-info-GAMENAME.ts`), which the framework already positions per player.

**Feasibility.**
- ⚠️ `boardgame-player-panel` has `overflow:hidden` (#57). A tumble that stays inside the die's
  box is fine; anything that lifts is clipped immediately. **This is the most likely place an
  author hits the clipping cliff without understanding why.**
- ⚠️ Blocking: only the active player's dice roll, so duration is bounded by one player's count.
- ✅ Reduced motion: trivial.

---

### Group C — On the shared surface

---

#### T7. Board throw

**What the player sees.** The dice are thrown onto the actual game board, among the pieces. They
skitter across the map, bounce, and settle wherever they land — possibly overlapping a territory,
a road, a pawn. Nothing about where they land matters mechanically, but the sense that the board
is a physical table is very strong. Players often *look* at where the dice landed relative to the
action.

**Exemplars.** Catan (dice thrown on the board); Monopoly; Risk; Warhammer 40k (dice on the
terrain); Backgammon (dice land in a quadrant, and *which* quadrant is real rules); Tabletop
Simulator and Tabletopia, whose entire model is "the board is the physics surface."

| | |
|---|---|
| **Rest** | Off-board in a corner, or wherever they last landed on the board. |
| **Rolls** | Across the board plane, in board coordinates. |
| **Camera/scale** | None, or a pan to follow if the board is panned/zoomed. |
| **Container** | The whole board; walls are the board edges (invisible bounds). |
| **Count/variety** | 2–6. |

**Author writes.**

```ts
<boardgame-spatial-board ...>
  <boardgame-dice-tray slot="surface" .stack=${this.state.Game.Dice} bounds="board"></boardgame-dice-tray>
</boardgame-spatial-board>
```

The interesting API question is that the tray here has **no visual body** — it is pure bounds.
That suggests `<boardgame-dice-tray surface="none">` rather than a separate element.

**Feasibility.**
- ⚠️ `boardgame-spatial-board` clips (#114, #152) and applies pan/zoom `transform`s. Dice inside
  it inherit the zoom, which is arguably correct (they're on the board) — but they cannot escape
  it, and a 3D-transformed child inside a transformed ancestor gets flattened into that ancestor's
  plane. **A perspective-correct die on a pan-zoomed board is genuinely hard.**
- ⚠️ If the board is zoomed out, dice become unreadably small at rest. Real implementations solve
  this by pairing T7 with T8 (roll big in the foreground, then place small on the board), which is
  a strong argument that these two treatments compose rather than compete.
- ⚠️ Backgammon's "which quadrant" case requires *landing position to be server-meaningful*, which
  the framework has no model for. Anti-goal.

---

#### T8. Foreground lift (zoom out and back)

**What the player sees.** The die (or the tray) grows smoothly out of its slot in the layout,
rising above the board with a shadow that lengthens, rotating as it comes forward, until it is
large and centered-ish and unmistakably the focus. It tumbles at that scale, lands, holds for a
beat so the result registers, then shrinks back down and settles into exactly the slot it came
from. **The position is continuous the entire time** — nothing ever jumps.

**Exemplars.** Digital Catan's roll; most mobile board-game adaptations (Ticket to Ride's dice-free
but the pattern is identical for its card reveals); Clash Royale-style "important thing comes
forward"; Gloomhaven Digital's modifier draw; Slay the Spire's card reveal; Marvel Champions app.

| | |
|---|---|
| **Rest** | Small, in the layout (in a tray or inline). |
| **Rolls** | In a foreground overlay above all game chrome. |
| **Camera/scale** | **Scale up and return** — the defining feature. |
| **Container** | Usually none while lifted (the overlay is the stage); optionally a scrim. |
| **Count/variety** | 1–6. |

**Author writes.**

```ts
<boardgame-dice-tray stage="lift" .stack=${...} .action=${...}>
```

One attribute. Everything hard about it is the framework's job.

**Feasibility.** ⚠️⚠️ **This is the portal treatment.** Everything about it is a hazard:
- Every plausible ancestor clips (`game-surface`, `game-board`, `player-panel`, `board-viewport`,
  `spatial-board`). Lifting *in place* is impossible; the dice must be **portaled to a document
  overlay**, which means a measured handoff: capture viewport-space geometry before the move,
  reparent, apply an inverse transform so frame N+1 is pixel-identical to frame N, then animate.
  The framework already has viewport-space geometry capture and center-to-center inversion
  (`motion/geometry.ts`, used by effect anchors) — this is the right primitive and it already
  exists.
- The return trip needs the *destination* geometry, which may have changed (layout reflow during
  the roll). Measure at settle time, not at lift time.
- Ancestors with `filter` or `opacity<1` (e.g. a dimmed inactive player panel) *trap* the child
  even without `overflow:hidden`. An author who dims non-active panels silently breaks lift.
- Reduced motion: **skip the lift entirely**, set the face in place, announce. Do not do a
  "reduced" lift — a half-lift is worse than none.
- Blocking: lift + tumble + hold + return is 1.6–2.5s. Acceptable once a turn; painful three times
  a turn (Yahtzee). Argues for `stage="lift"` being opt-in rather than default.

---

#### T9. Modal roll moment

**What the player sees.** The game dims. A large die (or set) occupies the center of the screen
against a scrim, tumbles with full drama, lands, and the result is displayed as a number/word with
emphasis — a crit gets particles, a fumble gets a shudder. Then it dismisses, either on a timer or
on a tap, and the game returns.

**Exemplars.** D&D Beyond's dice roller; Baldur's Gate 3's skill-check roll; Solasta; Gloomhaven
Digital's boss reveal; Marvel Champions' encounter card; any game where a single roll is *the*
dramatic beat of the turn.

| | |
|---|---|
| **Rest** | **Not present** — the dice exist only for the duration of the moment. |
| **Rolls** | Center screen, in a document-level overlay. |
| **Camera/scale** | Full takeover. |
| **Container** | Scrim; sometimes a rendered tray for flavor. |
| **Count/variety** | 1–3, often mixed polyhedra (a d20 plus modifier dice). |

**Author writes.** Because nothing rests anywhere, this is not a placed element at all — it is a
*transition response*, and should look like the existing effects hooks:

```ts
override effectsForTransition(ctx) {
  if (ctx.move?.AnimationKey !== MoveNames.RollAttack) return [];
  return [dice.moment({ die: ctx.after.Game.Attack.Components[0], tone: 'attention' })];
}
```

This is the one treatment whose natural authoring surface is a **hook returning a spec**, not an
element in `render()`. That asymmetry is worth designing for deliberately rather than discovering.

**Feasibility.**
- ✅ Portal is *not* a hazard here — there is no origin geometry to be continuous with, so the
  document overlay is the only home and no handoff is needed. Ironically the most dramatic
  treatment is easier than T8.
- ⚠️ "Dice must never jump" is satisfied vacuously (they fade/scale in from nothing), but the
  fade-in must be a real fade, not a pop.
- ⚠️ Blocking 2–3s including the dismiss beat. Must be skippable by tap.
- ⚠️ Focus management and `aria-modal` are real work. Reduced motion → a non-animated result
  panel with the same dismissal affordance.

---

### Group D — Persistence and post-roll interaction

---

#### T10. Keep-and-reroll pool

**What the player sees.** Five dice tumble into a tray. They settle and become *clickable*. The
player taps three to keep; those slide out of the tray into a "kept" row and dim slightly. The
player rolls again; only the two loose dice tumble. This repeats up to three times, then the whole
set is scored. Between rolls the dice are inert, persistent, and interactive.

**Exemplars.** Yahtzee (the canonical case); King of Tokyo; Zombie Dice; Dice Throne; Can't Stop
(reroll structure differs but the pool is the same); Elder Sign; Roll for the Galaxy.

| | |
|---|---|
| **Rest** | In the tray (loose) or in the kept row. Persistent across *several versions*. |
| **Rolls** | In the tray; only the loose subset animates. |
| **Camera/scale** | None (rolls happen ~3× per turn; lifting each time would be exhausting). |
| **Container** | Tray + a second "kept" zone. |
| **Count/variety** | 5–8, homogeneous. |

**Author writes.** The kept/loose split is **game state**, so it should be two stacks:

```ts
<boardgame-dice-tray .stack=${this.state.Game.LooseDice}
                     .action=${this.move(MoveNames.RollDice)}
                     .componentAction=${(c) => this.move(MoveNames.KeepDie).with({Die: c.ID})}>
</boardgame-dice-tray>
<boardgame-component-zone label="Kept" layout="spread" .stack=${this.state.Game.KeptDice}>
</boardgame-component-zone>
```

The migration from loose→kept is then **ordinary component motion** — the framework already
animates a component moving between stacks, and the author writes nothing for it. That is a
strong signal the dice tray should be a stack consumer, not a bespoke container.

**Feasibility.**
- ✅ Falls out of existing primitives almost entirely. The only new thing is "roll" as an
  animation on components that stay in the same stack.
- ⚠️ Requires the sim to animate a *subset* of a stack's components while the rest hold still —
  fine, but the tray must not relayout the held dice (a stack relayout would visibly nudge them).
- ⚠️ Blocking ×3 per turn. Duration budget must be per-turn-aware, not per-roll.
- ✅ Reduced motion: faces change in place; the keep/unkeep migration remains (it's structural).

---

#### T11. Dice as drafted resources

**What the player sees.** A handful of dice are rolled into a communal pool at the start of a
round. They then stop being dice-in-motion and start being *tokens with a number on them*: players
pick them up and place them onto board slots, player boards, or action spaces. The roll is a
one-time shuffle; the rest of the round the dice are inert draggable components.

**Exemplars.** Sagrada; Castles of Burgundy; Grand Austria Hotel; Troyes; Alien Frontiers;
Dice Forge; The Voyages of Marco Polo; Everdell's? (no) — but Kingsburg and Euphoria qualify.

| | |
|---|---|
| **Rest** | In a pool (post-roll), then on board slots (post-draft). Persistent for a whole round. |
| **Rolls** | In the pool region only. |
| **Camera/scale** | None. |
| **Container** | A pool zone; sometimes a bag-draw flourish first. |
| **Count/variety** | 5–20 in the pool. Often **mixed colors** with rules meaning. |

**Author writes.** The pool is a `boardgame-component-zone` of dice; placement is ordinary
stack-to-stack movement with existing `.action` bindings. The only dice-specific part is
"animate a roll for everything in this stack once":

```ts
<boardgame-component-zone label="Pool" layout="grid" .stack=${this.state.Game.Pool} roll-on-change>
```

i.e. a **die-level opt-in that says "when my face changes, tumble"** — which is arguably what
`boardgame-die` should do unconditionally, making this treatment free.

**Feasibility.**
- ✅ Everything except the tumble already exists.
- ⚠️ 20 dice rolling at once → see T13's budget problem.
- ⚠️ Color-as-rules-meaning again exceeds `Faces []int`; needs the same face/skin authoring hook
  as T3, plus a per-component color from the component's own `Values`.

---

### Group E — Scale and variety

---

#### T12. Mass roll ("bucket of dice")

**What the player sees.** Fifteen dice are dumped onto the table at once. They clatter, spread,
and settle in a scatter. The player then reads them as an *aggregate* — "seven hits" — not
individually. Good implementations immediately sort them: successes slide into one cluster,
failures into another, and a running tally appears.

**Exemplars.** Warhammer 40k / Age of Sigmar; Star Wars: Legion; Zombicide; Descent; Quarriors;
Dice Masters; Risk at its upper bound; Shadowrun and other d6-pool RPGs; Roll20/Foundry pool rolls.

| | |
|---|---|
| **Rest** | Not present until rolled; discarded after tallying. |
| **Rolls** | A wide surface, then sorted into result clusters. |
| **Camera/scale** | None — there's no room to zoom 15 dice. |
| **Container** | A wide tray or the whole board. |
| **Count/variety** | 8–50. Homogeneous but often two colors (attack/defense). |

**Author writes.**

```ts
<boardgame-dice-tray .stack=${this.state.Game.AttackDice}
                     budget="1200"
                     summarize=${(dice) => `${dice.filter(d => d.Value >= 4).length} hits`}>
```

The `summarize` hook is the interesting bit: at this cardinality the *tally* is the result and the
individual dice are texture.

**Feasibility.** ⚠️⚠️
- **Blocking duration is the binding constraint.** If each die takes 900ms and they're staggered,
  15 dice is 3–5s of frozen game. The API must express a **total budget** from which per-die
  timing is derived (fast stagger, short flights), not a per-die duration. `motion.stagger()`
  already exists as the deterministic-start-order primitive and is the right substrate.
- Baked WAAPI keyframes × 15 dice × ~60 keyframes each is a lot of DOM animation objects. Needs a
  cap (probably ~12 individually-simulated dice; beyond that, groups share baked trajectories with
  different offsets/rotations — visually indistinguishable, and deterministic).
- Die-vs-die collision at this count is out of the question. Reserved landing slots in a jittered
  grid read as a "scatter" convincingly enough.
- Reduced motion → all dice appear sorted into their result clusters with the tally. Arguably
  *better* than the animation.

---

#### T13. Mixed polyhedra (RPG set)

**What the player sees.** A d20 and a d6 tumble together; they are visibly different shapes, and
they read as a *set* — the d20 is the verdict and the d6 is the damage. Faces carry numerals, not
pips. Landing on a triangular face of a d20 is visually distinct from a cube coming to rest.

**Exemplars.** D&D Beyond; Roll20; Foundry VTT; Dice Throne (d6-heavy but mixed); Dungeon Crawl
Classics (the full weird-dice set); any digital RPG combat.

| | |
|---|---|
| **Rest** | Usually not present until rolled (see T9), or in a tray. |
| **Rolls** | Tray or modal. |
| **Camera/scale** | Often modal (T9). |
| **Container** | Tray or none. |
| **Count/variety** | 1–6, **heterogeneous shapes**. |

**Author writes.**

```ts
<boardgame-die .item=${d} shape="d20">
```

One attribute — but the server has to supply the shape somehow, because `Faces []int` of length 20
does not imply an icosahedron (it could be a 20-value spinner). Either the author states it
client-side (simple, dishonest) or `dice.Value` grows a `Shape` field (honest, a server change).

**Feasibility.** ⚠️
- d6 in CSS 3D is six divs. d8/d12 are tractable (8 triangles / 12 pentagons via `clip-path`).
  **d10 and d20 are a genuine slog** — 20 triangular faces each needing a hand-computed
  `rotate3d` + translate, plus numeral orientation on each, plus the fact that a d20's "up" face
  is a face-normal not an axis. This is a day of geometry per shape, not an hour.
- Relabeling to force the outcome (the project's stated mechanism) is *easier* on a d20, since
  players cannot track 20 faces.
- Reduced motion / low-end fallback: degrade any shape to the T1 reel showing the numeral. This is
  a genuinely good fallback and should probably also be the fallback for `shape` values the
  framework doesn't implement — **an author can write `shape="d100"` and get something reasonable**.

---

### Group F — Degenerate and adjacent

---

#### T14. Roll-and-move chain

**What the player sees.** The dice roll and settle; the moment they settle, the player's pawn
begins hopping along the track, one space per beat, the count visibly matching the pips. The roll
and the movement are one continuous sentence, not two events.

**Exemplars.** Monopoly; Snakes & Ladders; Camel Up; Formula D; Talisman; Pig's cousin games;
basically every roll-and-move ever printed.

| | |
|---|---|
| **Rest** | Anywhere (typically a small tray near the board). |
| **Rolls** | Tray or board. |
| **Camera/scale** | None for the roll; the board may pan to follow the pawn. |
| **Container** | Any. |
| **Count/variety** | 2. |

**Author writes.** Nothing dice-specific — the *framework already does this*. `moves.HopAlongPath`
emits one game version per hop, and the client animates each version separately. The roll is one
version; each hop is another. The author's only job is move decomposition on the **server**, which
TUTORIAL.md already teaches.

**Feasibility.**
- ✅ Falls out of existing architecture. Worth cataloging precisely *because* it needs no new API —
  it's evidence that "the roll blocks advancement" is already the right model.
- ⚠️ The handoff must not have a dead beat: the settle of the last die and the first hop should be
  adjacent. That's a version-queue timing question the existing gate already owns.

---

#### T15. Result-only (no dice)

**What the player sees.** No dice geometry at all. A number, a word, or a symbol appears with
emphasis — "8", "MISS", "★★☆" — perhaps with a brief pulse or a counter roll-up. Fast, unambiguous,
zero motion cost.

**Exemplars.** Board Game Arena's minimal/fast modes; chess.com-style clean UI; many euro
adaptations that treat dice as randomness rather than as props; every implementation's
accessibility mode.

| | |
|---|---|
| **Rest** | Not present. |
| **Rolls** | Nothing rolls. |
| **Camera/scale** | None. |
| **Container** | None. |
| **Count/variety** | N/A. |

**Author writes.** `<boardgame-status-text>` / their own markup plus an `fx.pulse`. Already possible.

**Feasibility.** ✅ Trivial. **This is the terminal degradation for every other treatment** and
should be the framework's automatic reduced-motion substitute — not a bespoke thing each treatment
invents.

---

#### T16. Drag-to-throw (player-authored trajectory)

**What the player sees.** The player presses on the dice, drags, and flicks. The throw direction
and force follow the gesture; the dice fly where you threw them and land on the result. Grabbing
and re-rolling feels tactile and personal.

**Exemplars.** Tabletop Simulator; Tabletopia; Dice by PCalc; most physical-feeling mobile
backgammon and Yahtzee apps; Board Game Arena's shake-to-roll on mobile.

| | |
|---|---|
| **Rest** | Tray or table. |
| **Rolls** | Wherever thrown. |
| **Camera/scale** | None. |
| **Container** | Table/tray with bounds. |
| **Count/variety** | 1–5. |

**Author writes.** `<boardgame-dice-tray throw="gesture">` — deceptively small.

**Feasibility.** ⚠️⚠️⚠️ **Recommended v1 anti-goal.**
- The outcome is server-authoritative and known *before* animation, but a gesture-thrown die starts
  moving *before* the server round-trip. Either the die flies with an unknown destination and the
  sim is re-solved mid-flight (possible — the relabeling trick means only the final orientation
  needs to change — but the trajectory must already be baked, and re-baking mid-flight breaks the
  "bake once into WAAPI" model), or the die does not move until the server answers, which makes the
  gesture feel broken.
- The animation regression harness can't drive gestures, so this treatment is untested by
  construction.
- Two sources of truth for "when did the roll start."

---

## 2. Capability axes

The treatments above are not 16 independent things; they are points in an 8-dimensional space.
**This is the key output — these are what the API must parameterize.**

| # | Axis | Values | Which treatments span it |
|---|---|---|---|
| **A** | **Face presentation** | pips / numerals / authored symbols; and geometry: flat reel / cube / other polyhedron | T1, T2, T3, T13 |
| **B** | **Container & collision surface** | none (free) → invisible bounds → rendered tray → cup/vessel → the whole board | T2, T4, T5, T7, T12 |
| **C** | **Staging** — where the roll happens relative to the layout | **in own box** / **in a placed container** / **lifted to overlay & returned** / **modal takeover** | T2, T4, T8, T9 |
| **D** | **Persistence & rest** | ephemeral (absent until rolled, gone after) / persistent in place / persistent in a container at last landing / migrates into another stack after settling | T9, T2, T4, T10, T11 |
| **E** | **Cardinality & heterogeneity** | 1 / few (2–8) / many (9–50); homogeneous / mixed color / mixed shape | T1, T4, T12, T13 |
| **F** | **Choreography** | simultaneous / staggered / sequential-per-version / multi-stage (reroll rounds) | T4, T10, T12, T14 |
| **G** | **Post-settle interactivity** | inert / selectable (keep) / draggable (place) | T4, T10, T11 |
| **H** | **Handoff** | settle → nothing / → another component's motion / → stack transfer / → summary tally | T14, T10, T12 |

Two derived observations:

- **Axis C is not continuous — it has a cliff between value 2 and value 3.** "In a placed
  container" requires no portal; "lifted" requires portal + measured continuous handoff against a
  component tree where every candidate ancestor clips. The API should make this boundary visible in
  its shape (a distinct `stage` attribute) rather than hiding it behind a size or duration knob
  that authors will cross accidentally.
- **Axes B, D, G and H are all already served by the stack/zone system.** A dice tray that consumes
  a stack gets rest positions, persistence, per-component actions, and stack-to-stack migration for
  free. The genuinely new capability is only Axis A (geometry), Axis C (staging), and Axis F
  (roll-specific choreography). **The dice API should be small.**

---

## 3. The 80% case

**Recommended zero-config default: T2, Tumble in place.**

Rationale:

- It's the only treatment that is *categorically* safe: it never leaves the die's own border box,
  so no portal, no ancestor coordination, no clipping hazard, no measurement handoff, and it works
  identically inside a player panel, a spatial board, a zone, or bare markup.
- It reads as a physical die (the thing authors reach for dice to communicate) without any of the
  staging cost.
- Its blocking duration (~800ms) is acceptable in every game shape including the reroll-heavy ones.
- It is a strict visual upgrade to what Pig does today with **zero change to Pig's source**.

**What `<boardgame-die>` with no configuration should do:**

```ts
<boardgame-die .item=${die}></boardgame-die>
```

1. Render a **cube** with pips (1–6) or numerals (anything else), sized from the host box, with a
   sensible default so a bare element is visible.
2. On `SelectedFace` change, **tumble in place**: 2–3 tumbles, one floor bounce, settle on the
   outcome face, ~800ms, deterministic per (component ID, version) so the regression harness sees
   the same frames every run.
3. Stay strictly within its own box (the box is pre-sized for the rotating extent).
4. Announce the result to a live region using `Value`, and keep `aria-label` current.
5. Under `prefers-reduced-motion`, **snap to the face and announce** — i.e. degrade to T15.
6. If `.action` is bound, be a button that rolls; otherwise be inert and non-focusable.
7. If the face count or shape isn't something it can render as a solid, **fall back to the T1 reel**
   rather than erroring. (An author writing `shape="d100"` or a 37-face die gets something sane.)

Point 7 matters: the framework already throws on unknown stack layouts
(`boardgame-component-stack.ts:1393`). Dice geometry should *not* follow that precedent — the
graceful-degradation ladder down to a reel down to a numeral is the whole reason this API can
promise "exotic is possible."

---

## 4. Progressive disclosure ladder

| Rung | Treatment | What the author adds |
|---|---|---|
| 0 | **T2** Tumble in place | `<boardgame-die .item=${d}>` — nothing. |
| 1 | **T1** Reel (opt *down*, for dense/small UI) | `motion="reel"` |
| 1 | **T13** Non-cube shape | `shape="d20"` |
| 1 | Bigger / calmer / snappier | `size=`, `duration=` — or better, `energy="low|normal|high"` so authors can't set 4s |
| 2 | **T3** Symbol faces | `<template slot="face" data-index="0">⚡</template>` × N, plus `.faceLabel` for a11y |
| 2 | **T4** Tray roll | swap the element: `<boardgame-dice-tray .stack=${s} .action=${a}>` |
| 2 | **T6** Per-player dice | put rung 0 or 2 in the player-info renderer — no new API |
| 2 | **T14** Roll-and-move | no client API; decompose the move server-side (already taught) |
| 3 | **T5** Cup | `vessel="cup"` on the tray (+ `conceal` and a server sanitization change for Liar's Dice) |
| 3 | **T7** Board throw | `surface="none" bounds="board"` on a tray slotted into the spatial board |
| 3 | **T10** Keep-and-reroll | second stack + `.componentAction` — mostly existing primitives |
| 3 | **T12** Mass roll | `budget="1200"` and optionally `.summarize` |
| 4 | **T8** Foreground lift | `stage="lift"` — one attribute, all the machinery framework-side |
| 4 | **T9** Modal moment | one **hook**: return `dice.moment({...})` from `effectsForTransition()` |
| 5 | Bespoke choreography | `motionPlanForRoll(context)` hook returning a `dice.throw({from, energy, spin, surface, stagger})` spec — sibling to the existing `motionCohortsForTransition()` / `motionTransfersForTransition()` family |

The shape to aim for, stated as a rule: **default is one element; common variations are one
attribute; container treatments are one different element; staging is one attribute; exotic is one
element plus one hook.** Rung 5 is deliberately the same *kind* of thing as the animation hooks
authors already know from TUTORIAL.md, so it costs no new concepts.

---

## 5. Anti-goals for v1

| Anti-goal | Why |
|---|---|
| **Die-vs-die collision** | N-body sim, ordering nondeterminism (bad for the regression harness), and a large blocking cost. Reserved landing slots with jitter are visually indistinguishable at the sizes involved. |
| **Drag-to-throw / gesture trajectory (T16)** | Requires motion to start before the server-authoritative outcome arrives, contradicting the "bake the trajectory once" model. Also untestable by the regression harness. |
| **d10 / d20 / d100 solids** | Days of per-shape 3D geometry for a payoff only RPG-shaped games need. Ship d6 (and possibly d8/d12); degrade everything else to the reel with numerals — which is legible and honest. |
| **Server-meaningful landing position** (Backgammon quadrants, "dice must land in the tray") | The server has no model of dice position, and inventing one makes presentation state authoritative. |
| **Dice tower** | The hidden interior means components vanish and reappear, which fights component continuity and `IDsLastSeen` inference for a purely decorative payoff. |
| **Free camera / orbit during a roll** | The "must never jump" constraint plus pan-zoom ancestors plus 3D flattening; and no exemplar needs it that `stage="lift"` doesn't cover. |
| **Canvas/WebGL dice** | Already excluded by project constraint (invisible to the regression harness). Restating it here because a mass roll (T12) is the case that will tempt someone. |
| **Author-settable raw physics constants** (restitution, friction, mass) | Invites unbounded blocking durations and non-deterministic-looking output. Expose `energy` and `budget` instead. |
| **A global dice-preset registry** | The motion-foundations doc explicitly prefers small typed helpers returning specs over a global preset registry. Dice should follow. |
| **Concealed dice as a client-side effect** (Liar's Dice) | Concealment is sanitization. Doing it in the renderer ships other players' faces to the browser. |

---

## 6. Accessibility and reduced motion

The framework already treats reduced motion as a **scheduling policy** with a settled end state
(`boardgame-animatable-item.ts:284`, `ARCHITECTURE.md:198`), not as a CSS media query — so the
correct model is "every treatment degrades to T15 plus its own structural residue."

| Treatment | Reduced-motion degradation |
|---|---|
| T1 Reel | Face swaps instantly. |
| T2 Tumble | Face swaps instantly, no rotation. |
| T3 Symbols | Same, but the live-region text must come from `faceLabel(index)`, **not** `Value`. |
| T4 Tray / T5 Cup | Dice appear at their landing slots with final faces. No cup choreography. |
| T6 Per-player | As T2/T4, per panel. |
| T7 Board throw | Dice appear at their board positions. |
| T8 Lift | **No lift at all.** Face changes in place. A partial lift is worse than none. |
| T9 Modal | Modal still appears (it's information, not decoration) but without motion; the dismiss affordance and focus trap are unchanged. |
| T10 Keep-and-reroll | Faces swap; the keep/unkeep migration remains as ordinary structural motion (which the framework already reduces correctly). |
| T11 Draft pool | Faces swap; drag/placement unaffected. |
| T12 Mass roll | Dice appear **already sorted into result clusters** with the tally. Plausibly better than the animation. |
| T13 Polyhedra | Degrade to reel/numeral. |
| T14 Roll-and-move | Roll snaps; hops remain (they're separate versions and carry real information). |

Cross-cutting a11y requirements, regardless of treatment:

- The result must reach a **live region** at settle time, once per roll, in the die's own vocabulary
  (numeral or symbol label). Today's `aria-label` is a static `'Die'` / `'Roll die'` and does not
  carry the value — that is a bug the dice work should fix regardless.
- The roll blocks game advancement, so the blocking interval must have a **hard cap** and must be
  skippable; a keyboard user should never be stuck watching 15 dice.
- `aria-busy` is already wired for pending submissions; the settle gate should extend it so
  assistive tech knows the outcome isn't final yet.
- Focus must not be stolen by a lift (T8) or lost when a modal (T9) dismisses.

---

## 7. Open questions for the design doc

1. **Does `dice.Value` grow a `Shape` field?** Client-only `shape=` is simpler but makes the die's
   physical identity a rendering opinion rather than a component property. `Faces []int` also can't
   express symbol faces, so T3 and T13 both push on the same seam.
2. **Who owns a die's transform in a tray?** Stacks own resting `layoutTransform` by design; the sim
   wants the transform during flight. Proposed answer, following the motion-foundations partition:
   tray owns resting slots, sim borrows the transform for the flight and returns it at settle — but
   this needs to be written down before anyone codes it.
3. **Are landing positions persisted?** They're presentation state with no server backing. Seeded
   from (component ID, version) makes them stable across reloads without inventing game state.
4. **Is `<boardgame-dice-tray>` a new element or a `layout="tray"` on the existing stack?** The
   tray needs collision walls and a floor plane, which no other layout has; but making it a stack
   layout gets zone/board/spatial composition for free.
5. **Duration budget granularity.** Per-die, per-roll, or per-turn? T10 (3 rolls/turn) and T12
   (15 dice/roll) push in opposite directions and both break a naive per-die `duration`.
