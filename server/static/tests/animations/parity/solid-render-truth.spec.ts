import { test, expect, type Page } from '@playwright/test';

/**
 * DOES THIS SOLID DRAW WHAT A Z-BUFFER WOULD? — the only test in this repo that
 * can answer that, and the reason it exists at all.
 *
 * `src/solid/` renders a solid as one `clip-path`-ed DOM element per planar
 * face, and `backface-visibility: hidden` is its ENTIRE hidden-surface removal.
 * CSS `preserve-3d` has no depth buffer: within a 3D context the browser sorts
 * by a plane heuristic, not per pixel, so the only thing keeping the picture
 * honest is the theorem that on a CLOSED CONVEX surface the camera-facing
 * facets tile the silhouette exactly once and every other facet is hidden.
 * Culling the back faces is then provably sufficient, and provably insufficient
 * for anything that is not convex.
 *
 * Every other check in this suite asks a weaker question. `die-shape.spec.ts`
 * asks whether the silhouette has a HOLE in it — which catches a facet that
 * went missing but not two facets drawn in the wrong order, because a wrongly
 * sorted solid is still perfectly opaque. `geometry.spec.ts` and the golden
 * corpus compare against RECORDED behaviour, so they pin whatever shipped,
 * correct or not. This spec compares against a COMPUTED reference instead:
 *
 *   1. Render the solid under the real pipeline — the actual `solidFacets()`
 *      output, the real `perspective(PERSPECTIVE_DEPTH_DIE_SIZES em)` camera on
 *      the `preserve-3d` carrier, the real `backface-visibility: hidden` and
 *      `will-change: transform` per facet — with an opaque, distinguishable
 *      flat fill per facet so that every pixel names the facet that drew it.
 *   2. Rasterize THE SAME GEOMETRY with a real z-buffer under an IDENTICAL
 *      camera: one ray per subsample from the eye through the pixel, intersect
 *      every face's plane, nearest fragment wins. No culling anywhere — if
 *      culling is right, the nearest fragment is always a front-facing facet
 *      and the two pictures agree.
 *   3. Diff, and report the largest wrong region that is more than a hairline.
 *
 * There is no golden file here and there must never be one: a recorded image
 * would freeze today's answer, and the whole point is that the answer is
 * derived from the geometry every run.
 *
 * ## Why the diff needs two separate antialiasing exclusions
 *
 * Naively differenced, every run is noise. The browser antialiases each facet's
 * `clip-path` edge and composites each facet as its own layer, so along every
 * silhouette edge and every shared facet edge there is a band of blended pixels
 * that belong to neither facet — and the die has a documented 1px hairline
 * residual on shared edges that is a known, accepted, purely cosmetic artifact.
 * A test that failed on those would be useless, and one that merely raised its
 * tolerance would stop being able to see a real mis-sort. So:
 *
 *   - The reference is supersampled 3x3. A pixel whose nine subsamples do not
 *     all name the same facet is a pixel the reference itself considers a seam,
 *     and it is NOT JUDGED AT ALL.
 *   - Those seams are then DILATED BY 2px, because the browser's antialiasing
 *     and layer compositing spread a little wider than the geometric edge, and
 *     because a sub-pixel disagreement between the reference camera and the
 *     browser's own rounding shifts every edge by a fraction of a pixel.
 *   - What survives is differenced exactly, and the wrong pixels are ERODED
 *     ONCE before being measured. A 1px hairline does not survive an erosion; a
 *     facet drawn in front of the facet that should have hidden it does.
 *
 * Both numbers are reported: the wrong-pixel fraction of the judged silhouette
 * (sensitive, noisy, informational) and the largest connected region surviving
 * the erosion (what is actually asserted).
 *
 * ## The two controls, which are assertions and not anecdotes
 *
 * A harness like this fails in two directions, and both failures look exactly
 * like a pass, so both are pinned by tests in this file rather than by a note
 * in a commit message:
 *
 *   - `a d20 is rendered exactly` is the NEGATIVE control. A d20 is convex, so
 *     culling renders it correctly, and any meaningful error this harness
 *     reports on a d20 is the harness being wrong about the camera — not the
 *     browser being wrong about the picture. It measures zero wrong pixels,
 *     not merely zero thick regions.
 *   - `a mis-painted solid is caught` is the POSITIVE control. Every facet is
 *     deliberately painted with a neighbour's colour; nothing about the
 *     geometry, the camera or the exclusions changes. If that does not light up
 *     nearly the whole silhouette, the harness is judging nothing — which is
 *     exactly what a vacuous test looks like from the outside, and this project
 *     has shipped five of those, each caught only by sabotage.
 *
 * Every shape additionally asserts a floor on how many pixels it judged, and on
 * how confidently it classified them, so "the exclusions ate the whole image"
 * and "every colour landed halfway between two entries" cannot pass either.
 *
 * ## What it measures when the real mechanism is broken
 *
 * The positive control breaks the PAINT. To check that the assertions also bite
 * on the mechanism actually under test, the facets' `backface-visibility` was
 * flipped to `visible` — removing the solid's entire hidden-surface removal —
 * and this spec re-run:
 *
 *   d20     2 of 8 poses fail, thickest wrong regions 40px and 62px
 *   d6      0 of 8 poses fail
 *   prism   0 of 8 poses fail
 *
 * Both halves of that are worth knowing. The d20 fails, so `thickest === 0` is
 * a real assertion against a real breakage and not just against mis-painting.
 * And the d6 and the prism do NOT fail, because with opaque flat facets CSS's
 * own plane sort happens to get those two shapes right without any culling at
 * all — which is precisely why a spec like this is needed rather than an eyeball
 * check on a cube: the shape whose rendering is easiest to look at is the shape
 * least able to tell you anything. The d20 carries the sensitivity here.
 */

// ---------------------------------------------------------------------------
// Scene constants. The die's own camera, at a size big enough that a mis-sorted
// facet is thousands of pixels rather than dozens.
// ---------------------------------------------------------------------------

/** `font-size` on the stage: what `1em`, and so the solid's nominal box, is. */
const DIE_PX = 150;

/**
 * The stage, and so the screenshot. Big enough for the widest thing measured
 * here: a 12-side prism at height ratio 0.55 has a bounding radius 1.141x its
 * nominal one, which the camera magnifies to about 90px — 110px of half-stage
 * leaves 20px of margin, and the margin matters because background pixels are
 * judged too.
 */
const STAGE_PX = 220;

/** 3x3 supersampling of the reference. See the file docs. */
const SUBSAMPLES = 3;

/** How far a reference seam is grown before anything near it is judged. */
const SEAM_DILATION_PX = 2;

