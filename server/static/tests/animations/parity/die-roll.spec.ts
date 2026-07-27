import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, expectCleanGate } from '../helpers';

// Task 10: <boardgame-die> stops pretending to roll and actually rolls.
//
// A roll now derives a seed from (component ID, roll identity), runs
// `dice-sim.ts` once, paints the server's value onto whichever face the
// simulation landed (`die-faces.ts`), and plays the baked trajectory
// (`dice-bake.ts`) as ONE sampled curve track on the visual channel.
//
// What these tests pin, and why each one exists:
//
//   1. One GATED animation, declared for the trajectory's own duration. The
//      declared duration is what the render-game watchdog arms itself from, and
//      `timing: 'immediate'` is what keeps a multi-second physics bake out of
//      the version slot's 600ms clamp.
//   2. The value the SERVER chose ends up on the face the physics turned up,
//      and that face is the one facing the player. Both halves matter:
//      `selectedFace` is an INDEX into `faces`, and the face assignment is
//      recomputed per roll, so "the die shows 4" is only true if the right
//      facet carries it AND that facet is readable.
//   3. Determinism. The renderer rebuilds the roll from scratch on every mount,
//      so the same (id, roll count) has to reproduce the same tumble bit for
//      bit, and a different roll count has to produce a different one.
//   4. Reduced motion: no tumble, right face.
//   5. Interruption: `finishGatedAnimations()` mid-roll leaves the right face
//      presented. This is the `fill: 'none'` resting-style contract.
//   6. The watchdog does not fire for a multi-second roll, in the real app.
//   7. The face labels swap at the START of the roll (see the label-swap
//      section below).
//   8. Every keyframe is a literal transform list -- no `var()`, no `calc()` --
//      because either one silently demotes the tumble off the compositor.
//
// Task 11 adds three groups below: what counts as a roll (the die's own
// `DynamicValues.RollCount`, so a throw landing on the face already showing
// still tumbles), what a roll's DURATION is (the throw with the simulator's
// dead trailing hold cut off), and the first browser witness of the gate
// watchdog's `expectedSettleMs` extension.



interface MountOptions {
  faceCount: number;
  selectedFace?: number;
  faces?: number[];
  componentId?: string;
  stateVersion?: number;
  dieSize?: string;
  /** `DynamicValues.RollCount`; omitted from the item entirely when null. */
  rollCount?: number | null;
}

/**
 * Mounts one <boardgame-die> in the served app page and installs its FIRST
 * item. Installing the first item is not a roll -- the die is being shown a
 * state it was already in -- so nothing animates here.
 *
 * Face VALUES are deliberately never equal to their own index (10, 20, 30...):
 * `SelectedFace` is an index and confusing it for a value is the silent bug
 * this component invites.
 */
async function mountDie(page: import('@playwright/test').Page, options: MountOptions) {
  await page.goto('/');
  await page.evaluate(async (opts) => {
    document.querySelectorAll('boardgame-die').forEach((el) => el.remove());
    await import('/src/components/boardgame-die.ts');
    const die = document.createElement('boardgame-die') as any;
    die.id = 'fixture-die';
    die.style.cssText = 'position:fixed;top:200px;left:200px;z-index:9999;';
    die.style.setProperty('--die-size', opts.dieSize ?? '100px');
    const faces = opts.faces ?? Array.from({ length: opts.faceCount }, (_, i) => (i + 1) * 10);
    if (opts.stateVersion !== undefined) die.stateVersion = opts.stateVersion;
    const dynamic: any = {
      SelectedFace: opts.selectedFace ?? 0, Value: faces[opts.selectedFace ?? 0],
    };
    // The real server always sends this (components/dice's DynamicValue); a
    // null asks for the pre-RollCount item shape, which the die still supports.
    if (opts.rollCount !== null) dynamic.RollCount = opts.rollCount ?? 0;
    die.item = {
      ID: opts.componentId ?? 'fixture-component',
      Values: { Faces: faces },
      DynamicValues: dynamic,
    };
    document.body.appendChild(die);
    await die.updateComplete;
    // _itemChanged runs in updated(), which schedules a second render pass.
    await die.updateComplete;
  }, options as any);
}

/** What one roll looked like, read out while the tumble is still running. */
interface RollCapture {
  /** Number of animations live on #inner. */
  animations: number;
  /** `will-animate` declarations the die dispatched for this roll. */
  declared: number[];
  /** The gated live count the animatable-item kernel is holding. */
  gatedCount: number;
  duration: number;
  easing: string;
  keyframeCount: number;
  first: string;
  last: string;
  frames: string[];
  /** #inner's inline resting transform, written by the motion-track kernel. */
  resting: string;
  /** Face values on the facets, by face index, DURING the tumble. */
  valuesDuringRoll: (string | undefined)[];
  /** The same, read at the instant the tumble's animation was created. */
  valuesAtPlay: (string | undefined)[];
  /** The trajectory the die should be playing, computed independently. */
  expected: {
    /** The TRIMMED duration: what the die animates. */
    durationMs: number;
    /** The simulator's own duration, dead trailing hold included. */
    rawDurationMs: number;
    presented: number;
    restingTransform: string;
    seed: number;
  };
  serverValue: number;
}

/**
 * Installs a SECOND item -- the real path a roll arrives by -- and reads the
 * roll out while it is still running. Everything the component is supposed to
 * be playing is recomputed here from the same pure modules, through the
 * component's own exported seed/trajectory derivation, so that a wrong seed, a
 * wrong tray or a wrong pixel radius all show up as a mismatch rather than as
 * "some animation ran".
 */
