import assert from 'node:assert/strict';
import test from 'node:test';

import {
  assertMoveInputSchemaFingerprint,
	creatorMoveInputFromLegacyStrings,
  serializeCreatorMoveInput,
	serializeCreatorMoveInputForServer,
  StaleMoveInputSchemaError,
  validateCreatorMoveInput,
  type MoveInputSchema,
} from './input.ts';

const schema = [{
  name: 'Choose',
  fields: [
    { name: 'Count', wireType: 'int', disposition: 'required', codec: 'integer' },
    { name: 'Enabled', wireType: 'bool', disposition: 'server-defaulted', codec: 'boolean' },
    { name: 'Mode', wireType: 'enum', disposition: 'required', codec: 'enum', enumName: 'mode', enumValues: ['Fast', 'Safe'] },
    { name: 'TargetPlayerIndex', wireType: 'playerIndex', disposition: 'context-owned', codec: 'player-index' },
  ],
}] as const satisfies MoveInputSchema;

test('validates and serializes native creator input', () => {
  assert.deepEqual(validateCreatorMoveInput(schema, 'Choose', { Count: 2, Mode: 'Fast', Enabled: false }), { valid: true, errors: [] });
  assert.deepEqual(serializeCreatorMoveInput(schema, 'Choose', { Count: 2, Mode: 'Fast', Enabled: false }), {
    Count: '2', Mode: 'Fast', Enabled: '0',
  });
});

test('rejects missing, extra, context, fractional, non-finite, and enum inputs', () => {
  const cases: Array<[Record<string, unknown>, string]> = [
    [{ Mode: 'Fast' }, 'missing-field'],
    [{ Count: 2, Mode: 'Fast', Surprise: true }, 'unknown-field'],
    [{ Count: 2, Mode: 'Fast', TargetPlayerIndex: 0 }, 'context-owned-field'],
    [{ Count: 1.5, Mode: 'Fast' }, 'invalid-type'],
    [{ Count: Number.POSITIVE_INFINITY, Mode: 'Fast' }, 'invalid-type'],
    [{ Count: 2, Mode: 'Unknown' }, 'invalid-value'],
  ];
  for (const [input, code] of cases) {
    const result = validateCreatorMoveInput(schema, 'Choose', input);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some(item => item.code === code), `${code}: ${JSON.stringify(result.errors)}`);
  }
});

test('rejects non-object roots with structured diagnostics', () => {
  for (const input of [null, [], 'bad']) {
    const result = validateCreatorMoveInput(schema, 'Choose', input);
    assert.equal(result.valid, false);
    assert.equal(result.errors[0]?.code, 'invalid-input');
  }
});

test('reserved proposal protocol fields are rejected as context-owned', () => {
  for (const name of ['MoveType', 'admin', 'player', 'ExpectedVersion']) {
    const reservedSchema = [{
      name: 'Reserved',
      fields: [{ name, wireType: 'string', disposition: 'required', codec: 'string' }],
    }] as const satisfies MoveInputSchema;
    const result = validateCreatorMoveInput(reservedSchema, 'Reserved', { [name]: 'value' });
    assert.equal(result.valid, false);
    assert.equal(result.errors[0]?.code, 'context-owned-field');
    assert.match(result.errors[0]?.message ?? '', /reserved/);
  }
});

test('fingerprints fail closed for mismatch and absence', () => {
  assert.doesNotThrow(() => assertMoveInputSchemaFingerprint('same', 'same'));
  assert.throws(() => assertMoveInputSchemaFingerprint('client', 'server'), StaleMoveInputSchemaError);
  assert.throws(() => assertMoveInputSchemaFingerprint('client', undefined), StaleMoveInputSchemaError);
});

test('safe proposal serialization checks server freshness before values', () => {
	assert.deepEqual(serializeCreatorMoveInputForServer(schema, 'same', 'same', 'Choose', { Count: 2, Mode: 'Fast' }), {
		Count: '2', Mode: 'Fast',
	});
	assert.throws(
		() => serializeCreatorMoveInputForServer(schema, 'client', undefined, 'Choose', { Count: 2, Mode: 'Fast' }),
		(error: unknown) => error instanceof StaleMoveInputSchemaError && error.code === 'BOARDGAME_STALE_MOVE_INPUT_SCHEMA',
	);
	assert.throws(
		() => serializeCreatorMoveInputForServer(schema, 'client', 'server', 'Choose', { Count: 2, Mode: 'Fast' }),
		StaleMoveInputSchemaError,
	);
});

test('legacy dataset strings are coerced into the native contract before validation', () => {
	assert.deepEqual(creatorMoveInputFromLegacyStrings(schema, 'Choose', {
		Count: '-2', Enabled: '0', Mode: 'Safe', Surprise: 'kept-for-validation',
	}), {
		Count: -2, Enabled: false, Mode: 'Safe', Surprise: 'kept-for-validation',
	});
	assert.deepEqual(creatorMoveInputFromLegacyStrings(schema, 'Choose', {
		Count: '2.5', Enabled: 'true', Mode: 'Fast',
	}), {
		Count: '2.5', Enabled: 'true', Mode: 'Fast',
	});
});
