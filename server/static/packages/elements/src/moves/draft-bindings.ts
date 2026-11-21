/**
 * The binding shapes the draft controllers hand to elements: what a stack,
 * zone, board, selection option or placement item needs in order to show and
 * drive a local selection or placement draft.
 *
 * These are layer-1 types (see scripts/layers.json) because layer-1 elements
 * take them as properties. The controllers that produce them
 * (selection-draft, placement-draft, source-destination) are renderer-kit
 * code and re-export these names.
 */
import type { TargetAction, TargetKey } from './target-action.js';

// ---- Selection drafts ----

export type SelectionDraftNotice<Key extends TargetKey> =
  | { readonly kind: 'cleared'; readonly message: string; readonly removed: readonly Key[] }
  | { readonly kind: 'pruned'; readonly message: string; readonly removed: readonly Key[] };

/** One candidate's snapshot-safe presentation and interaction binding. */
export interface SelectionOptionBinding<Key extends TargetKey> {
  readonly choice: Key;
  readonly selected: boolean;
  readonly capacityBlocked: boolean;
  toggle(): void;
}

export interface SelectionDraftSelectionBinding<Key extends TargetKey> {
  readonly candidates: readonly Key[];
  readonly selected: readonly Key[];
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

// ---- Placement drafts ----

export interface DraftPlacement<Item extends TargetKey, Target extends TargetKey> {
  readonly item: Item;
  readonly target: Target;
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

// ---- Source then destination ----

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
