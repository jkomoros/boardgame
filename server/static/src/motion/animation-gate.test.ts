import assert from 'node:assert/strict';
import test from 'node:test';
import { AnimationGate, type AnimationGateCallbacks } from './animation-gate.ts';

// A minimal deterministic fake clock/timer scheduler. AnimationGate never
// touches setTimeout/Date.now directly -- it goes through the injected
// setTimer/clearTimer/now callbacks -- so tests can drive the watchdog
// without any real elapsed time.
class FakeClock {
  nowMs = 0;
  private readonly timers = new Map<number, { at: number; cb: () => void }>();
  private nextId = 1;

  setTimer = (cb: () => void, ms: number): unknown => {
    const id = this.nextId++;
    this.timers.set(id, { at: this.nowMs + ms, cb });
    return id;
  };

  clearTimer = (handle: unknown): void => {
    this.timers.delete(handle as number);
  };

  now = (): number => this.nowMs;

  // Advances the clock by `ms`, firing any due timers in deadline order
  // (including ones (re-)armed by a firing callback), matching how a real
  // event loop would interleave a re-armed watchdog with the advancing clock.
  advance(ms: number): void {
    const target = this.nowMs + ms;
    for (;;) {
      let dueId: number | null = null;
      let dueAt = Infinity;
      for (const [id, t] of this.timers) {
        if (t.at <= target && t.at < dueAt) {
          dueAt = t.at;
          dueId = id;
        }
      }
      if (dueId === null) break;
      const timer = this.timers.get(dueId)!;
      this.timers.delete(dueId);
      this.nowMs = timer.at;
      timer.cb();
    }
    this.nowMs = target;
  }
}

function makeHarness(clock: FakeClock) {
  const events: string[] = [];
  const watchdogCalls: Array<{ pending: readonly string[]; budgetMs: number }> = [];
  let allDoneCount = 0;
  let openCount = 0;
  const cb: AnimationGateCallbacks = {
    onOpen: () => { openCount++; events.push('open'); },
    onAllDone: () => { allDoneCount++; events.push('all-done'); },
    onWatchdog: (pending, budgetMs) => {
      watchdogCalls.push({ pending, budgetMs });
      events.push('watchdog');
    },
    setTimer: clock.setTimer,
    clearTimer: clock.clearTimer,
    now: clock.now,
  };
  return {
    cb, events, watchdogCalls,
    allDoneCount: () => allDoneCount,
    openCount: () => openCount,
  };
}

test('open -> willAnimate -> animationDone fires onAllDone exactly once', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb);
  const ele = {};

  gate.open(1);
  assert.equal(gate.isOpen, true);
  assert.equal(gate.pendingCount, 0);

  gate.willAnimate(ele, 'boardgame-card#c1');
  assert.equal(gate.pendingCount, 1);

  gate.animationDone(ele);
  assert.equal(gate.isOpen, false);
  assert.equal(h.allDoneCount(), 1);
  assert.deepEqual(h.events, ['open', 'all-done']);

  // Duplicate done is a no-op: no second all-done fires.
  gate.animationDone(ele);
  assert.equal(h.allDoneCount(), 1);
});

test('done for an unknown element is ignored and does not close the gate', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb);
  const eleA = {};
  const eleB = {};

  gate.open(1);
  gate.willAnimate(eleA, 'a');
  gate.animationDone(eleB); // never registered
  assert.equal(gate.pendingCount, 1);
  assert.equal(gate.isOpen, true);
  assert.equal(h.allDoneCount(), 0);

  gate.animationDone(eleA);
  assert.equal(h.allDoneCount(), 1);

  // Once the map is empty (size === 0), a done() for any element -- known
  // or not -- is a no-op via the same early return.
  gate.animationDone(eleA);
  gate.animationDone(eleB);
  assert.equal(h.allDoneCount(), 1);
});

test('second open cancels the prior watchdog and resets bookkeeping', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb, { floorMs: 4000, marginMs: 1500 });
  const eleA = {};

  gate.open(1);
  gate.willAnimate(eleA, 'a', 5000); // extends deadline to 6500ms from t=0

  clock.advance(100);
  gate.open(2); // must cancel the 6500ms watchdog from cycle 1
  assert.equal(h.openCount(), 2);
  assert.equal(gate.pendingCount, 0); // bookkeeping reset

  // Advance to when the *old* cycle's extended watchdog (t=6500) would have
  // fired, but stop just short of the new cycle's floor deadline (t=4100).
  clock.advance(4000); // now at t=4100
  assert.equal(h.watchdogCalls.length, 1);
  assert.equal(h.watchdogCalls[0]?.budgetMs, 4000); // the new cycle's floor, not 6500
  assert.deepEqual(h.watchdogCalls[0]?.pending, []); // cycle 2 has no participants
});

test('watchdog fires onWatchdog then onAllDone after the floor when a participant never settles', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb, { floorMs: 4000, marginMs: 1500 });
  const ele = {};

  gate.open(1);
  gate.willAnimate(ele, 'div#stuck');
  clock.advance(4000);

  assert.deepEqual(h.events, ['open', 'watchdog', 'all-done']);
  assert.equal(h.watchdogCalls.length, 1);
  assert.deepEqual(h.watchdogCalls[0], { pending: ['div#stuck'], budgetMs: 4000 });
  assert.equal(gate.isOpen, false);
  assert.equal(h.allDoneCount(), 1);
});

