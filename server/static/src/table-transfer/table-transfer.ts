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

function currentOrigin(): string {
  return typeof globalThis.location?.origin === 'string'
    ? globalThis.location.origin
    : 'https://boardgame.invalid';
}

function sameOriginURL(value: unknown, label: string, maxLength: number): URL {
  const raw = nonempty(value, label, maxLength);
  let parsed: URL;
  try {
    parsed = new URL(raw, currentOrigin());
  } catch {
    throw new Error(`${label} must be a valid URL`);
  }
  if (parsed.origin !== currentOrigin() || parsed.username || parsed.password) {
    throw new Error(`${label} must be on the application origin`);
  }
  return parsed;
}

function webURL(value: unknown, label: string): string {
  const result = nonempty(value, label, 4096);
  return sameOriginURL(result, label, 4096).href;
}

function appURL(value: unknown, label: string): string {
  const parsed = sameOriginURL(value, label, 512);
  if (!parsed.pathname.startsWith('/game/') || parsed.hash) {
    throw new Error(`${label} must be a same-origin game path`);
  }
  return parsed.pathname + parsed.search;
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
  const pairingID = lowerHex(item['pairingID'], 'Table transfer offer.pairingID', 32);
  const token = transferToken(item['token'], 'Table transfer offer.token');
  const claimURL = webURL(item['claimURL'], 'Table transfer offer.claimURL');
  const expiresAtMs = timestamp(item['expiresAtMs'], 'Table transfer offer.expiresAtMs');
  const serverNowMs = timestamp(item['serverNowMs'], 'Table transfer offer.serverNowMs');
  if (expiresAtMs <= serverNowMs) throw new Error('Table transfer offer must expire in the future');
  if (token.split('.')[2] !== pairingID) throw new Error('Table transfer offer token does not match pairingID');
  const claim = new URL(claimURL);
  if (claim.pathname !== '/table' || claim.search || claim.hash !== `#transfer=${encodeURIComponent(token)}`) {
    throw new Error('Table transfer offer claimURL does not contain its exact token');
  }
  return {
    ok: true,
    pairingID,
    token,
    manualCode: transferCode(item['manualCode'], 'Table transfer offer.manualCode'),
    claimURL,
    qrDataURL,
    expiresAtMs,
    serverNowMs,
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

const pendingClaimKey = 'boardgame-table-transfer-pending-v1';
const pendingClaimMaxAgeMs = 10 * 60 * 1000;

export interface PendingTableTransferClaim {
  readonly input: TableTransferInput;
  readonly confirmed: boolean;
}

export function rememberPendingTableTransfer(input: TableTransferInput, confirmed: boolean): void {
  try {
    globalThis.sessionStorage?.setItem(pendingClaimKey, JSON.stringify({ version: 1, savedAtMs: Date.now(), input, confirmed }));
  } catch { /* An in-memory attempt still works when storage is unavailable. */ }
}

export function restorePendingTableTransfer(): PendingTableTransferClaim | null {
  try {
    const raw = globalThis.sessionStorage?.getItem(pendingClaimKey);
    if (!raw) return null;
    const item = record(JSON.parse(raw), 'Pending Table transfer');
    if (item['version'] !== 1 || typeof item['savedAtMs'] !== 'number' ||
      !Number.isSafeInteger(item['savedAtMs']) || Date.now() - item['savedAtMs'] > pendingClaimMaxAgeMs ||
      Date.now() < item['savedAtMs'] - 60_000 || typeof item['confirmed'] !== 'boolean') {
      clearPendingTableTransfer();
      return null;
    }
    const input = record(item['input'], 'Pending Table transfer.input');
    if (input['kind'] === 'token') {
      return { input: { kind: 'token', token: transferToken(input['token'], 'Pending Table transfer token') }, confirmed: item['confirmed'] };
    }
    if (input['kind'] === 'manual') {
      return {
        input: {
          kind: 'manual',
          roomCode: nonempty(input['roomCode'], 'Pending Table transfer room code', 16),
          manualCode: transferCode(input['manualCode'], 'Pending Table transfer manual code'),
        },
        confirmed: item['confirmed'],
      };
    }
  } catch { /* Corrupt or unavailable storage fails closed below. */ }
  clearPendingTableTransfer();
  return null;
}

export function clearPendingTableTransfer(): void {
  try { globalThis.sessionStorage?.removeItem(pendingClaimKey); } catch { /* ignored */ }
}

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
    case 'TABLE_TRANSFER_BUSY': return 'The shared Table is finishing an action. Try again in a moment.';
    case 'TABLE_TRANSFER_INVALID': return 'That transfer link or code is not valid.';
    case 'GAME_FINISHED': return 'This game has finished and cannot move to another screen.';
    default: return fallback || 'The shared Table could not be transferred. Please try again.';
  }
}
