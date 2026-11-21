/**
 * Characterization tests: each one pins a finding in FINDINGS.md, including
 * what goes WRONG when a mechanism the table adds is switched off.
 */
import { expect, test } from '@playwright/test';
import { expected, progress, ready } from './helpers.js';

const want = expected('memory-a');

test.describe('recreation', () => {
  for (const mode of ['all', 'changed', 'table']) {
    test(`without remembered geometry, mode=${mode} is a jump cut`, async ({ page }) => {
      await ready(page, `/memory-recreate.html?auto=0&mode=${mode}&measure=live`);
      const track = await page.evaluate(async ([index, ids]) => {
        const probe = (window as any).probe;
        await probe.stepTo(index as number);
        return probe.advanceAndTrack(ids);
      }, [want.bothUp, want.pairIds] as const);
      // The pair still ends up in the right stack...
      const won = await page.evaluate(([player]) => (window as any).probe.won(player), [want.winner]);
      expect(won.map((card: any) => card.id).sort()).toEqual([...want.pairIds].sort());
      // ...but it never passes through the middle of the path.
      for (let which = 0; which < 2; which++) {
        const { fractions } = progress(track, which);
        expect(fractions.filter(f => f > 0.1 && f < 0.9)).toHaveLength(0);
      }
    });
  }

  test('every card node really is replaced in mode=all', async ({ page }) => {
    await ready(page, '/memory-recreate.html?auto=0&mode=all');
    const same = await page.evaluate(async ([index]) => {
      const probe = (window as any).probe;
      await probe.stepTo(index as number);
      const before = new Set(probe.allCards());
      await probe.advanceAndTrack([]);
      return probe.allCards().filter((card: any) => before.has(card)).length;
    }, [want.bothUp] as const);
    expect(same).toBe(0);
  });

  test('a state arriving mid-flight still ends correct and animates', async ({ page }) => {
    const errors = await ready(page, '/memory-recreate.html?auto=0&mode=all');
    const result = await page.evaluate(async ([index, ids]) => {
      const probe = (window as any).probe;
      const d = (window as any).declarative;
      await probe.stepTo((index as number) - 1);
      // Deliver "both up" and then the capture 120 ms later, far inside the 400 ms flip.
      d.next();
      await new Promise(resolve => setTimeout(resolve, 120));
      const track = await probe.advanceAndTrack(ids);
      return { track, stats: { ...d.table.stats } };
    }, [want.bothUp, want.pairIds] as const);
    expect(result.stats.interrupted).toBeGreaterThan(0);
    const won = await page.evaluate(([player]) => (window as any).probe.won(player), [want.winner]);
    expect(won.map((card: any) => card.id).sort()).toEqual([...want.pairIds].sort());
    for (let which = 0; which < 2; which++) {
      const { fractions } = progress(result.track, which);
      expect(fractions.filter(f => f > 0.1 && f < 0.9).length).toBeGreaterThanOrEqual(3);
    }
    expect(errors).toEqual([]);
  });

  test('the "Match" callout survives being recreated with a new trigger', async ({ page }) => {
    await ready(page, '/memory-recreate.html?auto=0&mode=all');
    const played = await page.evaluate(async ([index]) => {
      const probe = (window as any).probe;
      const d = (window as any).declarative;
      await probe.stepTo(index as number);
      d.next();
      let seen = false;
      for (let i = 0; i < 30 && !seen; i++) {
        await new Promise(resolve => requestAnimationFrame(resolve));
        const callout = d.table.querySelector('bg-callout');
        const text = callout?.shadowRoot?.querySelector('boardgame-fading-text');
        seen = !!text?.shadowRoot?.getAnimations().length;
      }
      await d.table.whenSettled();
      return seen;
    }, [want.bothUp] as const);
    expect(played).toBe(true);
  });
});

test.describe('hints as data', () => {
  test('a stagger cohort given as JSON delays the second card', async ({ page }) => {
    await ready(page, '/memory.html?auto=0');
    const delays = await page.evaluate(async ([index, ids]) => {
      const probe = (window as any).probe;
      const d = (window as any).declarative;
      await probe.stepTo(index as number);
      d.frames[(index as number) + 1].transition = {
        move: 'Capture Cards',
        cohorts: [{ kind: 'stagger', subjects: ids, intervalMs: 150 }],
      };
      d.next();
      const seen = new Set<number>();
      for (let i = 0; i < 80; i++) {
        await new Promise(resolve => requestAnimationFrame(resolve));
        for (const animation of probe.animations()) {
          seen.add(Number((animation.effect as KeyframeEffect).getTiming().delay));
        }
      }
      await d.table.whenSettled();
      return [...seen].sort((a, b) => a - b);
    }, [want.bothUp, want.pairIds] as const);
    expect(delays).toContain(0);
    expect(delays).toContain(150);
  });
});

