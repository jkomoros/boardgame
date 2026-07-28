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
    // Registering the element is still a side effect this block wants; the
    // roll's seed, trajectory and settle-trim now live in the module that
    // owns them.
    await import('/src/components/boardgame-die.ts');
    const rollModule: any = await import('/src/motion/dice-roll.ts');
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const facesModule: any = await import('/src/motion/die-faces.ts');
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
    const seed = rollModule.dieRollSeed(opts.componentId ?? 'fixture-component', identity);
    const trajectory = rollModule.dieRollTrajectory(
      geometry, opts.componentId ?? 'fixture-component', identity);
    // What is PLAYED is the throw with its trailing dead hold cut off (see
    // settledTrajectory), so every expectation below is derived from that and
    // not from the raw trajectory -- including its duration, which is the
    // trimmed trajectory's last sample time.
    const settled = rollModule.settledTrajectory(trajectory.dice[0]);
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
        restingTransform: rollModule.rollScene(
          geometry, settled, presented, radiusPx, settledDurationMs).resting,
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
      // The chain carries the camera as a perspective() function in #inner's own
      // transform, so the composed matrix is PROJECTIVE. Every question asked of
      // it here -- which way a facet faces, which way three normals wind -- is a
      // question about the affine part, and it is also the part
      // `backface-visibility` itself reads. So the projection row is stripped
      // before anything is read out of it; leaving it in scales what comes back
      // by d / (d - z), which is 7% for a facet at the front of a resting die
      // and 13% at the top of a throw.
      const affine = (m: DOMMatrix) => {
        m.m14 = 0; m.m24 = 0; m.m34 = 0; m.m44 = 1;
        return m;
      };
    const composed = (element: HTMLElement) => {
      const chain: HTMLElement[] = [];
      for (let n: HTMLElement | null = element; n && n !== stage; n = n.parentElement) chain.unshift(n);
      let matrix = new DOMMatrix();
      for (const node of chain) {
        const value = getComputedStyle(node).transform;
        if (value && value !== 'none') matrix = matrix.multiply(new DOMMatrix(value));
      }
      return affine(matrix);
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
    // A real physics roll is a trajectory, not one 250ms animation-length slot
    // -- the assertion that the duration was not clamped into a version slot.
    expect(roll.duration).toBeGreaterThan(250);
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
  // companion cycle's slot -- 600ms, against a physics bake whose length is the
  // physics's and nobody else's (a few hundred ms for a d6, up to 2.8s for the
  // longest shapes). A clamped tumble is geometrically faithful and physically absurd
  // (a die falling at five times gravity), and the same policy can resolve to
  // SKIP outright, which makes playMotionTracks report 'not-started' and takes
  // its sibling tracks down with it. Neither is visible in solo play, because
  // the ambient animation context is null there and the policy has nothing to
  // clamp against -- so this test supplies one.
  test('a companion cycle does not clamp the roll into its slot', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      // A slot far shorter than any throw. The number is deliberately not the
      // companion cycle's real 600ms: what this pins is that an 'immediate'
      // roll is not clamped into WHATEVER slot is live, and a fixture that has
      // to be re-tuned every time the simulator is would stop pinning it.
      die.animationContext = {
        version: 12,
        startAtMs: Date.now(),
        slotDurationMs: 200,
        maxAnimationDurationMs: 200,
      };
    });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 3, stateVersion: 12 });
    // A live context, and one the kernel would otherwise have used.
    expect(roll.expected.durationMs).toBeGreaterThan(200);
    expect(roll.animations).toBe(1);
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    expect(roll.declared).toEqual([roll.expected.durationMs]);
  });

  test('every keyframe is a literal transform list, so the tumble composites', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1 });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 2, stateVersion: 7 });
    expect(roll.frames.length).toBeGreaterThan(10);
    // A var() or calc() anywhere in a transform keyframe forfeits compositing
    // and drops a multi-second tumble onto the main thread. This is exactly the
    // pattern the feature replaces, so it is asserted rather than assumed.
    expect(roll.frames.filter((value) => /var\(|calc\(/.test(value))).toEqual([]);
    // Every frame is the die's TRAVEL, then the camera, then its depth, then up
    // to two literal rotate3d turns, in front of the baked matrix3d -- all
    // literals. The order is the fix for the see-through holes: perspective()
    // projects everything to its right, so the travel in front of it moves the
    // solid AFTER it has been projected about its own centre, and the camera
    // rides with the die. (Chromium re-serializes small components in
    // exponential notation on the way back out of getKeyframes(); what was
    // WRITTEN never contains one, and it is `calc`/`var` that would cost the
    // compositor either way.)
    const shape = new RegExp(
      '^translate3d\\([-0-9.px, ]+\\)'
      + ' perspective\\([-0-9.]+px\\)'
      + ' translate3d\\([-0-9.px, ]+\\)'
      + '( rotate3d\\([-0-9., ]+deg\\))*'
      + ' matrix3d\\([-0-9., ]+\\)$');
    expect(roll.frames.filter((value) => !shape.test(value))).toEqual([]);
    // The curve's end IS the bake's own resting transform, byte for byte. This
    // is the contract that stops the die twitching when its animation is
    // removed: fill is 'none', so the element renders its resting style the
    // instant the tumble finishes.
    expect(roll.last).toBe(roll.expected.restingTransform);
  });

  // Trajectory positions are in die BOUNDING RADII; matrix3d translations are in
  // pixels. The only place that conversion happens is the `radiusPx` the die
  // measures off its own box, and getting it wrong is silent: the die still
  // tumbles, still lands on the right face, still squares up -- it just moves a
  // couple of pixels instead of across its own box, or a screenful instead. So
  // the distance the die actually covers is asserted, in units of its own size.
  test('the tumble travels a real fraction of the die across the screen', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, dieSize: '100px' });
    const roll = await rollDie(page, { faceCount: 6, selectedFace: 2, stateVersion: 21 });
    // The die's travel is the LEADING translate3d (outside the projection) plus
    // the depth term just inside it: between them they are where the solid is,
    // and the matrix3d that follows is now pure orientation.
    const translation = (value: string) => {
      const lateral = (value.match(/^translate3d\(([^)]*)\)/) as RegExpMatchArray)[1]
        .split(',').map(parseFloat);
      const depth = (value.match(/perspective\([^)]*\) translate3d\(([^)]*)\)/) as RegExpMatchArray)[1]
        .split(',').map(parseFloat);
      return [lateral[0], lateral[1], depth[2]];
    };
    const rest = translation(roll.last);
    const peak = Math.max(...roll.frames.map((frame) => {
      const at = translation(frame);
      return Math.hypot(at[0] - rest[0], at[1] - rest[1], at[2] - rest[2]);
    }));
    // The die is 100px across. A throw that covers well under half of that is
    // not a throw; one that covers three times it has left the board.
    //
    // THE FLOOR MOVED, and it is worth saying why rather than quietly fitting
    // it. It was 0.6, calibrated before the entry cap existed. `entrySimilarity`
    // now rescales the whole path so a roll enters from at most
    // MAX_ENTRY_OFFSET_DIE_WIDTHS of its own width out and from above, which
    // deliberately shortens the flight -- the old behaviour was a die appearing
    // ten frames' worth of travel away from where it lands. Measured over 60
    // seeds on a 100px d6 the peak travel is 0.50 to 1.21 die widths, median
    // 0.75; this seed is 0.54, near the low end, and the 0.6 bar had simply
    // stopped describing the shipped throw.
    //
    // 0.4 is below the measured minimum with room to spare and still kills the
    // sabotage this test exists for: leaving `radiusPx` at its default of 1
    // gives a travel under 0.03 die widths, more than an order of magnitude
    // under the bar.
    expect(peak / 100).toBeGreaterThan(0.4);
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
    // Read once the throw has SETTLED, which is where the composed matrix is a
    // rigid one. Mid-flight the die's travel sits in front of the camera in the
    // transform list, so the projection row multiplies into the 3x3 and the
    // matrix stops being a rotation -- correct on screen, where the divide by w
    // undoes it, but no longer something three normals can be read out of.
    // Handedness does not depend on the moment, so the moment chosen is the one
    // where the reading is exact.
    await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
    });
    const winding = await page.evaluate(async () => {
      const geometryModule: any = await import('/src/motion/die-geometry.ts');
      const geometry = geometryModule.dieGeometry(6);
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const stage = root.querySelector('#stage') as HTMLElement;
        // The chain carries the camera as a perspective() function in #inner's own
        // transform, so the composed matrix is PROJECTIVE. Every question asked of
        // it here -- which way a facet faces, which way three normals wind -- is a
        // question about the affine part, and it is also the part
        // `backface-visibility` itself reads. So the projection row is stripped
        // before anything is read out of it; leaving it in scales what comes back
        // by d / (d - z), which is 7% for a facet at the front of a resting die
        // and 13% at the top of a throw.
        const affine = (m: DOMMatrix) => {
          m.m14 = 0; m.m24 = 0; m.m34 = 0; m.m44 = 1;
          return m;
        };
        const composed = (element: HTMLElement) => {
          const chain: HTMLElement[] = [];
          for (let n: HTMLElement | null = element; n && n !== stage; n = n.parentElement) chain.unshift(n);
          let matrix = new DOMMatrix();
          for (const node of chain) {
            const value = getComputedStyle(node).transform;
            if (value && value !== 'none') matrix = matrix.multiply(new DOMMatrix(value));
          }
          return affine(matrix);
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
      const inner = root.querySelector('#inner') as HTMLElement;
      // The WHOLE chain, and the projection row stripped from it, so this is
      // the same number `restingRead` reports for the rolled die below. Reading
      // one of them through the camera and the other without it compares two
      // different quantities and disagrees by 7%.
      const m = new DOMMatrix(getComputedStyle(inner).transform)
        .multiply(new DOMMatrix(getComputedStyle(orient).transform))
        .multiply(new DOMMatrix(getComputedStyle(facet).transform));
      m.m14 = 0; m.m24 = 0; m.m34 = 0; m.m44 = 1;
      return m.m33;
    });
    const tilted = await page.evaluate(async () => {
      // Registering the element is still a side effect this block wants; the
      // roll's seed, trajectory and settle-trim now live in the module that
      // owns them.
      await import('/src/components/boardgame-die.ts');
      const rollModule: any = await import('/src/motion/dice-roll.ts');
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
        const trajectory = rollModule.dieRollTrajectory(geometry, 'fixture-component', rollCount);
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
    expect(early.last).toBe(early.expected.restingTransform);
    expect(late.last).toBe(late.expected.restingTransform);

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
    // Long enough that the interruption below lands mid-flight.
    expect(roll.duration).toBeGreaterThan(250);
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
    expect(roll.last).toBe(roll.expected.restingTransform);
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
    // What is played is the TRIMMED duration -- the die animates the throw's
    // motion and not whatever dead tail the simulator emitted behind it.
    expect(roll.duration).toBeCloseTo(roll.expected.durationMs, 6);
    // ...and never more than the simulator's own. How much is cut depends on
    // the simulator's rest detection and is pinned, on a trajectory built for
    // the purpose, by dice-roll.test.ts's `settledTrajectory` tests; what is
    // asserted HERE is that the component plays the trimmed throw and that the
    // cut costs nothing to look at.
    expect(roll.expected.durationMs).toBeLessThanOrEqual(roll.expected.rawDurationMs);
    // The cut costs nothing to LOOK at: the pose at the cut is the pose at the
    // end, to well under a tenth of a pixel on this 100px die. Measured through
    // the same bake the keyframes come from, so this is the rendered pose and
    // not an approximation of it.
    const drift = await page.evaluate(async (version: number) => {
      // Registering the element is still a side effect this block wants; the
      // roll's seed, trajectory and settle-trim now live in the module that
      // owns them.
      await import('/src/components/boardgame-die.ts');
      const rollModule: any = await import('/src/motion/dice-roll.ts');
      const geometryModule: any = await import('/src/motion/die-geometry.ts');
      const bakeModule: any = await import('/src/motion/dice-bake.ts');
      const geometry = geometryModule.dieGeometry(6);
      const trajectory = rollModule.dieRollTrajectory(geometry, 'fixture-component', version);
      const full = trajectory.dice[0];
      const cut = rollModule.settledTrajectory(full);
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
      expect(roll.expected.durationMs).toBeGreaterThan(250);
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
  // B2. WHAT THE DIE SAYS WHILE IT IS ROLLING, which used to be a line of red
  // text under it for the whole of every roll.
  //
  // Polled through a real pig roll: "Roll Dice is not possible right now" from
  // 78ms to 724ms -- the client displays state N while N+1 exists, so the
  // displayed state's legality snapshot says the move is not possible -- and
  // then "Wait for the current animation to finish" to 1233ms. Both are
  // legitimate transient states and neither is a failure, but the die rendered
  // EVERY reason into #action-status in --md-sys-color-error, drawn through the
  // tumbling solid.
  //
  // The fix is a severity on the reason itself (moves/action.ts), so this is
  // the browser end of the classification asserted in action.test.ts.
  test('says nothing in the error style while it rolls', async ({ page }) => {
    test.setTimeout(120000);
    await createOfflineGame(page, 'pig');
    const rollButton = page.getByRole('button', { name: /Roll die/ });
    await expect(rollButton).toBeEnabled({ timeout: 30000 });
    await rollButton.click();
    const observed = await page.evaluate(async () => {
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
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const shown: string[] = [];
      let sawTumble = false;
      const start = performance.now();
      while (performance.now() - start < 4000) {
        if (inner.getAnimations().length > 0) sawTumble = true;
        const status = root.querySelector('#action-status') as HTMLElement | null;
        const text = status?.textContent?.trim() ?? '';
        if (text) shown.push(text);
        if (sawTumble && inner.getAnimations().length === 0
          && performance.now() - start > 1500) break;
        await new Promise((resolve) => requestAnimationFrame(resolve));
      }
      await die.updateComplete;
      return {
        shown: [...new Set(shown)],
        sawTumble,
        // The title still carries the transient reason, where it costs no
        // layout and flashes nothing: quiet, not lost.
        title: (root.querySelector('#main') as HTMLElement).getAttribute('title'),
        announcement:
          (root.querySelector('.visually-hidden') as HTMLElement).textContent?.trim() ?? '',
        ariaLabel: (root.querySelector('#main') as HTMLElement).getAttribute('aria-label'),
      };
    });

    // The premise: a real roll really did run while this was polling.
    expect(observed.sawTumble).toBe(true);
    // THE ASSERTION. #action-status is the error-styled line under the die; it
    // is for a reason a person has to act on, and a die that is rolling is not
    // one.
    expect(observed.shown).toEqual([]);
    // ...and the result reached a screen reader, which an aria-label change on
    // a button does not: the label and the live region agree on the number.
    expect(observed.announcement).toMatch(/^Rolled \d+$/);
    // (The button says "Roll die showing N" while it is still rollable, and
    // "Die showing N" once it is not; what is asserted is the NUMBER agreeing.)
    expect(observed.ariaLabel).toMatch(
      new RegExp(`showing ${observed.announcement.replace('Rolled ', '')}$`));
  });

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
  // Until now nothing in a browser had ever exercised that -- the parity README
  // listed it as an accepted blind spot owned only by the gate's unit tests --
  // because no scenario ran long enough.
  //
  // A PHYSICS ROLL USED TO BE LONG ENOUGH AND IS NOT ANY MORE. The witness that
  // closed the blind spot was a d48 seed that ran to the simulator's own 5000ms
  // cap; the physics retune that followed cut every throw so far that the
  // longest roll over 4,400 seeded throws (11 shapes x 400 seeds, measured) is
  // 2761ms -- a d30 -- against the 4000ms floor. NO SEED CAN REACH IT. The test
  // did not fail when that landed, it just stopped testing anything: its own
  // guard (`toBeGreaterThan(4000)`) was the only thing that noticed.
  //
  // So the length comes from a DECLARED HOLD instead of from the tumble, which
  // is a real product path and not a test-only lever: `postAnimationDelay` is a
  // property of every animatable item, `boardgame-component-stack.ts` parses it
  // off an attribute so a game can set it, and `resolveMotionTiming` folds it
  // into the same `expectedSettleMs` the gate reads. It is also strictly better
  // suited to the job than a long roll was -- the occupancy is CHOSEN rather
  // than sampled from a distribution that a physics change can move.
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

      // The gate's own backstop floor. Not exported (`DEFAULT_FLOOR_MS` in
      // `animation-gate.ts`), so it is restated here; every number below is
      // derived from it rather than written out twice.
      const WATCHDOG_FLOOR_MS = 4000;
      // The hold the die declares after its tumble. Comfortably past the floor
      // on its own, so the test does not depend on how long the throw happens
      // to run -- which is precisely what stopped being dependable.
      const HOLD_MS = 4500;
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
        // a genuine participant in the real gate. Its OCCUPANCY, though, is
        // this test's own: the declared hold below is what carries the cycle
        // past the floor.
        const wrapper = document.createElement('div') as any;
        // pointer-events:none because this die is a d48, i.e. a BARREL, and a
        // barrel reserves a box its long axis fits in -- 2.63 die-sizes, so 263px
        // at the default. Laid over the board at (0, 0) that covers pig's own
        // "Roll die" button, and the click below would land on the fixture
        // instead. Nothing here clicks this die: it is thrown by installing an
        // item on `__watchdogDie`, so it has no use for pointer events at all.
        wrapper.style.cssText = 'position:absolute;top:0;left:0;pointer-events:none;';
        const die = document.createElement('boardgame-die') as any;
        die.id = 'watchdog-die';
        // The hold, declared the way a game declares one. It is folded into the
        // effect as `endDelay` and into `expectedSettleMs` by
        // `resolveMotionTiming`, so the animation genuinely stays live for it
        // and the die genuinely holds the gate for it -- this is occupancy, not
        // a number whispered to the watchdog.
        die.postAnimationDelay = opts.holdMs;
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
      }, { faces: LONG_FACES, id: LONG_ID, holdMs: HOLD_MS });

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
        // When the die reports itself finished, relative to the gate opening.
        // This is the PRODUCT consequence of the extension: without it the
        // cycle is force-closed at the floor and the registry sweep finishes
        // the die's animation early, so the hold it declared is cut short.
        const gateOpenedAt = performance.now();
        let settledAt = -1;
        die.addEventListener('roll-end', () => { settledAt = performance.now() - gateOpenedAt; });
        const faces = Array.from({ length: opts.faces }, (_, i) => i + 1);
        die.item = {
          ID: opts.id,
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 0, Value: 1, RollCount: opts.rollCount },
        };
        for (let pass = 0; pass < 4; pass++) await die.updateComplete;
        const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
        const animation = inner.getAnimations()[0];
        const timing = animation ? (animation.effect as KeyframeEffect).getTiming() : null;
        const duration = timing ? Number(timing.duration) : 0;
        const endDelay = timing ? Number(timing.endDelay ?? 0) : 0;
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
          endDelay,
          settledAt,
          gateOpenMs: closeAt - openAt,
          playState: animation ? animation.playState : 'none',
          watchdogFirings: hooks.watchdogFirings,
        };
      }, {
        faces: LONG_FACES, id: LONG_ID, rollCount: LONG_ROLL_COUNT, holdMs: HOLD_MS,
      });

      expect(observed.opened).toBe(true);
      // THE PREMISE, asserted rather than assumed: the die really is occupying
      // the cycle past the floor, and really did declare that. Without this the
      // test would go quietly green the way the previous version did when the
      // physics got faster -- the declaration is the only thing in the whole
      // scenario that can carry it past 4000ms.
      expect(observed.endDelay, 'the declared hold reached the effect').toBe(HOLD_MS);
      expect(observed.duration, 'and the tumble is real, not a stand-in').toBeGreaterThan(100);
      expect(observed.declared,
        'the declaration is the tumble plus the hold, and there is exactly one of it')
        .toEqual([observed.duration + HOLD_MS]);
      expect(observed.declared[0],
        'which has to be past the floor or the extension is never asked for')
        .toBeGreaterThan(WATCHDOG_FLOOR_MS);
      // THE ASSERTION. The gate stayed open for the die's own declared
      // occupancy, not for the floor. Without the extension the backstop fires
      // at the floor, the cycle is force-closed and the registry sweep
      // force-settles the die.
      expect(observed.watchdogFirings).toBe(before.watchdogFirings);
      expect(observed.gateOpenMs).toBeGreaterThan(WATCHDOG_FLOOR_MS + 400);
      // ...and the die was allowed to run its whole declared settle rather than
      // being cut off at the floor. This is what a player would see: the result
      // held for as long as the game said, instead of being swept away.
      expect(observed.settledAt, 'the die reported itself settled').toBeGreaterThan(0);
      expect(observed.settledAt,
        'and it settled on its own schedule, not at the watchdog floor')
        .toBeGreaterThan(WATCHDOG_FLOOR_MS + 200);

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
//   3. and the number the die landed on is the RIGHT WAY UP.
//
// (3) used to assert the opposite, on the grounds that a real die stops at a
// random angle. It was wrong, and looking at it is what settled it: across 8
// landed d20s showing 13, ZERO were upright and one read as "ει". The
// counter-argument is structural rather than a matter of taste -- the
// correction is a rotation ABOUT THE PRESENTED FACE'S OWN NORMAL, so it fixes
// that normal and preserves every pairwise angle between facets, which means
// (1) and (2) above hold verbatim under it and cannot be traded away for it.
// The app already normalized in its fresh-mount pose, so the old behaviour also
// meant the die visibly jumped to a tidier angle across a page reload.
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

  // THE LANDED NUMBER IS UPRIGHT, on every shape.
  //
  // The tolerance is deliberately tiny rather than generous, because the
  // correction is EXACT and not approximate: `readingPose` solves in closed
  // form for the roll, about the camera axis, that puts the landed facet's own
  // local +y on the screen's vertical, so anything left over is Chromium's own
  // serialization of the composed matrix (about six significant figures, worth
  // ~5e-4 degrees here) and nothing else. A tolerance of, say, 15 degrees would
  // pass under a correction that was merely close, and "merely close" on a d20
  // numeral is exactly the defect this replaces.
  //
  // Five shapes, not one: a d4 and a d12 present a face the naive derivation
  // happens to get right for some seeds (measured: seed 1 on a d4 came out
  // 0.005 degrees off under a wrong implementation), so a single shape or a
  // single seed is not evidence. Every face count that has ever been screenshot
  // in the critique is here, and 20 seeds each.
  for (const faceCount of [4, 6, 10, 12, 20]) {
    test(`a d${faceCount} lands with the presented face's content upright`, async ({ page }) => {
      test.setTimeout(180000);
      await page.goto('/');
      const rolled = await facetAngles(page, { faceCount, rolls: 20, settle: 'roll' });
      expect(rolled.length).toBe(20);
      const rolls = rolled.map((roll: any) => Math.abs(roll.presentedContentRoll));
      const detail = `d${faceCount} residual content roll per seed (deg): `
        + rolls.map((v: number) => v.toFixed(4)).join(', ');
      expect(Math.max(...rolls), detail).toBeLessThan(0.01);
      // ...and the value really was on the facet that got straightened, so this
      // cannot pass by measuring some other face very carefully.
      for (const roll of rolled) {
        expect(roll.presentedValue, `d${faceCount} seed ${roll.seed} presented value`)
          .toBe(String(roll.serverValue));
      }
    });
  }

  // The contrast the correction has to preserve: the pre-roll pose was ALREADY
  // upright, and a landed die must now agree with it rather than jump to a
  // tidier angle when the page is reloaded.
  test('a die that has never rolled was already upright, on every face', async ({ page }) => {
    test.setTimeout(120000);
    await page.goto('/');
    const rested = await facetAngles(page, { faceCount: 20, rolls: 0, settle: 'rest' });
    for (const pose of rested) {
      expect(Math.abs(pose.presentedContentRoll), `resting face ${pose.face} content roll`)
        .toBeLessThan(0.5);
    }
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

// ---------------------------------------------------------------------------
// WHAT A ROLL LEAVES BEHIND.
//
// Everything above is about the tumble itself. This block is about the four
// ways a roll ends up describing a die that is not on screen: a pose left on
// #inner by a roll that is over, a roll that could not be played at all, an
// item swapped for a different component's, and a component that has been
// sanitized away entirely. Each one was found by measurement, and each one ends
// with the die drawing one number while announcing another -- which is the worst
// thing this component can do, because the announcement is the only part a
// player cannot check.
test.describe('boardgame-die roll aftermath', () => {
  /**
   * A die's pose, and what it says about itself, read together.
   *
   * `front` is the facet the RENDER puts nearest the camera, so "what is drawn"
   * and "what is announced" can be compared without trusting either one.
   */
  async function poseAndLabel(page: import('@playwright/test').Page) {
    return await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const stage = root.querySelector('#stage') as HTMLElement;
      const inner = root.querySelector('#inner') as HTMLElement;
      const composed = (element: HTMLElement) => {
        const chain: HTMLElement[] = [];
        for (let n: HTMLElement | null = element; n && n !== stage; n = n.parentElement) chain.unshift(n);
        let matrix = new DOMMatrix();
        for (const node of chain) {
          const value = getComputedStyle(node).transform;
          if (value && value !== 'none') matrix = matrix.multiply(new DOMMatrix(value));
        }
        matrix.m14 = 0; matrix.m24 = 0; matrix.m34 = 0; matrix.m44 = 1;
        return matrix;
      };
      const facets = Array.from(root.querySelectorAll('.facet[data-face-index]')) as HTMLElement[];
      const scored = facets.map((el) => {
        const m = composed(el);
        return {
          faceIndex: Number(el.dataset.faceIndex),
          value: el.dataset.faceValue,
          towardsCamera: m.m33,
        };
      });
      const front = scored.reduce((best, row) => (row.towardsCamera > best.towardsCamera ? row : best));
      // How far the solid's own centre sits from the centre of the box it is
      // laid out in. A die that is showing nothing but a stale pose is outside
      // its own slot; a die that is where it belongs is at zero.
      const centre = composed(inner);
      return {
        front,
        offsetPx: Math.hypot(centre.m41, centre.m42),
        innerInline: inner.style.transform,
        ariaLabel: (root.querySelector('#main') as HTMLElement).getAttribute('aria-label'),
        values: facets.map((el) => el.dataset.faceValue),
        announcement:
          (root.querySelector('.visually-hidden') as HTMLElement)?.textContent?.trim() ?? '',
      };
    });
  }

  /** Install an item and let the die's two-pass roll planning run. */
  async function install(
    page: import('@playwright/test').Page,
    item: { id: string; faces: number[]; selectedFace: number; rollCount: number },
  ) {
    await page.evaluate(async (opts) => {
      const die = document.getElementById('fixture-die') as any;
      die.item = {
        ID: opts.id,
        Values: { Faces: opts.faces },
        DynamicValues: {
          SelectedFace: opts.selectedFace,
          Value: opts.faces[opts.selectedFace],
          RollCount: opts.rollCount,
        },
      };
      for (let pass = 0; pass < 4; pass++) await die.updateComplete;
    }, item);
  }

  async function settle(page: import('@playwright/test').Page) {
    await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
      await die.updateComplete;
    });
  }

  // B3. #orient is a CHILD of #inner, so the resting pose and the physics pose
  // COMPOSE; they are mutually exclusive only because _renderSolid renders
  // #orient as 'none' whenever a roll is set. Dropping the roll without clearing
  // #inner's inline transform therefore leaves the die posed by a throw it is no
  // longer showing -- measured at 60 to 106px outside its own 100px slot.
  test('a roll that cannot be planned clears the pose the last one left', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, rollCount: 0 });
    await install(page, { id: 'fixture-component', faces: [10, 20, 30, 40, 50, 60], selectedFace: 3, rollCount: 1 });
    await settle(page);
    const rolled = await poseAndLabel(page);
    // The premise: the throw really did leave a pose behind on #inner.
    expect(rolled.innerInline).not.toBe('');

    // A roll that cannot be planned. The die's size is what the trajectory is
    // scaled by, and a die with no measurable size has no plannable roll --
    // which is the same door a simulator or bake failure comes through.
    await page.evaluate(() => {
      (document.getElementById('fixture-die') as HTMLElement).style.setProperty('--die-size', '0px');
    });
    await install(page, { id: 'fixture-component', faces: [10, 20, 30, 40, 50, 60], selectedFace: 1, rollCount: 2 });
    const after = await poseAndLabel(page);
    expect(after.innerInline).toBe('');

    // ...and with the size restored the die is back inside its own box, showing
    // the face the server selected through the presentation pose.
    await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      die.style.setProperty('--die-size', '100px');
      await die.updateComplete;
    });
    const restored = await poseAndLabel(page);
    expect(restored.offsetPx).toBeLessThan(1);
    expect(restored.front.faceIndex).toBe(1);
    expect(restored.ariaLabel).toBe('Die showing 20');
  });

  test('a die given a different shape drops the old roll and its pose', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, rollCount: 0 });
    await install(page, { id: 'fixture-component', faces: [10, 20, 30, 40, 50, 60], selectedFace: 2, rollCount: 1 });
    await settle(page);
    expect((await poseAndLabel(page)).innerInline).not.toBe('');
    // The same element, a die with a different face count. The old roll's
    // assignment and pose are for a solid this one is not.
    await install(page, {
      id: 'fixture-component',
      faces: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
      selectedFace: 4,
      rollCount: 1,
    });
    const after = await poseAndLabel(page);
    expect(after.innerInline).toBe('');
    expect(after.offsetPx).toBeLessThan(1);
    expect(after.values.length).toBe(12);
    expect(after.front.faceIndex).toBe(4);
    expect(after.ariaLabel).toBe('Die showing 5');
  });

  // B4. playMotionTracks has three early returns that fire BEFORE it writes the
  // track's resting style, and the roll used to discard its result. In all
  // three the roll is already set, so #orient renders 'none' and #inner gets
  // nothing: the die draws in its RAW BODY FRAME, showing whichever facet
  // happens to point at the camera, while announcing the value the physics
  // landed.
  test('a roll whose playback never starts still lands where it should', async ({ page }) => {
    await mountDie(page, { faceCount: 20, selectedFace: 0, rollCount: 0 });
    // noAnimate is the one route that reaches those returns today: play()
    // returns null, and playMotionTracks reports 'not-started'.
    await page.evaluate(() => {
      (document.getElementById('fixture-die') as any).noAnimate = true;
    });
    await install(page, {
      id: 'fixture-component',
      faces: Array.from({ length: 20 }, (_, i) => (i + 1) * 10),
      selectedFace: 7,
      rollCount: 1,
    });
    const after = await poseAndLabel(page);
    // Nothing animated...
    const animations = await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      return (die.shadowRoot as ShadowRoot).querySelector('#inner')!.getAnimations().length;
    });
    expect(animations).toBe(0);
    // ...and the die is in the roll's own resting pose, not its body frame: the
    // facet nearest the camera is the one the physics landed, it carries the
    // server's value, it is readable, and it is what the die announces.
    expect(after.innerInline).not.toBe('');
    expect(after.front.value).toBe('80');
    expect(after.front.towardsCamera).toBeGreaterThan(0.7);
    expect(after.ariaLabel).toBe('Die showing 80');
  });

  // B5. The roll was reset only when the face COUNT changed, so swapping a d6
  // for a DIFFERENT d6 left the die drawing -- and announcing -- the previous
  // component's numbers. It self-corrected only if a throw arrived, and two
  // dice that have never been thrown are both at roll count 0.
  test('a die swapped for another of the same shape shows the new die', async ({ page }) => {
    await mountDie(page, {
      faceCount: 6, faces: [1, 2, 3, 4, 5, 6], selectedFace: 0, rollCount: 0, componentId: 'die-a',
    });
    await install(page, { id: 'die-a', faces: [1, 2, 3, 4, 5, 6], selectedFace: 5, rollCount: 1 });
    await settle(page);
    const first = await poseAndLabel(page);
    // The premise: die A really was THROWN, so it is carrying a roll of its own
    // -- an assignment and a pose -- for the swap below to have something to
    // leave behind.
    expect(first.innerInline).not.toBe('');
    expect(first.front.value).toBe('6');
    expect(first.ariaLabel).toBe('Die showing 6');

    // A different component, same shape, same roll count: not a throw, a
    // different die.
    await install(page, {
      id: 'die-b', faces: [10, 20, 30, 40, 50, 60], selectedFace: 1, rollCount: 0,
    });
    const second = await poseAndLabel(page);
    expect(second.values.map(Number).sort((a, b) => a - b)).toEqual([10, 20, 30, 40, 50, 60]);
    expect(second.innerInline).toBe('');
    expect(second.offsetPx).toBeLessThan(1);
    expect(second.front.faceIndex).toBe(1);
    expect(second.front.value).toBe('20');
    expect(second.ariaLabel).toBe('Die showing 20');
  });

  // B5, the other half. A game that wants to celebrate a number had to do it
  // from effectsForTransition at CYCLE START -- an effect fired at the die's
  // layout anchor while the solid is 60px away in the air, finishing on a 600ms
  // slot while the roll still has a second to run.
  test('dispatches roll-start and roll-end around the tumble', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, rollCount: 0 });
    const observed = await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const log: any[] = [];
      // Composed and bubbling: a game listens on an ancestor, across the shadow
      // boundary the die lives behind.
      const host = document.createElement('div');
      die.parentElement.insertBefore(host, die);
      host.appendChild(die);
      for (const name of ['roll-start', 'roll-end']) {
        host.addEventListener(name, (event: Event) => {
          const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
          log.push({
            name,
            detail: (event as CustomEvent).detail,
            animating: inner.getAnimations().length,
            announcement:
              (die.shadowRoot.querySelector('.visually-hidden') as HTMLElement).textContent.trim(),
          });
        });
      }
      const faces = [10, 20, 30, 40, 50, 60];
      die.item = {
        ID: 'fixture-component',
        Values: { Faces: faces },
        DynamicValues: { SelectedFace: 4, Value: faces[4], RollCount: 1 },
      };
      for (let pass = 0; pass < 4; pass++) await die.updateComplete;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      const startedWith = log.map((entry) => entry.name);
      await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
      await die.updateComplete;
      return {
        log,
        startedWith,
        announcement:
          (die.shadowRoot.querySelector('.visually-hidden') as HTMLElement).textContent.trim(),
      };
    });

    // roll-start fires as the tumble begins, roll-end only once it has stopped.
    expect(observed.startedWith).toEqual(['roll-start']);
    expect(observed.log.map((entry: any) => entry.name)).toEqual(['roll-start', 'roll-end']);
    const [start, end] = observed.log;
    expect(start.animating).toBe(1);
    expect(end.animating).toBe(0);
    // Both carry what a game needs to celebrate the number: the value on the
    // landed face, which face that is, whether the throw was cocked, and how
    // long the tumble runs.
    expect(start.detail.value).toBe(50);
    expect(start.detail.faceIndex).toBeGreaterThanOrEqual(0);
    expect(start.detail.cocked).toBe(false);
    expect(start.detail.durationMs).toBeGreaterThan(100);
    expect(end.detail).toEqual(start.detail);

    // THE ANNOUNCEMENT. A button's aria-label changing is not announced, so the
    // result never reached a screen reader; the live region is what carries it,
    // and it must not narrate the tumble on its way past.
    expect(start.announcement).toBe('');
    expect(observed.announcement).toBe('Rolled 50');
  });

  // THE LANDING BEAT.
  //
  // Nothing used to mark the result arriving. The tail of a throw decelerates,
  // so there is no frame a player can point at as the one it stopped on -- the
  // number simply becomes readable at some point, and a player who looked away
  // for a second cannot tell whether it is still going. A short pop at the
  // instant of settlement fixes that, and the critique rated it worth more than
  // making the tumble itself bouncier.
  //
  // Three things are pinned, and the last two are the ones that make it safe:
  // it happens WHEN THE ROLL ENDS (not at cycle start, where pig's own
  // celebration used to fire while the die was still 60px away in the air), it
  // is NOT a gate participant, and it is on #stage -- above the whole 3D scene,
  // so it composes with neither the tumble (#inner) nor the pose (#orient).
  test('marks the landing with a short accent that the gate never waits for',
    async ({ page }) => {
      await mountDie(page, { faceCount: 6, selectedFace: 0, rollCount: 0 });
      const observed = await page.evaluate(async () => {
        const die = document.getElementById('fixture-die') as any;
        const root = die.shadowRoot as ShadowRoot;
        const stage = root.querySelector('#stage') as HTMLElement;
        const inner = root.querySelector('#inner') as HTMLElement;
        const declared: number[] = [];
        die.addEventListener('will-animate', (event: Event) =>
          declared.push((event as CustomEvent).detail.expectedSettleMs));
        const snap = (label: string) => {
          const animation = stage.getAnimations()[0];
          const effect = animation ? (animation.effect as KeyframeEffect) : null;
          return {
            label,
            accents: stage.getAnimations().length,
            declared: declared.length,
            duration: effect ? Number(effect.getTiming().duration) : 0,
            fill: effect ? effect.getTiming().fill : null,
            frames: effect ? effect.getKeyframes().map((k: any) => k.transform) : [],
          };
        };
        const log: any[] = [];
        die.addEventListener('roll-start', () => log.push(snap('roll-start')));
        die.addEventListener('roll-end', () => log.push(snap('roll-end')));
        const faces = [10, 20, 30, 40, 50, 60];
        die.item = {
          ID: 'fixture-component',
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 4, Value: faces[4], RollCount: 1 },
        };
        for (let pass = 0; pass < 4; pass++) await die.updateComplete;
        await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
        await die.updateComplete;
        // Once the accent is over, #stage must be back to carrying nothing:
        // fill 'none', the same contract every other play in this component
        // holds to.
        const accent = stage.getAnimations()[0];
        if (accent) await accent.finished.catch(() => undefined);
        return { log, settled: getComputedStyle(stage).transform };
      });

      const start = observed.log.find((entry: any) => entry.label === 'roll-start');
      const end = observed.log.find((entry: any) => entry.label === 'roll-end');
      // WHEN. Nothing on #stage while the die is in the air; one thing the
      // instant it stops.
      expect(start.accents, 'nothing marks a roll that has only just started').toBe(0);
      expect(end.accents, 'the landing is marked as the die stops').toBe(1);
      // WHAT. A real pop -- a transform curve with an overshoot -- and not a
      // zero-length placeholder.
      expect(end.duration).toBeGreaterThan(100);
      expect(end.duration).toBeLessThan(500);
      expect(end.fill, 'the accent leaves #stage in its own resting style').toBe('none');
      const scales = end.frames.map((value: string) => Number(/scale\(([\d.]+)\)/.exec(value)![1]));
      expect(scales[0], 'starts where the die is').toBe(1);
      expect(scales[scales.length - 1], 'and ends there').toBe(1);
      expect(Math.max(...scales), 'with a visible overshoot in between').toBeGreaterThan(1.02);
      // NOT A GATE PARTICIPANT. The roll it punctuates has already settled;
      // holding the cycle open for a flourish would delay every other player's
      // board. Exactly one declaration was made for this throw -- the tumble's
      // -- and the accent added none.
      expect(start.declared, 'the tumble declares itself').toBe(1);
      expect(end.declared, 'and the accent declares nothing').toBe(1);
      expect(observed.settled, '#stage carries no transform once it is over').toBe('none');
    });

  test('runs no landing accent under reduced motion', async ({ browser }) => {
    const context = await browser.newContext({ reducedMotion: 'reduce' });
    const page = await context.newPage();
    try {
      await mountDie(page, { faceCount: 6, selectedFace: 0, rollCount: 0 });
      const accents = await page.evaluate(async () => {
        const die = document.getElementById('fixture-die') as any;
        const root = die.shadowRoot as ShadowRoot;
        const stage = root.querySelector('#stage') as HTMLElement;
        let seen = -1;
        die.addEventListener('roll-end', () => { seen = stage.getAnimations().length; });
        const faces = [10, 20, 30, 40, 50, 60];
        die.item = {
          ID: 'fixture-component',
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 4, Value: faces[4], RollCount: 1 },
        };
        for (let pass = 0; pass < 6; pass++) await die.updateComplete;
        return seen;
      });
      // The roll really did end (a -1 would mean the event never fired and the
      // assertion below would be vacuous), and nothing popped.
      expect(accents, 'the roll ended').not.toBe(-1);
      expect(accents, 'and reduced motion gets no flourish').toBe(0);
    } finally {
      await context.close();
    }
  });

  // B6. A die in a hidden stack is not absent: selectors.ts renders an occupied
  // but unreadable slot as {}, a TRUTHY object with no Values. Reaching for
  // Values.Faces threw a TypeError out of updated(), which Lit does not catch --
  // updateComplete rejects and the page takes an unhandled rejection on every
  // update from then on.
  test('a sanitized component renders an empty die instead of throwing', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 2, rollCount: 0 });
    const result = await page.evaluate(async () => {
      const errors: string[] = [];
      const onRejection = (event: PromiseRejectionEvent) => errors.push(String(event.reason));
      const onError = (event: ErrorEvent) => errors.push(String(event.message));
      window.addEventListener('unhandledrejection', onRejection);
      window.addEventListener('error', onError);
      const die = document.getElementById('fixture-die') as any;
      // What sanitization sends for a component nobody may look at.
      die.item = {};
      let settled = 'resolved';
      try {
        for (let pass = 0; pass < 3; pass++) await die.updateComplete;
      } catch (error) {
        settled = `rejected: ${String(error)}`;
      }
      // ...and a further update, because the reported failure was one rejection
      // PER UPDATE for as long as the die stayed on screen.
      die.requestUpdate();
      try {
        await die.updateComplete;
      } catch (error) {
        settled = `rejected: ${String(error)}`;
      }
      await new Promise((resolve) => setTimeout(resolve, 50));
      window.removeEventListener('unhandledrejection', onRejection);
      window.removeEventListener('error', onError);
      const root = die.shadowRoot as ShadowRoot;
      return {
        settled,
        errors,
        facets: root.querySelectorAll('.facet').length,
        reelFaces: root.querySelectorAll('#inner.reel .face').length,
        ariaLabel: (root.querySelector('#main') as HTMLElement).getAttribute('aria-label'),
      };
    });
    expect(result.settled).toBe('resolved');
    expect(result.errors).toEqual([]);
    // Nothing is known about the die, so nothing is drawn and nothing is
    // announced beyond what it is.
    expect(result.facets).toBe(0);
    expect(result.reelFaces).toBe(0);
    expect(result.ariaLabel).toBe('Die');
  });
});

