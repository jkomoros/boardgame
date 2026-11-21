/**
 * Reading a die and painting its outcome onto whichever face landed up.
 *
 * Pure arithmetic over `die-geometry.ts`: no DOM, no packages, no randomness.
 *
 * The server decides a die's value before any animation runs. Rejection
 * sampling — re-simulating until the physics happens to agree — costs
 * `sides^dice` simulations in expectation (~8,000 for three d20s) and is
 * unbounded, so it is not used. Instead the simulation runs ONCE, this module
 * works out which face it presented, and `assignFaceValues` paints the desired
 * value onto that face while permuting the rest into a still-legitimate die.
 * That is why `assignFaceValues` must be solvable for EVERY (presented face,
 * desired value) pair rather than merely usually: an unsolvable pair is a die
 * that cannot show a particular number.
 */

import {
  add,
  cross,
  dot,
  magnitude,
  scale,
  subtract,
  vec3,
  type DieGeometry,
  type Quat,
  type Vec3,
} from './die-geometry.ts';

/**
 * Which way is up in the world frame these functions read orientations in.
 *
 * +Y, matching the physics frame the simulator and the trajectory bake use.
 * Note that CSS's screen Y points DOWN, so a renderer composing a transform
 * from a body-frame pose has to flip Y itself; nothing here does it for it.
 */
export const WORLD_UP: Vec3 = vec3(0, 1, 0);

/**
 * How a resting die is read.
 *
 * `die-geometry.ts` reports only `'up-face' | 'top-vertex'` because it knows
 * the solid and not the convention. `'down-face'` is resolved here, for the
 * solids that have no up face at all — see `resolveReadingRule`.
 */
export type DieReadingRule = 'up-face' | 'down-face' | 'top-vertex';

/** Two normals count as opposite when they cancel to within this. */
const ANTIPODAL_TOLERANCE = 1e-9;

/**
 * A frame of three unit normals is treated as degenerate — carrying no
 * handedness to preserve — below this triple product. A barrel's side normals
 * are all perpendicular to its axis, so every triple of them is exactly 0.
 */
const CHIRALITY_TOLERANCE = 1e-9;

function conjugate(q: Quat): Quat {
  return [-q[0], -q[1], -q[2], q[3]];
}

/** Tolerate a drifted quaternion: an integrator's output is never exactly unit. */
function unitQuat(q: Quat): Quat {
  const length = Math.sqrt(q[0] * q[0] + q[1] * q[1] + q[2] * q[2] + q[3] * q[3]);
  if (!(length > 0)) throw new Error('cannot read a die from a zero-length quaternion');
  return [q[0] / length, q[1] / length, q[2] / length, q[3] / length];
}

/** Rotate `v` by the unit quaternion `q`. */
function rotate(q: Quat, v: Vec3): Vec3 {
  const axis = vec3(q[0], q[1], q[2]);
  const t = scale(cross(axis, v), 2);
  return add(add(v, scale(t, q[3])), cross(axis, t));
}

/**
 * For each face, the index of the face directly opposite it — or `null` when
 * some face has none.
 *
 * A d4 and every odd-sided barrel fail: a tetrahedron's normals point at its
 * vertices, and an odd barrel's side normals are spaced at angles that never
 * add to a half turn. Everything else this codebase builds (d6, d8, d10, d12,
 * d20 and even-sided barrels) pairs exactly.
 *
 * Face order is construction order, so the pairing has to be searched for; it
 * is never index adjacency.
 */
export function antipodalFacePairs(geometry: DieGeometry): readonly number[] | null {
  const count = geometry.faces.length;
  const partner = new Array<number>(count).fill(-1);
  for (let i = 0; i < count; i++) {
    if (partner[i] !== -1) continue;
    let found = -1;
    for (let j = i + 1; j < count; j++) {
      if (partner[j] !== -1) continue;
      const residual = magnitude(add(geometry.faces[i].normal, geometry.faces[j].normal));
      if (residual <= ANTIPODAL_TOLERANCE) {
        found = j;
        break;
      }
    }
    if (found === -1) return null;
    partner[i] = found;
    partner[found] = i;
  }
  return Object.freeze(partner);
}

