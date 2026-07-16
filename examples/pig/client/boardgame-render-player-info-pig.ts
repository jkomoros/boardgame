import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import '../../src/components/boardgame-status-text.js';
import type { PlayerState, State } from './_types.js';

class BoardgameRenderPlayerInfoPig extends LitElement {
  @property({ type: Object })
  state: State | null = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: PlayerState | null = null;

  override render() {
    return html`
      <div>Round Score <boardgame-status-text>${this.playerState?.RoundScore}</boardgame-status-text></div>
      <div>Total Score <boardgame-status-text>${this.playerState?.Score}</boardgame-status-text></div>
    `;
  }
}

customElements.define('boardgame-render-player-info-pig', BoardgameRenderPlayerInfoPig);