/**
 * A judged silhouette smaller than this means the solid did not render, or the
 * exclusions ate it. It is a COLLAPSE detector and not a tolerance: the
 * observed minimum over every shape and pose here is about 6,900 (a d6, which
 * shows the least of itself), and the prism runs to 17,000. It is set well
 * below the former so that legitimate variation between poses never trips it
 * and a blank stage always does.
 */
const MIN_JUDGED_SILHOUETTE_PX = 4000;

/**
 * The palette has to be separable enough that classifying a pixel to its
 * nearest entry is a decision and not a coin flip. Asserted, not assumed —
 * the generator below runs out of room somewhere past 20 facets.
 */
const MIN_PALETTE_DISTANCE = 40;

/**
 * How far a judged pixel may sit from the palette entry it was classified as.
 *
 * This is the assertion that the EXCLUSIONS are right, and it is the one that
 * would catch them silently drifting. A pixel that survived them is supposed to
 * be a pixel one facet painted flat and alone, so it should be that facet's
 * colour EXACTLY — and measured, every judged pixel of every honest pose here
 * is at distance 0, over three shapes and 24 poses. Anything blended would show
 * up here first: with the palette's entries at least 69 apart, an even mix of
 * two facets lands about 35 away. 20 leaves room for a PNG round trip or a
 * colour-space conversion moving a channel and no room at all for a blend.
 */
const MAX_JUDGED_COLOUR_DISTANCE = 20;

// ---------------------------------------------------------------------------
// Vector and matrix arithmetic, deliberately written out here rather than
// imported from `src/`.
//
// This is a REFERENCE implementation. If it shared code with the thing it is
// checking, a sign error in the shared part would cancel out and the test would
// agree with the bug. The only thing taken from `src/` is the geometry itself
// and the camera depth constant — the facts under test — and the one convention
// restated below, the physics-to-CSS axis flip, which `screen-frame.ts` owns
// and documents as `S = diag(1, -1, 1)`.
// ---------------------------------------------------------------------------

type V3 = [number, number, number];
/** Row-major 3x3. `m[r * 3 + c]`. */
type M3 = number[];

const IDENTITY_M3: M3 = [1, 0, 0, 0, 1, 0, 0, 0, 1];

function sub(a: V3, b: V3): V3 { return [a[0] - b[0], a[1] - b[1], a[2] - b[2]]; }
function dot(a: V3, b: V3): number { return a[0] * b[0] + a[1] * b[1] + a[2] * b[2]; }
function cross(a: V3, b: V3): V3 {
  return [
    a[1] * b[2] - a[2] * b[1],
    a[2] * b[0] - a[0] * b[2],
    a[0] * b[1] - a[1] * b[0],
  ];
}
function norm(a: V3): V3 {
  const length = Math.hypot(a[0], a[1], a[2]);
  return [a[0] / length, a[1] / length, a[2] / length];
}
function applyM3(m: M3, v: V3): V3 {
  return [
    m[0] * v[0] + m[1] * v[1] + m[2] * v[2],
    m[3] * v[0] + m[4] * v[1] + m[5] * v[2],
    m[6] * v[0] + m[7] * v[1] + m[8] * v[2],
  ];
}
function mulM3(a: M3, b: M3): M3 {
  const out: M3 = new Array(9).fill(0);
  for (let r = 0; r < 3; r++) {
    for (let c = 0; c < 3; c++) {
      let sum = 0;
      for (let k = 0; k < 3; k++) sum += a[r * 3 + k] * b[k * 3 + c];
      out[r * 3 + c] = sum;
    }
  }
  return out;
}

/** Rodrigues. `axis` need not be normalized; a zero axis gives the identity. */
function rotationAbout(axis: V3, angle: number): M3 {
  const length = Math.hypot(axis[0], axis[1], axis[2]);
  if (!(length > 1e-12)) return IDENTITY_M3.slice();
  const [x, y, z] = [axis[0] / length, axis[1] / length, axis[2] / length];
  const c = Math.cos(angle);
  const s = Math.sin(angle);
  const t = 1 - c;
  return [
    t * x * x + c, t * x * y - s * z, t * x * z + s * y,
    t * x * y + s * z, t * y * y + c, t * y * z - s * x,
    t * x * z - s * y, t * y * z + s * x, t * z * z + c,
  ];
}

/** The shortest rotation carrying unit `from` onto unit `to`. */
function rotationBetween(from: V3, to: V3): M3 {
  const axis = cross(from, to);
  const sine = Math.hypot(axis[0], axis[1], axis[2]);
  const cosine = Math.max(-1, Math.min(1, dot(from, to)));
  if (sine < 1e-12) {
    if (cosine > 0) return IDENTITY_M3.slice();
    // Antiparallel: any perpendicular axis, half a turn.
    const seed: V3 = Math.abs(from[0]) < 0.9 ? [1, 0, 0] : [0, 1, 0];
    return rotationAbout(cross(from, seed), Math.PI);
  }
  return rotationAbout(axis, Math.atan2(sine, cosine));
}

/** A deterministic uniform random rotation (Shoemake), from a 32-bit seed. */
function seededRotation(seed: number): M3 {
  let state = (seed * 2654435761) >>> 0;
  const next = () => {
    state = (state * 1664525 + 1013904223) >>> 0;
    return state / 4294967296;
  };
  // Burn a few, so nearby seeds do not give nearby poses.
  next(); next(); next();
  const u1 = next();
  const u2 = next();
  const u3 = next();
  const x = Math.sqrt(1 - u1) * Math.sin(2 * Math.PI * u2);
  const y = Math.sqrt(1 - u1) * Math.cos(2 * Math.PI * u2);
  const z = Math.sqrt(u1) * Math.sin(2 * Math.PI * u3);
  const w = Math.sqrt(u1) * Math.cos(2 * Math.PI * u3);
  return [
    1 - 2 * (y * y + z * z), 2 * (x * y - z * w), 2 * (x * z + y * w),
    2 * (x * y + z * w), 1 - 2 * (x * x + z * z), 2 * (y * z - x * w),
    2 * (x * z - y * w), 2 * (y * z + x * w), 1 - 2 * (x * x + y * y),
  ];
}

/**
 * `screen-frame.ts`'s `S = diag(1, -1, 1)`: body frame (+Y up, right handed) to
 * CSS frame (+Y down, left handed). Restated rather than imported — see the
 * note at the top of this section. `screen-frame.test.ts` is what pins that the
 * implementation still agrees with this.
 */
function toScreen(v: V3): V3 { return [v[0], -v[1], v[2]]; }

/** Column-major, which is the order `matrix3d()` takes its arguments in. */
function matrix3dOf(m: M3): string {
  return `matrix3d(${[
    m[0], m[3], m[6], 0,
    m[1], m[4], m[7], 0,
    m[2], m[5], m[8], 0,
    0, 0, 0, 1,
  ].join(',')})`;
}

