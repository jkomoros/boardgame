import { test, expect } from '@playwright/test';
import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  fingerprintFromSamples,
  type GeometryFingerprint,
  type MotionSample,
  type SampledAnimation,
} from './geometry-helpers.js';

// Pure (no page, no server) specs for the motion fingerprint's
// normalization math, driven with synthetic sample sequences.
//
// Motivation: a tumbling 3D die is the first animation in the app whose
// pose does NOT travel monotonically from A to B. It spins through many
// intermediate orientations, its path length far exceeds its net
// displacement, and it can land on the same face in the same place it
// started. A fingerprint normalized by NET displacement (chord from the
// start pose over the start->end distance) breaks on all of that:
// intermediate values run well above 1 on an absolute 0.08 tolerance, they
// are non-monotone, and a landing that returns to the start pose drives the
// denominator to zero — which nulls the channel and (all channels null)
// silently DROPS the most complex animation in the app from the golden.
//
// The fix these specs pin: normalize by CUMULATIVE PATH LENGTH, which is
// monotone non-decreasing and in [0,1] by construction.

const FRACTIONS = [0, 0.25, 0.5, 0.75, 1];

// A CSS 2D rotate(deg) + translate(tx,ty) as the 16-entry column-major 4x4
// matrix the sampler produces.
const pose = (deg: number, tx: number, ty: number): number[] => {
  const r = (deg * Math.PI) / 180;
  const c = Math.cos(r), s = Math.sin(r);
  return [c, s, 0, 0, -s, c, 0, 0, 0, 0, 1, 0, tx, ty, 0, 1];
};

const sample = (deg: number, tx: number, ty: number): MotionSample => ({
  // Bounding-rect center tracks the translation; the element is 100x100 at
  // the origin before transform.
  x: 50 + tx,
  y: 50 + ty,
  matrix: pose(deg, tx, ty),
  opacity: 1,
  z: 0,
});

const animation = (samples: MotionSample[]): SampledAnimation => ({
  samples,
  durationMs: 500,
  delayMs: 0,
});

// Channels that are normalized progress fractions, and so must obey
// [0,1] + monotone non-decreasing. `opacity` is a raw measured value (it
// legitimately decreases on a fade) and `zIndex` is an integer, so neither
// is a normalized channel.
const NORMALIZED_CHANNELS = ['progress', 'rotation', 'translation'];

const normalizedChannelsOf = (curve: object): [string, (number | null)[]][] =>
  Object.entries(curve as Record<string, unknown>)
    .filter(([k, v]) => NORMALIZED_CHANNELS.includes(k) && Array.isArray(v))
    .map(([k, v]) => [k, v as (number | null)[]]);

test.describe('motion fingerprint normalization', () => {
  test('a tumble that returns to its start pose is not dropped', () => {
    // A die spun through a full turn and landing on the same face in the
    // same spot: net displacement is exactly zero on every channel, but a
    // great deal of motion happened.
    const fp = fingerprintFromSamples(
      [animation([
        sample(0, 0, 0),
        sample(90, 40, -30),
        sample(180, 80, 0),
        sample(270, 40, 30),
        sample(360, 0, 0),
      ])],
      FRACTIONS,
    );
    expect(
      fp.curves.length,
      'a full-turn tumble must survive as a curve; a zero net-displacement '
      + 'denominator must not null every channel and drop it',
    ).toBe(1);
    const channels = normalizedChannelsOf(fp.curves[0]!);
    expect(
      channels.some(([, values]) => values.some((v) => v !== null)),
      'the surviving curve must carry at least one non-null normalized channel',
    ).toBe(true);
  });

  test('a non-monotone tumble stays in [0,1] and never decreases', () => {
    // Out and back and out again: path length far exceeds net displacement.
    const fp = fingerprintFromSamples(
      [animation([
        sample(0, 0, 0),
        sample(170, 200, 60),
        sample(20, 40, -20),
        sample(200, 180, 90),
        sample(45, 30, 10),
      ])],
      FRACTIONS,
    );
    expect(fp.curves.length).toBe(1);
    for (const [name, values] of normalizedChannelsOf(fp.curves[0]!)) {
      const present = values.filter((v): v is number => v !== null);
      if (present.length === 0) continue;
      for (const v of present) {
        expect(v, `${name} sample ${v} is outside [0,1]`).toBeGreaterThanOrEqual(0);
        expect(v, `${name} sample ${v} is outside [0,1]`).toBeLessThanOrEqual(1);
      }
      for (let i = 1; i < present.length; i++) {
        expect(
          present[i]!,
          `${name} decreases at index ${i}: ${JSON.stringify(values)}`,
        ).toBeGreaterThanOrEqual(present[i - 1]!);
      }
    }
  });

  test('rotation is reported separately from pixel-scale translation', () => {
    // The failure this pins: Frobenius-summing unitless rotation entries
    // (max 2.83) with translations in raw pixels means a 600px travel makes
    // the rotation ~0.5% of the norm, so a single `transform` channel
    // degenerates into a duplicate of the translation. Here translation is
    // linear across the whole flight while ALL the rotation happens in the
    // last quarter; the two must be distinguishable.
    const fp = fingerprintFromSamples(
      [animation([
        sample(0, 0, 0),
        sample(0, 150, 0),
        sample(0, 300, 0),
        sample(0, 450, 0),
        sample(180, 600, 0),
      ])],
      FRACTIONS,
    );
    expect(fp.curves.length).toBe(1);
    const curve = fp.curves[0]! as unknown as Record<string, (number | null)[]>;
    expect(
      Object.keys(curve),
      'transform must be split into independently normalized rotation and '
      + 'translation channels',
    ).toEqual(expect.arrayContaining(['rotation', 'translation']));
    expect(curve.translation).toEqual([0, 0.25, 0.5, 0.75, 1]);
    expect(
      curve.rotation,
      'rotation happens entirely in the final quarter and must not be '
      + 'flattened into the translation ramp',
    ).toEqual([0, 0, 0, 0, 1]);
  });
});