async function rollDie(
  page: import('@playwright/test').Page,
  options: {
    selectedFace: number; stateVersion?: number; faceCount: number; componentId?: string;
    /** `DynamicValues.RollCount`; omitted from the item entirely when null. */
    rollCount?: number | null;
  },
): Promise<RollCapture> {
  return await page.evaluate(async (opts) => {
    const dieModule: any = await import('/src/components/boardgame-die.ts');
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const facesModule: any = await import('/src/motion/die-faces.ts');
    const bakeModule: any = await import('/src/motion/dice-bake.ts');

    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const inner = root.querySelector('#inner') as HTMLElement;
    const stage = root.querySelector('#stage') as HTMLElement;
    const radiusPx = parseFloat(getComputedStyle(stage).fontSize) / 2;

    // Every animate() call landing on #inner, recorded with the keyframes and
    // the RESOLVED timing exactly as the kernel handed them to the browser.
    //
    // Not read back from getAnimations()/getKeyframes(), for two reasons: a
    // duration-0 effect (which is what reduced motion resolves to) is finished
    // and therefore not "relevant" the instant it starts, so it never appears
    // there at all; and Chromium re-serializes transform components on the way
    // out (six decimals become exponential notation), so a byte comparison
    // against what dice-bake.ts emitted is impossible through that door.
    const calls: any[] = [];
    const originalAnimate = Element.prototype.animate;
    Element.prototype.animate = function (this: Element, keyframes: any, options: any) {
      const animation = originalAnimate.call(this, keyframes, options);
      if (this === inner) {
        calls.push({
          keyframes,
          options,
          // What the facets carried at the instant the tumble was created --
          // before it could have painted a single frame.
          faceValues: (Array.from(root.querySelectorAll('.facet[data-face-index]')) as HTMLElement[])
            .map((el) => el.dataset.faceValue),
        });
      }
      return animation;
    } as typeof Element.prototype.animate;

    const declared: number[] = [];
    const onWillAnimate = (event: Event) =>
      declared.push((event as CustomEvent).detail.expectedSettleMs);
    die.addEventListener('will-animate', onWillAnimate);

    const faces = die.faces.slice();
    if (opts.stateVersion !== undefined) die.stateVersion = opts.stateVersion;
    const version = opts.stateVersion ?? die.stateVersion ?? 0;
    const dynamic: any = {
      SelectedFace: opts.selectedFace, Value: faces[opts.selectedFace],
    };
    // A throw the mount did not install, which is what the server sends for a
    // roll -- INCLUDING one that lands on the face already showing.
    //
    // It defaults to the state version rather than to 1 because the ROLL COUNT
    // is the seed (see `_rollIdentity`): a test that wants two scenarios to
    // throw differently has to vary the count, and every test here expresses
    // "a different situation" by varying the version. The mount installs count
    // 0, so any version >= 1 is a change and therefore a throw.
    if (opts.rollCount !== null) dynamic.RollCount = opts.rollCount ?? version;
    die.item = {
      ID: opts.componentId ?? 'fixture-component',
      Values: { Faces: faces },
      DynamicValues: dynamic,
    };
    // Three passes, all inside the same frame: the install writes the face, the
    // face change plans the roll, and the pass after that -- the first one whose
    // render carries the roll's own face values -- plays it.
    for (let pass = 0; pass < 4; pass++) await die.updateComplete;
    die.removeEventListener('will-animate', onWillAnimate);
    Element.prototype.animate = originalAnimate;

    const geometry = geometryModule.dieGeometry(opts.faceCount);
    // The die seeds from its ROLL COUNT where the item reports one, and from the
    // state version only where it does not. Derived here the same way the
    // component derives it, so a seed taken from the wrong number shows up as
    // every frame mismatching rather than as a test quietly agreeing with a bug.
    const identity = opts.rollCount === null ? version : (opts.rollCount ?? version);
    const seed = dieModule.dieRollSeed(opts.componentId ?? 'fixture-component', identity);
    const trajectory = dieModule.dieRollTrajectory(
      geometry, opts.componentId ?? 'fixture-component', identity);
    // What is PLAYED is the throw with its trailing dead hold cut off (see
    // settledTrajectory), so every expectation below is derived from that and
    // not from the raw trajectory -- including its duration, which is the
    // trimmed trajectory's last sample time.
    const settled = dieModule.settledTrajectory(trajectory.dice[0]);
    const settledDurationMs = settled.samples[settled.samples.length - 1].t;
    const presented = facesModule.presentedFaceIndex(
      geometry, settled.restingOrientation);

    const frames: string[] = (calls[0]?.keyframes ?? []).map((frame: any) => String(frame.transform));
    return {
      animations: calls.length,
      declared,
      gatedCount: die._liveGatedCount,
      duration: Number(calls[0]?.options?.duration ?? -1),
      easing: String(calls[0]?.options?.easing ?? ''),
      keyframeCount: frames.length,
      first: frames[0] ?? '',
      last: frames[frames.length - 1] ?? '',
      frames,
      valuesAtPlay: calls[0]?.faceValues ?? [],
      resting: inner.style.transform,
      valuesDuringRoll: (Array.from(root.querySelectorAll('.facet[data-face-index]')) as HTMLElement[])
        .map((el) => el.dataset.faceValue),
      expected: {
        durationMs: settledDurationMs,
        rawDurationMs: trajectory.durationMs,
        presented,
        restingTransform: bakeModule.restingTransform(settled, { radiusPx }),
        seed,
      },
      serverValue: faces[opts.selectedFace],
    };
  }, options as any);
}

/**
 * Where the die has actually come to rest, measured from the RENDER: the
 * composed matrix from #stage down to each facet, so the reading is the whole
 * chain (#inner's tumble, #orient's pose, the facet's own placement) and not
 * one link of it.
 */
async function restingRead(page: import('@playwright/test').Page) {
  return await page.evaluate(() => {
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const stage = root.querySelector('#stage') as HTMLElement;
    const composed = (element: HTMLElement) => {
      const chain: HTMLElement[] = [];
      for (let n: HTMLElement | null = element; n && n !== stage; n = n.parentElement) chain.unshift(n);
      let matrix = new DOMMatrix();
      for (const node of chain) {
        const value = getComputedStyle(node).transform;
        if (value && value !== 'none') matrix = matrix.multiply(new DOMMatrix(value));
      }
      return matrix;
    };
    const facets = Array.from(root.querySelectorAll('.facet[data-face-index]')) as HTMLElement[];
    // A facet's own +z after the whole chain IS its outward normal on screen;
    // the one with the largest z component is the one facing the player.
    const scored = facets.map((el) => {
      const m = composed(el);
      return {
        faceIndex: Number(el.dataset.faceIndex),
        value: el.dataset.faceValue,
        label: el.dataset.faceLabel,
        towardsCamera: m.m33,
      };
    });
    const front = scored.reduce((best, row) => (row.towardsCamera > best.towardsCamera ? row : best));
    const inner = root.querySelector('#inner') as HTMLElement;
    return {
      front,
      byIndex: Object.fromEntries(scored.map((row) => [row.faceIndex, row])),
      ariaLabel: (root.querySelector('#main') as HTMLElement).getAttribute('aria-label'),
      innerTransform: getComputedStyle(inner).transform,
      innerInline: inner.style.transform,
      liveAnimations: inner.getAnimations().length,
    };
  });
}

