import { test } from 'node:test';
import assert from 'node:assert/strict';
import { CompanionAnimationTimeline, usableAnimationContext } from './companion-sync.ts';

const policy = { slotDurationMs: 800, maxAnimationDurationMs: 600 };

function warm(timeline: CompanionAnimationTimeline, now = 10_000) {
  for (let version = 1; version <= 3; version++) {
    timeline.ingest('warmup', {
      version,
      serverSentAt: now - 25,
      serverPlayAt: now + 250,
      ...policy,
    }, now);
  }
}

test('schedule is bound to game and version, never the latest timing', () => {
  const timeline = new CompanionAnimationTimeline();
  warm(timeline);
  timeline.ingest('game-a', { version: 7, serverSentAt: 10_000, serverPlayAt: 11_000, ...policy }, 10_025);
  timeline.ingest('game-a', { version: 8, serverSentAt: 10_000, serverPlayAt: 12_000, ...policy }, 10_025);
  timeline.ingest('game-b', { version: 7, serverSentAt: 10_000, serverPlayAt: 13_000, ...policy }, 10_025);

  assert.deepEqual(timeline.schedule('game-a', 7, 10_100), {
    kind: 'scheduled', context: { version: 7, startAtMs: 11_025, ...policy },
  });
  assert.deepEqual(timeline.schedule('game-a', 8, 10_100), {
    kind: 'scheduled', context: { version: 8, startAtMs: 12_025, ...policy },
  });
  assert.deepEqual(timeline.schedule('game-b', 7, 10_100), {
    kind: 'scheduled', context: { version: 7, startAtMs: 13_025, ...policy },
  });
});

test('announced version waits only for the timing grace window', () => {
  const timeline = new CompanionAnimationTimeline();
  timeline.announce('game', 4, 1_000);
  assert.deepEqual(timeline.schedule('game', 4, 1_050), { kind: 'awaiting-timing', waitMs: 150 });
  assert.deepEqual(timeline.schedule('game', 4, 1_200), { kind: 'immediate' });
});

test('timing received during grace becomes the version schedule', () => {
  const timeline = new CompanionAnimationTimeline();
  warm(timeline);
  timeline.announce('game', 4, 10_000);
  timeline.ingest('game', { version: 4, serverSentAt: 10_050, serverPlayAt: 10_500, ...policy }, 10_075);
  assert.deepEqual(timeline.schedule('game', 4, 10_100), {
    kind: 'scheduled', context: { version: 4, startAtMs: 10_525, ...policy },
  });
});

test('cold estimator degrades to immediate playback', () => {
  const timeline = new CompanionAnimationTimeline();
  timeline.ingest('game', { version: 1, serverSentAt: 1_000, serverPlayAt: 1_250, ...policy }, 1_025);
  assert.deepEqual(timeline.schedule('game', 1, 1_050), { kind: 'immediate' });
});

test('resetGame removes only that game schedules', () => {
  const timeline = new CompanionAnimationTimeline();
  warm(timeline);
  timeline.ingest('a', { version: 1, serverSentAt: 1_000, serverPlayAt: 2_000, ...policy }, 1_025);
  timeline.ingest('b', { version: 1, serverSentAt: 1_000, serverPlayAt: 3_000, ...policy }, 1_025);
  timeline.resetGame('a');
  assert.deepEqual(timeline.schedule('a', 1, 1_100), { kind: 'immediate' });
  assert.equal(timeline.schedule('b', 1, 1_100).kind, 'scheduled');
});

test('clock sync uses the offset from the lowest-round-trip sample', () => {
  const timeline = new CompanionAnimationTimeline();
  timeline.ingestClockSync({ clientSentAt: 1_000, serverAt: 900 }, 1_100); // offset 150, RTT 100
  timeline.ingestClockSync({ clientSentAt: 2_000, serverAt: 1_950 }, 2_020); // offset 60, RTT 20
  timeline.ingestClockSync({ clientSentAt: 3_000, serverAt: 2_900 }, 3_060); // offset 130, RTT 60
  assert.equal(timeline.estimator.minOffset(), 60);
});

test('unusable future or stale targets discard the context completely', () => {
  const base = { version: 4, slotDurationMs: 800, maxAnimationDurationMs: 600 };
  assert.equal(usableAnimationContext({ ...base, startAtMs: 11_001 }, 1_000, 10_000), null);
  assert.equal(usableAnimationContext({ ...base, startAtMs: 199 }, 1_000, 10_000), null);
  assert.deepEqual(
    usableAnimationContext({ ...base, startAtMs: 1_500 }, 1_000, 10_000),
    { ...base, startAtMs: 1_500 },
  );
});
