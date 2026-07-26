import { Page, expect } from '@playwright/test';
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

// Motion-curve parity sampling.
//
// Raw bounding-rect goldens cannot work here: which components animate, and
// how far they travel, depends on per-game randomness (shuffled component
// ids feed the messy-stack rotation hash; FLIP skips no-op transforms), so
// rects recorded from one game never match a fresh game. What IS
// deterministic — and what the CSS-transition -> WAAPI migration must
// preserve exactly — is each animation's TIMING SHAPE: pause every live
// animation (element.animate Animations and CSSTransitions both appear in
// document.getAnimations()), seek to fixed fractions of its own active
// duration, and measure displacement-normalized progress and opacity at
// each fraction. Normalized progress at a fraction is a pure fingerprint of
// easing + timing (an ease-in-out mid-point is ~0.5 regardless of whether
// the card flies 40px or 400px), so curves from different games compare
// equal exactly when the animation timing behavior is unchanged.
//
// The in-page pass does MEASUREMENT only (it must run in the browser); the
// normalization math lives in module scope below so it can be exercised
// directly by unit-style specs with synthetic sample sequences.

export interface MotionCurve {
  // Normalized bounding-rect-center progress at each sampled fraction,
  // rounded to 2dp; null when the element does not travel (pure
  // opacity/rotation). Normalized by CUMULATIVE path length, so a curved or
  // out-and-back flight still reads as monotone progress in [0,1].
  progress: (number | null)[];
  // Normalized progress through the transform matrix's LINEAR part
  // (rotation/scale/skew/perspective — every entry except the translation
  // column), by cumulative path length; null when that part does not move.
  // Catches rotations/scales/flips that move no pixels of bounding-rect
  // center (e.g. a card flip's rotateY, a die's tumble).
  rotation: (number | null)[];
  // Normalized progress through the transform matrix's TRANSLATION column,
  // by cumulative path length; null when it does not move. Split out from
  // `rotation` because the two live on incomparable scales — rotation
  // entries are unitless (a whole matrix's worth is at most 2.83) while
  // translation is raw pixels, so a single Frobenius sum over both makes a
  // 60px travel drown the rotation to ~5% of the norm and degenerates into
  // a duplicate of `progress`.
  translation: (number | null)[];
  // Opacity at each sampled fraction, rounded to 2dp; null when opacity is
  // not animated (stays within 0.01 of constant).
  opacity: (number | null)[];
  // Declared [activeDurationMs, delayMs], each rounded to a 25ms grid.
  // Normalized progress deliberately divides absolute time out of the
  // other channels; this channel puts it back (harness critic gap 1: a
  // migration that misreads --animation-length would otherwise produce
  // identical normalized curves at 4x the duration). Also pins stagger
  // cadence: staggered cohort members carry distinct delays.
  timing: [number, number];
  // Computed z-index at each fraction; null when constant for the whole
  // flight (critic gap 7: a card passing UNDER instead of over mid-flight
  // is invisible to every other channel).
  zIndex: (number | null)[];
}

export interface GeometryFingerprint {
  fractions: number[];
  // Sorted unique curves (JSON identity). The COUNT of animating elements
  // varies per game; the SET of distinct timing shapes does not.
  curves: MotionCurve[];
}

// One measurement of one animated element at one sampled fraction.
export interface MotionSample {
  // Bounding-rect center, viewport px.
  x: number;
  y: number;
  // Computed transform as a 16-entry column-major 4x4 matrix.
  matrix: number[];
  opacity: number;
  z: number;
}

// The raw per-animation measurement record the in-page pass returns; the
// normalizer below turns a list of these into MotionCurves.
export interface SampledAnimation {
  samples: MotionSample[];
  durationMs: number;
  delayMs: number;
}

const GOLDEN_DIR = join(dirname(fileURLToPath(import.meta.url)), 'goldens');
const FRACTIONS = [0, 0.25, 0.5, 0.75, 1];

const round = (v: number) => Math.round(v * 100) / 100;
const grid25 = (v: number) => Math.round(v / 25) * 25;
const euclidean = (a: number[], b: number[]): number =>
  Math.sqrt(a.reduce((acc, v, i) => acc + (v - (b[i] ?? 0)) ** 2, 0));

