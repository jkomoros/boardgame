/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { styleMap } from 'lit/directives/style-map.js';
import type { PlayerChipPresentationChangedDetail } from './boardgame-base-player-info-renderer.js';
import './boardgame-player-chip.ts';
import './boardgame-render-player-info.js';

@customElement('boardgame-player-roster-item')
export class BoardgamePlayerRosterItem extends LitElement {
  static styles = css`
    :host {
      display: block;
    }

    .layout {
      display: flex;
    }

    .horizontal {
      flex-direction: row;
    }

    .vertical {
      flex-direction: column;
    }

    .center {
      align-items: center;
    }

    strong {
      font-size: var(--md-sys-typescale-title-medium-size, 16px);
      font-weight: var(--md-sys-typescale-title-medium-weight, 500);
      letter-spacing: 0.005em;
      color: var(--md-sys-color-on-surface, #1C1810);
    }

    boardgame-player-chip {
      padding-right: 10px;
    }

    .nobody {
      opacity: 0.5;
    }

    .loser {
      filter: saturate(0.5) brightness(1.5) blur(1px);
    }

    .inactive {
      opacity: 0.4;
      filter: grayscale(0.6);
    }

    strong.chip {
      font-size: 12px;
      font-weight: 400;
      background-color: var(--md-sys-color-outline, #857B6E);
      color: white;
      padding: 0.25em;
      height: 1em;
      width: 1em;
      box-sizing: content-box;
      text-align: center;
      border-radius: 50%;
      position: absolute;
      text-overflow: initial;
      line-height: 14px;
      bottom: 0.5em;
      right: 1.5em;
    }

    .current strong.chip {
      background-color: var(--light-accent-color, var(--md-sys-color-tertiary, #8B7432));
      box-shadow: 0 0 0 4px var(--light-accent-color, var(--md-sys-color-tertiary, #8B7432));
    }

    span {
      font-size: var(--md-sys-typescale-body-small-size, 12px);
      font-weight: 400;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
    }

    .viewing span {
      font-weight: bold;
      color: var(--accent-color, var(--md-sys-color-tertiary, #8B7432));
    }

    boardgame-render-player-info {
      font-size: 12px;
      font-weight: 400;
      overflow: visible;
    }
  `;

  @property({ type: String, attribute: 'game-name' })
  gameName = '';

  @property({ type: Boolean, attribute: 'is-empty' })
  isEmpty = false;

  @property({ type: Boolean, attribute: 'is-agent' })
  isAgent = false;

  @property({ type: Boolean })
  active = false;

  @property({ type: String, attribute: 'photo-url' })
  photoUrl = '';

  @property({ type: String, attribute: 'display-name' })
  displayName = '';

  @property({ type: Object })
  state: unknown = null;

