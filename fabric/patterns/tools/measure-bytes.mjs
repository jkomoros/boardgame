// Plays one fixed game through a table under Node and reports, for the deal
// and for each move, how much the table stores and how much each viewer is
// handed. No runtime is involved: these are JSON sizes of the values a
// pattern would store and release, so they do not move with machine load.
//
//   node fabric/patterns/tools/measure-bytes.mjs blackjack
//   node fabric/patterns/tools/measure-bytes.mjs memory
//
// The game is the one the browser scripts play: three players at blackjack
// for two rounds, hitting below 15 and standing otherwise; two players at
// Memory revealing cards in order. The deal's randomness is fixed.
//
// "written" models the store's diff: the JSON size of every path whose value
// differs from what was stored before, a changed scalar or a new subtree.
import { stepTable } from '../toolkit/table.ts';
import { looking, storedOf, viewOfTable } from './measure-adapter.ts';

const which = process.argv[2] ?? 'blackjack';
const { game, seats, options, next } = await import(`./measure-${which}.ts`).then(module => module.scenario());

const viewers = looking(seats, which);
const size = value => value === undefined ? 0 : Buffer.byteLength(JSON.stringify(value));

/** The JSON size of everything in `after` that is not already in `before`. */
function written(before, after) {
  if (before === after) return 0;
  const plain = value => value !== null && typeof value === 'object';
  if (!plain(before) || !plain(after) || Array.isArray(before) !== Array.isArray(after)) {
    return JSON.stringify(before) === JSON.stringify(after) ? 0 : size(after);
  }
  let total = 0;
  for (const key of Object.keys(after)) total += key in before ? written(before[key], after[key]) : size(after[key]) + key.length;
  return total;
}

let seed = 1;
const entropy = () => {
  seed = (seed * 1103515245 + 12345) % 2147483648;
  return seed / 2147483648;
};
const principal = seat => `did:measure:seat-${seat}`;

let table = undefined;
const rows = [];
function step(label, proposal, seat) {
  const before = table;
  const beforeViews = viewers.map(viewer => viewOfTable(game, before, viewer, principal));
  const result = stepTable(game, storedOf(before), proposal, { proposer: principal(seat), now: 1_700_000_000_000 + rows.length * 1000, entropy });
  if (result.error !== '') throw new Error(`${label}: ${result.error}`);
  table = JSON.parse(JSON.stringify(result.table));
  const views = viewers.map(viewer => viewOfTable(game, table, viewer, principal));
  rows.push({
    label,
    versions: table.version - (before?.version ?? -1),
    stored: size(table),
    storedWritten: written(before ?? {}, table),
    viewEach: Math.round(views.reduce((sum, view) => sum + size(view), 0) / views.length),
    viewsTotal: views.reduce((sum, view) => sum + size(view), 0),
    viewsWritten: views.reduce((sum, view, index) => sum + written(beforeViews[index] ?? {}, view), 0),
  });
}

for (let seat = 0; seat < seats; seat++) step(`sit ${seat}`, { kind: 'sit', label: `Player ${seat + 1}` }, seat);
step('deal', { kind: 'start', options }, 0);
for (let guard = 0; guard < 400 && table.phase === 'playing'; guard++) {
  const proposal = next(table);
  if (proposal === undefined) break;
  step(proposal.label, proposal.proposal, proposal.seat);
}

const moves = rows.filter(row => !row.label.startsWith('sit') && row.label !== 'deal');
const median = values => [...values].sort((a, b) => a - b)[Math.floor(values.length / 2)];
const column = name => ({ median: median(moves.map(row => row[name])), max: Math.max(...moves.map(row => row[name])) });
console.log(JSON.stringify({
  game: which,
  viewers: viewers.length,
  phase: table.phase,
  deal: rows.find(row => row.label === 'deal'),
  moves: moves.length,
  perMove: {
    versions: column('versions'),
    stored: column('stored'),
    storedWritten: column('storedWritten'),
    viewEach: column('viewEach'),
    viewsTotal: column('viewsTotal'),
    viewsWritten: column('viewsWritten'),
  },
  last: rows[rows.length - 1],
}, null, 1));
if (process.env.ROWS) for (const row of rows) console.log(JSON.stringify(row));
