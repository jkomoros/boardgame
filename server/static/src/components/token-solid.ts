/**
 * What a 3D `boardgame-token` IS: a shape, a resting pose, a size and ten
 * colours, all as pure functions of the token's own state.
 *
 * The deliberately token-specific half of the solid pipeline, the way
 * `die-face-marks.ts` is the deliberately die-specific half. `src/solid/` knows
 * how to place a polygon and nothing about what the polygon belongs to; this
 * module knows what a `chip` is and nothing about how a polygon is placed.
 * `boardgame-token.ts` is then only DOM.
 *
 * ## There is no 3D in the DOM, and that is the point
 *
 * A die tumbles, so its facets have to be re-projected by the browser every
 * frame and it pays for a live `preserve-3d` scene. A TOKEN'S POSE IS A
 * CONSTANT, so its projection is a constant, so it is done here — once, per
 * `(type, colour)` — and the DOM gets flat, coloured, untransformed polygons.
 *
 * That is a frame-rate decision, not a tidiness one. Chromium promotes every
 * element inside a live `preserve-3d` context to its own composited layer the
 * moment an ANCESTOR transform animates, which is exactly what a stack's FLIP
 * does to a component host on every move. Measured in `pass`: 55 tokens went
 * from 57 composited layers at rest to 1,047 during a move, 88.6 megapixels of
 * layer area, and 30fps against the flat art's 59.6. Projecting here is what
 * makes a token composite like the SVG it replaced. `src/solid/flat-facets.ts`
 * carries the measurements and the three things that all had to go.
 *
 * Back-facing facets are not hidden, they are never built: `backface-visibility`
 * was the whole hidden-surface removal and the same convexity that made it
 * sufficient makes culling here sufficient. A prism draws 6 or 7 elements
 * rather than 14, a cube 3 rather than 6.
 *
 * It is a separate module for one concrete reason: everything here is a pure
 * function of `(type, color)` with no DOM anywhere in it, so `token-solid.test.ts`
 * can run it under `node --test`. That matters more here than anywhere else on
 * this branch, because the property the design calls the biggest risk — the
 * resting pose being a pure function of current state rather than something a
 * node remembers — is exactly the property a pure module makes structural.
 *
 * ## The pose, and why there is only one camera
 *
 * Every shape is posed by the SAME camera: `CAMERA_LEAN_DEGREES` is how far the
 * piece's own up-axis leans towards the viewer, i.e. how high above the board
 * the camera sits. A board mixing a cube with a disc has to read as one scene,
 * so the lean is shared and only the shape's own alignment and spin differ.
 *
 * That is not a formality: the two families are built about DIFFERENT axes. A
 * die's up is body +Y; a prism's is body +Z, because `prism.ts` builds the
 * cross-section facing the camera (see its header). `SHAPES[...].align` is the
 * one rotation that reconciles them, and after it every shape's up-axis lands in
 * exactly the same place on screen — asserted in the tests, because a sign error
 * here shows up as a cube seen from below next to a disc seen from above, which
 * is a picture a person notices and no unit test otherwise would.
 *
 * ## Sizing: drawn extent, never circumsphere
 *
 * A token's `#inner` box IS its drawn extent — 30px by default, and every
 * stack margin, the board's `aspect-ratio: 1` clamp and the FLIP scale ratio key
 * off it, so a token may not reserve space the way `boardgame-die.ts`'s `#scaler`
 * does. The solid is therefore scaled so its POSED, PROJECTED SILHOUETTE fills
 * that box (`fitScale`). Sizing it the die's way instead — nominal sphere to box
 * — would draw every token about 40% smaller than the SVG it replaces, because a
 * cube's face spans only 1/sqrt(3) = 57.7% of its circumsphere.
 *
 * ## Colour: the same filters, arithmetic instead of pixels
 *
 * The flat tokens are red-family SVGs recoloured by CSS `filter`. A generated
 * solid has no art to filter, so the SAME filter chain is applied ARITHMETICALLY
 * to one representative red and the result becomes the solid's base colour. One
 * table, `TOKEN_COLOR_FILTERS`, is the single source of both: `boardgame-token.ts`
 * generates its `#outer.<color> img` rules from it, and `tokenBaseColor` runs it
 * through the filter arithmetic. So a 3D chip and a 2D meeple beside it are the
 * same hue by construction rather than by two lists agreeing.
 * `token-color-parity.spec.ts` pins the arithmetic against the browser's own
 * filter implementation.
 */

