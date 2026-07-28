import { test, expect, type Page } from '@playwright/test';

/**
 * A 3D TOKEN, AS A BROWSER ACTUALLY DRAWS IT.
 *
 * `token-solid.test.ts` proves the arithmetic — the pose is a rotation, the fit
 * is a fixed point, the colours are the same nine filters the flat art uses.
 * None of that says a browser draws anything. This spec is the other half: the
 * real component, in a real page, under the real stack.
 *
 * Four things are asserted here that cannot be asserted anywhere else.
 *
 * 1. THE POSE SURVIVES A RECYCLED HOST. This is the risk the design calls the
 *    biggest one. A stack pools component hosts across membership changes and
 *    NOTHING re-derives a pose on reuse, so a node carries whatever the previous
 *    occupant's last write left on it. It is the failure `boardgame-die.ts`'s
 *    `_clearRoll` exists to prevent — a die that dropped a roll without clearing
 *    the transform sat 60 to 106px outside its own slot, permanently — and here
 *    it would be worse, because a token's motion carriers are `noAnimate` and
 *    can never self-correct by playing anything. The test drives a real
 *    `boardgame-component-stack` through a shrink and a regrow, checks by NODE
 *    IDENTITY that the same element came back out of the pool, and then requires
 *    it to show the new component's shape, pose and colour. A test that only
 *    mounted a fresh token would pass while this was broken.
 *
 * 2. THE SILHOUETTE FILLS THE BOX. The whole sizing contract, measured off the
 *    facets the browser actually laid out rather than off the arithmetic that
 *    produced them. It fails if the 3D context collapses, if the camera is
 *    missing, if the fit is computed the die's way (which draws every token
 *    ~40% small), or if a facet is misplaced.
 *
 * 3. NO LAYER PROMOTION, WHILE AN ANCESTOR TRANSFORM IS ANIMATING. Not at rest
 *    — at rest the broken version passed. Chromium promotes every element in a
 *    live `preserve-3d` context the instant something above it animates, which
 *    is what a stack's FLIP does on every move: 55 tokens went from 57
 *    composited layers to 1,047 and from 59.6 to 30fps. Measured through the
 *    browser's own LayerTree, because a frame-rate assertion is a coin flip and
 *    the layer count is the thing that actually changed.
 *
 * 4. THE COLOURS AGREE WITH THE BROWSER'S OWN FILTERS. The arithmetic in
 *    `token-solid.ts` reimplements `hue-rotate`, `saturate` and `brightness`.
 *    Only Chromium can say whether it got them right, and it says so exactly:
 *    zero distance on all nine colours.
 */

/** Walk light and shadow DOM alike; every component here lives in a shadow root. */
async function deepQueryInstalled(page: Page): Promise<void> {
  await page.evaluate(() => {
    (window as any).__deep = (selector: string): Element[] => {
      const found: Element[] = [];
      const walk = (node: ParentNode) => {
        for (const element of node.querySelectorAll(selector)) found.push(element);
        for (const element of node.querySelectorAll('*')) {
          if (element.shadowRoot) walk(element.shadowRoot);
        }
      };
      walk(document);
      return found;
    };
  });
}

