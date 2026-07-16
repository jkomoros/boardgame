import type { ReactiveController } from 'lit';
import { isBoundMoveAction, type BoundMoveAction } from './action.ts';
import { gameSnapshotKey, type GameSnapshotHost } from './snapshot-controller.ts';
import type { TargetKey } from './target-action.js';

const MAX_DRAFT_ITEMS = 1024;
const MAX_DRAFT_TARGETS = 4096;
const MAX_DRAFT_HISTORY = 1024;

export type PlacementDraftRebasePolicy = 'clear' | 'keep-valid';

export interface DraftPlacement<Item extends TargetKey, Target extends TargetKey> {
  readonly item: Item;
  readonly target: Target;
}

export type PlacementDraftNotice<Item extends TargetKey, Target extends TargetKey> =
  | {
    readonly kind: 'cleared';
    readonly message: string;
    readonly removed: readonly DraftPlacement<Item, Target>[];
  }
  | {
    readonly kind: 'pruned';
    readonly message: string;
    readonly removed: readonly DraftPlacement<Item, Target>[];
  };

export interface PlacementDraftOptions<
  Item extends TargetKey,
  Target extends TargetKey,
  MoveName extends string,
  Input extends object,
> {
  /** Stable item identities currently available to place. */
  readonly items: readonly Item[];
  /** Stable destination identities currently available. */
  readonly targets: readonly Target[];
  /** Builds the one exact, typed move submitted for the complete local draft. */
  readonly action: (
    placements: readonly DraftPlacement<Item, Target>[],
  ) => BoundMoveAction<MoveName, Input>;
  readonly minPlacements?: number;
  readonly maxPlacements?: number;
  /** Clear by default; keeping still removes placements invalid in the new snapshot. */
  readonly rebase?: PlacementDraftRebasePolicy;
}

/** One placeable item's immutable presentation and interaction binding. */
export interface PlacementItemBinding<Item extends TargetKey, Target extends TargetKey> {
  readonly item: Item;
  readonly selected: boolean;
  readonly placedAt: Target | null;
  readonly capacityBlocked: boolean;
  select(): void;
  remove(): void;
}

/** One destination's immutable presentation and interaction binding. */
export interface PlacementTargetBinding<Item extends TargetKey, Target extends TargetKey> {
  readonly target: Target;
  readonly occupiedBy: Item | null;
  readonly canPlace: boolean;
  readonly reason: string | null;
  place(): void;
}

export interface PlacementDraftBinding<
  Item extends TargetKey,
  Target extends TargetKey,
  MoveName extends string,
  Input extends object,
> {
  readonly items: readonly Item[];
  readonly targets: readonly Target[];
  readonly placements: readonly DraftPlacement<Item, Target>[];
  readonly selectedItem: Item | null;
  readonly action: BoundMoveAction<MoveName, Input> | null;
  readonly notice: PlacementDraftNotice<Item, Target> | null;
  readonly minimumPlacements: number;
  readonly maximumPlacements: number;
  /** Common surface consumed by boardgame-draft-controls. */
  readonly count: number;
  readonly minimumCount: number;
  readonly maximumCount: number;
  readonly status: string;
  readonly canUndo: boolean;
  readonly canClear: boolean;
  item(item: Item): PlacementItemBinding<Item, Target>;
  target(target: Target): PlacementTargetBinding<Item, Target>;
  selectItem(item: Item): void;
  /** Keyboard/click workflow: place the selected item on this target. */
  place(target: Target): void;
  /** Pointer workflow: uses the same checked mutation as selectItem + place. */
  assign(item: Item, target: Target): void;
  removeItem(item: Item): void;
  undo(): void;
  clear(): void;
  dismissNotice(): void;
  targetFor(item: Item): Target | undefined;
  itemAt(target: Target): Item | undefined;
}

interface DraftState<Item extends TargetKey, Target extends TargetKey> {
  readonly placements: readonly DraftPlacement<Item, Target>[];
  readonly selectedItem: Item | null;
}

