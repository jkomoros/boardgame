/**
 * surfaceForGame returns which companion-mode surface this browser should
 * present for the given game: 'table' (the shared projector), 'hand' (a
 * player's phone), or null for the regular solo view.
 *
 * A ?display=table|hand query param takes precedence over tab-scoped intent
 * and the per-game surface cookie (surface_<gameId>, set by the server at
 * game-create for the host and at /api/join/seat for phones). Successful
 * framework transitions put the surface in the URL so tab restore remains
 * deterministic; authors can use the same override for local visual testing.
 *
 * The value is presentation intent only and never grants authority. Shared by boardgame-render-game (renderer module selection) and
 * boardgame-game-view (hiding solo chrome on companion surfaces) so the
 * two can never disagree about what surface is active.
 */
const ephemeralTableDeviceIDs = new Map<string, string>();

export function surfaceForGame(
  gameId: string,
  companionMode: boolean | undefined = undefined,
): 'table' | 'hand' | null {
  // Persisted presentation intent is never authoritative. In particular, a
  // tab that slept through switchToSolo must not resurrect a Hand/Table
  // renderer merely because its old browser state survived the socket event.
  if (companionMode === false) return null;
  const params = new URLSearchParams(window.location.search);
  const display = params.get('display');
  if (display === 'table' || display === 'hand') return display;
  if (gameId) {
    try {
      const tab = window.sessionStorage.getItem(`boardgame-surface:${gameId}`);
      if (tab === 'table' || tab === 'hand') return tab;
      // Working tab storage with no value means this tab never opted into a
      // companion surface. Do not inherit the origin-wide cookie written for
      // another tab in the same browser.
      return null;
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
 * Records presentation intent for this tab. The API also writes a cookie, but
 * that cookie is both origin-wide and invisible when API_HOST and the web
 * application use different origins. sessionStorage survives reload without
 * allowing one Hand/Table tab to silently change every other tab. This value
 * only chooses a renderer; server-side auth and sanitization remain authoritative.
 */
export function rememberSurfaceForGame(gameId: string, surface: 'table' | 'hand'): void {
  if (!gameId) throw new Error('gameId is required to remember a companion surface');
  try {
    window.sessionStorage.setItem(`boardgame-surface:${gameId}`, surface);
  } catch { /* cookie/query fallback still works when storage is unavailable */ }
}

export function forgetSurfaceForGame(gameId: string): void {
  if (!gameId) return;
  try { window.sessionStorage.removeItem(`boardgame-surface:${gameId}`); } catch { /* ignored */ }
}

export function forgetAllCompanionSurfaces(): void {
  try {
    const keys: string[] = [];
    for (let i = 0; i < window.sessionStorage.length; i++) {
      const key = window.sessionStorage.key(i);
      if (key?.startsWith('boardgame-surface:')) keys.push(key);
    }
    for (const key of keys) window.sessionStorage.removeItem(key);
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
		const sessionExisting = window.sessionStorage.getItem(key);
		if (sessionExisting && /^[a-f0-9]{32}$/.test(sessionExisting)) return sessionExisting;
		const bytes = new Uint8Array(16);
		window.crypto.getRandomValues(bytes);
		const created = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
		window.localStorage.setItem(key, created);
		window.sessionStorage.setItem(key, created);
		return created;
	} catch {
		// localStorage can be disabled independently of sessionStorage. Preserve
		// reload idempotency there before falling back to this document only.
		try {
			const existing = window.sessionStorage.getItem(key);
			if (existing && /^[a-f0-9]{32}$/.test(existing)) return existing;
			const bytes = new Uint8Array(16);
			window.crypto.getRandomValues(bytes);
			const created = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
			window.sessionStorage.setItem(key, created);
			return created;
		} catch { /* both browser storage mechanisms are unavailable */ }
		const bytes = new Uint8Array(16);
		window.crypto.getRandomValues(bytes);
		const existing = ephemeralTableDeviceIDs.get(gameId);
		if (existing) return existing;
		const created = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
		ephemeralTableDeviceIDs.set(gameId, created);
		return created;
	}
}
