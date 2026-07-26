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
//      triangles (the d7 barrel) -- the same code path for all three;
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
  // keyframes, same hooks, so the trace goldens cannot move -- but a solid has
  // no reel to scroll, and letting that translateY through would slide the die
  // by a multiple of its own size and back on every roll. #inner zeroes the
  // reel step instead, which both the CSS resting rule and the keyframes read.
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
      'translateY(calc(-1 * var(--effective-die-size) * 0))',
      'translateY(calc(-1 * var(--effective-die-size) * 5))',
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
