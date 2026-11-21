import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

// The renderer host owns a mount path for a game's player-info renderer that
// needs neither the webapp's roster nor its store (fabric/PROPOSAL.md, S7).
test('a host mounts a player-info renderer without the roster or the store', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { mountPlayerInfo } = await import('/src/components/boardgame-render-player-info.ts');
      const { BoardgameBasePlayerInfoRenderer } = await import('/src/components/boardgame-base-player-info-renderer.ts');
      const { html } = await import('/src/client.ts');
      const requests: unknown[] = [];
      const host = {
        loadPlayerInfoModule: async (request: unknown) => {
          requests.push(request);
          class MountedPlayerInfo extends (BoardgameBasePlayerInfoRenderer as unknown as new () => HTMLElement & {
            playerState: { Score: number } | null;
          }) {
            render() {
              return html`<span class="score">Score ${this.playerState?.Score ?? '?'}</span>`;
            }
          }
          customElements.define('boardgame-render-player-info-mountedgame', MountedPlayerInfo);
        },
      };
      const chips: unknown[] = [];
      document.body.addEventListener('player-chip-presentation-changed', event => {
        chips.push((event as CustomEvent).detail);
      });
      const mounts = [0, 1].map(playerIndex => mountPlayerInfo(document.body, {
        host,
        gameName: 'mountedgame',
        playerIndex,
        state: { Players: [{ Score: 3 }, { Score: 8 }] },
      }));
      const text = async (mount: HTMLElement & { updateComplete: Promise<unknown>; renderRoot: ParentNode }) => {
        for (let i = 0; i < 50; i++) {
          await mount.updateComplete;
          const renderer = mount.renderRoot.querySelector('boardgame-render-player-info-mountedgame') as
            (HTMLElement & { updateComplete: Promise<unknown>; renderRoot: ParentNode }) | null;
          if (renderer) {
            await renderer.updateComplete;
            return renderer.renderRoot.querySelector('.score')?.textContent ?? '';
          }
          await new Promise(resolve => setTimeout(resolve, 10));
        }
        return 'never mounted';
      };
      const initial = [await text(mounts[0]), await text(mounts[1])];
      mounts[1].state = { Players: [{ Score: 3 }, { Score: 9 }] };
      const updated = await text(mounts[1]);
      return {
        initial,
        updated,
        requests,
        storeLoaded: (globalThis as unknown as { __bgReduxStore?: unknown }).__bgReduxStore !== undefined,
        rosterDefined: customElements.get('boardgame-player-roster') !== undefined,
        chipEvents: chips.length > 0,
      };
    });
    expect(result).toEqual({
      initial: ['Score 3', 'Score 8'],
      updated: 'Score 9',
      // Two seats, one request: the host is asked once per game.
      requests: [{ gameName: 'mountedgame', attempt: 0 }],
      storeLoaded: false,
      rosterDefined: false,
      chipEvents: true,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a host-mounted player-info renderer shows a load failure in place', async ({ page }) => {
  await prepareRendererFixturePage(page);
  const message = await page.evaluate(async () => {
    const { mountPlayerInfo } = await import('/src/components/boardgame-render-player-info.ts');
    const mount = mountPlayerInfo(document.body, {
      host: { loadPlayerInfoModule: async () => { /* registers nothing */ } },
      gameName: 'absentgame',
      playerIndex: 0,
      state: null,
    }) as HTMLElement & { updateComplete: Promise<unknown>; renderRoot: ParentNode };
    for (let i = 0; i < 50; i++) {
      await mount.updateComplete;
      const alert = mount.renderRoot.querySelector('[role="alert"]');
      if (alert) return alert.textContent ?? '';
      await new Promise(resolve => setTimeout(resolve, 10));
    }
    return '';
  });
  expect(message).toBe(
    'Failed to load player renderer for absentgame: Player renderer module loaded but did not ' +
    'register <boardgame-render-player-info-absentgame>; use @registerPlayerInfoRenderer',
  );
});
