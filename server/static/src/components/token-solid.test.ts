import assert from 'node:assert/strict';
import test from 'node:test';
import {
  ART_ASSET_ASPECT,
  ART_DEPTH,
  CAMERA_DEPTH_WIDTHS,
  CAMERA_LEAN_DEGREES,
  PRISM_SIDES,
  SHADOW_DIRECTION,
  SHAPE_HEIGHT_RATIO,
  TOKEN_BASE_RED,
  TOKEN_COLOR_FILTERS,
  applyColorFilter,
  artDrawnWidth,
  facetShade,
  fitScale,
  isTokenSolidShape,
  posedNormals,
  posedUp,
  restingPose,
  shadowOffsetEm,
  silhouetteExtent,
  tokenBaseColor,
  tokenSolid,
  tokenSurface,
  type Rgb,
  type TokenSolidShape,
} from './token-solid.ts';

const SHAPES: readonly TokenSolidShape[] = ['cube', 'token', 'chip', 'disc'];

const dot = (a: readonly number[], b: readonly number[]) =>
  a[0] * b[0] + a[1] * b[1] + a[2] * b[2];

/** The style's `background:rgb(r,g,b)`, as numbers. */
function fillOf(style: string): Rgb {
  const match = /background:rgb\((\d+),(\d+),(\d+)\)/.exec(style);
  assert.ok(match, `no background in ${style}`);
  return [Number(match[1]), Number(match[2]), Number(match[3])];
}

// ---------------------------------------------------------------------------
// The facet budget.
// ---------------------------------------------------------------------------

test('only the facets the camera can see are built at all', () => {
  assert.equal(PRISM_SIDES, 12);
  // A closed surface is 6 and 14 polygons; a solid draws the front-facing half.
  assert.equal(tokenSurface('cube').faces.length + tokenSurface('cube').capFaces.length, 6);
  assert.equal(tokenSolid('cube', 'red').facets.length, 3);
  for (const shape of ['token', 'chip', 'disc'] as const) {
    const surface = tokenSurface(shape);
    assert.equal(surface.faces.length + surface.capFaces.length, PRISM_SIDES + 2);
    // One cap plus five walls: at a 50-degree lean with the spin half a side
    // off, six of the twelve walls turn away and so does the far cap.
    assert.equal(tokenSolid(shape, 'red').facets.length, 6, shape);
  }
});

test('a culled facet is one that faces away, and every kept one faces the camera', () => {
  // The cull must agree with the arithmetic that shades: `posedNormals` is in
  // surface order and a facet's key IS that order, so a kept key must name a
  // normal pointing back towards the eye. This is the property that would break
  // silently -- an off-by-one in the cull draws the WRONG six facets, which
  // still looks like a solid and is still six elements.
  for (const shape of SHAPES) {
    const normals = posedNormals(shape);
    const kept = new Set(tokenSolid(shape, 'red').facets.map((facet) => facet.key));
    assert.ok(kept.size > 0, shape);
    normals.forEach((normal, key) => {
      // Orthographically, `normal.z > 0` faces the camera. The real test is
      // perspective-correct so a facet within a couple of degrees of edge-on
      // may fall either way; nothing may be kept that points clearly away, and
      // nothing clearly towards may be dropped.
      if (normal[2] > 0.15) assert.ok(kept.has(key), `${shape} dropped facet ${key}`);
      if (normal[2] < -0.15) assert.ok(!kept.has(key), `${shape} kept back facet ${key}`);
    });
  }
});

test('55 tokens draw no more elements than the flat art they replace cost', () => {
  // The wall this used to be about was a facet-count wall, and it was the wrong
  // wall: the cost was compositor layers, not elements, and a token no longer
  // takes any (see src/solid/flat-facets.ts). What is left is the honest count.
  const worst = Math.max(...SHAPES.map((shape) => tokenSolid(shape, 'red').facets.length));
  assert.equal(worst * 55, 330);
});

