/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { PolymerElement } from '@polymer/polymer/polymer-element.js';

import './shared-styles.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgamePlayerChip extends PolymerElement {
  static get template() {
    return html`
    <style is="custom-style" include="iron-flex shared-styles">
      .photo {
        height: var(--player-chip-size, 50px);
        width: var(--player-chip-size, 50px);
        border-radius: 50%;
        margin-right: 0.5em;
        background-color:hsl(0, 0%, 90%);
        transition: background-color 1s ease-in-out;
      }

    </style>

    <img id="chip" src\$="[[_effectivePhotoUrl(photoUrl, isAgent)]]" class="photo">
`;
  }

  static get is() {
    return "boardgame-player-chip"
  }

  static get properties() {
    return {
      photoUrl: {
        type: String,
        value: "",
      },
      displayName: {
        type: String,
        observer: "_displayNameChanged"
      },  
      //Whether or not the user is an agent or not
      isAgent: {
        type: Boolean,
        value: false,
      }
    }
  }

  _effectivePhotoUrl(photoUrl, isAgent) {
    if (isAgent) return "src/assets/agent.svg";
    return (photoUrl) ? photoUrl : "src/assets/player.svg";
  }

  _displayNameChanged(newValue) {

    var result = "hsl(0, 0%, 90%)";

    if (newValue) {
      var hash = this._hashString(newValue);
      //Hash is between Number.MIN_VALUE and Number.MAX_VALUE, but needs to
      //be between 0 and 360

      var degree = hash % 360;
      result = "hsl(" + degree + ", 100%, 50%)";
    }

    this.$.chip.style.backgroundColor = result;

  }

  _hashString(str) {
    //Based on code at http://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript-jquery
    var hash = 0, i, chr;
    if (str.length === 0) return hash;
    for (i = 0; i < str.length; i++) {
      chr   = str.charCodeAt(i);
      hash  = ((hash << 5) - hash) + chr;
      hash |= 0; // Convert to 32bit integer
    }
    return hash;
  }
}

customElements.define(BoardgamePlayerChip.is, BoardgamePlayerChip);
