const MAX_FRAME_CHARACTERS = 4_096;
const MAX_TEXT_CHARACTERS = 1_024;
const MAX_DURATION_MS = 60_000;

type RecordValue = Readonly<Record<string, unknown>>;

export interface VersionTimingMessage {
  version: number;
  serverSentAt: number;
  serverPlayAt: number;
  slotDurationMs: number;
  maxAnimationDurationMs: number;
}

export interface ClockSyncMessage {
  clientSentAt: number;
  serverAt: number;
}

function record(value: unknown, path: string): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${path} must be an object`);
  }
  return value as RecordValue;
}

function string(value: unknown, path: string): string {
  if (typeof value !== 'string' || !value.trim() || value.length > MAX_TEXT_CHARACTERS) {
    throw new Error(`${path} must be a bounded non-empty string`);
  }
  return value;
}

function nonNegativeInteger(value: unknown, path: string): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) {
    throw new Error(`${path} must be a non-negative safe integer`);
  }
  return value;
}

function positiveFinite(value: unknown, path: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0 || value > MAX_DURATION_MS) {
    throw new Error(`${path} must be a positive finite duration no greater than ${MAX_DURATION_MS}ms`);
  }
  return value;
}

function nonNegativeFinite(value: unknown, path: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0 || value > MAX_DURATION_MS) {
    throw new Error(`${path} must be a non-negative finite duration no greater than ${MAX_DURATION_MS}ms`);
  }
  return value;
}

export type SocketFrame =
  | { type: 'version'; version: number; transport: 'json' | 'legacy' }
  | { type: 'version-timing'; timing: VersionTimingMessage }
  | { type: 'clock-sync'; clock: ClockSyncMessage }
  | { type: 'mode-changed'; gameID: string; newMode: 'solo' }
  | { type: 'presence-changed'; gameID: string }
  | { type: 'chat'; channel: string; messageID: string }
  | { type: 'unknown'; wireType: string };

function decodeJsonFrame(value: unknown): SocketFrame {
  const frame = record(value, 'Socket frame');
  const type = string(frame['type'], 'Socket frame.type');
  if (type === 'version') {
    return { type, version: nonNegativeInteger(frame['data'], 'Socket frame.data'), transport: 'json' };
  }
  if (type === 'version-timing') {
    const data = record(frame['data'], 'Socket frame.data');
    const slotDurationMs = positiveFinite(data['slotDurationMs'], 'Socket frame.data.slotDurationMs');
    const maxAnimationDurationMs = nonNegativeFinite(
      data['maxAnimationDurationMs'],
      'Socket frame.data.maxAnimationDurationMs',
    );
    if (maxAnimationDurationMs > slotDurationMs) {
      throw new Error('Socket frame.data.maxAnimationDurationMs must not exceed slotDurationMs');
    }
    const serverSentAt = nonNegativeInteger(data['serverSentAt'], 'Socket frame.data.serverSentAt');
    const serverPlayAt = nonNegativeInteger(data['serverPlayAt'], 'Socket frame.data.serverPlayAt');
    if (serverPlayAt < serverSentAt) {
      throw new Error('Socket frame.data.serverPlayAt must not precede serverSentAt');
    }
    return {
      type,
      timing: {
        version: nonNegativeInteger(data['version'], 'Socket frame.data.version'),
        serverSentAt,
        serverPlayAt,
        slotDurationMs,
        maxAnimationDurationMs,
      },
    };
  }
  if (type === 'clock-sync') {
    const data = record(frame['data'], 'Socket frame.data');
    return {
      type,
      clock: {
        clientSentAt: nonNegativeInteger(data['clientSentAt'], 'Socket frame.data.clientSentAt'),
        serverAt: nonNegativeInteger(data['serverAt'], 'Socket frame.data.serverAt'),
      },
    };
  }
  if (type === 'mode-changed') {
    const data = record(frame['data'], 'Socket frame.data');
    if (data['newMode'] !== 'solo') throw new Error('Socket frame.data.newMode must be "solo"');
    return {
      type,
      gameID: string(data['gameID'], 'Socket frame.data.gameID'),
      newMode: 'solo',
    };
  }
  if (type === 'presence-changed') {
    const data = record(frame['data'], 'Socket frame.data');
    return { type, gameID: string(data['gameID'], 'Socket frame.data.gameID') };
  }
  if (type === 'chat') {
    const data = record(frame['data'], 'Socket frame.data');
    return {
      type,
      channel: string(data['channel'], 'Socket frame.data.channel'),
      messageID: string(data['messageId'], 'Socket frame.data.messageId'),
    };
  }
  return { type: 'unknown', wireType: type };
}

export function decodeSocketFrame(value: unknown): SocketFrame {
  if (typeof value !== 'string' || value.length === 0 || value.length > MAX_FRAME_CHARACTERS) {
    throw new Error(`Socket frame must be a non-empty string of at most ${MAX_FRAME_CHARACTERS} characters`);
  }
  if (!value.startsWith('{')) {
    if (!/^(0|[1-9]\d*)$/.test(value)) throw new Error('Legacy socket frame must be an exact version integer');
    const version = Number(value);
    if (!Number.isSafeInteger(version)) throw new Error('Legacy socket frame version must be a safe integer');
    return { type: 'version', version, transport: 'legacy' };
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(value) as unknown;
  } catch {
    throw new Error('Socket frame must contain valid JSON');
  }
  return decodeJsonFrame(parsed);
}
