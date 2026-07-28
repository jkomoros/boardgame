import { LitElement, html } from 'lit';
import { property, query } from 'lit/decorators.js';
import {
  BoardgameBasePlayerInfoRenderer,
  type PlayerChipPresentationChangedDetail,
} from './boardgame-base-player-info-renderer.js';

interface PlayerInfoState {
  readonly Players?: readonly unknown[];
}

interface PlayerInfoRendererElement extends HTMLElement {
  state: PlayerInfoState | null;
  playerIndex: number;
}

/**
 * BoardgameRenderPlayerInfo dynamically loads and manages game-specific
 * player info renderers. It instantiates the appropriate renderer based on
 * the game name and handles state synchronization.
 */
class BoardgameRenderPlayerInfo extends LitElement {
  @property({ type: Object })
  state: PlayerInfoState | null = null;

  @property({ type: Boolean })
  active = false;

  @property({ type: String, attribute: 'game-name' })
  gameName = '';

  @property({ type: Object, attribute: false })
  renderer: PlayerInfoRendererElement | null = null;

  @property({ type: String, attribute: false })
  rendererGameName = '';

  @property({ type: Boolean, attribute: 'renderer-loaded' })
  rendererLoaded = false;

  @property({ type: Number, attribute: 'player-index' })
  playerIndex = 0;

  @query('#container')
  private _container?: HTMLElement;

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('active')) {
      this._activeChanged(this.active);
    }

    if (changedProperties.has('gameName')) {
      this._gameNameChanged(this.gameName);
    }

    if (changedProperties.has('rendererLoaded')) {
      this._rendererLoadedChanged(this.rendererLoaded);
    }

    if (changedProperties.has('state')) this._stateChanged(this.state);
    if (changedProperties.has('playerIndex')) this._playerIndexChanged(this.playerIndex);
  }

  private _activeChanged(newValue: boolean) {
    if (newValue) {
      this.instantiateRenderer();
      return;
    }
    this.resetRenderer();
  }

  private _gameNameChanged(newValue: string) {
    if (newValue !== this.rendererGameName) {
      this.resetRenderer();
    }
    if (newValue) this.instantiateRenderer();
  }

  private _rendererLoadedChanged(newValue: boolean) {
    if (!newValue) return;
    if (!this.renderer) {
      this.instantiateRenderer();
    }
  }

  private _stateChanged(newState: PlayerInfoState | null) {
    if (!this.renderer) return;
    this.renderer.state = newState;
  }

  private _playerIndexChanged(playerIndex: number): void {
    if (!this.renderer) return;
    this.renderer.playerIndex = playerIndex;
  }

  private _chipPresentationChanged(event: Event): void {
    event.stopPropagation();
    const detail = (event as CustomEvent<PlayerChipPresentationChangedDetail>).detail;
    this._publishChipPresentation(detail);
  }

  private _publishChipPresentation(detail: PlayerChipPresentationChangedDetail): void {
    this.dispatchEvent(new CustomEvent<PlayerChipPresentationChangedDetail>('player-chip-presentation-changed', {
      bubbles: true,
      composed: true,
      detail,
    }));
  }

  resetRenderer() {
    this.renderer?.remove();
    this.renderer = null;
    this.rendererGameName = '';
    this._publishChipPresentation({ text: '', color: '' });
  }

  instantiateRenderer() {
    if (!this.active || !this.rendererLoaded || !this._container || !this.gameName) return;
    if (this.renderer && this.rendererGameName === this.gameName) return;

    const tagName = `boardgame-render-player-info-${this.gameName}`;
    if (!customElements.get(tagName)) {
      throw new Error(`boardgame-render-player-info: ${tagName} is not registered even though rendererLoaded is true`);
    }
    const ele = document.createElement(tagName) as PlayerInfoRendererElement;
    if (!(ele instanceof BoardgameBasePlayerInfoRenderer)) {
      throw new Error(`boardgame-render-player-info: ${tagName} must extend the generated PlayerInfoRenderer base`);
    }

    ele.state = this.state;
    ele.playerIndex = this.playerIndex;
    ele.addEventListener('player-chip-presentation-changed', event => this._chipPresentationChanged(event));

    this.renderer = ele;
    this.rendererGameName = this.gameName;

    this._container.appendChild(ele);
  }

  override render() {
    return html`
      <div id="container">
        <!-- Dynamic renderer will be inserted here -->
      </div>
    `;
  }
}

customElements.define('boardgame-render-player-info', BoardgameRenderPlayerInfo);

export { BoardgameRenderPlayerInfo };
