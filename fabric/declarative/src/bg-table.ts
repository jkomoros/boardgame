/**
 * <bg-table>: the single animation root for a board.
 *
 * It owns one `boardgame-component-animator` for its subtree, the queue of
 * state versions, and the ambient registry and timing context that animatable
 * items look up by walking their ancestors. It is a wrapper: all measurement,
 * FLIP solving and playback is the existing animator's.
 *
 * Inputs are attributes and JSON-serializable properties. Two events leave
 * it: `propose-move` (a person activated something) and `timer-elapsed` (a
 * timer of the installed version has run out by this browser's clock).
 *
 * Three things here are NOT in the existing elements, and each is a finding
 * (see FINDINGS.md):
 *
 *  1. Scope. The animator reads a module-global list of every stack in the
 *     document. The table narrows that list to its own subtree.
 *  2. Remembered "before". The animator already pairs components by id, not
 *     by DOM node. What it cannot do is measure the old layout after the old
 *     nodes are gone, which is the situation when a renderer recreates
 *     children. The table therefore keeps the resting geometry of the last
 *     installed state, keyed by component id, and hands it to the animator.
 *  3. A queue. The table installs one version at a time and releases the next
 *     when motion settles, so the caller never needs a settle event.
 */
import { LitElement, html, css } from 'lit';
import { property } from 'lit/decorators.js';
import '@boardgame/elements/components/boardgame-component-animator.js';
import type { BoardgameComponentAnimator } from '@boardgame/elements/components/boardgame-component-animator.js';
import { BoardgameComponentStack } from '@boardgame/elements/components/boardgame-component-stack.js';
import { AnimatableRegistry } from '@boardgame/elements/motion/animatable-registry.js';
import { captureOffsetGeometry, captureViewportGeometry } from '@boardgame/elements/motion/geometry.js';
import { captureHistoricalPresentation } from '@boardgame/elements/motion/historical-presentation.js';
import type { MotionStaggerCohortSpec } from '@boardgame/elements/motion/cohort.js';
import { expandStack, resolvePath, isWireStack } from './expand.js';
import type { Chest, WireStack, WireView } from './expand.js';
import { frameOf, isTableView, seatOf } from './table-view.js';
import type { TableViewEntry, TableViewJSON } from './table-view.js';

/** Per-version presentation hints. All data; computed by whoever supplies the view. */
export interface Transition {
  /** Name of the move that produced this version. Informational. */
  readonly move?: string;
  /** Keep this version on screen this long after motion settles, before the next queued one. */
  readonly holdMs?: number;
  /** Animation length for this transition, overriding the `animation-length` attribute. */
  readonly animationLengthMs?: number;
  /** Stagger groups, by component id. Passed to the animator unchanged. */
  readonly cohorts?: readonly MotionStaggerCohortSpec[];
}

/** One move's candidate values for one argument, as the JSON API projects them. */
export interface ChoiceSet {
  readonly MoveName: string;
  readonly FieldName: string;
  readonly Candidates: readonly { readonly Value: number | string; readonly Available?: boolean }[];
}

/**
 * Everything the table is told about ONE version, as a single JSON value.
 * It is one property on purpose: a renderer may set separate properties in any
 * order, and a view paired with another version's legality would be wrong.
 */
export interface Frame {
  /** This viewer's released view: the JSON API's `game` object. */
  readonly view: WireView;
  /** Which slot values are legal, per move and argument. Absent means unknown. */
  readonly choices?: { readonly Sets?: readonly ChoiceSet[] } | null;
  /** Which moves need no argument and are legal now, by name. */
  readonly legality?: Readonly<Record<string, { readonly legalForPlayer?: boolean }>> | null;
  readonly transition?: Transition | null;
  /** Timers running in this version, by id, with the time each had left when it was produced. */
  readonly timers?: Readonly<Record<string, { readonly TimeLeft: number }>> | null;
}