// ---------------------------------------------------------------------------
// The palette.
// ---------------------------------------------------------------------------

function hslToRgb(hue: number, saturation: number, lightness: number): [number, number, number] {
  const channel = (n: number) => {
    const k = (n + hue * 12) % 12;
    const a = saturation * Math.min(lightness, 1 - lightness);
    return Math.round(255 * (lightness - a * Math.max(-1, Math.min(k - 3, 9 - k, 1))));
  };
  return [channel(0), channel(8), channel(4)];
}

/**
 * Entry 0 is the background; entry `i + 1` is facet `i`'s own colour. Hues walk
 * by the golden angle and lightness cycles through three tiers, so neighbouring
 * facets — which are the pairs a sort error confuses — are far apart in RGB
 * rather than merely different. `MIN_PALETTE_DISTANCE` is asserted per shape.
 */
function paletteFor(facetCount: number): [number, number, number][] {
  const palette: [number, number, number][] = [[8, 8, 8]];
  for (let i = 0; i < facetCount; i++) {
    const hue = ((i * 137.507764) % 360) / 360;
    palette.push(hslToRgb(hue, 0.92, [0.38, 0.56, 0.74][i % 3]));
  }
  return palette;
}

function minPaletteDistance(palette: [number, number, number][]): number {
  let smallest = Infinity;
  for (let i = 0; i < palette.length; i++) {
    for (let j = i + 1; j < palette.length; j++) {
      smallest = Math.min(smallest, Math.hypot(
        palette[i][0] - palette[j][0],
        palette[i][1] - palette[j][1],
        palette[i][2] - palette[j][2],
      ));
    }
  }
  return smallest;
}

// ---------------------------------------------------------------------------
// The reference rasterizer: the same geometry, with a real z-buffer.
// ---------------------------------------------------------------------------

/** One polygon of the surface, already posed and in screen pixels. */
interface PosedFace {
  readonly key: number;
  /**
   * Plane normal, from Newell over the posed loop, flipped if necessary to
   * point AWAY from the origin.
   *
   * The flip is not cosmetic. `toScreen` is a reflection, so it reverses every
   * loop's apparent winding and Newell's normal on the mapped loop comes out
   * pointing INWARD. Rather than restate that (and get it wrong the day a
   * convention moves), the direction is taken from the geometry: every solid
   * here is convex and centred on the origin, so the outward normal is the one
   * with a positive dot against the face's own centroid. The rasterizer does
   * not care about the sign at all — the poses do.
   */
  readonly normal: V3;
  /** `dot(normal, anyPointOnThePlane)`. */
  readonly offset: number;
  readonly origin: V3;
  readonly centroid: V3;
  readonly e1: V3;
  readonly e2: V3;
  /** The loop in the plane's own 2D frame, for the containment test. */
  readonly loop2d: readonly (readonly [number, number])[];
  /**
   * The projected screen-space bounding box, as a cheap rejection. A planar
   * loop entirely in front of the eye projects inside the box of its projected
   * vertices, so a subsample outside it cannot possibly land on this facet.
   * Pure speed: removing it changes no number this file reports.
   */
  readonly minX: number;
  readonly maxX: number;
  readonly minY: number;
  readonly maxY: number;
}

/**
 * Pose the surface and pixel-scale it, exactly as the DOM will be posed:
 * body -> `toScreen` -> `0.5 / nominalRadius` per em (`facet-placement.ts`'s
 * normalization, which is the contract `solidFacets` publishes) -> em to px ->
 * the `#orient` rotation.
 */
function poseSurface(
  polygons: readonly (readonly V3[])[],
  nominalRadius: number,
  pose: M3,
  depthPx: number,
): PosedFace[] {
  const unitsToPx = (0.5 / nominalRadius) * DIE_PX;
  return polygons.map((polygon, key) => {
    const points = polygon.map((vertex) => {
      const screen = toScreen(vertex as V3);
      return applyM3(pose, [
        screen[0] * unitsToPx,
        screen[1] * unitsToPx,
        screen[2] * unitsToPx,
      ]);
    });
    // Newell, so the plane comes from the whole loop rather than from three of
    // its vertices — a numerically better normal, and one that does not care
    // which three.
    let nx = 0;
    let ny = 0;
    let nz = 0;
    for (let i = 0; i < points.length; i++) {
      const current = points[i];
      const next = points[(i + 1) % points.length];
      nx += (current[1] - next[1]) * (current[2] + next[2]);
      ny += (current[2] - next[2]) * (current[0] + next[0]);
      nz += (current[0] - next[0]) * (current[1] + next[1]);
    }
    let normal = norm([nx, ny, nz]);
    const centroid: V3 = [
      points.reduce((s, p) => s + p[0], 0) / points.length,
      points.reduce((s, p) => s + p[1], 0) / points.length,
      points.reduce((s, p) => s + p[2], 0) / points.length,
    ];
    // Outward, taken from the geometry rather than from the winding: see the
    // note on `PosedFace.normal`.
    if (dot(normal, centroid) < 0) normal = [-normal[0], -normal[1], -normal[2]];
    const origin = points[0];
    const e1 = norm(sub(points[1], points[0]));
    const e2 = cross(normal, e1);
    const loop2d = points.map((point) => {
      const delta = sub(point, origin);
      return [dot(delta, e1), dot(delta, e2)] as const;
    });
    let minX = Infinity;
    let maxX = -Infinity;
    let minY = Infinity;
    let maxY = -Infinity;
    for (const point of points) {
      const magnify = depthPx / (depthPx - point[2]);
      const sx = point[0] * magnify;
      const sy = point[1] * magnify;
      if (sx < minX) minX = sx;
      if (sx > maxX) maxX = sx;
      if (sy < minY) minY = sy;
      if (sy > maxY) maxY = sy;
    }
    return {
      key,
      normal,
      offset: dot(normal, origin),
      origin,
      centroid,
      e1,
      e2,
      loop2d,
      minX,
      maxX,
      minY,
      maxY,
    };
  });
}

/** Crossing number. Works for any simple polygon, convex or not. */
function contains(loop: readonly (readonly [number, number])[], a: number, b: number): boolean {
  let inside = false;
  for (let i = 0, j = loop.length - 1; i < loop.length; j = i++) {
    const [ai, bi] = loop[i];
    const [aj, bj] = loop[j];
    if ((bi > b) !== (bj > b) && a < ((aj - ai) * (b - bi)) / (bj - bi) + ai) inside = !inside;
  }
  return inside;
}

/**
 * What the camera should see, per pixel: the facet's palette slot (`key + 1`),
 * 0 for background, or -1 for a pixel whose nine subsamples disagreed.
 *
 * The camera is CSS's: `perspective(d)` puts the eye at `(0, 0, d)` in the
 * transform origin's frame and projects onto `z = 0`, so the ray through a
 * screen point `(sx, sy)` measured from the stage centre is
 * `(0, 0, d) + t * (sx, sy, -d)` and `t` increases away from the eye. Nearest
 * fragment — smallest positive `t` — wins, with NO back-face culling, which is
 * the entire question being asked.
 */
