import assert from 'node:assert/strict';
import test from 'node:test';
import { cross, dot, magnitude, normalize, subtract, vec3, type Vec3 } from '../motion/die-geometry.ts';
import { CAMERA_AXIS, CSS_AXIS_SIGN, SCREEN_UP, cssNumber, facetBasis, toScreen } from './screen-frame.ts';

/** The determinant of the 3x3 matrix with these columns. */
function determinant(a: Vec3, b: Vec3, c: Vec3): number {
  return dot(a, cross(b, c));
}

/**
 * THE ASSERTION THAT WOULD HAVE CAUGHT THE SHIPPED BUG.
 *
 * The map from the right-handed physics frame to CSS's left-handed one must
 * reverse handedness, i.e. have determinant -1. `toScreen` originally shipped as
 * `(x, -y, -z)` — determinant +1, a proper ROTATION — which rendered the mirror
 * solid and made a die present the face opposite the one the physics turned up.
 * Every unit test in the repo passed while that was live, because none of them
 * looked at this number.
 */
test('the physics-to-CSS map is a reflection, not a rotation', () => {
  const columns = [0, 1, 2].map((axis) =>
    toScreen(vec3(axis === 0 ? 1 : 0, axis === 1 ? 1 : 0, axis === 2 ? 1 : 0)));
  assert.equal(determinant(columns[0], columns[1], columns[2]), -1);
  // Stated the other way: an odd number of axes is negated.
  assert.equal(CSS_AXIS_SIGN.filter((sign) => sign < 0).length % 2, 1);
  assert.equal(CSS_AXIS_SIGN[0] * CSS_AXIS_SIGN[1] * CSS_AXIS_SIGN[2], -1);
});

test('toScreen agrees with CSS_AXIS_SIGN componentwise, and nothing else moves', () => {
  const v = vec3(2, 3, 5);
  assert.deepEqual(toScreen(v), vec3(2 * CSS_AXIS_SIGN[0], 3 * CSS_AXIS_SIGN[1], 5 * CSS_AXIS_SIGN[2]));
  assert.deepEqual(toScreen(v), vec3(2, -3, 5));
  // Exactly one axis flips: x and z are untouched for every input.
  for (const sample of [vec3(1, 0, 0), vec3(0, 0, 1), vec3(-7, 11, -13)]) {
    assert.equal(toScreen(sample)[0], sample[0]);
    assert.equal(toScreen(sample)[2], sample[2]);
  }
});

test('toScreen is its own inverse and preserves lengths', () => {
  for (const v of [vec3(1, 2, 3), vec3(-4, 0, 5), vec3(0.5, -0.25, 0)]) {
    assert.deepEqual(toScreen(toScreen(v)), v);
    assert.equal(magnitude(toScreen(v)), magnitude(v));
  }
});

test('toScreen reverses the winding of a right-handed triple', () => {
  // Any right-handed body triple must come out left-handed on screen; that is
  // the whole content of "CSS is a left-handed frame".
  const a = normalize(vec3(1, 2, 3));
  const b = normalize(cross(a, vec3(0, 0, 1)));
  const c = cross(a, b);
  assert.ok(determinant(a, b, c) > 0.99, 'fixture triple is right handed');
  assert.ok(determinant(toScreen(a), toScreen(b), toScreen(c)) < -0.99);
});

/** Directions to build a basis for, including the two degenerate ones. */
const DIRECTIONS: readonly Vec3[] = [
  SCREEN_UP,
  vec3(0, 1, 0),
  CAMERA_AXIS,
  vec3(0, 0, -1),
  vec3(1, 0, 0),
  normalize(vec3(1, 1, 1)),
  normalize(vec3(-0.32, 0.26, 1)),
  normalize(vec3(0.001, 1, 0.001)),
  normalize(vec3(-3, -4, 5)),
];

test('facetBasis is orthonormal for every direction, degenerate ones included', () => {
  for (const w of DIRECTIONS) {
    const { u, v } = facetBasis(w);
    const where = `w=${JSON.stringify(w)}`;
    assert.ok(Math.abs(magnitude(u) - 1) < 1e-12, `|u| ${where}`);
    assert.ok(Math.abs(magnitude(v) - 1) < 1e-12, `|v| ${where}`);
    assert.ok(Math.abs(dot(u, v)) < 1e-12, `u.v ${where}`);
    assert.ok(Math.abs(dot(u, w)) < 1e-12, `u.w ${where}`);
    assert.ok(Math.abs(dot(v, w)) < 1e-12, `v.w ${where}`);
  }
});

/**
 * `(u, v, w)` becomes the upper 3x3 of the facet's `matrix3d`. A determinant of
 * -1 there would mirror the facet's own content — a numeral drawn backwards —
 * while leaving the solid's silhouette identical, which is precisely the defect
 * a screenshot comparison is worst at seeing.
 */
test('facetBasis is right handed, so a facet is never mirrored', () => {
  for (const w of DIRECTIONS) {
    const { u, v } = facetBasis(w);
    assert.ok(Math.abs(determinant(u, v, w) - 1) < 1e-12, `det ${JSON.stringify(w)}`);
    // Equivalently, and this is the form the doc comment states: v = w x u.
    assert.ok(magnitude(subtract(v, cross(w, u))) < 1e-12);
  }
});

test('facetBasis puts local +y down the screen wherever the facet faces the camera', () => {
  // A facet pointed at the viewer must lay its box out the same way up as the
  // screen, or every numeral on it is rotated.
  const { u, v } = facetBasis(CAMERA_AXIS);
  assert.ok(dot(u, vec3(1, 0, 0)) > 0.99, 'local +x is screen right');
  assert.ok(dot(v, vec3(0, 1, 0)) > 0.99, 'local +y is screen down (CSS y grows downward)');
});

test('cssNumber emits short plain decimals and never a signed zero', () => {
  assert.equal(cssNumber(0), '0');
  assert.equal(cssNumber(-0), '0');
  // A value that rounds to zero must not come back as "-0", which some CSS
  // parsers accept and every diff of generated styles finds noisy.
  assert.equal(cssNumber(-1e-9), '0');
  assert.equal(cssNumber(1), '1');
  assert.equal(cssNumber(0.5), '0.5');
  assert.equal(cssNumber(1 / 3), '0.33333');
  assert.equal(cssNumber(0.577350269), '0.57735');
  // No exponential notation for the magnitudes this pipeline emits.
  for (const value of [1e-5, 1234.56789, -0.000012345]) {
    assert.ok(!/e/i.test(cssNumber(value)), `${value} -> ${cssNumber(value)}`);
  }
});
