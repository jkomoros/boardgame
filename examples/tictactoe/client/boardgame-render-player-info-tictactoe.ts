import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';

class BoardgameRenderPlayerInfoTictactoe extends LitElement {
  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  playerIndex = 0;

  get chipText(): string {
    return this._computeChipText(this.state, this.playerIndex);
  }

  // chipColor intentionally not overridden - uses framework computed color

  private _computeChipText(state: any, playerIndex: number): string {
    return state?.Players?.[playerIndex]?.TokenValue || '';
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-render-player-info-tictactoe', BoardgameRenderPlayerInfoTictactoe);