function referenceRender(faces: readonly PosedFace[], depthPx: number): Int16Array {
  const expected = new Int16Array(STAGE_PX * STAGE_PX);
  const half = STAGE_PX / 2;
  const step = 1 / SUBSAMPLES;
  for (let y = 0; y < STAGE_PX; y++) {
    for (let x = 0; x < STAGE_PX; x++) {
      let agreed = 0;
      let first = 0;
      let seam = false;
      for (let sy = 0; sy < SUBSAMPLES && !seam; sy++) {
        for (let sx = 0; sx < SUBSAMPLES && !seam; sx++) {
          const px = x + (sx + 0.5) * step - half;
          const py = y + (sy + 0.5) * step - half;
          let bestT = Infinity;
          let bestKey = 0;
          for (const face of faces) {
            if (px < face.minX || px > face.maxX || py < face.minY || py > face.maxY) continue;
            const denominator = face.normal[0] * px + face.normal[1] * py
              - face.normal[2] * depthPx;
            if (Math.abs(denominator) < 1e-12) continue;
            const t = (face.offset - face.normal[2] * depthPx) / denominator;
            if (!(t > 0) || t >= bestT) continue;
            const hit: V3 = [t * px, t * py, depthPx - t * depthPx];
            const delta = sub(hit, face.origin);
            if (!contains(face.loop2d, dot(delta, face.e1), dot(delta, face.e2))) continue;
            bestT = t;
            bestKey = face.key + 1;
          }
          if (sx === 0 && sy === 0) first = bestKey;
          else if (bestKey !== first) seam = true;
          agreed = first;
        }
      }
      expected[y * STAGE_PX + x] = seam ? -1 : agreed;
    }
  }
  return expected;
}

// ---------------------------------------------------------------------------
// Morphology and the verdict.
// ---------------------------------------------------------------------------

/** Chebyshev dilation by `radius`, separably. */
function dilate(mask: Uint8Array, radius: number): Uint8Array {
  const horizontal = new Uint8Array(mask.length);
  for (let y = 0; y < STAGE_PX; y++) {
    for (let x = 0; x < STAGE_PX; x++) {
      let hit = 0;
      for (let d = -radius; d <= radius && !hit; d++) {
        const nx = x + d;
        if (nx >= 0 && nx < STAGE_PX && mask[y * STAGE_PX + nx]) hit = 1;
      }
      horizontal[y * STAGE_PX + x] = hit;
    }
  }
  const out = new Uint8Array(mask.length);
  for (let y = 0; y < STAGE_PX; y++) {
    for (let x = 0; x < STAGE_PX; x++) {
      let hit = 0;
      for (let d = -radius; d <= radius && !hit; d++) {
        const ny = y + d;
        if (ny >= 0 && ny < STAGE_PX && horizontal[ny * STAGE_PX + x]) hit = 1;
      }
      out[y * STAGE_PX + x] = hit;
    }
  }
  return out;
}

/**
 * One 3x3 erosion: a pixel survives only if all eight of its neighbours are
 * also set. A 1px hairline cannot survive it; a 3px-wide region can. Pixels off
 * the edge of the image count as unset, which is the conservative direction.
 */
function erodeOnce(mask: Uint8Array): Uint8Array {
  const out = new Uint8Array(mask.length);
  for (let y = 1; y < STAGE_PX - 1; y++) {
    for (let x = 1; x < STAGE_PX - 1; x++) {
      const p = y * STAGE_PX + x;
      if (!mask[p]) continue;
      let all = 1;
      for (let dy = -1; dy <= 1 && all; dy++) {
        for (let dx = -1; dx <= 1 && all; dx++) {
          if (!mask[p + dy * STAGE_PX + dx]) all = 0;
        }
      }
      out[p] = all;
    }
  }
  return out;
}

function largestRegion(mask: Uint8Array): { size: number; bbox: string } {
  const seen = new Uint8Array(mask.length);
  let best = 0;
  let bbox = 'none';
  for (let start = 0; start < mask.length; start++) {
    if (!mask[start] || seen[start]) continue;
    const stack = [start];
    seen[start] = 1;
    let count = 0;
    let x0 = STAGE_PX;
    let y0 = STAGE_PX;
    let x1 = -1;
    let y1 = -1;
    while (stack.length) {
      const p = stack.pop()!;
      count++;
      const px = p % STAGE_PX;
      const py = (p / STAGE_PX) | 0;
      if (px < x0) x0 = px;
      if (px > x1) x1 = px;
      if (py < y0) y0 = py;
      if (py > y1) y1 = py;
      const neighbours = [
        px > 0 ? p - 1 : -1,
        px < STAGE_PX - 1 ? p + 1 : -1,
        py > 0 ? p - STAGE_PX : -1,
        py < STAGE_PX - 1 ? p + STAGE_PX : -1,
      ];
      for (const q of neighbours) if (q >= 0 && mask[q] && !seen[q]) { seen[q] = 1; stack.push(q); }
    }
    if (count > best) {
      best = count;
      bbox = `${x1 - x0 + 1}x${y1 - y0 + 1} at ${x0},${y0}`;
    }
  }
  return { size: best, bbox };
}

interface Verdict {
  /** Pixels judged at all: not a reference seam and not within the dilation. */
  readonly judged: number;
  /** Judged pixels the reference says are part of the solid. */
  readonly silhouette: number;
  /** Judged pixels naming the wrong facet (or the wrong presence of one). */
  readonly wrong: number;
  /** `wrong / silhouette`, the sensitive-but-noisy number. */
  readonly wrongFractionOfSilhouette: number;
  /** `wrong / judged`, which is what the positive control is quoted in. */
  readonly wrongFractionOfJudged: number;
  /** Largest connected wrong region surviving one erosion. THE assertion. */
  readonly thickest: number;
  readonly thickestBox: string;
  /** How far the worst judged pixel was from any palette entry. */
  readonly worstColourDistance: number;
  /** A few wrong pixels, spelled out, so a failure says what went wrong where. */
  readonly samples: readonly string[];
}

