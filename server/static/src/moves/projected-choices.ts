import {
  hydrateMoveActionPreview,
  type BoundMoveAction,
} from './action.ts';
import type { ProjectedMoveChoiceSetWire, ProjectedMoveChoicesWire } from '../types/api.js';
import type { PlayerPresentation } from '../status/player-presentation.js';

export const MOVE_CHOICE_PROJECTION_SCHEMA_VERSION = 1;

export interface MessageDescriptor {
  readonly id: string;
  readonly defaultMessage: string;
}

export type MessageResolver = (message: MessageDescriptor) => string;

export const defaultMessageResolver: MessageResolver = message => message.defaultMessage;

export interface MoveChoiceProjectionSchemaEntry {
  readonly moveName: string;
  readonly fieldName: string;
  readonly source: 'players' | 'enum-values';
  readonly candidateValues?: readonly string[];
  readonly disclosure: 'actor-exact';
}

/** Generated games specialize this map with one property per projected move. */
export type MoveChoiceProjectionTypes = Readonly<Record<string, {
  readonly field: string;
  readonly value: string | number;
  readonly input: object;
}>>;

type ProjectionFor<
  Projections extends MoveChoiceProjectionTypes,
  MoveName extends keyof Projections,
> = Projections[MoveName];

export interface ProjectedMoveChoiceCandidate<
  MoveName extends string,
  Value extends string | number,
  Input extends object,
> {
  /** Stable semantic identity, suitable for keyed rendering and analytics. */
  readonly id: string;
  readonly value: Value;
  /** Client-owned, localizable candidate presentation. */
  readonly message: MessageDescriptor;
  readonly available: boolean;
  /** The canonical ordinary action, already hydrated legal or illegal. */
  readonly action: BoundMoveAction<MoveName, Input>;
}

export interface ProjectedMoveChoiceSet<
  MoveName extends string,
  Projection extends { readonly field: string; readonly value: string | number; readonly input: object },
> {
  readonly move: MoveName;
  readonly field: Projection['field'];
  readonly message: MessageDescriptor;
  readonly candidates: readonly ProjectedMoveChoiceCandidate<
    MoveName,
    Projection['value'],
    Projection['input']
  >[];
}

type AnyProjectedSet<Projections extends MoveChoiceProjectionTypes> = {
  [MoveName in keyof Projections & string]: ProjectedMoveChoiceSet<MoveName, Projections[MoveName]>
}[keyof Projections & string];

export type ProjectedMoveChoicesStatus = 'ready' | 'failed';

/** Immutable, exact typed view over one validated projection snapshot. */
export class ProjectedMoveChoices<Projections extends MoveChoiceProjectionTypes> {
  readonly status: ProjectedMoveChoicesStatus;
  readonly message: MessageDescriptor | null;
  readonly #sets: ReadonlyMap<string, AnyProjectedSet<Projections>>;

  private constructor(
    status: ProjectedMoveChoicesStatus,
    sets: ReadonlyMap<string, AnyProjectedSet<Projections>>,
    message: MessageDescriptor | null,
  ) {
    this.status = status;
    this.#sets = sets;
    this.message = message;
    Object.freeze(this);
  }

  static failed<Projections extends MoveChoiceProjectionTypes>(): ProjectedMoveChoices<Projections> {
    return new ProjectedMoveChoices('failed', new Map(), PROJECTION_FAILED_MESSAGE);
  }

  static ready<Projections extends MoveChoiceProjectionTypes>(
    sets: ReadonlyMap<string, AnyProjectedSet<Projections>>,
  ): ProjectedMoveChoices<Projections> {
    return new ProjectedMoveChoices('ready', sets, null);
  }

  get<MoveName extends keyof Projections & string>(
    move: MoveName,
  ): ProjectedMoveChoiceSet<MoveName, ProjectionFor<Projections, MoveName>> | null {
    return (this.#sets.get(move) as ProjectedMoveChoiceSet<
      MoveName,
      ProjectionFor<Projections, MoveName>
    > | undefined) ?? null;
  }