test.describe('geometry', () => {
  for (const [label, css] of [
    ['a scaled ancestor', 'transform: scale(0.7); transform-origin: top left;'],
    ['a scrolled container', 'height: 300px; overflow: auto;'],
  ] as const) {
    test(`a flight inside ${label} starts and ends in the right place`, async ({ page }) => {
      await ready(page, '/memory.html?auto=0');
      const track = await page.evaluate(async ([index, ids, style]) => {
        const probe = (window as any).probe;
        const d = (window as any).declarative;
        const wrapper = document.createElement('div');
        wrapper.setAttribute('style', style as string);
        d.table.before(wrapper);
        wrapper.append(d.table);
        wrapper.scrollTop = 60;
        await new Promise(resolve => setTimeout(resolve, 50));
        await probe.stepTo(index as number);
        return probe.advanceAndTrack(ids);
      }, [want.bothUp, want.pairIds, css] as const);
      for (let which = 0; which < 2; which++) {
        const { fractions } = progress(track, which);
        // The table holds 600 ms before a capture, so skip the still frames.
        const moving = fractions.filter(f => f > 0.02);
        expect(moving[0]!).toBeLessThan(0.35);
        expect(Math.max(...fractions)).toBeLessThan(1.05);
        expect(Math.min(...fractions)).toBeGreaterThan(-0.05);
        expect(fractions[fractions.length - 1]!).toBeGreaterThan(0.98);
      }
    });
  }
});

test.describe('two tables', () => {
  const running = (side: string) => `(window.probe.animations('${side}').length)`;

  test('scoped: one table installing a state does not touch the other mid-flight', async ({ page }) => {
    const errors = await ready(page, '/two-tables.html?auto=0');
    const counts = await page.evaluate(async ([index]) => {
      const probe = (window as any).probe;
      const { left, right } = (window as any).declarative;
      await probe.stepTo(index as number, 'left');
      await right.table.whenSettled();
      left.next(); // the capture: two cards fly for 600 ms
      await new Promise(resolve => setTimeout(resolve, 200));
      const before = probe.animations('left').length;
      right.next(); // the other table installs a state
      await new Promise(resolve => setTimeout(resolve, 80));
      const after = probe.animations('left').length;
      await left.table.whenSettled(); await right.table.whenSettled();
      return { before, after, left: { ...left.table.stats }, right: { ...right.table.stats } };
    }, [want.bothUp] as const);
    expect(counts.before).toBeGreaterThan(0);
    expect(counts.after).toBe(counts.before);
    expect(errors).toEqual([]);
  });

  test('unscoped (today): the other table\'s install cuts this table\'s animation short', async ({ page }) => {
    await ready(page, '/two-tables.html?auto=0&scope=document');
    const counts = await page.evaluate(async ([index]) => {
      const probe = (window as any).probe;
      const { left, right } = (window as any).declarative;
      await probe.stepTo(index as number, 'left');
      await right.table.whenSettled();
      left.next();
      await new Promise(resolve => setTimeout(resolve, 200));
      const before = probe.animations('left').length;
      right.next();
      await new Promise(resolve => setTimeout(resolve, 80));
      const after = probe.animations('left').length;
      await left.table.whenSettled(); await right.table.whenSettled();
      return { before, after };
    }, [want.bothUp] as const);
    expect(counts.before).toBeGreaterThan(0);
    expect(counts.after).toBe(0);
  });

  test('unscoped (today): when both tables change together, each animator also plays the other\'s cards', async ({ page }) => {
    const plays = async (scope: string) => {
      await page.goto('about:blank');
      await ready(page, `/two-tables.html?auto=0&scope=${scope}`);
      return page.evaluate(async () => {
        const { left, right } = (window as any).declarative;
        await left.table.whenSettled(); await right.table.whenSettled();
        for (let step = 0; step < 4; step++) {
          left.next(); right.next();
          await new Promise(resolve => setTimeout(resolve, 50));
          await left.table.whenSettled(); await right.table.whenSettled();
        }
        return left.table.stats.animationsStarted + right.table.stats.animationsStarted;
      });
    };
    const scoped = await plays('table');
    const unscoped = await plays('document');
    expect(unscoped).toBeGreaterThan(scoped);
  });

  test('the same component ids on both tables: scoped animates, unscoped does not animate at all', async ({ page }) => {
    const animated = async (scope: string) => {
      await page.goto('about:blank');
      await ready(page, `/two-tables.html?auto=0&same=1&scope=${scope}`);
      return page.evaluate(async ([index]) => {
        const probe = (window as any).probe;
        const { left, right } = (window as any).declarative;
        await probe.stepTo(index as number, 'left');
        await probe.stepTo(index as number, 'right');
        return [left.table.stats.animatedCycles, right.table.stats.animatedCycles];
      }, [want.capture] as const);
    };
    const scoped = await animated('table');
    expect(scoped[0]).toBeGreaterThan(0);
    expect(scoped[1]).toBeGreaterThan(0);
    expect(await animated('document')).toEqual([0, 0]);
  });

  test('propose-move comes from the table that was clicked', async ({ page }) => {
    await ready(page, '/two-tables.html?auto=0');
    await page.evaluate(() => {
      const seen: string[] = [];
      (window as any).origins = seen;
      for (const id of ['left', 'right']) {
        document.getElementById(id)!.addEventListener('propose-move', () => seen.push(id));
      }
    });
    await page.evaluate(async () => {
      await (window as any).probe.stepTo(1, 'right');
    });
    await page.locator('#right bg-stack').first()
      .locator('boardgame-card[boardgame-component]').nth(5).click();
    expect(await page.evaluate(() => (window as any).origins)).toEqual(['right']);
  });
});