/**
 * Every pixel the reference itself considers ambiguous — BOTH ways, and the
 * second way is not optional.
 *
 * The obvious source is subsample disagreement: nine rays, more than one
 * answer, so the geometric edge crosses this pixel. What that misses is an edge
 * that lands exactly ON a pixel boundary, where no pixel has partial coverage
 * and all nine subsamples of every pixel agree — which is not a corner case
 * here, it is the COMMON case, because the poses that matter most are the
 * symmetric ones and a symmetric pose puts a shared facet edge straight down
 * the middle of a stage whose width is even. Measured: a d20 with only the
 * disagreement test left a 1px column of blended pixels at x=109 on the exact
 * centre line, classified to the neighbouring facet, and it came and went
 * between runs of the identical pose — a flake, in a spec whose whole value is
 * that it does not flake.
 *
 * So a pixel is also a seam if any of its eight neighbours resolves to a
 * DIFFERENT facet. That is the honest statement of what is being excluded:
 * not "pixels the reference could not resolve" but "pixels near a boundary the
 * browser is entitled to antialias".
 */
function seamMask(expected: Int16Array): Uint8Array {
  const seam = new Uint8Array(expected.length);
  for (let y = 0; y < STAGE_PX; y++) {
    for (let x = 0; x < STAGE_PX; x++) {
      const p = y * STAGE_PX + x;
      if (expected[p] < 0) { seam[p] = 1; continue; }
      for (let dy = -1; dy <= 1 && !seam[p]; dy++) {
        for (let dx = -1; dx <= 1 && !seam[p]; dx++) {
          const nx = x + dx;
          const ny = y + dy;
          if (nx < 0 || nx >= STAGE_PX || ny < 0 || ny >= STAGE_PX) continue;
          const q = ny * STAGE_PX + nx;
          if (expected[q] >= 0 && expected[q] !== expected[p]) seam[p] = 1;
        }
      }
    }
  }
  return seam;
}

function compare(expected: Int16Array, observed: Uint8Array, distance: Uint8Array): Verdict {
  const excluded = dilate(seamMask(expected), SEAM_DILATION_PX);

  const wrongMask = new Uint8Array(expected.length);
  const samples: string[] = [];
  let judged = 0;
  let silhouette = 0;
  let wrong = 0;
  let worstColourDistance = 0;
  for (let p = 0; p < expected.length; p++) {
    if (excluded[p]) continue;
    judged++;
    if (expected[p] > 0) silhouette++;
    if (distance[p] > worstColourDistance) worstColourDistance = distance[p];
    if (observed[p] !== expected[p]) {
      wrongMask[p] = 1;
      wrong++;
      if (samples.length < 12) {
        samples.push(`(${p % STAGE_PX},${(p / STAGE_PX) | 0}) want ${expected[p]}`
          + ` got ${observed[p]} at colour distance ${distance[p]}`);
      }
    }
  }
  const { size, bbox } = largestRegion(erodeOnce(wrongMask));
  return {
    judged,
    silhouette,
    wrong,
    wrongFractionOfSilhouette: silhouette ? wrong / silhouette : 1,
    wrongFractionOfJudged: judged ? wrong / judged : 1,
    thickest: size,
    thickestBox: bbox,
    worstColourDistance,
    samples,
  };
}

// ---------------------------------------------------------------------------
// The page side: the real pipeline, and the only PNG decoder to hand.
// ---------------------------------------------------------------------------

interface ShapeSpec {
  /** `die` takes a face count; `prism` takes a side count and a height ratio. */
  readonly kind: 'die' | 'prism';
  readonly count: number;
  readonly heightRatio?: number;
}

interface PreparedShape {
  /** `[...faces, ...capFaces]` in body coordinates: the surface under test. */
  readonly polygons: V3[][];
  readonly nominalRadius: number;
  /** `PERSPECTIVE_DEPTH_DIE_SIZES` — read from `src/`, never restated. */
  readonly perspectiveDieSizes: number;
}

/**
 * Install the in-page harness: builds the scene out of the REAL `solidFacets()`
 * output under the REAL camera, and classifies a screenshot's pixels.
 *
 * All of it lives in one `page.evaluate` because a `page.evaluate` callback
 * cannot close over anything in this file.
 */
async function installHarness(page: Page, stagePx: number, diePx: number): Promise<void> {
  await page.evaluate(async ({ stagePx: STAGE, diePx: SIZE }) => {
    const facetPlacement = await import('/src/solid/facet-placement.ts');
    const prism = await import('/src/solid/prism.ts');
    const dieGeometry = await import('/src/motion/die-geometry.ts');
    const diceRoll = await import('/src/motion/dice-roll.ts');

    interface Surface {
      faces: readonly { readonly polygon: readonly (readonly number[])[] }[];
      capFaces: readonly { readonly polygon: readonly (readonly number[])[] }[];
      nominalRadius: number;
    }

    const surfaceFor = (kind: string, count: number, heightRatio: number): Surface =>
      (kind === 'prism'
        ? prism.prismSurface(count, heightRatio)
        : dieGeometry.dieGeometry(count)) as unknown as Surface;

    let current: { surface: Surface; facets: readonly { key: number; style: string }[] } | null = null;

    (window as any).__renderTruth = {
      /** Build the surface, and hand its raw body-frame geometry back. */
      prepare(kind: string, count: number, heightRatio: number) {
        const surface = surfaceFor(kind, count, heightRatio);
        current = { surface, facets: facetPlacement.solidFacets(surface as any) };
        return {
          polygons: [...surface.faces, ...surface.capFaces]
            .map((face) => face.polygon.map((vertex) => [vertex[0], vertex[1], vertex[2]])),
          nominalRadius: surface.nominalRadius,
          perspectiveDieSizes: diceRoll.PERSPECTIVE_DEPTH_DIE_SIZES,
        };
      },

      /**
       * Mount the scene, posed by `matrix3d`, one flat opaque fill per facet.
       *
       * The element chain is `boardgame-die.ts`'s own, minus the die: a stage
       * whose `font-size` IS the solid's size (so every generated `em` resolves
       * against it), the `preserve-3d` carrier that holds the camera, the
       * `preserve-3d` pose carrier, and the facets — `left/top: 50%` with the
       * generated negative margins, so every element in the chain shares one
       * transform origin at the stage's centre, and `backface-visibility:
       * hidden` plus `will-change: transform`, which are the culling and the
       * per-facet layer whose interaction this whole spec exists to check.
       */
      async render(pose: string, colours: readonly string[], background: string) {
        (window as any).__renderTruth.teardown();
        const stage = document.createElement('div');
        stage.id = 'render-truth-stage';
        stage.style.cssText = `position:fixed;left:0;top:0;width:${STAGE}px;height:${STAGE}px;`
          + `font-size:${SIZE}px;background:${background};overflow:hidden;`
          + 'z-index:2147483647;contain:none;';
        const carrier = document.createElement('div');
        carrier.style.cssText = 'position:absolute;inset:0;transform-style:preserve-3d;'
          + `transform:perspective(${(window as any).__renderTruth.depthEm}em);`;
        const orient = document.createElement('div');
        orient.style.cssText = `position:absolute;inset:0;transform-style:preserve-3d;transform:${pose};`;
        for (const facet of current!.facets) {
          const el = document.createElement('div');
          el.setAttribute('style', `${facet.style};position:absolute;left:50%;top:50%;`
            + `backface-visibility:hidden;will-change:transform;background:${colours[facet.key]};`);
          orient.appendChild(el);
        }
        carrier.appendChild(orient);
        stage.appendChild(carrier);
        document.body.appendChild(stage);
        // Two frames plus a beat: `will-change: transform` promotes each facet
        // to its own layer, and a promotion applied in the same task that
        // mounts the scene is not reliably in place for the first frame — the
        // same reason `boardgame-die.ts` declares it unconditionally.
        await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
        await new Promise<void>((r) => setTimeout(r, 40));
        const box = stage.getBoundingClientRect();
        return { x: box.x, y: box.y, width: box.width, height: box.height };
      },

      depthEm: 0,

      teardown() {
        document.getElementById('render-truth-stage')?.remove();
      },

      /**
       * Decode the screenshot and name, per pixel, the palette entry it is
       * nearest to. Nearest-entry rather than exact-match on purpose: the PNG
       * round trip and the canvas colour space can each move a channel by one,
       * and a threshold that rejected those would silently stop judging
       * pixels. The distance to the winner is returned alongside so that "every
       * pixel classified confidently" is a fact the caller can assert on
       * instead of an assumption.
       */
      async classify(base64: string, palette: readonly (readonly number[])[]) {
        const image = new Image();
        image.src = 'data:image/png;base64,' + base64;
        await image.decode();
        const canvas = document.createElement('canvas');
        canvas.width = image.width;
        canvas.height = image.height;
        const context = canvas.getContext('2d', { willReadFrequently: true })!;
        context.drawImage(image, 0, 0);
        const data = context.getImageData(0, 0, canvas.width, canvas.height).data;
        const total = canvas.width * canvas.height;
        const classes = new Uint8Array(total);
        const distances = new Uint8Array(total);
        for (let p = 0; p < total; p++) {
          const r = data[p * 4];
          const g = data[p * 4 + 1];
          const b = data[p * 4 + 2];
          let best = 0;
          let bestDistance = Infinity;
          for (let k = 0; k < palette.length; k++) {
            const dr = r - palette[k][0];
            const dg = g - palette[k][1];
            const db = b - palette[k][2];
            const distance = dr * dr + dg * dg + db * db;
            if (distance < bestDistance) { bestDistance = distance; best = k; }
          }
          classes[p] = best;
          distances[p] = Math.min(255, Math.round(Math.sqrt(bestDistance)));
        }
        const pack = (bytes: Uint8Array) => {
          let text = '';
          for (let i = 0; i < bytes.length; i += 4096) {
            text += String.fromCharCode(...bytes.subarray(i, i + 4096));
          }
          return btoa(text);
        };
        return {
          width: canvas.width,
          height: canvas.height,
          classes: pack(classes),
          distances: pack(distances),
        };
      },
    };
  }, { stagePx, diePx });
}

