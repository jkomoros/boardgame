import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, expectCleanGate } from '../helpers';

// Task 10: <boardgame-die> stops pretending to roll and actually rolls.
//
// A face change now derives a seed from (component ID, state version), runs
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
//      so the same (id, version) has to reproduce the same tumble bit for bit,
//      and a different version has to produce a different one.
//   4. Reduced motion: no tumble, right face.
//   5. Interruption: `finishGatedAnimations()` mid-roll leaves the right face
//      presented. This is the `fill: 'none'` resting-style contract.
//   6. The watchdog does not fire for a multi-second roll, in the real app.
//   7. The face labels swap at the START of the roll (see the label-swap
//      section below).
//   8. Every keyframe is a literal transform list -- no `var()`, no `calc()` --
//      because either one silently demotes the tumble off the compositor.



interface MountOptions {
  faceCount: number;
  selectedFace?: number;
  faces?: number[];
  componentId?: string;
  stateVersion?: number;
  dieSize?: string;
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
    die.item = {
      ID: opts.componentId ?? 'fixture-component',
      Values: { Faces: faces },
      DynamicValues: { SelectedFace: opts.selectedFace ?? 0, Value: faces[opts.selectedFace ?? 0] },
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
    durationMs: number;
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
  options: { selectedFace: number; stateVersion?: number; faceCount: number; componentId?: string },
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
    die.item = {
      ID: opts.componentId ?? 'fixture-component',
      Values: { Faces: faces },
      DynamicValues: { SelectedFace: opts.selectedFace, Value: faces[opts.selectedFace] },
    };
    // Three passes, all inside the same frame: the install writes the face, the
    // face change plans the roll, and the pass after that -- the first one whose
    // render carries the roll's own face values -- plays it.
    for (let pass = 0; pass < 4; pass++) await die.updateComplete;
    die.removeEventListener('will-animate', onWillAnimate);
    Element.prototype.animate = originalAnimate;

    const geometry = geometryModule.dieGeometry(opts.faceCount);
    const version = opts.stateVersion ?? die.stateVersion ?? 0;
    const seed = dieModule.dieRollSeed(opts.componentId ?? 'fixture-component', version);
    const trajectory = dieModule.dieRollTrajectory(
      geometry, opts.componentId ?? 'fixture-component', version);
    const presented = facesModule.presentedFaceIndex(
      geometry, trajectory.dice[0].restingOrientation);

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
        durationMs: trajectory.durationMs,
        presented,
        restingTransform: bakeModule.restingTransform(trajectory.dice[0], { radiusPx }),
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
  // The answer is to make it up: the scene squares the landed face onto the
  // reading axis, always, so the face carrying the value is exactly level
  // whether the throw settled at 0.0 degrees (nearly all of them) or at 2.4
  // (measured worst for a d12 over 40 throws) or at the 3-plus that counts as
  // cocked. There is no floor drawn, so a couple of degrees of world tilt is
  // invisible, and the alternative -- refusing to animate a rare roll -- trades
  // an unreadable die for a die that sometimes does not move.
  //
  // Pinned on a d12, whose throws are the ones that actually land tilted, at the
  // first state version that produces one.
  test('lands the reading face exactly square, even when the throw did not', async ({ page }) => {
    await mountDie(page, { faceCount: 12, selectedFace: 0, stateVersion: 1, dieSize: '160px' });
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
      for (let version = 2; version < 80; version++) {
        const trajectory = dieModule.dieRollTrajectory(geometry, 'fixture-component', version);
        const die = trajectory.dice[0];
        const presented = facesModule.presentedFaceIndex(geometry, die.restingOrientation);
        const world = rotate(
          die.restingOrientation as number[], geometry.faces[presented].normal as number[]);
        const degrees = (Math.acos(Math.min(1, Math.max(-1, world[1]))) * 180) / Math.PI;
        if (degrees > 0.5) return { version, degrees };
      }
      return null;
    });
    // If this ever comes back null the fixture has stopped exercising the case
    // it exists for, which is a louder failure than a silently vacuous pass.
    expect(tilted, 'no d12 throw in the first 78 versions landed off-square').not.toBeNull();
    expect(tilted!.degrees).toBeGreaterThan(0.5);

    const roll = await rollDie(page, {
      faceCount: 12, selectedFace: 7, stateVersion: tilted!.version,
    });
    await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
      await Promise.all(inner.getAnimations().map((a) => a.finished.catch(() => undefined)));
    });
    const read = await restingRead(page);
    // Read the LANDED facet by index, not the most camera-facing one: a d12's
    // neighbouring faces are only 63 degrees apart, so at this camera elevation
    // one of them is nearer face-on than the top face is. (A d6's are 90 apart,
    // which is why the tests above can use the front-most facet.)
    const landed = read.byIndex[roll.expected.presented];
    expect(landed.value).toBe(String(roll.serverValue));
    // Squared up: the reading face sits at exactly the camera's own elevation
    // off the view axis, cos(35 degrees), and NOT at cos(35 +/- the throw's own
    // tilt). The throw above is off by `tilted.degrees`, which without the
    // square-up moves this by ~0.01 -- hundreds of times the tolerance.
    expect(landed.towardsCamera).toBeCloseTo(Math.cos((35 * Math.PI) / 180), 4);
  });

  test('is deterministic in (component id, state version), and only in those', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-a' });
    const first = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, componentId: 'die-a',
    });
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-a' });
    const again = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, componentId: 'die-a',
    });
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-a' });
    const later = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 10, componentId: 'die-a',
    });
    await mountDie(page, { faceCount: 6, selectedFace: 0, stateVersion: 1, componentId: 'die-b' });
    const other = await rollDie(page, {
      faceCount: 6, selectedFace: 2, stateVersion: 9, componentId: 'die-b',
    });

    // Same identity, same roll -- the whole reason the simulation is seeded
    // rather than random: a remount mid-roll must not re-throw the die.
    expect([again.first, again.last, again.duration])
      .toEqual([first.first, first.last, first.duration]);
    // A NEW state version is a new throw, and so is a different die.
    expect(later.first).not.toBe(first.first);
    expect(other.first).not.toBe(first.first);
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
    // A roll landing on the face the die is already showing does not tumble
    // (~1 in 6; see the report), so this retries until one does. P(4 in a row)
    // is under a percent.
    let before = await gateSnapshot(page);
    let declared = 0;
    for (let attempt = 0; attempt < 4 && declared === 0; attempt++) {
      before = await gateSnapshot(page);
      const die = page.getByRole('button', { name: 'Roll die' });
      await expect(die).toBeEnabled({ timeout: 20000 });
      await die.click();
      declared = await longRollDuration(page);
      if (declared === 0) await expectCleanGate(page, before, 30000);
    }
    expect(declared).toBeGreaterThan(600);
    // The watchdog's floor is well under a physics roll, so the die's
    // will-animate declaration is the only thing keeping the cycle from being
    // force-closed mid-tumble.
    await expectCleanGate(page, before, 60000);
    const after = await gateSnapshot(page);
    expect(after.watchdogFirings).toBe(before.watchdogFirings);
  });
});

/**
 * The duration the die's live tumble was declared for, or 0 if it did not
 * tumble. The die lives several shadow roots down, so this walks for it rather
 * than querying the document, which sees nothing inside a shadow tree.
 */
async function longRollDuration(page: import('@playwright/test').Page): Promise<number> {
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
    if (!die) return 0;
    const inner = (die.shadowRoot as ShadowRoot).querySelector('#inner') as HTMLElement;
    const effect = inner?.getAnimations()[0]?.effect as KeyframeEffect | undefined;
    const duration = effect ? Number(effect.getTiming().duration) : 0;
    // A tumble is seconds long; nothing else this die plays is. Settle for 0
    // once the client has gone quiet, so a same-face roll can be retried.
    if (duration > 600) return duration;
    const hooks = (window as any).__bgAnimTestHooks;
    return hooks && hooks.gateCloses >= hooks.gateOpens ? 0 : false;
  }, undefined, { timeout: 25000 }).then((handle) => handle.jsonValue()) as Promise<number>;
}
