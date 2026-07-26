import { test, expect } from '@playwright/test';
import { createOfflineGame } from '../helpers';

// Task 8: <boardgame-die> stops being a vertical reel of flat squares and
// becomes an actual solid -- one absolutely positioned element per polygon of
// the surface `die-geometry.ts` generates, placed by a transform derived from
// that polygon's own normal and centroid and cut to its outline with
// clip-path.
//
// These are FIXTURE assertions (no goldens): the die's *animation* is
// unchanged by this task, so nothing here samples motion curves. What it pins
// is the DOM/CSS contract that the later physics tasks build on:
//
//   1. one element per surface polygon, for shapes whose facets are
//      rectangles (d6), triangles (d20) and non-square rectangles + cap
//      triangles (the d7 barrel) -- the same code path for all three, each
//      drawn at the geometry's OWN in-plane extents (test 2: counting
//      clip-path vertices is not enough -- forcing every facet's box square
//      leaves every vertex count untouched while the solid falls apart);
//   1b. the resting pose shows more than one facet, so the die reads as a
//      solid rather than as the flat polygon it replaces;
//   2. `#inner` -- which `motionTrackTarget('visual')` returns and which a
//      later task animates the tumble on -- carries a LIVE preserve-3d
//      context, verified by a consequence (depth sorting) and not merely by
//      a computed-style string that Chromium reports from the specified
//      value even when a grouping property has flattened it;
//   3. every facet is cut to its polygon;
//   4. `--die-size` is a real custom property a caller can set;
//   5. the reel survives as the degenerate fallback for a die with no valid
//      geometry, rather than throwing during a render pass.
//
// Test 5 mounts the die in the REAL app (pig) rather than in an isolated
// fixture, because an isolated fixture cannot see an ancestor that flattens
// the 3D context.

const DIE_SIZE_DEFAULT_PX = 50;

// Mounts one <boardgame-die> in the served app page. Everything else in this
// spec reads through page.evaluate on the same element id.
//
// `SelectedFace` is an INDEX into `Faces`, not a value -- confusing the two is
// the silent bug this component invites, so the fixtures deliberately use face
// VALUES that never coincide with their own indices (10, 20, 30, ...).
async function mountDie(
  page: import('@playwright/test').Page,
  options: { faceCount: number; selectedFace?: number; dieSize?: string },
): Promise<void> {
  await page.goto('/');
  await page.evaluate(async (opts) => {
    document.querySelectorAll('boardgame-die').forEach((el) => el.remove());
    await import('/src/components/boardgame-die.ts');
    const die = document.createElement('boardgame-die') as any;
    die.id = 'fixture-die';
    // z-index because the app shell paints over a bare fixed-position child
    // of <body>, and one test hit-tests the die's own centre.
    die.style.cssText = 'position:fixed;top:120px;left:120px;z-index:9999;';
    if (opts.dieSize) die.style.setProperty('--die-size', opts.dieSize);
    const faces = Array.from({ length: opts.faceCount }, (_, i) => (i + 1) * 10);
    die.item = {
      Values: { Faces: faces },
      DynamicValues: { SelectedFace: opts.selectedFace ?? 0, Value: faces[opts.selectedFace ?? 0] },
    };
    document.body.appendChild(die);
    await die.updateComplete;
    // _itemChanged runs in updated(), which schedules a second render pass.
    await die.updateComplete;
  }, { faceCount: options.faceCount, selectedFace: options.selectedFace, dieSize: options.dieSize } as any);
}

// Reads the facet inventory of the mounted fixture die, plus the surface the
// geometry module says it should have.
async function facetInventory(page: import('@playwright/test').Page, faceCount: number) {
  return await page.evaluate(async (count) => {
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const geometry = geometryModule.dieGeometry(count);
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    return {
      expectedReadable: geometry.faces.length,
      expectedTotal: geometry.faces.length + geometry.capFaces.length,
      expectedPolygonSizes: [...geometry.faces, ...geometry.capFaces]
        .map((f: any) => f.polygon.length)
        .sort((a: number, b: number) => a - b),
      readable: facets.filter((el) => el.dataset.faceIndex !== undefined).length,
      total: facets.length,
      clipPathPolygonSizes: facets
        .map((el) => {
          const clip = getComputedStyle(el).clipPath;
          if (!clip.startsWith('polygon(')) return -1;
          return clip.slice('polygon('.length, -1).split(',').length;
        })
        .sort((a, b) => a - b),
      faceValues: facets
        .filter((el) => el.dataset.faceIndex !== undefined)
        .map((el) => Number(el.dataset.faceValue)),
    };
  }, faceCount);
}

