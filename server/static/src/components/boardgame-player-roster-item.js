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

import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import './boardgame-player-chip.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-styles/color.js';
import './boardgame-render-player-info.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgamePlayerRosterItem extends PolymerElement {
  static get template() {
    return html`
    <style is="custom-style" include="iron-flex">
      strong {
        @apply --paper-font-title;
        color: var(--primary-text-color);
      }

      boardgame-player-chip {
        padding-right: 10px;
      }

      .nobody {
        opacitY:0.5;
      }

      .loser {
        filter: saturate(0.5) brightness(1.5) blur(1px);
      }

      strong.chip {
        @apply --paper-font-caption;
        background-color: var(--disabled-text-color);
        color: white;
        padding:0.25em;
        height:1em;
        width:1em;
        box-sizing:content-box;
        text-align:center;
        border-radius:50%;
        position:absolute;
        text-overflow: initial;


        /* TODO: the following are all a nudging hack */
        line-height:14px;
        bottom:0.5em;
        right:1.5em;
      }

      .current strong.chip {
        background-color: var(--light-accent-color);
        box-shadow: 0 0 0 4px var(--light-accent-color);
      }

      span {
        @apply --paper-font-caption;
        color: var(--secondary-text-color);
      }

      .viewing span {
        font-weight:bold;
        color: var(--accent-color);
      }

      boardgame-render-player-info {
        @apply --paper-font-caption;
        /* --paper-font-caption sets overflow, but we want it to not be set so boardgame-status-text will not clip */
        overflow:visible;
      }

    </style>
    <div class\$="layout horizontal center [[classForPlayer(playerIndex, viewingAsPlayer, currentPlayerIndex, finished, winner)]]">
      <div style="position:relative">
        <boardgame-player-chip display-name="[[displayName]]" is-agent="[[isAgent]]" photo-url="[[photoUrl]]"></boardgame-player-chip>
        <strong class="chip" style\$="[[_styleForChip(chipColor, finished, winner)]]">[[_textForChip(chipText, playerIndex, finished, winner)]]</strong>
      </div>
      <div class="layout vertical">
        <strong class\$="[[classForName(displayName)]]">[[nameOrNobody(displayName)]]</strong>
        <span>[[playerDescription(isEmpty, isAgent, playerIndex, viewingAsPlayer)]]</span>
        <boardgame-render-player-info state="[[state]]" player-index="[[playerIndex]]" renderer-loaded="[[rendererLoaded]]" game-name="[[gameName]]" chip-text="{{chipText}}" chip-color="{{chipColor}}" active="[[active]]"></boardgame-render-player-info>
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-player-roster-item"
  }

  static get properties() {
    return {
      gameName: String,
      isEmpty: {
        type: Boolean,
        value: false,
      },
      isAgent: {
        type: Boolean,
        value: false,
      },
      active: Boolean,
      photoUrl: String,
      displayName: String,
      state: Object,
      playerIndex: Number,
      viewingAsPlayer: Number,
      currentPlayerIndex: Number,
      finished: Boolean,
      winner: Boolean,
      rendererLoaded: {
        type: Boolean,
        value: false,
      },
      chipText: {
        type: String,
        value: "",
      },
      chipColor: {
        type: String,
        value: "",
      }
    }
  }

  nameOrNobody(displayName) {
    return (displayName) ? displayName : "Nobody"
  }

  classForName(displayName) {
    if (!displayName) return "nobody"
    return ""
  }

  _styleForChip(chipColor, finished, winner) {
    if (finished) {
      return "box-shadow: none; background-color: " + (winner ? "var(--paper-green-800)" : "var(--paper-red-300)");
    }
    if (!chipColor) return "box-shadow: none";
    return "background-color: " + chipColor;
  }

  _textForChip(chipText, playerIndex, finished, winner) {
    if (finished) {
      return winner ? "\u2605" : "\u2715";
    }
    return (chipText) ? chipText : playerIndex;
  }

  playerDescription(isEmpty, isAgent, index,  viewingAsPlayer) {
    if (isEmpty) return "No one";
    if (isAgent) return "Robot";
    if (index == viewingAsPlayer) return "You";
    return "Human";
  }

  classForPlayer(index, viewingAsPlayer, currentPlayerIndex, finished, winner) {
    var result = [];
    if (finished) result.push(winner ? "winner" : "loser");
    if (index == viewingAsPlayer) result.push("viewing");
    if (index == currentPlayerIndex) result.push("current");
    return result.join(" ");
  }
}

customElements.define(BoardgamePlayerRosterItem.is, BoardgamePlayerRosterItem);
