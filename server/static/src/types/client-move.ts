import type { ClientMove, JsonValue } from './api.js';

const MAX_ANIMATION_KEY_LENGTH = 256;

function isJsonValue(value: unknown): boolean {
  if (value === null || typeof value === 'string' || typeof value === 'boolean') return true;
  if (typeof value === 'number') return Number.isFinite(value);
  if (Array.isArray(value)) return value.every(isJsonValue);
  if (!value || typeof value !== 'object') return false;
  const prototype = Object.getPrototypeOf(value);
  if (prototype !== Object.prototype && prototype !== null) return false;
  return Object.values(value as Readonly<Record<string, unknown>>).every(isJsonValue);
}

function copyJsonValue(value: JsonValue): JsonValue {
  if (value === null || typeof value !== 'object') return value;
  if (Array.isArray(value)) return Object.freeze(value.map(copyJsonValue)) as unknown as JsonValue;
  return Object.freeze(Object.fromEntries(
    Object.entries(value).map(([key, child]) => [key, copyJsonValue(child)]),
  ));
}

/** Validate untrusted bundle metadata and copy only the animation-safe fields. */
export function clientMoveFromWire(value: unknown): ClientMove | null {
  if (value === null) return null;
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Client move metadata must be an object or null');
  }
  const record = value as Readonly<Record<string, unknown>>;
  if (typeof record.AnimationKey !== 'string' || !record.AnimationKey.trim()) {
    throw new Error('Client move metadata AnimationKey must be a non-empty string');
  }
  if (record.AnimationKey.length > MAX_ANIMATION_KEY_LENGTH) {
    throw new Error(`Client move metadata AnimationKey exceeds ${MAX_ANIMATION_KEY_LENGTH} characters`);
  }
  if (!Number.isSafeInteger(record.Version) || (record.Version as number) < 0) {
    throw new Error('Client move metadata Version must be a non-negative safe integer');
  }
  let properties: Readonly<Record<string, JsonValue>> | undefined;
  if (record.Properties !== undefined) {
    if (!record.Properties || typeof record.Properties !== 'object' || Array.isArray(record.Properties) || !isJsonValue(record.Properties)) {
      throw new Error('Client move metadata Properties must contain only JSON values');
    }
    properties = copyJsonValue(record.Properties as Record<string, JsonValue>) as Readonly<Record<string, JsonValue>>;
  }
  return Object.freeze({
    AnimationKey: record.AnimationKey,
    Version: record.Version as number,
    ...(properties ? { Properties: properties } : {}),
  });
}
