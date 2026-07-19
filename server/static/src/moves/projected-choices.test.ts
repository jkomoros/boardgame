import assert from 'node:assert/strict';
import test from 'node:test';
import {
  MoveSubmissionGate,
  createMoveAction,
  type BoundMoveAction,
  type MoveActionService,
  type MoveActionSnapshot,
} from './action.ts';
import { serializeCreatorMoveInput, validateCreatorMoveInput } from './input.ts';
import { buildProjectedMoveChoices } from './projected-choices.ts';

type Names = 'Choose Player' | 'Guess Card';
type Inputs = {
  'Choose Player': { TargetPlayer: number };
  'Guess Card': { GuessedCard: 'Guard' | 'Priest' };
};
type Projections = {
  'Choose Player': { readonly field: 'TargetPlayer'; readonly value: number; readonly input: Inputs['Choose Player'] };
  'Guess Card': { readonly field: 'GuessedCard'; readonly value: 'Guard' | 'Priest'; readonly input: Inputs['Guess Card'] };
};

const inputSchema = [
  { name: 'Choose Player', fields: [{ name: 'TargetPlayer', wireType: 'int', disposition: 'required', codec: 'player-index' }] },
  { name: 'Guess Card', fields: [{ name: 'GuessedCard', wireType: 'string', disposition: 'required', codec: 'enum', enumName: 'Card', enumValues: ['Guard', 'Priest'] }] },
] as const;

const projectionSchema = [
  { moveName: 'Choose Player', fieldName: 'TargetPlayer', source: 'players', disclosure: 'actor-exact' },
  { moveName: 'Guess Card', fieldName: 'GuessedCard', source: 'enum-values', candidateValues: ['Guard', 'Priest'], disclosure: 'actor-exact' },
] as const;

function actions(): <K extends keyof Projections & string>(
  move: K,
  input: Projections[K]['input'],
) => BoundMoveAction<K, Projections[K]['input']> {
  const service: MoveActionService = {
    currentClientSchemaFingerprint: () => 'input-fingerprint',
    currentServerSchemaFingerprint: () => 'input-fingerprint',
    currentTransport: () => ({ submit: async () => ({ kind: 'success' }) }),
    currentPreviewTransport: () => null,
    currentTargetPreviewTransport: () => null,
    currentGate: () => new MoveSubmissionGate(),
    nextRequestID: () => 'request',
    validate: (name, input) => validateCreatorMoveInput(inputSchema, name, input).errors,
    serialize: (name, input) => serializeCreatorMoveInput(inputSchema, name, input),
    actionCache: new Map(),
  };
  const snapshot: MoveActionSnapshot = {
    snapshotKey: 'v7', currentSnapshotKey: () => 'v7',
    snapshotVersion: 7, currentSnapshotVersion: () => 7,
    viewingAsPlayer: 0, proposingAsPlayer: 0, proposingAsAdmin: false,
    currentLegality: () => ({ legalForPlayer: false, legalForAnyone: true }),
    currentAnimating: () => false, baselineLegalityApplies: true,
  };
  return (move, input) => {
    const builder = createMoveAction<keyof Projections & string, Names, Inputs>(move, service, snapshot);
    return builder.with(input) as BoundMoveAction<typeof move, Projections[typeof move]['input']>;
  };
}

function readyWire() {
  return {
    StateVersion: 7,
    MoveChoiceProjectionSchemaFingerprint: 'projection-fingerprint',
    ProjectionSchemaVersion: 1,
    Status: 'ready' as const,
    Sets: [
      {
        MoveName: 'Choose Player', FieldName: 'TargetPlayer', Source: 'players' as const,
        Candidates: [{ Value: 0, Available: false }, { Value: 1, Available: true }],
      },
      {
        MoveName: 'Guess Card', FieldName: 'GuessedCard', Source: 'enum-values' as const,
        Candidates: [{ Value: 'Guard', Available: true }, { Value: 'Priest', Available: false }],
      },
    ],
  };
}

