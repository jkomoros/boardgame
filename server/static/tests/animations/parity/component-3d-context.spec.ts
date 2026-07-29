import { test, expect } from '@playwright/test';

// Every BoardgameComponent must be able to host a 3D scene on #inner.
//
// #inner is what motionTrackTarget('visual') returns, so it is the element a
// component-owned solid mounts on and the element `transform-style:
// preserve-3d` has to go on -- exactly the role #inner plays in
// boardgame-die.ts's own solid ("Nothing from #stage down may carry a grouping
// property ... each of those forces transform-style back to flat and collapses
// the solid into a pile of overlapping outlines").
//
// A `filter` is such a grouping property. The alt-shadow elevation
// (`#outer.alt-shadow #inner`) and boardgame-token's highlight throb both used
// to put one on #inner, which flattened any 3D context rooted there for every
// token in the app. Both now live on #outer.
//
// The probe below is geometric, not declarative: it measures whether a
// translateZ'd child is actually projected, so it fails if the browser
// flattens the context for ANY reason, not merely if a particular CSS rule
// moved. It depends on nothing outside #inner -- #inner supplies its own
// perspective -- so the only variable is what filter #inner carries.
//
//   perspective 200px, child at z = 100px  ->  scale 200 / (200 - 100) = 2
//   flattened                              ->  translateZ discarded, scale 1
//
// A 10px child therefore measures 20px when the 3D context is honored and
// 10px when it is not.
test.describe('component-owned 3D scenes', () => {
  test('a preserve-3d context on #inner survives a token\'s elevation and throb', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));
      await frame();

      const inner = el.renderRoot.querySelector('#inner') as HTMLElement;
      const outer = el.renderRoot.querySelector('#outer') as HTMLElement;

      // The probe. #inner is the 3D carrier: preserve-3d plus a perspective of
      // its OWN, so the measurement depends on nothing outside this element
      // and the only variable is what filter #inner carries. One child sits
      // 100px toward the viewer.
      //
      //   perspective 200px, child at z = 100px -> scale 200 / (200 - 100) = 2
      //   flattened -> translateZ discarded, scale 1
      //
      // so a 10px child measures 20px when the 3D context is honored and 10px
      // when the browser forced it flat. Geometric, not declarative: it fails
      // if the scene actually collapses, for any reason, rather than if a
      // particular CSS rule moved.
      inner.style.transformStyle = 'preserve-3d';
      inner.style.transform = 'perspective(200px)';
      const face = document.createElement('div');
      face.style.cssText =
        'position:absolute;left:0;top:0;width:10px;height:10px;transform:translateZ(100px)';
      inner.appendChild(face);
      const measure = () => face.getBoundingClientRect().width;

      // 1. As a token actually renders: altShadow is on for every token
      // (boardgame-token sets it in firstUpdated), so this is the ordinary
      // resting state of every token in the app.
      const resting = {
        innerFilter: getComputedStyle(inner).filter,
        outerFilter: getComputedStyle(outer).filter,
        width: measure(),
      };

      // 2. Still highlighted: the throb is an infinite filter animation, and
      // it must not reintroduce a filter on #inner either.
      el.highlighted = true;
      await el.updateComplete;
      await frame();
      const throbbing = {
        innerFilter: getComputedStyle(inner).filter,
        outerFilter: getComputedStyle(outer).filter,
        width: measure(),
      };
      el.highlighted = false;
      await el.updateComplete;
      await frame();

      // 3. The BEFORE state, reconstructed in place: put the very same
      // elevation filter back on #inner, where `#outer.alt-shadow #inner`
      // used to put it, and re-measure. This is what every token did until
      // this commit.
      const elevation = getComputedStyle(inner)
        .getPropertyValue('--alt-shadow-elevation-normal').trim();
      inner.style.filter = elevation;
      await frame();
      const withFilterOnInner = {
        innerFilter: getComputedStyle(inner).filter,
        width: measure(),
      };

      // 4. And the same element recovers the instant the filter goes away,
      // which pins the filter -- not the probe, the layout or the shadow
      // root -- as the thing doing the flattening.
      inner.style.filter = '';
      await frame();
      const recovered = measure();

      el.remove();
      return { elevation, resting, throbbing, withFilterOnInner, recovered };
    });

    // The elevation is still applied -- just one level up. Losing it entirely
    // would also "fix" the flattening, so this assertion is what makes the
    // rest meaningful.
    expect(result.elevation, 'the alt-shadow elevation custom property must still exist')
      .toContain('drop-shadow');
    expect(result.resting.outerFilter, 'the elevation must now be on #outer')
      .toContain('drop-shadow');
    expect(result.resting.innerFilter, '#inner must carry no filter at rest').toBe('none');
    expect(result.resting.width, 'a preserve-3d context on #inner must be honored')
      .toBeCloseTo(20, 0);

    expect(result.throbbing.innerFilter, 'the throb must not put a filter back on #inner')
      .toBe('none');
    expect(result.throbbing.outerFilter, 'the throb pulses #outer\'s filter')
      .toContain('drop-shadow');
    expect(result.throbbing.width, 'a highlighted token must still hold its 3D scene')
      .toBeCloseTo(20, 0);

    // The before state: this is the assertion that fails on the old code,
    // where resting.innerFilter WAS this same drop-shadow stack.
    expect(result.withFilterOnInner.innerFilter, 'the reconstructed before state')
      .toContain('drop-shadow');
    expect(
      result.withFilterOnInner.width,
      'a filter on #inner forces transform-style: flat, collapsing the scene',
    ).toBeCloseTo(10, 0);

    expect(result.recovered, 'removing the filter restores the 3D context').toBeCloseTo(20, 0);
  });

  // The same trap, in the other component that owns #inner's transform.
  //
  // boardgame-card's ROTATED alt-shadow pair pointed the same drop-shadow stack
  // at #inner. Nothing sets `altShadow` on a card -- not in this repo, not in
  // ../games -- so it was unreachable rather than live, which is precisely the
  // argument for moving it while it still costs nothing: a 3D card has to mount
  // its scene on #inner, and #inner is where the flip's own transform already
  // lives.
  //
  // The probe is the same geometric one, and the properties are driven the way
  // a game would drive them (`altShadow` and `rotated`), not by writing CSS
  // classes by hand.
  test('a preserve-3d context on #inner survives a rotated card\'s alt-shadow', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-card.ts');
      const el = document.createElement('boardgame-card') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;
      // The combination the rules key off, and the one no game currently
      // produces: a rotated card asking for the filter-based elevation.
      el.altShadow = true;
      el.rotated = true;
      await el.updateComplete;
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));
      await frame();

      const inner = el.renderRoot.querySelector('#inner') as HTMLElement;
      const outer = el.renderRoot.querySelector('#outer') as HTMLElement;
      // #inner already carries the card's own flip transform, so the probe
      // supplies its own 3D context on a CHILD of it rather than overwriting
      // that. That is not the same arithmetic as the token's probe, and the
      // assertions below say what it is instead: the card's #outer carries
      // `perspective: 1000px` for the flip, and while #inner really is
      // preserve-3d that outer perspective COMPOUNDS with the probe's own.
      const scene = document.createElement('div');
      scene.style.cssText =
        'position:absolute;left:0;top:0;transform-style:preserve-3d;transform:perspective(200px)';
      const face = document.createElement('div');
      face.style.cssText =
        'position:absolute;left:0;top:0;width:10px;height:10px;transform:translateZ(100px)';
      scene.appendChild(face);
      inner.appendChild(scene);
      await frame();
      const measure = () => face.getBoundingClientRect().width;

      const resting = {
        innerFilter: getComputedStyle(inner).filter,
        outerFilter: getComputedStyle(outer).filter,
        width: measure(),
      };

      // The BEFORE state, reconstructed in place: the same rotated elevation
      // back on #inner, where `#outer.alt-shadow.rotated #inner` used to put it.
      const elevation = getComputedStyle(inner)
        .getPropertyValue('--alt-shadow-elevation-normal-rotated').trim();
      inner.style.filter = elevation;
      await frame();
      const withFilterOnInner = { innerFilter: getComputedStyle(inner).filter, width: measure() };
      inner.style.filter = '';
      await frame();
      const recovered = measure();

      // A disabled rotated card must still look disabled: the elevation and
      // the saturate share one filter slot on #outer now, so the composition
      // has to be spelled out or one silently replaces the other.
      el.disabled = true;
      await el.updateComplete;
      await frame();
      const disabledOuterFilter = getComputedStyle(outer).filter;

      el.remove();
      return { elevation, resting, withFilterOnInner, recovered, disabledOuterFilter };
    });

    expect(result.elevation, 'the rotated alt-shadow elevation must still exist')
      .toContain('drop-shadow');
    expect(result.resting.outerFilter, 'the rotated elevation must now be on #outer')
      .toContain('drop-shadow');
    expect(result.resting.innerFilter, '#inner must carry no filter at rest').toBe('none');
    // 25, NOT the token's 20, and the difference is real rather than noise.
    // The 10px face at translateZ(100px) is magnified 2x by the probe's own
    // `perspective(200px)` and then a further 1.25x by #outer's
    // `perspective: 1000px`, which reaches it because #inner is preserve-3d.
    // Measured directly: setting `#outer { perspective: none }` in place drops
    // this same probe from 25 to exactly 20. `boardgame-token`'s #outer has no
    // perspective, which is why the test above reads 20 and this one does not.
    expect(result.resting.width, 'a preserve-3d context under #inner must be honored')
      .toBeCloseTo(25, 0);

    expect(result.withFilterOnInner.innerFilter, 'the reconstructed before state')
      .toContain('drop-shadow');
    // 20, and this is the whole tooth: a filter on #inner takes #inner out of
    // the card's own 3D rendering context, so the probe keeps its own 2x and
    // loses the 1.25x that came from #outer's perspective. It is a smaller
    // signal than the token's 20 -> 10, because the card's probe cannot mount
    // ON #inner (the flip transform lives there) and so only the OUTER
    // contribution is at stake -- but it is deterministic, it is 25%, and it is
    // exactly the regression this test exists for: putting the rotated
    // alt-shadow back on #inner reproduces this number.
    expect(
      result.withFilterOnInner.width,
      'a filter on #inner takes it out of the card\'s 3D context',
    ).toBeCloseTo(20, 0);
    expect(result.recovered, 'removing the filter restores the 3D context').toBeCloseTo(25, 0);

    expect(result.disabledOuterFilter, 'a disabled rotated card keeps its elevation')
      .toContain('drop-shadow');
    expect(result.disabledOuterFilter, 'and still looks disabled').toContain('saturate');
  });
});
