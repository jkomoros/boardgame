import { LitElement, html } from 'lit';
import { property, query } from 'lit/decorators.js';
import {
  BoardgameBasePlayerInfoRenderer,
  type PlayerChipPresentationChangedDetail,
} from './boardgame-base-player-info-renderer.js';
import type { RendererHost } from '../renderer-host/renderer-host.js';
import { TIMER_SERVICE_REQUEST_EVENT, type TimerServiceRequestDetail } from '@boardgame/elements/timers/timer-service.js';

const pendingModuleLoads = new WeakMap<RendererHost, Map<string, Promise<unknown>>>();

/** The custom element a game's player-info renderer registers. */
export function playerInfoRendererTag(gameName: string): string {
  return `boardgame-render-player-info-${gameName}`;
}

/**
 * Makes sure a game's player-info renderer is registered and extends the
 * generated base, loading its module through the host if it is not there yet.
 * Rejects with a message fit to show a game author.
 */
export async function ensurePlayerInfoRenderer(
  host: RendererHost | null,
  gameName: string,
  attempt = 0,
): Promise<void> {
  if (!/^[a-z][a-z0-9]*$/.test(gameName)) {
    throw new Error(`Invalid player renderer game name ${JSON.stringify(gameName)}; expected lowercase letters and digits`);
  }
  const tagName = playerInfoRendererTag(gameName);
  const validate = (constructor: CustomElementConstructor): void => {
    if (!(constructor.prototype instanceof BoardgameBasePlayerInfoRenderer)) {
      throw new Error(`Player renderer <${tagName}> must extend the generated PlayerInfoRenderer base`);
    }
  };
  const existing = customElements.get(tagName);
  if (existing) {
    validate(existing);
    return;
  }
  if (!host?.loadPlayerInfoModule) {
    throw new Error(`the renderer host provides no player-info module loader for <${tagName}>`);
  }
  // Every seat asks at once; the host is asked once per game and attempt.
  let loads = pendingModuleLoads.get(host);
  if (!loads) pendingModuleLoads.set(host, loads = new Map());
  const key = `${gameName}:${attempt}`;
  let load = loads.get(key);
  if (!load) {
    load = Promise.resolve(host.loadPlayerInfoModule({ gameName, attempt }));
    loads.set(key, load);
    const settled = (): void => { loads.delete(key); };
    load.then(settled, settled);
  }
  await load;
  const constructor = customElements.get(tagName);
  if (!constructor) {
    throw new Error(
      `Player renderer module loaded but did not register <${tagName}>; ` +
      'use @registerPlayerInfoRenderer',
    );
  }
  validate(constructor);
}

interface PlayerInfoState {
  readonly Players?: readonly unknown[];
}

interface PlayerInfoRendererElement extends HTMLElement {
  state: PlayerInfoState | null;
  playerIndex: number;
}

/**
 * BoardgameRenderPlayerInfo mounts a game-specific player info renderer for
 * one seat and keeps its state in step.
 *
 * It can be driven two ways. The webapp's roster loads the renderer module
 * once for every seat and sets `rendererLoaded` on each of these elements.
 * Any other host sets `host` instead (or calls `mountPlayerInfo()`), and the
 * element loads the module through `RendererHost.loadPlayerInfoModule` by
 * itself: no roster and no store involved.
 */
class BoardgameRenderPlayerInfo extends LitElement {
  /**
   * When set, this element loads the game's player-info renderer through the
   * host and manages `rendererLoaded` itself.
   */
  @property({ attribute: false })
  host: RendererHost | null = null;

  /** Why the host-driven load failed; empty when it has not. */
  @property({ type: String, attribute: false })
  rendererError = '';

  private _hostLoadGeneration = 0;

  constructor() {
    super();
    // A player-info renderer may read timers. Mounted outside a
    // boardgame-render-game, nothing above it need answer, so answer from the
    // host's service when there is one.
    this.addEventListener(TIMER_SERVICE_REQUEST_EVENT, (event: Event) => {
      const service = this.host?.timerService;
      const accept = (event as CustomEvent<TimerServiceRequestDetail>).detail?.accept;
      if (!service || typeof accept !== 'function') return;
      event.stopPropagation();
      accept(service);
    });
  }

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

    if ((changedProperties.has('host') || changedProperties.has('gameName')) && this.host) {
      void this._loadThroughHost();
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

  private async _loadThroughHost(): Promise<void> {
    const generation = ++this._hostLoadGeneration;
    const host = this.host;
    const gameName = this.gameName;
    this.rendererError = '';
    this.rendererLoaded = false;
    if (!host || !gameName) return;
    try {
      await ensurePlayerInfoRenderer(host, gameName);
      if (generation !== this._hostLoadGeneration) return;
      this.rendererLoaded = true;
    } catch (error) {
      if (generation !== this._hostLoadGeneration) return;
      const message = error instanceof Error ? error.message : String(error);
      this.rendererError = `Failed to load player renderer for ${gameName}: ${message.trim() || 'unknown renderer error'}`;
      console.error(this.rendererError, error);
    }
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

    const tagName = playerInfoRendererTag(this.gameName);
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
      ${this.rendererError ? html`<div role="alert">${this.rendererError}</div>` : ''}
      <div id="container">
        <!-- Dynamic renderer will be inserted here -->
      </div>
    `;
  }
}

customElements.define('boardgame-render-player-info', BoardgameRenderPlayerInfo);

export interface PlayerInfoMountOptions {
  /** Supplies the module loader and, for renderers that read timers, the timer service. */
  readonly host: RendererHost;
  readonly gameName: string;
  readonly playerIndex: number;
  /** The expanded state for the viewer; update it later through the returned element's `state`. */
  readonly state: PlayerInfoState | null;
}

/**
 * The host-owned mount path for a game's player-info renderer: appends a
 * `boardgame-render-player-info` for one seat to `container`, wherever the
 * host's layout wants it. The element loads the renderer through the host,
 * shows a load failure in place, and emits `player-chip-presentation-changed`.
 * Assign `state` and `playerIndex` on the returned element to update it, and
 * `remove()` it to unmount.
 */
export function mountPlayerInfo(
  container: Element | DocumentFragment,
  options: PlayerInfoMountOptions,
): BoardgameRenderPlayerInfo {
  const element = document.createElement('boardgame-render-player-info') as BoardgameRenderPlayerInfo;
  element.host = options.host;
  element.gameName = options.gameName;
  element.playerIndex = options.playerIndex;
  element.state = options.state;
  element.active = true;
  container.appendChild(element);
  return element;
}

export { BoardgameRenderPlayerInfo };
