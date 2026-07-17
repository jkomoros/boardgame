import assert from 'node:assert/strict';
import test from 'node:test';
import { codeFromJoinRoute, JoinSessionScope } from './join-session.ts';

test('join route parsing is scoped to the current route query', () => {
  assert.equal(codeFromJoinRoute('?code=abcd'), 'ABCD');
  assert.equal(codeFromJoinRoute('anything?code=abcde&other=1'), 'ABCDE');
  assert.equal(codeFromJoinRoute('?other=1'), null);
});

test('a new join route aborts and invalidates work from the old room', () => {
  const scope = new JoinSessionScope();
  assert.equal(scope.activate('?code=AAAA'), true);
  const first = scope.begin();
  assert.equal(scope.isCurrent(first), true);

  assert.equal(scope.activate('?code=BBBB'), true);
  assert.equal(first.controller.signal.aborted, true);
  assert.equal(scope.isCurrent(first), false);
  const second = scope.begin();
  assert.equal(scope.isCurrent(second), true);

  scope.deactivate();
  assert.equal(second.controller.signal.aborted, true);
  assert.equal(scope.isCurrent(second), false);
});

test('duplicate activation and request supersession are deterministic', () => {
  const scope = new JoinSessionScope();
  assert.equal(scope.activate('?code=AAAA'), true);
  assert.equal(scope.activate('?code=AAAA'), false);
  const first = scope.begin();
  const second = scope.begin();
  assert.equal(first.controller.signal.aborted, true);
  assert.equal(scope.isCurrent(first), false);
  assert.equal(scope.isCurrent(second), true);
});
