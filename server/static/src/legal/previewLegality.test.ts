// Unit tests for disabledSpacesFromResults — the pure decision at the heart of
// board legality preview: given the candidates sent to movePreviewBatch and the
// results that came back (in order), which board spaces should be grayed. Run
// with `node --test`. The fail-safe rule (a missing/short result grays the space
// rather than inviting a click the server would reject) is the load-bearing
// behavior pinned here.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { disabledSpacesFromResults, type PreviewCandidate } from './previewLegality.ts';

const cand = (space: number): PreviewCandidate => ({ space, args: { Slot: String(space) } });

test('all legal -> nothing grayed', () => {
  const candidates = [cand(0), cand(1), cand(2)];
  const results = [{ Legal: true }, { Legal: true }, { Legal: true }];
  assert.deepEqual(disabledSpacesFromResults(candidates, results), []);
});

test('grays exactly the illegal candidates, by their space (not loop index)', () => {
  // Non-contiguous spaces prove the result maps back to candidate.space.
  const candidates = [cand(10), cand(20), cand(30)];
  const results = [{ Legal: false, Error: 'occupied' }, { Legal: true }, { Legal: false }];
  assert.deepEqual(disabledSpacesFromResults(candidates, results), [10, 30]);
});

test('a short results array grays the uncovered candidates (fail safe)', () => {
  const candidates = [cand(0), cand(1), cand(2)];
  const results = [{ Legal: true }]; // server returned fewer than we asked
  assert.deepEqual(disabledSpacesFromResults(candidates, results), [1, 2]);
});

test('undefined results grays every candidate (fail safe)', () => {
  const candidates = [cand(5), cand(6)];
  assert.deepEqual(disabledSpacesFromResults(candidates, undefined), [5, 6]);
});

test('no candidates -> no disabled spaces', () => {
  assert.deepEqual(disabledSpacesFromResults([], [{ Legal: true }]), []);
});
