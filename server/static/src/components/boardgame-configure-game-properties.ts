import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@material/web/iconbutton/icon-button.js';
import '@material/web/icon/icon.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import { configureGame } from '../actions/game.js';

interface GameRoute {
  name: string;
  id: string;
}

@customElement('boardgame-configure-game-properties')
export class BoardgameConfigureGameProperties extends connect(store)(LitElement) {
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

  get disabled(): boolean {
    return !(this.admin || this.isOwner || this.configurable);
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

  private async _submit(open: boolean, visible: boolean): Promise<void> {
    if (!this.gameRoute) return;

    const response = await store.dispatch(
      configureGame(this.gameRoute, open, visible, this.admin)
    );

    if (response.error) {
      // Dispatch error event
      this.dispatchEvent(new CustomEvent("show-error", {
        composed: true,
        detail: {
          message: response.error,
          friendlyMessage: response.friendlyError,
          title: "Couldn't toggle"
        }
      }));
    } else {
      // Tell game-view to fetch data now
      this.dispatchEvent(new CustomEvent("refresh-info", { composed: true }));
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
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-configure-game-properties': BoardgameConfigureGameProperties;
  }
}
