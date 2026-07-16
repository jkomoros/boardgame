const MAX_READINESS_PARTICIPANTS = 256;

export type ReadinessState = 'not-required' | 'ready' | 'waiting';
export type ReadinessKey = string | number;

export interface ReadinessParticipant<Key extends ReadinessKey = ReadinessKey> {
  readonly key: Key;
  readonly label: string;
  readonly state: ReadinessState;
}

export interface ReadinessPresentation<Key extends ReadinessKey = ReadinessKey> {
  readonly participants: readonly ReadinessParticipant<Key>[];
  readonly readyCount: number;
  readonly requiredCount: number;
  readonly complete: boolean;
  readonly empty: boolean;
  readonly message: string;
}

export interface ReadinessLabels {
  readonly complete?: string;
  readonly empty?: string;
  /** Phrase after “N of M”, for example “ready” or “votes cast”. */
  readonly progress?: string;
}

/** Validate and summarize creator-provided, already-sanitized public readiness. */
export function readinessPresentation<Key extends ReadinessKey>(
  participants: readonly ReadinessParticipant<Key>[],
  labels: ReadinessLabels = {},
): ReadinessPresentation<Key> {
  if (!Array.isArray(participants)) {
    throw new Error('boardgame-readiness: participants must be an array');
  }
  if (participants.length > MAX_READINESS_PARTICIPANTS) {
    throw new Error(`boardgame-readiness: participants exceeds the maximum of ${MAX_READINESS_PARTICIPANTS}`);
  }
  const completeLabel = optionalLabel('completeLabel', labels.complete);
  const emptyLabel = optionalLabel('emptyLabel', labels.empty);
  const progressLabel = optionalLabel('progressLabel', labels.progress) || 'ready';
  const seen = new Set<ReadinessKey>();
  const copy = participants.map((participant, index) => {
    if (!participant || typeof participant !== 'object') {
      throw new Error(`boardgame-readiness: participant at index ${index} must be an object`);
    }
    const key: Key = participant.key;
    const label: string = participant.label;
    const state: ReadinessState = participant.state;
    if ((typeof key !== 'string' && typeof key !== 'number')
      || (typeof key === 'number' && !Number.isFinite(key))
      || (typeof key === 'string' && !key.trim())) {
      throw new Error(`boardgame-readiness: participant at index ${index} has an invalid key`);
    }
    if (seen.has(key)) {
      throw new Error(`boardgame-readiness: duplicate participant key ${JSON.stringify(key)}`);
    }
    seen.add(key);
    if (typeof label !== 'string' || !label.trim()) {
      throw new Error(`boardgame-readiness: participant ${JSON.stringify(key)} needs a non-empty label`);
    }
    if (state !== 'ready' && state !== 'waiting' && state !== 'not-required') {
      throw new Error(`boardgame-readiness: participant ${JSON.stringify(key)} has unknown state ${JSON.stringify(state)}`);
    }
    return Object.freeze({ key, label: label.trim(), state });
  });
  const requiredCount = copy.filter(participant => participant.state !== 'not-required').length;
  const readyCount = copy.filter(participant => participant.state === 'ready').length;
  const empty = requiredCount === 0;
  const complete = !empty && readyCount === requiredCount;
  const message = empty
    ? (emptyLabel || 'No participants are required')
    : complete
      ? (completeLabel || `All ${requiredCount} ready`)
      : `${readyCount} of ${requiredCount} ${progressLabel}`;
  return Object.freeze({
    participants: Object.freeze(copy),
    readyCount,
    requiredCount,
    complete,
    empty,
    message,
  });
}

function optionalLabel(name: string, value: string | undefined): string {
  if (value === undefined || value === '') return '';
  if (typeof value !== 'string' || !value.trim()) {
    throw new Error(`boardgame-readiness: ${name} must be omitted or a non-empty string`);
  }
  return value.trim();
}
