/** Pure wall-clock countdown math shared by the Redux timer adapter. */
export function remainingTimerMs(initialTimeLeft: number, startedAt: number, now: number): number {
  if (!Number.isFinite(initialTimeLeft) || initialTimeLeft < 0) {
    throw new Error('timer clock: initialTimeLeft must be a finite non-negative number');
  }
  if (!Number.isFinite(startedAt) || !Number.isFinite(now)) {
    throw new Error('timer clock: wall-clock times must be finite numbers');
  }
  const elapsed = Math.max(0, now - startedAt);
  return Math.max(0, initialTimeLeft - elapsed);
}
