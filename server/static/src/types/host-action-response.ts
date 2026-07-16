type RecordValue = Readonly<Record<string, unknown>>;

function record(value: unknown): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Host action response must be an object');
  }
  return value as RecordValue;
}

export interface HostActionResponse {
  ok: true;
  locked?: boolean;
}

/** Copies the small conventional-HTTP success contract used by host actions. */
export function decodeHostActionResponse(value: unknown): HostActionResponse {
  const item = record(value);
  if (item['ok'] !== true) throw new Error('Host action response.ok must be true');
  if (item['locked'] !== undefined && typeof item['locked'] !== 'boolean') {
    throw new Error('Host action response.locked must be a boolean when provided');
  }
  return {
    ok: true,
    ...(typeof item['locked'] === 'boolean' ? { locked: item['locked'] } : {}),
  };
}