test('validates projections before creating exact typed ordinary actions', () => {
  const choices = buildProjectedMoveChoices<Projections>({
    wire: readyWire(), stateVersion: 7, schema: projectionSchema,
    schemaFingerprint: 'projection-fingerprint',
    playerPresentations: [
      { playerIndex: 0, label: 'Ada' },
      { playerIndex: 1, label: 'Grace' },
    ],
    action: actions(),
    messages: { 'Guess Card': { id: 'valentine.guess', defaultMessage: 'Name their card' } },
  });
  assert.equal(choices.status, 'ready');
  const players = choices.get('Choose Player');
  assert.deepEqual(players?.candidates.map(candidate => [
    candidate.id, candidate.value, candidate.message.defaultMessage, candidate.action.preview.kind,
  ]), [
    ['choose-player:target-player:0', 0, 'Ada', 'illegal'],
    ['choose-player:target-player:1', 1, 'Grace', 'legal'],
  ]);
  const cards = choices.get('Guess Card');
  assert.equal(cards?.message.defaultMessage, 'Name their card');
  assert.deepEqual(cards?.candidates.map(candidate => [candidate.id, candidate.message.defaultMessage]), [
    ['guess-card:guessed-card:guard', 'Guard'], ['guess-card:guessed-card:priest', 'Priest'],
  ]);
  // Complete bound legality supersedes the default move form's false baseline.
  assert.equal(cards?.candidates[0].action.canPropose, true);
  assert.equal(cards?.candidates[1].action.reason?.code, 'preview-illegal');
});

test('accepts a ready subset but rejects spoofed candidate universes', () => {
  const base = readyWire();
  const subset = { ...base, Sets: [base.Sets[1]] };
  const choices = buildProjectedMoveChoices<Projections>({
    wire: subset, stateVersion: 7, schema: projectionSchema,
    schemaFingerprint: 'projection-fingerprint', playerPresentations: [], action: actions(),
  });
  assert.equal(choices.get('Choose Player'), null);
  assert.equal(choices.get('Guess Card')?.candidates.length, 2);

  const spoofed = {
    ...subset,
    Sets: [{ ...base.Sets[1], Candidates: [{ Value: 'Princess', Available: true }] }],
  };
  assert.throws(() => buildProjectedMoveChoices<Projections>({
    wire: spoofed, stateVersion: 7, schema: projectionSchema,
    schemaFingerprint: 'projection-fingerprint', playerPresentations: [], action: actions(),
  }), /do not match the generated universe/);
});

test('preserves an explicit projection failure as renderable state', () => {
  const choices = buildProjectedMoveChoices<Projections>({
    wire: {
      StateVersion: 7,
      MoveChoiceProjectionSchemaFingerprint: 'projection-fingerprint',
      ProjectionSchemaVersion: 1,
      Status: 'failed',
      Sets: [],
    },
    stateVersion: 7, schema: projectionSchema,
    schemaFingerprint: 'projection-fingerprint', playerPresentations: [], action: actions(),
  });
  assert.equal(choices.status, 'failed');
  assert.match(choices.message?.defaultMessage ?? '', /temporarily unavailable/);
});

test('rejects all-disabled included sets so the renderer surfaces failure', () => {
  const base = readyWire();
  const allDisabled = {
    ...base,
    Sets: [{
      ...base.Sets[1],
      Candidates: base.Sets[1].Candidates.map(candidate => ({ ...candidate, Available: false })),
    }],
  };
  assert.throws(() => buildProjectedMoveChoices<Projections>({
    wire: allDisabled, stateVersion: 7, schema: projectionSchema,
    schemaFingerprint: 'projection-fingerprint', playerPresentations: [], action: actions(),
  }), /no available candidate/);
});

test('rejects stale snapshots and generated-schema mismatches', () => {
  const common = {
    wire: readyWire(), schema: projectionSchema,
    schemaFingerprint: 'projection-fingerprint',
    playerPresentations: [
      { playerIndex: 0, label: 'Ada' }, { playerIndex: 1, label: 'Grace' },
    ],
    action: actions(),
  };
  assert.throws(() => buildProjectedMoveChoices<Projections>({ ...common, stateVersion: 8 }), /state version/);
  assert.throws(() => buildProjectedMoveChoices<Projections>({
    ...common, stateVersion: 7, schemaFingerprint: 'wrong',
  }), /fingerprint/);
});
