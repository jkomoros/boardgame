type RecordValue = Readonly<Record<string, unknown>>;

export const TABLE_LEASE_FAILURE_CODES = [
  'TABLE_LEASE_ACTIVE',
  'TABLE_LEASE_NOT_ELIGIBLE',
  'GAME_NOT_COMPANION',
  'GAME_FINISHED',
	'GAME_NOT_FOUND',
	'TABLE_LEASE_STORAGE',
	'TABLE_LEASE_RANDOM',
	'TABLE_LEASE_ALREADY_RESTORED',
	'TABLE_LEASE_INVALID_REQUEST',
] as const;

export type TableLeaseFailureCode = typeof TABLE_LEASE_FAILURE_CODES[number];

export interface TableLeaseAcquireResponse {
  ok: true;
  alreadyHeld: boolean;
  expiresAtMs: number;
}

function record(value: unknown): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Table lease response must be an object');
  }
  return value as RecordValue;
}

export function decodeTableLeaseAcquireResponse(value: unknown): TableLeaseAcquireResponse {
  const item = record(value);
  if (item['ok'] !== true) throw new Error('Table lease response.ok must be true');
  if (typeof item['alreadyHeld'] !== 'boolean') {
    throw new Error('Table lease response.alreadyHeld must be a boolean');
  }
  if (typeof item['expiresAtMs'] !== 'number'
    || !Number.isSafeInteger(item['expiresAtMs'])
    || item['expiresAtMs'] <= 0) {
    throw new Error('Table lease response.expiresAtMs must be a positive safe integer');
  }
  return {
    ok: true,
    alreadyHeld: item['alreadyHeld'],
    expiresAtMs: item['expiresAtMs'],
  };
}

export function isTableLeaseFailureCode(value: string | undefined): value is TableLeaseFailureCode {
  return value !== undefined
    && (TABLE_LEASE_FAILURE_CODES as readonly string[]).includes(value);
}

export function tableLeaseFailureMessage(code: TableLeaseFailureCode): string {
  switch (code) {
    case 'TABLE_LEASE_ACTIVE':
      return 'Another shared screen is still active. This page will offer takeover if it does not reconnect.';
    case 'TABLE_LEASE_NOT_ELIGIBLE':
      return 'Only a seated player can take over the shared Table.';
    case 'GAME_NOT_COMPANION':
      return 'This game is no longer using companion mode.';
    case 'GAME_FINISHED':
      return 'This game has finished, so its shared Table cannot be taken over.';
	case 'GAME_NOT_FOUND':
		return 'This game no longer exists.';
	case 'TABLE_LEASE_STORAGE':
	case 'TABLE_LEASE_RANDOM':
		return 'The server could not safely restore the shared Table. Please try again.';
	case 'TABLE_LEASE_ALREADY_RESTORED':
		return 'Another player restored the shared Table first.';
	case 'TABLE_LEASE_INVALID_REQUEST':
		return 'This browser could not create a safe Table recovery request. Reload and try again.';
  }
}