// ---------------------------------------------------------------------------
// A LANDED DIE STILL READS AS A SOLID.
//
// The resting pose has a fix for a solid whose other facets are too far round
// to be seen -- `companionTilt`, which leans the pose until the most face-on of
// them reaches `COMPANION_VIEW_LIMIT` (75 degrees off the camera axis) and not
// one degree further. It is generic, not face-count-specific, and on a d4 --
// whose other three normals are 109.47 degrees away, past where
// `backface-visibility: hidden` culls them -- it is the difference between a
// flat triangle and a tetrahedron.
//
// It did not survive LANDING. The landed numeral used to be straightened by a
// separate turn about the presented facet's OWN normal, emitted on `#orient`
// INSIDE the scene's pose. That turn fixes the presented normal and preserves
// every pairwise angle, so it provably could not cost the presented facet its
// dominance -- and there was a test asserting exactly that, passing. What it
// does move is every other facet's DEPTH, and measured analytically over 12
// seeded landings it carried the companion the tilt had just placed at exactly
// 75.0 degrees to anywhere between 71.0 and 88.4, leaving the second facet a
// hairline in 3 of them.
//
// The correction is a roll of the PICTURE about the camera axis now, from the
// same `readingPose` a die that has never rolled uses, so it cannot: a rotation
// about +Z leaves every direction's z alone.
//
// Measured here from the RENDER, by hit-testing a grid over the die -- not from
// normals, which is what the analytic argument above already covers, and not
// from the animation's frames, which is a different question (see
// `die-shape.spec.ts`'s hull-deficit test). The pose under test is a static
// state: `fill: 'none'` means a finished tumble renders the resting style.

