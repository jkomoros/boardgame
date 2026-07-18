import type {
  FlipGeometry,
  OffsetGeometry,
  ViewportGeometry,
} from './geometry.js';
import type { AnimationTimingPolicy } from './timing.js';

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

export interface StructuralPropertyChange {
  readonly name: string;
  readonly before: StructuralValueSnapshot;
  readonly after: StructuralValueSnapshot;
}

export type StructuralValueSnapshot =
  | string
  | number
  | boolean
  | null
  | undefined
  | Readonly<{ kind: 'opaque' | 'non-finite-number' }>;

export interface StructuralSpatialChange {
  readonly offsetFrom?: OffsetGeometry;
  readonly offsetTo?: OffsetGeometry;
  /** Historical visual endpoints captured in the same measurement transaction. */
  readonly viewportFrom: ViewportGeometry;
  readonly viewportTo: ViewportGeometry;
  readonly inversion: FlipGeometry;
}

export interface StructuralMotionDraft {
  readonly subjectId: string;
  readonly presence: StructuralPresence;
  readonly provenance: StructuralProvenance;
  readonly spatial?: StructuralSpatialChange;
  readonly transform?: Readonly<{ before: string; after: string }>;
  readonly properties: readonly StructuralPropertyChange[];
  readonly opacity?: Readonly<{ before: number; after: number }>;
}

export interface StructuralTimingRequest {
  readonly policy: AnimationTimingPolicy;
  readonly delayMs: number;
  readonly durationMs: number;
}

export interface StructuralExecutedTiming {
  readonly delayMs: number;
  readonly durationMs: number;
  readonly endDelayMs: number;
  readonly iterations: number;
  readonly easing: string;
  readonly fill: FillMode;
}

export type StructuralExecution =
  | Readonly<{ status: 'planned' }>
  | Readonly<{
    status: 'started';
    animations: readonly StructuralExecutedTiming[];
  }>
  | Readonly<{
    status: 'skipped';
    reason: 'not-started' | 'missing-endpoint' | 'no-spatial-change' | 'timing';
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
  readonly timingRequest: StructuralTimingRequest;
  readonly execution: StructuralExecution;
}

export interface StructuralMotionPlan {
  readonly source: 'flip' | 'explicit';
  readonly generation: number;
  readonly phase: 'planned' | 'executing' | 'settled';
  readonly segments: readonly StructuralMotionSegment[];
}

export type StructuralMotionObserver = (plan: StructuralMotionPlan) => void;

function finiteOpacity(value: string | undefined): number {
  const parsed = Number.parseFloat(value || '1');
  return Number.isFinite(parsed) ? parsed : 1;
}

function snapshotValue(value: unknown): StructuralValueSnapshot {
  if (value === null || value === undefined) return value;
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : Object.freeze({ kind: 'non-finite-number' });
  }
  if (typeof value === 'string' || typeof value === 'boolean') return value;
  return Object.freeze({ kind: 'opaque' });
}

export function createStructuralMotionDraft(input: Readonly<{
  subjectId: string;
  presence: StructuralPresence;
  provenance: StructuralProvenance;
  from?: OffsetGeometry;
  to?: OffsetGeometry;
  viewportFrom?: ViewportGeometry;
  viewportTo?: ViewportGeometry;
  inversion?: FlipGeometry;
  beforeTransform?: string;
  afterTransform?: string;
  beforeProperties?: Readonly<Record<string, unknown>>;
  afterProperties?: Readonly<Record<string, unknown>>;
  animatingProperties?: readonly string[];
  beforeOpacity?: string;
  afterOpacity?: string;
}>): StructuralMotionDraft {
  const properties = (input.animatingProperties ?? []).flatMap(name => {
    const before = input.beforeProperties?.[name];
    const after = input.afterProperties?.[name];
    return before === after
      ? []
      : [Object.freeze({
        name,
        before: snapshotValue(before),
        after: snapshotValue(after),
      })];
  });
  const beforeTransform = input.beforeTransform ?? '';
  const afterTransform = input.afterTransform ?? '';
  const beforeOpacity = finiteOpacity(input.beforeOpacity);
  const afterOpacity = finiteOpacity(input.afterOpacity);
  return Object.freeze({
    subjectId: input.subjectId,
    presence: input.presence,
    provenance: Object.freeze({ ...input.provenance }),
    ...(input.inversion?.changed && input.viewportFrom && input.viewportTo ? {
      spatial: Object.freeze({
        ...(input.from ? { offsetFrom: input.from } : {}),
        ...(input.to ? { offsetTo: input.to } : {}),
        viewportFrom: input.viewportFrom,
        viewportTo: input.viewportTo,
        inversion: input.inversion,
      }),
    } : {}),
    ...(beforeTransform !== afterTransform ? {
      transform: Object.freeze({ before: beforeTransform, after: afterTransform }),
    } : {}),
    properties: Object.freeze(properties),
    ...(Math.abs(beforeOpacity - afterOpacity) > 0.01 ? {
      opacity: Object.freeze({ before: beforeOpacity, after: afterOpacity }),
    } : {}),
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
  const segments = entries.map(({ draft, timingRequest }) => Object.freeze({
    ...draft,
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
  updates: ReadonlyMap<string, StructuralExecution>,
): StructuralMotionPlan {
  const segments = plan.segments.map(segment => {
    const execution = updates.get(segment.subjectId);
    const frozenExecution = execution && 'animations' in execution
      ? Object.freeze({
        ...execution,
        animations: Object.freeze(execution.animations.map(timing => Object.freeze({ ...timing }))),
      })
      : execution ? Object.freeze({ ...execution }) : null;
    return execution
      ? Object.freeze({ ...segment, execution: frozenExecution! })
      : segment;
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
