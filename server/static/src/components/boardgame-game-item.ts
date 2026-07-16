import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import './boardgame-configure-game-properties.ts';
import './boardgame-player-chip.ts';
import { GamePathMixin } from './boardgame-game-path.js';
import type { GameListItem, ManagerInfo, PlayerInfo } from '../types/store';

@customElement('boardgame-game-item')
export class BoardgameGameItem extends GamePathMixin(LitElement) {
  static styles = css`
    :host {
      display: block;
      --player-chip-size: 32px;
    }

    .card {
      background: linear-gradient(180deg, var(--md-sys-color-surface-container-low, #F5F0E8) 0%, var(--md-sys-color-surface-container, #F0EBE3) 100%);
      padding: 16px;
      margin: 8px;
      border-radius: 12px;
      box-shadow: 0 1px 3px 0 rgba(60, 40, 20, 0.10),
                  0 1px 2px 0 rgba(60, 40, 20, 0.06),
                  inset 0 1px 0 rgba(255, 255, 255, 0.5);
      color: var(--md-sys-color-on-surface, #1C1810);
      transition: box-shadow 0.25s ease, transform 0.25s ease;
    }

    .card:hover {
      box-shadow: 0 4px 12px 0 rgba(60, 40, 20, 0.14),
                  0 2px 4px 0 rgba(60, 40, 20, 0.08),
                  inset 0 1px 0 rgba(255, 255, 255, 0.5);
      transform: translateY(-2px);
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
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      margin-left: 8px;
    }

    .empty {
      font-style: italic;
    }

    boardgame-player-chip {
      margin-left: 0.5em;
    }

    a {
      color: var(--accent-color, #8B7432);
      text-decoration: none;
      font-weight: 600;
      font-family: var(--md-sys-typescale-title-medium-font, 'Source Sans 3', sans-serif);
      transition: color 0.2s ease;
    }

    a:hover {
      color: var(--md-sys-color-primary, #2E6B4F);
    }
  `;

  @property({ type: Object })
  item: GameListItem | null = null;

  @property({ type: Array })
  managers: ManagerInfo[] = [];

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
              .photoUrl="${player.PhotoURL || ''}"
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