// The same invariants, asserted over the RECORDED CORPUS rather than
// synthetic input. A normalizer regression, or a hand-edited golden, could
// otherwise reintroduce out-of-range or doubling-back samples into the
// baseline that every future comparison is measured against — silently,
// because the set comparison only asks whether observed and golden agree,
// never whether either is well-formed.
const GOLDEN_DIR = join(dirname(fileURLToPath(import.meta.url)), 'goldens');

test.describe('recorded geometry goldens are well-formed', () => {
  const files = readdirSync(GOLDEN_DIR)
    .filter((f) => f.startsWith('geometry-') && f.endsWith('.json'));

  test('the corpus is non-empty', () => {
    expect(files.length).toBeGreaterThan(0);
  });

  for (const file of files) {
    test(file, () => {
      const golden: GeometryFingerprint = JSON.parse(
        readFileSync(join(GOLDEN_DIR, file), 'utf-8'));
      expect(golden.fractions.length).toBeGreaterThan(1);
      expect(golden.curves.length, 'a golden with no curves asserts nothing')
        .toBeGreaterThan(0);
      for (const curve of golden.curves) {
        const c = curve as unknown as Record<string, unknown>;
        for (const name of ['progress', 'rotation', 'translation', 'opacity', 'zIndex']) {
          expect(Array.isArray(c[name]), `${name} must be a per-fraction array`).toBe(true);
          expect((c[name] as unknown[]).length).toBe(golden.fractions.length);
        }
        expect(
          [...curve.progress, ...curve.rotation, ...curve.translation, ...curve.opacity]
            .some((v) => v !== null),
          'an all-null curve asserts nothing and must have been dropped',
        ).toBe(true);
        for (const name of NORMALIZED_CHANNELS) {
          const values = c[name] as (number | null)[] | undefined;
          if (!values) continue;
          const present = values.filter((v): v is number => v !== null);
          if (present.length === 0) continue;
          expect(
            present.length,
            `${file}: ${name} mixes nulls with values; a channel is on or off`,
          ).toBe(values.length);
          expect(present[0], `${file}: ${name} must start at 0`).toBe(0);
          expect(present[present.length - 1], `${file}: ${name} must end at 1`).toBe(1);
          for (let i = 0; i < present.length; i++) {
            expect(present[i]!, `${file}: ${name} sample out of [0,1]`)
              .toBeGreaterThanOrEqual(0);
            expect(present[i]!, `${file}: ${name} sample out of [0,1]`)
              .toBeLessThanOrEqual(1);
            if (i > 0) {
              expect(
                present[i]!,
                `${file}: ${name} decreases at index ${i}: ${JSON.stringify(values)}`,
              ).toBeGreaterThanOrEqual(present[i - 1]!);
            }
          }
        }
      }
    });
  }
});