test.describe('boardgame-die physics roll', () => {
  test('plays exactly one gated animation for the trajectory\'s own duration', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 2 });

    expect(roll.animations).toBe(1);
    expect(roll.gatedCount).toBe(1);
    // ONE will-animate, declaring the physics duration. This is what the
    // render-game watchdog arms itself from; a declaration short of the real
    // duration force-closes the cycle mid-tumble.
    expect(roll.declared).toEqual([roll.expected.durationMs]);
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    // A real physics roll is seconds, not one 250ms animation-length slot --
    // the assertion that the duration was not clamped into a version slot.
    expect(roll.duration).toBeGreaterThan(600);
    // A sampled track carries its own timing; the kernel must pin linear or
    // the trajectory is time-warped by the default ease-in-out.
    expect(roll.easing).toBe('linear');
    // resolution: Math.round(durationMs / 16.7), clamped to [2, 256].
    const expectedFrames = Math.min(256, Math.max(2, Math.round(roll.expected.durationMs / 16.7)));
    expect(roll.keyframeCount).toBe(expectedFrames);
  });

  // Trap 2, and the reason the roll asks for `timing: 'immediate'`.
  //
  // The kernel's DEFAULT policy is 'version', which clamps an animation into the
  // companion cycle's slot -- 600ms, against a physics bake of two or three
  // seconds. A clamped tumble is geometrically faithful and physically absurd
  // (a die falling at five times gravity), and the same policy can resolve to
  // SKIP outright, which makes playMotionTracks report 'not-started' and takes
  // its sibling tracks down with it. Neither is visible in solo play, because
  // the ambient animation context is null there and the policy has nothing to
  // clamp against -- so this test supplies one.
  test('a companion cycle does not clamp the roll into its slot', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      die.animationContext = {
        version: 12,
        startAtMs: Date.now(),
        slotDurationMs: 600,
        maxAnimationDurationMs: 600,
      };
    });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 12 });
    // A live context, and one the kernel would otherwise have used.
    expect(roll.expected.durationMs).toBeGreaterThan(600);
    expect(roll.animations).toBe(1);
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    expect(roll.declared).toEqual([roll.expected.durationMs]);
  });

  test('every keyframe is a literal transform list, so the tumble composites', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 2, stateVersion: 7 });
    expect(roll.frames.length).toBeGreaterThan(20);
    // A var() or calc() anywhere in a transform keyframe forfeits compositing
    // and drops a multi-second tumble onto the main thread. This is exactly the
    // pattern the feature replaces, so it is asserted rather than assumed.
    expect(roll.frames.filter((value) => /var\(|calc\(/.test(value))).toEqual([]);
    // Every frame is the constant scene prefix -- a literal translate3d in px
    // and up to two literal rotate3d turns -- in front of the baked matrix3d.
    // (Chromium re-serializes small components in exponential notation on the
    // way back out of getKeyframes(); what was WRITTEN never contains one, and
    // it is `calc`/`var` that would cost the compositor either way.)
    const shape = /^translate3d\([-0-9.px, ]+\)( rotate3d\([-0-9., ]+deg\))* matrix3d\([-0-9., ]+\)$/;
    expect(roll.frames.filter((value) => !shape.test(value))).toEqual([]);
    // The curve's end IS the bake's own resting transform, byte for byte. This
    // is the contract that stops the die twitching when its animation is
    // removed: fill is 'none', so the element renders its resting style the
    // instant the tumble finishes.
    expect(roll.last.endsWith(roll.expected.restingTransform)).toBe(true);
  });

  // Trajectory positions are in die CIRCUMRADII; matrix3d translations are in
  // pixels. The only place that conversion happens is the `radiusPx` the die
  // measures off its own box, and getting it wrong is silent: the die still
  // tumbles, still lands on the right face, still squares up -- it just moves a
  // couple of pixels instead of across its own box, or a screenful instead. So
  // the distance the die actually covers is asserted, in units of its own size.
  test('the tumble travels a real fraction of the die across the screen', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, dieSize: '100px' });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 2, stateVersion: 21 });
    const translation = (value: string) => {
      const parts = (value.match(/matrix3d\(([^)]*)\)/) as RegExpMatchArray)[1].split(',').map(Number);
      return parts.slice(12, 15);
    };
    const rest = translation(roll.last);
    const peak = Math.max(...roll.frames.map((frame) => {
      const at = translation(frame);
      return Math.hypot(at[0] - rest[0], at[1] - rest[1], at[2] - rest[2]);
    }));
    // The die is 100px across. A throw that covers less than half of that is
    // not a throw; one that covers three times it has left the board.
    expect(peak / 100).toBeGreaterThan(0.6);
    expect(peak / 100).toBeLessThan(3);
  });

  for (const selectedFace of [0, 1, 3, 5]) {
    test(`lands face ${selectedFace}'s value on the face physics turned up`, async ({ page }) => {
      await mountDie(page, { faceCount: 6, selectedFace: 4, stateVersion: 1 });
      const roll = await rollDie(page, {
        faceCount: 6, selectedFace, stateVersion: 100 + selectedFace,
      });
      await page.evaluate(async () => {
        const die = document.getElementById('fixture-die') as any;
        const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
        await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
      });
      const read = await restingRead(page);

      // The physics decided WHICH face lands up; the value assignment paints
      // the server's value onto exactly that face.
      expect(read.front.faceIndex).toBe(roll.expected.presented);
      expect(read.front.value).toBe(String(roll.serverValue));
      expect(read.ariaLabel).toBe(`Die showing ${roll.serverValue}`);
      // ...and that face is genuinely readable. Without a camera looking DOWN
      // at the tray, the physics up-face is edge-on to the viewer (the world's
      // up axis is screen-up, not screen-out) and the player reads a side face.
      expect(read.front.towardsCamera).toBeGreaterThan(0.7);
    });
  }

  // The die on screen has to be the die the physics threw, not its mirror
  // image.
  //
  // The trap is that CSS's own frame -- x right, y DOWN, z toward the viewer --
  // is LEFT handed, so the map from the geometry's right-handed body frame has
  // to have determinant -1 (flip Y, and only Y) to keep the solid's handedness.
  // `(x, -y, -z)`, which looks like the obvious "turn it the right way up", is a
  // proper rotation and renders the MIRROR of the solid: every pair of faces
  // along the axis the two conventions disagree on comes out swapped, so a die
  // whose physics landed one face up draws the opposite one. Nothing about a
  // stationary d6 shows it; the roll does, because `dice-bake.ts` maps a
  // simulated pose the other way.
  //
  // Measured from the render, and stated physically: a right-handed triple of
  // outward face normals (which is what `assignFaceValues` gives the three
  // lowest values, and why a real Western die reads 1-2-3 counter-clockwise)
  // must wind COUNTER-CLOCKWISE around the corner they share, seen from
  // outside. The 2D cross product of their screen projections is then
  // NEGATIVE -- CSS y points down, which reverses the sign of a winding -- and
  // in fact equals minus the corner's own z, which is what the assertion pins.
  test('renders the die the physics threw, not its mirror image', async ({ page }) => {
    await mountDie(page, { faceCount: 6, faces: [1, 2, 3, 4, 5, 6], selectedFace: 0, stateVersion: 1 });
    await rollDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 3 });
    const winding = await page.evaluate(async () => {
      const geometryModule: any = await import('/src/motion/die-geometry.ts');
      const geometry = geometryModule.dieGeometry(6);
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const stage = root.querySelector('#stage') as HTMLElement;
      const composed = (element: HTMLElement) => {
        const chain: HTMLElement[] = [];
        for (let n: HTMLElement | null = element; n && n !== stage; n = n.parentElement) chain.unshift(n);
        let matrix = new DOMMatrix();
        for (const node of chain) {
          const value = getComputedStyle(node).transform;
          if (value && value !== 'none') matrix = matrix.multiply(new DOMMatrix(value));
        }
        return matrix;
      };
      const facets = Array.from(root.querySelectorAll('.facet[data-face-index]')) as HTMLElement[];
      // The facets carrying values 1, 2 and 3, which assignFaceValues winds
      // right-handed in the BODY frame.
      const carrying = [1, 2, 3].map((value) =>
        facets.find((el) => el.dataset.faceValue === String(value)) as HTMLElement);
      const dot = (a: number[], b: number[]) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
      const cross = (a: number[], b: number[]) => [
        a[1] * b[2] - a[2] * b[1], a[2] * b[0] - a[0] * b[2], a[0] * b[1] - a[1] * b[0]];
      const bodyNormals = carrying.map((el) =>
        geometry.faces[Number(el.dataset.faceIndex)].normal as number[]);
      const rendered = carrying.map((el) => {
        const m = composed(el);
        return [m.m31, m.m32, m.m33];
      });
      const [p, q, r] = rendered;
      return {
        bodyHandedness: dot(bodyNormals[0], cross(bodyNormals[1], bodyNormals[2])),
        screenWinding: (q[0] - p[0]) * (r[1] - p[1]) - (q[1] - p[1]) * (r[0] - p[0]),
        cornerTowardsCamera: p[2] + q[2] + r[2],
      };
    });

    // The premise: die-faces.ts really did wind 1-2-3 right-handed.
    expect(winding.bodyHandedness).toBeGreaterThan(0.5);
    // The corner they share is well off the screen plane, so the winding below
    // is being read from one side of it or the other and its sign means
    // something.
    expect(Math.abs(winding.cornerTowardsCamera)).toBeGreaterThan(0.2);
    // For three mutually perpendicular normals the projected winding is exactly
    // MINUS the corner's own z when the solid is drawn as it is, and exactly
    // PLUS it when the drawing is mirrored. No slack to hide in, and no
    // dependence on which side of the die the corner happens to be on.
    expect(winding.screenWinding).toBeCloseTo(-winding.cornerTowardsCamera, 5);
  });

  // Trap 8: `RollTrajectory.cocked` says the simulator could not settle the die
  // flat in eight throws, and the renderer must not then quietly print the value
  // on a face that is not really up.
  //
  // The answer is to make it up: the scene's reading pose AIMS at the face the
  // throw landed, so the face carrying the value ends up exactly where the pose
  // puts it whether the throw settled at 0.0 degrees (nearly all of them) or at
  // 2.4 (measured worst for a d12 over 40 throws) or at the 3-plus that counts
  // as cocked. There is no floor drawn, so a couple of degrees of world tilt is
  // invisible, and the alternative -- refusing to animate a rare roll -- trades
  // an unreadable die for a die that sometimes does not move.
  //
  // Pinned on a d12, whose throws are the ones that actually land tilted, at the
  // first state version that produces one.
  test('lands the reading face exactly square, even when the throw did not', async ({ page }) => {
    await mountDie(page, { faceCount: 12, selectedFace: 0, stateVersion: 1, dieSize: '160px' });
    // Where this die's own resting pose puts the face it presents, read off a
    // d12 that has not been thrown. The cocked throw below has to land there
    // too -- exactly there, not near it -- and reading it rather than writing a
    // number down is what keeps this test pinned to the pose the player sees
    // instead of to a constant that has to be edited whenever the pose moves.
    const restingTowardsCamera = await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const facet = root.querySelector('.facet[data-face-index="0"]') as HTMLElement;
      const orient = root.querySelector('#orient') as HTMLElement;
      const m = new DOMMatrix(getComputedStyle(orient).transform)
        .multiply(new DOMMatrix(getComputedStyle(facet).transform));
      return m.m33;
    });
    const tilted = await page.evaluate(async () => {
      const dieModule: any = await import('/src/components/boardgame-die.ts');
      const geometryModule: any = await import('/src/motion/die-geometry.ts');
      const facesModule: any = await import('/src/motion/die-faces.ts');
      const geometry = geometryModule.dieGeometry(12);
      const rotate = (q: number[], v: number[]) => {
        const t = [
          2 * (q[1] * v[2] - q[2] * v[1]),
          2 * (q[2] * v[0] - q[0] * v[2]),
          2 * (q[0] * v[1] - q[1] * v[0]),
        ];
        return [
          v[0] + q[3] * t[0] + (q[1] * t[2] - q[2] * t[1]),
          v[1] + q[3] * t[1] + (q[2] * t[0] - q[0] * t[2]),
          v[2] + q[3] * t[2] + (q[0] * t[1] - q[1] * t[0]),
        ];
      };
      for (let rollCount = 2; rollCount < 80; rollCount++) {
        const trajectory = dieModule.dieRollTrajectory(geometry, 'fixture-component', rollCount);
        const die = trajectory.dice[0];
        const presented = facesModule.presentedFaceIndex(geometry, die.restingOrientation);
        const world = rotate(
          die.restingOrientation as number[], geometry.faces[presented].normal as number[]);
        const degrees = (Math.acos(Math.min(1, Math.max(-1, world[1]))) * 180) / Math.PI;
        if (degrees > 0.5) return { rollCount, degrees };
      }
      return null;
    });
    // If this ever comes back null the fixture has stopped exercising the case
    // it exists for, which is a louder failure than a silently vacuous pass.
    expect(tilted, 'no d12 throw in the first 78 roll counts landed off-square').not.toBeNull();
    expect(tilted!.degrees).toBeGreaterThan(0.5);

    const roll = await rollDie(page, {
      faceCount: 12, selectedFace: 7, stateVersion: 2, rollCount: tilted!.rollCount,
    });
    await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
    });
    const read = await restingRead(page);
    const landed = read.byIndex[roll.expected.presented];
    expect(landed.value).toBe(String(roll.serverValue));
    // Squared up: the reading face sits at exactly the tilt the resting pose
    // gives it, and NOT at that tilt plus or minus the throw's own lean. The
    // throw above is off by `tilted.degrees`, which without the pose aiming at
    // the LANDED face moves this by ~0.01 -- hundreds of times the tolerance.
    expect(landed.towardsCamera).toBeCloseTo(restingTowardsCamera, 4);
    // ...and, since the pose aims at that face, it is now genuinely the most
    // square-on facet on the die. It was not: a d12's neighbours are 63 degrees
    // apart, and at the old fixed camera elevation one of them was nearer
    // face-on than the face carrying the value.
    expect(read.front.faceIndex).toBe(roll.expected.presented);
  });

  test('is deterministic in (component id, roll count), and only in those', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-a' });
    const first = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, rollCount: 5, componentId: 'die-a',
    });
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-a' });
    const again = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, rollCount: 5, componentId: 'die-a',
    });
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-a' });
    const later = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, rollCount: 6, componentId: 'die-a',
    });
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-b' });
    const other = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, rollCount: 5, componentId: 'die-b',
    });

    // Same identity, same roll -- the whole reason the simulation is seeded
    // rather than random: a remount mid-roll must not re-throw the die.
    expect([again.first, again.last, again.duration])
      .toEqual([first.first, first.last, first.duration]);
    // The NEXT throw is a different one, and so is another die's.
    expect(later.first).not.toBe(first.first);
    expect(other.first).not.toBe(first.first);
  });

  // THE SEED IS THE ROLL COUNT, NOT THE STATE VERSION.
  //
  // Both numbers are deterministic, so the test above passes either way. What
  // separates them is that the state version moves while ONE throw is on
  // screen: it advances for every move any player makes, and a game view
  // mounting installs a die three times at three different versions with the
  // die untouched. Seeded on the version, a component that re-planned its roll
  // after any of those -- a remount, a tab returning to the foreground, a
  // replay scrubbing -- would derive a DIFFERENT trajectory for the SAME throw
  // and the die would change its path in mid-air.
  //
  // So: the same roll count at two unrelated state versions must produce the
  // same throw, byte for byte, and the count must still be doing real work.
  test('re-derives the same throw for one roll count at any state version', async ({ page }) => {
    await mountDie(page, {
      faceCount: 6, selectedFace: 0, stateVersion: 3, rollCount: 0, componentId: 'die-a',
    });
    const early = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 4, rollCount: 6, componentId: 'die-a',
    });
    await mountDie(page, {
      faceCount: 6, selectedFace: 0, stateVersion: 30, rollCount: 0, componentId: 'die-a',
    });
    const late = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 77, rollCount: 6, componentId: 'die-a',
    });

    // Seeded on the version these are two different throws; seeded on the count
    // they are the same throw seen twice. Every keyframe, not just the ends.
    expect(late.frames).toEqual(early.frames);
    expect(late.duration).toBe(early.duration);
    expect(late.expected.presented).toBe(early.expected.presented);
    // ...and the throw that was played is the one the roll COUNT derives, which
    // is what a seed silently taken from the version would fail even if the two
    // captures above happened to agree.
    expect(early.last.endsWith(early.expected.restingTransform)).toBe(true);
    expect(late.last.endsWith(late.expected.restingTransform)).toBe(true);

    // The count is not being ignored either: the next throw, at the very same
    // state version, is a different one.
    await mountDie(page, {
      faceCount: 6, selectedFace: 0, stateVersion: 30, rollCount: 0, componentId: 'die-a',
    });
    const next = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 77, rollCount: 7, componentId: 'die-a',
    });
    expect(next.first).not.toBe(late.first);
  });

  // WHEN the die's other faces take their new numbers.
  //
  // The assignment is recomputed per roll, so every face except the landed one
  // carries a different number afterwards. That swap has to happen either as
  // the roll STARTS or as it ENDS, and it is deliberately the start: at t=0 the
  // die is already airborne and tumbling in the first painted frame, so no
  // number that changes is legible; at t=1 the swap would land on a die the
  // player is reading, and the number under their eye would change after it had
  // stopped moving.
  test('swaps the face values as the roll starts, not as it ends', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 4, stateVersion: 42 });
    // Read at the instant element.animate() was called, which is before the
    // tumble can have painted anything: the landed face already carries the
    // server's value, and so does every facet in the first frame anyone sees.
    expect(roll.valuesAtPlay[roll.expected.presented]).toBe(String(roll.serverValue));
    expect(roll.valuesAtPlay).toEqual(roll.valuesDuringRoll);
    // ...and the whole set is still a permutation of the die's own faces.
    expect(roll.valuesDuringRoll.map(Number).sort((a, b) => a - b))
      .toEqual([10, 20, 30, 40, 50, 60]);
  });

  test('an interrupted roll still leaves the right face presented', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 1, stateVersion: 5 });
    expect(roll.duration).toBeGreaterThan(600);
    // Mid-flight: the cycle sweep force-settles gated animations, and with
    // fill:'none' the element then renders its RESTING style. If the resting
    // style and the curve's end disagreed by so much as a rounding digit, the
    // die would visibly jump -- and if it were not written at all, the die
    // would snap back to its pre-roll pose.
    await page.waitForTimeout(120);
    const midway = await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      die.finishGatedAnimations();
      return true;
    });
    expect(midway).toBe(true);
    const read = await restingRead(page);
    expect(read.liveAnimations).toBe(0);
    expect(read.front.faceIndex).toBe(roll.expected.presented);
    expect(read.front.value).toBe(String(roll.serverValue));
    expect(read.front.towardsCamera).toBeGreaterThan(0.7);
    // ...and the pose it is left in is where the curve ENDS, not where it was
    // interrupted and not the pre-roll pose. Compared as resolved matrices,
    // because the two strings go through different serializers.
    const jump = await page.evaluate((expected: string) => {
      const die = document.getElementById('fixture-die') as any;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      const held = new DOMMatrix(getComputedStyle(inner).transform);
      const ends = new DOMMatrix(expected);
      return held.toFloat64Array().reduce(
        (worst, value, index) => Math.max(worst, Math.abs(value - ends.toFloat64Array()[index])), 0);
    }, roll.last);
    expect(jump).toBeLessThan(1e-3);
  });
});

