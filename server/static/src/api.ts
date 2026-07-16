/**
 * Typed API utilities for making HTTP requests to the boardgame server.
 * Replaces iron-ajax/boardgame-ajax with modern fetch-based approach.
 */

import type { MoveForm } from './types/api';

// API_HOST is defined in index.html
declare const API_HOST: string;

/**
 * Standard API response wrapper
 */
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  friendlyError?: string;
  code?: string;
  expectedVersion?: number;
  actualVersion?: number;
  status: number;
}

function responseRecord(value: unknown): Readonly<Record<string, unknown>> | null {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
    ? value as Readonly<Record<string, unknown>>
    : null;
}

function optionalVersion(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0
    ? value
    : undefined;
}

async function unwrapApiResponse<T>(response: Response): Promise<ApiResponse<T>> {
  const status = response.status;
  let parsed: unknown;
  try {
    parsed = await response.json();
  } catch {
    return {
      status,
      error: `HTTP ${status}: ${response.statusText}`,
      friendlyError: 'The server returned an invalid response',
    };
  }

  const envelope = responseRecord(parsed);
  if (!envelope) {
    return {
      status,
      error: 'Invalid API response: expected an object envelope',
      friendlyError: 'The server returned an invalid response',
    };
  }

  if (envelope['Status'] === 'Success') {
    return { status, data: parsed as T };
  }
  if (envelope['Status'] !== 'Failure') {
    return {
      status,
      error: 'Invalid API response: Status must be "Success" or "Failure"',
      friendlyError: 'The server returned an invalid response',
    };
  }

  return {
    status,
    error: typeof envelope['Error'] === 'string' && envelope['Error']
      ? envelope['Error']
      : `Request failed with status ${status}`,
    friendlyError: typeof envelope['FriendlyError'] === 'string' && envelope['FriendlyError']
      ? envelope['FriendlyError']
      : 'An error occurred',
    code: typeof envelope['Code'] === 'string' ? envelope['Code'] : undefined,
    expectedVersion: optionalVersion(envelope['ExpectedVersion']),
    actualVersion: optionalVersion(envelope['ActualVersion']),
  };
}

async function unwrapHttpJsonResponse(response: Response): Promise<ApiResponse<unknown>> {
  const status = response.status;
  let parsed: unknown;
  try {
    parsed = await response.json();
  } catch {
    return {
      status,
      error: `HTTP ${status}: ${response.statusText}`,
      friendlyError: 'The server returned an invalid response',
    };
  }
  if (status >= 200 && status < 300) return { status, data: parsed };
  const body = responseRecord(parsed);
  return {
    status,
    error: body && typeof body['error'] === 'string' && body['error'].trim()
      ? body['error']
      : `Request failed with status ${status}`,
    friendlyError: 'The request could not be completed',
  };
}

export interface HttpJsonOptions {
  headers?: Readonly<Record<string, string>>;
  signal?: AbortSignal;
}

async function apiHttpJson(
  method: 'GET' | 'POST',
  url: string,
  body: Readonly<Record<string, unknown>> | undefined,
  options: HttpJsonOptions,
): Promise<ApiResponse<unknown>> {
  try {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...options.headers,
    };
    if (body) headers['Content-Type'] = 'application/json';
    const response = await fetch(url, {
      method,
      credentials: 'include',
      headers,
      ...(body ? { body: JSON.stringify(body) } : {}),
      ...(options.signal ? { signal: options.signal } : {}),
    });
    return await unwrapHttpJsonResponse(response);
  } catch (error) {
    return {
      status: 0,
      error: error instanceof Error ? error.message : 'Network error',
      friendlyError: 'Unable to connect to the server',
    };
  }
}

// apiHttpGet/apiHttpPost are for conventional HTTP endpoints whose successful
// payload is the JSON body and whose failures use status codes + {error}. Most
// legacy boardgame API endpoints instead use the Status envelope and should
// continue to use apiGet/apiPost; keeping the names distinct prevents a payload
// from silently being interpreted under the wrong protocol.
export function apiHttpGet(url: string, options: HttpJsonOptions = {}): Promise<ApiResponse<unknown>> {
  return apiHttpJson('GET', url, undefined, options);
}

