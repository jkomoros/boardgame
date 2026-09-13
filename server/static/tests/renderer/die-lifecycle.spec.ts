import { test, expect, type Page } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test.use({ reducedMotion: 'no-preference' });

async function mount(page: Page, direct = false) {
  await prepareRendererFixturePage(page);
  await page.evaluate(async (direct) => {
    await import('/src/components/boardgame-die.ts');
    const die = document.createElement('boardgame-die');
    die.id = 'die';
    const faces = [1, 2, 3, 4, 5, 6];
    if (direct) { die.faces = faces; die.selectedFaceIndex = 3; }
    else die.item = { ID: 'die', Values: { Faces: faces }, DynamicValues: { SelectedFace: 3, Value: 4, RollCount: 0 } } as any;
    document.body.append(die);
    for (let i = 0; i < 4; i++) await die.updateComplete;
  }, direct);
}

async function roll(page: Page, count: number) {
  await page.evaluate(async count => {
    const die = document.querySelector('boardgame-die')!;
    die.item = { ...die.item, DynamicValues: { SelectedFace: 3, Value: 4, RollCount: count } } as any;
    for (let i = 0; i < 4; i++) await die.updateComplete;
  }, count);
}

test('direct properties mount quietly and replace the shape without keeping its old roll', async ({ page }) => {
  await mount(page, true);
  expect(await page.locator('#die').evaluate(die => die.shadowRoot!.querySelector('#inner')!.getAnimations().length)).toBe(0);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    die.selectedFaceIndex = 4;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    const old = die.shadowRoot!.querySelector('#inner')!.getAnimations();
    die.faces = [10, 20, 30, 40];
    die.selectedFaceIndex = 2;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    return { oldCount: old.length, oldStates: old.map(a => a.playState), value: die.value, live: die.shadowRoot!.querySelector('#inner')!.getAnimations().length };
  });
  expect(result).toEqual({ oldCount: 1, oldStates: ['idle'], value: 30, live: 0 });
});

test('a repeated result clears the old announcement before the next tumble', async ({ page }) => {
  await mount(page);
  await roll(page, 1);
  await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    die.finishGatedAnimations(); await die.settled(); await die.updateComplete;
  });
  await expect(page.locator('#die [aria-live]')).toHaveText('Rolled 4');
  await roll(page, 2);
  await expect(page.locator('#die [aria-live]')).toHaveText('');
  await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    die.finishGatedAnimations(); await die.settled(); await die.updateComplete;
  });
  await expect(page.locator('#die [aria-live]')).toHaveText('Rolled 4');
});

test('cancelling the current tumble lands it and releases the gate', async ({ page }) => {
  await mount(page); await roll(page, 1);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    let ends = 0; die.addEventListener('roll-end', () => ends++);
    die.shadowRoot!.querySelector('#inner')!.getAnimations().forEach(a => a.cancel());
    await die.settled(); await die.updateComplete;
    return { value: die.value, ends, animating: die.isAnimating, label: die.shadowRoot!.querySelector('#main')!.getAttribute('aria-label') };
  });
  expect(result).toEqual({ value: 4, ends: 1, animating: false, label: 'Die showing 4' });
});

test('a new throw owns the only tumble and cancels the preceding landing accent', async ({ page }) => {
  await mount(page); await roll(page, 1);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    const inner = die.shadowRoot!.querySelector('#inner')!;
    const old = inner.getAnimations()[0];
    die.item = { ...die.item, DynamicValues: { SelectedFace: 4, Value: 5, RollCount: 2 } } as any;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    const oldState = old.playState; const live = inner.getAnimations().length;
    die.finishGatedAnimations(); await die.settled(); await die.updateComplete;
    const accent = die.shadowRoot!.querySelector('#stage')!.getAnimations()[0];
    die.item = { ...die.item, DynamicValues: { SelectedFace: 3, Value: 4, RollCount: 3 } } as any;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    return { oldState, live, hadAccent: !!accent, accentState: accent?.playState, accents: die.shadowRoot!.querySelector('#stage')!.getAnimations().length };
  });
  expect(result).toEqual({ oldState: 'idle', live: 1, hadAccent: true, accentState: 'idle', accents: 0 });
});

for (const replacement of ['component', 'hidden', 'correction'] as const) {
  test(`a ${replacement} snapshot supersedes a running die without a new throw`, async ({ page }) => {
    await mount(page); await roll(page, 1);
    const result = await page.evaluate(async replacement => {
      const die = document.querySelector('boardgame-die')!;
      const old = die.shadowRoot!.querySelector('#inner')!.getAnimations()[0];
      die.item = replacement === 'hidden' ? {} : {
        ...die.item, ID: replacement === 'component' ? 'new-die' : 'die',
        DynamicValues: { SelectedFace: 1, Value: 2, RollCount: 1 },
      } as any;
      for (let i = 0; i < 4; i++) await die.updateComplete;
      await die.settled();
      return { oldState: old.playState, value: die.value, live: die.shadowRoot!.querySelector('#inner')!.getAnimations().length, label: die.shadowRoot!.querySelector('#main')!.getAttribute('aria-label') };
    }, replacement);
    expect(result.oldState).toBe('idle'); expect(result.live).toBe(0);
    expect(result.value).toBe(replacement === 'hidden' ? null : 2);
    expect(result.label).not.toContain('rolling');
  });
}