// ---------------------------------------------------------------------------
// The pose sets, and why these poses.
// ---------------------------------------------------------------------------

interface Pose {
  readonly name: string;
  readonly matrix: M3;
}

/** A face's outward normal in the CSS frame, at rest. */
function faceNormals(polygons: readonly (readonly V3[])[]): V3[] {
  // The depth here only feeds the projected bounding box, which poses ignore.
  return poseSurface(polygons, 1, IDENTITY_M3, 1e6).map((face) => face.normal);
}

/**
 * The poses, chosen rather than swept.
 *
 * The original experiment swept sixteen randomized poses per tilt band across
 * many bands — 253 for the d20 alone — which is the right shape for an
 * EXPERIMENT deciding whether a rendering approach works at all, and the wrong
 * one for a spec that has to run in the normal suite. Randomly sampled poses
 * are also mostly uninformative: a facet whose normal is 40 degrees off the
 * camera is not a facet any renderer gets wrong. So this picks the poses where
 * the two renderers could actually disagree, and adds a few random ones only as
 * a net for whatever this reasoning missed.
 *
 *   - `rest` — the identity. It is the pose a die that has never rolled and a
 *     token that has never moved actually render in, so it is the one a user is
 *     most likely to be looking at, and the one an off-by-one in the facet
 *     chain shows up in most plainly.
 *   - `edge-on -0.4deg` / `edge-on +0.4deg` — the front-most facet's normal put
 *     89.6 and 90.4 degrees from the camera axis. THE CULL DECISION IS THE
 *     TEST: at 90 degrees exactly `backface-visibility` flips, and these two
 *     poses straddle it by less than half a degree on either side, so a facet
 *     whose cull decision has gone the wrong way is one of them. What they
 *     cannot show is that facet ITSELF: a facet within half a degree of edge-on
 *     projects to a sliver a pixel or two wide, which the seam exclusions eat
 *     whichever way it went. What they do show is everything AROUND it — that
 *     the solid stays closed and correctly sorted while one of its facets sits
 *     on the threshold, which is the failure that would actually be visible.
 *     Whether culling works at all, as opposed to at the boundary, is what the
 *     poses below are for.
 *   - `twin-depth ...` — the bisector of a face's normal and its most nearly
 *     parallel neighbour's aimed straight at the camera, which makes those two
 *     facets exactly equally tilted and therefore at very nearly the same
 *     depth, while sharing an edge. That is the configuration a plane-sorting
 *     renderer without a z-buffer has to guess at, so it is where a solid that
 *     was relying on sorting rather than on culling would come apart. Two of
 *     them per shape: one on face 0, one on the face halfway through the
 *     surface — which on the prism is deliberately one of each kind, a flat cap
 *     and a side wall, since a cap-and-wall pair meets at 90 degrees and a
 *     wall-and-wall pair at 30.
 *   - `random N` — three uniformly random rotations from fixed seeds. Fixed,
 *     because a flaky render-truth test is worse than none; three, because they
 *     are insurance and not the argument.
 */
