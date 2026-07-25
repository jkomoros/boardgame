// Pure, DOM-free kernel for the animation-completion gate previously inlined
// in boardgame-render-game.ts (see _resetAnimating / _armWatchdog /
// _componentWillAnimate / _componentAnimationDone / _nextStateIfNoAnimations /
// _notifyAnimationsDone). Extracted verbatim (behavior-for-behavior) so the
// gate's timing/bookkeeping invariants -- in particular the watchdog-deadline
// extension, which has no e2e coverage long enough to exercise it -- can be
// unit-tested with fake timers instead of only indirectly through the DOM.
//
// The gate tracks one open "cycle" of in-flight animations at a time,
// identified by an opaque `cycleId` supplied by the caller (render-game's
// motionCycleId). Participants are registered via willAnimate()/keyed by
// object identity (the animating element) and cleared via animationDone().
// When the last participant clears, or the watchdog backstop fires, the gate
// "closes" and reports onAllDone() exactly once per cycle.

export interface AnimationGateCallbacks {
  // Gate transitions closed -> open (open() was called).
  onOpen(): void;
  // The last participant settled, the watchdog force-fired, or a caller
  // explicitly closed the cycle. Fires at most once per open() cycle.
  onAllDone(): void;
  // The watchdog backstop fired: `pending` lists the still-unsettled
  // participants' labels (in registration order), `budgetMs` is the
  // watchdog's armed budget at fire time (the floor, or an extended
  // deadline). Called immediately before onAllDone().
  onWatchdog(pending: readonly string[], budgetMs: number): void;
  // Injectable timer so tests can use a fake clock instead of real time.
  setTimer(cb: () => void, ms: number): unknown;
  clearTimer(handle: unknown): void;
  now(): number;
}

export interface AnimationGateOptions {
  floorMs?: number;
  marginMs?: number;
}

// The watchdog floor: the gate never gets less than this before firing, even
// for trivially short cycles. Longer declared cycles push it out.
const DEFAULT_FLOOR_MS = 4000;
// Slack added past a declared long cycle's expected settle instant, so
// normal per-animation jitter/scheduling never trips the watchdog.
const DEFAULT_MARGIN_MS = 1500;

export class AnimationGate {
  private readonly cb: AnimationGateCallbacks;
  private readonly floorMs: number;
  private readonly marginMs: number;

  // Participants of the current cycle: animating element -> its label (used
  // only for the watchdog's diagnostic message). Mirrors the previous
  // _activeAnimations map's keys; the label is precomputed by the caller
  // (render-game derives tag#id) since the kernel has no DOM access.
  //
  // Starts as a fresh, empty map with allDoneFired = false -- NOT the
  // "never opened" defaults -- because this mirrors render-game's
  // firstUpdated(), which used to set exactly this state directly (before
  // any real open() cycle) so that a renderer mounting with state already
  // installed still gets a single completion signal at cycleId 0.
  private activeAnimations = new Map<object, string>();
  private allDoneFired = false;
  private cycleId = 0;

  private watchdogTimer: unknown = null;
  // Largest declared settle time (delay + duration + endDelay) reported by
  // any willAnimate() in the current cycle. Used to extend the watchdog past
  // a legitimately long cycle (stagger + post-animation-delay + long
  // animation length) rather than force-close mid-animation. Reset to 0 at
  // each open().
  private maxExpectedSettleMs = 0;
  // Absolute clock instant (cb.now()-comparable) the current watchdog is
  // armed to fire at. Tracked so an incoming willAnimate() can tell whether
  // a longer play would outlast the deadline and re-arm.
  private watchdogDeadlineEpoch = 0;

  constructor(cb: AnimationGateCallbacks, opts: AnimationGateOptions = {}) {
    this.cb = cb;
    this.floorMs = opts.floorMs ?? DEFAULT_FLOOR_MS;
    this.marginMs = opts.marginMs ?? DEFAULT_MARGIN_MS;
  }

  get isOpen(): boolean {
    return !this.allDoneFired;
  }

