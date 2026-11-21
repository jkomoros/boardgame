/** Page C: two tables, two replays, two clocks, one page. */
import '../src/index.js';
import type { BgTable, Frame } from '../src/index.js';
import { framesFor, loadReplay, logProposals } from './replay.js';

const params = new URLSearchParams(location.search);
const scope = params.get('scope') ?? 'table';
const same = params.get('same') === '1';
logProposals(document, document.querySelector<HTMLElement>('#log')!);
const board = document.querySelector<HTMLTemplateElement>('#board')!;

async function drive(id: string, replayName: string, viewer: number) {
  const replay = await loadReplay(replayName);
  const frames: Frame[] = framesFor(replay, viewer);
  const table = document.querySelector<BgTable>(`#${id}`)!;
  table.setAttribute('scope', scope);
  table.chest = replay.chest;
  table.viewer = viewer;
  table.append(board.content.cloneNode(true));
  let index = 0;
  const next = (): void => {
    if (index < frames.length) table.frame = frames[index++]!;
  };
  next();
  return { table, frames, next, get position() { return index; } };
}

const left = await drive('left', 'memory-a', 0);
const right = await drive('right', same ? 'memory-a' : 'memory-b', 0);
document.querySelector('#next-left')!.addEventListener('click', left.next);
document.querySelector('#next-right')!.addEventListener('click', right.next);
if (params.get('auto') !== '0') {
  setInterval(left.next, 1300);
  setTimeout(() => setInterval(right.next, 1900), 450);
}
Object.assign(window, { declarative: { left, right } });
