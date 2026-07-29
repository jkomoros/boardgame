import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

/**
 * `<boardgame-stat>`, exercised in a browser.
 *
 * The derivation rules live in `src/components/stat-values.test.ts`, which runs
 * without a DOM. What can only be checked here is the part that matters to an
 * author: that the element COMPOSES `boardgame-status-text` rather than
 * reimplementing it, that its escape hatches actually reach through the shadow
 * boundary, and that it refuses a capacity number loudly.
 */

/** Six slots, three filled -- the `Food 3/6` shape, in stack.go's wire format. */
const SIZED_STACK = {
  Deck: 'food',
  Indexes: [4, -1, 9, -1, 2, -1],
  IDs: ['a', '', 'b', '', 'c', ''],
  IDsLastSeen: {},
  ShuffleCount: 0,
  Size: 6,
  GameName: 'g',
  Components: [{ Index: 4 }, null, { Index: 9 }, null, { Index: 2 }, null],
};

/** A growable stack with no cap: MaxSize is omitempty'd off the wire. */
const GROWABLE_STACK = {
  Deck: 'cards',
  Indexes: [3, 7],
  IDs: ['a', 'b'],
  IDsLastSeen: {},
  ShuffleCount: 0,
  GameName: 'g',
  Components: [{ Index: 3 }, { Index: 7 }],
};

/**
 * The text a person actually reads: the FLAT tree, so it descends into
 * `boardgame-status-text`'s own shadow root (where the value lives, because the
 * stat composes it rather than reimplementing it) and skips a slot's fallback
 * content whenever the slot has assigned nodes. `shadowRoot.textContent` sees
 * neither of those and would have made every assertion below a lie.
 *
 * Whitespace-only nodes are DROPPED rather than collapsed to a space, because
 * the stat contains no text spaces at all: every gap a reader sees is flex
 * `gap`. The template's own line breaks and the ones inside status-text's
 * shadow root collapse away in layout (measured: the value and the capacity
 * render exactly 0px apart), so joining them with a space would be inventing
 * separation that is not on screen.
 */
declare const flatText: (node: Node) => string;

async function setup(page: import('@playwright/test').Page) {
  const diagnostics = await prepareRendererFixturePage(page);
  await page.evaluate(async () => {
    await import('/src/client.ts');
    await customElements.whenDefined('boardgame-stat');
    const walk = (node: Node): string => {
      if (node.nodeType === Node.TEXT_NODE) {
        const text = node.textContent ?? '';
        return text.trim() === '' ? '' : text;
      }
      if (node.nodeType !== Node.ELEMENT_NODE) return '';
      // The transient fading-text callout status-text plays on a change is
      // aria-hidden decoration, not the stat's text.
      if ((node as Element).getAttribute?.('aria-hidden') === 'true') return '';
      if (node instanceof HTMLSlotElement) {
        const assigned = node.assignedNodes();
        return (assigned.length ? assigned : [...node.childNodes]).map(walk).join('');
      }
      const element = node as Element;
      const kids = element.shadowRoot ? element.shadowRoot.childNodes : element.childNodes;
      return [...kids].map(walk).join('');
    };
    (globalThis as unknown as { flatText: (node: Node) => string }).flatText =
      node => walk(node);
  });
  return diagnostics;
}