// Reads every facet's RENDERED box and compares it against the geometry's own
// polygons.
//
// The comparison is deliberately not "does the component's own facetStyle say
// what facetStyle says". For each facet it takes:
//
//   - from the DOM: the box's used width/height in px, and the 4x4 matrix the
//     browser resolved for it, whose first three columns are the axes the box
//     is rotated onto (u, v) and its normal (w), and whose last column is
//     where the box centre was moved to;
//   - from `die-geometry.ts`: the facet's polygon, scaled the one way the
//     component documents (`0.5 / circumradius` per em, `1em` = --die-size)
//     and turned into CSS space by the documented (x, -y, -z).
//
// and requires that the polygon, projected onto the rendered u/v axes about
// the rendered box centre, spans that box EXACTLY -- from -width/2 to +width/2
// and -height/2 to +height/2 -- and that every clip-path vertex lands on the
// polygon vertex it is cut from. Reading (u, v) off the render rather than
// recomputing them is what makes this basis-agnostic: the component may orient
// a facet's box however it likes inside the facet's plane, and the assertion
// still holds; what it may NOT do is choose the box's SIZE or CENTRE, which
// belong to the geometry. A square-facet assumption fails the extent check on
// the shorter axis of every non-square facet, and a d7's side facets are
// 2.7:1.
async function facetBoxes(page: import('@playwright/test').Page, faceCount: number) {
  return await page.evaluate(async (count) => {
    const geometryModule: any = await import('/src/motion/die-geometry.ts');
    const geometry = geometryModule.dieGeometry(count);
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
    const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
    // 1em is the die's size: #stage sets font-size, so this is the one number
    // that turns the geometry's own units into the px the browser laid out.
    const emPx = parseFloat(getComputedStyle(root.querySelector('#stage') as HTMLElement).fontSize);
    const unitsToPx = (0.5 / geometry.circumradius) * emPx;
    const surface = [...geometry.faces, ...geometry.capFaces];

    const dot = (a: number[], b: number[]) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
    const rows: any[] = [];
    for (const [index, face] of surface.entries()) {
      const element = facets[index];
      if (!element) { rows.push({ index, missing: true }); continue; }
      const style = getComputedStyle(element);
      const matrix = new DOMMatrix(style.transform);
      const u = [matrix.m11, matrix.m12, matrix.m13];
      const v = [matrix.m21, matrix.m22, matrix.m23];
      const w = [matrix.m31, matrix.m32, matrix.m33];
      const centre = [matrix.m41, matrix.m42, matrix.m43];
      const width = parseFloat(style.width);
      const height = parseFloat(style.height);

      // The geometry's polygon, in the same frame the box centre is given in:
      // px, about the solid's centre, CSS axes (x, -y, -z).
      const projected: number[][] = face.polygon.map((point: number[]) => {
        const screen = [point[0] * unitsToPx, -point[1] * unitsToPx, -point[2] * unitsToPx];
        const offset = [0, 1, 2].map((i) => screen[i] - centre[i]);
        return [dot(offset, u), dot(offset, v)];
      });
      const spans = [0, 1].map((axis) => {
        const values = projected.map((p) => p[axis]);
        return [Math.min(...values), Math.max(...values)];
      });

      // Where the clip-path puts each vertex, in the same box-centred frame.
      const clip = style.clipPath;
      const clipPoints = clip.startsWith('polygon(')
        ? clip.slice('polygon('.length, -1).split(',').map((pair) => {
            const [x, y] = pair.trim().split(/\s+/).map((value) => parseFloat(value) / 100);
            return [x * width - width / 2, y * height - height / 2];
          })
        : null;

      rows.push({
        index,
        faceIndex: element.dataset.faceIndex ?? null,
        width,
        height,
        // The box must be exactly the polygon's extent along the axes it was
        // rotated onto, centred where the polygon's extent is centred.
        boxError: Math.max(
          Math.abs(spans[0][0] + width / 2), Math.abs(spans[0][1] - width / 2),
          Math.abs(spans[1][0] + height / 2), Math.abs(spans[1][1] - height / 2),
        ),
        clipError: clipPoints === null || clipPoints.length !== projected.length
          ? Number.POSITIVE_INFINITY
          : Math.max(...clipPoints.map((p, i) =>
              Math.max(Math.abs(p[0] - projected[i][0]), Math.abs(p[1] - projected[i][1])))),
        // The placement must be a rotation, never a reflection: a mirrored
        // facet draws its glyphs backwards.
        orthonormalError: Math.max(
          Math.abs(dot(u, u) - 1), Math.abs(dot(v, v) - 1), Math.abs(dot(w, w) - 1),
          Math.abs(dot(u, v)), Math.abs(dot(u, w)), Math.abs(dot(v, w)),
          Math.abs(dot([
            u[1] * v[2] - u[2] * v[1],
            u[2] * v[0] - u[0] * v[2],
            u[0] * v[1] - u[1] * v[0],
          ], w) - 1),
        ),
        // Reported so a failure says WHICH way the box is wrong, and so the
        // "these facets are not square in the first place" guard below has
        // something to measure.
        aspect: Math.max(width, height) / Math.min(width, height),
        geometryAspect: Math.max(spans[0][1] - spans[0][0], spans[1][1] - spans[1][0])
          / Math.min(spans[0][1] - spans[0][0], spans[1][1] - spans[1][0]),
      });
    }
    return { rows, emPx, facetCount: facets.length, surfaceCount: surface.length };
  }, faceCount);
}