/**
 * Local, immutable item-to-target drafting for word tiles and similar turns.
 * The resulting action remains the ordinary snapshot-bound move action, so the
 * controller cannot weaken legality, the submission gate, or ExpectedVersion.
 */
export class PlacementDraftController<Item extends TargetKey, Target extends TargetKey>
implements ReactiveController {
  readonly #host: GameSnapshotHost;
  #stateObject: object | null | undefined;
  #snapshotKey = '';
  #placements: readonly DraftPlacement<Item, Target>[] = Object.freeze([]);
  #selectedItem: Item | null = null;
  #history: DraftState<Item, Target>[] = [];
  #items = new Set<Item>();
  #targets = new Set<Target>();
  #maxPlacements = 1;
  #rebase: PlacementDraftRebasePolicy = 'clear';
  #notice: PlacementDraftNotice<Item, Target> | null = null;

  constructor(host: GameSnapshotHost) {
    this.#host = host;
    host.addController(this);
  }

  bind<MoveName extends string, Input extends object>(
    options: PlacementDraftOptions<Item, Target, MoveName, Input>,
  ): PlacementDraftBinding<Item, Target, MoveName, Input> {
    const items = validateKeys('item', options.items, MAX_DRAFT_ITEMS);
    const targets = validateKeys('target', options.targets, MAX_DRAFT_TARGETS);
    const minPlacements = validateBound('minPlacements', options.minPlacements ?? 1, 1, MAX_DRAFT_ITEMS);
    const maximum = Math.min(items.length, targets.length, MAX_DRAFT_ITEMS);
    const maxPlacements = validateBound(
      'maxPlacements', options.maxPlacements ?? Math.max(minPlacements, maximum), minPlacements, MAX_DRAFT_ITEMS,
    );
    const rebase = options.rebase ?? 'clear';
    if (rebase !== 'clear' && rebase !== 'keep-valid') {
      throw new Error(`PlacementDraftController has unknown rebase policy ${JSON.stringify(rebase)}`);
    }

    const nextItems = new Set(items);
    const nextTargets = new Set(targets);
    const nextSnapshotKey = gameSnapshotKey(this.#host);
    const snapshotChanged = this.#snapshotKey !== ''
      && (this.#snapshotKey !== nextSnapshotKey || this.#stateObject !== this.#host.state);
    const availabilityChanged = this.#placements.some(
      placement => !nextItems.has(placement.item) || !nextTargets.has(placement.target),
    ) || this.#placements.length > maxPlacements
      || (this.#selectedItem !== null && !nextItems.has(this.#selectedItem));

    this.#items = nextItems;
    this.#targets = nextTargets;
    this.#maxPlacements = maxPlacements;
    this.#rebase = rebase;
    if (snapshotChanged || availabilityChanged) this.#reconcile(snapshotChanged);
    this.#snapshotKey = nextSnapshotKey;
    this.#stateObject = this.#host.state;

    let action: BoundMoveAction<MoveName, Input> | null = null;
    if (this.#placements.length >= minPlacements) {
      try {
        action = options.action(this.#placements);
      } catch (error) {
        const detail = error instanceof Error ? `: ${error.message}` : '';
        throw new Error(`PlacementDraftController action failed${detail}`);
      }
      if (!isBoundMoveAction(action)) {
        throw new Error('PlacementDraftController action must return a bound move action');
      }
    }

    const byItem = new Map(this.#placements.map(placement => [placement.item, placement.target] as const));
    const byTarget = new Map(this.#placements.map(placement => [placement.target, placement.item] as const));
    return Object.freeze({
      items,
      targets,
      placements: this.#placements,
      selectedItem: this.#selectedItem,
      action,
      notice: this.#notice,
      minimumPlacements: minPlacements,
      maximumPlacements: maxPlacements,
      count: this.#placements.length,
      minimumCount: minPlacements,
      maximumCount: maxPlacements,
      status: this.#selectedItem !== null
        ? 'Item selected. Choose a destination.'
        : `${this.#placements.length} placement${this.#placements.length === 1 ? '' : 's'} drafted.`,
      canUndo: this.#history.length > 0,
      canClear: this.#placements.length > 0 || this.#selectedItem !== null,
      item: (item: Item): PlacementItemBinding<Item, Target> => {
        if (!nextItems.has(item)) {
          throw new Error(`PlacementDraftController cannot bind unknown item ${JSON.stringify(item)}`);
        }
        const placedAt = byItem.get(item) ?? null;
        const selected = this.#selectedItem === item;
        return Object.freeze({
          item,
          selected,
          placedAt,
          capacityBlocked: !selected && placedAt === null && this.#placements.length >= maxPlacements,
          select: () => this.selectItem(item),
          remove: () => this.removeItem(item),
        });
      },
      target: (target: Target): PlacementTargetBinding<Item, Target> => {
        if (!nextTargets.has(target)) {
          throw new Error(`PlacementDraftController cannot bind unknown target ${JSON.stringify(target)}`);
        }
        const occupiedBy = byTarget.get(target) ?? null;
        const selected = this.#selectedItem;
        const selectedIsPlaced = selected !== null && byItem.has(selected);
        const atCapacity = !selectedIsPlaced && this.#placements.length >= maxPlacements;
        const canPlace = selected !== null && !atCapacity
          && (occupiedBy === null || occupiedBy === selected);
        return Object.freeze({
          target,
          occupiedBy,
          canPlace,
          reason: selected === null
            ? 'Select an item first'
            : occupiedBy !== null && occupiedBy !== selected
              ? 'Destination is occupied'
              : atCapacity ? 'Maximum placements reached' : null,
          place: () => this.place(target),
        });
      },
      selectItem: this.selectItem,
      place: this.place,
      assign: this.assign,
      removeItem: this.removeItem,
      undo: this.undo,
      clear: this.clear,
      dismissNotice: this.dismissNotice,
      targetFor: (item: Item) => byItem.get(item),
      itemAt: (target: Target) => byTarget.get(target),
    });
  }

  readonly selectItem = (item: Item): void => {
    this.#requireItem(item);
    this.#selectedItem = this.#selectedItem === item ? null : item;
    this.#host.requestUpdate();
  };

  readonly place = (target: Target): void => {
    if (this.#selectedItem === null) {
      throw new Error('PlacementDraftController cannot place a target before selecting an item');
    }
    this.assign(this.#selectedItem, target);
  };

  readonly assign = (item: Item, target: Target): void => {
    this.#requireItem(item);
    this.#requireTarget(target);
    const occupied = this.#placements.find(placement => placement.target === target);
    if (occupied && occupied.item !== item) {
      throw new Error(`PlacementDraftController target ${JSON.stringify(target)} is already occupied`);
    }
    const existing = this.#placements.findIndex(placement => placement.item === item);
    if (existing >= 0 && this.#placements[existing]!.target === target) {
      if (this.#selectedItem !== null) {
        this.#selectedItem = null;
        this.#host.requestUpdate();
      }
      return;
    }
    if (existing < 0 && this.#placements.length >= this.#maxPlacements) {
      throw new Error(`PlacementDraftController cannot exceed ${this.#maxPlacements} placements`);
    }
    this.#remember();
    const next = [...this.#placements];
    const placement = freezePlacement(item, target);
    if (existing >= 0) next[existing] = placement;
    else next.push(placement);
    this.#placements = Object.freeze(next);
    this.#selectedItem = null;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly removeItem = (item: Item): void => {
    this.#requireItem(item);
    const index = this.#placements.findIndex(placement => placement.item === item);
    if (index < 0) return;
    this.#remember();
    this.#placements = Object.freeze(this.#placements.filter((_, candidate) => candidate !== index));
    if (this.#selectedItem === item) this.#selectedItem = null;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly undo = (): void => {
    const previous = this.#history.pop();
    if (!previous) return;
    this.#placements = previous.placements;
    this.#selectedItem = previous.selectedItem;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly clear = (): void => {
    if (this.#placements.length === 0 && this.#selectedItem === null) return;
    if (this.#placements.length > 0) this.#remember();
    this.#placements = Object.freeze([]);
    this.#selectedItem = null;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  readonly dismissNotice = (): void => {
    if (!this.#notice) return;
    this.#notice = null;
    this.#host.requestUpdate();
  };

  hostDisconnected(): void {
    this.#placements = Object.freeze([]);
    this.#selectedItem = null;
    this.#history = [];
    this.#notice = null;
  }

  #reconcile(snapshotChanged: boolean): void {
    const previous = this.#placements;
    if (this.#rebase === 'clear') {
      if (previous.length > 0 || this.#selectedItem !== null) {
        this.#notice = Object.freeze({
          kind: 'cleared',
          message: snapshotChanged
            ? 'The game changed, so the local draft was cleared.'
            : 'Available items or targets changed, so the local draft was cleared.',
          removed: previous,
        });
      }
      this.#placements = Object.freeze([]);
      this.#selectedItem = null;
    } else {
      const selectedWasRemoved = this.#selectedItem !== null && !this.#items.has(this.#selectedItem);
      const kept = previous.filter(
        placement => this.#items.has(placement.item) && this.#targets.has(placement.target),
      ).slice(0, this.#maxPlacements);
      const removed = previous.filter(placement => !kept.includes(placement));
      this.#placements = Object.freeze(kept);
      if (this.#selectedItem !== null && !this.#items.has(this.#selectedItem)) this.#selectedItem = null;
      if (removed.length > 0 || selectedWasRemoved) {
        this.#notice = Object.freeze({
          kind: 'pruned',
          message: snapshotChanged
            ? 'The game changed, so unavailable placements were removed from the local draft.'
            : 'Available items, targets, or limits changed, so invalid placements were removed from the local draft.',
          removed: Object.freeze(removed),
        });
      }
    }
    this.#history = [];
  }

  #remember(): void {
    this.#history.push(Object.freeze({
      placements: this.#placements,
      selectedItem: this.#selectedItem,
    }));
    if (this.#history.length > MAX_DRAFT_HISTORY) this.#history.shift();
  }

  #requireItem(item: Item): void {
    if (!this.#items.has(item)) {
      throw new Error(`PlacementDraftController cannot use unknown item ${JSON.stringify(item)}`);
    }
  }

  #requireTarget(target: Target): void {
    if (!this.#targets.has(target)) {
      throw new Error(`PlacementDraftController cannot use unknown target ${JSON.stringify(target)}`);
    }
  }
}

function validateKeys<Key extends TargetKey>(
  kind: 'item' | 'target', values: readonly Key[], maximum: number,
): readonly Key[] {
  if (!Array.isArray(values)) throw new Error(`PlacementDraftController ${kind}s must be an array`);
  if (values.length > maximum) {
    throw new Error(`PlacementDraftController has ${values.length} ${kind}s; maximum is ${maximum}`);
  }
  const seen = new Set<TargetKey>();
  const copy = [...values];
  copy.forEach((key, index) => {
    if ((typeof key !== 'string' && typeof key !== 'number')
      || (typeof key === 'number' && !Number.isFinite(key))) {
      throw new Error(`PlacementDraftController ${kind} at index ${index} must be a string or finite number`);
    }
    if (seen.has(key)) throw new Error(`PlacementDraftController ${kind} ${JSON.stringify(key)} is duplicated`);
    seen.add(key);
  });
  return Object.freeze(copy);
}

function validateBound(name: string, value: number, minimum: number, maximum: number): number {
  if (!Number.isSafeInteger(value) || value < minimum || value > maximum) {
    throw new Error(`PlacementDraftController ${name} must be a safe integer from ${minimum} through ${maximum}`);
  }
  return value;
}

function freezePlacement<Item extends TargetKey, Target extends TargetKey>(
  item: Item, target: Target,
): DraftPlacement<Item, Target> {
  return Object.freeze({ item, target });
}