/**
 * Which reading convention this solid actually admits.
 *
 * The geometry module reports `'top-vertex'` for the d4 and `'up-face'` for
 * everything else, which is right as far as it goes but overstates the
 * `'up-face'` case: an ODD-sided barrel resting on a side face points an EDGE
 * at the ceiling, not a face, and its two best up-face candidates score
 * identically to within 1e-16 — a coin flip decided by floating-point noise,
 * and a different face on the next roll for no visible reason.
 *
 * Such a die has a perfectly unambiguous DOWN face (it is resting on it), and
 * physical odd-sided barrel dice are read from below for exactly this reason,
 * so that is the rule used. The test suite pins both halves of the fact: the
 * up-face tie is real, and the down face wins by a wide margin.
 *
 * The distinction is drawn from the geometry (does every face have an
 * antipode?) rather than from the face count, so a new solid gets the right
 * rule without being added to a list here. It would be better still for
 * `die-geometry.ts` to report `'down-face'` itself; see the task report.
 */
export function resolveReadingRule(geometry: DieGeometry): DieReadingRule {
  if (geometry.readingRule === 'top-vertex') return 'top-vertex';
  return antipodalFacePairs(geometry) === null ? 'down-face' : 'up-face';
}

/** The face whose normal is most aligned with `direction` (ties: lowest index). */
function extremeFace(geometry: DieGeometry, direction: Vec3): number {
  let best = 0;
  let bestScore = -Infinity;
  for (let i = 0; i < geometry.faces.length; i++) {
    const score = dot(geometry.faces[i].normal, direction);
    if (score > bestScore) {
      bestScore = score;
      best = i;
    }
  }
  return best;
}

/**
 * The d4 rule. A tetrahedron resting on a face has three faces tilted equally
 * upward — there is no up face to pick — but exactly one strictly highest
 * vertex, which is what a real d4 is read from.
 *
 * A vertex of a tetrahedron determines one face: the one it does not touch.
 * Every vertex of a convex solid lies on or INSIDE every face plane, so the
 * face a vertex does not touch is the one it is farthest inside — the minimum
 * signed distance, not the maximum. (For the d4 the three faces the apex sits
 * on measure 0 and the fourth measures -2.31.)
 *
 * That fourth face is also the face the die is resting on, so the value ends up
 * on the hidden underside; a top-read d4 prints it at the apex corner of the
 * three visible faces instead. Task 8's renderer has to make that choice, and
 * the report flags it.
 */
function faceOppositeTopVertex(geometry: DieGeometry, bodyUp: Vec3): number {
  let apex = geometry.vertices[0];
  let apexHeight = -Infinity;
  for (const vertex of geometry.vertices) {
    const height = dot(vertex, bodyUp);
    if (height > apexHeight) {
      apexHeight = height;
      apex = vertex;
    }
  }
  let best = 0;
  let bestDistance = Infinity;
  for (let i = 0; i < geometry.faces.length; i++) {
    const face = geometry.faces[i];
    const distance = dot(face.normal, subtract(apex, face.centroid));
    if (distance < bestDistance) {
      bestDistance = distance;
      best = i;
    }
  }
  return best;
}

/**
 * Which face index a die in `orientation` is READ FROM — which is not always a
 * face a viewer can see.
 *
 * For a d6/d8/d10/d12/d20 and every even-sided barrel the returned face points
 * up, as you would expect. For a d4 and every ODD-SIDED BARREL it is the face
 * the die is RESTING ON: those solids present no single face upward (see
 * `resolveReadingRule` and `faceOppositeTopVertex`), and reading from below is
 * the only unambiguous rule. `resolveReadingRule(geometry)` distinguishes the
 * cases: `'up-face'` is visible, `'down-face'` and `'top-vertex'` are not.
 *
 * A renderer must not assume it can print the value at the centre of the
 * returned face: on those solids that centre is face-down against the table.
 * A physical d4 prints its value at the apex corner of the three faces that ARE
 * visible, and a physical odd barrel prints it beside the resting face; a
 * renderer that draws one glyph per face centre has to make the same choice, or
 * the die will land showing nothing.
 *
 * Rather than rotating every normal into the world frame, the world up vector
 * is rotated once into the body frame by the inverse orientation; the two are
 * identical (`dot(R n, up) === dot(n, R⁻¹ up)`) and this is O(1).
 */