// Column-major 4x4 index groups. 12/13/14 are the translation column
// (tx, ty, tz); everything else is the linear part (rotation, scale, skew
// and the perspective row) — reported as the `rotation` channel.
const TRANSLATION_INDICES = [12, 13, 14];
const LINEAR_INDICES = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 15];
const pick = (m: number[], indices: number[]): number[] => indices.map((i) => m[i] ?? 0);

// Normalized CUMULATIVE path length along a sequence of points in some
// metric space: sum the step lengths and divide by the total.
//
// This replaces the old chord-from-start over net-displacement ratio, which
// silently assumed motion travels monotonically from A to B. A tumbling die
// breaks that assumption four ways: its path length far exceeds its net
// displacement (values ran well past 1 against an absolute 0.08 tolerance),
// the values were non-monotone, a landing near the start pose drove the
// denominator toward zero (exploding the ratios, or — just under the
// threshold — nulling the channel, and an all-null curve is DROPPED
// entirely), and the landing pose depends on the server-rolled value so the
// denominator moved run to run. Cumulative arc length is monotone
// non-decreasing and lands in [0,1] by construction, which kills all four.
//
// Returns null when the total path is at or below `epsilon` — i.e. this
// channel did not move at all, so it fingerprints nothing.
const cumulativeProfile = (
  points: number[][],
  epsilon: number,
): number[] | null => {
  const travelled: number[] = [0];
  let total = 0;
  for (let i = 1; i < points.length; i++) {
    total += euclidean(points[i]!, points[i - 1]!);
    travelled.push(total);
  }
  if (!(total > epsilon)) return null;
  return travelled.map((d) => round(d / total));
};

// Extracts the pure ROTATION from a column-major 4x4 transform matrix, as
// the three columns of an orthonormal 3x3 (column-major, 9 entries).
// Gram-Schmidt, i.e. the standard matrix-decomposition first step: divide
// out scale and skew so a `scale(...) rotateY(...)` composite reports the
// same orientation as the bare rotation. A degenerate (zero-scale) column
// yields identity — no orientation is recoverable from a collapsed matrix.
const rotationBasis = (m: number[]): number[] => {
  const col = (i: number): number[] => [m[i] ?? 0, m[i + 1] ?? 0, m[i + 2] ?? 0];
  const dot = (a: number[], b: number[]) => a[0]! * b[0]! + a[1]! * b[1]! + a[2]! * b[2]!;
  const sub = (a: number[], b: number[], k: number) => a.map((v, i) => v - k * b[i]!);
  const norm = (a: number[]) => Math.sqrt(dot(a, a));
  const IDENTITY = [1, 0, 0, 0, 1, 0, 0, 0, 1];
  let x = col(0), y = col(4), z = col(8);
  const sx = norm(x);
  if (!(sx > 1e-6)) return IDENTITY;
  x = x.map((v) => v / sx);
  y = sub(y, x, dot(x, y));
  const sy = norm(y);
  if (!(sy > 1e-6)) return IDENTITY;
  y = y.map((v) => v / sy);
  z = sub(sub(z, x, dot(x, z)), y, dot(y, z));
  const sz = norm(z);
  if (!(sz > 1e-6)) return IDENTITY;
  z = z.map((v) => v / sz);
  // Left-handed (mirrored) basis: flip the third axis so the result is a
  // proper rotation and the trace formula below stays valid.
  const det = x[0]! * (y[1]! * z[2]! - y[2]! * z[1]!)
    - y[0]! * (x[1]! * z[2]! - x[2]! * z[1]!)
    + z[0]! * (x[1]! * y[2]! - x[2]! * y[1]!);
  if (det < 0) z = z.map((v) => -v);
  return [...x, ...y, ...z];
};

// Angle in DEGREES between two orientations. For column-major bases A and
// B, the relative rotation is AᵀB, whose trace is just the sum of the dot
// products of corresponding columns; angle = acos((trace − 1) / 2).
const angleBetweenBases = (a: number[], b: number[]): number => {
  let trace = 0;
  for (let c = 0; c < 3; c++) {
    for (let r = 0; r < 3; r++) trace += a[c * 3 + r]! * b[c * 3 + r]!;
  }
  const cos = Math.min(1, Math.max(-1, (trace - 1) / 2));
  return (Math.acos(cos) * 180) / Math.PI;
};

