import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import '@material/web/iconbutton/icon-button.js';
import '@material/web/icon/icon.js';
import './boardgame-ajax.ts';

import type { BoardgameAjax } from './boardgame-ajax.ts';

interface GameRoute {
  name?: string;
  id?: string;
}

interface ApiResponse {
  Status?: string;
  Error?: string;
  FriendlyError?: string;
}

@customElement('boardgame-configure-game-properties')
export class BoardgameConfigureGameProperties extends LitElement {
  static styles = css`
    :host {
      display: inline-flex;
      gap: 8px;
    }
  `;

  @property({ type: Boolean })
  gameVisible = false;

  @property({ type: Boolean })
  gameOpen = false;

  @property({ type: Boolean })
  admin = false;

  @property({ type: Boolean })
  isOwner = false;

  @property({ type: Object })
  gameRoute: GameRoute | null = null;

  @property({ type: Boolean })
  configurable = false;

  @property({ type: Object })
  private _response: ApiResponse | null = null;

  @query('#ajax')
  private ajax!: BoardgameAjax;

  get disabled(): boolean {
    return !(this.admin || this.isOwner || this.configurable);
  }

  protected updated(changedProperties: Map<string, unknown>): void {
    if (changedProperties.has('_response')) {
      this._responseChanged(this._response);
    }
  }

  private _visibleIcon(gameVisible: boolean): string {
    return gameVisible ? "visibility" : "visibility_off";
  }

  private _openIcon(gameOpen: boolean): string {
    return gameOpen ? "people" : "people_outline";
  }

  private _openAlt(gameOpen: boolean): string {
    return gameOpen
      ? "Anyone who has the link can join"
      : "Only specifically invited people may join";
  }

  private _visibleAlt(gameVisible: boolean): string {
    return gameVisible
      ? "Your game is publicly listed so random people can find it"
      : "Your game is unlisted so only people you share the link with can find it";
  }

  private _handleOpenTapped(): void {
    this._submit(!this.gameOpen, this.gameVisible);
  }

  private _handleVisibleTapped(): void {
    this._submit(this.gameOpen, !this.gameVisible);
  }

  private _submit(open: boolean, visible: boolean): void {
    this.ajax.body = {
      "open": open ? 1 : 0,
      "visible": visible ? 1 : 0,
      "admin": this.admin ? 1 : 0
    };
    this.ajax.generateRequest();
  }

  private _responseChanged(newValue: ApiResponse | null): void {
    if (!newValue) return;

    if (newValue.Status === "Success") {
      // Tell game-view to fetch data now
      this.dispatchEvent(new CustomEvent("refresh-info", { composed: true }));
    } else {
      this.dispatchEvent(new CustomEvent("show-error", {
        composed: true,
        detail: {
          message: newValue.Error,
          friendlyMessage: newValue.FriendlyError,
          title: "Couldn't toggle"
        }
      }));
    }
  }

  render() {
    return html`
      <md-icon-button
        ?disabled="${this.disabled}"
        @click="${this._handleOpenTapped}"
        title="${this._openAlt(this.gameOpen)}">
        <md-icon>${this._openIcon(this.gameOpen)}</md-icon>
      </md-icon-button>
      <md-icon-button
        ?disabled="${this.disabled}"
        @click="${this._handleVisibleTapped}"
        title="${this._visibleAlt(this.gameVisible)}">
        <md-icon>${this._visibleIcon(this.gameVisible)}</md-icon>
      </md-icon-button>
      <boardgame-ajax
        id="ajax"
        game-path="configure"
        .gameRoute="${this.gameRoute}"
        method="POST"
        content-type="application/x-www-form-urlencoded"
        .lastResponse="${this._response}">
      </boardgame-ajax>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-configure-game-properties': BoardgameConfigureGameProperties;
  }
}