// Task 11: WHAT COUNTS AS A ROLL.
//
// Task 10 triggered the tumble on `selectedFace`, which is an INDEX into the
// face list. A throw landing on the face already showing leaves that index
// alone -- one throw in six for a d6, measured at 4 of 20 rolls in pig -- so
// the player clicked Roll and the die did not move. The state version cannot
// stand in (it moves three times during one mount with the die untouched), and
// neither can "the move that produced this state", which says nothing about
// WHICH die was thrown. `components/dice`'s `DynamicValue.RollCount` is the
// server saying it, per die, and it is now the trigger.
test.describe('boardgame-die roll trigger', () => {
  test('a throw landing on the face already showing still tumbles', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 1, rollCount: 4 });
    // Same face, same value, same everything the pre-RollCount die could see.
    const roll = await rollDie(page, {
      faceCount: 6, selectedFace: 3, stateVersion: 2, rollCount: 5,
    });
    expect(roll.animations).toBe(1);
    expect(roll.gatedCount).toBe(1);
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    expect(roll.declared).toEqual([roll.expected.durationMs]);
    await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
    });
    const read = await restingRead(page);
    // ...and it lands showing what it was already showing, on whichever facet
    // the physics turned up, not on the one it started from.
    expect(read.front.faceIndex).toBe(roll.expected.presented);
    expect(read.front.value).toBe(String(roll.serverValue));
    expect(read.ariaLabel).toBe(`Die showing ${roll.serverValue}`);
  });

  // The other half of the contract, and the reason a heuristic like "animate
  // whenever the game version moved and this die is rollable" is not acceptable:
  // a die's item is re-installed for every state a game reaches, and almost none
  // of them threw it.
  test('an install that did not throw the die does not tumble', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, rollCount: 4 });
    // A later state version, a fresh item object, even a different face -- but
    // the same roll count, so nothing threw this die.
    const quiet = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, rollCount: 4,
    });
    expect(quiet.animations).toBe(0);
    expect(quiet.declared).toEqual([]);
    expect(quiet.gatedCount).toBe(0);
  });

  // A die driven by hand, or by a game that does not use `components/dice`,
  // reports no roll count at all. It keeps the pre-RollCount trigger, which is
  // the best available signal when there is no better one -- and which every
  // other test in this file exercises through the same fallback.
  test('a die that reports no roll count still tumbles on a face change', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, rollCount: null });
    const roll = await rollDie(page, {
      faceCount: 6, selectedFace: 3, stateVersion: 2, rollCount: null,
    });
    expect(roll.animations).toBe(1);
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    // ...and it is seeded from the STATE VERSION, which is the only identity
    // such a die has. `expected` is derived from the version here, so this is
    // the fallback seed asserted and not merely tolerated.
    expect(roll.last.endsWith(roll.expected.restingTransform)).toBe(true);
  });
});

