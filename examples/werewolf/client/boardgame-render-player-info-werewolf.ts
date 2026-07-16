import { html, css } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

/**
 * Player-info renderer for werewolf's roster tiles. Shown for every game
 * surface (solo, Table, and Hand rosters). Deliberately reveals NOTHING
 * about roles: it only receives the viewer-sanitized playerState, where
 * other players' Role reads as the enum zero value, so showing Role here
 * would leak nothing useful and mislead. It surfaces only public status —
 * eliminated / voted / thinking — which is safe for all viewers.
 */
@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoWerewolf extends PlayerInfoRenderer {
  static override styles = css`
    .status {
      font-size: 12px;
      opacity: 0.8;
    }
    .eliminated {
      color: #c62828;
      font-weight: 600;
    }
  `;

  override render() {
    const p = this.playerState;
    if (p?.Eliminated) {
      return html`<div class="status eliminated">Eliminated</div>`;
    }
    // Intentionally NOT showing vote status: playerState.Vote is currently
    // unsanitized, so during Night it would reveal exactly which players
    // are werewolves (see the filed vote-sanitization fix). Eliminated is
    // the only unambiguously-public status; the roster shows just that.
    // \xa0 keeps the tile height stable so the roster doesn't jump.
    return html`<div class="status">\xa0</div>`;
  }
}