// TOTAL SWEPT ROTATION, in degrees, of one animation's transform across the
// sampled fractions: accumulate the angle between successive orientations.
//
// This is the one thing the path-length normalizer structurally cannot see.
// That normalizer is magnitude-INVARIANT by construction — a 180° turn and
// a 90° turn under the same easing produce an identical normalized
// `rotation` channel — so nothing in the curve set pins how FAR a rotation
// actually turns. This does, as a plain scalar.
//
// It is deliberately NOT a fingerprint channel: the curve set is compared
// as a SET across every animating element, and debuganimations' messy-stack
// tilts are per-game RANDOM in magnitude, so a general magnitude channel
// would flake on every run. Use this only from a scenario whose rotation
// magnitude is a genuine invariant of the product (memory's reveal is
// always exactly a half turn; a fixed-seed die roll lands a fixed pose).
//
// Accumulating step-wise, rather than measuring start-to-end, is what makes
// a multi-turn tumble measurable at all: a 360° roll returns to its start
// orientation, so the start-to-end angle would read 0.
export function sweptRotationDegrees(samples: MotionSample[]): number {
  let total = 0;
  for (let i = 1; i < samples.length; i++) {
    total += angleBetweenBases(
      rotationBasis(samples[i - 1]!.matrix),
      rotationBasis(samples[i]!.matrix),
    );
  }
  return total;
}

// Swept rotation of every sampled animation, descending. A scenario picks
// the entry it means (usually the largest — the element it triggered) and
// asserts its magnitude; the sampler cannot label animations by role.
export function sweptRotationsDegrees(sampled: SampledAnimation[]): number[] {
  return sampled
    .map((e) => sweptRotationDegrees(e.samples))
    .sort((a, b) => b - a);
}

// Turns raw per-animation samples into the deduplicated curve set. Pure —
// no DOM, no page — so specs can feed it synthetic sequences directly.
export function fingerprintFromSamples(
  sampledAnimations: SampledAnimation[],
  fractions: number[],
): GeometryFingerprint {
  const curves = sampledAnimations
    .filter((e) => e.samples.length === fractions.length)
    .map((e) => {
      const s = e.samples;
      const first = s[0]!, last = s[s.length - 1]!;
      const nulls = s.map(() => null);
      // The travel gate stays on NET bounding-rect displacement (an
      // in-place tumble is not a travel animation, and widening this gate
      // to path length would newly open the channel on curves that only
      // jitter). What changed is the VALUE: cumulative path length, so a
      // curved or doubling-back flight reads monotone. Nothing is lost for
      // an in-place tumble — the `rotation`/`translation` channels below
      // are path-gated and carry it.
      const travels = Math.hypot(last.x - first.x, last.y - first.y) > 2; // px
      const progress = travels ? cumulativeProfile(s.map((p) => [p.x, p.y]), 0) : null;
      const rotation = cumulativeProfile(
        s.map((p) => pick(p.matrix, LINEAR_INDICES)), 0.01);
      const translation = cumulativeProfile(
        s.map((p) => pick(p.matrix, TRANSLATION_INDICES)), 0.01);
      const opacities = s.map((p) => p.opacity);
      const opacityAnimates = Math.max(...opacities) - Math.min(...opacities) > 0.01;
      const zs = s.map((p) => p.z);
      const zChanges = Math.max(...zs) !== Math.min(...zs);
      return {
        progress: progress ?? nulls,
        rotation: rotation ?? nulls,
        translation: translation ?? nulls,
        opacity: opacities.map((o) => (opacityAnimates ? round(o) : null)),
        timing: [grid25(e.durationMs), grid25(e.delayMs)] as [number, number],
        zIndex: zs.map((z) => (zChanges ? z : null)),
      };
    });
  // Unique by JSON identity, sorted for stable comparison. All-null
  // curves are dropped: they are sub-threshold noise (near-no-op FLIPs
  // whose presence is per-game random) and assert nothing.
  const seen = new Map<string, MotionCurve>();
  for (const c of curves) {
    // Timing/zIndex don't count toward "observable motion": a curve whose
    // visual channels are all null asserts nothing regardless of its
    // declared timing.
    const allNull = [...c.progress, ...c.rotation, ...c.translation, ...c.opacity]
      .every((v) => v === null);
    if (!allNull) seen.set(JSON.stringify(c), c);
  }
  return {
    fractions: [...fractions],
    curves: [...seen.keys()].sort().map((k) => seen.get(k)!),
  };
}