test('a long declared cycle extends the watchdog deadline and is NOT force-closed at the floor', () => {
  // MANDATORY (harness-critic gap 9): this is the only coverage anywhere of
  // the expectedSettleMs deadline-extension mechanism -- no e2e scenario is
  // long enough to exercise it.
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb, { floorMs: 4000, marginMs: 1500 });
  const ele = {};

  gate.open(1);
  // Declared settle time (stagger + duration + endDelay) well beyond the
  // floor: deadline should extend to 6000 + 1500 = 7500ms, not stay at 4000.
  gate.willAnimate(ele, 'long-cycle', 6000);

  clock.advance(4000); // the old floor instant
  assert.equal(gate.isOpen, true, 'must not force-close at the floor');
  assert.equal(h.watchdogCalls.length, 0);

  // The participant settles normally before the extended deadline: the
  // extended watchdog must be cancelled, not merely deferred.
  gate.animationDone(ele);
  assert.equal(gate.isOpen, false);
  assert.equal(h.allDoneCount(), 1);

  clock.advance(10_000);
  assert.equal(h.watchdogCalls.length, 0, 'cancelled watchdog must never fire');
  assert.equal(h.allDoneCount(), 1);
});

test('a shorter willAnimate does not shrink an already-extended deadline', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb, { floorMs: 4000, marginMs: 1500 });
  const eleA = {};
  const eleB = {};

  gate.open(1);
  gate.willAnimate(eleA, 'a', 6000); // deadline -> 7500
  gate.willAnimate(eleB, 'b', 2000); // would ask for only 3500 -- must not shrink

  clock.advance(4000); // past the floor and past the shorter ask
  assert.equal(gate.isOpen, true);
  assert.equal(h.watchdogCalls.length, 0);

  clock.advance(3500); // total 7500: the extended deadline
  assert.equal(h.watchdogCalls.length, 1);
  assert.equal(h.watchdogCalls[0]?.budgetMs, 7500);
  assert.deepEqual(h.watchdogCalls[0]?.pending.slice().sort(), ['a', 'b']);
});

test('settleIfEmpty fires onAllDone only with no participants and a matching cycleId', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb);
  const ele = {};

  gate.open(5);
  gate.willAnimate(ele, 'a');

  gate.settleIfEmpty(999); // wrong cycle: no-op
  assert.equal(gate.isOpen, true);

  gate.settleIfEmpty(5); // right cycle, but a participant is still pending
  assert.equal(gate.isOpen, true);
  assert.equal(h.allDoneCount(), 0);

  gate.animationDone(ele);
  assert.equal(h.allDoneCount(), 1);

  gate.settleIfEmpty(5); // already fired: no double-fire
  assert.equal(h.allDoneCount(), 1);
});

test('settleIfEmpty with no participants closes an open cycle immediately, defaulting to the current cycleId', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb);

  gate.open(5);
  gate.settleIfEmpty();
  assert.equal(gate.isOpen, false);
  assert.equal(h.allDoneCount(), 1);
});

test('close() force-closes regardless of pending participants, guarded by cycleId and the fired latch', () => {
  // Backs the interrupted-cycle close in boardgame-render-game's
  // _stateChanged: a genuine cycle handoff must close the OLD cycle's gate
  // even though it still has unsettled participants.
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb);
  const ele = {};

  gate.open(1);
  gate.willAnimate(ele, 'a');
  assert.equal(gate.pendingCount, 1);

  gate.close(1);
  assert.equal(gate.isOpen, false);
  assert.equal(h.allDoneCount(), 1);

  gate.close(1); // already fired: no-op
  assert.equal(h.allDoneCount(), 1);

  gate.open(2);
  gate.close(999); // mismatched cycle: no-op
  assert.equal(gate.isOpen, true);
  assert.equal(h.allDoneCount(), 1);
});

test('dispose() clears an armed watchdog without force-closing the gate', () => {
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb, { floorMs: 4000, marginMs: 1500 });

  gate.open(1);
  gate.dispose();

  clock.advance(10_000);
  assert.equal(h.watchdogCalls.length, 0);
  assert.equal(h.allDoneCount(), 0);
  assert.equal(gate.isOpen, true); // dispose only clears the timer, not state
});

test('a freshly constructed gate (before any open()) starts in the unfired, no-participants state', () => {
  // Mirrors boardgame-render-game's firstUpdated(), which used to set
  // _allAnimationsDoneFired = false and a fresh _activeAnimations map
  // directly -- before any real cycle had opened -- so that a renderer
  // mounting with state already installed (cycleId 0) still gets a single
  // completion signal.
  const clock = new FakeClock();
  const h = makeHarness(clock);
  const gate = new AnimationGate(h.cb);

  assert.equal(gate.isOpen, true);
  assert.equal(gate.pendingCount, 0);
  assert.equal(h.openCount(), 0); // no onOpen callback for this implicit state

  gate.settleIfEmpty(0);
  assert.equal(h.allDoneCount(), 1);
  assert.equal(gate.isOpen, false);
});