import { dieGeometry, dot, normalize, scale as scaleVec, vec3, type Vec3 } from '../motion/die-geometry.ts';
import { type SolidSurface } from '../solid/facet-placement.ts';
import { flatFacetStyle, type FlatPoint } from '../solid/flat-facets.ts';
import { prismSurface } from '../solid/prism.ts';
import { SCREEN_UP, toScreen } from '../solid/screen-frame.ts';

// ---------------------------------------------------------------------------
// The camera.
// ---------------------------------------------------------------------------

/**
 * How high above the board the camera sits, in degrees: 90 is straight down
 * (and draws a disc as a flat polygon with no rim at all), 0 is level with the
 * table (and draws a checkerboard as a line).
 *
 * Chosen by looking, at the sizes these actually render at, which is the only
 * way this could have been chosen. The boards are drawn top-down, which argues
 * for a high camera — but a `disc` is a tenth as thick as it is wide, and a high
 * camera draws its rim at zero pixels and the whole solid as a flat polygon.
 * Rendered side by side at 30, 60 and 120px and on a real checkers board, 65
 * degrees leaves a disc and a chip indistinguishable from the flat art they
 * replace, 58 is marginal at 30px, and 50 puts a legible rim under every prism
 * while still reading as a piece standing on a board rather than one lying on
 * it. A cube at 50 shows its top and two sides, which is also what the
 * isometric `token_cube.svg` draws.
 */
export const CAMERA_LEAN_DEGREES = 50;

/**
 * How far the camera is from the piece, in TOKEN WIDTHS — deliberately relative,
 * where `boardgame-card.ts` uses a flat `perspective: 1000px`.
 *
 * `pass` puts 55 tokens on screen at two different scales (`--component-scale:
 * 0.5` on the unused ones). A fixed pixel depth would foreshorten those two sets
 * differently — the small ones nearly orthographic, the big ones not — so the
 * same piece would be a different shape depending on where it sat. Expressed in
 * token widths, the projection is scale-invariant and `fitScale` can be a
 * constant per shape rather than a measurement per element.
 *
 * The value is the die's (`PERSPECTIVE_DEPTH_DIE_SIZES`), for the plain reason
 * that a die and a token on the same table want the same lens.
 */
export const CAMERA_DEPTH_WIDTHS = 6;

// ---------------------------------------------------------------------------
// Lighting.
// ---------------------------------------------------------------------------

/**
 * Where the light comes from, in CSS space (+x right, +y DOWN, +z at the
 * viewer): above, a little to the left, a little in front.
 *
 * Upper-left is the house convention the authored art states outright —
 * `token_disc.svg` says "Light source: upper-left" and draws its bevel as a
 * gradient from a bright upper-left to a dark lower-right. The solids inherit it
 * so that a 3D disc and a 2D meeple beside it are lit from the same place.
 * (`token_cube.svg`, which is much older Illustrator output, disagrees and lights
 * from the right; the authored assets win.)
 */
const LIGHT: Vec3 = normalize(vec3(-0.3, -0.75, 0.59));

/**
 * Lambert, with a floor: `AMBIENT + DIFFUSE * (normal . light)`, clamped.
 *
 * The two constants are not free. `token_disc.svg`'s bevel gradient runs from
 * #C80000 to #700000 against a #C20000 face, i.e. from 1.03 down to 0.58 of the
 * face's own brightness, and its cap is 1.00 by definition. These reproduce that
 * range on the shape the SVG draws: at the resting pose a disc's cap comes out
 * at 1.02 and the rim the camera can see sweeps from 0.83 on the lit side down
 * to 0.59 at the bottom — 1.02..0.59 against the art's 1.03..0.58. So the solid
 * is shaded like the art it replaces, rather than like a renderer.
 *
 * The floor exists for the shapes that are not discs: a cube presents faces
 * pointing much further from the light than a rim ever does, and without it the
 * darkest face goes to nearly nothing at the small sizes these draw at.
 */