test.describe('blackjack', () => {
  test('per-seat hands, face-down cards and mapped faces from the same vocabulary', async ({ page }) => {
    const errors = await ready(page, '/blackjack.html?auto=0&viewer=1');
    const board = await page.evaluate(async () => {
      const d = (window as any).declarative;
      // Version 9: both seats dealt, seat 0 has revealed, seat 1 has not.
      while (d.position < 10) { d.next(); await new Promise(r => setTimeout(r, 0)); await d.table.whenSettled(); }
      const read = (path: string) => {
        const stack = [...d.table.querySelectorAll('bg-stack')].find((s: any) => s.path === path) as any;
        return [...stack.shadowRoot.querySelectorAll('boardgame-component-stack > boardgame-card')]
          .filter((card: any) => !card.spacer)
          .map((card: any) => card.faceUp ? `${card.suit}${card.rank}` : '##');
      };
      const seats = [...d.table.querySelectorAll('bg-seat')].map((seat: any) => ({
        active: seat.hasAttribute('active'), viewer: seat.hasAttribute('viewer'),
      }));
      const actions = [...d.table.querySelectorAll('bg-action')].map((action: any) =>
        action.shadowRoot.querySelector('button').disabled);
      return {
        draw: read('Game.DrawStack'), p0: read('Players.0.VisibleHand'),
        p1Hidden: read('Players.1.HiddenHand'), p1: read('Players.1.VisibleHand'),
        seats, actions, stats: { ...d.table.stats },
      };
    });
    // Seat 1 is the viewer, so its own hidden card is face up to it.
    expect(board.draw).toHaveLength(48);
    expect(board.draw.every((face: string) => face === '##')).toBe(true);
    expect(board.p0).toHaveLength(2);
    expect(board.p0.every((face: string) => /^[♠♥♣♦](A|J|Q|K|[0-9]+)$/.test(face))).toBe(true);
    expect(board.p1Hidden).toHaveLength(1);
    expect(board.p1Hidden[0]).not.toBe('##');
    expect(board.seats[1].viewer).toBe(true);
    expect(board.stats.animatedCycles).toBeGreaterThan(3);
    expect(errors).toEqual([]);
  });

  test('another viewer sees that hidden card face down, and Hit proposes a move', async ({ page }) => {
    await ready(page, '/blackjack.html?auto=0&viewer=0');
    const hidden = await page.evaluate(async () => {
      const d = (window as any).declarative;
      while (d.position < 10) { d.next(); await new Promise(r => setTimeout(r, 0)); await d.table.whenSettled(); }
      const stack = [...d.table.querySelectorAll('bg-stack')]
        .find((s: any) => s.path === 'Players.1.HiddenHand') as any;
      return [...stack.shadowRoot.querySelectorAll('boardgame-component-stack > boardgame-card')]
        .filter((card: any) => !card.spacer)
        .map((card: any) => ({ faceUp: card.faceUp, suit: card.suit, rank: card.rank, text: card.textContent.trim() }));
    });
    expect(hidden).toEqual([{ faceUp: false, suit: '', rank: '', text: '' }]);
    await page.locator('bg-action[move="Current Player Hit"] button').click();
    expect(await page.evaluate(() => (window as any).proposals)).toEqual([
      { move: 'Current Player Hit', args: {}, version: 9 },
    ]);
  });
});
