import type { FlipGeometry, OffsetGeometry } from './geometry.js';
import type { AnimationTimingPolicy } from './timing.js';

export type StructuralPresence = 'retained' | 'appearing' | 'departing';

export type StructuralProvenance =
  | Readonly<{ kind: 'identity' }>
  | Readonly<{
    kind: 'stack-history';
    endpoint: 'source' | 'destination';
    stackId: string;
    evidence: 'only-candidate' | 'runner-up' | 'latest-seen';
  }>;

export interface StructuralPropertyChange {
  readonly name: string;
  readonly before: unknown;
  readonly after: unknown;
}

export interface StructuralSpatialChange {
  readonly from: OffsetGeometry;
  readonly to: OffsetGeometry;
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

export interface StructuralMotionSegment extends StructuralMotionDraft {
  readonly timingRequest: StructuralTimingRequest;
}

export interface StructuralMotionPlan {
  readonly generation: number;
  readonly phase: 'ready-to-play';
  readonly segments: readonly StructuralMotionSegment[];
}

function finiteOpacity(value: string | undefined): number {
  const parsed = Number.parseFloat(value || '1');
  return Number.isFinite(parsed) ? parsed : 1;
}

export function createStructuralMotionDraft(input: Readonly<{
  subjectId: string;
  presence: StructuralPresence;
  provenance: StructuralProvenance;
  from: OffsetGeometry;
  to: OffsetGeometry;
  inversion: FlipGeometry;
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
      : [Object.freeze({ name, before, after })];
  });
  const beforeTransform = input.beforeTransform ?? '';
  const afterTransform = input.afterTransform ?? '';
  const beforeOpacity = finiteOpacity(input.beforeOpacity);
  const afterOpacity = finiteOpacity(input.afterOpacity);
  return Object.freeze({
    subjectId: input.subjectId,
    presence: input.presence,
    provenance: Object.freeze({ ...input.provenance }),
    ...(input.inversion.changed ? {
      spatial: Object.freeze({
        from: input.from,
        to: input.to,
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
): StructuralMotionPlan {
  const segments = entries.map(({ draft, timingRequest }) => Object.freeze({
    ...draft,
    timingRequest: Object.freeze({ ...timingRequest }),
  }));
  return Object.freeze({
    generation,
    phase: 'ready-to-play',
    segments: Object.freeze(segments),
  });
}