// Runs `trigger`, waits for animations to exist and their population to
// stabilize across two frames, pauses them all, seeks each through the
// sample fractions of its own (delay + activeDuration), measuring targets'
// positions and opacity, then finishes everything so the gate settles
// normally. Total paused wall-time stays well under the 4s watchdog floor.
export async function sampleRawMotion(
  page: Page,
  trigger: () => Promise<void>,
): Promise<SampledAnimation[]> {
  await trigger();
  // One atomic in-page pass: find animations, wait for their population to
  // stabilize, pause, seek, sample, finish. document.getAnimations() does
  // NOT see shadow-tree animations in this Chromium (verified empirically:
  // 0 at document level while 141 ran inside component shadow roots), so
  // every step walks the shadow trees and collects per-element
  // getAnimations() instead.
  const sampledAnimations: SampledAnimation[] = await page.evaluate(async (fractions) => {
    const deepAnimations = (): Animation[] => {
      const out: Animation[] = [];
      const walk = (root: Document | ShadowRoot) => {
        for (const el of Array.from(root.querySelectorAll('*'))) {
          const anims = (el as Element & { getAnimations?: (o?: object) => Animation[] })
            .getAnimations?.({ subtree: false });
          if (anims) out.push(...anims);
          const sr = (el as Element & { shadowRoot: ShadowRoot | null }).shadowRoot;
          if (sr) walk(sr);
        }
      };
      walk(document);
      return out;
    };
    const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));
    const targetOf = (a: Animation): HTMLElement | null => {
      const t = (a.effect as KeyframeEffect | null)?.target;
      return t instanceof HTMLElement ? t : null;
    };
    const parseMatrix = (raw: string): number[] => {
      if (!raw || raw === 'none') return [1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1];
      const nums = raw.slice(raw.indexOf('(') + 1, -1).split(',').map(Number);
      if (nums.length === 6) {
        const [a, b, c, d, e, f] = nums as [number, number, number, number, number, number];
        return [a, b, 0, 0, c, d, 0, 0, 0, 0, 1, 0, e, f, 0, 1];
      }
      return nums;
    };
    interface Sample { x: number; y: number; matrix: number[]; opacity: number; z: number }
    interface WaveEntry { samples: Sample[]; durationMs: number; delayMs: number }

    // A cycle's animations arrive in WAVES (staggered cohorts, arrival
    // decorations, departure fades chained on earlier completions). A
    // single pause-at-first-stability latches an arbitrary wave, so the
    // sampled set varies run to run. Instead: iterate — pause and sample
    // each newly appeared wave, finish it (letting the choreography chain
    // to the next wave), and accumulate the UNION of curves until no new
    // animations appear for ~1s. Requires at least one wave overall.
    const samplesAll: WaveEntry[] = [];
    const sampled = new Set<Animation>();
    const overallDeadline = performance.now() + 20000;
    let sawAny = false;
    let quietSince = performance.now();
    for (;;) {
      const now = performance.now();
      if (now > overallDeadline) {
        if (sawAny) break;
        throw new Error('no animations appeared within 20s of the trigger');
      }
      const freshCount = deepAnimations().filter((a) => !sampled.has(a)).length;
      // Quiet-window exit is checked AFTER the fresh probe (review finding:
      // checking before could skip a wave born in the final frame of the
      // quiet window and flake the existential comparison).
      if (freshCount === 0) {
        if (sawAny && now - quietSince > 1000) break;
        await frame(); continue;
      }
      // Two frames so co-scheduled stragglers of this wave are included.
      await frame(); await frame();
      const wave = deepAnimations().filter((a) => !sampled.has(a));
      if (wave.length === 0) { await frame(); continue; }
      sawAny = true;
      for (const a of wave) { sampled.add(a); a.pause(); }
      const waveEntries: WaveEntry[] = wave.map((a) => {
        const t = a.effect?.getComputedTiming();
        return {
          samples: [],
          durationMs: Number(t?.activeDuration ?? 0),
          delayMs: Number(t?.delay ?? 0),
        };
      });
      for (const frac of fractions) {
        for (const a of wave) {
          const t = a.effect?.getComputedTiming();
          const total = Number(t?.delay ?? 0) + Number(t?.activeDuration ?? 0);
          try { a.currentTime = frac * total; } catch { /* infinite or detached */ }
        }
        wave.forEach((a, i) => {
          const el = targetOf(a);
          if (!el) return;
          const r = el.getBoundingClientRect();
          const style = getComputedStyle(el);
          waveEntries[i]!.samples.push({
            x: r.x + r.width / 2,
            y: r.y + r.height / 2,
            matrix: parseMatrix(style.transform),
            opacity: parseFloat(style.opacity) || 0,
            z: parseInt(style.zIndex, 10) || 0,
          });
        });
      }
      for (const a of wave) { try { a.finish(); } catch { try { a.cancel(); } catch { /* dead */ } } }
      samplesAll.push(...waveEntries);
      quietSince = performance.now();
    }
    return samplesAll;
  }, FRACTIONS);
  return sampledAnimations;
}

