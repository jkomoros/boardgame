/**
 * boardgame-gathering-share
 *
 * Shows a "Copy invite link" button that copies the game URL to the clipboard.
 */
import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import '@material/web/button/outlined-button.js';
import '@material/web/icon/icon.js';

interface GameRoute {
  name: string;
  id: string;
}

@customElement('boardgame-gathering-share')
export class BoardgameGatheringShare extends LitElement {
  static styles = css`
    :host {
      display: inline-block;
    }
    .copied {
      font-size: 12px;
      color: var(--md-sys-color-primary, #2E6B4F);
      margin-left: 8px;
      opacity: 0;
      transition: opacity 0.3s;
    }
    .copied.show {
      opacity: 1;
    }
  `;

  @property({ type: Object, attribute: 'game-route' })
  gameRoute: GameRoute | null = null;

  @state()
  private _copied = false;

  private get _gameUrl(): string {
    if (!this.gameRoute) return window.location.href;
    return `${window.location.origin}/game/${this.gameRoute.name}/${this.gameRoute.id}`;
  }

  private async _handleCopy(): Promise<void> {
    try {
      await navigator.clipboard.writeText(this._gameUrl);
      this._copied = true;
      setTimeout(() => { this._copied = false; }, 2000);
    } catch {
      // Fallback: select the URL text
      const input = document.createElement('input');
      input.value = this._gameUrl;
      document.body.appendChild(input);
      input.select();
      document.execCommand('copy');
      document.body.removeChild(input);
      this._copied = true;
      setTimeout(() => { this._copied = false; }, 2000);
    }
  }

  render() {
    return html`
      <md-outlined-button @click=${this._handleCopy}>
        <md-icon slot="icon">link</md-icon>
        Copy invite link
      </md-outlined-button>
      <span class="copied ${this._copied ? 'show' : ''}">Copied!</span>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-gathering-share': BoardgameGatheringShare;
  }
}