// ---------------------------------------------------------------------------
// The pose.
// ---------------------------------------------------------------------------

test('every resting pose is a rotation, never a reflection', () => {
  for (const shape of SHAPES) {
    const m = restingPose(shape);
    // Orthonormal rows.
    for (let i = 0; i < 3; i++) {
      for (let j = 0; j < 3; j++) {
        const expected = i === j ? 1 : 0;
        const product = dot(m.slice(i * 3, i * 3 + 3), m.slice(j * 3, j * 3 + 3));
        assert.ok(Math.abs(product - expected) < 1e-12, `${shape} rows ${i},${j}: ${product}`);
      }
    }
    // A reflection here would mirror every facet's own content and leave the
    // silhouette alone -- the exact defect `screen-frame.ts` documents.
    const determinant =
      m[0] * (m[4] * m[8] - m[5] * m[7])
      - m[1] * (m[3] * m[8] - m[5] * m[6])
      + m[2] * (m[3] * m[7] - m[4] * m[6]);
    assert.ok(Math.abs(determinant - 1) < 1e-12, `${shape} determinant ${determinant}`);
  }
});

test('every shape stands the same way up under one camera', () => {
  // The two families are built about different axes -- a die's up is body +Y, a
  // prism's is body +Z -- so this is the assertion that `align` reconciles them
  // rather than merely existing. A sign error puts a cube's top face where a
  // disc's bottom is, which no other test here would notice.
  //
  // Two halves, and the second is the one with teeth: every shape's own up-axis
  // lands in the same place on screen, AND each shape actually HAS a facet
  // facing that way (the cube's +Y face, the prism's cap). Without the second,
  // an `align` that pointed a prism sideways would still satisfy the first.
  const lean = (CAMERA_LEAN_DEGREES * Math.PI) / 180;
  const expected = [0, -Math.cos(lean), Math.sin(lean)];
  for (const shape of SHAPES) {
    const up = posedUp(shape);
    for (let i = 0; i < 3; i++) {
      assert.ok(
        Math.abs(up[i] - expected[i]) < 1e-9,
        `${shape}'s up-axis points ${up} rather than ${expected}`,
      );
    }
    const facing = posedNormals(shape).some((normal) =>
      Math.abs(normal[0] - up[0]) < 1e-9
      && Math.abs(normal[1] - up[1]) < 1e-9
      && Math.abs(normal[2] - up[2]) < 1e-9);
    assert.ok(facing, `${shape} has no facet facing its own up-axis`);
  }
});

test('the camera leans towards the viewer, so the up face is visible', () => {
  // Sign check, stated as the thing a person would see: the face the pose puts
  // on top of the piece has to face the CAMERA (+z in the CSS frame), not away
  // from it. Leaning the wrong way renders every piece from underneath, and
  // `backface-visibility` then culls the top face out of every one of them.
  for (const shape of SHAPES) {
    assert.ok(posedUp(shape)[2] > 0, `${shape}'s top face points away from the camera`);
    assert.ok(posedUp(shape)[1] < 0, `${shape}'s top face is not at the top`);
  }
});

test('the pose depends on nothing but the shape', () => {
  // `TokenSolid` has no `pose` field and has not had one since the projection
  // moved to build time -- the pose is baked into every facet's already-
  // projected outline. This test compared `.pose` on two solids, so both sides
  // were `undefined` and it asserted nothing; `*.test.ts` is outside
  // tsconfig.json's `include`, so the type checker never saw it either.
  //
  // What must not vary with colour is therefore the GEOMETRY of the facet
  // styles: their boxes, their offsets and their clip paths, i.e. everything
  // but the fill.
  const geometry = (shape: TokenSolidShape, color: string) =>
    tokenSolid(shape, color).facets.map((facet) => facet.style.replace(/;background:.*$/, ''));
  const fills = (shape: TokenSolidShape, color: string) =>
    tokenSolid(shape, color).facets.map((facet) => fillOf(facet.style));
  for (const shape of SHAPES) {
    assert.deepEqual(restingPose(shape), restingPose(shape));
    assert.deepEqual(geometry(shape, 'red'), geometry(shape, 'blue'),
      `${shape} is posed differently depending on its colour`);
    // ...and not vacuous in the other direction: the colours DO differ, so the
    // two solids being compared are genuinely two different solids.
    assert.notDeepEqual(fills(shape, 'red'), fills(shape, 'blue'));
  }
});

