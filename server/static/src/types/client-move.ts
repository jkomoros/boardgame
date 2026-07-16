import type { ClientMove } from './api.js';

const MAX_CLIENT_MOVE_NAME_LENGTH = 256;

/** Validate untrusted bundle metadata and copy only the animation-safe fields. */
export function clientMoveFromWire(value: unknown): ClientMove | null {
  if (value === null) return null;
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Client move metadata must be an object or null');
  }
  const record = value as Readonly<Record<string, unknown>>;
  if (typeof record.Name !== 'string' || !record.Name.trim()) {
    throw new Error('Client move metadata Name must be a non-empty string');
  }
  if (record.Name.length > MAX_CLIENT_MOVE_NAME_LENGTH) {
    throw new Error(`Client move metadata Name exceeds ${MAX_CLIENT_MOVE_NAME_LENGTH} characters`);
  }
  if (!Number.isSafeInteger(record.Version) || (record.Version as number) < 0) {
    throw new Error('Client move metadata Version must be a non-negative safe integer');
  }
  return Object.freeze({ Name: record.Name, Version: record.Version as number });
}
