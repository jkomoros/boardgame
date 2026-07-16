import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { PlayerState, State } from './_types.js';
import '../../src/components/boardgame-status-text.js';

class BoardgameRenderPlayerInfoCheckers extends LitElement {
  @property({ type: Object }) state: State | null = null;
  @property({ type: Number }) playerIndex = 0;
  @property({ type: Object }) playerState: PlayerState | null = null;

  override render() {
    return html`
      Number of cards:
      <boardgame-status-text>${this.playerState?.Hand.Indexes.length ?? 0}</boardgame-status-text>
    `;
  }
}

customElements.define('boardgame-render-player-info-checkers', BoardgameRenderPlayerInfoCheckers);