export function apiHttpPost(
  url: string,
  body: Readonly<Record<string, unknown>>,
  options: HttpJsonOptions = {},
): Promise<ApiResponse<unknown>> {
  return apiHttpJson('POST', url, body, options);
}

/**
 * Gets the base API URL, respecting API_HOST configuration
 */
function getBaseUrl(): string {
  const host = typeof API_HOST !== 'undefined' ? API_HOST : '';
  return host + '/api/';
}

/**
 * Constructs a URL for game-agnostic API endpoints
 * @param path - API path (e.g., "auth", "list")
 * @param params - Optional query parameters
 * @returns Full API URL
 *
 * @example
 * buildApiUrl('auth') // "/api/auth"
 * buildApiUrl('list', { type: 'active' }) // "/api/list?type=active"
 */
export function buildApiUrl(path: string, params?: Record<string, string | number | boolean>): string {
  const base = getBaseUrl() + path;
  if (!params) return base;

  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    query.append(key, String(value));
  });

  const queryString = query.toString();
  return queryString ? `${base}?${queryString}` : base;
}

/**
 * Constructs a URL for game-specific API endpoints
 * @param gameName - Game type (e.g., "blackjack", "memory")
 * @param gameId - Game instance ID
 * @param path - Game-specific path (e.g., "move", "info", "configure")
 * @param params - Optional query parameters
 * @returns Full game API URL
 *
 * @example
 * buildGameUrl('memory', '123', 'info') // "/api/game/memory/123/info"
 * buildGameUrl('blackjack', '456', 'move', { player: 0 }) // "/api/game/blackjack/456/move?player=0"
 */
export function buildGameUrl(
  gameName: string,
  gameId: string,
  path: string,
  params?: Record<string, string | number | boolean>
): string {
  const base = `${getBaseUrl()}game/${gameName}/${gameId}/${path}`;
  if (!params) return base;

  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    query.append(key, String(value));
  });

  const queryString = query.toString();
  return queryString ? `${base}?${queryString}` : base;
}

/**
 * Performs a GET request to the API
 * @param url - Full API URL (use buildApiUrl or buildGameUrl)
 * @param signal - Optional cancellation signal for lifecycle-bound reads
 * @returns Typed API response
 *
 * @example
 * const response = await apiGet<GameInfoResponse>(
 *   buildGameUrl('memory', '123', 'info', { player: 0 })
 * );
 * if (response.data) {
 *   console.log(response.data.Game);
 * }
 */
export async function apiGet<T>(url: string, signal?: AbortSignal): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(url, {
      method: 'GET',
      credentials: 'include', // Matches iron-ajax withCredentials: true
      headers: {
        'Accept': 'application/json',
      },
      signal,
    });

    return await unwrapApiResponse<T>(response);
  } catch (error) {
    // Network errors
    return {
      status: 0,
      error: error instanceof Error ? error.message : 'Network error',
      friendlyError: 'Unable to connect to the server',
    };
  }
}

/**
 * Performs a POST request to the API
 * @param url - Full API URL (use buildApiUrl or buildGameUrl)
 * @param body - Request body (will be JSON-encoded or form-encoded)
 * @param contentType - Content type ('application/json' or 'application/x-www-form-urlencoded')
 * @returns Typed API response
 *
 * @example
 * const response = await apiPost<ConfigureResponse>(
 *   buildGameUrl('memory', '123', 'configure'),
 *   { open: 1, visible: 1, admin: 0 },
 *   'application/x-www-form-urlencoded'
 * );
 */
export async function apiPost<T>(
  url: string,
  body: Readonly<Record<string, unknown>>,
  contentType: 'application/json' | 'application/x-www-form-urlencoded' = 'application/json',
  signal?: AbortSignal,
): Promise<ApiResponse<T>> {
  try {
    let requestBody: string;
    const headers: Record<string, string> = {
      'Accept': 'application/json',
    };

    if (contentType === 'application/x-www-form-urlencoded') {
      // Form-encoded body
      const formData = new URLSearchParams();
      Object.entries(body).forEach(([key, value]) => {
        formData.append(key, String(value));
      });
      requestBody = formData.toString();
      headers['Content-Type'] = 'application/x-www-form-urlencoded';
    } else {
      // JSON body
      requestBody = JSON.stringify(body);
      headers['Content-Type'] = 'application/json';
    }

    const response = await fetch(url, {
      method: 'POST',
      credentials: 'include', // Matches iron-ajax withCredentials: true
      headers,
      body: requestBody,
      signal,
    });

    return await unwrapApiResponse<T>(response);
  } catch (error) {
    // Network errors
    return {
      status: 0,
      error: error instanceof Error ? error.message : 'Network error',
      friendlyError: 'Unable to connect to the server',
    };
  }
}

