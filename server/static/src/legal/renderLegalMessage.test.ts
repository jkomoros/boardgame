// Unit tests for renderLegalMessage — the TS mirror of Go's RenderLegalMessage
// (legal_error.go). Run with `node --test` (Node >=23 native TS). These pin
// BYTE-FOR-BYTE parity with the Go renderer: same placeholder regex, same
// raw-key fallback, same bare-placeholder-name behavior for a missing binding,
// same int/bool formatting. Any divergence here is a client bug.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { renderLegalMessage } from './renderLegalMessage.ts';
import type { PreconditionMessage } from '../types/api';

const msg = (
  template: string,
  bindings?: Record<string, string | number | boolean>,
): PreconditionMessage => (bindings ? { template, bindings } : { template });

test('nil/undefined message renders empty string (Go: m == nil -> "")', () => {
  assert.equal(renderLegalMessage(undefined, {}), '');
  assert.equal(renderLegalMessage(null as unknown as undefined, {}), '');
});

test('template present in table, no placeholders -> body verbatim', () => {
  const table = { 'reveal.no_cards_left': 'You have no cards left to reveal this turn' };
  assert.equal(
    renderLegalMessage(msg('reveal.no_cards_left'), table),
    'You have no cards left to reveal this turn',
  );
});

test('template ABSENT from table -> falls back to the raw key string (Go: !ok -> body = template)', () => {
  assert.equal(renderLegalMessage(msg('checkers.black_spaces_only'), {}), 'checkers.black_spaces_only');
  // undefined table behaves the same as an empty table
  assert.equal(renderLegalMessage(msg('some.key'), undefined), 'some.key');
});

test('string binding is substituted', () => {
  const table = { k: 'the space at {where} is taken' };
  assert.equal(renderLegalMessage(msg('k', { where: 'A1' }), table), 'the space at A1 is taken');
});

test('integer binding formats as the integer (Go strconv.Itoa)', () => {
  const table = { k: 'requires at least {min}' };
  assert.equal(renderLegalMessage(msg('k', { min: 3 }), table), 'requires at least 3');
  assert.equal(renderLegalMessage(msg('k', { min: 0 }), table), 'requires at least 0');
});

test('boolean binding formats as true/false (Go strconv.FormatBool)', () => {
  const table = { k: 'requires {prop} to be {want}' };
  assert.equal(
    renderLegalMessage(msg('k', { prop: 'DieCounted', want: false }), table),
    'requires DieCounted to be false',
  );
  assert.equal(renderLegalMessage(msg('k', { want: true }), { k: '{want}' }), 'true');
});

test('MISSING binding renders the bare placeholder NAME (not blank, never throws)', () => {
  const table = { k: 'you have {value} cards' };
  assert.equal(renderLegalMessage(msg('k', {}), table), 'you have value cards');
  // multiple, mixed present/absent
  assert.equal(
    renderLegalMessage(msg('k2', { a: 'X' }), { k2: '{a}-{b}-{a}' }),
    'X-b-X',
  );
});

test('bindings undefined (stripped by #693 guard) -> every placeholder renders bare', () => {
  const table = { k: '{detail} at {index}' };
  assert.equal(renderLegalMessage(msg('k'), table), 'detail at index');
});

test('a PRESENT but falsy binding (0, false, "") uses the value, not the bare name', () => {
  assert.equal(renderLegalMessage(msg('k', { n: 0 }), { k: '{n}' }), '0');
  assert.equal(renderLegalMessage(msg('k', { b: false }), { k: '{b}' }), 'false');
  assert.equal(renderLegalMessage(msg('k', { s: '' }), { k: '{s}' }), '');
});

test('placeholder grammar matches Go regex \\{([A-Za-z0-9_]+)\\} exactly', () => {
  // underscores + digits in names are valid
  assert.equal(renderLegalMessage(msg('k', { a_1: 'Z' }), { k: '{a_1}' }), 'Z');
  // empty braces and names with spaces are NOT placeholders -> left literal
  assert.equal(renderLegalMessage(msg('k'), { k: 'a {} b {x y} c' }), 'a {} b {x y} c');
  // hyphen is not a word char -> not a placeholder
  assert.equal(renderLegalMessage(msg('k'), { k: '{a-b}' }), '{a-b}');
});

test('present-in-table but EMPTY body stays empty (comma-ok, not truthiness)', () => {
  assert.equal(renderLegalMessage(msg('k'), { k: '' }), '');
});