// ---------------------------------------------------------------------------
// Sizing.
// ---------------------------------------------------------------------------

test('the fit scale is the exact fixed point of its own silhouette', () => {
  // `fitScale` iterates, because the silhouette depends on the camera distance
  // and the camera distance depends on the size. This says the answer is right
  // without saying how many iterations it took to get there.
  for (const shape of SHAPES) {
    const fit = fitScale(shape);
    const filled = fit * silhouetteExtent(shape, fit);
    assert.ok(Math.abs(filled - 1) < 1e-12, `${shape} fills ${filled} of its box`);
  }
});

test('the perspective actually magnifies, so the fit is not the orthographic one', () => {
  // If this ever came out equal, the camera would have quietly become
  // orthographic and the fixed point above would be vacuously satisfied.
  for (const shape of SHAPES) {
    const orthographic = 1 / silhouetteExtent(shape, 0);
    assert.ok(
      orthographic > fitScale(shape),
      `${shape}: perspective must widen the silhouette, so the fit must shrink`,
    );
    assert.ok(orthographic / fitScale(shape) < 1.2, `${shape}: ${CAMERA_DEPTH_WIDTHS} widths is not a fisheye`);
  }
});

test('a cube is sized by its silhouette, so its own face is smaller than the box', () => {
  // What "size by drawn extent" MEANS, said about the shape that makes it
  // clearest. The silhouette is one box across; a cube's face spans only
  // 1/sqrt(3) = 58% of its circumsphere, so the face is much less than the box
  // -- which is exactly what the isometric `token_cube.svg` draws, and exactly
  // what makes the solid a drop-in for it. Sizing so the FACE filled the box
  // instead would draw the piece 73% larger and hang it out of its own slot.
  const fit = fitScale('cube');
  const surface = tokenSurface('cube');
  const pose = restingPose('cube');
  // The widest pair of vertices across the posed cube, orthographically.
  const unitsToEm = 0.5 / surface.nominalRadius;
  let widest = 0;
  for (const face of surface.faces) {
    for (const vertex of face.polygon) {
      const screen = [vertex[0], -vertex[1], vertex[2]].map((v) => v * unitsToEm);
      const x = dot(pose.slice(0, 3), screen);
      const y = dot(pose.slice(3, 6), screen);
      widest = Math.max(widest, Math.abs(x), Math.abs(y));
    }
  }
  // Drawn at `fit`, the silhouette is one box across (within the perspective
  // term the exact test above pins), and the cube's own face is much less.
  assert.ok(Math.abs(2 * widest * fit - 1) < 0.02, `cube silhouette ${2 * widest * fit}`);
  const faceSpan = (1 / Math.sqrt(3)) * fit;
  assert.ok(faceSpan < 0.6, `a cube's face spans ${faceSpan} of the box, as it must`);
});

test('a taller prism is drawn smaller across, because its silhouette is taller', () => {
  // The token is 0.55 as thick as it is wide and the disc a tenth, so leaning
  // them by the same camera makes the token's silhouette the taller one -- and a
  // fit that ignored the pose would give all three prisms the same number.
  assert.ok(SHAPE_HEIGHT_RATIO.token > SHAPE_HEIGHT_RATIO.chip);
  assert.ok(SHAPE_HEIGHT_RATIO.chip > SHAPE_HEIGHT_RATIO.disc);
  assert.ok(fitScale('token') < fitScale('chip'));
  assert.ok(fitScale('chip') < fitScale('disc'));
});

