type JsonRecord = Readonly<Record<string, unknown>>;

function record(value: unknown, label: string): JsonRecord {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value as JsonRecord;
}

function nonempty(value: unknown, label: string, maxLength = 512): string {
  if (typeof value !== 'string' || !value.trim()) throw new Error(`${label} must be a non-empty string`);
  if (value.length > maxLength) throw new Error(`${label} is too long`);
  return value;
}

function timestamp(value: unknown, label: string): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value <= 0) {
    throw new Error(`${label} must be a positive safe integer`);
  }
  return value;
}

function boolean(value: unknown, label: string): boolean {
  if (typeof value !== 'boolean') throw new Error(`${label} must be boolean`);
  return value;
}

function webURL(value: unknown, label: string): string {
  const result = nonempty(value, label, 4096);
  if (!(result.startsWith('/') || /^https:\/\//i.test(result) || /^http:\/\//i.test(result))) {
    throw new Error(`${label} must be a relative or HTTP(S) URL`);
  }
  return result;
}

function appURL(value: unknown, label: string): string {
  const result = nonempty(value, label, 512);
  if (!result.startsWith('/') || result.startsWith('//')) {
    throw new Error(`${label} must be a same-origin absolute path`);
  }
  return result;
}

function lowerHex(value: unknown, label: string, length: number): string {
  const result = nonempty(value, label, length);
  if (result.length !== length || !/^[0-9a-f]+$/.test(result)) {
    throw new Error(`${label} must be ${length} lowercase hexadecimal characters`);
  }
  return result;
}

function transferToken(value: unknown, label: string): string {
  const result = nonempty(value, label, 512);
  if (!/^v1\.[A-Za-z0-9_-]+\.[0-9a-f]{32}\.[0-9a-f]{64}$/.test(result)) {
    throw new Error(`${label} has an invalid transfer-token shape`);
  }
  return result;
}

function transferCode(value: unknown, label: string): string {
  const result = nonempty(value, label, 10);
  if (!/^[0-9A-HJKMNP-TV-Z]{10}$/.test(result)) {
    throw new Error(`${label} must be a ten-character Crockford code`);
  }
  return result;
}

export interface TableTransferOffer {
  readonly ok: true;
  readonly pairingID: string;
  readonly token: string;
  readonly manualCode: string;
  readonly claimURL: string;
  readonly qrDataURL: string;
  readonly expiresAtMs: number;
  readonly serverNowMs: number;
}

export interface TableTransferInspection {
  readonly ok: true;
  readonly pairingID: string;
  readonly gameID: string;
  readonly gameName: string;
  readonly gameDisplayName: string;
  readonly expiresAtMs: number;
  readonly serverNowMs: number;
  readonly alreadyRedeemed: boolean;
}

export interface TableTransferRedemption {
  readonly ok: true;
  readonly gameID: string;
  readonly gameName: string;
  readonly gameURL: string;
}

export function decodeTableTransferCancel(value: unknown): Readonly<{ ok: true }> {
  const item = record(value, 'Table transfer cancellation');
  if (item['ok'] !== true) throw new Error('Table transfer cancellation.ok must be true');
  return { ok: true };
}

export function decodeTableTransferOffer(value: unknown): TableTransferOffer {
  const item = record(value, 'Table transfer offer');
  if (item['ok'] !== true) throw new Error('Table transfer offer.ok must be true');
  const qrDataURL = nonempty(item['qrDataURL'], 'Table transfer offer.qrDataURL', 2_000_000);
  if (!qrDataURL.startsWith('data:image/png;base64,')) throw new Error('Table transfer offer.qrDataURL must be a PNG data URL');
  return {
    ok: true,
    pairingID: lowerHex(item['pairingID'], 'Table transfer offer.pairingID', 32),
    token: transferToken(item['token'], 'Table transfer offer.token'),
    manualCode: transferCode(item['manualCode'], 'Table transfer offer.manualCode'),
    claimURL: webURL(item['claimURL'], 'Table transfer offer.claimURL'),
    qrDataURL,
    expiresAtMs: timestamp(item['expiresAtMs'], 'Table transfer offer.expiresAtMs'),
    serverNowMs: timestamp(item['serverNowMs'], 'Table transfer offer.serverNowMs'),
  };
}

export function decodeTableTransferInspection(value: unknown): TableTransferInspection {
  const item = record(value, 'Table transfer inspection');
  if (item['ok'] !== true) throw new Error('Table transfer inspection.ok must be true');
  return {
    ok: true,
    pairingID: lowerHex(item['pairingID'], 'Table transfer inspection.pairingID', 32),
    gameID: nonempty(item['gameID'], 'Table transfer inspection.gameID', 128),
    gameName: nonempty(item['gameName'], 'Table transfer inspection.gameName', 128),
    gameDisplayName: nonempty(item['gameDisplayName'], 'Table transfer inspection.gameDisplayName', 256),
    expiresAtMs: timestamp(item['expiresAtMs'], 'Table transfer inspection.expiresAtMs'),
    serverNowMs: timestamp(item['serverNowMs'], 'Table transfer inspection.serverNowMs'),
    alreadyRedeemed: boolean(item['alreadyRedeemed'], 'Table transfer inspection.alreadyRedeemed'),
  };
}

export function decodeTableTransferRedemption(value: unknown): TableTransferRedemption {
  const item = record(value, 'Table transfer redemption');
  if (item['ok'] !== true) throw new Error('Table transfer redemption.ok must be true');
  return {
    ok: true,
    gameID: nonempty(item['gameID'], 'Table transfer redemption.gameID', 128),
    gameName: nonempty(item['gameName'], 'Table transfer redemption.gameName', 128),
    gameURL: appURL(item['gameURL'], 'Table transfer redemption.gameURL'),
  };
}

export type TableTransferInput =
  | Readonly<{ kind: 'token'; token: string }>
  | Readonly<{ kind: 'manual'; roomCode: string; manualCode: string }>;

/** Owns async work for one visible transfer route. Superseded work cannot
 * commit into a newly-entered token or manual-code attempt. */
export class TableTransferScope {
  private generation = 0;
  private controller: AbortController | null = null;

  begin(): Readonly<{ signal: AbortSignal; isCurrent: () => boolean }> {
    this.controller?.abort();
    const generation = ++this.generation;
    const controller = new AbortController();
    this.controller = controller;
    return {
      signal: controller.signal,
      isCurrent: () => !controller.signal.aborted && this.generation === generation,
    };
  }

  invalidate(): void {
    this.generation++;
    this.controller?.abort();
    this.controller = null;
  }
}

export function transferTokenFromFragment(fragment: string): string | null {
  const raw = fragment.startsWith('#') ? fragment.slice(1) : fragment;
  if (raw.length > 4096) return null;
  const tokens = new URLSearchParams(raw).getAll('transfer');
  if (tokens.length !== 1) return null;
  const token = tokens[0]?.trim();
  return token && token.length <= 1024 ? token : null;
}

export function transferFailureMessage(code: string | undefined, fallback?: string): string {
  switch (code) {
    case 'TABLE_TRANSFER_EXPIRED': return 'This transfer code has expired. Create a new one on the current shared Table.';
    case 'TABLE_TRANSFER_CANCELLED': return 'This transfer was cancelled on the current shared Table.';
    case 'TABLE_TRANSFER_ALREADY_REDEEMED': return 'This transfer was already used by another screen.';
    case 'TABLE_TRANSFER_INVALID': return 'That transfer link or code is not valid.';
    case 'GAME_FINISHED': return 'This game has finished and cannot move to another screen.';
    default: return fallback || 'The shared Table could not be transferred. Please try again.';
  }
}
