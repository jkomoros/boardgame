export interface MotionExactSighting {
  readonly subjectId: string;
  readonly collectionId: string;
}

export interface MotionCollectionHistory {
  readonly collectionId: string;
  readonly lastSeen: Readonly<Record<string, number>>;
}

export type MotionContinuityEndpoint =
  | Readonly<{ kind: 'subject'; phase: 'before' | 'after'; collectionId: string }>
  | Readonly<{ kind: 'collection'; collectionId: string }>;

export type MotionContinuityResolution =
  | Readonly<{
    status: 'resolved';
    subjectId: string;
    presence: 'retained' | 'appearing' | 'departing';
    from: MotionContinuityEndpoint;
    to: MotionContinuityEndpoint;
    evidence: 'identity' | 'history';
  }>
  | Readonly<{
    status: 'unresolved';
    subjectId: string;
    endpoint: 'source' | 'destination' | 'identity';
    reason: 'absent-both-sides' | 'duplicate-exact-sighting'
      | 'missing-history' | 'ambiguous-history' | 'invalid-history';
  }>;

function subjectEndpoint(
  phase: 'before' | 'after',
  collectionId: string,
): MotionContinuityEndpoint {
  return Object.freeze({ kind: 'subject', phase, collectionId });
}

function collectionEndpoint(collectionId: string): MotionContinuityEndpoint {
  return Object.freeze({ kind: 'collection', collectionId });
}

function unresolved(
  subjectId: string,
  endpoint: 'source' | 'destination' | 'identity',
  reason: Extract<MotionContinuityResolution, { status: 'unresolved' }>['reason'],
): MotionContinuityResolution {
  return Object.freeze({ status: 'unresolved', subjectId, endpoint, reason });
}

/**
 * Resolve logical continuity without retaining DOM, geometry, or history
 * versions. Exact sightings dominate history. Inferred endpoints are selected
 * only when the highest remaining history version names exactly one collection.
 */
export function resolveStructuralContinuity(
  subjectId: string,
  beforeExact: readonly MotionExactSighting[],
  afterExact: readonly MotionExactSighting[],
  histories: readonly MotionCollectionHistory[],
): MotionContinuityResolution {
  const before = beforeExact.filter(sighting => sighting.subjectId === subjectId);
  const after = afterExact.filter(sighting => sighting.subjectId === subjectId);
  if (before.length > 1 || after.length > 1) {
    return unresolved(subjectId, 'identity', 'duplicate-exact-sighting');
  }
  if (before.length === 1 && after.length === 1) {
    return Object.freeze({
      status: 'resolved',
      subjectId,
      presence: 'retained',
      from: subjectEndpoint('before', before[0].collectionId),
      to: subjectEndpoint('after', after[0].collectionId),
      evidence: 'identity',
    });
  }
  if (before.length === 0 && after.length === 0) {
    return unresolved(subjectId, 'identity', 'absent-both-sides');
  }

  const endpoint = before.length === 0 ? 'source' as const : 'destination' as const;
  const exactCollectionId = (before[0] ?? after[0]).collectionId;
  const candidates: Array<{ collectionId: string; version: number }> = [];
  let invalid = false;
  const collectionIds = new Set<string>();
  for (const history of histories) {
    if (!history.collectionId || collectionIds.has(history.collectionId)) {
      invalid = true;
      continue;
    }
    collectionIds.add(history.collectionId);
    const version = history.lastSeen[subjectId];
    if (version === undefined) continue;
    if (!Number.isSafeInteger(version) || version < 0) {
      invalid = true;
      continue;
    }
    if (history.collectionId !== exactCollectionId) {
      candidates.push({ collectionId: history.collectionId, version });
    }
  }
  if (invalid) return unresolved(subjectId, endpoint, 'invalid-history');
  if (candidates.length === 0) return unresolved(subjectId, endpoint, 'missing-history');
  const highestVersion = Math.max(...candidates.map(candidate => candidate.version));
  const highest = candidates.filter(candidate => candidate.version === highestVersion);
  if (highest.length !== 1) return unresolved(subjectId, endpoint, 'ambiguous-history');
  const inferred = collectionEndpoint(highest[0].collectionId);

  return before.length === 0
    ? Object.freeze({
      status: 'resolved',
      subjectId,
      presence: 'appearing',
      from: inferred,
      to: subjectEndpoint('after', after[0].collectionId),
      evidence: 'history',
    })
    : Object.freeze({
      status: 'resolved',
      subjectId,
      presence: 'departing',
      from: subjectEndpoint('before', before[0].collectionId),
      to: inferred,
      evidence: 'history',
    });
}
