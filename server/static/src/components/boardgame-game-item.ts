import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import './boardgame-configure-game-properties.ts';
import './boardgame-player-chip.ts';
import { GamePathMixin } from './boardgame-game-path.js';

interface PlayerInfo {
  PhotoUrl: string;
  DisplayName: string;
  IsAgent: boolean;
  IsEmpty: boolean;
}

interface GameItem {
  Name: string;
  ID: string;
  Players: PlayerInfo[];
  ReadableLastActivity: string;
  Open: boolean;
  Visible: boolean;
}

interface Manager {
  Name: string;
  DisplayName: string;
}

@customElement('boardgame-game-item')
export class BoardgameGameItem extends GamePathMixin(LitElement) {
  static styles = css`
    :host {
      display: block;
      --player-chip-size: 32px;
    }

    .card {
      background: var(--md-sys-color-surface-container-low, #f7f2fa);
      padding: 16px;
      margin: 8px;
      border-radius: 12px;
      box-shadow: var(--md-sys-elevation-1, 0 1px 3px 1px rgba(0,0,0,.15), 0 1px 2px rgba(0,0,0,.3));
      color: var(--md-sys-color-on-surface, #1c1b1f);
    }

    .layout {
      display: flex;
    }

    .horizontal {
      flex-direction: row;
    }

    .center {
      align-items: center;
    }

    .flex {
      flex: 1;
    }

    .minor {
      font-size: 12px;
      font-weight: 400;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      margin-left: 8px;
    }

    .empty {
      font-style: italic;
    }

    boardgame-player-chip {
      margin-left: 0.5em;
    }

    a {
      color: var(--accent-color, #ff4081);
      text-decoration: none;
      font-weight: 500;
    }

    a:hover {
      text-decoration: underline;
    }
  `;

  @property({ type: Object })
  item: GameItem | null = null;

  @property({ type: Array })
  managers: Manager[] = [];

  get gameDisplayName(): string {
    if (!this.item) return "";
    if (!this.managers) return "";
    for (let i = 0; i < this.managers.length; i++) {
      const manager = this.managers[i];
      if (manager.Name === this.item.Name) {
        return manager.DisplayName;
      }
    }
    return this.item.Name;
  }

  private _playerItemClasses(playerItem: PlayerInfo): string {
    return playerItem.IsEmpty ? "empty" : "";
  }

  private _displayNameForPlayerItem(playerItem: PlayerInfo): string {
    return playerItem.IsEmpty ? "No one" : playerItem.DisplayName;
  }

  render() {
    if (!this.item) return html``;

    return html`
      <div class="card layout horizontal center">
        <a href="${this.GamePath(this.item.Name, this.item.ID)}">
          ${this.gameDisplayName}
        </a>
        ${repeat(
          this.item.Players || [],
          (player) => player.DisplayName || 'empty',
          (player) => html`
            <boardgame-player-chip
              .photoUrl="${player.PhotoUrl}"
              .displayName="${player.DisplayName}"
              ?is-agent="${player.IsAgent}">
            </boardgame-player-chip>
          `
        )}
        <span class="minor">Last activity ${this.item.ReadableLastActivity}</span>
        <div class="flex"></div>
        <span class="minor">${this.item.ID}</span>
        <boardgame-configure-game-properties
          ?game-open="${this.item.Open}"
          ?game-visible="${this.item.Visible}">
        </boardgame-configure-game-properties>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-game-item': BoardgameGameItem;
  }
}
