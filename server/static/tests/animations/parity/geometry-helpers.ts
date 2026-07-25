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

export interface MotionCurve {
  // Normalized displacement progress at each sampled fraction, rounded to
  // 2dp; null when the element does not translate (pure opacity/rotation).
  progress: (number | null)[];
  // Normalized computed-transform-matrix progress (Frobenius distance from
  // the start matrix over total matrix travel); null when the transform
  // channel does not change. Catches rotations/scales/flips that move no
  // pixels of bounding-rect center (e.g. a card flip's rotateY).
  transform: (number | null)[];
  // Opacity at each sampled fraction, rounded to 2dp; null when opacity is
  // not animated (stays within 0.01 of constant).
  opacity: (number | null)[];
}

export interface GeometryFingerprint {
  fractions: number[];
  // Sorted unique curves (JSON identity). The COUNT of animating elements
  // varies per game; the SET of distinct timing shapes does not.
  curves: MotionCurve[];
}

const GOLDEN_DIR = join(dirname(fileURLToPath(import.meta.url)), 'goldens');
const FRACTIONS = [0, 0.25, 0.5, 0.75, 1];

// Runs `trigger`, waits for animations to exist and their population to
// stabilize across two frames, pauses them all, seeks each through the
// sample fractions of its own (delay + activeDuration), measuring targets'
// positions and opacity, then finishes everything so the gate settles
// normally. Total paused wall-time stays well under the 4s watchdog floor.
export async function sampleMotionCurves(
  page: Page,
  trigger: () => Promise<void>,
): Promise<GeometryFingerprint> {
  await trigger();
  // One atomic in-page pass: find animations, wait for their population to
  // stabilize, pause, seek, sample, finish. document.getAnimations() does
  // NOT see shadow-tree animations in this Chromium (verified empirically:
  // 0 at document level while 141 ran inside component shadow roots), so
  // every step walks the shadow trees and collects per-element
  // getAnimations() instead.
  const fingerprint: GeometryFingerprint = await page.evaluate(async (fractions) => {
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
    interface Sample { x: number; y: number; matrix: number[]; opacity: number }

    // A cycle's animations arrive in WAVES (staggered cohorts, arrival
    // decorations, departure fades chained on earlier completions). A
    // single pause-at-first-stability latches an arbitrary wave, so the
    // sampled set varies run to run. Instead: iterate — pause and sample
    // each newly appeared wave, finish it (letting the choreography chain
    // to the next wave), and accumulate the UNION of curves until no new
    // animations appear for ~1s. Requires at least one wave overall.
    const samplesAll: Sample[][] = [];
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
      if (sawAny && now - quietSince > 1000) break;
      const freshCount = deepAnimations().filter((a) => !sampled.has(a)).length;
      if (freshCount === 0) { await frame(); continue; }
      // Two frames so co-scheduled stragglers of this wave are included.
      await frame(); await frame();
      const wave = deepAnimations().filter((a) => !sampled.has(a));
      if (wave.length === 0) { await frame(); continue; }
      sawAny = true;
      for (const a of wave) { sampled.add(a); a.pause(); }
      const waveSamples: Sample[][] = wave.map(() => []);
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
          waveSamples[i]!.push({
            x: r.x + r.width / 2,
            y: r.y + r.height / 2,
            matrix: parseMatrix(style.transform),
            opacity: parseFloat(style.opacity) || 0,
          });
        });
      }
      for (const a of wave) { try { a.finish(); } catch { try { a.cancel(); } catch { /* dead */ } } }
      samplesAll.push(...waveSamples);
      quietSince = performance.now();
    }

    const round = (v: number) => Math.round(v * 100) / 100;
    const curves = samplesAll
      .filter((s) => s.length === fractions.length)
      .map((s) => {
        const first = s[0]!, last = s[s.length - 1]!;
        const totalDist = Math.hypot(last.x - first.x, last.y - first.y);
        const translates = totalDist > 2; // px; below this it's not a travel animation
        const matrixDist = (a: number[], b: number[]): number =>
          Math.sqrt(a.reduce((acc, v, i) => acc + (v - (b[i] ?? 0)) ** 2, 0));
        const totalMatrix = matrixDist(last.matrix, first.matrix);
        const transforms = totalMatrix > 0.01;
        const opacities = s.map((p) => p.opacity);
        const opacityAnimates = Math.max(...opacities) - Math.min(...opacities) > 0.01;
        return {
          progress: s.map((p) => translates
            ? round(Math.hypot(p.x - first.x, p.y - first.y) / totalDist)
            : null),
          transform: s.map((p) => transforms
            ? round(matrixDist(p.matrix, first.matrix) / totalMatrix)
            : null),
          opacity: opacities.map((o) => (opacityAnimates ? round(o) : null)),
        };
      });
    // Unique by JSON identity, sorted for stable comparison. All-null
    // curves are dropped: they are sub-threshold noise (near-no-op FLIPs
    // whose presence is per-game random) and assert nothing.
    const seen = new Map<string, MotionCurveLike>();
    interface MotionCurveLike {
      progress: (number | null)[];
      transform: (number | null)[];
      opacity: (number | null)[];
    }
    for (const c of curves) {
      const allNull = [...c.progress, ...c.transform, ...c.opacity].every((v) => v === null);
      if (!allNull) seen.set(JSON.stringify(c), c);
    }
    return {
      fractions: [...fractions],
      curves: [...seen.keys()].sort().map((k) => seen.get(k)!),
    };
  }, FRACTIONS);
  return fingerprint;
}

// Compares (or with PARITY_RECORD=1, rewrites) the golden. Each golden
// curve must be matched by some observed curve within tolerance, and vice
// versa — set equivalence under a per-sample numeric tolerance, so minor
// sub-pixel jitter cannot flake while an easing/duration/keyframe change
// (which shifts mid-fraction progress by far more than the tolerance)
// always fails.
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
    return channel(a.progress, b.progress)
      && channel(a.transform, b.transform)
      && channel(a.opacity, b.opacity);
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
