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
export function surfaceForGame(gameId: string): 'table' | 'hand' | null {
  const params = new URLSearchParams(window.location.search);
  const display = params.get('display');
  if (display === 'table' || display === 'hand') return display;
  if (gameId) {
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