  all(): readonly AnyProjectedSet<Projections>[] {
    return Object.freeze([...this.#sets.values()]);
  }
}

export const PROJECTION_FAILED_MESSAGE: MessageDescriptor = Object.freeze({
  id: 'boardgame.projected-choices.failed',
  defaultMessage: 'Choices are temporarily unavailable. Refresh the game to try again.',
});

export function defaultProjectedChoiceMessage(moveName: string, fieldName: string): MessageDescriptor {
  return Object.freeze({
    id: `boardgame.projected-choices.${semanticToken(moveName)}.${semanticToken(fieldName)}`,
    defaultMessage: `Choose ${humanize(fieldName)}`,
  });
}

export interface BuildProjectedMoveChoicesOptions<Projections extends MoveChoiceProjectionTypes> {
  readonly wire: ProjectedMoveChoicesWire;
  readonly stateVersion: number;
  readonly schema: readonly MoveChoiceProjectionSchemaEntry[];
  readonly schemaFingerprint: string;
  readonly playerPresentations: readonly PlayerPresentation[];
  readonly action: <MoveName extends keyof Projections & string>(
    move: MoveName,
    input: Projections[MoveName]['input'],
  ) => BoundMoveAction<MoveName, Projections[MoveName]['input']>;
  readonly messages?: Readonly<Partial<Record<keyof Projections & string, MessageDescriptor>>>;
}

/** Validate untrusted wire data before narrowing values into generated types. */
export function buildProjectedMoveChoices<Projections extends MoveChoiceProjectionTypes>(
  options: BuildProjectedMoveChoicesOptions<Projections>,
): ProjectedMoveChoices<Projections> {
  const { wire } = options;
  if (wire.StateVersion !== options.stateVersion) {
    throw new Error(`Projected choices state version ${wire.StateVersion} does not match snapshot ${options.stateVersion}`);
  }
  if (wire.ProjectionSchemaVersion !== MOVE_CHOICE_PROJECTION_SCHEMA_VERSION) {
    throw new Error(`Unsupported projected-choice schema version ${wire.ProjectionSchemaVersion}`);
  }
  if (!options.schemaFingerprint || wire.MoveChoiceProjectionSchemaFingerprint !== options.schemaFingerprint) {
    throw new Error('Projected-choice schema fingerprint does not match the generated client');
  }
  if (wire.Status === 'failed') return ProjectedMoveChoices.failed<Projections>();

  const schemaByMove = new Map(options.schema.map(entry => [entry.moveName, entry]));
  if (schemaByMove.size !== options.schema.length) {
    throw new Error('Generated projected-choice schema contains duplicate move names');
  }
  const result = new Map<string, AnyProjectedSet<Projections>>();
  for (const wireSet of wire.Sets) {
    const schema = schemaByMove.get(wireSet.MoveName);
    if (!schema) throw new Error(`Projected choices contains unknown move ${JSON.stringify(wireSet.MoveName)}`);
    if (result.has(wireSet.MoveName)) {
      throw new Error(`Projected choices contains duplicate move ${JSON.stringify(wireSet.MoveName)}`);
    }
    if (wireSet.FieldName !== schema.fieldName || wireSet.Source !== schema.source) {
      throw new Error(`Projected choices for ${JSON.stringify(wireSet.MoveName)} does not match its generated field/source`);
    }
    const values = validateCandidateValues(wireSet, schema, options.playerPresentations);
    if (!wireSet.Candidates.some(candidate => candidate.Available)) {
      throw new Error(`Projected choices for ${JSON.stringify(wireSet.MoveName)} contains no available candidate`);
    }
    const move = wireSet.MoveName as keyof Projections & string;
    const candidates = wireSet.Candidates.map((candidate, index) => {
      const value = values[index] as Projections[typeof move]['value'];
      const input = { [schema.fieldName]: value } as Projections[typeof move]['input'];
      const action = options.action(move, input);
      hydrateMoveActionPreview(
        action as BoundMoveAction<string, object>,
        candidate.Available
          ? { kind: 'legal' }
          : {
            kind: 'illegal',
            reason: { code: 'preview-illegal', message: 'This choice is not available.' },
          },
      );
      return Object.freeze({
        id: candidateID(move, schema.fieldName, value),
        value,
        message: candidateMessage(move, schema, value, options.playerPresentations),
        available: candidate.Available,
        action,
      });
    });
    result.set(move, Object.freeze({
      move,
      field: schema.fieldName,
      message: options.messages?.[move] ?? defaultProjectedChoiceMessage(move, schema.fieldName),
      candidates: Object.freeze(candidates),
    }) as AnyProjectedSet<Projections>);
  }
  return ProjectedMoveChoices.ready(result);
}

function validateCandidateValues(
  wireSet: ProjectedMoveChoiceSetWire,
  schema: MoveChoiceProjectionSchemaEntry,
  players: readonly PlayerPresentation[],
): readonly (string | number)[] {
  const seen = new Set<string | number>();
  const rawValues = wireSet.Candidates.map(candidate => {
    if (seen.has(candidate.Value)) {
      throw new Error(`Projected choices for ${JSON.stringify(wireSet.MoveName)} duplicates candidate ${JSON.stringify(candidate.Value)}`);
    }
    seen.add(candidate.Value);
    return candidate.Value;
  });
  if (schema.source === 'enum-values') {
    const expected = schema.candidateValues ?? [];
    if (rawValues.some(value => typeof value !== 'string')
      || rawValues.length !== expected.length
      || rawValues.some(value => !expected.includes(value as string))) {
      throw new Error(`Projected enum candidates for ${JSON.stringify(wireSet.MoveName)} do not match the generated universe`);
    }
    return rawValues;
  }
  if (rawValues.length !== players.length) {
    throw new Error(`Projected player candidates for ${JSON.stringify(wireSet.MoveName)} do not match the player roster`);
  }
  return rawValues.map((value, index) => {
    if (!Number.isSafeInteger(value) || value !== index) {
      throw new Error(`Projected player candidate ${JSON.stringify(value)} is not canonical roster index ${index}`);
    }
    return index;
  });
}

function candidateID(move: string, field: string, value: string | number): string {
  return `${semanticToken(move)}:${semanticToken(field)}:${semanticToken(String(value))}`;
}

function candidateMessage(
  move: string,
  schema: MoveChoiceProjectionSchemaEntry,
  value: string | number,
  players: readonly PlayerPresentation[],
): MessageDescriptor {
  const defaultMessage = schema.source === 'players'
    ? players[value as number]?.label ?? `Player ${Number(value) + 1}`
    : humanize(String(value));
  return Object.freeze({
    id: `boardgame.projected-choices.${semanticToken(move)}.${semanticToken(schema.fieldName)}.candidate.${semanticToken(String(value))}`,
    defaultMessage,
  });
}

function semanticToken(value: string): string {
  return value.trim()
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/[^a-zA-Z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .toLowerCase() || 'choice';
}

function humanize(value: string): string {
  const spaced = value.replace(/([a-z0-9])([A-Z])/g, '$1 $2').replace(/[_-]+/g, ' ').trim();
  return spaced ? spaced[0].toUpperCase() + spaced.slice(1) : 'choice';
}