function posesFor(polygons: readonly (readonly V3[])[]): Pose[] {
  const normals = faceNormals(polygons);
  const camera: V3 = [0, 0, 1];

  const tiltFaceTo = (index: number, degrees: number): M3 => {
    const normal = normals[index];
    // Rotate about the axis perpendicular to both, which moves the normal
    // straight towards (or past) the camera axis by the angle asked for. When
    // the face already points at the camera — a prism's cap does, exactly —
    // that axis is degenerate and ANY perpendicular will do; without this
    // branch the pose silently comes back as the identity and the edge-on
    // cases quietly stop being tested.
    let axis = cross(normal, camera);
    if (Math.hypot(axis[0], axis[1], axis[2]) < 1e-9) {
      axis = cross(normal, Math.abs(normal[0]) < 0.9 ? [1, 0, 0] : [0, 1, 0]);
    }
    const current = Math.acos(Math.max(-1, Math.min(1, dot(normal, camera))));
    return rotationAbout(axis, current - (degrees * Math.PI) / 180);
  };

  const twinDepth = (index: number): M3 => {
    let partner = -1;
    let bestDot = -Infinity;
    for (let j = 0; j < normals.length; j++) {
      if (j === index) continue;
      const d = dot(normals[index], normals[j]);
      if (d > bestDot) { bestDot = d; partner = j; }
    }
    const bisector = norm([
      normals[index][0] + normals[partner][0],
      normals[index][1] + normals[partner][1],
      normals[index][2] + normals[partner][2],
    ]);
    // A fixed roll about the camera axis, composed AFTER the aim. It changes no
    // facet's depth — a rotation about the view direction cannot — so the pose
    // is still exactly the twin-depth one, but the shared edge no longer runs
    // down a pixel column. Without it a d20's twin-depth pose is bit-identical
    // to its rest pose (the d20 is built with two faces already straddling the
    // camera axis), which spends one of eight poses proving the same thing
    // twice.
    return mulM3(rotationAbout(camera, (7 * Math.PI) / 180), rotationBetween(bisector, camera));
  };

  // Tilt the facet that STARTS square-on to the camera, not face 0. Face 0 on a
  // d6 or a d20 already sits near 90 degrees at rest, so tilting it to 89.6
  // moves the whole solid by four tenths of a degree and the pose is a
  // duplicate of `rest`. Taking the front-most facet instead makes the pose a
  // large rotation that happens to leave one facet balanced on the cull
  // threshold, which is both a distinct pose and the case being asked about.
  let frontMost = 0;
  for (let i = 1; i < normals.length; i++) {
    if (dot(normals[i], camera) > dot(normals[frontMost], camera)) frontMost = i;
  }

  const halfway = Math.floor(polygons.length / 2);
  return [
    { name: 'rest', matrix: IDENTITY_M3.slice() },
    { name: 'edge-on -0.4deg (just front-facing)', matrix: tiltFaceTo(frontMost, 89.6) },
    { name: 'edge-on +0.4deg (just back-facing)', matrix: tiltFaceTo(frontMost, 90.4) },
    { name: 'twin-depth on face 0', matrix: twinDepth(0) },
    { name: `twin-depth on face ${halfway}`, matrix: twinDepth(halfway) },
    // Composed with the twin-depth pose, so a random rotation still lands on a
    // solid that is not trivially axis-aligned rather than merely elsewhere.
    { name: 'random 1', matrix: seededRotation(0x5eed01) },
    { name: 'random 2', matrix: mulM3(seededRotation(0x5eed02), twinDepth(0)) },
    { name: 'random 3', matrix: seededRotation(0x5eed03) },
  ];
}

// ---------------------------------------------------------------------------
// The measurement.
// ---------------------------------------------------------------------------

function unpack(base64: string): Uint8Array {
  return new Uint8Array(Buffer.from(base64, 'base64'));
}

interface PoseMeasurement extends Verdict {
  readonly pose: string;
}

/**
 * Render every pose of one shape twice — once in the browser, once against a
 * z-buffer — and return the diff of each.
 *
 * `mispaint` is the positive control: it permutes which colour each facet is
 * painted with, changing nothing else at all. The reference and the exclusions
 * are computed identically either way, so the difference between the two runs
 * is exactly the harness's own sensitivity.
 */