const SHADE_AMBIENT = 0.71;
const SHADE_DIFFUSE = 0.33;
const SHADE_MIN = 0.5;
const SHADE_MAX = 1.08;

/** How bright a facet with this posed, CSS-space outward normal is drawn. */
export function facetShade(posedNormal: Vec3): number {
  const lambert = SHADE_AMBIENT + SHADE_DIFFUSE * dot(posedNormal, LIGHT);
  return Math.min(SHADE_MAX, Math.max(SHADE_MIN, lambert));
}

// ---------------------------------------------------------------------------
// 3x3 rotations, in the CSS frame.
// ---------------------------------------------------------------------------

/** Row-major 3x3: `m[row * 3 + col]`. */
export type Mat3 = readonly number[];

const IDENTITY: Mat3 = Object.freeze([1, 0, 0, 0, 1, 0, 0, 0, 1]);

/** CSS `rotateX(deg)`, exactly: the matrix the browser will build from it. */
function rotateX(degrees: number): Mat3 {
  const a = (degrees * Math.PI) / 180;
  const c = Math.cos(a);
  const s = Math.sin(a);
  return Object.freeze([1, 0, 0, 0, c, -s, 0, s, c]);
}

/** CSS `rotateY(deg)`. */
function rotateY(degrees: number): Mat3 {
  const a = (degrees * Math.PI) / 180;
  const c = Math.cos(a);
  const s = Math.sin(a);
  return Object.freeze([c, 0, s, 0, 1, 0, -s, 0, c]);
}

function multiply(a: Mat3, b: Mat3): Mat3 {
  const out = new Array<number>(9).fill(0);
  for (let row = 0; row < 3; row++) {
    for (let col = 0; col < 3; col++) {
      let sum = 0;
      for (let k = 0; k < 3; k++) sum += a[row * 3 + k] * b[k * 3 + col];
      out[row * 3 + col] = sum;
    }
  }
  return Object.freeze(out);
}

function apply(m: Mat3, v: Vec3): Vec3 {
  return vec3(
    m[0] * v[0] + m[1] * v[1] + m[2] * v[2],
    m[3] * v[0] + m[4] * v[1] + m[5] * v[2],
    m[6] * v[0] + m[7] * v[1] + m[8] * v[2],
  );
}

/*
 * There is deliberately no `matrix3d(pose)` emitter here any more. The pose is
 * never handed to CSS: it is applied in `visibleFacetPolygons`, once, and what
 * reaches the DOM is the result. See `src/solid/flat-facets.ts` for why -- a
 * live 3D transform anywhere in a token's subtree is a thousand composited
 * layers the first time a stack animates the host.
 */

// ---------------------------------------------------------------------------
// The shape catalogue.
// ---------------------------------------------------------------------------

/** The `boardgame-token` types that render as real solids. */
export type TokenSolidShape = 'cube' | 'token' | 'chip' | 'disc';

interface ShapeSpec {
  /**
   * The rotation that carries the shape's own up-axis onto the die frame's up
   * (body +Y), applied BEFORE the spin and the camera lean. Identity for a die,
   * a quarter turn for a prism, whose axis is +Z as built.
   */
  readonly align: Mat3;
  /** A turn about the up-axis, after alignment: which side faces the viewer. */
  readonly spinDegrees: number;
  /** Builds the closed surface. */
  readonly surface: () => SolidSurface;
}

