import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('rectangular game boards align layers, preserve label order, and reject a short layer', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const views = await import('/src/components/component-view.ts');
      await import('/src/components/boardgame-game-board.ts');
      document.body.innerHTML = '';

      const component = (id: string, layer: string) => ({
        ID: id,
        Index: 0,
        Deck: layer,
        GameName: 'layer-test',
        Values: { Layer: layer },
      });
      const stack = (layer: string, length: number, occupiedIndex = -1) => {
        const components = Array.from({ length }, (_, index) =>
          index === occupiedIndex ? component(`${layer}-${index}`, layer) : null);
        return {
          Deck: layer,
          Indexes: components.map(item => item ? item.Index : -1),
          IDs: components.map(item => item?.ID ?? ''),
          IDsLastSeen: {},
          ShuffleCount: 0,
          Size: length,
          GameName: 'layer-test',
          Components: components,
        };
      };

      const board = document.createElement('boardgame-game-board') as any;
      board.rows = 2;
      board.cols = 3;
      board.boardLabel = 'Two by three layered board';
      board.stacks = [stack('words', 6, 4), stack('keys', 6, 4)];
      board.componentViews = [
        views.tokenView({ properties: () => ({ type: 'disc', color: 'red' }) }),
        views.tokenView({ properties: () => ({ type: 'disc', color: 'blue' }) }),
      ];
      board.labelFor = ({ index, occupants }: { index: number; occupants: any[] }) =>
        `${index}: ${occupants.map(item => item?.Values.Layer ?? 'empty').join(' then ')}`;
      document.body.appendChild(board);
      await board.updateComplete;
      const renderedLayers = [...board.shadowRoot.querySelectorAll('boardgame-component-stack')] as any[];
      await Promise.all(renderedLayers.map(layer => layer.updateComplete));
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));

      const centers = renderedLayers.map(layer => {
        const rect = layer.Components[4].getBoundingClientRect();
        return { x: rect.x + rect.width / 2, y: rect.y + rect.height / 2 };
      });
      const boardArea = board.shadowRoot.querySelector('.board-area').getBoundingClientRect();
      const label = board.shadowRoot.querySelector('[data-index="4"]').getAttribute('aria-label');

      const short = document.createElement('boardgame-game-board') as any;
      short.rows = 2;
      short.cols = 3;
      short.boardLabel = 'Invalid short layer';
      short.stacks = [stack('words', 6), stack('keys', 5)];
      let shortLayerError = '';
      try {
        short._validateConfiguration();
      } catch (error) {
        shortLayerError = error instanceof Error ? error.message : String(error);
      }

      board.remove();
      return {
        layers: renderedLayers.map(layer => ({ cols: layer.boardCols, rows: layer.boardRows })),
        label,
        centers,
        ratio: boardArea.width / boardArea.height,
        shortLayerError,
      };
    });

    expect(result.layers).toEqual([{ cols: 3, rows: 2 }, { cols: 3, rows: 2 }]);
    expect(result.label).toBe('4: words then keys');
    expect(result.ratio).toBeCloseTo(3 / 2, 2);
    expect(Math.abs(result.centers[0].x - result.centers[1].x)).toBeLessThan(1);
    expect(Math.abs(result.centers[0].y - result.centers[1].y)).toBeLessThan(1);
    expect(result.shortLayerError).toContain('expected 6 stack components but received 5');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
