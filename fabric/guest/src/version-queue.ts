/**
 * The guest's replacement for the Redux pending-bundle queue.
 *
 * The eventual host delivers the latest value of a cell, not an ordered log,
 * so each delivery is a window of recent versions. This queue turns
 * overlapping windows back into "every version exactly once, in order".
 * It has no DOM and no timers, so it is testable by itself.
 */

export interface Versioned {
  readonly version: number;
}

export interface OfferResult {
  /** The key changed (viewer or epoch), so everything shown is void. */
  readonly reset: boolean;
  /** How many entries were new and are now queued. */
  readonly accepted: number;
  /**
   * Set when the window began after the version that would have come next:
   * the versions in between were never delivered and cannot be played.
   */
  readonly gap: { readonly after: number; readonly next: number } | null;
}

export class VersionQueue<Entry extends Versioned> {
  private _pending: Entry[] = [];
  private _key: string | null = null;
  private _lastQueued: number | null = null;

  /** Versions waiting to be shown. */
  get length(): number {
    return this._pending.length;
  }

  /** The highest version ever accepted under the current key. */
  get lastQueued(): number | null {
    return this._lastQueued;
  }

  /**
   * Offer a window. `key` names the history the entries belong to; a
   * different key discards the queue. Entries need not be sorted. The first
   * window under a key contributes only its newest entry.
   */
  offer(key: string, entries: readonly Entry[]): OfferResult {
    const reset = this._key !== null && this._key !== key;
    if (this._key !== key) {
      this._key = key;
      this._pending = [];
      this._lastQueued = null;
    }
    for (const entry of entries) {
      if (!Number.isSafeInteger(entry.version) || entry.version < 0) {
        throw new Error(`state entry has an invalid version: ${JSON.stringify(entry.version)}`);
      }
    }
    let sorted = [...entries].sort((a, b) => a.version - b.version);
    if (this._lastQueued === null && sorted.length > 1) {
      // Nothing has been shown under this key, so the older entries are
      // history the viewer never saw. Cut to the newest instead of replaying
      // however much of the past the window happens to hold.
      sorted = sorted.slice(-1);
    }
    let accepted = 0;
    let gap: OfferResult['gap'] = null;
    for (const entry of sorted) {
      if (this._lastQueued !== null && entry.version <= this._lastQueued) continue;
      if (this._lastQueued !== null && entry.version > this._lastQueued + 1 && gap === null) {
        gap = { after: this._lastQueued, next: entry.version };
      }
      this._pending.push(entry);
      this._lastQueued = entry.version;
      accepted++;
    }
    return { reset, accepted, gap };
  }

  peek(): Entry | undefined {
    return this._pending[0];
  }

  shift(): Entry | undefined {
    return this._pending.shift();
  }

  /** Forget everything, including which versions were seen. */
  clear(): void {
    this._pending = [];
    this._key = null;
    this._lastQueued = null;
  }
}