/**
 * Every prism is 12-sided, and that number is now a LOOK, not a budget.
 *
 * It was a budget. Measured in a real game before the flattening, 55 tokens as
 * 24-sided prisms was 1,430 `clip-path`-ed elements and 42.8fps at rest, against
 * 770 and 60fps at 12 sides, and it read as a wall somewhere near 800 elements.
 * That wall was misread: the cost was not the element count, it was that each of
 * those elements sat in a live `preserve-3d` context and took a composited layer
 * of its own the instant a stack animated the host (`src/solid/flat-facets.ts`
 * has the layer measurements). Flattened and culled, 55 chips draw 330 elements
 * and hold 60fps through a move, and doubling the sides would draw about 660 —
 * still an order of magnitude under anything measured to hurt.
 *
 * What DOES bound it is the picture: 12 sides reads as a circle up to about
 * 100px and as a visible dodecagon above it, and the boards this draws for
 * (checkers at ~50px, `pass` at 30 and 60) sit under that. Raising it is a
 * legibility decision now, and should still be made by looking.
 */
export const PRISM_SIDES = 12;

/**
 * Height over diameter, per shape.
 *
 * `token` is the chunky stackable one, `chip` a casino chip, `disc` the flattest
 * — a checkers counter. They are the proportions the SVGs they replace draw:
 * `token_token.svg` is a squat cylinder, `token_chip.svg` a thin one, and
 * `token_disc.svg` a top-down checker whose bevel is a thin ring.
 */
export const SHAPE_HEIGHT_RATIO: Readonly<Record<'token' | 'chip' | 'disc', number>> = Object.freeze({
  token: 0.55,
  chip: 0.13,
  disc: 0.1,
});

const SHAPES: Readonly<Record<TokenSolidShape, ShapeSpec>> = Object.freeze({
  // Zero new geometry: the cube a 3D token renders is exactly the cube a d6 is.
  // The 45-degree spin is what makes it read as a solid rather than as a square:
  // face-on it presents one facet and its silhouette is a rectangle.
  cube: { align: IDENTITY, spinDegrees: 45, surface: () => dieGeometry(6) },
  // A prism is built about +Z, so `rotateX(90deg)` stands it up. The spin is
  // half a side, which puts a flat facet at the bottom of the silhouette rather
  // than a vertex — the difference between a piece resting on its rim and one
  // balanced on a corner.
  token: {
    align: rotateX(90),
    spinDegrees: 180 / PRISM_SIDES,
    surface: () => prismSurface(PRISM_SIDES, SHAPE_HEIGHT_RATIO.token),
  },
  chip: {
    align: rotateX(90),
    spinDegrees: 180 / PRISM_SIDES,
    surface: () => prismSurface(PRISM_SIDES, SHAPE_HEIGHT_RATIO.chip),
  },
  disc: {
    align: rotateX(90),
    spinDegrees: 180 / PRISM_SIDES,
    surface: () => prismSurface(PRISM_SIDES, SHAPE_HEIGHT_RATIO.disc),
  },
});

/** Whether this `type` renders as a solid at all. */
export function isTokenSolidShape(type: string): type is TokenSolidShape {
  return Object.prototype.hasOwnProperty.call(SHAPES, type);
}

/** The closed surface a shape is, for a caller that wants the geometry itself. */
export function tokenSurface(shape: TokenSolidShape): SolidSurface {
  return SHAPES[shape].surface();
}

/** Every facet's outward normal in the CSS frame, already posed, in facet order. */
export function posedNormals(shape: TokenSolidShape): readonly Vec3[] {
  const surface = SHAPES[shape].surface();
  const pose = restingPose(shape);
  return [...surface.faces, ...surface.capFaces]
    .map((face) => apply(pose, toScreen(face.normal)));
}

/**
 * Where the shape's OWN up-axis ends up on screen, in the CSS frame.
 *
 * It is NOT the up-most facet, and that distinction is why this exists: on a
 * leaned prism the back wall tilts further up the screen than the cap does, so
 * "the facet with the smallest y" answers a different question and answers it
 * wrong. `align` is defined as the rotation carrying the shape's up-axis onto
 * screen-up, so undoing it and running the result back through the whole pose
 * gives the direction the camera puts up — the same one for every shape, which
 * is the point.
 */
