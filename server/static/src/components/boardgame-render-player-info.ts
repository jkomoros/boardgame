import { LitElement, html } from 'lit';
import { property, query } from 'lit/decorators.js';

interface PlayerInfoState {
  readonly Players?: readonly unknown[];
}

interface PlayerInfoRendererElement extends HTMLElement {
  state: PlayerInfoState | null;
  playerIndex: number;
  playerState: unknown;
  readonly chipText?: unknown;
  readonly chipColor?: unknown;
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

  @property({ type: String })
  gameName = '';

  @property({ type: Object, attribute: false })
  renderer: PlayerInfoRendererElement | null = null;

  @property({ type: String, attribute: false })
  rendererGameName = '';

  @property({ type: Boolean })
  rendererLoaded = false;

  @property({ type: String, attribute: true })
  chipText = '';

  @property({ type: String, attribute: true })
  chipColor = '';

  @property({ type: Number })
  playerIndex = 0;

  @query('#container')
  private _container?: HTMLElement;

  private instantiateWhenReady = false;
  private instantiateWhenGameNameSet = false;

  get playerState(): unknown {
    return this._computePlayerState(this.state, this.playerIndex);
  }

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

    if (changedProperties.has('state')) {
      this._stateChanged(this.state, changedProperties.get('state'));
    }

    if (changedProperties.has('playerIndex') || changedProperties.has('state')) {
      this._playerStateChanged(this.playerState);
    }
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);

    if (this.instantiateWhenReady) {
      this.instantiateRenderer();
    }
  }

  private _activeChanged(newValue: boolean) {
    if (!newValue) {
      if (!this.renderer) return;
      if (this.renderer.parentElement) {
        this.renderer.parentElement.removeChild(this.renderer);
      }
      this.renderer = null;
    }
  }

  private _chipTextChanged(e: CustomEvent<{ value?: unknown }>) {
    if (typeof e.detail.value === 'string') this.chipText = e.detail.value;
  }

  private _chipColorChanged(e: CustomEvent<{ value?: unknown }>) {
    if (typeof e.detail.value === 'string') this.chipColor = e.detail.value;
  }

  private _gameNameChanged(newValue: string) {
    if (!newValue) return;
    if (newValue !== this.rendererGameName) {
      this.resetRenderer();
    }
    if (!this.instantiateWhenGameNameSet) return;
    this.instantiateRenderer();
  }

  private _rendererLoadedChanged(newValue: boolean) {
    if (!newValue) return;
    if (!this.renderer) {
      this.instantiateRenderer();
    }
  }

  private _stateChanged(newState: PlayerInfoState | null, oldState: unknown) {
    if (!this.renderer) return;

    // If state changed from non-null to null, skip
    if (!newState && oldState) {
      return;
    }

    // For Lit renderers, just set the property directly
    this.renderer.state = newState;
    this.requestUpdate();
  }

  private _computePlayerState(state: PlayerInfoState | null, playerIndex: number): unknown {
    if (!state) return null;
    return state.Players?.[playerIndex];
  }

  private _playerStateChanged(newValue: unknown) {
    if (!this.renderer) return;
    this.renderer.playerState = newValue;
  }

  resetRenderer() {
    if (!this._container) return;
    if (this.renderer) {
      this._container.removeChild(this.renderer);
    }
    this.renderer = null;
    this.rendererGameName = '';
  }

  instantiateRenderer() {
    if (!this.rendererLoaded) return;
    if (!this._container) {
      this.instantiateWhenReady = true;
      return;
    }

    if (!this.gameName) {
      this.instantiateWhenGameNameSet = true;
      return;
    }

    const ele = document.createElement(`boardgame-render-player-info-${this.gameName}`) as PlayerInfoRendererElement;

    ele.state = this.state;
    ele.playerIndex = this.playerIndex;
    ele.playerState = this.playerState;

    this.chipText = typeof ele.chipText === 'string' ? ele.chipText : '';
    this.chipColor = typeof ele.chipColor === 'string' ? ele.chipColor : '';

    this.renderer = ele;
    this.rendererGameName = this.gameName;

    // Listen for chip property changes from the renderer
    if (this.renderer) {
      this.renderer.addEventListener('chip-text-changed', (e: Event) => this._chipTextChanged(e as CustomEvent<{ value?: unknown }>));
      this.renderer.addEventListener('chip-color-changed', (e: Event) => this._chipColorChanged(e as CustomEvent<{ value?: unknown }>));
    }

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
