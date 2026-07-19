import type {
  FlipGeometry,
  ViewportGeometry,
} from './geometry.js';
import type { AnimationTimingPolicy } from './timing.js';
import { sanitizeMotionSubjectSnapshot } from './subject.ts';
import type { MotionSubjectSnapshot } from './subject.js';
import type { ComponentMotionChannel } from './component-track.js';

export type StructuralPresence = 'retained' | 'appearing' | 'departing';

export type StructuralProvenance =
  | Readonly<{ kind: 'identity' }>
  | Readonly<{
    kind: 'unresolved';
    endpoint: 'source' | 'destination';
  }>
  | Readonly<{
    kind: 'stack-history';
    endpoint: 'source' | 'destination';
    stackId: string;
    evidence: 'only-candidate' | 'runner-up' | 'latest-seen' | 'ambiguous';
  }>;

/** Privacy-safe visual path captured in one measurement transaction. */
export interface StructuralMotionPath {
  readonly kind: 'stationary' | 'travel';
  readonly from: ViewportGeometry;
  readonly to: ViewportGeometry;
}

export interface StructuralMotionDraft {
  readonly subjectId: string;
  /** Transition-local presentation declaration key; never segment identity. */
  readonly declarationKey?: string;
  readonly presence: StructuralPresence;
  readonly provenance: StructuralProvenance;
  readonly visualSubject?: MotionSubjectSnapshot;
  readonly path?: StructuralMotionPath;
  /** Exact, single-owner channels intended for this segment. */
  readonly channels: readonly ComponentMotionChannel[];
}

export interface StructuralTimingRequest {
  readonly policy: AnimationTimingPolicy;
  readonly delayMs: number;
  readonly durationMs: number;
}

export interface StructuralExecutedTiming {
  readonly channel: ComponentMotionChannel;
  readonly delayMs: number;
  readonly durationMs: number;
  readonly endDelayMs: number;
  readonly iterations: number;
  readonly easing: string;
  readonly fill: FillMode;
}

/** Stable identity for one segment throughout its immutable lifecycle. */
export interface StructuralMotionSegmentRef {
  readonly source: 'flip' | 'explicit';
  readonly generation: number;
  readonly segmentIndex: number;
}

export type StructuralExecution =
  | Readonly<{ status: 'planned' }>
  | Readonly<{
    status: 'armed' | 'active-observed';
    animations: readonly StructuralExecutedTiming[];
  }>
  | Readonly<{
    status: 'skipped';
    reason: 'not-started' | 'playback-error' | 'missing-endpoint' | 'no-spatial-change' | 'timing' | 'superseded' | 'ownership-conflict';
  }>
  | Readonly<{
    status: 'finished';
    animations: readonly StructuralExecutedTiming[];
  }>
  | Readonly<{
    status: 'cancelled';
    animations: readonly StructuralExecutedTiming[];
  }>;

export interface StructuralMotionSegment extends StructuralMotionDraft {
  readonly ref: StructuralMotionSegmentRef;
  readonly timingRequest: StructuralTimingRequest;
  readonly execution: StructuralExecution;
}

export interface StructuralMotionPlan {
  readonly source: 'flip' | 'explicit';
  readonly generation: number;
  readonly phase: 'planned' | 'executing' | 'settled';
  readonly segments: readonly StructuralMotionSegment[];
}

function snapshotChannels(
  channels: readonly Readonly<{ target: string; property: string }>[] | undefined,
): readonly ComponentMotionChannel[] {
  const seen = new Set<string>();
  const result: ComponentMotionChannel[] = [];
  for (const channel of channels ?? []) {
    if ((channel.target !== 'host' && channel.target !== 'visual')
      || (channel.property !== 'transform' && channel.property !== 'opacity')) continue;
    const key = `${channel.target}:${channel.property}`;
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(key as ComponentMotionChannel);
  }
  return Object.freeze(result);
}

export function createStructuralMotionDraft(input: Readonly<{
  subjectId: string;
  declarationKey?: string;
  presence: StructuralPresence;
  provenance: StructuralProvenance;
  visualSubject?: unknown;
  viewportFrom?: ViewportGeometry;
  viewportTo?: ViewportGeometry;
  inversion?: FlipGeometry;
  channels?: readonly Readonly<{ target: string; property: string }>[];
}>): StructuralMotionDraft {
  const visualSubject = sanitizeMotionSubjectSnapshot(input.visualSubject);
  return Object.freeze({
    subjectId: input.subjectId,
    ...(input.declarationKey ? { declarationKey: input.declarationKey } : {}),
    presence: input.presence,
    provenance: Object.freeze({ ...input.provenance }),
    ...(visualSubject ? { visualSubject } : {}),
    ...(input.viewportFrom && input.viewportTo ? {
      path: Object.freeze({
        kind: input.inversion?.changed ? 'travel' as const : 'stationary' as const,
        from: input.viewportFrom,
        to: input.viewportTo,
      }),
    } : {}),
    channels: snapshotChannels(input.channels),
  });
}

export function publishStructuralMotionPlan(
  generation: number,
  entries: readonly Readonly<{
    draft: StructuralMotionDraft;
    timingRequest: StructuralTimingRequest;
  }>[],
  source: StructuralMotionPlan['source'] = 'flip',
): StructuralMotionPlan {
  const segments = entries.map(({ draft, timingRequest }, segmentIndex) => Object.freeze({
    ...draft,
    ref: Object.freeze({ source, generation, segmentIndex }),
    timingRequest: Object.freeze({ ...timingRequest }),
    execution: Object.freeze({ status: 'planned' as const }),
  }));
  return Object.freeze({
    source,
    generation,
    phase: 'planned',
    segments: Object.freeze(segments),
  });
}

export function updateStructuralMotionExecutions(
  plan: StructuralMotionPlan,
  updates: ReadonlyMap<number, StructuralExecution>,
): StructuralMotionPlan {
  const transitions: Readonly<Record<StructuralExecution['status'], readonly StructuralExecution['status'][]>> = {
    planned: ['armed', 'skipped'],
    armed: ['active-observed', 'cancelled'],
    'active-observed': ['finished', 'cancelled'],
    skipped: [],
    finished: [],
    cancelled: [],
  };
  const segments = plan.segments.map((segment, segmentIndex) => {
    const execution = updates.get(segmentIndex);
    if (!execution || !transitions[segment.execution.status].includes(execution.status)) {
      return segment;
    }
    const frozenExecution = execution && 'animations' in execution
      ? Object.freeze({
        ...execution,
        animations: Object.freeze(execution.animations.map(timing => Object.freeze({ ...timing }))),
      })
      : execution ? Object.freeze({ ...execution }) : null;
    return Object.freeze({ ...segment, execution: frozenExecution! });
  });
  const terminal = segments.every(segment => (
    segment.execution.status === 'finished'
    || segment.execution.status === 'cancelled'
    || segment.execution.status === 'skipped'
  ));
  const started = segments.some(segment => segment.execution.status !== 'planned');
  return Object.freeze({
    ...plan,
    phase: terminal ? 'settled' : started ? 'executing' : 'planned',
    segments: Object.freeze(segments),
  });
}
