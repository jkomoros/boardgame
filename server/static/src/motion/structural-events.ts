import type {
  StructuralExecution,
  StructuralMotionPlan,
  StructuralMotionSegment,
} from './structural-plan.js';

export type StructuralMotionEventKind = StructuralExecution['status'];

/**
 * One observed execution-state transition for a structural motion segment.
 *
 * The event is a projection of immutable plan data, not a command. Consumers
 * may decorate it, log it, or ignore it; they cannot influence structural
 * playback through this value.
 */
export interface StructuralMotionEvent {
  /** Stable for this segment/status transition within one plan generation. */
  readonly id: string;
  readonly source: StructuralMotionPlan['source'];
  readonly generation: number;
  readonly segmentIndex: number;
  readonly subjectId: string;
  readonly kind: StructuralMotionEventKind;
  readonly segment: StructuralMotionSegment;
}

function samePlan(
  previous: StructuralMotionPlan | null,
  next: StructuralMotionPlan,
): previous is StructuralMotionPlan {
  return previous?.source === next.source
    && previous.generation === next.generation;
}

/**
 * Compile plan revisions into the execution transitions newly visible in
 * `next`. The function is pure and reconstructs nothing: if an observer misses
 * an intermediate revision, it reports the status it actually received rather
 * than inventing a `started` event.
 *
 * Passing null (or a different generation) treats every segment in `next` as
 * newly observed. Segment index is part of identity so malformed plans with
 * duplicate subject IDs still produce distinct events.
 */
export function compileStructuralMotionEvents(
  previous: StructuralMotionPlan | null,
  next: StructuralMotionPlan,
): readonly StructuralMotionEvent[] {
  const continuing = samePlan(previous, next);
  const events = next.segments.flatMap((segment, segmentIndex) => {
    const before = continuing ? previous.segments[segmentIndex] : undefined;
    if (before?.subjectId === segment.subjectId
      && before.execution.status === segment.execution.status) return [];
    const kind = segment.execution.status;
    return [Object.freeze({
      id: `${next.source}:${next.generation}:${segmentIndex}:${kind}`,
      source: next.source,
      generation: next.generation,
      segmentIndex,
      subjectId: segment.subjectId,
      kind,
      segment,
    })];
  });
  return Object.freeze(events);
}