  get pendingCount(): number {
    return this.activeAnimations.size;
  }

  // Opens a new cycle under `cycleId`. Mirrors the old _resetAnimating():
  // cancels any watchdog left over from a previous cycle, clears
  // bookkeeping, and arms the watchdog at the floor. If animations complete
  // normally, the gate closes (and cancels the watchdog) before it fires.
  open(cycleId: number): void {
    this.cycleId = cycleId;
    if (this.watchdogTimer !== null) {
      this.cb.clearTimer(this.watchdogTimer);
      this.watchdogTimer = null;
    }
    this.maxExpectedSettleMs = 0;
    this.activeAnimations = new Map();
    this.allDoneFired = false;
    this.cb.onOpen();
    this.armWatchdog(this.floorMs);
  }

  // Registers a participant of the current cycle. `expectedSettleMs`, if
  // longer than any declared so far, extends (never shrinks) the watchdog
  // deadline to expectedSettleMs + marginMs.
  willAnimate(ele: object, label: string, expectedSettleMs?: number): void {
    this.activeAnimations.set(ele, label);
    if (typeof expectedSettleMs === 'number' && expectedSettleMs > this.maxExpectedSettleMs) {
      this.maxExpectedSettleMs = expectedSettleMs;
      const targetEpoch = this.cb.now() + expectedSettleMs + this.marginMs;
      if (targetEpoch > this.watchdogDeadlineEpoch) {
        this.armWatchdog(targetEpoch - this.cb.now());
      }
    }
  }

  // Clears a settled participant. Closes the cycle once the last
  // participant clears. A duplicate/unknown-element call after the map is
  // already empty is a no-op (the size === 0 early return).
  animationDone(ele: object): void {
    if (this.activeAnimations.size === 0) return;
    this.activeAnimations.delete(ele);
    if (this.activeAnimations.size === 0) {
      this.close(this.cycleId);
    }
  }

  // Closes the cycle if (and only if) it has no pending participants and
  // `cycleId` still matches the currently open cycle. Replaces
  // _nextStateIfNoAnimations.
  settleIfEmpty(cycleId: number = this.cycleId): void {
    if (cycleId !== this.cycleId) return;
    if (this.activeAnimations.size === 0) {
      this.close(cycleId);
    }
  }

  // Force-closes the cycle regardless of pending participants, guarded only
  // by cycleId matching and the already-fired latch. Backs the interrupted-
  // cycle close in render-game's _stateChanged (a genuine cycle handoff must
  // close the OLD cycle's gate even with unsettled participants) and the
  // post-mount completion signal for a renderer instantiated after state was
  // already installed. Replaces the unconditional part of
  // _notifyAnimationsDone.
  close(cycleId: number = this.cycleId): void {
    if (cycleId !== this.cycleId) return;
    if (this.allDoneFired) return;
    if (this.watchdogTimer !== null) {
      this.cb.clearTimer(this.watchdogTimer);
      this.watchdogTimer = null;
    }
    this.allDoneFired = true;
    this.cb.onAllDone();
  }

  // Clears any armed watchdog without otherwise touching gate state.
  // Replaces disconnectedCallback's watchdog cleanup.
  dispose(): void {
    if (this.watchdogTimer !== null) {
      this.cb.clearTimer(this.watchdogTimer);
      this.watchdogTimer = null;
    }
  }

  private armWatchdog(fromNowMs: number): void {
    if (this.watchdogTimer !== null) {
      this.cb.clearTimer(this.watchdogTimer);
      this.watchdogTimer = null;
    }
    this.watchdogDeadlineEpoch = this.cb.now() + fromNowMs;
    this.watchdogTimer = this.cb.setTimer(() => {
      this.watchdogTimer = null;
      if (this.allDoneFired) return;
      const pending: string[] = [];
      for (const label of this.activeAnimations.values()) {
        pending.push(label);
      }
      this.cb.onWatchdog(pending, fromNowMs);
      this.close(this.cycleId);
    }, fromNowMs);
  }
}