// Task 11: the trailing hold.
//
// `dice-sim.ts` only calls a die at rest after 0.3s of continuous stillness, and
// it emits that hold as samples. Animating it is not free: the roll is GATED, so
// it is 0.3s of a median 1.0s roll during which the whole game's animation cycle
// waits on a die that has already stopped. It is cut off, and this is what says
// so and what says it costs nothing to look at.
test.describe('boardgame-die roll duration', () => {
  test('plays the motion and not the dead hold at the end of it', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 2 });
    // What is played is the trimmed duration, and it is materially shorter --
    // for this seed 633ms against the simulator's 933ms.
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    expect(roll.expected.rawDurationMs - roll.expected.durationMs).toBeGreaterThan(250);
    // The cut costs nothing to LOOK at: the pose at the cut is the pose at the
    // end, to well under a tenth of a pixel on this 100px die. Measured through
    // the same bake the keyframes come from, so this is the rendered pose and
    // not an approximation of it.
    const drift = await page.evaluate(async (version: number) => {
      const dieModule: any = await import('/src/components/boardgame-die.ts');
      const geometryModule: any = await import('/src/motion/die-geometry.ts');
      const bakeModule: any = await import('/src/motion/dice-bake.ts');
      const geometry = geometryModule.dieGeometry(6);
      const trajectory = dieModule.dieRollTrajectory(geometry, 'fixture-component', version);
      const full = trajectory.dice[0];
      const cut = dieModule.settledTrajectory(full);
      const atCut = new DOMMatrix(bakeModule.restingTransform(cut, { radiusPx: 50 }));
      const atEnd = new DOMMatrix(bakeModule.restingTransform(full, { radiusPx: 50 }));
      const a = atCut.toFloat64Array();
      const b = atEnd.toFloat64Array();
      return a.reduce((worst, value, index) => Math.max(worst, Math.abs(value - b[index])), 0);
    }, 2);
    // Translation components are pixels, rotation components are unit-scale, so
    // this bound is a pixel bound on the worst of the two.
    expect(drift).toBeLessThan(0.1);
  });
});

