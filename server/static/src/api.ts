/**
 * Typed API utilities for making HTTP requests to the boardgame server.
 * Replaces iron-ajax/boardgame-ajax with modern fetch-based approach.
 */

// API_HOST is defined in index.html
declare const API_HOST: string;

/**
 * Standard API response wrapper
 */
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  friendlyError?: string;
  status: number;
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
export async function apiGet<T>(url: string): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(url, {
      method: 'GET',
      credentials: 'include', // Matches iron-ajax withCredentials: true
      headers: {
        'Accept': 'application/json',
      },
    });

    const status = response.status;

    // Try to parse JSON response
    let jsonData: any;
    try {
      jsonData = await response.json();
    } catch {
      // If JSON parsing fails, treat as error
      return {
        status,
        error: `HTTP ${status}: ${response.statusText}`,
        friendlyError: 'The server returned an invalid response',
      };
    }

    // Check for application-level errors (boardgame server format)
    if (jsonData.Status === 'Success') {
      return {
        status,
        data: jsonData as T,
      };
    }

    // Handle application errors
    return {
      status,
      error: jsonData.Error || `Request failed with status ${status}`,
      friendlyError: jsonData.FriendlyError || 'An error occurred',
    };
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
  body: Record<string, any>,
  contentType: 'application/json' | 'application/x-www-form-urlencoded' = 'application/json'
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
    });

    const status = response.status;

    // Try to parse JSON response
    let jsonData: any;
    try {
      jsonData = await response.json();
    } catch {
      // If JSON parsing fails, treat as error
      return {
        status,
        error: `HTTP ${status}: ${response.statusText}`,
        friendlyError: 'The server returned an invalid response',
      };
    }

    // Check for application-level errors (boardgame server format)
    if (jsonData.Status === 'Success') {
      return {
        status,
        data: jsonData as T,
      };
    }

    // Handle application errors
    return {
      status,
      error: jsonData.Error || `Request failed with status ${status}`,
      friendlyError: jsonData.FriendlyError || 'An error occurred',
    };
  } catch (error) {
    // Network errors
    return {
      status: 0,
      error: error instanceof Error ? error.message : 'Network error',
      friendlyError: 'Unable to connect to the server',
    };
  }
}