/** Detail of `timer-elapsed`: the table's own clock says this timer has run out. */
export interface TimerElapsedDetail {
  readonly timerId: string;
  readonly version: number | null;
}

/** How long after a timer's time is up the table first says so, and how it repeats. */
const TIMER_GRACE_MS = 150;
const TIMER_REPEAT_MS = 1000;
const TIMER_REPEATS = 12;

/** An element inside a table that re-renders when the installed view changes. */
export interface TableChild extends HTMLElement {
  tableChanged(): void;
  readonly updateComplete: Promise<boolean>;
  readonly isUpdatePending?: boolean;
}

/** The private animator state the table reads and writes. Each field is a finding. */
interface AnimatorInternals {
  _infoById: Record<string, Record<string, unknown>>;
  _beforeSeenIds: Set<string>;
  _beforeCollectionOffsets: Map<string, unknown>;
  _beforeCollectionViewportOffsets: Map<string, unknown>;
  _historicalPresentationById: Map<string, unknown>;
}

interface Snapshot {
  readonly infoById: Record<string, Record<string, unknown>>;
  readonly collectionOffsets: Map<string, unknown>;
  readonly collectionViewportOffsets: Map<string, unknown>;
}

interface Retained {
  snapshot: Snapshot | null;
  memory: Map<string, unknown>;
}

/** Survives recreation of the table element itself, keyed by its `name`. */
const retainedByName = new Map<string, Retained>();

const globalStackList = Object.getOwnPropertyDescriptor(
  BoardgameComponentStack.prototype, '_sharedStackList',
)!.get as (this: BoardgameComponentStack) => BoardgameComponentStack[];

/** Events of the wrapped elements that carry layout or timing. None may leave the table. */
const INTERNAL_EVENTS = [
  'will-animate', 'animation-done', 'stack-slot-geometry-changed',
  'component-tapped', 'motion-cycle-release',
] as const;

/** Re-assign values set as own properties before the element's class was defined. */
export function adoptEarlyProperties(element: HTMLElement, names: readonly string[]): void {
  const bag = element as unknown as Record<string, unknown>;
  for (const name of names) {
    if (!Object.prototype.hasOwnProperty.call(element, name)) continue;
    const value = bag[name];
    delete bag[name];
    bag[name] = value;
  }
}

export function findTable(from: Node): BgTable | null {
  let node: Node | null = from;
  while (node) {
    if (node instanceof BgTable) return node;
    if (node instanceof ShadowRoot) node = node.host;
    else node = (node as Element).assignedSlot ?? node.parentNode;
  }
  return null;
}

export class BgTable extends LitElement {
  static override styles = css`
    :host { display: block; position: relative; }
    boardgame-component-animator { display: none; }
  `;

  /** The game's static component catalogue. Set once. */
  @property({ attribute: false }) chest: Chest | null = null;

  @property({ type: String, attribute: 'game-name' }) gameName = '';

  /** Stable identity. Required for animation to survive recreation of the table itself. */
  @property({ type: String }) name = '';

  /** Default animation length in milliseconds. 0 leaves the element default (250). */
  @property({ type: Number, attribute: 'animation-length' }) animationLength = 0;

  /** Seat this table is shown to; -1 is an observer. */
  @property({ type: Number }) viewer = -1;

  /**
   * EXPERIMENT SWITCH. `remembered` (default) pairs recreated children with the
   * geometry the table remembered. `live` is the animator as it is today: it
   * can only measure nodes that still exist.
   */
  @property({ type: String }) measure: 'remembered' | 'live' = 'remembered';

  /**
   * EXPERIMENT SWITCH. `table` (default) animates only this table's stacks.
   * `document` is the animator as it is today: every stack on the page.
   */
  @property({ type: String }) scope: 'table' | 'document' = 'table';