test('a stat is a labelled boardgame-status-text, not a second animating number', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const stat = document.createElement('boardgame-stat');
      stat.label = 'Score';
      stat.icon = '🎲';
      stat.value = 21;
      document.body.append(stat);
      await stat.updateComplete;
      const root = stat.shadowRoot!;
      const status = root.querySelector('boardgame-status-text');
      return {
        // Composition, not duplication: the value IS a status-text.
        composesStatusText: !!status,
        statusValue: (status as { value?: unknown } | null)?.value ?? null,
        // The animating-number policy passes straight through.
        autoMessage: (status as { autoMessage?: unknown } | null)?.autoMessage ?? null,
        text: flatText(stat),
        parts: [...root.querySelectorAll('[part]')].map(node => node.getAttribute('part')),
      };
    });

    expect(result.composesStatusText).toBe(true);
    expect(result.statusValue).toBe(21);
    expect(result.autoMessage).toBe('diff-up');
    expect(result.text).toBe('🎲Score21');
    // Every piece is addressable from outside the shadow root.
    expect(result.parts).toEqual(['icon', 'label', 'value']);

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('capacity comes from the stack itself, and the count is the filled slots', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async ([sized, growable]) => {
      const make = async (stack: unknown) => {
        const stat = document.createElement('boardgame-stat');
        stat.label = 'Food';
        stat.stack = stack as never;
        document.body.append(stat);
        await stat.updateComplete;
        return {
          text: flatText(stat),
          capacityPart: !!stat.shadowRoot!.querySelector('[part="capacity"]'),
          value: stat.displayValue,
          capacity: stat.displayCapacity ?? null,
          // "3/6" must read as one token, not as two words with a gap.
          capacityGapPx: (() => {
            const status = stat.shadowRoot!.querySelector('boardgame-status-text');
            const cap = stat.shadowRoot!.querySelector('[part="capacity"]');
            if (!status || !cap) return null;
            return Math.round(cap.getBoundingClientRect().left - status.getBoundingClientRect().right);
          })(),
        };
      };
      // A computed value over the same stack still takes its capacity from it.
      const computed = document.createElement('boardgame-stat');
      computed.label = 'Food';
      computed.stack = sized as never;
      computed.value = 99;
      document.body.append(computed);
      await computed.updateComplete;
      return {
        sized: await make(sized),
        growable: await make(growable),
        computed: flatText(computed),
      };
    }, [SIZED_STACK, GROWABLE_STACK] as const);

    // Six slots, three filled. Indexes.length is 6 -- the trap several
    // renderers fell into by using it as a count.
    expect(result.sized.value).toBe(3);
    expect(result.sized.capacity).toBe(6);
    expect(result.sized.text).toBe('Food3/6');
    expect(result.sized.capacityGapPx).toBe(0);

    // An uncapped growable stack renders a bare value, never "2/0".
    expect(result.growable.value).toBe(2);
    expect(result.growable.capacity).toBe(null);
    expect(result.growable.capacityPart).toBe(false);
    expect(result.growable.text).toBe('Food2');

    expect(result.computed).toBe('Food99/6');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a capacity number where the stack goes is refused, by name', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const message = await page.evaluate(async () => {
      const stat = document.createElement('boardgame-stat');
      stat.label = 'Food';
      (stat as unknown as { stack: unknown }).stack = 6;
      document.body.append(stat);
      try {
        await stat.updateComplete;
      } catch (error) {
        return (error as Error).message;
      }
      // Lit reports a render throw through updateComplete; if it did not, the
      // element must at least not have rendered a capacity of 6.
      return `NO THROW: ${stat.shadowRoot?.textContent?.trim()}`;
    });

    expect(message).toContain('takes the stack itself, not a capacity number');

    diagnostics.stop();
  } finally {
    diagnostics.stop();
  }
});

test('the icon and label slots replace the defaults, and a part restyles the value', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const host = document.createElement('div');
      const sheet = document.createElement('style');
      sheet.textContent = 'boardgame-stat::part(value) { color: rgb(1, 2, 3); }';
      document.head.append(sheet);
      host.innerHTML = '<boardgame-stat label="ignored" icon="ignored">'
        + '<span slot="icon">★</span><em slot="label">Supply</em></boardgame-stat>';
      document.body.append(host);
      const stat = host.querySelector('boardgame-stat')!;
      stat.value = 4;
      await stat.updateComplete;
      const root = stat.shadowRoot!;
      return {
        // Slotted content wins; the attribute value is slot fallback only, and
        // the flat-tree read is what proves the fallback is not what shows.
        text: flatText(stat),
        valueColor: getComputedStyle(root.querySelector('[part="value"]')!).color,
      };
    });

    expect(result.text).toBe('★Supply4');
    // ::part reaches through the shadow boundary from the game's own stylesheet.
    expect(result.valueColor).toBe('rgb(1, 2, 3)');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('memory adopted the stack form rather than counting Indexes by hand', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const rendered = await page.evaluate(async () => {
      await import('/game-src/memory/boardgame-render-player-info-memory.ts');
      await customElements.whenDefined('boardgame-render-player-info-memory');
      const info = document.createElement('boardgame-render-player-info-memory') as HTMLElement & {
        state: unknown;
        playerIndex: number;
        updateComplete: Promise<unknown>;
      };
      info.state = {
        Players: [{
          WonCards: {
            Deck: 'cards', Indexes: [1, 2, 3, 4], IDs: ['a', 'b', 'c', 'd'],
            IDsLastSeen: {}, ShuffleCount: 0, GameName: 'memory',
            Components: [{ Index: 1 }, { Index: 2 }, { Index: 3 }, { Index: 4 }],
          },
        }],
      };
      info.playerIndex = 0;
      document.body.append(info);
      await info.updateComplete;
      const stat = info.shadowRoot!.querySelector('boardgame-stat')!;
      await stat.updateComplete;
      return {
        text: flatText(stat),
        boundStack: stat.stack !== null,
        boundValue: stat.value,
      };
    });

    expect(rendered.text).toBe('Won Cards4');
    // The stack, not a pre-counted number: the point of the migration.
    expect(rendered.boundStack).toBe(true);
    expect(rendered.boundValue).toBe(undefined);

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