export function posedUp(shape: TokenSolidShape): Vec3 {
  const align = SHAPES[shape].align;
  // A rotation's inverse is its transpose.
  const inverse = [0, 3, 6, 1, 4, 7, 2, 5, 8].map((i) => align[i]);
  return apply(restingPose(shape), apply(inverse, SCREEN_UP));
}

/**
 * The resting pose, as a matrix in the CSS frame.
 *
 * `rotateX(-lean) * rotateY(spin) * align`, read right to left: stand the shape
 * up, turn it about its own up-axis, then lean the whole thing towards the
 * camera. A pure function of the type and two constants — there is no state to
 * remember and nothing to write down, which is the entire point.
 */
export function restingPose(shape: TokenSolidShape): Mat3 {
  const spec = SHAPES[shape];
  return multiply(
    multiply(rotateX(-CAMERA_LEAN_DEGREES), rotateY(spec.spinDegrees)),
    spec.align,
  );
}

// ---------------------------------------------------------------------------
// Sizing.
// ---------------------------------------------------------------------------

/**
 * How far the camera is from the solid's centre, in the solid's own `em`, when
 * the solid is drawn at `fit` times the token's box.
 *
 * `CAMERA_DEPTH_WIDTHS` is stated in token WIDTHS and `1em` is `fit` widths, so
 * the two differ by exactly `fit` — the same conversion `fitScale` solves its
 * fixed point against, named once so the projection and the sizing cannot drift.
 */
function cameraDepthEm(fit: number): number {
  return CAMERA_DEPTH_WIDTHS / fit;
}

/**
 * Half the projected silhouette's widest half-extent, in the solid's own `em`
 * (where `0.5 / nominalRadius` has normalized the nominal sphere to 1em).
 *
 * Measured from the ORIGIN outwards rather than as a bounding box, because the
 * box is centred on the token's box and a silhouette that is off-centre sticks
 * out on one side only. `depthEm` of `Infinity` gives the orthographic extent.
 */
function projectedHalfExtent(surface: SolidSurface, pose: Mat3, depthEm: number): number {
  const unitsToEm = 0.5 / surface.nominalRadius;
  let widest = 0;
  for (const face of [...surface.faces, ...surface.capFaces]) {
    for (const vertex of face.polygon) {
      const p = apply(pose, scaleVec(toScreen(vertex), unitsToEm));
      if (!(depthEm > p[2])) {
        throw new Error(`token solid: the camera at ${depthEm}em is inside the solid`);
      }
      const magnify = Number.isFinite(depthEm) ? depthEm / (depthEm - p[2]) : 1;
      widest = Math.max(widest, Math.abs(p[0] * magnify), Math.abs(p[1] * magnify));
    }
  }
  return widest;
}

/**
 * How many times the token's box the solid must be drawn at for its silhouette
 * to exactly fill that box — the number `#solid`'s `font-size` multiplies
 * `--component-effective-width` by.
 *
 * It is a fixed point, and it has to be solved rather than evaluated: the
 * silhouette depends on the perspective magnification, the magnification depends
 * on how far away the camera is in the SOLID's units, and that distance depends
 * on how big the solid was drawn. `CAMERA_DEPTH_WIDTHS` is stated in token
 * widths, so `depth_em = CAMERA_DEPTH_WIDTHS / fit`. The map is a strong
 * contraction (the whole perspective term is a few percent), so a handful of
 * iterations converges to double precision; the test asserts the result is
 * self-consistent rather than trusting the loop count.
 */
export function fitScale(shape: TokenSolidShape): number {
  let fit = 1 / silhouetteExtent(shape, 0);
  for (let i = 0; i < 12; i++) fit = 1 / silhouetteExtent(shape, fit);
  return fit;
}

/**
 * The posed, projected silhouette's widest extent in the solid's own `em`, when
 * the solid is drawn at `fit` times the token's box — i.e. with the camera
 * `CAMERA_DEPTH_WIDTHS / fit` of the solid's own em away. A `fit` of 0 puts the
 * camera at infinity, which is the orthographic silhouette.
 *
 * Exported because it is what makes `fitScale`'s fixed point CHECKABLE:
 * `fit * silhouetteExtent(shape, fit)` must be exactly 1, and a test can say so
 * without knowing how many iterations the solver took.
 */
