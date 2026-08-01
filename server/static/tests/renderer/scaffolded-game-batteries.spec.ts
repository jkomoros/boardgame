import { expect, test } from '@playwright/test';
import { mkdirSync, rmSync, writeFileSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

/**
 * THE ACCEPTANCE TEST FOR THE SCAFFOLD.
 *
 * Two client batteries exist so renderers stop rebuilding them: the shared
 * layout/state vocabulary (`src/styles/renderer-styles.ts`, wired onto the
 * renderer base classes) and `<boardgame-stat>` (which reads a count and a
 * capacity off the stack itself). The generator that writes every new game's
 * renderer received neither, so the default path did not get the default —
 * which is the whole objection.
 *
 * `boardgame-util/lib/stub/client_batteries_test.go` pins the generated TEXT.
 * This pins the generated BEHAVIOUR: it takes the committed stub goldens
 * VERBATIM off disk, drops them into a game-src directory laid out exactly as a
 * scaffolded game is (so their own `'../../src/client.js'` import resolves with
 * no rewriting at all), mounts them in a real browser, and asks the two
 * questions the batteries exist to answer:
 *
 *   1. Does `.horizontal` / `.active` mean anything inside a scaffolded game's
 *      shadow root, without the game importing a thing?
 *   2. Does the player summary show the COUNT of a sized stack, rather than
 *      `Indexes.length`, which on a sized stack is the capacity?
 *
 * The only file written that is not a golden is `_game_renderer.ts`, which a
 * real scaffolded game gets from `boardgame-util codegen` rather than from the
 * stub. The shim reproduces exactly the two things the goldens use from it: a
 * base class per surface, and a registration decorator that defines the tag.
 */

const repoRoot = fileURLToPath(new URL('../../../..', import.meta.url));
const goldenDir = resolve(repoRoot, 'boardgame-util/lib/stub/testdata');
// game-src is already gitignored, and its depth is the one a scaffolded game
// has: `../../src/client.js` from inside it is `server/static/src/client.ts`.
const fixtureDir = resolve(repoRoot, 'server/static/game-src/scaffoldacceptance');
// `boardgame-util serve` runs Vite from an assembled temp directory, so the
// fixture is reached through Vite's filesystem escape hatch rather than a root-
// relative path. The files themselves are untouched: their own
// `'../../src/client.js'` import resolves on disk from where they are written.
const fixtureURL = `/@fs${fixtureDir}`;

/** Six slots, three filled. `Indexes.length` is 6 — the trap the stat kills. */
const SIZED_HAND = {
  Deck: 'examplecards',
  Indexes: [4, -1, 9, -1, 2, -1],
  IDs: ['a', '', 'b', '', 'c', ''],
  IDsLastSeen: {},
  ShuffleCount: 0,
  Size: 6,
  GameName: 'checkers',
  Components: [{ Index: 4 }, null, { Index: 9 }, null, { Index: 2 }, null],
};

const GAME_RENDERER_SHIM = `
import {
  BoardgameBaseGameRenderer,
  BoardgameBasePlayerInfoRenderer,
} from '../../src/client.js';

export abstract class GameRenderer extends BoardgameBaseGameRenderer<any, any, any, any> {}
export abstract class PlayerInfoRenderer extends BoardgameBasePlayerInfoRenderer<any, any> {}

export function registerGameRenderer(constructor: any): void {
  customElements.define('boardgame-render-game-checkers', constructor);
}

export function registerPlayerInfoRenderer(constructor: any): void {
  customElements.define('boardgame-render-player-info-checkers', constructor);
}
`;

// One worker for the whole file: the fixture directory is shared state, and a
// parallel worker's afterAll would delete it out from under the other test.
test.describe.configure({ mode: 'serial' });

test.beforeAll(() => {
  rmSync(fixtureDir, { recursive: true, force: true });
  mkdirSync(fixtureDir, { recursive: true });
  writeFileSync(resolve(fixtureDir, '_game_renderer.ts'), GAME_RENDERER_SHIM);
  // The DEFAULT scaffold (no tutorial flag) for the game surface: this is the
  // renderer a person gets when they type `boardgame-util stub` and answer
  // nothing special, and it is the one that used to declare `static styles`
  // outright and so throw the vocabulary away.
  writeFileSync(
    resolve(fixtureDir, 'boardgame-render-game-checkers.ts'),
    readFileSync(resolve(goldenDir, 'default/checkers/client/boardgame-render-game-checkers.ts')),
  );
  // The TUTORIAL scaffold for the player summary, because that is the variant
  // that renders a stack at all.
  writeFileSync(
    resolve(fixtureDir, 'boardgame-render-player-info-checkers.ts'),
    readFileSync(resolve(goldenDir, 'tutorial/checkers/client/boardgame-render-player-info-checkers.ts')),
  );
});

test.afterAll(() => {
  rmSync(fixtureDir, { recursive: true, force: true });
});

test('a scaffolded game renderer receives the shared style vocabulary', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);

  const computed = await page.evaluate(async (base: string) => {
    await import(`${base}/boardgame-render-game-checkers.ts`);
    await customElements.whenDefined('boardgame-render-game-checkers');
    const element = document.createElement('boardgame-render-game-checkers');
    document.body.appendChild(element);
    await (element as unknown as { updateComplete: Promise<unknown> }).updateComplete;
    const root = element.shadowRoot;
    if (!root) throw new Error('scaffolded renderer has no shadow root');
    // Probe the vocabulary where it has to work: inside this game's own shadow
    // root, on markup the game wrote, with the game importing nothing.
    const probe = document.createElement('div');
    probe.className = 'horizontal center active';
    root.appendChild(probe);
    const style = getComputedStyle(probe);
    return {
      display: style.display,
      flexDirection: style.flexDirection,
      alignItems: style.alignItems,
      outlineStyle: style.outlineStyle,
      outlineWidth: style.outlineWidth,
    };
  }, fixtureURL);

  expect(computed.display, '.horizontal did not reach a scaffolded shadow root').toBe('flex');
  expect(computed.flexDirection).toBe('row');
  expect(computed.alignItems, '.center did not reach a scaffolded shadow root').toBe('center');
  expect(computed.outlineStyle, '.active did not reach a scaffolded shadow root').toBe('solid');
  expect(computed.outlineWidth).toBe('2px');

  diagnostics.assertEmpty();
  diagnostics.stop();
});