/**
 * Response from the single-move legality preview endpoint: the same moveForm the
 * /info endpoint ships, with its legality fields
 * (LegalForPlayer/LegalForPlayerError/Preconditions) computed against the
 * current state for the given args.
 */
export interface MovePreviewResponse {
  Form: MoveForm;
}

/**
 * Asks the server whether one move — with the given field args — is currently
 * legal for the requesting player, WITHOUT applying it. Returns the same
 * moveForm shape (legality booleans + the advisory Preconditions ledger) the
 * /info forms carry, so a client can show enabled/disabled + a reason as the
 * player edits a move's args. Side-effect-free on the server, so safe to call on
 * every keystroke. Sends form-encoded args (MoveType plus one entry per field),
 * matching the real move endpoint's parsing.
 *
 * Intentionally DIFFERENT from movePreviewBatch, which is not accidental drift:
 * this single primitive is the RICH one — it returns the full moveForm including
 * the Preconditions ledger, for a focused "explain why this exact move is/isn't
 * legal as you edit its fields" UI (the intended consumer; none ships yet, so
 * the board graying uses the batch). movePreviewBatch is the COMPACT one —
 * {Legal, Error} per candidate, nested Args JSON — for graying many board cells
 * in one round-trip, where the ledger would be dead weight ×N. Generated
 * contracts reject reserved proposal-protocol field names, and the selector is
 * written after creator arguments as defense in depth.
 *
 * @param args - field name -> raw string value, exactly as the move form submits
 * @param params - optional query params (e.g. { player } to preview as a
 *   specific seat); omit to preview as the session's player, like submitMove
 */
export async function movePreview(
  gameName: string,
  gameId: string,
  moveType: string,
  args: Record<string, string> = {},
  params?: Record<string, string | number | boolean>,
  signal?: AbortSignal,
): Promise<ApiResponse<MovePreviewResponse>> {
  return apiPost<MovePreviewResponse>(
    buildGameUrl(gameName, gameId, 'movePreview', params),
    { ...args, MoveType: moveType },
    'application/x-www-form-urlencoded',
    signal,
  );
}

/**
 * One arg-set to evaluate in a batch preview: the field values (fieldName ->
 * string) to bind before checking legality.
 */
export interface MovePreviewCandidate {
  /** Opaque correlation token echoed by the server. */
  ID?: string;
  Args: Record<string, string>;
}

/**
 * One candidate's legality, returned in the same order as the request's
 * candidates.
 */
export interface MovePreviewBatchResult {
  /** Opaque correlation token from the corresponding candidate. */
  ID?: string;
  Legal: boolean;
  Error?: string;
}

/**
 * Response from the batch legality preview endpoint: one result per requested
 * candidate, in order.
 */
export interface MovePreviewBatchResponse {
  Results: MovePreviewBatchResult[];
}

/**
 * Evaluates many candidate arg-sets of ONE move type in a single round-trip —
 * the call that lets a board gray all its illegal targets at once instead of one
 * request per cell. Results come back in candidate order (correlate by index).
 * Sends JSON {MoveType, Candidates}. Never applies anything.
 *
 * @param candidates - the arg-sets to test, e.g. one per board cell
 * @param params - optional query params (e.g. { player }); omit to preview as
 *   the session's player
 */
export async function movePreviewBatch(
  gameName: string,
  gameId: string,
  moveType: string,
  candidates: MovePreviewCandidate[],
  params?: Record<string, string | number | boolean>,
  expectedVersion?: number,
  signal?: AbortSignal,
): Promise<ApiResponse<MovePreviewBatchResponse>> {
  return apiPost<MovePreviewBatchResponse>(
    buildGameUrl(gameName, gameId, 'movePreviewBatch', params),
    {
      MoveType: moveType,
      Candidates: candidates,
      ...(expectedVersion === undefined ? {} : { ExpectedVersion: expectedVersion }),
    },
    'application/json',
    signal,
  );
}