/**
 * Which facets a LANDED die shows, over `rolls` seeded throws, by share of the
 * die's own hit area.
 *
 * Counted by share and not by "was hit at all": a facet a few degrees off
 * edge-on is a sliver a pixel or two wide, which is not a facet a player can
 * see. `minShare` is the same 0.05 `die-shape.spec.ts` uses for the resting
 * pose, so the two poses are judged by one standard.
 */
async function landedVisibleFacets(
  page: import('@playwright/test').Page,
  options: { faceCount: number; rolls: number; minShare: number },
) {
  return await page.evaluate(async (opts) => {
    await import('/src/components/boardgame-die.ts');
    document.querySelectorAll('boardgame-die').forEach((el) => el.remove());
    const die = document.createElement('boardgame-die') as any;
    die.id = 'landed-visibility-die';
    die.style.cssText = 'position:fixed;top:200px;left:200px;z-index:9999;';
    // 200px, the size the resting-pose measurement uses: big enough that a
    // facet's projected size is a real number of hit-test cells.
    die.style.setProperty('--die-size', '200px');
    const faces = Array.from({ length: opts.faceCount }, (_, i) => (i + 1) * 10);
    die.item = {
      ID: 'landed-visibility-component',
      Values: { Faces: faces },
      DynamicValues: { SelectedFace: 0, Value: faces[0], RollCount: 0 },
    };
    document.body.appendChild(die);
    await die.updateComplete;
    await die.updateComplete;
    const root = die.shadowRoot as ShadowRoot;

    const sample = (presented: number) => {
      const box = (root.querySelector('#main') as HTMLElement).getBoundingClientRect();
      const hits = new Map<string, number>();
      const N = 60;
      for (let i = 0; i < N; i++) {
        for (let j = 0; j < N; j++) {
          const hit = (root as any).elementFromPoint(
            box.left + ((i + 0.5) / N) * box.width,
            box.top + ((j + 0.5) / N) * box.height,
          ) as Element | null;
          const facet = hit?.closest?.('.facet') as HTMLElement | null;
          if (!facet) continue;
          const key = facet.dataset.faceIndex
            ?? `cap${Array.from(root.querySelectorAll('.facet')).indexOf(facet)}`;
          hits.set(key, (hits.get(key) ?? 0) + 1);
        }
      }
      const total = [...hits.values()].reduce((a, b) => a + b, 0);
      return {
        total,
        distinctFacetsVisible:
          [...hits.values()].filter((value) => value / total >= opts.minShare).length,
        presentedShare: total ? (hits.get(String(presented)) ?? 0) / total : 0,
        shares: Object.fromEntries(
          [...hits.entries()].map(([key, value]) => [key, Number((value / total).toFixed(3))]),
        ),
      };
    };

    const results: any[] = [];
    for (let seed = 1; seed <= opts.rolls; seed++) {
      const selected = seed % opts.faceCount;
      die.item = {
        ID: 'landed-visibility-component',
        Values: { Faces: faces },
        DynamicValues: { SelectedFace: selected, Value: faces[selected], RollCount: seed },
      };
      for (let pass = 0; pass < 5; pass++) await die.updateComplete;
      const inner = root.querySelector('#inner') as HTMLElement;
      // Finish rather than wait: `fill: 'none'` renders the resting style the
      // instant the tumble is done, and that resting style IS the pose here.
      inner.getAnimations().forEach((animation) => animation.finish());
      await die.updateComplete;
      const presented = die._roll ? die._roll.presented : -1;
      const facet = root.querySelector(`.facet[data-face-index="${presented}"]`) as HTMLElement;
      results.push({
        seed,
        presented,
        serverValue: faces[selected],
        presentedValue: facet?.dataset.faceValue ?? null,
        ...sample(presented),
      });
    }
    die.remove();
    return results;
  }, options);
}