async function measureShape(
  page: Page,
  shape: ShapeSpec,
  options: { mispaint?: boolean } = {},
): Promise<{ measurements: PoseMeasurement[]; facetCount: number; paletteDistance: number }> {
  const prepared = await page.evaluate(
    ({ kind, count, heightRatio }) =>
      (window as any).__renderTruth.prepare(kind, count, heightRatio) as PreparedShape,
    { kind: shape.kind, count: shape.count, heightRatio: shape.heightRatio ?? 1 },
  );
  const polygons = prepared.polygons as V3[][];
  const facetCount = polygons.length;
  const palette = paletteFor(facetCount);
  const paletteDistance = minPaletteDistance(palette);
  const depthPx = prepared.perspectiveDieSizes * DIE_PX;

  await page.evaluate((depth) => { (window as any).__renderTruth.depthEm = depth; },
    prepared.perspectiveDieSizes);

  // Which palette slot each facet is actually painted with. Identity is honest;
  // the control shifts every facet onto a different facet's colour.
  const painted = polygons.map((_, i) => (options.mispaint ? ((i + 1) % facetCount) + 1 : i + 1));
  const colours = painted.map((slot) => `rgb(${palette[slot].join(',')})`);
  const background = `rgb(${palette[0].join(',')})`;

  const measurements: PoseMeasurement[] = [];
  for (const pose of posesFor(polygons)) {
    const rect = await page.evaluate(
      ({ pose: poseCss, colours: fills, background: bg }) =>
        (window as any).__renderTruth.render(poseCss, fills, bg),
      { pose: matrix3dOf(pose.matrix), colours, background },
    );
    expect(
      { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
      'the stage must sit at the viewport origin at its declared size, or the clip is wrong',
    ).toEqual({ x: 0, y: 0, width: STAGE_PX, height: STAGE_PX });

    const shot = (await page.screenshot({
      clip: { x: 0, y: 0, width: STAGE_PX, height: STAGE_PX },
    })).toString('base64');
    const classified = await page.evaluate(
      ({ base64, palette: entries }) => (window as any).__renderTruth.classify(base64, entries),
      { base64: shot, palette },
    );
    expect(
      [classified.width, classified.height],
      'the screenshot must be 1 device pixel per CSS pixel',
    ).toEqual([STAGE_PX, STAGE_PX]);

    const expected = referenceRender(
      poseSurface(polygons, prepared.nominalRadius, pose.matrix, depthPx), depthPx);
    measurements.push({
      pose: pose.name,
      ...compare(expected, unpack(classified.classes), unpack(classified.distances)),
    });
  }
  await page.evaluate(() => (window as any).__renderTruth.teardown());
  return { measurements, facetCount, paletteDistance };
}

function report(label: string, measurements: readonly PoseMeasurement[]): string {
  return [`${label}:`, ...measurements.map((m) =>
    `  ${m.pose}: thickest ${m.thickest}px (${m.thickestBox}), wrong ${m.wrong}px`
    + ` = ${(m.wrongFractionOfSilhouette * 100).toFixed(3)}% of a ${m.silhouette}px silhouette`
    + ` (${(m.wrongFractionOfJudged * 100).toFixed(3)}% of ${m.judged}px judged),`
    + ` worst colour distance ${m.worstColourDistance}`
    + (m.samples.length ? `\n      ${m.samples.join('\n      ')}` : ''))].join('\n');
}

// ---------------------------------------------------------------------------

test.describe('a solid renders what a z-buffer would', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await installHarness(page, STAGE_PX, DIE_PX);
  });

  /**
   * THE NEGATIVE CONTROL, and the only reason any other number in this file can
   * be believed.
   *
   * A d20 is convex and its 20 facets are all the same size, which makes it the
   * shape with the most nearly-equal-depth neighbours in the dice set and the
   * one `die-shape.spec.ts` measured as the WORST affected by the culling tear.
   * It is also rendered correctly, necessarily, because culling a closed convex
   * surface is exactly right. So a d20 is the shape on which any error this
   * harness reports is the harness's own — a camera that does not match CSS's,
   * a normalization applied twice, an axis flipped — and it is checked at zero
   * wrong pixels rather than at zero thick regions for exactly that reason. The
   * original experiment measured 0 across 253 poses; if this ever reports a
   * non-trivial number, fix the harness before believing anything it says about
   * a shape whose rendering is actually in question.
   */
  test('a d20 is rendered exactly', async ({ page }) => {
    test.setTimeout(180000);
    const { measurements, facetCount, paletteDistance } = await measureShape(page, { kind: 'die', count: 20 });
    expect(facetCount, 'a d20 is 20 facets').toBe(20);
    expect(paletteDistance, 'facet colours must be separable').toBeGreaterThan(MIN_PALETTE_DISTANCE);
    const summary = report('d20', measurements);
    console.log(summary);
    for (const m of measurements) {
      expect(m.silhouette, `${m.pose}: judged too little to mean anything\n${summary}`)
        .toBeGreaterThan(MIN_JUDGED_SILHOUETTE_PX);
      expect(m.worstColourDistance, `${m.pose}: a judged pixel was not a flat facet fill\n${summary}`)
        .toBeLessThanOrEqual(MAX_JUDGED_COLOUR_DISTANCE);
      expect(m.thickest, `${m.pose}: a thick wrong region\n${summary}`).toBe(0);
    }
    // The strong form, which only the negative control gets to claim.
    const totalWrong = measurements.reduce((sum, m) => sum + m.wrong, 0);
    expect(totalWrong, `a d20 must diff to nothing at all\n${summary}`).toBe(0);
  });

  /**
   * A d6 is the shape every other check in this suite calls the easy one — its
   * facets are squares, so its facet boxes never overhang a neighbour, and it
   * is the only shape that never tore. That makes it the shape whose 1px
   * hairline residual on shared edges is cleanest to see, and therefore the one
   * that proves the erosion is doing its job rather than the tolerance being
   * loose: the hairline is measured here, and it is required not to survive.
   */
  test('a d6 is rendered exactly', async ({ page }) => {
    test.setTimeout(180000);
    const { measurements, facetCount } = await measureShape(page, { kind: 'die', count: 6 });
    expect(facetCount, 'a d6 is 6 facets').toBe(6);
    const summary = report('d6', measurements);
    console.log(summary);
    for (const m of measurements) {
      expect(m.silhouette, `${m.pose}: judged too little to mean anything\n${summary}`)
        .toBeGreaterThan(MIN_JUDGED_SILHOUETTE_PX);
      expect(m.worstColourDistance, `${m.pose}: a judged pixel was not a flat facet fill\n${summary}`)
        .toBeLessThanOrEqual(MAX_JUDGED_COLOUR_DISTANCE);
      expect(m.thickest, `${m.pose}: a thick wrong region\n${summary}`).toBe(0);
    }
  });

  /**
   * THE ONE THAT MATTERS GOING FORWARD.
   *
   * A 12-side flat-capped prism is the shape `boardgame-token`'s `token`,
   * `chip` and `disc` become, at 12 sides because that is the facet budget's
   * ceiling. It is the first solid in this pipeline that was NOT built by
   * `die-geometry.ts`, and it is the only one whose facets come in two
   * radically different kinds — twelve narrow walls meeting at 30 degrees, and
   * two flat caps meeting each wall at 90 — so it exercises the facet placement
   * on aspect ratios a die never produces. Its bounding radius also exceeds its
   * nominal one by 1.14x, unlike every closed-form die, which is precisely the
   * kind of difference a renderer quietly assuming a circumsphere would get
   * wrong.
   */
  test('a 12-side flat-capped prism is rendered exactly', async ({ page }) => {
    test.setTimeout(180000);
    const { measurements, facetCount } = await measureShape(
      page, { kind: 'prism', count: 12, heightRatio: 0.55 });
    expect(facetCount, 'a 12-side prism is 14 facets').toBe(14);
    const summary = report('prism-12 @ 0.55', measurements);
    console.log(summary);
    for (const m of measurements) {
      expect(m.silhouette, `${m.pose}: judged too little to mean anything\n${summary}`)
        .toBeGreaterThan(MIN_JUDGED_SILHOUETTE_PX);
      expect(m.worstColourDistance, `${m.pose}: a judged pixel was not a flat facet fill\n${summary}`)
        .toBeLessThanOrEqual(MAX_JUDGED_COLOUR_DISTANCE);
      expect(m.thickest, `${m.pose}: a thick wrong region\n${summary}`).toBe(0);
    }
  });

  /**
   * THE POSITIVE CONTROL. Without it, everything above is unfalsifiable.
   *
   * The three tests above pass when the browser agrees with the z-buffer. They
   * would ALSO pass if the exclusions grew until nothing was judged, if the
   * classifier collapsed every colour onto one entry, if the reference silently
   * returned all-background, or if the screenshot were blank — every one of
   * which is a way this file could become a test that cannot fail. The floor on
   * the judged silhouette closes some of those; this closes the rest, by
   * breaking the ONE thing under test and requiring the harness to notice.
   *
   * Every facet is painted with the next facet's colour. Geometry, camera,
   * poses, reference and exclusions are byte-for-byte the ones the honest run
   * uses, so a mis-painted facet is indistinguishable from a facet drawn where
   * another facet should have been — which is what a sort error looks like.
   * Nearly every judged silhouette pixel must light up, and the thick-region
   * measure — the number the other three assert on — must be enormous.
   */
  test('a deliberately mis-painted solid is caught', async ({ page }) => {
    test.setTimeout(180000);
    const { measurements } = await measureShape(page, { kind: 'die', count: 20 }, { mispaint: true });
    const summary = report('d20 mis-painted', measurements);
    console.log(summary);
    for (const m of measurements) {
      expect(m.wrongFractionOfSilhouette, `${m.pose}: the harness missed a mis-painted facet\n${summary}`)
        .toBeGreaterThan(0.9);
      expect(m.thickest, `${m.pose}: the thick-region measure missed a mis-painted facet\n${summary}`)
        .toBeGreaterThan(1000);
    }
  });
});