  /**
   * Milliseconds to keep the current version on screen before installing one
   * produced by the named move: `{"Capture Cards": 1000}`. Static data that
   * replaces a renderer computing `post-animation-delay` from state.
   */
  @property({ type: Object, attribute: 'hold-before' })
  holdBefore: Readonly<Record<string, number>> = {};

  /** Animation length in milliseconds by move name, overriding `animation-length`. */
  @property({ type: Object, attribute: 'animation-lengths' })
  animationLengths: Readonly<Record<string, number>> = {};

  // Ambient providers, found by BoardgameAnimatableItem's ancestor walk.
  readonly animatableRegistry = new AnimatableRegistry();
  animationContext: null = null;
  get gameVersion(): number { return this._installed?.Version ?? 0; }

  private _queue: Frame[] = [];
  private _frame: Frame | null = null;
  private get _installed(): WireView | null { return this._frame?.view ?? null; }
  private get _installedTransition(): Transition | null { return this._frame?.transition ?? null; }
  private _pumping = false;
  private _children = new Set<TableChild>();
  private _expanded = new WeakMap<object, unknown>();
  private _memory = new Map<string, unknown>();

  private _ready = false;
  private _hasRested = false;
  private _snapshot: Snapshot | null = null;
  private _changeScheduled = false;
  private _structural = false;
  private _beforeSource: 'live' | 'snapshot' = 'snapshot';
  private _token = 0;
  private _open = 0;
  private _flying = false;
  private _waiters: Array<() => void> = [];
  private _resize: ResizeObserver | null = null;

  /** Counters for tests and for the findings. Not part of the declarative API. */
  readonly stats = { cycles: 0, animatedCycles: 0, animationsStarted: 0, installs: 0, interrupted: 0 };

  // ---- the declarative input -------------------------------------------

  private _tableView: TableViewJSON | null = null;
  private _principal = '';
  private _epoch: number | undefined = undefined;
  private _queuedVersion = -1;
  private _entryOf = new WeakMap<Frame, TableViewEntry>();
  private _timerHandles: ReturnType<typeof setTimeout>[] = [];

  /**
   * The view a Common Fabric table releases, whole and untouched: seats, and
   * a window of the most recent versions. The table queues every version it
   * has not queued yet, so a caller that only ever holds "the newest value"
   * loses no intermediate version. A value that is not a table view (the
   * `undefined` before the first move) is ignored.
   */
  set tableView(value: TableViewJSON | null | undefined) {
    if (!isTableView(value)) return;
    this._tableView = value;
    this.setAttribute('phase', value.phase ?? 'seating');
    if (value.gameName && !this.gameName) this.gameName = value.gameName;
    this._syncViewer();
    const entries = value.entries ?? [];
    if (value.epoch !== this._epoch) {
      // A new deal: nothing before it is comparable.
      this._epoch = value.epoch;
      this._queuedVersion = -1;
      this._queue = [];
    }
    const fresh = entries.filter(entry => entry.version > this._queuedVersion);
    // Someone arriving mid-game is shown the present, not a replay.
    const queue = this._queuedVersion < 0 ? fresh.slice(-1) : fresh;
    for (const entry of queue) {
      const frame = frameOf(entry, this.viewer);
      this._entryOf.set(frame, entry);
      this._queuedVersion = entry.version;
      this.frame = frame;
    }
    for (const child of this._children) child.tableChanged();
  }

  get tableView(): TableViewJSON | null {
    return this._tableView;
  }

  /** Who is looking, as the table names its seats' holders. Decides `viewer`. */
  set principal(value: string | null | undefined) {
    this._principal = typeof value === 'string' ? value : '';
    this._syncViewer();
  }

  get principal(): string {
    return this._principal;
  }

  /** Seat labels of the table view, in seat order. */
  get seatLabels(): string[] {
    return (this._tableView?.seats ?? []).map(seat => seat.label);
  }