test('a scaffolded player summary shows the count of a sized stack, not its capacity', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);

  const result = await page.evaluate(async ({ base, hand }: { base: string; hand: unknown }) => {
    await import(`${base}/boardgame-render-player-info-checkers.ts`);
    await customElements.whenDefined('boardgame-render-player-info-checkers');
    const element = document.createElement('boardgame-render-player-info-checkers');
    (element as unknown as { state: unknown }).state = { Players: [{ Hand: hand }] };
    (element as unknown as { playerIndex: number }).playerIndex = 0;
    document.body.appendChild(element);
    await (element as unknown as { updateComplete: Promise<unknown> }).updateComplete;
    const root = element.shadowRoot;
    if (!root) throw new Error('scaffolded player-info renderer has no shadow root');
    const stat = root.querySelector('boardgame-stat');
    if (!stat) throw new Error('scaffolded player-info renderer does not use <boardgame-stat>');
    await (stat as unknown as { updateComplete: Promise<unknown> }).updateComplete;
    const statRoot = stat.shadowRoot;
    if (!statRoot) throw new Error('boardgame-stat has no shadow root');
    const statusText = statRoot.querySelector('boardgame-status-text');
    await (statusText as unknown as { updateComplete: Promise<unknown> }).updateComplete;
    return {
      indexesLength: (hand as { Indexes: readonly number[] }).Indexes.length,
      displayValue: (stat as unknown as { displayValue: unknown }).displayValue,
      displayCapacity: (stat as unknown as { displayCapacity: unknown }).displayCapacity,
      label: (statRoot.querySelector('[part="label"]')?.textContent ?? '').trim(),
      valueText: (statusText?.shadowRoot?.textContent ?? '').trim(),
      capacityText: (statRoot.querySelector('[part="capacity"]')?.textContent ?? '').trim(),
    };
  }, { base: fixtureURL, hand: SIZED_HAND });

  // The bug in one line: the number the old scaffold printed was this one.
  expect(result.indexesLength).toBe(6);
  expect(result.displayValue, 'the scaffold printed the capacity as if it were the count').toBe(3);
  expect(result.valueText).toContain('3');
  expect(result.displayCapacity, 'the stat did not read the capacity off the stack').toBe(6);
  expect(result.capacityText).toBe('/6');
  expect(result.label).toBe('Cards');

  diagnostics.assertEmpty();
  diagnostics.stop();
});
