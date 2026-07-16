import type { ReactiveController } from 'lit';
import { isBoundMoveAction, type BoundMoveAction } from './action.ts';
import { gameSnapshotKey, type GameSnapshotHost } from './snapshot-controller.ts';
import type { TargetKey } from './target-action.js';

const MAX_SELECTION_CANDIDATES = 4096;
const MAX_SELECTIONS = 1024;
const MAX_SELECTION_HISTORY = 1024;

export type SelectionDraftRebasePolicy = 'clear' | 'keep-valid';

export type SelectionDraftNotice<Key extends TargetKey> =
  | { readonly kind: 'cleared'; readonly message: string; readonly removed: readonly Key[] }
  | { readonly kind: 'pruned'; readonly message: string; readonly removed: readonly Key[] };

export interface SelectionDraftOptions<
  Key extends TargetKey,
  MoveName extends string,
  Input extends object,
> {
  readonly candidates: readonly Key[];
  readonly action: (selected: readonly Key[]) => BoundMoveAction<MoveName, Input>;
  readonly minSelected?: number;
  readonly maxSelected?: number;
  /** Clear by default; keeping still removes unavailable selections. */
  readonly rebase?: SelectionDraftRebasePolicy;
}

/** One candidate's snapshot-safe presentation and interaction binding. */
export interface SelectionOptionBinding<Key extends TargetKey> {
  readonly choice: Key;
  readonly selected: boolean;
  readonly capacityBlocked: boolean;
  toggle(): void;
}

export interface SelectionDraftBinding<
  Key extends TargetKey,
  MoveName extends string,
  Input extends object,
> {
  readonly candidates: readonly Key[];
  readonly selected: readonly Key[];
  readonly action: BoundMoveAction<MoveName, Input> | null;
  readonly notice: SelectionDraftNotice<Key> | null;
  readonly minimumSelected: number;
  readonly maximumSelected: number;
  readonly count: number;
  readonly minimumCount: number;
  readonly maximumCount: number;
  readonly status: string;
  readonly canUndo: boolean;
  readonly canClear: boolean;
  option(key: Key): SelectionOptionBinding<Key>;
  toggle(key: Key): void;
  select(key: Key): void;
  deselect(key: Key): void;
  clear(): void;
  undo(): void;
  dismissNotice(): void;
  isSelected(key: Key): boolean;
}

/** Snapshot-safe local multi-selection whose only commit is a typed move action. */
export class SelectionDraftController<Key extends TargetKey> implements ReactiveController {
  readonly #host: GameSnapshotHost;
  #stateObject: object | null | undefined;
  #snapshotKey = '';
  #selected: readonly Key[] = Object.freeze([]);
  #history: (readonly Key[])[] = [];
  #candidates = new Set<Key>();
  #maximumSelected = 1;
  #rebase: SelectionDraftRebasePolicy = 'clear';
  #notice: SelectionDraftNotice<Key> | null = null;

  constructor(host: GameSnapshotHost) {
    this.#host = host;
    host.addController(this);
  }

