import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';
import type { State } from './_types.js';

class BoardgameRenderPlayerInfoTictactoe extends PlayerInfoRenderer {
  get chipText(): string {
    return this._computeChipText(this.state, this.playerIndex);
  }

  // chipColor intentionally not overridden - uses framework computed color

  private _computeChipText(state: State | null, playerIndex: number): string {
    return state?.Players?.[playerIndex]?.TokenValue || '';
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-render-player-info-tictactoe', BoardgameRenderPlayerInfoTictactoe);
