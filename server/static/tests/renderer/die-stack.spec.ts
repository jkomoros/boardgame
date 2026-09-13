import { test, expect } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test.use({ reducedMotion: 'no-preference' });

async function mount(page: import('@playwright/test').Page, count = 5, budget = false) {
  await prepareRendererFixturePage(page);
  await page.evaluate(async ({ count, budget }) => {
    const { dieView } = await import('/src/components/component-view.ts');
    await import('/src/components/boardgame-component-stack.ts');
    const die = (id: string, roll = 0) => ({ ID: id, Index: 0, Deck: 'dice', GameName: 'dice-fixture',
      Values: { Faces: [1, 2, 3, 4, 5, 6] }, DynamicValues: { SelectedFace: 3, Value: 4, RollCount: roll } });
    const state = (items: any[]) => ({ Deck: 'dice', GameName: 'dice-fixture', Components: items,
      Indexes: items.map((_, i) => i), IDs: items.map(item => item?.ID ?? ''), IDsLastSeen: {}, ShuffleCount: 0, Size: items.length });
    const stack = document.createElement('boardgame-component-stack');
    stack.componentView = dieView(budget ? { rollBudget: { durationMs: 900, maxSolidDice: 5 } } : {});
    stack.layout = 'spread';
    stack.stack = state(Array.from({ length: count }, (_, i) => die(`die-${i}`)));
    document.body.append(stack);
    const drain = async () => { for (let i = 0; i < 6; i++) {
      await stack.updateComplete;
      await Promise.all([...stack.querySelectorAll('boardgame-die')].map(d => d.updateComplete));
    }};
    await drain();
    (window as any).diceFixture = { stack, die, state, drain };
  }, { count, budget });
}

test('dice are real component hosts; repeated rolls affect only changed counters', async ({ page }) => {
  await mount(page);
  const result = await page.evaluate(async () => {
    const { BoardgameComponent } = await import('/src/components/boardgame-component.ts');
    const { stack, die, state, drain } = (window as any).diceFixture;
    const before = [...stack.querySelectorAll('boardgame-die')] as any[];
    const initialized = before.every(d => d instanceof BoardgameComponent && !d.isAnimating && d.value === 4);
    const starts: string[] = [];
    stack.addEventListener('roll-start', (e: Event) => starts.push((e.target as HTMLElement).id));
    stack.stack = state([die('die-0'), die('die-1', 1), die('die-2'), die('die-3', 1), die('die-4')]);
    await drain();
    const after = [...stack.querySelectorAll('boardgame-die')] as any[];
    const result = { initialized, retained: before.every((d, i) => d === after[i]), starts,
      nestedButtons: stack.querySelectorAll('boardgame-die').length && after.some(d => d.shadowRoot.querySelector('button')),
      values: after.map(d => d.value), rollTracks: after.map(d => d.shadowRoot.querySelector('#inner').getAnimations().length) };
    after.forEach(d => d.finishGatedAnimations());
    await Promise.all(after.map(d => d.settled()));
    return { ...result, released: after.every(d => !d.isAnimating) };
  });
  expect(result).toEqual({ initialized: true, retained: true, starts: ['die-1', 'die-3'], nestedButtons: false,
    values: [4, 4, 4, 4, 4], rollTracks: [0, 1, 0, 1, 0], released: true });
});

test('hidden, empty, and replacement slots clear roll state and historical carriers preserve only display', async ({ page }) => {
  await mount(page);
  const result = await page.evaluate(async () => {
    const { captureHistoricalPresentation, installHistoricalPresentation, clearHistoricalPresentation } = await import('/src/motion/historical-presentation.ts');
    const { stack, die, state, drain } = (window as any).diceFixture;
    stack.stack = state([die('die-0', 1), die('die-1'), die('die-2'), die('die-3'), die('die-4')]);
    await drain();
    const source = stack.querySelector('boardgame-die');
    source.finishGatedAnimations(); await source.settled(); await source.updateComplete;
    const appearance = captureHistoricalPresentation(source)!;
    const carrier = stack.newMotionCarrier().component;
    const installed = installHistoricalPresentation(carrier, appearance);
    for (let i = 0; i < 4; i++) await carrier.updateComplete;
    const history = { installed, value: carrier.value, item: carrier.item ?? null,
      animations: carrier.shadowRoot.querySelector('#inner').getAnimations().length };
    clearHistoricalPresentation(carrier);
    for (let i = 0; i < 4; i++) await carrier.updateComplete;
    const cleared = carrier.value;
    stack.stack = state([{ ID: 'die-0' }, null, die('replacement', 8)]);
    await drain();
    const current = [...stack.querySelectorAll('boardgame-die')] as any[];
    return { history, cleared, values: current.map(d => d.value), spacers: current.map(d => d.spacer),
      animated: current.some(d => d.isAnimating), hiddenText: current[0].shadowRoot.textContent.includes('Rolled') };
  });
  expect(result.history).toEqual({ installed: true, value: 4, item: null, animations: 0 });
  expect(result.cleared).toBe(null);
  expect(result.values).toEqual([null, null, 4]);
  expect(result.spacers).toEqual([false, true, false]);
  expect(result.animated).toBe(false);
  expect(result.hiddenText).toBe(false);
});