test('reduced motion reports only the landed result and disables hover transitions', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' }); await mount(page);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    const events: string[] = [];
    die.addEventListener('roll-start', () => events.push('start'));
    die.addEventListener('roll-end', () => events.push('end'));
    die.item = { ...die.item, DynamicValues: { SelectedFace: 3, Value: 4, RollCount: 1 } } as any;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    await die.settled();
    return { events, value: die.value, transition: getComputedStyle(die.shadowRoot!.querySelector('#main')!).transitionDuration };
  });
  expect(result).toEqual({ events: ['end'], value: 4, transition: '0s' });
});

test('disconnecting during the landing accent releases its animation', async ({ page }) => {
  await mount(page); await roll(page, 1);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    die.finishGatedAnimations(); await die.settled(); await die.updateComplete;
    const accent = die.shadowRoot!.querySelector('#stage')!.getAnimations()[0];
    die.remove(); await Promise.resolve(); await Promise.resolve();
    return { hadAccent: !!accent, state: accent?.playState };
  });
  expect(result.hadAccent).toBe(true); expect(['idle', 'finished']).toContain(result.state);
});

for (const policy of ['normal', 'disabled', 'reduced'] as const) {
  test(`the reel fallback has the same completion contract under ${policy} motion`, async ({ page }) => {
    if (policy === 'reduced') await page.emulateMedia({ reducedMotion: 'reduce' });
    await prepareRendererFixturePage(page);
    const result = await page.evaluate(async policy => {
      await import('/src/components/boardgame-die.ts');
      const die = document.createElement('boardgame-die');
      die.noAnimate = policy === 'disabled';
      die.item = { ID: 'coin', Values: { Faces: [10, 20] }, DynamicValues: { SelectedFace: 0, RollCount: 0 } } as any;
      document.body.append(die);
      for (let i = 0; i < 4; i++) await die.updateComplete;
      const events: string[] = [];
      die.addEventListener('roll-start', () => events.push('start'));
      die.addEventListener('roll-end', event => events.push(`end:${(event as CustomEvent).detail.value}`));
      die.item = { ...die.item, DynamicValues: { SelectedFace: 1, RollCount: 1 } } as any;
      for (let i = 0; i < 4; i++) await die.updateComplete;
      await die.settled(); await die.updateComplete;
      const first = events.slice(); events.length = 0;
      die.item = { ...die.item, DynamicValues: { SelectedFace: 1, RollCount: 2 } } as any;
      for (let i = 0; i < 4; i++) await die.updateComplete;
      await die.settled(); await die.updateComplete;
      return { first, repeated: events, value: die.value, announcement: die.shadowRoot!.querySelector('[aria-live]')!.textContent?.trim() };
    }, policy);
    expect(result.first).toEqual(policy === 'normal' ? ['start', 'end:20'] : ['end:20']);
    expect(result.repeated).toEqual(['end:20']);
    expect(result.value).toBe(20); expect(result.announcement).toBe('Rolled 20');
  });
}

test('switching between item and direct authoring installs a new baseline', async ({ page }) => {
  await mount(page, true);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    let starts = 0; die.addEventListener('roll-start', () => starts++);
    die.item = { Values: { Faces: [1, 2, 3, 4, 5, 6] }, DynamicValues: { SelectedFace: 1 } } as any;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    die.selectedFaceIndex = 5;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    const authoritative = die.value;
    die.item = null; die.faces = [10, 20, 30, 40]; die.selectedFaceIndex = 2;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    return { authoritative, direct: die.value, starts, live: die.shadowRoot!.querySelector('#inner')!.getAnimations().length };
  });
  expect(result).toEqual({ authoritative: 2, direct: 30, starts: 0, live: 0 });
});

test('an unmeasurable die still reports its authoritative result', async ({ page }) => {
  await mount(page);
  const result = await page.evaluate(async () => {
    const die = document.querySelector('boardgame-die')!;
    die.style.setProperty('--die-size', '0px');
    const events: string[] = [];
    die.addEventListener('roll-start', () => events.push('start'));
    die.addEventListener('roll-end', event => events.push(`end:${(event as CustomEvent).detail.value}`));
    die.item = { ...die.item, DynamicValues: { SelectedFace: 2, Value: 3, RollCount: 1 } } as any;
    for (let i = 0; i < 4; i++) await die.updateComplete;
    return { events, value: die.value, live: die.isAnimating };
  });
  expect(result).toEqual({ events: ['end:3'], value: 3, live: false });
});

for (const source of ['item', 'direct'] as const) {
  test(`invalid ${source} reel indexes draw the same fallback they announce`, async ({ page }) => {
    await prepareRendererFixturePage(page);
    const results = await page.evaluate(async source => {
      await import('/src/components/boardgame-die.ts');
      const results = [];
      for (const index of [0.5, 99]) {
        const die = document.createElement('boardgame-die');
        if (source === 'item') die.item = { Values: { Faces: [10, 20] }, DynamicValues: { SelectedFace: index } } as any;
        else { die.faces = [10, 20]; die.selectedFaceIndex = index; }
        document.body.append(die);
        for (let i = 0; i < 4; i++) await die.updateComplete;
        results.push({ value: die.value, index: die.shadowRoot!.querySelector<HTMLElement>('#main')!.style.getPropertyValue('--selected-face-index') });
      }
      return results;
    }, source);
    expect(results).toEqual([{ value: 10, index: '0' }, { value: 10, index: '0' }]);
  });
}