  /** A viewer who sat down sees the same versions with their own legality. */
  private _syncViewer(): void {
    if (!this._tableView) return;
    const seat = seatOf(this._tableView, this._principal);
    if (seat === this.viewer) return;
    this.viewer = seat;
    const again = (frame: Frame): Frame => {
      const entry = this._entryOf.get(frame);
      if (!entry) return frame;
      const next = { ...frameOf(entry, seat), view: frame.view };
      this._entryOf.set(next, entry);
      return next;
    };
    this._queue = this._queue.map(again);
    if (this._frame) {
      this._frame = again(this._frame);
      // Same stacks, different legality: have every stack re-stamp its slots.
      this._expanded = new WeakMap();
      for (const child of this._children) child.tableChanged();
    }
  }

  private _armTimers(frame: Frame): void {
    for (const handle of this._timerHandles) clearTimeout(handle);
    this._timerHandles = [];
    for (const [timerId, timer] of Object.entries(frame.timers ?? {})) {
      const fire = (left: number): void => {
        if (this._frame !== frame || !this.isConnected) return;
        const detail: TimerElapsedDetail = Object.freeze({
          timerId, version: frame.view.Version ?? null,
        });
        this.dispatchEvent(new CustomEvent('timer-elapsed', { detail, bubbles: true, composed: true }));
        if (left > 0) this._timerHandles.push(setTimeout(() => fire(left - 1), TIMER_REPEAT_MS));
      };
      this._timerHandles.push(setTimeout(() => fire(TIMER_REPEATS),
        Math.max(0, timer.TimeLeft) + TIMER_GRACE_MS));
    }
  }

  /**
   * THE input. Assigning a frame queues that version; the table installs
   * versions one at a time and releases the next when motion settles.
   */
  set frame(value: Frame | null) {
    if (!value?.view) return;
    this._queue.push(value);
    void this._pump();
  }

  get frame(): Frame | null {
    return this._frame;
  }

  /** Shorthand for a frame that is only a view. */
  set view(value: WireView | null) {
    if (value) this.frame = { view: value };
  }

  get view(): WireView | null {
    return this._installed;
  }

  /**
   * Legal values for one move argument in the installed version, or null when
   * the frame said nothing about legality (then every target slot is enabled).
   */
  candidates(move: string, arg: string): readonly (number | string)[] | null {
    if (!this._frame?.choices) return null;
    const sets = this._frame.choices.Sets ?? [];
    return sets.filter(set => set.MoveName === move && set.FieldName === arg)
      .flatMap(set => set.Candidates.filter(candidate => candidate.Available !== false)
        .map(candidate => candidate.Value));
  }

  /** Whether an argument-less move is legal now; null when the frame did not say. */
  legal(move: string): boolean | null {
    const legality = this._frame?.legality;
    if (!legality) return null;
    return !!legality[move]?.legalForPlayer;
  }

  /** The installed view's state. Children resolve their `path` against this. */
  get state(): unknown {
    return this._installed?.CurrentState ?? null;
  }

  get animating(): boolean {
    return this._open > 0 || this._pumping;
  }

  resolve(path: string): unknown {
    return resolvePath(this._installed?.CurrentState, path);
  }

  /** Expand a wire stack once per object, so unchanged stacks keep their identity. */
  expand(stack: unknown): unknown {
    if (!isWireStack(stack)) return null;
    let expanded = this._expanded.get(stack);
    if (!expanded) {
      expanded = expandStack(stack as WireStack, this.chest, this.gameName,
        this._installed?.CurrentState ?? null);
      this._expanded.set(stack, expanded);
    }
    return expanded;
  }

  // ---- services for child elements -------------------------------------

  register(child: TableChild): void {
    this._children.add(child);
  }

  unregister(child: TableChild): void {
    this._children.delete(child);
  }

  /** Small per-name memory so a recreated child can know its previous value. */
  recall(key: string): unknown { return this._memory.get(key); }
  remember(key: string, value: unknown): void { this._memory.set(key, value); }