// Which facets the die actually SHOWS, by hit-testing a grid over its box.
//
// Counted by share of the hit area, not by "was hit at all": a facet a few
// degrees off edge-on is a sliver a pixel or two wide, which is not a facet a
// player can see. Before the resting pose was fixed, a d4's second facet
// covered 0% to 1.7% of the die (it was a back-face, culled outright, or the
// sliver at the silhouette); after, 16% to 33%.
async function visibleFacets(
  page: import('@playwright/test').Page,
  options: { minShare: number },
) {
  return await page.evaluate(async (opts) => {
    const die = document.getElementById('fixture-die') as any;
    const root = die.shadowRoot as ShadowRoot;
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
        const key = facet.dataset.faceIndex ?? `cap${Array.from(root.querySelectorAll('.facet')).indexOf(facet)}`;
        hits.set(key, (hits.get(key) ?? 0) + 1);
      }
    }
    const total = [...hits.values()].reduce((a, b) => a + b, 0);
    const shares = Object.fromEntries(
      [...hits.entries()].map(([key, value]) => [key, Number((value / total).toFixed(3))]),
    );
    return {
      distinctFacetsVisible: [...hits.values()].filter((value) => value / total >= opts.minShare).length,
      presentedShare: total ? (hits.get(String(die.selectedFace)) ?? 0) / total : 0,
      shares,
    };
  }, options);
}

