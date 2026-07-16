import {
  MoveSubmissionGate,
  type MovePreviewTransport,
  type MoveSubmissionRequest,
} from '../moves/action.js';
import type { TargetPreviewTransport } from '../moves/target-action.js';
import {
  validatePlayerPresentations,
  type PlayerPresentation,
} from '../status/player-presentation.js';

export const RENDERER_FIXTURE_SCHEMA_VERSION = 1 as const;

export type RendererFixtureSurface = 'game' | 'table' | 'hand';

export interface RendererFixtureLegality {
  readonly legalForPlayer: boolean;
  readonly legalForAnyone: boolean;
  readonly error?: string;
}

export interface RendererFixtureGameContract {
  readonly State: object;
  readonly MoveName: string;
  readonly RendererTag: `boardgame-render-game-${string}`;
}

export interface RendererFixtureSnapshot<Contract extends RendererFixtureGameContract> {
  readonly schemaVersion: typeof RENDERER_FIXTURE_SCHEMA_VERSION;
  readonly state: Contract['State'];
  readonly viewingAsPlayer: number;
  readonly currentPlayerIndex: number;
  readonly moveLegality: Readonly<Record<Contract['MoveName'], RendererFixtureLegality>>;
  readonly version: number;
  readonly outcome: {
    readonly finished: boolean;
    readonly winners: readonly number[];
  };
  readonly surface: RendererFixtureSurface;
  readonly serverMoveInputSchemaFingerprint: string;
  readonly previewDisabledSpaces?: readonly number[];
  /** Explicit identities available through renderer.playerPresentation(index). */
  readonly playerPresentations?: readonly PlayerPresentation[];
}

export interface RendererFixtureDefinition<Contract extends RendererFixtureGameContract> {
  readonly tagName: Contract['RendererTag'];
  readonly snapshot: RendererFixtureSnapshot<Contract>;
}

export function defineRendererFixture<Contract extends RendererFixtureGameContract>(
  definition: RendererFixtureDefinition<Contract>,
): RendererFixtureDefinition<Contract> {
  return Object.freeze(definition);
}

export interface RendererFixtureProposal<MoveName extends string = string> {
  readonly requestID: string;
  readonly snapshotVersion: number;
  readonly name: MoveName;
  readonly arguments: Readonly<Record<string, string>>;
}

interface RendererFixtureTarget<State extends object> extends HTMLElement {
  state: State | null;
  viewingAsPlayer: number;
  currentPlayerIndex: number;
  moveLegality: Record<string, RendererFixtureLegality>;
  gameFinished: boolean;
  gameWinners: number[];
  serverMoveInputSchemaFingerprint: string | null;
  previewDisabledSpaces: number[];
  playerPresentations: readonly PlayerPresentation[];
  playerPresentation(playerIndex: number): PlayerPresentation;
  gameName: string;
  gameId: string;
  gameVersion: number;
  snapshotEpoch: number;
  proposingAsPlayer: number;
  proposingAsAdmin: boolean;
  moveTransport: { submit(request: MoveSubmissionRequest): Promise<{ readonly kind: 'success' }> };
  movePreviewTransport: MovePreviewTransport;
  targetPreviewTransport: TargetPreviewTransport;
  moveSubmissionGate: MoveSubmissionGate;
  readonly updateComplete: Promise<unknown>;
}

interface ProposalEventDetail<MoveName extends string> {
  readonly name: MoveName;
  readonly arguments: Readonly<Record<string, string>>;
}

export class RendererFixtureHandle<Contract extends RendererFixtureGameContract> {
  readonly host: HTMLElement;
  readonly renderer: RendererFixtureTarget<Contract['State']>;

  readonly #proposals: RendererFixtureProposal<Contract['MoveName']>[] = [];
  #snapshot: RendererFixtureSnapshot<Contract>;
  #sequence = 0;
  readonly #submissionGate = new MoveSubmissionGate();
  readonly #proposalListener: EventListener;