// The fingerprint of an already-sampled run, at the harness's fractions.
// Split from `sampleRawMotion` so a scenario that also needs a scalar
// measurement (e.g. `sweptRotationDegrees`) can take both off ONE trigger
// rather than driving the scenario twice.
export function fingerprintOf(sampled: SampledAnimation[]): GeometryFingerprint {
  return fingerprintFromSamples(sampled, FRACTIONS);
}

export async function sampleMotionCurves(
  page: Page,
  trigger: () => Promise<void>,
): Promise<GeometryFingerprint> {
  return fingerprintOf(await sampleRawMotion(page, trigger));
}

// Compares (or with PARITY_RECORD=1, rewrites) the golden. Each golden
// curve must be matched by some observed curve within tolerance, and vice
// versa — set equivalence under a per-sample numeric tolerance, so minor
// sub-pixel jitter cannot flake while an easing/duration/keyframe change
// (which shifts mid-fraction progress by far more than the tolerance)
// always fails.
//
// Deliberate non-goal: HOW MANY elements animate. Set semantics mean one
// member of a near-duplicate curve family disappearing is invisible here
// (element counts are per-game random). Count regressions are owned by the
// trace suite: memory's exact gateDelta.plays/settles and per-element event
// sequences, and debuganimations' exact cycle counts.
export function expectCurvesMatchGolden(
  actual: GeometryFingerprint,
  name: string,
  tolerance = 0.08,
): void {
  const goldenPath = join(GOLDEN_DIR, `${name}.json`);
  expect(actual.curves.length, 'scenario must produce at least one motion curve')
    .toBeGreaterThan(0);
  if (process.env.PARITY_RECORD === '1') {
    mkdirSync(dirname(goldenPath), { recursive: true });
    writeFileSync(goldenPath, JSON.stringify(actual, null, 2) + '\n');
    return;
  }
  if (!existsSync(goldenPath)) {
    throw new Error(`missing golden ${name}; record with PARITY_RECORD=1`);
  }
  const golden: GeometryFingerprint = JSON.parse(readFileSync(goldenPath, 'utf-8'));
  expect(actual.fractions).toEqual(golden.fractions);
  const matches = (a: MotionCurve, b: MotionCurve): boolean => {
    const channel = (xs: (number | null)[], ys: (number | null)[]): boolean =>
      xs.length === ys.length && xs.every((x, i) => {
        const y = ys[i]!;
        if (x === null || y === null) return x === y;
        return Math.abs(x - y) <= tolerance;
      });
    // Timing is declared (not measured), so after 25ms-grid rounding it
    // must match exactly; zIndex values are integers compared exactly.
    const exact = (xs: (number | null)[], ys: (number | null)[]): boolean =>
      xs.length === ys.length && xs.every((x, i) => x === ys[i]);
    return channel(a.progress, b.progress)
      && channel(a.rotation, b.rotation)
      && channel(a.translation, b.translation)
      && channel(a.opacity, b.opacity)
      && a.timing[0] === b.timing[0]
      && a.timing[1] === b.timing[1]
      && exact(a.zIndex, b.zIndex);
  };
  for (const g of golden.curves) {
    expect(
      actual.curves.some((a) => matches(a, g)),
      `golden curve ${JSON.stringify(g)} has no tolerant match among observed curves ${JSON.stringify(actual.curves)}`,
    ).toBe(true);
  }
  for (const a of actual.curves) {
    expect(
      golden.curves.some((g) => matches(a, g)),
      `observed curve ${JSON.stringify(a)} has no tolerant match among golden curves`,
    ).toBe(true);
  }
}