test.describe('boardgame-die physics roll, reduced motion', () => {
  test('runs no tumble and still presents the right face', async ({ browser }) => {
    const context = await browser.newContext({ reducedMotion: 'reduce' });
    const page = await context.newPage();
    try {
      await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
      const roll = await rollDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 11 });
      // Reduced motion is a complete scheduling policy, not a default an
      // explicit duration can override: the effect resolves to duration 0 and
      // occupies the cycle for no time at all, so nothing tumbles...
      expect(roll.duration).toBe(0);
      expect(roll.declared).toEqual([0]);
      expect(roll.expected.durationMs).toBeGreaterThan(600);
      await page.evaluate(async () => {
        const die = document.getElementById('fixture-die') as any;
        const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
        await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
      });
      const read = await restingRead(page);
      expect(read.liveAnimations).toBe(0);
      // ...and the die is left holding the roll's resting pose, showing the
      // value the server chose on the face that landed up.
      expect(read.front.faceIndex).toBe(roll.expected.presented);
      expect(read.front.value).toBe(String(roll.serverValue));
      expect(read.front.towardsCamera).toBeGreaterThan(0.7);
      expect(read.ariaLabel).toBe(`Die showing ${roll.serverValue}`);
    } finally {
      await context.close();
    }
  });
});

test.describe('boardgame-die physics roll, in the app', () => {
  test('a multi-second roll never trips the gate watchdog', async ({ page }) => {
    test.setTimeout(120000);
    await createOfflineGame(page, 'pig');
    await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 30000 });
    // Every roll tumbles now, including one that lands on the face already
    // showing: `DynamicValues.RollCount` is the trigger, not the face index.
    // The retry is bounded and is NOT about that -- it is tolerance for an API
    // binary built before `RollCount` existed, against which the die falls back
    // to the face change and a same-face roll still moves nothing.
    let before = await gateSnapshot(page);
    let declared = 0;
    for (let attempt = 0; attempt < 4 && declared === 0; attempt++) {
      before = await gateSnapshot(page);
      const die = page.getByRole('button', { name: 'Roll die' });
      await expect(die).toBeEnabled({ timeout: 20000 });
      await die.click();
      declared = await liveRollDuration(page);
      if (declared === 0) await expectCleanGate(page, before, 30000);
    }
    expect(declared).toBeGreaterThan(0);
    // The watchdog's floor is well under a physics roll, so the die's
    // will-animate declaration is the only thing keeping the cycle from being
    // force-closed mid-tumble.
    await expectCleanGate(page, before, 60000);
    const after = await gateSnapshot(page);
    expect(after.watchdogFirings).toBe(before.watchdogFirings);
  });

  // Task 11: THE WATCHDOG EXTENSION, WITNESSED.
  //
  // `AnimationGate` arms a backstop at a 4000ms FLOOR when a cycle opens and
  // force-closes the cycle if it is still open then -- which, for an animation
  // that was legitimately going to run longer, means the tumble is cut off
  // mid-air and the game moves on. The escape hatch is `willAnimate`'s
  // `expectedSettleMs`: a participant that declares a longer settle re-arms the
  // backstop at `declaration + 1500ms`.
  //
  // Until now nothing in a browser had ever exercised that -- the parity
  // README listed it as an accepted blind spot owned only by the gate's unit
  // tests -- because no scenario ran long enough. A physics roll can: the
  // simulator caps a throw at 5000ms, and a big barrel reaches the cap. The
  // seed below is a d48 that is STILL TUMBLING when the cap cuts it off, so it
  // declares 5000ms against a 4000ms floor, and a full second of it exists only
  // because the declaration moved the deadline.
  //
  // The die is a fixture, but nothing else here is: it is mounted inside the
  // live renderer, so it registers with the real ambient registry, its
  // `will-animate` reaches the real gate through the real listener, and it is
  // thrown inside a real cycle opened by a real move.
  test('a roll past the watchdog floor extends the deadline instead of being cut off',
    async ({ page }) => {
      test.setTimeout(180000);
      await createOfflineGame(page, 'pig');
      await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 30000 });

      // A d48 whose throw runs to the simulator's own 5000ms cap. The seed is
      // (component ID, roll count), so the count is what selects it.
      const LONG_FACES = 48;
      const LONG_ROLL_COUNT = 118;
      const LONG_ID = 'watchdog-die';
      // Mount it inside the live renderer, so it joins the real gate.
      // `<boardgame-die>` is already defined by the app; importing the module
      // again through a second URL would re-run customElements.define.
      await page.evaluate(async (opts) => {
        const find = (root: DocumentFragment | Document): HTMLElement | null => {
          const direct = root.querySelector('boardgame-render-game') as HTMLElement | null;
          if (direct) return direct;
          for (const element of Array.from(root.querySelectorAll('*'))) {
            const shadow = (element as HTMLElement).shadowRoot;
            if (!shadow) continue;
            const found = find(shadow);
            if (found) return found;
          }
          return null;
        };
        const renderGame = find(document) as HTMLElement;
        const host = renderGame.shadowRoot as ShadowRoot;
        // The die is mounted inside the real renderer so that its
        // `animatableRegistry` and `animationContext` climbs reach it and it is
        // a genuine participant in the real gate. Its THROW, though, is this
        // test's own: the seed comes from the item's `RollCount`, which is set
        // below to one that is known to run past the watchdog floor.
        const wrapper = document.createElement('div') as any;
        wrapper.style.cssText = 'position:absolute;top:0;left:0;';
        const die = document.createElement('boardgame-die') as any;
        die.id = 'watchdog-die';
        const faces = Array.from({ length: opts.faces }, (_, i) => i + 1);
        die.item = {
          ID: opts.id,
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 0, Value: 1, RollCount: 0 },
        };
        wrapper.appendChild(die);
        host.appendChild(wrapper);
        // Kept on window: the die now lives in a shadow root, where
        // document.getElementById cannot reach it.
        (window as any).__watchdogDie = die;
        await die.updateComplete;
        await die.updateComplete;
      }, { faces: LONG_FACES, id: LONG_ID });

      const before = await gateSnapshot(page);
      // Throw the long die INSIDE the cycle pig's own roll opens: a declaration
      // made while no cycle is open is not a gate participant at all, and the
      // watchdog it would have to outlast never gets armed.
      await page.getByRole('button', { name: 'Roll die' }).click();
      const observed = await page.evaluate(async (opts) => {
        const hooks = (window as any).__bgAnimTestHooks;
        const opensBefore = hooks.gateOpens;
        const start = performance.now();
        while (hooks.gateOpens === opensBefore && performance.now() - start < 20000) {
          await new Promise((resolve) => requestAnimationFrame(resolve));
        }
        if (hooks.gateOpens === opensBefore) return { opened: false } as any;
        const die = (window as any).__watchdogDie as any;
        const declared: number[] = [];
        die.addEventListener('will-animate', (event: Event) =>
          declared.push((event as CustomEvent).detail.expectedSettleMs));
        const faces = Array.from({ length: opts.faces }, (_, i) => i + 1);
        die.item = {
          ID: opts.id,
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 0, Value: 1, RollCount: opts.rollCount },
        };
        for (let pass = 0; pass < 4; pass++) await die.updateComplete;
        const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
        const animation = inner.getAnimations()[0];
        const duration = animation
          ? Number((animation.effect as KeyframeEffect).getTiming().duration) : 0;
        // Wait for the cycle to finish one way or the other, then read how long
        // the gate actually stayed open and whether the tumble reached its end.
        const closesBefore = hooks.gateCloses;
        while (hooks.gateCloses === closesBefore && performance.now() - start < 30000) {
          await new Promise((resolve) => requestAnimationFrame(resolve));
        }
        const log = hooks.log as { t: number; ev: string }[];
        let openAt = 0;
        let closeAt = 0;
        for (const entry of log) {
          if (entry.ev === 'gate-open') { openAt = entry.t; closeAt = 0; }
          if (entry.ev === 'gate-close' && openAt && !closeAt) closeAt = entry.t;
        }
        return {
          opened: true,
          declared,
          duration,
          gateOpenMs: closeAt - openAt,
          animationProgress: animation
            ? Number(animation.currentTime ?? 0) / Math.max(1, duration) : 0,
          playState: animation ? animation.playState : 'none',
          watchdogFirings: hooks.watchdogFirings,
        };
      }, { faces: LONG_FACES, id: LONG_ID, rollCount: LONG_ROLL_COUNT });

      expect(observed.opened).toBe(true);
      // The die really did play a throw past the floor, and really did declare
      // it. If the physics ever stops producing one this test would be vacuous,
      // so the length is asserted rather than assumed.
      expect(observed.duration,
        'the fixture seed no longer produces a roll past the 4000ms watchdog floor')
        .toBeGreaterThan(4000);
      expect(observed.declared).toEqual([observed.duration]);
      // THE ASSERTION. The gate stayed open for the tumble's own length, not
      // for the 4000ms floor. Without the extension the backstop fires at the
      // floor, the cycle is force-closed and the registry sweep force-settles
      // the die a second short of its landing.
      expect(observed.watchdogFirings).toBe(before.watchdogFirings);
      expect(observed.gateOpenMs).toBeGreaterThan(4500);
      // ...and the tumble was allowed to finish rather than being cut off.
      expect(observed.animationProgress).toBeGreaterThan(0.98);

      await page.evaluate(() => {
        (window as any).__watchdogDie?.parentElement?.remove();
        delete (window as any).__watchdogDie;
      });
    });
});