// ---------------------------------------------------------------------------
// Lighting.
// ---------------------------------------------------------------------------

test('the light comes from above and from the left', () => {
  // Stated as three comparisons rather than as the light vector, so it is a fact
  // about the PICTURE: up beats down, left beats right, and the range matches
  // the bevel the authored disc art draws (1.03 down to 0.58 of its face).
  const up = facetShade([0, -1, 0]);
  const down = facetShade([0, 1, 0]);
  const left = facetShade([-1, 0, 0]);
  const right = facetShade([1, 0, 0]);
  assert.ok(up > down, `up ${up} must be brighter than down ${down}`);
  assert.ok(left > right, `left ${left} must be brighter than right ${right}`);
  assert.ok(up <= 1.08 && down >= 0.5, `the shade range is ${down}..${up}`);
});

test('a shadow falls the opposite way from the light, and mostly downwards', () => {
  // The two families are lit by the SAME light and this is the only place the
  // art half can express it: meeple and pawn have no mesh to shade, so all of
  // their depth is shadow OFFSETS, and every one of them points this way.
  const [x, y] = SHADOW_DIRECTION;
  assert.ok(x > 0, `the light is on the left, so a shadow falls right (${x})`);
  assert.ok(y > 0, `and downwards (${y})`);
  assert.ok(y > x, `and further down than sideways (${y} vs ${x})`);
  assert.ok(Math.abs(Math.hypot(x, y) - 1) < 1e-12, 'it is a unit direction');
  // The same three comparisons facetShade is asserted with, so a light moved in
  // one place and not the other cannot pass both.
  assert.ok(facetShade([-1, 0, 0]) > facetShade([1, 0, 0]));
  assert.ok(facetShade([0, -1, 0]) > facetShade([0, 1, 0]));
});

test('a shadow offset stays parallel to the light at any distance', () => {
  for (const distance of [0.01, ART_DEPTH.edgeEm, ART_DEPTH.groundEm, 3]) {
    const offset = shadowOffsetEm(distance);
    assert.ok(Math.abs(Math.hypot(offset.x, offset.y) - distance) < 1e-12,
      `${distance}em came out ${Math.hypot(offset.x, offset.y)}`);
    assert.ok(Math.abs(offset.x * SHADOW_DIRECTION[1] - offset.y * SHADOW_DIRECTION[0]) < 1e-12,
      'the cross product with the light direction must be zero');
  }
});

test('the tilt is small enough to leave a piece standing', () => {
  // The scene camera sits CAMERA_LEAN_DEGREES above the board, and fully
  // reprojecting a standing piece to it would foreshorten it by cos(50) = 0.64
  // and lay it down -- the mesh work this design declines. The tilt has to be a
  // long way short of that and still do something.
  const flat = Math.cos((CAMERA_LEAN_DEGREES * Math.PI) / 180);
  assert.ok(ART_DEPTH.lean < 1, 'the art is foreshortened at all');
  assert.ok(ART_DEPTH.lean > (1 + flat) / 2,
    `lean ${ART_DEPTH.lean} is nearer lying down than standing (flat is ${flat})`);
});

test('a piece\'s contact shadow is sized off the piece, not off its box', () => {
  // The box is square and an SVG keeps its own proportions inside it, so a pawn
  // is drawn at 0.43 of its box's width. Sized off the box, its shadow came out
  // nearly twice as wide as the pawn.
  assert.equal(artDrawnWidth('meeple'), ART_ASSET_ASPECT.meeple);
  assert.equal(artDrawnWidth('pawn'), ART_ASSET_ASPECT.pawn);
  assert.ok(artDrawnWidth('pawn') < 0.5, 'a pawn is a narrow piece in a square box');
  // A shape at least as tall as it is wide fills the box across, and a solid --
  // which has no asset to letterbox -- is the same case.
  assert.equal(artDrawnWidth('cube'), 1);
});

