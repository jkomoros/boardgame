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
import { customElement, property, query } from 'lit/decorators.js';

@customElement('boardgame-player-chip')
export class BoardgamePlayerChip extends LitElement {
  static styles = css`
    .photo {
      height: var(--player-chip-size, 50px);
      width: var(--player-chip-size, 50px);
      border-radius: 50%;
      margin-right: 0.5em;
      background-color: hsl(0, 0%, 90%);
      transition: background-color 1s ease-in-out;
    }
  `;

  @property({ type: String })
  photoUrl = '';

  @property({ type: String })
  displayName = '';

  @property({ type: Boolean })
  isAgent = false;

  @query('#chip')
  private chip!: HTMLImageElement;

  protected updated(changedProperties: Map<string, any>): void {
    if (changedProperties.has('displayName')) {
      this._updateBackgroundColor();
    }
  }

  private _effectivePhotoUrl(): string {
    if (this.isAgent) return 'src/assets/agent.svg';
    return this.photoUrl ? this.photoUrl : 'src/assets/player.svg';
  }

  private _updateBackgroundColor(): void {
    let result = 'hsl(0, 0%, 90%)';

    if (this.displayName) {
      const hash = this._hashString(this.displayName);
      // Hash is between Number.MIN_VALUE and Number.MAX_VALUE, but needs to
      // be between 0 and 360
      const degree = hash % 360;
      result = `hsl(${degree}, 100%, 50%)`;
    }

    if (this.chip) {
      this.chip.style.backgroundColor = result;
    }
  }

  private _hashString(str: string): number {
    // Based on code at http://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript-jquery
    let hash = 0;
    if (str.length === 0) return hash;

    for (let i = 0; i < str.length; i++) {
      const chr = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + chr;
      hash |= 0; // Convert to 32bit integer
    }
    return hash;
  }

  render() {
    return html`
      <img id="chip" src="${this._effectivePhotoUrl()}" class="photo">
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-chip': BoardgamePlayerChip;
  }
}
