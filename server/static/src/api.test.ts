// Unit tests for the legality-preview API client methods (movePreview /
// movePreviewBatch). Run with `node --test` (Node >=23 native TS). These pin the
// wire contract with the server's two preview endpoints: the single-move
// preview parses FORM-encoded args like the real move endpoint, the batch parses
// JSON {MoveType, Candidates}; both flow through apiPost so the Success/Failure
// envelope is unwrapped for free. fetch is stubbed — no network, no DOM.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { movePreview, movePreviewBatch } from './api.ts';

interface Captured {
  url: string;
  opts: { method: string; headers: Record<string, string>; body: string };
}

// stubFetch installs a fake global fetch that captures the one request made and
// returns the given JSON envelope. Returns a getter for the captured request.
function stubFetch(responseJson: unknown, status = 200): () => Captured {
  let captured: Captured | undefined;
  (globalThis as unknown as { fetch: unknown }).fetch = async (url: string, opts: Captured['opts']) => {
    captured = { url, opts };
    return { status, statusText: 'OK', json: async () => responseJson };
  };
  return () => {
    if (!captured) throw new Error('fetch was never called');
    return captured;
  };
}

test('movePreview posts form-encoded MoveType+args to the movePreview path and unwraps the envelope', async () => {
  const get = stubFetch({ Status: 'Success', Form: { Name: 'Place Token', LegalForPlayer: false, LegalForPlayerError: 'nope' } });

  const res = await movePreview('checkers', 'g1', 'Place Token', { Slot: '5' });

  const cap = get();
  assert.equal(cap.url, '/api/game/checkers/g1/movePreview');
  assert.equal(cap.opts.method, 'POST');
  assert.equal(cap.opts.headers['Content-Type'], 'application/x-www-form-urlencoded');
  // Form-encoded body carries MoveType plus one entry per field — exactly what
  // getMoveFromForm reads via c.PostForm.
  const params = new URLSearchParams(cap.opts.body);
  assert.equal(params.get('MoveType'), 'Place Token');
  assert.equal(params.get('Slot'), '5');
  // The Success envelope is unwrapped into .data.
  assert.equal(res.data?.Form.LegalForPlayer, false);
  assert.equal(res.data?.Form.LegalForPlayerError, 'nope');
  assert.equal(res.error, undefined);
});

test('movePreviewBatch posts versioned, correlated candidates to the batch path', async () => {
  const get = stubFetch({ Status: 'Success', Results: [
    { ID: 'cell:0', Legal: true },
    { ID: 'cell:1', Legal: false, Error: 'blocked' },
  ] });

  const controller = new AbortController();
  const res = await movePreviewBatch('tictactoe', 'g2', 'Place Token', [
    { ID: 'cell:0', Args: { Slot: '0' } },
    { ID: 'cell:1', Args: { Slot: '1' } },
  ], undefined, 7, controller.signal);

  const cap = get();
  assert.equal(cap.url, '/api/game/tictactoe/g2/movePreviewBatch');
  assert.equal(cap.opts.method, 'POST');
  assert.equal(cap.opts.headers['Content-Type'], 'application/json');
  const body = JSON.parse(cap.opts.body);
  assert.equal(body.MoveType, 'Place Token');
  assert.equal(body.ExpectedVersion, 7);
  assert.deepEqual(body.Candidates, [
    { ID: 'cell:0', Args: { Slot: '0' } },
    { ID: 'cell:1', Args: { Slot: '1' } },
  ]);
  assert.equal(cap.opts.signal, controller.signal);
  assert.equal(res.data?.Results.length, 2);
  assert.equal(res.data?.Results[0].Legal, true);
  assert.equal(res.data?.Results[0].ID, 'cell:0');
  assert.equal(res.data?.Results[1].Legal, false);
  assert.equal(res.data?.Results[1].Error, 'blocked');
});

test('a Failure envelope preserves structured stale metadata, never as legal data', async () => {
  stubFetch({
    Status: 'Failure', Error: 'bad move type', FriendlyError: 'That move is not available',
    Code: 'STALE_SNAPSHOT', ExpectedVersion: 3, ActualVersion: 4,
  });

  const res = await movePreview('checkers', 'g1', 'Nope', {});

  assert.equal(res.data, undefined);
  assert.equal(res.error, 'bad move type');
  assert.equal(res.friendlyError, 'That move is not available');
  assert.equal(res.code, 'STALE_SNAPSHOT');
  assert.equal(res.expectedVersion, 3);
  assert.equal(res.actualVersion, 4);
});

test('optional params (e.g. previewing as a specific player) pass through as a query string', async () => {
  const get = stubFetch({ Status: 'Success', Results: [] });

  await movePreviewBatch('memory', 'g3', 'Reveal Card', [], { player: 2 });

  assert.equal(get().url, '/api/game/memory/g3/movePreviewBatch?player=2');
});

test('malformed API envelopes fail closed with actionable diagnostics', async () => {
  stubFetch(['Success']);
  const nonObject = await movePreview('memory', 'g4', 'Reveal Card', {});
  assert.equal(nonObject.data, undefined);
  assert.equal(nonObject.error, 'Invalid API response: expected an object envelope');
  assert.equal(nonObject.friendlyError, 'The server returned an invalid response');

  stubFetch({ Status: 'Maybe', Form: {} });
  const unknownStatus = await movePreview('memory', 'g4', 'Reveal Card', {});
  assert.equal(unknownStatus.data, undefined);
  assert.equal(unknownStatus.error, 'Invalid API response: Status must be "Success" or "Failure"');
});

test('malformed failure metadata cannot masquerade as structured stale data', async () => {
  stubFetch({
    Status: 'Failure',
    Error: 7,
    FriendlyError: false,
    Code: 9,
    ExpectedVersion: '3',
    ActualVersion: Number.NaN,
  }, 409);

  const result = await movePreview('memory', 'g5', 'Reveal Card', {});
  assert.equal(result.error, 'Request failed with status 409');
  assert.equal(result.friendlyError, 'An error occurred');
  assert.equal(result.code, undefined);
  assert.equal(result.expectedVersion, undefined);
  assert.equal(result.actualVersion, undefined);
});