  /**
   * A child reports that the board is changing. The FIRST report in a task
   * decides where "before" comes from: if nothing has been removed yet the
   * animator can still measure it; otherwise the remembered snapshot is used.
   */
  noteChange(structural: boolean): void {
    if (this._changeScheduled) return;
    this._changeScheduled = true;
    this._structural = structural;
    if (this._open > 0) this.stats.interrupted++;
    this._open++;
    if (!structural && this._ready) {
      this._animator.prepare();
      this._beforeSource = 'live';
    } else {
      this._beforeSource = 'snapshot';
    }
    const token = ++this._token;
    queueMicrotask(() => {
      this._changeScheduled = false;
      void this._finish(token);
    });
  }

  /** Resolves when no motion cycle is open and the queue is empty. For hosts and tests. */
  whenSettled(): Promise<void> {
    if (this._open === 0 && !this._pumping && this._queue.length === 0) return Promise.resolve();
    return new Promise(resolve => this._waiters.push(resolve));
  }

  // ---- lifecycle --------------------------------------------------------

  override connectedCallback(): void {
    super.connectedCallback();
    // A host may assign these before the element is defined. Lit re-applies
    // its own reactive properties after an upgrade; plain accessors it does
    // not know about stay shadowed by the instance value unless moved here.
    adoptEarlyProperties(this, ['chest', 'principal', 'tableView', 'frame', 'view']);
    for (const type of INTERNAL_EVENTS) this.addEventListener(type, this._swallow);
    const retained = this.name ? retainedByName.get(this.name) : undefined;
    if (retained) {
      this._snapshot = retained.snapshot;
      this._memory = retained.memory;
      this._hasRested = retained.snapshot !== null;
    }
  }

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    for (const type of INTERNAL_EVENTS) this.removeEventListener(type, this._swallow);
    this._resize?.disconnect();
    this._resize = null;
    for (const handle of this._timerHandles) clearTimeout(handle);
    this._timerHandles = [];
    if (this.name) retainedByName.set(this.name, { snapshot: this._snapshot, memory: this._memory });
    this._token++;
    this._open = 0;
    this._flush();
  }

  private _gatedPlays = 0;

  private readonly _swallow = (event: Event): void => {
    if (event.type === 'will-animate') this._gatedPlays++;
    event.stopPropagation();
  };

  protected override firstUpdated(): void {
    void this._becomeReady();
  }

  protected override updated(changed: Map<string, unknown>): void {
    if (changed.has('animationLength')) this._applyAnimationLength();
    if (changed.has('chest') || changed.has('gameName')) {
      this._expanded = new WeakMap();
      for (const child of this._children) child.tableChanged();
    }
  }

  private get _animator(): BoardgameComponentAnimator {
    return this.renderRoot.querySelector('boardgame-component-animator')!;
  }

  private get _internals(): AnimatorInternals {
    return this._animator as unknown as AnimatorInternals;
  }

  private _readyPromise: Promise<void> | null = null;

  private _becomeReady(): Promise<void> {
    this._readyPromise ??= (async () => {
      await this.updateComplete;
      const animator = this._animator;
      await animator.updateComplete;
      // FINDING (S9): the animator reaches every stack in the document through
      // its inner stack's `_sharedStackList`. Narrow it to this table.
      const inner = animator.shadowRoot!.querySelector('boardgame-component-stack')!;
      Object.defineProperty(inner, '_sharedStackList', {
        configurable: true,
        get: () => this._stacks(),
      });
      this._ready = true;
      this._resize = new ResizeObserver(() => {
        // Remembered geometry is stale after a reflow. Re-measure at rest.
        if (this._open === 0 && this._hasRested) this._snapshot = this._measure();
      });
      this._resize.observe(this);
    })();
    return this._readyPromise;
  }

  /** Stacks that belong to this table: the nearest table ancestor is this one. */
  private _stacks(): BoardgameComponentStack[] {
    const everything = globalStackList.call(
      this._animator.shadowRoot!.querySelector('boardgame-component-stack')!);
    if (this.scope === 'document') return everything;
    return everything.filter(stack => findTable(stack) === this);
  }

  private _applyAnimationLength(): void {
    const move = this._installedTransition?.move ?? '';
    const ms = this._installedTransition?.animationLengthMs
      ?? this.animationLengths[move] ?? this.animationLength;
    if (ms > 0) this.style.setProperty('--animation-length', `${ms}ms`);
    else this.style.removeProperty('--animation-length');
  }

  // ---- the queue ---------------------------------------------------------

  private async _pump(): Promise<void> {
    if (this._pumping) return;
    this._pumping = true;
    try {
      while (this._queue.length > 0 && this.isConnected) {
        const entry = this._queue.shift()!;
        const hold = this.holdBefore[entry.transition?.move ?? ''] ?? 0;
        if (hold > 0 && this._frame) await new Promise(resolve => setTimeout(resolve, hold));
        this.stats.installs++;
        // "Before" has to be captured while the old state is still rendered.
        this.noteChange(false);
        this._frame = entry;
        this._applyAnimationLength();
        for (const child of this._children) child.tableChanged();
        await this._cycleDone();
        // Only the newest version's timers are still running.
        if (this._queue.length === 0) this._armTimers(entry);
        const after = entry.transition?.holdMs ?? 0;
        if (after > 0 && this._queue.length > 0) {
          await new Promise(resolve => setTimeout(resolve, after));
        }
      }
    } finally {
      this._pumping = false;
      this._flush();
    }
  }

  private _cycleDone(): Promise<void> {
    if (this._open === 0) return Promise.resolve();
    return new Promise(resolve => this._cycleWaiters.push(resolve));
  }

  private _cycleWaiters: Array<() => void> = [];

  private _flush(): void {
    if (this._open > 0) return;
    const cycle = this._cycleWaiters;
    this._cycleWaiters = [];
    for (const resolve of cycle) resolve();
    if (this._pumping || this._queue.length > 0) return;
    const waiters = this._waiters;
    this._waiters = [];
    for (const resolve of waiters) resolve();
  }

  // ---- one motion cycle ---------------------------------------------------

  private async _finish(token: number): Promise<void> {
    try {
      await this._becomeReady();
      await this._rendered();
      if (token !== this._token || !this.isConnected) return;
      if (this._hasRested) {
        const animator = this._animator;
        if (this._beforeSource === 'snapshot') {
          // The old nodes may be gone. Open the animator's cycle, then replace
          // what it just measured (the NEW layout) with the remembered OLD one.
          animator.prepare();
          if (this.measure === 'remembered' && this._snapshot) this._restore(this._snapshot);
        }
        const cohorts = this._installedTransition?.cohorts;
        if (cohorts?.length) animator.installMotionCohorts(cohorts);
        const playsBefore = this._gatedPlays;
        this._flying = true;
        const settled = animator.animateFlip();
        // animateFlip measures the new layout two microtasks from now, with
        // every component at rest. Remember it then: it is the next "before".
        await null; await null; await null;
        if (token === this._token) this._snapshot = this._measure();
        await settled;
        await this._itemsSettled();
        if (token !== this._token) return;
        this._flying = false;
        this.stats.cycles++;
        const started = this._gatedPlays - playsBefore;
        this.stats.animationsStarted += started;
        if (started > 0) this.stats.animatedCycles++;
      } else {
        // First render: there is no previous layout to animate from. A stack
        // still eases its components into a messy or fanned pose by itself.
        await this._itemsSettled();
        if (token !== this._token) return;
        this._hasRested = true;
      }
      this._snapshot = this._measure();
    } finally {
      if (token === this._token) {
        this._open = 0;
        this._flush();
      }
    }
  }

  /** Every animatable item under this table (cards, callouts) has no gated animation left. */
  private async _itemsSettled(): Promise<void> {
    await Promise.all(this.animatableRegistry.items().map(item =>
      (item as unknown as { settled?: () => Promise<void> }).settled?.()));
  }

  /** Wait until every table child, stack and component has finished rendering. */
  private async _rendered(): Promise<void> {
    for (let pass = 0; pass < 12; pass++) {
      const pending: Array<Promise<unknown>> = [];
      for (const child of this._children) pending.push(child.updateComplete);
      await Promise.all(pending);
      const stacks = this._stacks();
      await Promise.all(stacks.map(stack => stack.updateComplete));
      const components = stacks.flatMap(stack => stack.Components as LitElement[]);
      await Promise.all(components.map(component => component.updateComplete));
      await null;
      const busy = [...this._children, ...this._stacks(),
        ...this._stacks().flatMap(stack => stack.Components as LitElement[])]
        .some(element => (element as { isUpdatePending?: boolean }).isUpdatePending);
      if (!busy) return;
    }
  }

  /**
   * Record where every identified component rests right now, in exactly the
   * shape the animator's prepare() records it. This is the piece prepare()
   * cannot provide once the nodes it would measure have been replaced.
   */
  private _measure(): Snapshot {
    const animator = this._animator;
    const ancestor = animator.ancestorOffsetParent;
    const infoById: Record<string, Record<string, unknown>> = {};
    const collectionOffsets = new Map<string, unknown>();
    const collectionViewportOffsets = new Map<string, unknown>();
    for (const stack of this._stacks()) {
      const offsetComponent = stack.offsetComponent as HTMLElement;
      collectionOffsets.set(stack.id, captureOffsetGeometry(offsetComponent, ancestor));
      collectionViewportOffsets.set(stack.id, captureViewportGeometry(offsetComponent));
      for (const component of stack.Components as Array<HTMLElement & Record<string, any>>) {
        if (!component.id) continue;
        const record = infoById[component.id] ?? {};
        (record.beforeCollectionIds ??= [] as string[]);
        (record.beforeCollectionIds as string[]).push(stack.id);
        record.offsets = captureOffsetGeometry(component, ancestor);
        record.viewportOffsets = captureViewportGeometry(component);
        record.visualSubject = component.motionSubjectSnapshot?.() ?? undefined;
        // Inline, not computed: a host animation may be mid-flight, and the
        // authored style is the resting value.
        record.beforeTransform = component.style.transform;
        record.beforeInlineTransform = component.style.transform;
        record.beforeOpacity = component.style.opacity || '1';
        record.before = component.animatingPropValues();
        try {
          record.historicalPresentation = captureHistoricalPresentation(component as never)
            ?? undefined;
        } catch {
          record.historicalPresentation = undefined;
        }
        infoById[component.id] = record;
      }
    }
    return { infoById, collectionOffsets, collectionViewportOffsets };
  }

  /** Hand a remembered snapshot to the animator as this cycle's "before". */
  private _restore(snapshot: Snapshot): void {
    const internals = this._internals;
    const infoById: Record<string, Record<string, unknown>> = {};
    for (const [id, record] of Object.entries(snapshot.infoById)) {
      infoById[id] = {
        ...record,
        historicalPresentation: record.historicalPresentation
          ?? internals._historicalPresentationById.get(id),
      };
    }
    internals._infoById = infoById;
    internals._beforeSeenIds = new Set(Object.keys(infoById));
    internals._beforeCollectionOffsets = new Map(snapshot.collectionOffsets);
    internals._beforeCollectionViewportOffsets = new Map(snapshot.collectionViewportOffsets);
  }

  override render() {
    return html`<slot></slot><boardgame-component-animator></boardgame-component-animator>`;
  }
}

customElements.define('bg-table', BgTable);

declare global {
  interface HTMLElementTagNameMap { 'bg-table': BgTable; }
}
