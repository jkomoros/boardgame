import { motionTransfer } from './transfer.ts';
import type { MotionTransferDeclaration } from './transfer.ts';

/** An explicit, deterministic start order for automatic structural motion. */
export interface MotionStaggerCohortSpec {
  readonly kind: 'stagger';
  readonly subjects: readonly string[];
  readonly intervalMs: number;
  readonly key?: string;
}

export interface MotionScheduleEntry {
  readonly subjectId: string;
  readonly legacyDelayMs: number;
}

export interface ScheduledMotionEntry {
  readonly subjectId: string;
  readonly delayMs: number;
}

export type MotionCohortSchedule = Readonly<{
  status: 'applied' | 'fallback';
  entries: readonly ScheduledMotionEntry[];
  reason?: 'invalid-entry' | 'invalid-cohort' | 'duplicate-subject';
}>;

function nonEmpty(value: string, label: string): string {
  const normalized = value.trim();
  if (!normalized) throw new Error(`${label} must not be empty`);
  return normalized;
}

function finiteNonnegative(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0;
}

function frozenEntries(entries: readonly MotionScheduleEntry[]): readonly ScheduledMotionEntry[] {
  return Object.freeze(entries.map(entry => Object.freeze({
    subjectId: entry.subjectId,
    delayMs: entry.legacyDelayMs,
  })));
}

function fallback(
  entries: readonly MotionScheduleEntry[],
  reason: NonNullable<MotionCohortSchedule['reason']>,
): MotionCohortSchedule {
  return Object.freeze({ status: 'fallback', reason, entries: frozenEntries(entries) });
}

/**
 * Apply explicit cohort cadence over already-computed compatibility delays.
 *
 * Explicit timing replaces the legacy delay for cohort members. A malformed
 * declaration rejects the complete explicit set, so author configuration can
 * never leave structural playback half-retimed.
 */
export function compileMotionCohortSchedule(
  entries: readonly MotionScheduleEntry[],
  cohorts: readonly MotionStaggerCohortSpec[],
): MotionCohortSchedule {
  const entryIds = new Set<string>();
  for (const entry of entries) {
    if (typeof entry.subjectId !== 'string' || !entry.subjectId.trim()
      || !finiteNonnegative(entry.legacyDelayMs)
      || entryIds.has(entry.subjectId)) return fallback(entries, 'invalid-entry');
    entryIds.add(entry.subjectId);
  }

  const explicitDelays = new Map<string, number>();
  const declaredIds = new Set<string>();
  for (const input of cohorts) {
    const cohort = input as Partial<MotionStaggerCohortSpec> | null;
    if (!cohort || cohort.kind !== 'stagger' || !finiteNonnegative(cohort.intervalMs)
      || !Array.isArray(cohort.subjects) || cohort.subjects.length === 0) {
      return fallback(entries, 'invalid-cohort');
    }
    const withinCohort = new Set<string>();
    for (let rank = 0; rank < cohort.subjects.length; rank++) {
      const subjectId = cohort.subjects[rank];
      const delayMs = rank * cohort.intervalMs;
      if (typeof subjectId !== 'string' || !subjectId.trim() || !finiteNonnegative(delayMs)) {
        return fallback(entries, 'invalid-cohort');
      }
      if (withinCohort.has(subjectId) || declaredIds.has(subjectId)) {
        return fallback(entries, 'duplicate-subject');
      }
      withinCohort.add(subjectId);
      declaredIds.add(subjectId);
      if (entryIds.has(subjectId)) explicitDelays.set(subjectId, delayMs);
    }
  }

  return Object.freeze({
    status: 'applied',
    entries: Object.freeze(entries.map(entry => Object.freeze({
      subjectId: entry.subjectId,
      delayMs: explicitDelays.get(entry.subjectId) ?? entry.legacyDelayMs,
    }))),
  });
}

type StaggerOptions = Readonly<{
  subjects: readonly string[];
  intervalMs: number;
  key?: string;
}>;

export const motion = Object.freeze({
  transfer(options: MotionTransferDeclaration) {
    return motionTransfer(options);
  },
  stagger(options: StaggerOptions): MotionStaggerCohortSpec {
    if (!finiteNonnegative(options.intervalMs)) {
      throw new Error('motion stagger intervalMs must be finite and non-negative');
    }
    if (options.subjects.length === 0) {
      throw new Error('motion stagger subjects must not be empty');
    }
    const subjects = options.subjects.map(subject => nonEmpty(subject, 'motion stagger subject ID'));
    if (new Set(subjects).size !== subjects.length) {
      throw new Error('motion stagger subjects must be unique');
    }
    return Object.freeze({
      kind: 'stagger',
      subjects: Object.freeze(subjects),
      intervalMs: options.intervalMs,
      ...(options.key === undefined ? {} : { key: nonEmpty(options.key, 'motion stagger key') }),
    });
  },
});
