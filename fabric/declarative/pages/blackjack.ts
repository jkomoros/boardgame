/** Blackjack's whole script: parse the replay, assign one frame per state. */
import '../src/index.js';
import type { BgStack } from '../src/index.js';
import { framesFor, loadReplay, logProposals } from './replay.js';

const params = new URLSearchParams(location.search);
const viewer = Number(params.get('viewer') ?? 0);
const replay = await loadReplay('blackjack');
const frames = framesFor(replay, viewer);

const table = document.querySelector('bg-table')!;
table.chest = replay.chest;
table.viewer = viewer;
// The face description is one JSON value shared by every card stack. It is
// written once in the markup and copied to each stack's `face` attribute.
const face = document.querySelector<HTMLTemplateElement>('#cards')!.content.textContent!;
for (const stack of document.querySelectorAll<BgStack>('bg-stack.cards')) stack.setAttribute('face', face);
logProposals(document, document.querySelector<HTMLElement>('#log')!);

// `start` jumps straight to a version, for a static look at one state.
let index = Math.min(Number(params.get('start') ?? 0), frames.length - 1);
const next = (): void => {
  if (index >= frames.length) return;
  table.frame = frames[index++]!;
  document.querySelector('#position')!.textContent =
    `${index} / ${frames.length}: ${replay.frames[index - 1]!.move || replay.frames[index - 1]!.label}`;
};
document.querySelector('#next')!.addEventListener('click', next);
next();
if (params.get('auto') !== '0') setInterval(next, 700);
Object.assign(window, { declarative: { table, frames, next, get position() { return index; } } });