export function silhouetteExtent(shape: TokenSolidShape, fit: number): number {
  const depthEm = fit > 0 ? cameraDepthEm(fit) : Infinity;
  return 2 * projectedHalfExtent(SHAPES[shape].surface(), restingPose(shape), depthEm);
}

// ---------------------------------------------------------------------------
// Colour.
// ---------------------------------------------------------------------------

export type Rgb = readonly [number, number, number];

/**
 * The red every token asset is drawn in, and so the colour a 3D token is before
 * any filter: `#C50000`.
 *
 * The four solid-rendered assets paint their primary face #C70000 (`token`),
 * #C50000 (`cube`), #C20000 (`disc`'s face gradient midpoint) and #C50000
 * (`chip`); this is the median, and it is within 2/255 of every one of them. It
 * matters that it is one of THEIR reds and not a nice red: `meeple` and `pawn`
 * keep their art, so on a board that mixes them the filtered SVG and the shaded
 * solid have to land on the same hue.
 */
export const TOKEN_BASE_RED: Rgb = Object.freeze([0xc5, 0x00, 0x00] as const);

/**
 * The CSS filter that turns the base red into each legal colour — the SINGLE
 * source of truth for it.
 *
 * `boardgame-token.ts` generates its `#outer.<color> img` rules from this table
 * and `tokenBaseColor` evaluates the same strings arithmetically, so the flat
 * art and the solid cannot drift apart. `red` is absent because red is the
 * colour the art is already drawn in, which is why it never had a rule.
 */
export const TOKEN_COLOR_FILTERS: Readonly<Record<string, string>> = Object.freeze({
  gray: 'saturate(0.0) brightness(3.0)',
  green: 'hue-rotate(130deg) brightness(2.0)',
  teal: 'hue-rotate(185deg) brightness(2.4)',
  purple: 'hue-rotate(300deg) brightness(1.0)',
  pink: 'hue-rotate(-93deg) brightness(4) saturate(0.8)',
  blue: 'hue-rotate(220deg) brightness(2.0) saturate(1.5)',
  orange: 'hue-rotate(50deg) brightness(2.5)',
  yellow: 'hue-rotate(70deg) brightness(4)',
  black: 'saturate(0.0) brightness(1.7)',
});

/** The luminance weights every CSS colour-matrix filter is defined against. */
const LUMA: Rgb = Object.freeze([0.213, 0.715, 0.072] as const);

function clampChannel(value: number): number {
  return Math.min(255, Math.max(0, value));
}

function applyMatrix(rgb: Rgb, m: readonly number[]): Rgb {
  return Object.freeze([0, 1, 2].map((row) => clampChannel(
    m[row * 3] * rgb[0] + m[row * 3 + 1] * rgb[1] + m[row * 3 + 2] * rgb[2],
  )) as unknown as Rgb);
}

/** `saturate(s)`, as the filter spec's colour matrix. */
function saturateMatrix(s: number): readonly number[] {
  const [lr, lg, lb] = LUMA;
  return [
    lr + (1 - lr) * s, lg - lg * s, lb - lb * s,
    lr - lr * s, lg + (1 - lg) * s, lb - lb * s,
    lr - lr * s, lg - lg * s, lb + (1 - lb) * s,
  ];
}

/** `hue-rotate(deg)`, as the filter spec's colour matrix. */
function hueRotateMatrix(degrees: number): readonly number[] {
  const a = (degrees * Math.PI) / 180;
  const c = Math.cos(a);
  const s = Math.sin(a);
  return [
    0.213 + c * 0.787 - s * 0.213, 0.715 - c * 0.715 - s * 0.715, 0.072 - c * 0.072 + s * 0.928,
    0.213 - c * 0.213 + s * 0.143, 0.715 + c * 0.285 + s * 0.140, 0.072 - c * 0.072 - s * 0.283,
    0.213 - c * 0.213 - s * 0.787, 0.715 - c * 0.715 + s * 0.715, 0.072 + c * 0.928 + s * 0.072,
  ];
}