test.describe('a landed die reads as a solid', () => {
  // Every shape the critique screenshotted. The d4 is the one the companion
  // tilt exists for; the other four are the controls that say the fix did not
  // buy the tetrahedron anything at the rest of the set's expense.
  for (const faceCount of [4, 6, 10, 12, 20]) {
    test(`a landed d${faceCount} shows more than one facet`, async ({ page }) => {
      test.setTimeout(180000);
      await page.goto('/');
      const landed = await landedVisibleFacets(page, { faceCount, rolls: 8, minShare: 0.05 });
      expect(landed.length).toBe(8);

      // The premise: these really are landed dice, showing the value the
      // server chose. A measurement of the wrong facet proves nothing.
      for (const pose of landed) {
        expect(pose.presented, `d${faceCount} seed ${pose.seed} did not land`)
          .toBeGreaterThanOrEqual(0);
        expect(pose.presentedValue, `d${faceCount} seed ${pose.seed} presented value`)
          .toBe(String(pose.serverValue));
        expect(pose.total, `d${faceCount} seed ${pose.seed} hit nothing`).toBeGreaterThan(100);
      }

      const flat = landed.filter((pose: any) => pose.distinctFacetsVisible < 2);
      expect(flat.map((pose: any) =>
        `seed ${pose.seed}: ${JSON.stringify(pose.shares)}`),
        `${flat.length}/${landed.length} landed d${faceCount} poses read as a single flat polygon`)
        .toEqual([]);

      // ...and the face carrying the value is still the dominant one, which is
      // what stops "show a second facet" being satisfied by drowning the first.
      // Asserted for the d4 only, for the same reason the resting-pose test
      // does: it is the shape the tilt touches, and its facets are big enough
      // that Chromium's hit testing of a preserve-3d subtree is dependable.
      if (faceCount === 4) {
        for (const pose of landed) {
          expect(pose.presentedShare,
            `d4 seed ${pose.seed} presented share, shares ${JSON.stringify(pose.shares)}`)
            .toBeGreaterThan(0.5);
          expect(pose.presentedShare,
            `d4 seed ${pose.seed} presented share, shares ${JSON.stringify(pose.shares)}`)
            .toBeLessThan(0.95);
        }
      }
    });
  }
});
