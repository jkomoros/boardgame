import type { ReactiveController } from 'lit';
import type { TargetAction, TargetKey } from './target-action.js';
import { gameSnapshotKey, type GameSnapshotHost } from './snapshot-controller.ts';

// Keep source selection bounded independently of the destination collection.
// A type-only dependency also lets Node execute this controller's TypeScript
// unit tests from a clean checkout without relying on stale emitted .js files.
const MAX_SOURCE_DESTINATION_SOURCES = 1024;

export interface SourceDestinationBinding<
  Key extends TargetKey,
  MoveName extends string = string,
  Input extends object = object,
> {
  readonly selectedSource: Key | null;
  readonly sources: readonly Key[];
  readonly action: TargetAction<Key, MoveName, Input> | null;
  selectSource(key: Key): void;
  clear(): void;
}

export interface SourceDestinationOptions<
  Key extends TargetKey,
  MoveName extends string,
  Input extends object,
> {
  readonly sources: readonly Key[];
  readonly destinations: (source: Key) => TargetAction<Key, MoveName, Input>;
}

export type SourceDestinationHost = GameSnapshotHost;

/**
 * Owns the small piece of local state in a source-then-destination interaction.
 * TargetAction remains responsible for legality, preview, staleness, and submission.
 */
export class SourceDestinationController<Key extends TargetKey> implements ReactiveController {
  readonly #host: SourceDestinationHost;
  #state: object | null | undefined;
  #snapshotKey = '';
  #sourceSet = new Set<Key>();
  #selectedSource: Key | null = null;

  constructor(host: SourceDestinationHost) {
    this.#host = host;
    host.addController(this);
  }

  bind<MoveName extends string, Input extends object>(
    options: SourceDestinationOptions<Key, MoveName, Input>,
  ): SourceDestinationBinding<Key, MoveName, Input> {
    const sources = validateSources(options.sources);
    const snapshotKey = gameSnapshotKey(this.#host);
    if (this.#state !== this.#host.state || this.#snapshotKey !== snapshotKey) {
      this.#state = this.#host.state;
      this.#snapshotKey = snapshotKey;
      this.#selectedSource = null;
    }
    this.#sourceSet = new Set(sources);
    if (this.#selectedSource !== null && !this.#sourceSet.has(this.#selectedSource)) {
      this.#selectedSource = null;
    }

    let action: TargetAction<Key, MoveName, Input> | null = null;
    if (this.#selectedSource !== null) {
      try {
        action = options.destinations(this.#selectedSource);
      } catch (error) {
        const detail = error instanceof Error ? `: ${error.message}` : '';
        throw new Error(`SourceDestinationController destinations failed for ${JSON.stringify(this.#selectedSource)}${detail}`);
      }
      for (const candidate of action.candidates) {
        if (this.#sourceSet.has(candidate.key)) {
          throw new Error(`SourceDestinationController key ${JSON.stringify(candidate.key)} is ambiguously both a source and destination`);
        }
      }
    }

    return Object.freeze({
      selectedSource: this.#selectedSource,
      sources,
      action,
      selectSource: this.selectSource,
      clear: this.clear,
    });
  }

  readonly selectSource = (key: Key): void => {
    if (!this.#sourceSet.has(key)) {
      throw new Error(`SourceDestinationController cannot select unknown source ${JSON.stringify(key)}`);
    }
    this.#selectedSource = Object.is(this.#selectedSource, key) ? null : key;
    this.#host.requestUpdate();
  };

  readonly clear = (): void => {
    if (this.#selectedSource === null) return;
    this.#selectedSource = null;
    this.#host.requestUpdate();
  };

  hostDisconnected(): void {
    this.#selectedSource = null;
  }
}

function validateSources<Key extends TargetKey>(sources: readonly Key[]): readonly Key[] {
  if (!Array.isArray(sources)) throw new Error('SourceDestinationController sources must be an array');
  const copy = [...sources];
  if (copy.length > MAX_SOURCE_DESTINATION_SOURCES) {
    throw new Error(`SourceDestinationController has ${copy.length} sources; maximum is ${MAX_SOURCE_DESTINATION_SOURCES}`);
  }
  const seen = new Set<TargetKey>();
  copy.forEach((key, index) => {
    if (typeof key !== 'string' && typeof key !== 'number') {
      throw new Error(`SourceDestinationController source at index ${index} must be a string or number`);
    }
    if (typeof key === 'number' && !Number.isFinite(key)) {
      throw new Error(`SourceDestinationController source at index ${index} must be finite`);
    }
    if (seen.has(key)) {
      throw new Error(`SourceDestinationController source ${JSON.stringify(key)} is duplicated`);
    }
    seen.add(key);
  });
  return Object.freeze(copy);
}
