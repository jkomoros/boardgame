import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import './shared-styles.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import './boardgame-configure-game-properties.js';
import '@polymer/paper-styles/typography.js';
import './boardgame-player-chip.js';
import {GamePathMixin} from './boardgame-game-path.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameGameItem extends GamePathMixin(PolymerElement) {
  static get template() {
    return html`
    <style include="shared-styles iron-flex">
      .minor {
        @apply --paper-font-caption;
      }
      :host {
        --player-chip-size: 32px;
      }
      .empty {
        font-style: italic;
      }
      boardgame-player-chip {
        margin-left: 0.5em;
      }
    </style>
    <div class="card layout horizontal center">
      <a href="{{GamePath(item.Name, item.ID)}}">{{gameDisplayName}}</a>
      <template is="dom-repeat" items="{{item.Players}}">
        <boardgame-player-chip photo-url="{{item.PhotoUrl}}" display-name="{{item.DisplayName}}" is-agent="{{item.IsAgent}}"></boardgame-player-chip>
      </template>
      <span class="minor">Last activity {{item.ReadableLastActivity}}</span>
      <div class="flex"></div>
      <span class="minor">{{item.ID}}</span>
      <boardgame-configure-game-properties game-open="{{item.Open}}" game-visible="{{item.Visible}}"></boardgame-configure-game-properties>
    </div>
`;
  }

  static get is() {
    return "boardgame-game-item"
  }

  static get properties() {
    return {
      item: Object,
      managers: Array,
      gameDisplayName: {
        type: String,
        computed: "_computeGameDisplayName(item, managers)"
      }
    }
  }

  _playerItemClasses(playerItem) {
    return playerItem.IsEmpty ? "empty" : "";
  }

  _displayNameForPlayerItem(playerItem) {
    return playerItem.IsEmpty ? "No one" : playerItem.DisplayName;
  }

  _computeGameDisplayName(item, managers) {
    if (!item) return "";
    if (!managers) return "";
    for (let i = 0; i < managers.length; i++) {
      let manager = managers[i];
      if (manager.Name == item.Name) {
        return manager.DisplayName;
      }
    }
    return item.Name;
  }
}

customElements.define(BoardgameGameItem.is, BoardgameGameItem);