// ---------------------------------------------------------------------------
// POST-ROLL READABILITY.
//
// A die exists to be read. Every test above pins that the LANDED FACET carries
// the server's value -- a DOM fact -- and `die-shape.spec.ts` pins the pre-roll
// pose, but neither one asks the question a player asks: once the tumble stops,
// is the face carrying the value the one that reads?
//
// It was not. The pre-roll pose and the post-roll scene were two independent
// producers of a resting pose and they disagreed by 51.7 degrees: the pre-roll
// pose leaves the presented facet 22.4 degrees off the camera axis, the scene
// left it at exactly 35 and tilted the other way. On a d20, whose facets are
// only 41.8 degrees apart in normal angle, a NEIGHBOUR is then up to 27.5
// degrees more square-on than the face the player is meant to read -- measured
// over 25 seeded rolls, 24 of 25 for a d20 and 12 of 25 for a d10. Rendered at
// 300px showing 17, the "17" was small and edge-on at the top of the die while
// "18" was big and central. `aria-label` was right the whole time, which is why
// nothing caught it.
//
// So this measures the thing itself, from the render:
//
//   1. the presented facet is the MOST square-on facet on the die, by a real
//      margin, over many seeds and three shapes with very different facet
//      spacings;
//   2. the post-roll framing is the SAME framing as the pre-roll one -- a die
//      must not jump between two views of itself when a roll starts;
//   3. and the die still stops at a PHYSICAL angle: squaring the face towards
//      the camera is a re-aim of the whole scene about an axis perpendicular to
//      that face's normal, so it cannot spin the numeral upright, and the roll
//      the die comes to rest at stays whatever the simulation says.
//
// Everything is read through `getComputedStyle` after the tumble has been
// finished, so what is measured is what Chromium actually renders once the
// animation is gone (`fill: 'none'`) -- not a string the component emitted.

/**
 * Every facet's outward normal on screen after `rolls` seeded rolls, plus the
 * presented facet's content roll, measured from the composed render.
 *
 * `settle: 'roll'` throws the die and finishes the tumble; `settle: 'rest'`
 * mounts it without one, which is the pre-roll pose.
 */
async function facetAngles(
  page: import('@playwright/test').Page,
  options: { faceCount: number; rolls: number; settle: 'roll' | 'rest' },
) {
  return await page.evaluate(async (opts) => {
    await import('/src/components/boardgame-die.ts');
    document.querySelectorAll('boardgame-die').forEach((el) => el.remove());
    const die = document.createElement('boardgame-die') as any;
    die.id = 'readability-die';
    die.style.cssText = 'position:fixed;top:200px;left:200px;z-index:9999;';
    // 300px: the size the failure was rendered at, and big enough that a
    // facet's projected size is a real number of pixels.
    die.style.setProperty('--die-size', '300px');
    const faces = Array.from({ length: opts.faceCount }, (_, i) => (i + 1) * 10);
    die.item = {
      ID: 'readability-component',
      Values: { Faces: faces },
      DynamicValues: { SelectedFace: 0, Value: faces[0], RollCount: 0 },
    };
    document.body.appendChild(die);
    await die.updateComplete;
    await die.updateComplete;

    const root = die.shadowRoot as ShadowRoot;
    const stage = root.querySelector('#stage') as HTMLElement;
    // The composed transform from #stage down, exactly as the browser resolved
    // it: #inner (the tumble, or the trajectory's resting transform once the
    // tumble is gone), #orient (the pre-roll pose), and the facet's own place
    // on the solid.
    const composed = (element: HTMLElement) => {
      const chain: HTMLElement[] = [];
      for (let n: HTMLElement | null = element; n && n !== stage; n = n.parentElement) chain.unshift(n);
      let matrix = new DOMMatrix();
      for (const node of chain) {
        const value = getComputedStyle(node).transform;
        if (value && value !== 'none') matrix = matrix.multiply(new DOMMatrix(value));
      }
      return matrix;
    };
    const readPose = (presented: number) => {
      const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
      const rows = facets.map((el) => {
        const m = composed(el);
        const length = Math.hypot(m.m31, m.m32, m.m33) || 1;
        return {
          faceIndex: el.dataset.faceIndex === undefined ? -1 : Number(el.dataset.faceIndex),
          value: el.dataset.faceValue ?? null,
          // The facet's outward normal after the whole chain; its angle off the
          // camera axis (+Z) is how square-on it is.
          offAxis: (Math.acos(Math.min(1, Math.max(-1, m.m33 / length))) * 180) / Math.PI,
          // The facet's local +y on screen: how far the CONTENT is from upright.
          contentRoll: (Math.atan2(m.m21, m.m22) * 180) / Math.PI,
        };
      });
      const shown = rows.find((row) => row.faceIndex === presented)!;
      const rivals = rows.filter((row) => row !== shown);
      const best = rivals.reduce((a, b) => (b.offAxis < a.offAxis ? b : a));
      return {
        presentedOffAxis: shown.offAxis,
        presentedValue: shown.value,
        presentedContentRoll: shown.contentRoll,
        bestRivalOffAxis: best.offAxis,
        bestRivalFace: best.faceIndex,
        margin: best.offAxis - shown.offAxis,
      };
    };

    if (opts.settle === 'rest') {
      const poses: any[] = [];
      for (let face = 0; face < opts.faceCount; face++) {
        die.item = {
          ID: 'readability-component',
          Values: { Faces: faces },
          // The roll count never moves, so no throw: this is a die being SHOWN
          // a face, which is the pose `presentationTransform` owns.
          DynamicValues: { SelectedFace: face, Value: faces[face], RollCount: 0 },
        };
        for (let pass = 0; pass < 4; pass++) await die.updateComplete;
        poses.push({ face, ...readPose(face) });
      }
      die.remove();
      return poses;
    }

    const poses: any[] = [];
    for (let seed = 1; seed <= opts.rolls; seed++) {
      const selected = seed % opts.faceCount;
      die.item = {
        ID: 'readability-component',
        Values: { Faces: faces },
        DynamicValues: { SelectedFace: selected, Value: faces[selected], RollCount: seed },
      };
      for (let pass = 0; pass < 5; pass++) await die.updateComplete;
      const inner = root.querySelector('#inner') as HTMLElement;
      // Finish rather than wait: `fill: 'none'` means a finished tumble renders
      // the resting style, which is the pose under test, and 25 multi-second
      // throws per shape is not a test anyone runs.
      inner.getAnimations().forEach((animation) => animation.finish());
      await die.updateComplete;
      poses.push({
        seed,
        serverValue: faces[selected],
        ...readPose(die._roll ? die._roll.presented : -1),
      });
    }
    die.remove();
    return poses;
  }, options);
}