  constructor(
    host: HTMLElement,
    renderer: RendererFixtureTarget<Contract['State']>,
    snapshot: RendererFixtureSnapshot<Contract>,
  ) {
    this.host = host;
    this.renderer = renderer;
    this.#snapshot = snapshot;
    this.#proposalListener = (event: Event): void => {
      if (!(event instanceof CustomEvent) || !isProposalDetail(event.detail)) {
        throw new Error('Renderer fixture received malformed propose-move detail');
      }
      if (!Object.prototype.hasOwnProperty.call(this.#snapshot.moveLegality, event.detail.name)) {
        throw new Error(`Renderer fixture received unknown move proposal: ${event.detail.name}`);
      }
      this.recordProposal({
        requestID: `fixture-v${this.#snapshot.version}-request-${++this.#sequence}`,
        snapshotVersion: this.#snapshot.version,
        viewingAsPlayer: this.#snapshot.viewingAsPlayer,
        proposingAsPlayer: this.#snapshot.viewingAsPlayer,
        proposingAsAdmin: this.#snapshot.viewingAsPlayer === -2,
        name: event.detail.name,
        arguments: event.detail.arguments,
      });
    };
    renderer.moveTransport = {
      submit: async request => {
        this.recordProposal(request);
        return { kind: 'success' };
      },
    };
    renderer.movePreviewTransport = {
      preview: async request => {
        const legality = this.#snapshot.moveLegality[request.name as Contract['MoveName']];
        return {
          kind: 'success',
          legal: legality?.legalForPlayer ?? false,
          ...(legality?.error ? { error: legality.error } : {}),
        };
      },
    };
    renderer.targetPreviewTransport = {
      previewTargets: async request => {
        const legality = this.#snapshot.moveLegality[request.name as Contract['MoveName']];
        const disabled = new Set(this.#snapshot.previewDisabledSpaces ?? []);
        return {
          kind: 'success',
          results: request.candidates.map((candidate, index) => ({
            id: candidate.id,
            legal: (legality?.legalForPlayer ?? false) && !disabled.has(index),
            ...((legality?.error || disabled.has(index))
              ? { error: legality?.error ?? 'This target is occupied' }
              : {}),
          })),
        };
      },
    };
    renderer.moveSubmissionGate = this.#submissionGate;
    this.install(snapshot);
    renderer.addEventListener('propose-move', this.#proposalListener);
  }

  get proposals(): readonly RendererFixtureProposal<Contract['MoveName']>[] {
    return Object.freeze([...this.#proposals]);
  }

  async update(snapshot: RendererFixtureSnapshot<Contract>): Promise<void> {
    this.install(snapshot);
    await this.renderer.updateComplete;
  }

  dispose(): void {
    this.renderer.removeEventListener('propose-move', this.#proposalListener);
    this.host.remove();
  }

  private install(snapshot: RendererFixtureSnapshot<Contract>): void {
    validateSnapshot(snapshot);
    this.#snapshot = snapshot;
    this.host.dataset['fixtureSchemaVersion'] = String(snapshot.schemaVersion);
    this.host.dataset['fixtureVersion'] = String(snapshot.version);
    this.host.dataset['fixtureSurface'] = snapshot.surface;
    this.renderer.state = snapshot.state;
    this.renderer.viewingAsPlayer = snapshot.viewingAsPlayer;
    this.renderer.currentPlayerIndex = snapshot.currentPlayerIndex;
    this.renderer.moveLegality = cloneLegality(snapshot.moveLegality);
    this.renderer.gameFinished = snapshot.outcome.finished;
    this.renderer.gameWinners = [...snapshot.outcome.winners];
    this.renderer.serverMoveInputSchemaFingerprint = snapshot.serverMoveInputSchemaFingerprint;
    this.renderer.previewDisabledSpaces = [...(snapshot.previewDisabledSpaces ?? [])];
    this.renderer.playerPresentations = validatePlayerPresentations(snapshot.playerPresentations ?? []);
    this.renderer.gameName = fixtureGameName(this.host.dataset['rendererFixture'] ?? '');
    this.renderer.gameId = 'fixture';
    this.renderer.gameVersion = snapshot.version;
    this.renderer.snapshotEpoch = snapshot.version;
    this.renderer.proposingAsPlayer = snapshot.viewingAsPlayer;
    this.renderer.proposingAsAdmin = snapshot.viewingAsPlayer === -2;
  }

  private recordProposal(request: MoveSubmissionRequest): void {
    if (!Object.prototype.hasOwnProperty.call(this.#snapshot.moveLegality, request.name)) {
      throw new Error(`Renderer fixture received unknown move proposal: ${request.name}`);
    }
    this.#proposals.push(Object.freeze({
      requestID: request.requestID,
      snapshotVersion: request.snapshotVersion,
      name: request.name as Contract['MoveName'],
      arguments: Object.freeze({ ...request.arguments }),
    }));
  }
}

export async function mountRendererFixture<Contract extends RendererFixtureGameContract>(
  definition: RendererFixtureDefinition<Contract>,
  parent: ParentNode = document.body,
): Promise<RendererFixtureHandle<Contract>> {
  const { tagName, snapshot } = definition;
  if (!tagName.includes('-')) {
    throw new Error(`Renderer fixture tag must be a custom-element name; received ${tagName}`);
  }
  validateSnapshot(snapshot);
  validateSurfaceTag(tagName, snapshot.surface);
  if (!customElements.get(tagName)) {
    throw new Error(`${tagName} is not registered; import the renderer module before mounting it`);
  }
  const renderer = document.createElement(tagName);
  if (!isRendererFixtureTarget<Contract['State']>(renderer)) {
    throw new Error(`${tagName} does not expose the boardgame renderer fixture contract`);
  }
  const host = document.createElement('section');
  host.dataset['rendererFixture'] = tagName;
  host.append(renderer);
  parent.append(host);
  let handle: RendererFixtureHandle<Contract> | undefined;
  try {
    handle = new RendererFixtureHandle(host, renderer, snapshot);
    await renderer.updateComplete;
    return handle;
  } catch (error) {
    if (handle) handle.dispose();
    else host.remove();
    throw error;
  }
}

function validateSnapshot<Contract extends RendererFixtureGameContract>(
  snapshot: RendererFixtureSnapshot<Contract>,
): void {
  if (snapshot.schemaVersion !== RENDERER_FIXTURE_SCHEMA_VERSION) {
    throw new Error(`Unsupported renderer fixture schema version: ${snapshot.schemaVersion}`);
  }
  if (!isRecord(snapshot.state)) {
    throw new Error('Renderer fixture state must be a non-null object');
  }
  for (const [label, value] of [
    ['viewingAsPlayer', snapshot.viewingAsPlayer],
    ['currentPlayerIndex', snapshot.currentPlayerIndex],
    ['version', snapshot.version],
  ] as const) {
    if (!Number.isSafeInteger(value)) {
      throw new Error(`Renderer fixture ${label} must be a safe integer`);
    }
  }
  if (snapshot.version < 0) {
    throw new Error('Renderer fixture version must be non-negative');
  }
  if (!['game', 'table', 'hand'].includes(snapshot.surface)) {
    throw new Error(`Unsupported renderer fixture surface: ${snapshot.surface}`);
  }
  if (!/^sha256:[0-9a-f]{64}$/.test(snapshot.serverMoveInputSchemaFingerprint)) {
    throw new Error('Renderer fixture requires the generated move-input schema fingerprint');
  }
  if (!isRecord(snapshot.moveLegality)) {
    throw new Error('Renderer fixture moveLegality must be a record');
  }
  for (const [name, legality] of Object.entries(snapshot.moveLegality)) {
    if (!name || !isRecord(legality) || typeof legality['legalForPlayer'] !== 'boolean'
      || typeof legality['legalForAnyone'] !== 'boolean'
      || (legality['error'] !== undefined && typeof legality['error'] !== 'string')) {
      throw new Error(`Renderer fixture has malformed legality for move ${name || '<empty>'}`);
    }
    if (legality['legalForPlayer'] && !legality['legalForAnyone']) {
      throw new Error(`Renderer fixture legality is contradictory for move ${name}`);
    }
  }
  if (!isRecord(snapshot.outcome)
    || typeof snapshot.outcome['finished'] !== 'boolean'
    || !Array.isArray(snapshot.outcome['winners'])) {
    throw new Error('Renderer fixture outcome must contain finished and winners');
  }
  for (const winner of snapshot.outcome.winners) {
    if (!Number.isSafeInteger(winner)) {
      throw new Error('Renderer fixture winner indexes must be safe integers');
    }
  }
  if (snapshot.previewDisabledSpaces !== undefined
    && !Array.isArray(snapshot.previewDisabledSpaces)) {
    throw new Error('Renderer fixture previewDisabledSpaces must be an array');
  }
  for (const space of snapshot.previewDisabledSpaces ?? []) {
    if (!Number.isSafeInteger(space) || space < 0) {
      throw new Error('Renderer fixture disabled-space indexes must be non-negative safe integers');
    }
  }
  validatePlayerPresentations(snapshot.playerPresentations ?? []);
}

function validateSurfaceTag(tagName: string, surface: RendererFixtureSurface): void {
  const suffix = tagName.endsWith('-table') ? 'table'
    : tagName.endsWith('-hand') ? 'hand'
      : 'game';
  if (suffix !== surface) {
    throw new Error(`Renderer fixture surface ${surface} does not match renderer tag ${tagName}`);
  }
}

function cloneLegality(
  source: Readonly<Record<string, RendererFixtureLegality>>,
): Record<string, RendererFixtureLegality> {
  return Object.fromEntries(Object.entries(source).map(([name, legality]) => [
    name,
    legality.error === undefined
      ? { legalForPlayer: legality.legalForPlayer, legalForAnyone: legality.legalForAnyone }
      : { ...legality },
  ]));
}

function isRendererFixtureTarget<State extends object>(
  element: HTMLElement,
): element is RendererFixtureTarget<State> {
  const candidate: Partial<RendererFixtureTarget<State>> = element;
  return 'state' in candidate
    && 'viewingAsPlayer' in candidate
    && 'currentPlayerIndex' in candidate
    && 'moveLegality' in candidate
    && 'gameFinished' in candidate
    && 'gameWinners' in candidate
    && 'serverMoveInputSchemaFingerprint' in candidate
    && 'previewDisabledSpaces' in candidate
    && 'playerPresentations' in candidate
    && typeof candidate.playerPresentation === 'function'
    && candidate.updateComplete instanceof Promise;
}

function isProposalDetail(value: unknown): value is ProposalEventDetail<string> {
  if (!isRecord(value) || typeof value['name'] !== 'string' || !isRecord(value['arguments'])) {
    return false;
  }
  return Object.values(value['arguments']).every((argument) => typeof argument === 'string');
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function fixtureGameName(tagName: string): string {
  return tagName.replace(/^boardgame-render-game-/, '').replace(/-(?:table|hand)$/, '');
}
