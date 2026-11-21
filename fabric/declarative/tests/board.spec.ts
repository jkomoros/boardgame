/**
 * The same assertions for every way of driving the Memory board:
 * in place (page A), with recreated children in three modes (page B), and on
 * each of two tables that share a page (page C).
 */
import { expect, test } from '@playwright/test';
import { expected, progress, ready } from './helpers.js';

interface Variant {
  readonly name: string;
  readonly url: string;
  readonly replay: string;
  readonly side?: string;
  readonly scopeSelector: string;
}

const variants: Variant[] = [
  { name: 'A in place', url: '/memory.html?auto=0', replay: 'memory-a', scopeSelector: 'bg-table' },
  { name: 'B recreate every stack', url: '/memory-recreate.html?auto=0&mode=all', replay: 'memory-a', scopeSelector: 'bg-table' },
  { name: 'B recreate changed stacks', url: '/memory-recreate.html?auto=0&mode=changed', replay: 'memory-a', scopeSelector: 'bg-table' },
  { name: 'B recreate the table', url: '/memory-recreate.html?auto=0&mode=table', replay: 'memory-a', scopeSelector: 'bg-table' },
  { name: 'C left table', url: '/two-tables.html?auto=0', replay: 'memory-a', side: 'left', scopeSelector: '#left' },
  { name: 'C right table', url: '/two-tables.html?auto=0', replay: 'memory-b', side: 'right', scopeSelector: '#right' },
];

for (const variant of variants) {
  test.describe(variant.name, () => {
    const want = expected(variant.replay);
    const side = variant.side;

    test('cards render with the right faces, and a reveal turns one face up', async ({ page }) => {
      const errors = await ready(page, variant.url);
      await page.evaluate(([s]) => (window as any).probe.handle(s).table.whenSettled(), [side]);
      const setup = await page.evaluate(([s]) => (window as any).probe.grid(s), [side]);
      expect(setup).toHaveLength(20);
      expect(setup.every((card: any) => !card.faceUp && card.text === '')).toBe(true);

      const hiddenBefore = setup.find((card: any) => card.index === want.revealSlot);
      expect(hiddenBefore.faceUp).toBe(false);

      const track = await page.evaluate(async ([s, index]) => {
        const probe = (window as any).probe;
        await probe.stepTo((index as number) - 1, s);
        return probe.advanceAndTrack([], s);
      }, [side, want.firstReveal] as const);
      const grid = await page.evaluate(([s]) => (window as any).probe.grid(s), [side]);
      const revealed = grid.find((card: any) => card.index === want.revealSlot);
      expect(revealed.faceUp).toBe(true);
      expect(revealed.text).toBe(want.revealFace);
      // Every other card is still face down and shows nothing.
      expect(grid.filter((card: any) => card.faceUp)).toHaveLength(1);
      expect(grid.filter((card: any) => !card.faceUp).every((card: any) => card.text === '')).toBe(true);
      // The flip was a real animation of a transform, not a property snap.
      expect(track.transformAnimations).toBeGreaterThan(0);
      expect(track.ms).toBeGreaterThan(200);
      expect(errors).toEqual([]);
    });

    test('a captured pair flies to the right won-cards stack', async ({ page }) => {
      const errors = await ready(page, variant.url);
      const before = await page.evaluate(async ([s, index]) => {
        const probe = (window as any).probe;
        await probe.stepTo(index as number, s);
        return { grid: probe.grid(s), won: [probe.won(0, s), probe.won(1, s)] };
      }, [side, want.bothUp] as const);
      expect(before.grid.filter((card: any) => card.faceUp).map((card: any) => card.id).sort())
        .toEqual([...want.pairIds].sort());
      expect(before.won[want.winner]).toHaveLength(0);

      const track = await page.evaluate(([s, ids]) =>
        (window as any).probe.advanceAndTrack(ids, s), [side, want.pairIds] as const);

      const after = await page.evaluate(([s]) => {
        const probe = (window as any).probe;
        return { grid: probe.grid(s), won: [probe.won(0, s), probe.won(1, s)] };
      }, [side]);
      expect(after.won[want.winner].map((card: any) => card.id).sort())
        .toEqual([...want.pairIds].sort());
      expect(after.won[want.winner].every((card: any) =>
        card.faceUp && card.text === want.pairFace)).toBe(true);
      expect(after.won[1 - want.winner]).toHaveLength(0);
      expect(after.grid).toHaveLength(18);

      // An animation ran between the two states: it was not a jump cut.
      expect(track.transformAnimations).toBeGreaterThan(0);
      for (let which = 0; which < want.pairIds.length; which++) {
        const { travel, fractions } = progress(track, which);
        expect(travel).toBeGreaterThan(50);
        // The first painted frame is still near where the card WAS.
        expect(fractions[0]!).toBeLessThan(0.35);
        // It passed through the middle of its path on several frames.
        expect(fractions.filter(f => f > 0.1 && f < 0.9).length).toBeGreaterThanOrEqual(3);
        // No painted frame shows the card at its destination and then back
        // on its path: progress never drops by more than rounding noise.
        const backwards = fractions.slice(1).map((f, i) => fractions[i]! - f);
        expect(Math.max(...backwards)).toBeLessThan(0.05);
        // And it ended where the new state puts it.
        expect(fractions[fractions.length - 1]!).toBeGreaterThan(0.98);
      }
      expect(errors).toEqual([]);
    });

    test('a click on a face-down card emits propose-move with JSON detail', async ({ page }) => {
      const errors = await ready(page, variant.url);
      await page.evaluate(async ([s, index]) => (window as any).probe.stepTo(index as number, s),
        [side, want.firstReveal] as const);
      const grid = await page.evaluate(([s]) => (window as any).probe.grid(s), [side]);
      const target = grid.find((card: any) => card.index === want.hiddenSlot);
      expect(target.role).toBe('button');
      expect(target.disabled).toBe(false);

      // A script-made click is not a gesture and proposes nothing.
      await page.evaluate(([s, id]) => {
        const card = (window as any).probe.allCards(s).find((c: any) => c.id === id);
        card.shadowRoot.querySelector('#outer').click();
      }, [side, target.id] as const);
      expect(await page.evaluate(() => (window as any).proposals)).toEqual([]);

      // The already revealed card is not a target.
      const cards = page.locator(`${variant.scopeSelector} bg-stack`).first()
        .locator('boardgame-card[boardgame-component]');
      await cards.nth(want.revealSlot).click({ force: true });
      expect(await page.evaluate(() => (window as any).proposals)).toEqual([]);

      await cards.nth(want.hiddenSlot).click();
      const proposals = await page.evaluate(() => (window as any).proposals);
      expect(proposals).toHaveLength(1);
      expect(proposals[0]).toMatchObject({
        move: 'Reveal Card',
        args: { CardIndex: want.hiddenSlot },
        index: want.hiddenSlot,
        version: want.versions[want.firstReveal],
      });
      // The detail survives a JSON round trip unchanged: it is data.
      expect(JSON.parse(JSON.stringify(proposals[0]))).toEqual(proposals[0]);
      // Nothing that carries layout or timing reached the document.
      expect(await page.evaluate(() => (window as any).leaked)).toEqual([]);
      expect(errors).toEqual([]);
    });
  });
}