export function presentedFaceIndex(geometry: DieGeometry, orientation: Quat): number {
  if (geometry.faces.length === 0) throw new Error('die geometry has no readable faces');
  const bodyUp = rotate(conjugate(unitQuat(orientation)), WORLD_UP);
  switch (resolveReadingRule(geometry)) {
    case 'top-vertex':
      return faceOppositeTopVertex(geometry, bodyUp);
    case 'down-face':
      return extremeFace(geometry, scale(bodyUp, -1));
    default:
      return extremeFace(geometry, bodyUp);
  }
}

/**
 * Pair the values so each pair sums to `min + max`, or `null` when they do not
 * admit such a pairing. Consecutive runs (1..6, 0..9, 1..20) always do; an
 * arbitrary author-supplied face set may not, and then there is no sum rule to
 * honour and any bijection will do.
 */
function complementaryValuePairs(
  values: readonly number[],
): readonly (readonly [number, number])[] | null {
  if (values.length === 0 || values.length % 2 !== 0) return null;
  const ascending = [...values].sort((a, b) => a - b);
  const target = ascending[0] + ascending[ascending.length - 1];
  const pairs: [number, number][] = [];
  for (let i = 0; i < ascending.length / 2; i++) {
    const low = ascending[i];
    const high = ascending[ascending.length - 1 - i];
    if (low + high !== target) return null;
    pairs.push([low, high]);
  }
  // Ascending by low value: pairs[i] is the pair holding the i'th lowest value.
  return pairs;
}

/**
 * The values in the order the caller GAVE them — not sorted — with
 * `desiredValue` pulled out, placed on `presented`, and the rest laid down
 * across the remaining faces in that same given order.
 */
function plainBijection(
  values: readonly number[],
  presented: number,
  desiredValue: number,
): number[] {
  const remaining = [...values];
  remaining.splice(remaining.indexOf(desiredValue), 1);
  const result = new Array<number>(values.length);
  result[presented] = desiredValue;
  let next = 0;
  for (let i = 0; i < values.length; i++) {
    if (i !== presented) result[i] = remaining[next++];
  }
  return result;
}

/** Distinct face indices carrying `wanted`, in order; null if they are not all there. */
function facesCarrying(values: readonly number[], wanted: readonly number[]): number[] | null {
  const taken = new Set<number>();
  const found: number[] = [];
  for (const value of wanted) {
    const index = values.findIndex((candidate, i) => candidate === value && !taken.has(i));
    if (index < 0) return null;
    taken.add(index);
    found.push(index);
  }
  return found;
}

/** Where one complementary value pair was placed. */
interface Placement {
  /** The antipodal face indices holding its low and high value. */
  readonly low: number;
  readonly high: number;
  /** True for the pair holding the desired value, whose placement is forced. */
  readonly forced: boolean;
}

/**
 * A labelling in which opposite faces sum to `min + max` and — where the solid
 * has a handedness to have — the three lowest values wind the way a Western
 * die's 1, 2, 3 do.
 */
