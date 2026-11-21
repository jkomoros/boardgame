/**
 * Page B: the same Memory board, driven the way a pattern renderer would
 * drive it. The script holds no element across states. It describes the board
 * as plain data and a small reconciler DESTROYS AND RECREATES elements.
 */
import '../src/index.js';
import type { Frame } from '../src/index.js';
import { framesFor, loadReplay, logProposals } from './replay.js';

const params = new URLSearchParams(location.search);
const mode = (params.get('mode') ?? 'all') as 'all' | 'changed' | 'table';
const measure = params.get('measure') ?? 'remembered';
const viewer = Number(params.get('viewer') ?? 0);
const replay = await loadReplay(params.get('replay') ?? 'memory-a');
const frames = framesFor(replay, viewer);
const mount = document.querySelector<HTMLElement>('#mount')!;
logProposals(document, document.querySelector<HTMLElement>('#log')!);

/** One element as a renderer sees it: a tag, attributes, JSON properties. */
interface Described {
  readonly host: string;
  readonly tag: string;
  readonly attrs: Readonly<Record<string, string>>;
  readonly props: Readonly<Record<string, unknown>>;
}

/** The board for one frame, as data. Everything here is JSON-serializable. */
function describe(frame: Frame): Described[] {
  const state = frame.view.CurrentState as {
    Game: { Cards: { Indexes: number[] } };
    Players: { WonCards: unknown }[];
  };
  const candidates = (frame.choices?.Sets ?? [])
    .filter(set => set.MoveName === 'Reveal Card' && set.FieldName === 'CardIndex')
    .flatMap(set => set.Candidates.filter(c => c.Available !== false).map(c => c.Value));
  return [
    {
      host: 'grid', tag: 'bg-stack',
      attrs: { name: 'cards', layout: 'grid', messy: '', face: 'Type' },
      props: {
        stack: state.Game.Cards,
        version: frame.view.Version,
        pick: { move: 'Reveal Card', arg: 'CardIndex', slots: 'hidden', candidates },
      },
    },
    {
      host: 'grid', tag: 'bg-callout',
      attrs: { name: 'match', message: 'Match' },
      props: { trigger: state.Game.Cards.Indexes.filter(index => index !== -1).length },
    },
    ...state.Players.map((player, index) => ({
      host: `won${index}`, tag: 'bg-stack',
      attrs: { name: `won${index}`, layout: 'stack', messy: '', face: 'Type' },
      props: { stack: player.WonCards },
    })),
  ];
}

const SHELL = `
  <div class="board" data-host="grid"></div>
  <div class="won">
    <div class="pile"><div class="label">Player 0</div><div data-host="won0"></div></div>
    <div class="pile"><div class="label">Player 1</div><div data-host="won1"></div></div>
  </div>`;

function newTable(): HTMLElement & Record<string, unknown> {
  const table = document.createElement('bg-table') as unknown as HTMLElement & Record<string, unknown>;
  table.setAttribute('name', 'memory');
  table.setAttribute('game-name', 'memory');
  table.setAttribute('animation-length', '400');
  table.setAttribute('measure', measure);
  table.chest = replay.chest;
  table.viewer = viewer;
  return table;
}

function create(item: Described): HTMLElement {
  const element = document.createElement(item.tag);
  for (const [name, value] of Object.entries(item.attrs)) element.setAttribute(name, value);
  Object.assign(element, item.props);
  return element;
}

let table = newTable();
table.innerHTML = SHELL;
mount.append(table);
const live = new Map<string, { element: HTMLElement; json: string }>();
const counts = { created: 0, kept: 0 };

function render(frame: Frame): void {
  if (mode === 'table') {
    // Worst case: the animation root itself is thrown away.
    table.remove();
    table = newTable();
    table.innerHTML = `
      <div class="board">
        <bg-stack name="cards" path="Game.Cards" layout="grid" messy face="Type"
                  pick='{"move":"Reveal Card","arg":"CardIndex","slots":"hidden"}'></bg-stack>
        <bg-callout name="match" message="Match" path="Game.Cards.$count"></bg-callout>
      </div>
      <div class="won">
        <div class="pile"><div class="label">Player 0</div>
          <bg-stack name="won0" path="Players.0.WonCards" layout="stack" messy face="Type"></bg-stack></div>
        <div class="pile"><div class="label">Player 1</div>
          <bg-stack name="won1" path="Players.1.WonCards" layout="stack" messy face="Type"></bg-stack></div>
      </div>`;
    mount.append(table);
    table.frame = frame;
    counts.created++;
    return;
  }
  for (const item of describe(frame)) {
    const key = `${item.host}/${item.attrs.name}`;
    const json = JSON.stringify([item.attrs, item.props]);
    const existing = live.get(key);
    if (existing && mode === 'changed' && existing.json === json) {
      counts.kept++;
      continue;
    }
    // Destroy, then create. Never update in place.
    existing?.element.remove();
    const element = create(item);
    table.querySelector(`[data-host="${item.host}"]`)!.append(element);
    live.set(key, { element, json });
    counts.created++;
  }
}

let index = 0;
const next = (): void => {
  if (index >= frames.length) return;
  render(frames[index++]!);
  document.querySelector('#position')!.textContent =
    `${index} / ${frames.length}: ${replay.frames[index - 1]!.label} (mode ${mode}, measure ${measure})`;
};
document.querySelector('#next')!.addEventListener('click', next);
next();
if (params.get('auto') !== '0') setInterval(next, 1400);
Object.assign(window, {
  declarative: {
    get table() { return table; },
    get position() { return index; },
    frames, next, counts,
  },
});
