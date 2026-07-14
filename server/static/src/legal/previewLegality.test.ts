// Unit tests for disabledSpacesFromResults — the pure decision at the heart of
// board legality preview: given the candidates sent to movePreviewBatch and the
// results that came back (in order), which board spaces should be grayed. Run
// with `node --test`. The fail-safe rule (a missing/short result grays the space
// rather than inviting a click the server would reject) is the load-bearing
// behavior pinned here.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  disabledSpacesFromResults,
  samePreviewSpaces,
  previewOutcome,
  type PreviewCandidate,
} from './previewLegality.ts';

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

// samePreviewSpaces — the guard that skips re-rendering the board when the
// grayed set hasn't actually changed (disabledSpacesFromResults yields spaces in
// stable candidate order, so an order-sensitive compare is correct).
test('samePreviewSpaces: identical arrays are equal', () => {
  assert.equal(samePreviewSpaces([0, 4, 8], [0, 4, 8]), true);
  assert.equal(samePreviewSpaces([], []), true);
});

test('samePreviewSpaces: different length or members are not equal', () => {
  assert.equal(samePreviewSpaces([0, 4], [0, 4, 8]), false);
  assert.equal(samePreviewSpaces([0, 4, 8], [0, 5, 8]), false);
  assert.equal(samePreviewSpaces([0, 4, 8], [8, 4, 0]), false); // order matters (stable candidate order)
});

// previewOutcome — the completed-refresh decision: apply only when this refresh
// is still the latest (seq unchanged), its renderer is still mounted, and the
// server actually returned data; otherwise drop it (superseded) or keep prior
// graying (transient error).
test('previewOutcome: latest + mounted + data -> apply', () => {
  assert.equal(previewOutcome({ startedSeq: 3, currentSeq: 3, rendererStillMounted: true, hasData: true }), 'apply');
});

test('previewOutcome: a newer refresh superseded this one -> drop-stale', () => {
  assert.equal(previewOutcome({ startedSeq: 3, currentSeq: 4, rendererStillMounted: true, hasData: true }), 'drop-stale');
});

test('previewOutcome: renderer no longer mounted -> drop-stale (even with data)', () => {
  assert.equal(previewOutcome({ startedSeq: 3, currentSeq: 3, rendererStillMounted: false, hasData: true }), 'drop-stale');
});

test('previewOutcome: latest + mounted but no data (error) -> keep-on-error', () => {
  assert.equal(previewOutcome({ startedSeq: 3, currentSeq: 3, rendererStillMounted: true, hasData: false }), 'keep-on-error');
});

test('previewOutcome: staleness beats the error branch', () => {
  // Superseded AND errored: we drop (don't touch graying), we don't "keep".
  assert.equal(previewOutcome({ startedSeq: 1, currentSeq: 2, rendererStillMounted: true, hasData: false }), 'drop-stale');
});
