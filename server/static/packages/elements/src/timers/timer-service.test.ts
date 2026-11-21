import assert from 'node:assert/strict';
import test from 'node:test';
import { TimerService, type TimerReading } from './timer-service.ts';

test('timer service isolates frame and second cadence subscribers', () => {
  const service = new TimerService();
  const frames: TimerReading[] = [];
  const seconds: TimerReading[] = [];
  service.subscribe('timer-1', 'frame', reading => frames.push(reading));
  service.subscribe('timer-1', 'second', reading => seconds.push(reading));

  service.update({ 'timer-1': { TimeLeft: 2500, originalTimeLeft: 3000 } });
  service.update({ 'timer-1': { TimeLeft: 2400, originalTimeLeft: 3000 } });
  service.update({ 'timer-1': { TimeLeft: 1900, originalTimeLeft: 3000 } });

  assert.deepEqual(frames.map(reading => reading.timeLeftMs), [0, 2500, 2400, 1900]);
  assert.deepEqual(seconds.map(reading => reading.secondsLeft), [0, 3, 2]);
  assert.equal(frames.at(-1)?.progress, 1900 / 3000);
});

test('timer service reports elapsed/removal, unsubscribes, and rejects malformed clocks', () => {
  const service = new TimerService();
  const statuses: string[] = [];
  const unsubscribe = service.subscribe('timer-1', 'second', reading => statuses.push(reading.status));
  service.update({ 'timer-1': { TimeLeft: 1, originalTimeLeft: 10 } });
  service.update({ 'timer-1': { TimeLeft: 0, originalTimeLeft: 10 } });
  service.update({});
  unsubscribe();
  service.update({ 'timer-1': { TimeLeft: 5, originalTimeLeft: 10 } });
  assert.deepEqual(statuses, ['unavailable', 'running', 'elapsed', 'unavailable']);
  assert.throws(() => service.update({ bad: { TimeLeft: -1 } }), /TimeLeft/);
  assert.throws(() => service.update({ bad: { TimeLeft: 2, originalTimeLeft: 1 } }), /originalTimeLeft/);
});