test('a facing cap is drawn at the flat art\'s own brightness', () => {
  // The cap of a resting prism is what replaces the flat SVG's face, so it has
  // to come out at essentially 1.0 -- otherwise every 3D token on a board is a
  // different shade of red from every meeple beside it.
  for (const shape of ['token', 'chip', 'disc'] as const) {
    const shade = facetShade(posedUp(shape));
    assert.ok(Math.abs(shade - 1) < 0.05, `${shape}'s cap is drawn at ${shade}`);
  }
});

test('a shape is drawn with more than one brightness', () => {
  // A flat-shaded solid is a polygon. This is the cheapest possible statement
  // that the shading is doing anything at all.
  for (const shape of SHAPES) {
    const fills = tokenSolid(shape, 'red').facets.map((facet) => fillOf(facet.style)[0]);
    assert.ok(new Set(fills).size >= 3, `${shape} draws ${new Set(fills).size} distinct fills`);
  }
});

// ---------------------------------------------------------------------------
// Colour.
// ---------------------------------------------------------------------------

test('an unnamed colour is the red the art is drawn in', () => {
  // What a stack passes for a component whose colour is hidden, and what the
  // flat art shows in the same case: no class matches, so no filter applies.
  assert.deepEqual(tokenBaseColor(''), TOKEN_BASE_RED);
  assert.deepEqual(tokenBaseColor('red'), TOKEN_BASE_RED);
  assert.deepEqual(tokenBaseColor('not-a-colour'), TOKEN_BASE_RED);
});

test('every legal colour is its own colour', () => {
  const seen = new Map<string, string>();
  for (const name of Object.keys(TOKEN_COLOR_FILTERS)) {
    const key = tokenBaseColor(name).join(',');
    assert.ok(!seen.has(key), `${name} and ${seen.get(key)} are the same colour`);
    seen.set(key, name);
  }
  assert.equal(seen.size, Object.keys(TOKEN_COLOR_FILTERS).length);
});

test('the colour name is case-insensitive, as the class is', () => {
  assert.deepEqual(tokenBaseColor('Blue'), tokenBaseColor('blue'));
  assert.deepEqual(
    fillOf(tokenSolid('disc', 'Blue').facets[0].style),
    fillOf(tokenSolid('disc', 'blue').facets[0].style),
  );
});

test('filter functions apply left to right and clamp as they go', () => {
  // Order matters, and WHY it matters is the clamp: brightness and the colour
  // matrices are both linear, so they would commute exactly if nothing ever
  // clipped. Brightening a saturated red past 255 and then desaturating is not
  // the same colour as desaturating first -- 175 against 255 -- and the CSS
  // shorthand is defined left to right.
  const forwards = applyColorFilter('brightness(3) saturate(0)', [200, 20, 20]);
  const backwards = applyColorFilter('saturate(0) brightness(3)', [200, 20, 20]);
  assert.notDeepEqual(forwards.map(Math.round), backwards.map(Math.round));
  // Clamped per function, which is what a filter primitive's result buffer does.
  assert.deepEqual(applyColorFilter('brightness(4)', [200, 200, 200]), [255, 255, 255]);
  assert.deepEqual(applyColorFilter('brightness(0)', [200, 200, 200]), [0, 0, 0]);
  // Identity cases, so a typo in the matrices shows up here rather than on screen.
  assert.deepEqual(applyColorFilter('brightness(1)', [12, 34, 56]), [12, 34, 56]);
  assert.deepEqual(applyColorFilter('hue-rotate(0deg)', [12, 34, 56]).map(Math.round), [12, 34, 56]);
  assert.deepEqual(applyColorFilter('saturate(1)', [12, 34, 56]).map(Math.round), [12, 34, 56]);
  // A grey has no hue to rotate.
  assert.deepEqual(applyColorFilter('hue-rotate(137deg)', [80, 80, 80]).map(Math.round), [80, 80, 80]);
});

