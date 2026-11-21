/** Page A's whole script: parse the replay, assign one frame per state. */
import '../src/index.js';
import { framesFor, loadReplay, logProposals } from './replay.js';

const params = new URLSearchParams(location.search);
const viewer = Number(params.get('viewer') ?? 0);
const replay = await loadReplay(params.get('replay') ?? 'memory-a');
const frames = framesFor(replay, viewer);

const table = document.querySelector('bg-table')!;
table.chest = replay.chest;
table.viewer = viewer;
logProposals(document, document.querySelector<HTMLElement>('#log')!);

let index = 0;
const next = (): void => {
  if (index >= frames.length) return;
  table.frame = frames[index++]!;
  document.querySelector('#position')!.textContent =
    `${index} / ${frames.length}: ${replay.frames[index - 1]!.label}`;
};
document.querySelector('#next')!.addEventListener('click', next);
next();
// The table queues, so a plain timer is enough: no settle signal is needed.
if (params.get('auto') !== '0') setInterval(next, 1400);
Object.assign(window, { declarative: { table, frames, next, get position() { return index; } } });
