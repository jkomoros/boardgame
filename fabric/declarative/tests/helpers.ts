import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import type { Page } from '@playwright/test';

/** What a replay says should be on screen, computed without any element code. */
export interface Expected {
  /** Frame index of the first reveal, of the frame where both cards of the captured pair are up, and of the capture. */
  readonly firstReveal: number;
  readonly bothUp: number;
  readonly capture: number;
  /** Slot and face of the first revealed card. */
  readonly revealSlot: number;
  readonly revealFace: string;
  /** The captured pair. */
  readonly pairIds: readonly string[];
  readonly pairFace: string;
  readonly winner: number;
  readonly versions: readonly number[];
  /** A slot that is still face down after the first reveal. */
  readonly hiddenSlot: number;
}

export function expected(replayName: string): Expected {
  const path = fileURLToPath(new URL(`../replays/${replayName}.json`, import.meta.url));
  const replay = JSON.parse(readFileSync(path, 'utf8'));
  const state = (index: number) =>
    replay.frames[index].viewers.find((v: any) => v.viewer === 0).game.CurrentState;
  const deck = replay.chest.Decks.cards;
  const firstReveal = replay.frames.findIndex((f: any) => f.move === 'Reveal Card');
  const capture = replay.frames.findIndex((f: any) => f.move === 'Capture Cards');
  const bothUp = capture - 1;
  const cards = state(firstReveal).Game.Cards;
  const revealSlot = cards.Indexes.findIndex((index: number) => index >= 0);
  const before = state(bothUp);
  const after = state(capture);
  const winner = after.Players.findIndex((player: any, index: number) =>
    player.WonCards.Indexes.filter((i: number) => i !== -1).length
      > before.Players[index].WonCards.Indexes.filter((i: number) => i !== -1).length);
  const pairIds = after.Players[winner].WonCards.IDs.filter((id: string) => id);
  const pairIndex = after.Players[winner].WonCards.Indexes.find((i: number) => i >= 0);
  return {
    firstReveal, bothUp, capture, revealSlot,
    revealFace: deck[cards.Indexes[revealSlot]].Values.Type,
    pairIds, pairFace: deck[pairIndex].Values.Type, winner,
    versions: replay.frames.map((f: any) => f.version),
    hiddenSlot: cards.Indexes.findIndex((index: number) => index === -2),
  };
}

/**
 * Page-side helpers. `side` selects a table on the two-table page.
 * Everything looks elements up fresh, by component id, so it keeps working
 * when a page destroys and recreates its elements.
 */
export async function installProbe(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const handle = (side?: string) => {
      const d = (window as any).declarative;
      return side ? d[side] : d;
    };
    const stacks = (side?: string): any[] => [...handle(side).table.querySelectorAll('bg-stack')];
    const stack = (kind: string, side?: string): any => {
      const keys: Record<string, string[]> = {
        grid: ['cards', 'Game.Cards'],
        won0: ['won0', 'Players.0.WonCards'],
        won1: ['won1', 'Players.1.WonCards'],
      };
      return stacks(side).find(s => keys[kind]!.includes(s.key));
    };
    const cardsIn = (s: any): any[] =>
      [...s.shadowRoot.querySelectorAll('boardgame-component-stack > boardgame-card[boardgame-component]')]
        .filter((card: any) => !card.spacer);
    const allCards = (side?: string): any[] => stacks(side).flatMap(cardsIn);
    const describe = (card: any) => ({
      id: card.id, index: card.index, faceUp: card.faceUp, text: card.textContent.trim(),
      disabled: card.disabled, role: card.getAttribute('role'),
    });
    const animations = (side?: string): Animation[] => {
      const found: Animation[] = [];
      for (const s of stacks(side)) {
        found.push(...s.shadowRoot.getAnimations());
        const inner = s.shadowRoot.querySelector('boardgame-component-stack');
        found.push(...inner.shadowRoot.getAnimations());
        for (const card of inner.querySelectorAll('boardgame-card')) {
          found.push(...card.shadowRoot.getAnimations());
        }
        for (const carrier of inner.shadowRoot.querySelectorAll('boardgame-card')) {
          found.push(...carrier.shadowRoot.getAnimations());
        }
      }
      return found.filter(animation => animation.playState === 'running');
    };
    const center = (id: string, side?: string) => {
      const card = allCards(side).find(c => c.id === id);
      if (!card) return null;
      const rect = card.getBoundingClientRect();
      return { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
    };
    /** Advance one state and sample, every frame until settled, where `ids` are and what is animating. */
    const advanceAndTrack = async (ids: string[], side?: string) => {
      const h = handle(side);
      const start = ids.map(id => center(id, side));
      const samples: any[] = [];
      let running = 0;
      let transformAnimations = 0;
      let done = false;
      const t0 = performance.now();
      h.next();
      // Sample AFTER the frame's other callbacks have run, i.e. what was
      // painted: the animator starts playback from its own rAF callback.
      const channel = new MessageChannel();
      channel.port1.onmessage = () => sample();
      const tick = () => {
        if (done) return;
        channel.port2.postMessage(0);
        requestAnimationFrame(tick);
      };
      const sample = () => {
        if (done) return;
        samples.push(ids.map(id => center(id, side)));
        const live = animations(side);
        running = Math.max(running, live.length);
        transformAnimations = Math.max(transformAnimations, live.filter(animation => {
          const effect = animation.effect as KeyframeEffect;
          return effect.getKeyframes().some(frame => 'transform' in frame)
            && Number(effect.getComputedTiming().duration) > 0;
        }).length);
      };
      requestAnimationFrame(tick);
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      await h.table.whenSettled();
      done = true;
      const end = ids.map(id => center(id, side));
      return { start, end, samples, running, transformAnimations, ms: performance.now() - t0 };
    };
    const stepTo = async (index: number, side?: string) => {
      const h = handle(side);
      await h.table.whenSettled();
      // `position` counts frames delivered; the page delivers frame 0 itself.
      while (h.position < index + 1) {
        h.next();
        await new Promise(resolve => setTimeout(resolve, 0));
        await h.table.whenSettled();
      }
    };
    (window as any).probe = {
      handle, stack, cardsIn, allCards, describe, animations, center, advanceAndTrack, stepTo,
      grid: (side?: string) => cardsIn(stack('grid', side)).map(describe),
      won: (player: number, side?: string) => cardsIn(stack(`won${player}`, side)).map(describe),
    };
    // Record every event that reaches the document, to prove what does not leak.
    const leaked: string[] = [];
    (window as any).leaked = leaked;
    for (const type of ['will-animate', 'animation-done', 'stack-slot-geometry-changed',
      'component-tapped', 'motion-cycle-release']) {
      document.addEventListener(type, () => leaked.push(type));
    }
  });
}

export async function ready(page: Page, url: string): Promise<string[]> {
  const errors: string[] = [];
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()); });
  page.on('pageerror', error => errors.push(String(error)));
  await installProbe(page);
  await page.goto(url);
  await page.waitForFunction(() => (window as any).declarative);
  return errors;
}

/** How a tracked component travelled, as fractions of the straight line from start to end. */
export function progress(
  track: { start: any[]; end: any[]; samples: any[][] },
  which: number,
): { travel: number; fractions: number[] } {
  const start = track.start[which];
  const end = track.end[which];
  const travel = Math.hypot(end.x - start.x, end.y - start.y);
  const fractions = track.samples.map(sample => sample[which]).filter(Boolean).map(point =>
    ((point.x - start.x) * (end.x - start.x) + (point.y - start.y) * (end.y - start.y))
      / (travel * travel));
  return { travel, fractions };
}