function standardArrangement(
  geometry: DieGeometry,
  opposite: readonly number[],
  valuePairs: readonly (readonly [number, number])[],
  presented: number,
  desiredValue: number,
): number[] {
  const count = geometry.faces.length;
  const facePairs: [number, number][] = [];
  for (let i = 0; i < count; i++) {
    if (opposite[i] > i) facePairs.push([i, opposite[i]]);
  }

  const forcedPairIndex = valuePairs.findIndex(
    ([low, high]) => low === desiredValue || high === desiredValue,
  );
  const [forcedLow, forcedHigh] = valuePairs[forcedPairIndex];

  const result = new Array<number>(count);
  const placements: Placement[] = [];

  result[presented] = desiredValue;
  result[opposite[presented]] = desiredValue === forcedLow ? forcedHigh : forcedLow;
  placements.push({ low: presented, high: opposite[presented], forced: true });

  // Everything else is unconstrained, so pair up in index order: deterministic,
  // and with no floating-point comparison anywhere in the choice. For a d6 the
  // sum rule plus the chirality fix below already force the result to be one of
  // the 24 rotations of a standard die, whichever order this loop runs in.
  const freeFacePairs = facePairs.filter(([a, b]) => a !== presented && b !== presented);
  const freePairIndices = valuePairs
    .map((_, index) => index)
    .filter((index) => index !== forcedPairIndex);
  for (let k = 0; k < freeFacePairs.length; k++) {
    const [a, b] = freeFacePairs[k];
    const pairIndex = freePairIndices[k];
    const [low, high] = valuePairs[pairIndex];
    result[a] = low;
    result[b] = high;
    placements.push({ low: a, high: b, forced: false });
  }

  // Chirality. `valuePairs[i]` holds the i'th lowest value, so the three lowest
  // values live in valuePairs 0, 1 and 2. Swapping the two faces of any one of
  // those three pairs moves exactly one of the three onto the opposite (hence
  // negated) normal, which inverts the triple product and nothing else: the
  // pair still occupies the same two antipodal faces, so the sum rule survives.
  //
  // The first non-forced placement is always such a pair. Free placements are
  // pushed in ascending pair index, and at most one pair is forced, so the
  // first free one holds valuePairs[0] or valuePairs[1].
  if (valuePairs.length >= 3) {
    const lowest = [valuePairs[0][0], valuePairs[1][0], valuePairs[2][0]];
    const carriers = facesCarrying(result, lowest);
    const flip = placements.find((placement) => !placement.forced);
    if (carriers && flip) {
      const [a, b, c] = carriers.map((index) => geometry.faces[index].normal);
      // A degenerate frame has no handedness to preserve. Every triple of a
      // barrel's side normals is exactly coplanar, so this is not hypothetical.
      if (dot(a, cross(b, c)) < -CHIRALITY_TOLERANCE) {
        const swap = result[flip.low];
        result[flip.low] = result[flip.high];
        result[flip.high] = swap;
      }
    }
  }

  return result;
}

/**
 * Values for every face index, such that face `presented` carries
 * `desiredValue` and the whole array is a permutation of `faces`.
 *
 * Where the solid pairs its faces antipodally AND the values pair up to
 * `min + max` (both true of a standard d6/d8/d10/d12/d20 with a consecutive
 * value run), opposite faces sum to `min + max` and the arrangement is a real
 * die: for a d6 specifically it is always one of the 24 rotations of the
 * Western standard die, right-handed, never its mirror. Otherwise — a d4, an
 * odd-sided barrel, an author-supplied value set with no complement — there is
 * no convention to honour and any bijection is correct.
 *
 * Solvable for every (presented, desiredValue) pair by construction: the
 * desired value is placed first and the rest is filled in around it.
 */
export function assignFaceValues(
  geometry: DieGeometry,
  faces: readonly number[],
  presented: number,
  desiredValue: number,
): readonly number[] {
  const count = geometry.faces.length;
  if (faces.length !== count) {
    throw new Error(
      `face count mismatch: geometry has ${count} readable faces, got ${faces.length} values`,
    );
  }
  if (!Number.isInteger(presented) || presented < 0 || presented >= count) {
    throw new Error(`presented face ${presented} is not a face index of a d${count}`);
  }
  if (!faces.includes(desiredValue)) {
    throw new Error(`desired value ${desiredValue} is not one of this die's face values`);
  }

  const opposite = antipodalFacePairs(geometry);
  const valuePairs = complementaryValuePairs(faces);
  if (opposite && valuePairs) {
    return Object.freeze(
      standardArrangement(geometry, opposite, valuePairs, presented, desiredValue),
    );
  }
  return Object.freeze(plainBijection(faces, presented, desiredValue));
}