  @property({ type: Number, attribute: 'player-index' })
  playerIndex = 0;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Number, attribute: 'current-player-index' })
  currentPlayerIndex = 0;

  @property({ type: Boolean })
  finished = false;

  @property({ type: Boolean })
  winner = false;

  @property({ type: Boolean, attribute: 'renderer-loaded' })
  rendererLoaded = false;

  @property({ type: String, attribute: 'chip-text' })
  chipText = '';

  @property({ type: String, attribute: 'chip-color' })
  chipColor = '';

  // Framework-computed color from FrameworkComputedPlayerProperties.
  @property({ type: String, attribute: 'computed-color' })
  computedColor = '';

  // Whether this player may be active (from FrameworkComputedPlayerProperties).
  @property({ type: Boolean, attribute: 'may-be-active' })
  mayBeActive = true;

  // Priority: chipColor (game renderer) > computedColor (framework) > '' (chip hash fallback)
  private get _effectiveColor(): string {
    return this.chipColor || this.computedColor || '';
  }

  private nameOrNobody(displayName: string): string {
    return displayName ? displayName : "Nobody";
  }

  private classForName(displayName: string): string {
    if (!displayName) return "nobody";
    return "";
  }

  private _styleForChip(chipColor: string, finished: boolean, winner: boolean): Readonly<Record<string, string>> {
    if (finished) {
      return {
        boxShadow: 'none',
        backgroundColor: winner
          ? 'var(--md-sys-color-primary, #2E6B4F)'
          : 'var(--md-sys-color-error, #BA1A1A)',
      };
    }
    if (!chipColor) return { boxShadow: 'none' };
    if (!CSS.supports('color', chipColor)) {
      throw new Error(`boardgame-player-roster-item: chip color ${JSON.stringify(chipColor)} is not valid CSS`);
    }
    return { backgroundColor: chipColor };
  }

  private _chipPresentationChanged(event: CustomEvent<PlayerChipPresentationChangedDetail>): void {
    const detail: unknown = event.detail;
    if (!detail || typeof detail !== 'object') {
      throw new Error('boardgame-player-roster-item: player chip presentation event requires text and color strings');
    }
    const record = detail as Readonly<Record<string, unknown>>;
    if (typeof record['text'] !== 'string' || typeof record['color'] !== 'string') {
      throw new Error('boardgame-player-roster-item: player chip presentation event requires text and color strings');
    }
    this.chipText = record['text'];
    this.chipColor = record['color'];
  }

  private _textForChip(chipText: string, playerIndex: number, finished: boolean, winner: boolean): string {
    if (finished) {
      return winner ? "\u2605" : "\u2715";
    }
    return chipText ? chipText : String(playerIndex);
  }

  private playerDescription(isEmpty: boolean, isAgent: boolean, index: number, viewingAsPlayer: number): string {
    if (!this.mayBeActive && isEmpty) return "Waiting to be seated";
    if (!this.mayBeActive && !isEmpty) return "Sitting out";
    if (isEmpty) return "No one";
    if (isAgent) return "Robot";
    if (index === viewingAsPlayer) return "You";
    return "Human";
  }

  private classForPlayer(
    index: number,
    viewingAsPlayer: number,
    currentPlayerIndex: number,
    finished: boolean,
    winner: boolean
  ): string {
    const result: string[] = [];
    if (!this.mayBeActive) result.push("inactive");
    if (finished) result.push(winner ? "winner" : "loser");
    if (index === viewingAsPlayer) result.push("viewing");
    // AnyPlayerIndex (-3) means all players are "current" (simultaneous phase)
    if (currentPlayerIndex === -3 || index === currentPlayerIndex) result.push("current");
    return result.join(" ");
  }

  render() {
    return html`
      <div class="layout horizontal center ${this.classForPlayer(
        this.playerIndex,
        this.viewingAsPlayer,
        this.currentPlayerIndex,
        this.finished,
        this.winner
      )}">
        <div style="position:relative">
          <boardgame-player-chip
            .displayName="${this.displayName}"
            ?is-agent="${this.isAgent}"
            .photoUrl="${this.photoUrl}"
            .color="${this._effectiveColor}">
          </boardgame-player-chip>
          <strong
            class="chip"
            style=${styleMap(this._styleForChip(this._effectiveColor, this.finished, this.winner))}>
            ${this._textForChip(this.chipText, this.playerIndex, this.finished, this.winner)}
          </strong>
        </div>
        <div class="layout vertical">
          <strong class="${this.classForName(this.displayName)}">
            ${this.nameOrNobody(this.displayName)}
          </strong>
          <span>
            ${this.playerDescription(this.isEmpty, this.isAgent, this.playerIndex, this.viewingAsPlayer)}
          </span>
          <boardgame-render-player-info
            .state="${this.state}"
            .playerIndex="${this.playerIndex}"
            .rendererLoaded="${this.rendererLoaded}"
            .gameName="${this.gameName}"
            @player-chip-presentation-changed=${this._chipPresentationChanged}
            ?active="${this.active}">
          </boardgame-render-player-info>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-roster-item': BoardgamePlayerRosterItem;
  }
}