test.describe('a token that is a solid', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await deepQueryInstalled(page);
  });

  test('draws one element per VISIBLE surface polygon, and no more', async ({ page }) => {
    const counts = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const out: Record<string, { facets: number; img: number; solid: boolean }> = {};
      for (const type of ['cube', 'token', 'chip', 'disc', 'meeple', 'pawn']) {
        const el = document.createElement('boardgame-token') as any;
        el.type = type;
        el.item = { ID: type };
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        out[type] = {
          facets: el.renderRoot.querySelectorAll('.facet').length,
          img: el.renderRoot.querySelectorAll('img').length,
          solid: el.renderRoot.querySelector('#outer').classList.contains('solid'),
        };
        el.remove();
      }
      return out;
    });

    // A cube is 6 polygons and a prism 12 walls + 2 caps, but a solid only
    // BUILDS the ones the camera can see: `backface-visibility: hidden` is gone
    // and token-solid.ts culls instead. 3 of a cube's 6, and one cap plus five
    // walls of a prism's 14. 55 tokens is 330 elements, not 770.
    expect(counts.cube).toEqual({ facets: 3, img: 0, solid: true });
    for (const type of ['token', 'chip', 'disc']) {
      expect(counts[type], `${type} is a 12-side prism`).toEqual({ facets: 6, img: 0, solid: true });
    }
    // The non-convex pair keep their authored art, and build no scene: culling
    // is provably wrong for them (see the design's measured tilt table).
    for (const type of ['meeple', 'pawn']) {
      expect(counts[type], `${type} keeps its art`).toEqual({ facets: 0, img: 1, solid: false });
    }
  });

  test('builds no scene at all for a spacer', async ({ page }) => {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.type = 'disc';
      // No item: a spacer, which holds a slot open and is visibility: hidden.
      document.body.appendChild(el);
      await el.updateComplete;
      await el.updateComplete;
      const asSpacer = {
        spacer: el.spacer,
        facets: el.renderRoot.querySelectorAll('.facet').length,
        visibility: getComputedStyle(el.renderRoot.querySelector('#outer')).visibility,
      };
      // ...and the moment it stands for something, it is a solid.
      el.item = { ID: 'real' };
      await el.updateComplete;
      await el.updateComplete;
      const asComponent = {
        spacer: el.spacer,
        facets: el.renderRoot.querySelectorAll('.facet').length,
        visibility: getComputedStyle(el.renderRoot.querySelector('#outer')).visibility,
      };
      el.remove();
      return { asSpacer, asComponent };
    });

    expect(result.asSpacer.spacer, 'a token with no item is a spacer').toBe(true);
    expect(result.asSpacer.visibility).toBe('hidden');
    expect(result.asSpacer.facets, 'an invisible spacer must not build a scene').toBe(0);
    expect(result.asComponent.spacer).toBe(false);
    expect(result.asComponent.facets, 'and the scene appears when it stands for something').toBe(6);
  });

  /**
   * THE FRAME-RATE TEST, AND THE REASON THIS FILE EXISTS AT ALL.
   *
   * A token used to draw itself in a live `preserve-3d` scene: a `perspective`
   * on #outer, `transform-style: preserve-3d` on #inner and #solid, and a
   * `matrix3d` on every facet. Nothing about that asked to be composited, and at
   * rest nothing was. But CHROMIUM PROMOTES EVERY ELEMENT INSIDE A LIVE 3D
   * SORTING CONTEXT TO ITS OWN COMPOSITED LAYER THE MOMENT AN ANCESTOR
   * TRANSFORM ANIMATES -- and a stack's FLIP animates the component host on
   * every single move. Measured in `pass`, 55 tokens at 14 facets: 57
   * composited layers at rest, 1,047 during a move, and 88.6 megapixels of
   * layer area, because a clip-path'ed facet seen through a perspective gets
   * conservative layer bounds two thousand pixels across. 30fps against the
   * flat art's 59.6, linear in facet count, and unmoved by promoting the
   * container -- because the promotion was never the container's.
   *
   * A token's pose is a constant, so the fix was to stop asking a browser to
   * project it: token-solid.ts does the perspective divide once and emits flat,
   * already-projected, untransformed polygons. `pass` went to 60.5fps during a
   * move, at parity with the flat SVG art.
   *
   * THIS IS A STRUCTURAL TEST, NOT A FRAME-RATE ONE, deliberately. A frames-
   * per-second assertion on a shared machine is a coin flip; the layer count is
   * the thing that actually changed and it is exact -- 55 tokens promoted 57
   * layers before this test could fail and 1,047 after. It drives the layers
   * through the browser's own LayerTree via CDP rather than inferring them, and
   * it animates a real ancestor transform, because AT REST the broken version
   * passes every assertion here.
   */
  test('promotes no layers, even while an ancestor transform animates', async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== 'chromium', 'CDP LayerTree is Chromium-only');

    const TOKENS = 55;
    await page.evaluate(async (count) => {
      await import('/src/components/boardgame-token.ts');
      document.body.style.cssText = 'margin:0;background:#eee;height:100vh;overflow:hidden';
      document.body.innerHTML = '';
      const tokens: any[] = [];
      for (let i = 0; i < count; i++) {
        const el = document.createElement('boardgame-token') as any;
        el.type = 'chip';
        el.color = ['red', 'blue', 'green', 'yellow'][i % 4];
        el.item = { ID: `t${i}` };
        el.style.cssText = `position:absolute;left:${(i % 10) * 95 + 10}px;`
          + `top:${Math.floor(i / 10) * 95 + 10}px;--component-width:60px`;
        document.body.appendChild(el);
        tokens.push(el);
      }
      await Promise.all(tokens.map((t) => t.updateComplete));
      await Promise.all(tokens.map((t) => t.updateComplete));
      (window as any).__tokens = tokens;
      // Exactly what a stack's FLIP does to a host, and the only thing that
      // provokes the promotion: an ANIMATING ANCESTOR TRANSFORM.
      (window as any).__drive = () => tokens.map((t) => t.animate(
        [{ transform: 'translate(0px,0px)' }, { transform: 'translate(8px,5px)' }],
        { duration: 900, iterations: Infinity, direction: 'alternate', composite: 'add' },
      ));
    }, TOKENS);

    const facets = await page.evaluate(() => (window as any).__tokens
      .reduce((n: number, t: any) => n + t.renderRoot.querySelectorAll('.facet').length, 0));
    expect(facets, 'the scene under test must actually be 55 solids').toBe(TOKENS * 6);

    // The measured half: the browser's own layer tree, while the hosts animate.
    const client = await page.context().newCDPSession(page);
    const snapshots: any[][] = [];
    client.on('LayerTree.layerTreeDidChange', (event: any) => {
      if (event.layers) snapshots.push(event.layers);
    });
    await client.send('LayerTree.enable');
    await page.waitForTimeout(500);
    await page.evaluate(() => (window as any).__drive());
    await page.waitForTimeout(1200);
    await client.send('LayerTree.disable');
    await client.detach();

    const layers = snapshots[snapshots.length - 1] ?? [];
    const painted = layers.filter((layer) => layer.drawsContent);
    const megapixels = painted.reduce((sum, l) => sum + l.width * l.height, 0) / 1e6;

    // The animating hosts themselves are promoted, and should be: that is one
    // layer per token plus the document's own, which is what the flat SVG art
    // took too (57 painted layers, 1.6 megapixels, measured). A facet that took
    // a layer of its own would put this in four figures.
    expect(layers.length, `the layer tree did not arrive (${snapshots.length} snapshots)`)
      .toBeGreaterThan(0);
    expect(painted.length,
      `55 animating tokens must promote about one layer each, not one per facet`)
      .toBeLessThan(TOKENS * 2);
    // And the layers that do exist must be token-sized. The old version's
    // facets were 2062x2062 each; this catches a promotion that somehow kept
    // the count down but not the area.
    expect(megapixels, 'the promoted layer area must stay in the low megapixels')
      .toBeLessThan(10);

    // The declarative half: nothing a token renders may name a 3D context or
    // ask to be promoted. Each of these on its own was measured to leave ~1,000
    // layers standing, so all of them have to hold.
    const styles = await page.evaluate(() => {
      const el = (window as any).__tokens[0];
      const facet = el.renderRoot.querySelector('.facet');
      return {
        facetWillChange: getComputedStyle(facet).willChange,
        facetTransform: getComputedStyle(facet).transform,
        facetBackface: getComputedStyle(facet).backfaceVisibility,
        innerStyle: getComputedStyle(el.renderRoot.querySelector('#inner')).transformStyle,
        solidStyle: getComputedStyle(el.renderRoot.querySelector('#solid')).transformStyle,
        solidTransform: getComputedStyle(el.renderRoot.querySelector('#solid')).transform,
        outerPerspective: getComputedStyle(el.renderRoot.querySelector('#outer')).perspective,
        innerTransform: getComputedStyle(el.renderRoot.querySelector('#inner')).transform,
      };
    });
    expect(styles.facetWillChange, 'a token must promote nothing').toBe('auto');
    expect(styles.facetTransform, 'a facet is already projected; it takes no transform')
      .toBe('none');
    expect(styles.solidTransform, 'the pose is in the polygons, not on #solid').toBe('none');
    expect(styles.outerPerspective, 'the camera is applied in JavaScript, not by CSS')
      .toBe('none');
    expect(styles.innerStyle, '#inner must not open a 3D sorting context').toBe('flat');
    expect(styles.solidStyle, 'nor may #solid').toBe('flat');
    expect(styles.facetBackface, 'culling happens in token-solid.ts now').toBe('visible');
    // #inner's transform belongs to the animation kernel. Nothing here may put
    // anything on it -- see beforeOrphaned.
    expect(styles.innerTransform).toBe('none');

  });

  /**
   * Measured off PIXELS, and it has to be.
   *
   * A facet's `getBoundingClientRect()` is the rect of its unclipped element
   * BOX, and `clip-path` is what cuts that box down to the polygon: a prism's
   * cap box is the whole cross-section square, so the union of the rects
   * overstates a leaned token's silhouette by 16%. The only honest measurement
   * of a drawn silhouette is what got drawn.
   *
   * The threshold ignores the elevation drop-shadow, which is a filter on
   * #outer, is not part of the solid, and darkens white by at most about 20%;
   * every token colour is more than 120 away from white.
   */
  test('draws a silhouette that fills the token\'s own box', async ({ page }) => {
    const BOX = 200;
    for (const type of ['cube', 'token', 'chip', 'disc']) {
      const box = await page.evaluate(async ({ type: shape, size }) => {
        await import('/src/components/boardgame-token.ts');
        document.body.innerHTML = '';
        document.body.style.cssText = 'margin:0;background:#fff';
        const el = document.createElement('boardgame-token') as any;
        el.type = shape;
        el.item = { ID: shape };
        el.style.cssText = 'position:fixed;left:0;top:0;display:block';
        el.style.setProperty('--component-width', `${size}px`);
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
        const rect = el.renderRoot.querySelector('#inner').getBoundingClientRect();
        return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
      }, { type, size: BOX });
      expect(box, `${type} keeps the box the flat art had`)
        .toEqual({ x: 0, y: 0, width: BOX, height: BOX });

      // A margin around the box, so an overflowing solid is measured rather
      // than clipped by the screenshot.
      const margin = 40;
      const shot = (await page.screenshot({
        clip: { x: 0, y: 0, width: BOX + margin, height: BOX + margin },
      })).toString('base64');
      const drawn = await page.evaluate(async ({ base64 }) => {
        const image = new Image();
        image.src = 'data:image/png;base64,' + base64;
        await image.decode();
        const canvas = document.createElement('canvas');
        canvas.width = image.width;
        canvas.height = image.height;
        const context = canvas.getContext('2d', { willReadFrequently: true })!;
        context.drawImage(image, 0, 0);
        const data = context.getImageData(0, 0, canvas.width, canvas.height).data;
        let left = Infinity;
        let right = -Infinity;
        let top = Infinity;
        let bottom = -Infinity;
        let painted = 0;
        for (let y = 0; y < canvas.height; y++) {
          for (let x = 0; x < canvas.width; x++) {
            const p = (y * canvas.width + x) * 4;
            const distance = Math.hypot(255 - data[p], 255 - data[p + 1], 255 - data[p + 2]);
            if (distance < 120) continue;
            painted++;
            left = Math.min(left, x);
            right = Math.max(right, x);
            top = Math.min(top, y);
            bottom = Math.max(bottom, y);
          }
        }
        return { painted, width: right - left + 1, height: bottom - top + 1, left, top };
      }, { base64: shot });

      expect(drawn.painted, `${type} drew nothing`).toBeGreaterThan(1000);
      // Filling means the LARGER extent is the box: a leaned prism is taller
      // than it is wide, and it may not overflow the slot the stack gave it
      // (stack margins, the board's aspect-ratio clamp and the FLIP scale ratio
      // all key off that box).
      const widest = Math.max(drawn.width, drawn.height);
      expect(widest, `${type} drew ${drawn.width}x${drawn.height} in a ${BOX}px box`)
        .toBeGreaterThan(BOX * 0.97);
      expect(widest, `${type} overflowed its own box`).toBeLessThan(BOX * 1.03);
      // ...and it is not a degenerate sliver in the other direction either.
      expect(Math.min(drawn.width, drawn.height)).toBeGreaterThan(BOX * 0.55);
      // The cube is the shape that says the sizing MODEL is the right one:
      // sized the die's way (nominal sphere to box) it would draw at
      // 1/sqrt(3) = 58% of this and the assertion above would fail by 40%.
    }
  });

  test('is coloured by the same filters the flat art is', async ({ page }) => {
    const comparison = await page.evaluate(async () => {
      const solid = await import('/src/components/token-solid.ts');
      const base = solid.TOKEN_BASE_RED;
      const names = Object.keys(solid.TOKEN_COLOR_FILTERS);
      // One swatch per colour, filled with the base red and filtered by the very
      // string the token's `#outer.<color> img` rule uses.
      const row = document.createElement('div');
      row.style.cssText = 'position:fixed;left:0;top:0;display:flex;z-index:2147483647';
      for (const name of names) {
        const swatch = document.createElement('div');
        swatch.style.cssText = `width:20px;height:20px;background:rgb(${base.join(',')});`
          + `filter:${solid.TOKEN_COLOR_FILTERS[name]}`;
        row.appendChild(swatch);
      }
      document.body.appendChild(row);
      await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
      return {
        names,
        computed: names.map((name) => solid.tokenBaseColor(name).map(Math.round)),
        width: names.length * 20,
      };
    });

    const shot = (await page.screenshot({
      clip: { x: 0, y: 0, width: comparison.width, height: 20 },
    })).toString('base64');
    const painted = await page.evaluate(async ({ base64, count }) => {
      const image = new Image();
      image.src = 'data:image/png;base64,' + base64;
      await image.decode();
      const canvas = document.createElement('canvas');
      canvas.width = image.width;
      canvas.height = image.height;
      const context = canvas.getContext('2d', { willReadFrequently: true })!;
      context.drawImage(image, 0, 0);
      const out: number[][] = [];
      for (let i = 0; i < count; i++) {
        const data = context.getImageData(i * 20 + 10, 10, 1, 1).data;
        out.push([data[0], data[1], data[2]]);
      }
      document.querySelectorAll('div').forEach((element) => element.remove());
      return out;
    }, { base64: shot, count: comparison.names.length });

    expect(comparison.names.length, 'nine recoloured tokens plus the unfiltered red')
      .toBe(9);
    comparison.names.forEach((name, index) => {
      // Exact. Measured at distance 0 on every colour, which is the point: the
      // arithmetic is not an approximation of the filter, it IS the filter.
      expect(painted[index], `${name}: the arithmetic must match the browser's own filter`)
        .toEqual(comparison.computed[index]);
    });
    // ...and the whole thing would be vacuous if every colour were the same.
    expect(new Set(painted.map((rgb) => rgb.join(','))).size).toBe(9);
  });

  /**
   * THE ONE THAT MATTERS. See the file docs.
   *
   * A real `boardgame-component-stack`, driven the way a game drives it, through
   * the exact membership change that recycles a host: two components, then one
   * (the second host is orphaned into `_componentPool`), then two again with a
   * DIFFERENT shape and colour in the second slot (`newComponent()` pops the
   * pool). The middle snapshot is taken while the host is orphaned, and it is
   * expected to STILL be wearing the old cube: that is the whole hazard, and
   * seeing it there is what makes the last snapshot mean something.
   */
  test('shows the new component\'s pose on a recycled host', async ({ page }) => {
    const result = await page.evaluate(async () => {
      const view = await import('/src/components/component-view.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const stack = document.createElement('boardgame-component-stack') as any;
      document.body.appendChild(stack);
      stack.componentView = view.tokenView({
        properties: ({ kind, component }: any) => kind === 'visible'
          ? { type: component.Values.Type, color: component.Values.Color }
          : { type: 'token', color: 'red' },
      });

      const component = (id: string, Type: string, Color: string) => ({
        ID: id, Index: 0, Deck: 'd', GameName: 'g',
        Values: { Type, Color }, DynamicValues: {},
      });
      const setMembership = async (components: unknown[]) => {
        stack.stack = {
          Deck: 'd',
          Size: components.length,
          Components: components,
          IDs: components.map((c: any) => c?.ID ?? ''),
        };
        await stack.updateComplete;
        await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
        for (const child of [...stack.children]) await (child as any).updateComplete;
        for (const child of [...stack.children]) await (child as any).updateComplete;
      };
      const snapshot = (el: any) => ({
        type: el.type,
        color: el.color,
        facets: el.renderRoot.querySelectorAll('.facet').length,
        // The pose is no longer a transform anywhere: it is baked into every
        // facet's clip-path, so the facet STYLES are what carries it and what a
        // stale host would be wearing. See src/solid/flat-facets.ts.
        pose: [...el.renderRoot.querySelectorAll('.facet')]
          .map((facet: any) => facet.style.clipPath).join('|') || null,
        fontSize: el.renderRoot.querySelector('#solid')?.style.fontSize ?? null,
        fill: el.renderRoot.querySelector('.facet')?.style.background ?? null,
      });

      await setMembership([component('a', 'cube', 'red'), component('b', 'cube', 'red')]);
      const host = [...stack.querySelectorAll('boardgame-token')][1] as any;
      const asCube = snapshot(host);

      // Shrink: this host is removed and pushed into the pool.
      await setMembership([component('a', 'cube', 'red')]);
      const whilePooled = { orphaned: host.parentElement === null, ...snapshot(host) };

      // Regrow with a different component. `newComponent()` pops the pool.
      await setMembership([component('a', 'cube', 'red'), component('c', 'disc', 'blue')]);
      const hosts = [...stack.querySelectorAll('boardgame-token')];
      const asDisc = { recycled: hosts[1] === host, ...snapshot(host) };

      stack.remove();
      return { asCube, whilePooled, asDisc };
    });

    expect(result.asCube.facets, 'the first occupant is a cube').toBe(3);
    expect(result.asCube.pose).toMatch(/^polygon\(/);

    // The hazard, made visible: a pooled host keeps its last occupant's scene.
    expect(result.whilePooled.orphaned, 'the host must actually leave the stack').toBe(true);
    expect(result.whilePooled.pose, 'and it still wears the cube it was')
      .toBe(result.asCube.pose);

    // THE ASSERTION. Same DOM node, entirely the new component's presentation.
    expect(result.asDisc.recycled, 'the stack must have reused the pooled host').toBe(true);
    expect(result.asDisc.type).toBe('disc');
    expect(result.asDisc.facets, 'a visible disc is 6 facets, a visible cube 3').toBe(6);
    expect(result.asDisc.pose, 'the pose must be the disc\'s, not the cube\'s')
      .not.toBe(result.asCube.pose);
    expect(result.asDisc.fontSize, 'and so must the size')
      .not.toBe(result.asCube.fontSize);
    expect(result.asDisc.fill, 'and the colour').not.toBe(result.asCube.fill);
    // Not merely "different": exactly what a freshly mounted blue disc shows.
    const fresh = await page.evaluate(async () => {
      // No import: `component-view.ts` already pulled the element in, and
      // importing it again by a different specifier registers it twice.
      const el = document.createElement('boardgame-token') as any;
      el.type = 'disc';
      el.color = 'blue';
      el.item = { ID: 'fresh' };
      document.body.appendChild(el);
      await el.updateComplete;
      await el.updateComplete;
      const out = {
        pose: [...el.renderRoot.querySelectorAll('.facet')]
          .map((facet: any) => facet.style.clipPath).join('|'),
        fontSize: el.renderRoot.querySelector('#solid').style.fontSize,
        fill: el.renderRoot.querySelector('.facet').style.background,
      };
      el.remove();
      return out;
    });
    expect({
      pose: result.asDisc.pose,
      fontSize: result.asDisc.fontSize,
      fill: result.asDisc.fill,
    }, 'a recycled host must be indistinguishable from a fresh one').toEqual(fresh);
  });

  /**
   * The other half of "explicitly cleared on orphan or recycle".
   *
   * `#inner` is what `motionTrackTarget('visual')` returns, so the animation
   * kernel writes a track's resting value there — and a token compiles no visual
   * tracks today, which is exactly the situation in which a stale write is
   * invisible until it is permanent. Measured precedent: a stale write to a
   * card's #inner transform was invisible for a whole flight and became
   * permanent the instant the animation ended, leaving the card stuck at 45
   * degrees. This writes one by hand and requires the orphan hook to clear it.
   */
  test('clears the visual carrier before the host is orphaned', async ({ page }) => {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.type = 'disc';
      el.item = { ID: 'a' };
      document.body.appendChild(el);
      await el.updateComplete;
      await el.updateComplete;
      const inner = el.renderRoot.querySelector('#inner');
      // What a visual motion track's resting value would look like.
      inner.style.transform = 'rotate(45deg) scale(2)';
      const before = getComputedStyle(inner).transform;
      el.beforeOrphaned();
      const after = getComputedStyle(inner).transform;
      el.remove();
      return { before, after };
    });

    expect(result.before, 'the probe must actually have written something')
      .not.toBe('none');
    expect(result.after, 'an orphaned host must not carry a pose into its next life')
      .toBe('none');
  });
});