/**
 * Run one `filter` shorthand chain over one colour, in sRGB and left to right,
 * clamping after every function — which is what a filter primitive's result
 * buffer does, and what makes `brightness(4)` on a saturated red produce a
 * clipped colour rather than an out-of-gamut one.
 *
 * Only the three functions the token palette uses are implemented, and anything
 * else throws rather than being ignored: a silently skipped function would
 * recolour the flat art and not the solid, which is the exact class of drift
 * this table exists to prevent.
 */
export function applyColorFilter(filter: string, rgb: Rgb): Rgb {
  let out = rgb;
  const pattern = /([a-z-]+)\(\s*(-?[\d.]+)(deg)?\s*\)/g;
  let consumed = 0;
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(filter)) !== null) {
    consumed += match[0].length;
    const amount = Number(match[2]);
    switch (match[1]) {
      case 'brightness':
        out = Object.freeze([0, 1, 2].map((i) => clampChannel(out[i] * amount)) as unknown as Rgb);
        break;
      case 'saturate':
        out = applyMatrix(out, saturateMatrix(amount));
        break;
      case 'hue-rotate':
        out = applyMatrix(out, hueRotateMatrix(amount));
        break;
      default:
        throw new Error(`token colour filter: unsupported function ${match[1]}`);
    }
  }
  if (consumed === 0 && filter.trim() !== '') {
    throw new Error(`token colour filter: could not parse "${filter}"`);
  }
  return out;
}

/**
 * The base colour of a solid token of this colour name.
 *
 * An unknown name — including the empty string, which is what a stack passes for
 * a component whose colour is hidden from this player (checkers does exactly
 * that) — is the unfiltered red, which is precisely what the flat art shows in
 * the same case: no class matches, so no filter applies.
 */
export function tokenBaseColor(color: string): Rgb {
  const filter = TOKEN_COLOR_FILTERS[color.toLowerCase()];
  return filter === undefined ? TOKEN_BASE_RED : applyColorFilter(filter, TOKEN_BASE_RED);
}

// ---------------------------------------------------------------------------
// The whole thing, as the DOM needs it.
// ---------------------------------------------------------------------------

/** One facet, as the complete inline style of the one element that draws it. */
export interface TokenFacet {
  /** Stable key: index into the surface's `[...faces, ...capFaces]`. */
  readonly key: number;
  readonly style: string;
}

/** Everything `boardgame-token.ts` needs to draw one solid, and nothing else. */
export interface TokenSolid {
  /**
   * The FRONT-FACING facets only, already projected. A back-facing one is not
   * hidden here, it is never built: see `visibleFacetPolygons`.
   */
  readonly facets: readonly TokenFacet[];
  /** What `#solid`'s `font-size` multiplies `--component-effective-width` by. */
  readonly fit: number;
}

/**
 * Every facet the camera can see, as its projected outline in `em` from the
 * solid's centre — and NOTHING for the ones it cannot.
 *
 * Two things happen here that used to be the browser's job, and both are only
 * possible because a token's pose is a constant:
 *
 * 1. FACING. `backface-visibility: hidden` culls a facet whose outward normal
 *    has turned away from the eye. The same test, done here: the eye sits at
 *    `(0, 0, depth)` and a facet is drawn exactly when its outward normal points
 *    against the ray from the eye to its centroid. It is the PERSPECTIVE-correct
 *    test — the eye is a point, not a direction — which is what CSS does too, so
 *    the two agree facet for facet.
 *
 * 2. PROJECTION. The perspective divide CSS would apply per frame, applied once:
 *    a point at depth `z` is magnified by `depth / (depth - z)`. This is the
 *    same expression `projectedHalfExtent` sizes the solid with, which is what
 *    keeps the silhouette exactly one box wide.
 *
 * Nothing is sorted, and nothing needs to be: the front-facing facets of a
 * CONVEX solid tile the silhouette exactly once. Every shape that renders as a
 * solid is convex — that is the same precondition the culled 3D version rested
 * on, and it is why `meeple` and `pawn` keep their authored art.
 */