test.describe('boardgame-die post-roll readability', () => {
  // Three shapes chosen for how far apart their facet normals are: a d6's are
  // 90 degrees apart (which is why pig, the only shipping game with dice, never
  // showed this), a d10's 51.7 and a d20's 41.8. The tighter the spacing, the
  // less tilt a pose may spend before a neighbour wins.
  for (const faceCount of [6, 10, 20]) {
    test(`a d${faceCount} settles with the landed face the most square-on facet`, async ({ page }) => {
      test.setTimeout(120000);
      await page.goto('/');
      const rolls = await facetAngles(page, { faceCount, rolls: 25, settle: 'roll' });
      expect(rolls.length).toBe(25);

      const failures = rolls.filter((roll: any) => roll.margin <= 1);
      const worst = rolls.reduce((a: any, b: any) => (b.margin < a.margin ? b : a));
      const detail = (roll: any) =>
        `seed ${roll.seed}: presented facet ${roll.presentedOffAxis.toFixed(1)} deg off the`
        + ` camera axis, facet ${roll.bestRivalFace} only ${roll.bestRivalOffAxis.toFixed(1)}`;
      expect(failures.map(detail),
        `${failures.length}/${rolls.length} rolls left a rival facet more square-on than the`
        + ` face carrying the value; worst margin ${worst.margin.toFixed(1)} deg`)
        .toEqual([]);
      // ...and the value really was on it, so this cannot pass by presenting
      // the wrong facet very well.
      for (const roll of rolls) {
        expect(roll.presentedValue, `seed ${roll.seed} presented value`)
          .toBe(String(roll.serverValue));
      }
    });

    // One object, one framing. The two poses are produced by the same routine;
    // this is the measurement that says so, and it is what a player sees as
    // "the die did not jump when I clicked Roll".
    test(`a d${faceCount} lands in the same framing it rests in`, async ({ page }) => {
      test.setTimeout(120000);
      await page.goto('/');
      const rested = await facetAngles(page, { faceCount, rolls: 0, settle: 'rest' });
      const rolled = await facetAngles(page, { faceCount, rolls: 8, settle: 'roll' });
      const restingAngle = rested[0].presentedOffAxis;
      // Every face rests at the same angle off the camera axis: the pose is a
      // property of the SOLID, not of which face is up.
      for (const pose of rested) {
        expect(pose.presentedOffAxis, `d${faceCount} face ${pose.face} resting tilt`)
          // Two decimals, not more: Chromium serializes a computed matrix to
          // about six significant figures, which is worth 5e-4 degrees here.
          // The disagreement this exists to catch was 12.6.
          .toBeCloseTo(restingAngle, 2);
      }
      for (const roll of rolled) {
        expect(roll.presentedOffAxis,
          `d${faceCount} seed ${roll.seed} lands at a different tilt than it rests at`)
          // Two decimals, not more: Chromium serializes a computed matrix to
          // about six significant figures, which is worth 5e-4 degrees here.
          // The disagreement this exists to catch was 12.6.
          .toBeCloseTo(restingAngle, 2);
      }
    });
  }

  // The other half of the constraint. A real die stops at a random angle, and
  // squaring the landed face towards the camera must not quietly turn that into
  // a die that always stops with its numeral upright: the re-aim rotates about
  // an axis perpendicular to the presented normal, so it carries no twist about
  // that normal at all. The pre-roll pose DOES normalize the content roll -- it
  // has to read like the flat 2D die it replaces -- so the two are asserted
  // together, in opposite directions.
  test('the landed face keeps the roll the physics left it at', async ({ page }) => {
    test.setTimeout(120000);
    await page.goto('/');
    const rested = await facetAngles(page, { faceCount: 20, rolls: 0, settle: 'rest' });
    const rolled = await facetAngles(page, { faceCount: 20, rolls: 25, settle: 'roll' });
    // Pre-roll: upright, every face. (die-shape.spec.ts pins this too; repeated
    // here because it is the contrast the assertion below depends on.)
    for (const pose of rested) {
      expect(Math.abs(pose.presentedContentRoll), `resting face ${pose.face} content roll`)
        .toBeLessThan(0.5);
    }
    // Post-roll: not upright, and not clustered near upright either.
    const rolls = rolled.map((roll: any) => Math.abs(roll.presentedContentRoll));
    const upright = rolls.filter((value: number) => value < 15).length;
    expect(Math.max(...rolls),
      `post-roll content rolls: ${rolls.map((v: number) => v.toFixed(0)).join(', ')}`)
      .toBeGreaterThan(60);
    expect(upright,
      `${upright}/25 rolls stopped within 15 degrees of upright, which is not what a die does`)
      .toBeLessThan(8);
  });
});

/**
 * The duration of the die's live tumble, or 0 if it did not tumble. The die
 * lives several shadow roots down, so this walks for it rather than querying the
 * document, which sees nothing inside a shadow tree.
 *
 * Any animation on `#inner` IS the tumble: a solid die plays nothing else there
 * (the reel's scroll is the fallback for a die with no geometry). Duration is
 * deliberately NOT used as the discriminator -- trimming the trailing hold puts
 * a fast d6 roll under 400ms, well inside the version slot's own 600ms.
 */
async function liveRollDuration(page: import('@playwright/test').Page): Promise<number> {
  return await page.waitForFunction(() => {
    const find = (root: DocumentFragment | Document): any => {
      const direct = root.querySelector('boardgame-die');
      if (direct) return direct;
      for (const element of Array.from(root.querySelectorAll('*'))) {
        const shadow = (element as HTMLElement).shadowRoot;
        if (!shadow) continue;
        const found = find(shadow);
        if (found) return found;
      }
      return null;
    };
    const die = find(document);
    if (!die) return false;
    const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
    const effect = inner?.getAnimations()[0]?.effect as KeyframeEffect | undefined;
    const duration = effect ? Number(effect.getTiming().duration) : 0;
    if (duration > 0) return duration;
    const hooks = (window as any).__bgAnimTestHooks;
    return hooks && hooks.gateCloses >= hooks.gateOpens ? 0 : false;
  }, undefined, { timeout: 25000 }).then((handle) => handle.jsonValue()) as Promise<number>;
}
