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
 * 3. NO LAYER PROMOTION. `will-change: transform` is a dice-specific need — it
 *    exists so a promotion is in place before a tumble starts. A token does not
 *    tumble, and declining promotion is what keeps 55 of them at 60fps. Asserted
 *    as a computed style, because it is the kind of thing that gets copied in
 *    from the die by accident.
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

  test('draws one element per surface polygon, and no more', async ({ page }) => {
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

    // 6 for a cube, 12 walls + 2 caps for a prism. The budget the whole design
    // is built around: 55 tokens at 14 facets is 770 elements, and ~800 is where
    // the measured cliff starts.
    expect(counts.cube).toEqual({ facets: 6, img: 0, solid: true });
    for (const type of ['token', 'chip', 'disc']) {
      expect(counts[type], `${type} is a 12-side prism`).toEqual({ facets: 14, img: 0, solid: true });
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
    expect(result.asSpacer.facets, 'an invisible spacer must not build 14 elements').toBe(0);
    expect(result.asComponent.spacer).toBe(false);
    expect(result.asComponent.facets, 'and the scene appears when it stands for something').toBe(14);
  });

  test('promotes no layers', async ({ page }) => {
    const styles = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.type = 'chip';
      el.item = { ID: 'a' };
      document.body.appendChild(el);
      await el.updateComplete;
      await el.updateComplete;
      const facet = el.renderRoot.querySelector('.facet');
      const inner = el.renderRoot.querySelector('#inner');
      const outer = el.renderRoot.querySelector('#outer');
      const solid = el.renderRoot.querySelector('#solid');
      const result = {
        facetWillChange: getComputedStyle(facet).willChange,
        facetBackface: getComputedStyle(facet).backfaceVisibility,
        innerStyle: getComputedStyle(inner).transformStyle,
        solidStyle: getComputedStyle(solid).transformStyle,
        outerPerspective: getComputedStyle(outer).perspective,
        innerTransform: getComputedStyle(inner).transform,
      };
      el.remove();
      return result;
    });

    // The one that matters for the frame rate.
    expect(styles.facetWillChange, 'a static token must promote nothing').toBe('auto');
    // And the one that makes the picture right: culling IS the hidden-surface
    // removal, and it is sufficient exactly because these shapes are convex.
    expect(styles.facetBackface).toBe('hidden');
    expect(styles.innerStyle, '#inner is the 3D carrier').toBe('preserve-3d');
    expect(styles.solidStyle, 'and the pose carrier must not flatten it').toBe('preserve-3d');
    // 30px * 6 widths. A camera that had quietly gone away would read 'none',
    // and every solid would render orthographically at the wrong size.
    expect(styles.outerPerspective).toBe('180px');
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
        pose: el.renderRoot.querySelector('#solid')?.style.transform ?? null,
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

    expect(result.asCube.facets, 'the first occupant is a cube').toBe(6);
    expect(result.asCube.pose).toMatch(/^matrix3d\(/);

    // The hazard, made visible: a pooled host keeps its last occupant's scene.
    expect(result.whilePooled.orphaned, 'the host must actually leave the stack').toBe(true);
    expect(result.whilePooled.pose, 'and it still wears the cube it was')
      .toBe(result.asCube.pose);

    // THE ASSERTION. Same DOM node, entirely the new component's presentation.
    expect(result.asDisc.recycled, 'the stack must have reused the pooled host').toBe(true);
    expect(result.asDisc.type).toBe('disc');
    expect(result.asDisc.facets, 'a disc is 14 facets, a cube 6').toBe(14);
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
        pose: el.renderRoot.querySelector('#solid').style.transform,
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