  bind<MoveName extends string, Input extends object>(
    options: SelectionDraftOptions<Key, MoveName, Input>,
  ): SelectionDraftBinding<Key, MoveName, Input> {
    const candidates = validateCandidates(options.candidates);
    const minimumSelected = validateBound('minSelected', options.minSelected ?? 1, 1, MAX_SELECTIONS);
    const maximumSelected = validateBound(
      'maxSelected', options.maxSelected ?? Math.max(minimumSelected, candidates.length),
      minimumSelected,
      MAX_SELECTIONS,
    );
    const rebase = options.rebase ?? 'clear';
    if (rebase !== 'clear' && rebase !== 'keep-valid') {
      throw new Error(`SelectionDraftController has unknown rebase policy ${JSON.stringify(rebase)}`);
    }
    const nextCandidates = new Set(candidates);
    const nextSnapshotKey = gameSnapshotKey(this.#host);
    const snapshotChanged = this.#snapshotKey !== ''
      && (this.#snapshotKey !== nextSnapshotKey || this.#stateObject !== this.#host.state);
    const availabilityChanged = this.#selected.some(key => !nextCandidates.has(key))
      || this.#selected.length > maximumSelected;
    this.#candidates = nextCandidates;
    this.#maximumSelected = maximumSelected;
    this.#rebase = rebase;
    if (snapshotChanged || availabilityChanged) this.#reconcile(snapshotChanged);
    this.#snapshotKey = nextSnapshotKey;
    this.#stateObject = this.#host.state;

    let action: BoundMoveAction<MoveName, Input> | null = null;
    if (this.#selected.length >= minimumSelected) {
      try {
        action = options.action(this.#selected);
      } catch (error) {
        const detail = error instanceof Error ? `: ${error.message}` : '';
        throw new Error(`SelectionDraftController action failed${detail}`);
      }
      if (!isBoundMoveAction(action)) {
        throw new Error('SelectionDraftController action must return a bound move action');
      }
    }
    return Object.freeze({
      candidates,
      selected: this.#selected,
      action,
      notice: this.#notice,
      minimumSelected,
      maximumSelected,
      count: this.#selected.length,
      minimumCount: minimumSelected,
      maximumCount: maximumSelected,
      status: `${this.#selected.length} selection${this.#selected.length === 1 ? '' : 's'} drafted.`,
      canUndo: this.#history.length > 0,
      canClear: this.#selected.length > 0,
      option: (key: Key): SelectionOptionBinding<Key> => {
        if (!nextCandidates.has(key)) {
          throw new Error(`SelectionDraftController cannot bind unknown candidate ${JSON.stringify(key)}`);
        }
        const selected = this.#selected.includes(key);
        return Object.freeze({
          choice: key,
          selected,
          capacityBlocked: !selected && this.#selected.length >= maximumSelected,
          toggle: () => this.toggle(key),
        });
      },
      toggle: this.toggle,
      select: this.select,
      deselect: this.deselect,
      clear: this.clear,
      undo: this.undo,
      dismissNotice: this.dismissNotice,
      isSelected: (key: Key) => this.#selected.includes(key),
    });
  }

  readonly toggle = (key: Key): void => {
    if (this.#selected.includes(key)) this.deselect(key);
    else this.select(key);
  };

  readonly select = (key: Key): void => {
    this.#requireCandidate(key);
    if (this.#selected.includes(key)) return;
    if (this.#selected.length >= this.#maximumSelected) {
      throw new Error(`SelectionDraftController cannot exceed ${this.#maximumSelected} selections`);
    }
    this.#remember();
    this.#selected = Object.freeze([...this.#selected, key]);
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly deselect = (key: Key): void => {
    this.#requireCandidate(key);
    if (!this.#selected.includes(key)) return;
    this.#remember();
    this.#selected = Object.freeze(this.#selected.filter(candidate => candidate !== key));
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly clear = (): void => {
    if (this.#selected.length === 0) return;
    this.#remember();
    this.#selected = Object.freeze([]);
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly undo = (): void => {
    const previous = this.#history.pop();
    if (!previous) return;
    this.#selected = previous;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly dismissNotice = (): void => {
    if (!this.#notice) return;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  hostDisconnected(): void {
    this.#selected = Object.freeze([]);
    this.#history = [];
    this.#notice = null;
  }

  #requireCandidate(key: Key): void {
    if (!this.#candidates.has(key)) {
      throw new Error(`SelectionDraftController cannot select unknown candidate ${JSON.stringify(key)}`);
    }
  }

  #remember(): void {
    this.#history.push(this.#selected);
    if (this.#history.length > MAX_SELECTION_HISTORY) this.#history.shift();
  }

  #reconcile(snapshotChanged: boolean): void {
    const previous = this.#selected;
    if (this.#rebase === 'clear') {
      if (previous.length > 0) {
        this.#notice = Object.freeze({
          kind: 'cleared',
          message: snapshotChanged
            ? 'The game changed, so the local draft was cleared.'
            : 'Available choices changed, so the local draft was cleared.',
          removed: previous,
        });
      }
      this.#selected = Object.freeze([]);
    } else {
      const kept = previous.filter(key => this.#candidates.has(key)).slice(0, this.#maximumSelected);
      const removed = previous.filter(key => !kept.includes(key));
      this.#selected = Object.freeze(kept);
      if (removed.length > 0) {
        this.#notice = Object.freeze({
          kind: 'pruned',
          message: snapshotChanged
            ? 'The game changed; unavailable selections were removed.'
            : 'Unavailable selections were removed from the draft.',
          removed: Object.freeze(removed),
        });
      }
    }
    this.#history = [];
  }
}

function validateCandidates<Key extends TargetKey>(candidates: readonly Key[]): readonly Key[] {
  if (!Array.isArray(candidates)) throw new Error('SelectionDraftController candidates must be an array');
  if (candidates.length > MAX_SELECTION_CANDIDATES) {
    throw new Error(`SelectionDraftController has ${candidates.length} candidates; maximum is ${MAX_SELECTION_CANDIDATES}`);
  }
  const seen = new Set<TargetKey>();
  const copy = [...candidates];
  copy.forEach((key, index) => {
    if ((typeof key !== 'string' && typeof key !== 'number')
      || (typeof key === 'number' && !Number.isFinite(key))) {
      throw new Error(`SelectionDraftController candidate at index ${index} must be a finite string or number`);
    }
    if (seen.has(key)) throw new Error(`SelectionDraftController candidate ${JSON.stringify(key)} is duplicated`);
    seen.add(key);
  });
  return Object.freeze(copy);
}

function validateBound(name: string, value: number, minimum: number, maximum: number): number {
  if (!Number.isSafeInteger(value) || value < minimum || value > maximum) {
    throw new Error(`SelectionDraftController ${name} must be an integer from ${minimum} through ${maximum}`);
  }
  return value;
}