test.describe('boardgame-die solid', () => {
  for (const faceCount of [6, 20, 7]) {
    test(`d${faceCount} renders one element per surface polygon, each cut to its outline`, async ({ page }) => {
      await mountDie(page, { faceCount });
      const inventory = await facetInventory(page, faceCount);

      expect(inventory.readable).toBe(faceCount);
      expect(inventory.readable).toBe(inventory.expectedReadable);
      expect(inventory.total).toBe(inventory.expectedTotal);
      // Every facet is clipped, and clipped to the SAME polygon arity the
      // geometry reports -- a triangle stays a triangle and a pentagon stays
      // a pentagon, which is what "one general routine" has to mean.
      expect(inventory.clipPathPolygonSizes).toEqual(inventory.expectedPolygonSizes);
      // Face VALUES, not indices: a d7 fixture's faces are 10..70.
      expect(inventory.faceValues.slice().sort((a, b) => a - b))
        .toEqual(Array.from({ length: faceCount }, (_, i) => (i + 1) * 10));
    });
  }

  // The assertion the vertex-count check above cannot make. Forcing every
  // facet's box square -- `const sq = Math.max(width, height)` in facetStyle,
  // the exact square-facet assumption this component exists to avoid -- leaves
  // every clip-path vertex COUNT untouched, so the test above stays green
  // while a d7 renders as disconnected slabs and a d6/d12/d20 opens seam gaps.
  // This one measures the box itself against the geometry's own extents.
  for (const faceCount of [6, 20, 12, 10, 7]) {
    test(`d${faceCount} draws each facet at the geometry's own in-plane extents`, async ({ page }) => {
      // 200px so a facet's shorter side is tens of px: the tolerances below
      // are absolute px, and a 50px die would put them near the noise floor.
      await mountDie(page, { faceCount, dieSize: '200px' });
      const { rows, emPx, facetCount, surfaceCount } = await facetBoxes(page, faceCount);

      expect(emPx).toBe(200);
      expect(facetCount).toBe(surfaceCount);
      // DOM order is surface order -- `[...faces, ...capFaces]` -- and the
      // readable prefix of it carries the face indices. Everything below pairs
      // facet i with surface polygon i, so pin that pairing first.
      expect(rows.map((row: any) => row.faceIndex)).toEqual(
        rows.map((_: any, index: number) => (index < faceCount ? String(index) : null)),
      );

      // Sub-pixel on a 200px die. Not exact-equal because the browser reports
      // used px for an em length and the component rounds its own output to
      // five decimals.
      const worst = (key: string) => rows.reduce(
        (best: any, row: any) => (row[key] > best[key] ? row : best), rows[0]);
      const detail = (row: any) =>
        `facet ${row.index}: ${row.width.toFixed(2)}x${row.height.toFixed(2)}px,`
        + ` rendered aspect ${row.aspect.toFixed(3)} vs geometry ${row.geometryAspect.toFixed(3)}`;
      const worstBox = worst('boxError');
      const worstClip = worst('clipError');
      expect(worstBox.boxError, `box vs geometry extents -- ${detail(worstBox)}`)
        .toBeLessThanOrEqual(0.05);
      expect(worstClip.clipError, `clip vertices vs geometry vertices -- ${detail(worstClip)}`)
        .toBeLessThanOrEqual(0.05);
      // Loose because Chromium serializes a computed matrix to about six
      // significant figures (measured worst case 9e-6), and tight enough by a
      // wide margin for what it is looking for: a reflection puts the
      // determinant at -1, i.e. an error of 2.
      expect(Math.max(...rows.map((row: any) => row.orthonormalError)),
        'facet placement is a rotation, not a reflection or a scale')
        .toBeLessThan(1e-3);

      // And the guard that keeps the above from passing vacuously: it only
      // says anything if the geometry's facets are NOT square to begin with.
      // Measured from the geometry, so it moves with the geometry: a d7's
      // side facets are 2.73:1 today, a d10's kites 1.27:1, a d20's triangles
      // 1.07 and 1.16:1, a d12's pentagons 1.05:1 -- and a d6's squares 1:1,
      // which is why the d6 case asserts nothing here.
      if (faceCount !== 6) {
        expect(Math.max(...rows.map((row: any) => row.geometryAspect))).toBeGreaterThan(1.05);
      }
      for (const row of rows) {
        expect(row.aspect, detail(row)).toBeCloseTo(row.geometryAspect, 2);
      }
    });
  }

  // A solid whose faces are far apart in normal angle -- a tetrahedron's are
  // 109.47 degrees apart -- puts every non-presented face behind the camera at
  // the fixed resting tilt, where backface-visibility culls it, so a d4 used to
  // render as ONE flat triangle: the 2D die this component replaces. The pose
  // now tilts as far as it takes for the most face-on of the others to come
  // into view (and not one degree further: shapes with closer normals, which is
  // every other shape, are untouched).
  for (const faceCount of [4, 6, 8, 12, 20]) {
    test(`d${faceCount} rests showing more than one facet`, async ({ page }) => {
      // Every face of the d4 in turn: which of the other three is best placed
      // depends on the presented face's own orientation, and the pose has to
      // work for all of them.
      const selectedFaces = faceCount === 4 ? [0, 1, 2, 3] : [1];
      for (const selectedFace of selectedFaces) {
        await mountDie(page, { faceCount, selectedFace, dieSize: '200px' });
        const visible = await visibleFacets(page, { minShare: 0.05 });
        expect(visible.distinctFacetsVisible,
          `d${faceCount} presenting face ${selectedFace}, hit-area share per facet: ${JSON.stringify(visible.shares)}`)
          .toBeGreaterThanOrEqual(2);
        // The presented face is the value the player reads, so it stays the
        // dominant one. Asserted for the d4 only: it is the shape the tilt
        // touches, and it is the one whose facets are big enough that
        // Chromium's hit testing of a preserve-3d subtree is dependable --
        // on a d12 or d20 it misses whole facets in some orientations.
        if (faceCount === 4) expect(visible.presentedShare).toBeGreaterThan(0.5);
      }
    });
  }

  test('#inner carries a live preserve-3d context that depth-sorts the facets', async ({ page }) => {
    // selectedFace 3 is deliberately neither the first nor the last facet in
    // DOM order: under a FLATTENED context the facets paint in DOM order and
    // the last one wins the centre of the die, so this assertion fails.
    // Mounted large so the perspective probe below has room: the projection
    // shift it measures is a fraction of the die's size.
    await mountDie(page, { faceCount: 6, selectedFace: 3, dieSize: '160px' });
    const result = await page.evaluate(() => {
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const main = root.querySelector('#main') as HTMLElement;
      const box = main.getBoundingClientRect();
      const hit = (root as any).elementFromPoint(
        box.left + box.width / 2,
        box.top + box.height / 2,
      ) as Element | null;
      const facet = hit?.closest?.('.facet') as HTMLElement | null;

      // Second, occlusion-independent probe that the 3D context is LIVE all
      // the way from the perspective wrapper down to the facets: switch the
      // perspective off and the projection must change. It cannot, if
      // anything between #stage and a facet has flattened -- a flattened
      // context ignores the perspective entirely. Concretely, on a cube
      // without perspective the three pairs of opposite faces project to
      // pairwise-IDENTICAL rectangles; with it, the nearer of each pair is
      // measurably bigger.
      const stage = root.querySelector('#stage') as HTMLElement;
      const sizes = () => (Array.from(root.querySelectorAll('.facet')) as HTMLElement[])
        .map((el) => { const r = el.getBoundingClientRect(); return [r.width, r.height]; });
      const projected = sizes();
      stage.style.perspective = 'none';
      const flattened = sizes();
      stage.style.perspective = '';
      const worstShift = Math.max(...projected.flatMap((pair, i) =>
        pair.map((value, j) => Math.abs(value - flattened[i][j]))));

      return {
        transformStyle: getComputedStyle(inner).transformStyle,
        frontFaceIndex: facet?.dataset.faceIndex ?? null,
        frontFaceValue: facet?.dataset.faceValue ?? null,
        hitDescription: hit ? `${hit.tagName.toLowerCase()}.${hit.className}` : 'nothing',
        domOrderLastFaceIndex: (Array.from(root.querySelectorAll('.facet')) as HTMLElement[])
          .filter((el) => el.dataset.faceIndex !== undefined)
          .at(-1)?.dataset.faceIndex ?? null,
        worstShift,
      };
    });

    expect(result.transformStyle).toBe('preserve-3d');
    // Measured 7.0px on a 160px d6; 0 exactly if anything has flattened.
    expect(result.worstShift).toBeGreaterThan(4);
    // The die presents the SELECTED face, and selectedFace is an index: face
    // index 3 carries value 40 in this fixture.
    expect(`${result.frontFaceIndex} (hit ${result.hitDescription})`)
      .toBe(`3 (hit ${result.hitDescription})`);
    expect(result.frontFaceValue).toBe('40');
    expect(result.domOrderLastFaceIndex).not.toBe('3');
  });

  test('--die-size is a real custom property a caller can set', async ({ page }) => {
    await mountDie(page, { faceCount: 6 });
    const defaults = await page.evaluate(() => {
      const root = (document.getElementById('fixture-die') as any).shadowRoot as ShadowRoot;
      const main = root.querySelector('#main') as HTMLElement;
      const facet = root.querySelector('.facet') as HTMLElement;
      return { box: main.offsetWidth, facet: facet.offsetWidth };
    });
    expect(defaults.box).toBe(DIE_SIZE_DEFAULT_PX);

    await mountDie(page, { faceCount: 6, dieSize: '160px' });
    const scaled = await page.evaluate(() => {
      const root = (document.getElementById('fixture-die') as any).shadowRoot as ShadowRoot;
      const main = root.querySelector('#main') as HTMLElement;
      const facet = root.querySelector('.facet') as HTMLElement;
      return { box: main.offsetWidth, facet: facet.offsetWidth };
    });
    expect(scaled.box).toBe(160);
    // The SOLID scales with the property, not just the box around it.
    expect(scaled.facet / defaults.facet).toBeCloseTo(160 / DIE_SIZE_DEFAULT_PX, 1);
  });

  // The reel's face-change spin is UNCHANGED by this task -- same track, same
  // hooks, so the trace goldens cannot move -- but a solid has no reel to
  // scroll, and letting that translateY through would slide the die by a
  // multiple of its own size and back on every roll. #inner.solid zeroes
  // --reel-step instead, which both the CSS resting rule and the keyframes
  // read, and which is a variable of its own precisely so that zeroing it
  // cannot zero anything else the solid measures against the die's size.
  test('the face-change spin still plays, and does not displace the solid', async ({ page }) => {
    await mountDie(page, { faceCount: 6, selectedFace: 0 });
    const result = await page.evaluate(async () => {
      const die = document.getElementById('fixture-die') as any;
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const at = () => getComputedStyle(inner).transform;
      const before = at();
      die.item = {
        Values: { Faces: die.faces },
        DynamicValues: { SelectedFace: 5, Value: die.faces[5] },
      };
      await die.updateComplete;
      await die.updateComplete;
      const animations = inner.getAnimations();
      const keyframes = animations[0]?.effect instanceof KeyframeEffect
        ? animations[0].effect.getKeyframes().map((frame: any) => frame.transform)
        : [];
      await new Promise((resolve) => setTimeout(resolve, 80));
      const during = at();
      await Promise.all(animations.map((a) => a.finished.catch(() => undefined)));
      return { before, during, after: at(), count: animations.length, keyframes };
    });

    expect(result.count).toBe(1);
    expect(result.keyframes).toEqual([
      'translateY(calc(-1 * var(--reel-step) * 0))',
      'translateY(calc(-1 * var(--reel-step) * 5))',
    ]);
    expect([result.before, result.during, result.after])
      .toEqual(['matrix(1, 0, 0, 1, 0, 0)', 'matrix(1, 0, 0, 1, 0, 0)', 'matrix(1, 0, 0, 1, 0, 0)']);
  });

  test('falls back to the reel for a die with fewer than three faces', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-die.ts');
      const errors: string[] = [];
      const onError = (e: ErrorEvent) => errors.push(String(e.message));
      window.addEventListener('error', onError);
      const mount = async (faces: number[] | null) => {
        const die = document.createElement('boardgame-die') as any;
        die.item = faces === null ? null : {
          Values: { Faces: faces },
          DynamicValues: { SelectedFace: 0, Value: faces[0] },
        };
        document.body.appendChild(die);
        await die.updateComplete;
        await die.updateComplete;
        const root = die.shadowRoot as ShadowRoot;
        const out = {
          facets: root.querySelectorAll('.facet').length,
          reelFaces: root.querySelectorAll('#inner.reel .face').length,
          hasInner: !!root.querySelector('#inner'),
        };
        die.remove();
        return out;
      };
      const twoFaced = await mount([1, 2]);
      const empty = await mount(null);
      window.removeEventListener('error', onError);
      return { twoFaced, empty, errors };
    });

    expect(result.errors).toEqual([]);
    expect(result.twoFaced).toEqual({ facets: 0, reelFaces: 2, hasInner: true });
    expect(result.empty).toEqual({ facets: 0, reelFaces: 0, hasInner: true });
  });

  test('the solid survives mounted in the real app (pig)', async ({ page }) => {
    test.setTimeout(120000);
    await createOfflineGame(page, 'pig');
    const result = await page.evaluate(() => {
      const deepQueryFirst = (root: Document | ShadowRoot | Element, selector: string): Element | null => {
        const direct = root.querySelector(selector);
        if (direct) return direct;
        for (const el of Array.from(root.querySelectorAll('*'))) {
          if ((el as any).shadowRoot) {
            const found = deepQueryFirst((el as any).shadowRoot, selector);
            if (found) return found;
          }
        }
        return null;
      };
      const die = deepQueryFirst(document, 'boardgame-die') as HTMLElement | null;
      if (!die) return { found: false } as any;
      // The game view mounts scrolled to the admin panel, so the board can be
      // outside the viewport, where elementFromPoint answers null.
      die.scrollIntoView({ block: 'center' });
      const root = die.shadowRoot as ShadowRoot;
      const inner = root.querySelector('#inner') as HTMLElement;
      const facets = Array.from(root.querySelectorAll('.facet')) as HTMLElement[];
      const main = root.querySelector('#main') as HTMLElement;
      const box = main.getBoundingClientRect();
      const hit = (root as any).elementFromPoint(
        box.left + box.width / 2,
        box.top + box.height / 2,
      ) as Element | null;
      const facet = hit?.closest?.('.facet') as HTMLElement | null;
      const selectedFace = (die as any).selectedFace;
      // Ancestors that flatten a 3D context do so by carrying a grouping
      // property; report the chain so a failure names the culprit.
      const flatteners: string[] = [];
      let node: Element | null = inner;
      while (node) {
        const style = getComputedStyle(node);
        const reasons: string[] = [];
        if (style.overflow !== 'visible') reasons.push(`overflow:${style.overflow}`);
        if (style.filter !== 'none') reasons.push(`filter:${style.filter}`);
        if (Number(style.opacity) < 1) reasons.push(`opacity:${style.opacity}`);
        if (style.clipPath !== 'none' && node !== inner) reasons.push(`clip-path:${style.clipPath}`);
        if (reasons.length) flatteners.push(`${node.tagName.toLowerCase()}#${node.id || ''}: ${reasons.join(',')}`);
        node = node.parentElement
          ?? ((node.getRootNode() as ShadowRoot).host as Element | null)
          ?? null;
      }
      return {
        found: true,
        transformStyle: getComputedStyle(inner).transformStyle,
        facetCount: facets.length,
        clipped: facets.every((el) => getComputedStyle(el).clipPath.startsWith('polygon(')),
        frontFaceIndex: facet?.dataset.faceIndex ?? null,
        hitDescription: hit ? `${hit.tagName.toLowerCase()}.${hit.className}` : 'nothing',
        selectedFace: String(selectedFace),
        innerFlatteners: flatteners,
      };
    });

    expect(result.found).toBe(true);
    expect(result.transformStyle).toBe('preserve-3d');
    // pig's die is a d6: six readable faces, no cap triangles.
    expect(result.facetCount).toBe(6);
    expect(result.clipped).toBe(true);
    // Depth sorting works in situ: the presented face is the one at the
    // centre of the die, not whichever facet paints last.
    expect(`${result.frontFaceIndex} (hit ${result.hitDescription}; flatteners above #inner: ${JSON.stringify(result.innerFlatteners)})`)
      .toBe(`${result.selectedFace} (hit ${result.hitDescription}; flatteners above #inner: ${JSON.stringify(result.innerFlatteners)})`);
  });
});
