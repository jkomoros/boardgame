/**
 * surfaceForGame returns which companion-mode surface this browser should
 * present for the given game: 'table' (the shared projector), 'hand' (a
 * player's phone), or null for the regular solo view.
 *
 * A ?display=table|hand query param takes precedence over the per-game
 * surface cookie (surface_<gameId>, set by the server at game-create for
 * the host and at /api/join/seat for phones). The param override exists
 * for developer testing — it lets one machine show both surfaces — and is
 * harmless in prod because no production client sets it.
 *
 * Shared by boardgame-render-game (renderer module selection) and
 * boardgame-game-view (hiding solo chrome on companion surfaces) so the
 * two can never disagree about what surface is active.
 */
const ephemeralTableDeviceIDs = new Map<string, string>();

export function surfaceForGame(gameId: string): 'table' | 'hand' | null {
  const params = new URLSearchParams(window.location.search);
  const display = params.get('display');
  if (display === 'table' || display === 'hand') return display;
  if (gameId) {
    try {
      const local = window.localStorage.getItem(`boardgame-surface:${gameId}`);
      if (local === 'table' || local === 'hand') return local;
    } catch { /* storage may be disabled; the cookie remains the fallback */ }
    const cookieName = `surface_${gameId}=`;
    const cookies = document.cookie.split('; ');
    for (const c of cookies) {
      if (c.startsWith(cookieName)) {
        const value = c.slice(cookieName.length);
        if (value === 'table' || value === 'hand') return value;
      }
    }
  }
  return null;
}

/**
 * Records presentation intent on the frontend origin. The API also writes a
 * cookie, but that cookie is invisible when API_HOST and the web application
 * use different origins. This value only chooses a renderer; server-side auth
 * and state sanitization remain authoritative for private information.
 */
export function rememberSurfaceForGame(gameId: string, surface: 'table' | 'hand'): void {
  if (!gameId) throw new Error('gameId is required to remember a companion surface');
  try {
    window.localStorage.setItem(`boardgame-surface:${gameId}`, surface);
  } catch { /* cookie/query fallback still works when storage is unavailable */ }
}

export function forgetSurfaceForGame(gameId: string): void {
  if (!gameId) return;
  try { window.localStorage.removeItem(`boardgame-surface:${gameId}`); } catch { /* ignored */ }
}

export function forgetAllCompanionSurfaces(): void {
  try {
    const keys: string[] = [];
    for (let i = 0; i < window.localStorage.length; i++) {
      const key = window.localStorage.key(i);
      if (key?.startsWith('boardgame-surface:')) keys.push(key);
    }
    for (const key of keys) window.localStorage.removeItem(key);
  } catch { /* ignored */ }
}

/** Stable, non-secret browser identity used only to make Table acquisition
 * idempotent when a committed HTTP response is lost. Authority still requires
 * the server-issued HttpOnly credential. */
export function tableRecoveryDeviceID(gameId: string): string {
	if (!gameId) throw new Error('gameId is required for a Table recovery device ID');
	const key = `boardgame-table-device:${gameId}`;
	try {
		const existing = window.localStorage.getItem(key);
		if (existing && /^[a-f0-9]{32}$/.test(existing)) return existing;
		const bytes = new Uint8Array(16);
		window.crypto.getRandomValues(bytes);
		const created = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
		window.localStorage.setItem(key, created);
		return created;
	} catch {
		// Storage can be unavailable in privacy modes. A per-page ID still
		// makes ordinary retries idempotent for the life of this document.
		const bytes = new Uint8Array(16);
		window.crypto.getRandomValues(bytes);
		const existing = ephemeralTableDeviceIDs.get(gameId);
		if (existing) return existing;
		const created = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
		ephemeralTableDeviceIDs.set(gameId, created);
		return created;
	}
}