test('an unimplemented filter function is refused rather than skipped', () => {
  // Skipping it silently would recolour the flat art and not the solid, which is
  // the exact drift the shared table exists to prevent.
  assert.throws(() => applyColorFilter('sepia(0.5)', TOKEN_BASE_RED), /unsupported function sepia/);
  assert.throws(() => applyColorFilter('nonsense', TOKEN_BASE_RED), /could not parse/);
});

test('a shaded facet keeps its colour\'s hue', () => {
  // Shading multiplies, so the hue survives and only the brightness moves --
  // which is why a 3D blue chip and a flat blue meeple read as the same blue.
  // The trap is the ceiling: orange is (255, 91, 0), so a highlight above 1.0
  // clips the red and lifts only the green, and the lit facet goes yellow.
  for (const name of Object.keys(TOKEN_COLOR_FILTERS).concat(['red'])) {
    const base = tokenBaseColor(name);
    const channels = [0, 1, 2].filter((i) => base[i] >= 16);
    for (const facet of tokenSolid('disc', name).facets) {
      const fill = fillOf(facet.style);
      for (const i of channels) {
        assert.ok(fill[i] <= 255, `${name} channel ${i} clipped at ${fill[i]}`);
        // Every channel scaled by the SAME factor, to within one part in 255.
        const drift = Math.abs(fill[i] / base[i] - fill[channels[0]] / base[channels[0]]);
        assert.ok(drift < 0.02, `${name} channel ${i} drifted by ${drift}`);
      }
    }
  }
});

// ---------------------------------------------------------------------------
// The whole thing.
// ---------------------------------------------------------------------------

test('a solid is a pure function of its type and colour', () => {
  // The property the whole design rests on: nothing here remembers a previous
  // caller, so a recycled host cannot inherit one.
  for (const shape of SHAPES) {
    assert.deepEqual(tokenSolid(shape, 'green'), tokenSolid(shape, 'green'));
    assert.notDeepEqual(
      tokenSolid(shape, 'green').facets[0].style,
      tokenSolid(shape, 'orange').facets[0].style,
    );
  }
  // ...and the geometry does not depend on the colour at all.
  const strip = (style: string) => style.replace(/;background:rgb\([\d,]+\)/, '');
  for (const shape of SHAPES) {
    assert.deepEqual(
      tokenSolid(shape, 'green').facets.map((f) => strip(f.style)),
      tokenSolid(shape, 'black').facets.map((f) => strip(f.style)),
    );
  }
});

test('every facet is a clipped, filled box and carries NO transform', () => {
  for (const shape of SHAPES) {
    const solid = tokenSolid(shape, 'teal');
    let previousKey = -1;
    for (const facet of solid.facets) {
      assert.ok(facet.key > previousKey, `${shape} facet keys must stay in surface order`);
      previousKey = facet.key;
      assert.match(facet.style, /clip-path:polygon\(/);
      assert.match(facet.style, /background:rgb\(/);
      assert.match(facet.style, /width:[\d.]+em;height:[\d.]+em/);
      // THE ONE THAT KEEPS THE FRAME RATE. A transform of any kind here -- and
      // a 3D one in particular -- is what hands every facet its own composited
      // layer the moment a stack's FLIP animates the host: 1,047 layers and
      // 88.6 megapixels for 55 tokens, measured. See src/solid/flat-facets.ts.
      assert.ok(!facet.style.includes('transform'), `${shape} facet ${facet.key}: ${facet.style}`);
      assert.ok(!facet.style.includes('backface'), `${shape} facet ${facet.key}: ${facet.style}`);
    }
  }
});

test('only the four convex types are solids', () => {
  for (const shape of SHAPES) assert.ok(isTokenSolidShape(shape));
  for (const other of ['meeple', 'pawn', '', 'Cube', 'sphere']) {
    assert.ok(!isTokenSolidShape(other), `${other} must not render as a solid`);
  }
});