export function visibleFacetPolygons(
  shape: TokenSolidShape,
  fit: number = fitScale(shape),
): readonly { readonly key: number; readonly points: readonly FlatPoint[] }[] {
  const surface = SHAPES[shape].surface();
  const pose = restingPose(shape);
  const unitsToEm = 0.5 / surface.nominalRadius;
  const depthEm = cameraDepthEm(fit);
  const out: { key: number; points: FlatPoint[] }[] = [];
  [...surface.faces, ...surface.capFaces].forEach((face, key) => {
    const normal = apply(pose, toScreen(face.normal));
    const centroid = apply(pose, scaleVec(toScreen(face.centroid), unitsToEm));
    // The ray from the eye to the facet. Front-facing means the outward normal
    // opposes it.
    const facing = normal[0] * centroid[0]
      + normal[1] * centroid[1]
      + normal[2] * (centroid[2] - depthEm);
    if (facing >= 0) return;
    const points = face.polygon.map((vertex) => {
      const p = apply(pose, scaleVec(toScreen(vertex), unitsToEm));
      if (!(depthEm > p[2])) {
        throw new Error(`token solid: the camera at ${depthEm}em is inside the solid`);
      }
      const magnify = depthEm / (depthEm - p[2]);
      return { x: p[0] * magnify, y: p[1] * magnify };
    });
    out.push({ key, points });
  });
  if (out.length === 0) {
    throw new Error(`token solid: ${shape} presents no facet to the camera`);
  }
  return out;
}

/**
 * Building a surface runs a convex hull for the die shapes and trigonometry for
 * the prisms, and the result depends on nothing but the two strings in the key,
 * so it is computed once per `(type, colour)` pair. Bounded by six types times
 * eleven colours; nothing a player can drive touches the key.
 */
const SOLID_CACHE = new Map<string, TokenSolid>();

/**
 * The solid for a token of this type and colour: one style string per VISIBLE
 * facet, and the sizing multiplier.
 *
 * Everything returned is derived; nothing is remembered between calls except as
 * a cache keyed by the same inputs. A caller that renders this on every update —
 * which `boardgame-token.ts` does, from its template — cannot leave a stale pose
 * on a pooled element, because there is no state anywhere in this file for a
 * previous occupant to have written.
 */
export function tokenSolid(shape: TokenSolidShape, color: string): TokenSolid {
  const key = `${shape}|${color.toLowerCase()}`;
  const cached = SOLID_CACHE.get(key);
  if (cached) return cached;

  const surface = SHAPES[shape].surface();
  const pose = restingPose(shape);
  const base = tokenBaseColor(color);
  const fit = fitScale(shape);
  const polygons = [...surface.faces, ...surface.capFaces];
  // How bright this colour can be lit before a channel clips. Shading MULTIPLIES,
  // which is what keeps a 3D blue chip the same blue as a flat blue meeple beside
  // it -- but only while every channel scales by the same number. Orange is
  // (255, 91, 0), so a highlight above 1.0 would clip the red and lift only the
  // green, turning the lit facet yellow. Ceilinged instead, per colour.
  const headroom = 255 / Math.max(base[0], base[1], base[2], 1);
  const facets = visibleFacetPolygons(shape, fit).map(({ key: facetKey, points }) => {
    const shade = Math.min(
      facetShade(normalize(apply(pose, toScreen(polygons[facetKey].normal)))),
      headroom,
    );
    const fill = [0, 1, 2].map((i) => Math.round(clampChannel(base[i] * shade))).join(',');
    return Object.freeze({
      key: facetKey,
      style: `${flatFacetStyle(points)};background:rgb(${fill})`,
    });
  });

  const solid: TokenSolid = Object.freeze({
    facets: Object.freeze(facets),
    fit,
  });
  SOLID_CACHE.set(key, solid);
  return solid;
}