for (const budget of [false, true]) for (const count of [5, 20]) test(`${count} concurrent dice (budget ${budget}) report bounded duration and allocation measurements`, async ({ page }) => {
  await mount(page, count, budget);
  const result = await page.evaluate(async count => {
    const { stack, die, state, drain } = (window as any).diceFixture;
    const started = performance.now();
    stack.stack = state(Array.from({ length: count }, (_, i) => die(`die-${i}`, 1)));
    await drain();
    const dice = [...stack.querySelectorAll('boardgame-die')] as any[];
    const tracks = dice.flatMap(d => d.shadowRoot.querySelector('#inner').getAnimations()) as Animation[];
    const durations = tracks.map(a => Number(a.effect!.getTiming().duration));
    const measurement = { count, planningMs: performance.now() - started, tracks: tracks.length,
      maximumDurationMs: Math.max(...durations), keyframes: tracks.reduce((n, a) => n + (a.effect as KeyframeEffect).getKeyframes().length, 0),
      facets: dice.reduce((n, d) => n + d.shadowRoot.querySelectorAll('.facet').length, 0) };
    dice.forEach(d => d.finishGatedAnimations()); await Promise.all(dice.map(d => d.settled()));
    return { ...measurement, released: dice.every(d => !d.isAnimating), values: dice.map(d => d.value) };
  }, count);
  console.log('dice pool measurement', result);
  expect(result.tracks).toBe(budget ? Math.min(count, 5) : count);
  expect(result.maximumDurationMs).toBeLessThanOrEqual(budget ? 900 : 5000);
  expect(result.facets).toBe(6 * (budget ? Math.min(count, 5) : count));
  expect(result.released).toBe(true);
  expect(result.values).toEqual(Array(count).fill(4));
});

test('host structural motion composes with the roll and reduced motion avoids all simulation tracks', async ({ page }) => {
  await mount(page);
  const result = await page.evaluate(async () => {
    const { stack, die, state, drain } = (window as any).diceFixture;
    stack.stack = state([die('die-0', 1)]); await drain();
    const host = stack.querySelector('boardgame-die');
    const roll = host.shadowRoot.querySelector('#inner').getAnimations()[0];
    const tracks = host.playAnimation({ before: {}, after: {}, invertedTransform: 'translateX(100px)',
      finalTransform: '', beforeOpacity: '', finalOpacity: '', needsHostTransition: true, durationMs: 500, timingPolicy: 'immediate' });
    const targets = tracks.map((a: Animation) => (a.effect as KeyframeEffect).target === host);
    const rollStillOwned = host.shadowRoot.querySelector('#inner').getAnimations()[0] === roll;
    host.finishGatedAnimations(); await host.settled();
    return { targets, rollStillOwned, released: !host.isAnimating, value: host.value };
  });
  expect(result).toEqual({ targets: [true], rollStillOwned: true, released: true, value: 4 });
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const quiet = await page.evaluate(async () => {
    const { stack, die, state, drain } = (window as any).diceFixture;
    stack.stack = state([die('die-0', 2)]); await drain();
    const host = stack.querySelector('boardgame-die');
    return { track: host._roll?.track, value: host.value, running: host.isAnimating };
  });
  expect(quiet).toEqual({ track: null, value: 4, running: false });
});

test('hidden slots do not spend the solid budget and moving across its boundary keeps correct faces', async ({ page }) => {
  await mount(page, 6, true);
  const result = await page.evaluate(async () => {
    const { stack, die, state, drain } = (window as any).diceFixture;
    stack.stack = state([{ ID: 'hidden' }, die('die-5'), die('die-4'), die('die-3'), die('die-2'), die('die-1')]);
    await drain();
    const first = [...stack.querySelectorAll('boardgame-die')] as any[];
    const solids = first.map(d => !!d.shadowRoot.querySelector('#inner.solid'));
    const starts: string[] = []; stack.addEventListener('roll-start', (e: Event) => starts.push((e.target as HTMLElement).id));
    stack.stack = state([die('die-0'), die('die-1'), die('die-2'), die('die-3'), die('die-4'), die('die-5')]);
    await drain();
    const second = [...stack.querySelectorAll('boardgame-die')] as any[];
    return { solids, starts, values: second.map(d => d.value), lastSolid: !!second[5].shadowRoot.querySelector('#inner.solid') };
  });
  expect(result).toEqual({ solids: [false, true, true, true, true, true], starts: [], values: [4,4,4,4,4,4], lastSolid: false });
});

test('pool budgets reject invalid bounds before they can allocate roll work', async ({ page }) => {
  await prepareRendererFixturePage(page);
  const errors = await page.evaluate(async () => {
    const { dieView } = await import('/src/components/component-view.ts');
    return [{ durationMs: NaN, maxSolidDice: 5 }, { durationMs: 900, maxSolidDice: 1.5 },
      { durationMs: 900, maxSolidDice: 33 }, { durationMs: 5001, maxSolidDice: 5 }].map(rollBudget => {
      try { dieView({ rollBudget }); return ''; } catch (error) { return String(error); }
    });
  });
  expect(errors.every(message => message.includes('rollBudget requires'))).toBe(true);
});

test('retained hosts keep their landed pose through cosmetic recipe updates', async ({ page }) => {
  await mount(page);
  const result = await page.evaluate(async () => {
    const { stack, die, state, drain } = (window as any).diceFixture;
    stack.stack = state([die('die-0', 1)]); await drain();
    const host = stack.querySelector('boardgame-die');
    host.finishGatedAnimations(); await host.settled(); await host.updateComplete;
    const resting = host.shadowRoot.querySelector('#inner').style.transform;
    let starts = 0; host.addEventListener('roll-start', () => starts++);
    stack.componentView = stack.componentView.withProperties({ symbols: { '4': '★' } });
    await drain();
    return { retained: host === stack.querySelector('boardgame-die'), samePose: host.shadowRoot.querySelector('#inner').style.transform === resting,
      value: host.value, starts, symbol: host.shadowRoot.textContent.includes('★') };
  });
  expect(result).toEqual({ retained: true, samePose: true, value: 4, starts: 0, symbol: true });
});
