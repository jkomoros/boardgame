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
      font-size: 20px;
      font-weight: 500;
      letter-spacing: 0.005em;
      color: var(--primary-text-color, #212121);
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

    strong.chip {
      font-size: 12px;
      font-weight: 400;
      background-color: var(--disabled-text-color, #9e9e9e);
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
      background-color: var(--light-accent-color, #ff4081);
      box-shadow: 0 0 0 4px var(--light-accent-color, #ff4081);
    }

    span {
      font-size: 12px;
      font-weight: 400;
      color: var(--secondary-text-color, #757575);
    }

    .viewing span {
      font-weight: bold;
      color: var(--accent-color, #ff4081);
    }

    boardgame-render-player-info {
      font-size: 12px;
      font-weight: 400;
      overflow: visible;
    }
  `;

  @property({ type: String })
  gameName = '';

  @property({ type: Boolean })
  isEmpty = false;

  @property({ type: Boolean })
  isAgent = false;

  @property({ type: Boolean })
  active = false;

  @property({ type: String })
  photoUrl = '';

  @property({ type: String })
  displayName = '';

  @property({ type: Object })
  state: unknown = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  currentPlayerIndex = 0;

  @property({ type: Boolean })
  finished = false;

  @property({ type: Boolean })
  winner = false;

  @property({ type: Boolean })
  rendererLoaded = false;

  @property({ type: String })
  chipText = '';

  @property({ type: String })
  chipColor = '';

  private nameOrNobody(displayName: string): string {
    return displayName ? displayName : "Nobody";
  }

  private classForName(displayName: string): string {
    if (!displayName) return "nobody";
    return "";
  }

  private _styleForChip(chipColor: string, finished: boolean, winner: boolean): string {
    if (finished) {
      return "box-shadow: none; background-color: " +
        (winner ? "#2e7d32" : "#e57373"); // Material green-800 / red-300
    }
    if (!chipColor) return "box-shadow: none";
    return "background-color: " + chipColor;
  }

  private _textForChip(chipText: string, playerIndex: number, finished: boolean, winner: boolean): string {
    if (finished) {
      return winner ? "\u2605" : "\u2715";
    }
    return chipText ? chipText : String(playerIndex);
  }

  private playerDescription(isEmpty: boolean, isAgent: boolean, index: number, viewingAsPlayer: number): string {
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
    if (finished) result.push(winner ? "winner" : "loser");
    if (index === viewingAsPlayer) result.push("viewing");
    if (index === currentPlayerIndex) result.push("current");
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
            .photoUrl="${this.photoUrl}">
          </boardgame-player-chip>
          <strong
            class="chip"
            style="${this._styleForChip(this.chipColor, this.finished, this.winner)}">
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
            ?renderer-loaded="${this.rendererLoaded}"
            .gameName="${this.gameName}"
            .chipText="${this.chipText}"
            .chipColor="${this.chipColor}"
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
