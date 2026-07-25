// Ambient discovery registry for animatable items that are NOT tracked by
// the shared component animator's own stack-component bookkeeping (#714's
// "non-component" gap: standalone dice, status-text wrappers, fading-text,
// tokens, and any other game-authored BoardgameAnimatableItem that lives
// outside a boardgame-component-stack). A single instance lives on
// boardgame-render-game as `animatableRegistry`; BoardgameAnimatableItem's
// connectedCallback/disconnectedCallback walk up to find it (the same
// parent/slot walk shape as _ambientAnimationContext(), factored into the
// shared `_ambientLookup` helper -- see boardgame-animatable-item.ts) and
// self-register/unregister.
//
// Typed via a minimal structural interface rather than importing
// BoardgameAnimatableItem: boardgame-animatable-item.ts imports THIS module
// (to discover and register with an ambient registry), so this module must
// not import back, or the two files would form an import cycle.
export interface RegistrableAnimatableItem {
  // The cycle-start registry sweep force-settles a stale cycle's GATED
  // participants only, so an ungated ambient loop (an infinite highlight
  // throb) survives the sweep. finishAllAnimations (kill everything,
  // including ambient) is reserved for the tree-departure paths and is not
  // part of the sweep contract.
  finishGatedAnimations(): void;
  animationContext: unknown;
}

export class AnimatableRegistry {
  private readonly _items = new Set<RegistrableAnimatableItem>();

  // register/unregister are idempotent: a Set already ignores a duplicate
  // add, and deleting an absent item is a harmless no-op. This matters
  // because an item's cached provider reference (captured at connect time)
  // is unregistered unconditionally at disconnect, even if it was never
  // actually registered (no provider was found) or was already removed.
  register(item: RegistrableAnimatableItem): void {
    this._items.add(item);
  }

  unregister(item: RegistrableAnimatableItem): void {
    this._items.delete(item);
  }

  // items() returns a fresh snapshot array, not a live view onto the
  // internal Set. Callers (render-game's cycle-start reset) iterate the
  // result while calling into each item's finishAllAnimations(), which can
  // itself trigger connect/disconnect (and thus register/unregister) as a
  // side effect of settling animations synchronously. Iterating a snapshot
  // makes that safe: mutating the live Set mid-iteration would otherwise
  // risk skipping or double-visiting entries (or throwing, depending on the
  // engine's Set-iterator invalidation semantics).
  items(): readonly RegistrableAnimatableItem[] {
    return [...this._items];
  }
}
